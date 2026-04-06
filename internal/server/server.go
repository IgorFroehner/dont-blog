package server

import (
	"bytes"
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/igor/my-go-site/internal/builder"
)

func Serve(configPath string, port string, templateFS, staticFS embed.FS) error {
	// Initial build
	fmt.Println("Building site...")
	if err := builder.Build(configPath, templateFS, staticFS); err != nil {
		return fmt.Errorf("initial build failed: %w", err)
	}

	// Reload state
	var (
		mu         sync.Mutex
		reloadChan = make(chan struct{}, 1)
	)

	rebuild := func() {
		mu.Lock()
		defer mu.Unlock()

		fmt.Println("Rebuilding...")
		start := time.Now()
		if err := builder.Build(configPath, templateFS, staticFS); err != nil {
			log.Printf("rebuild error: %v", err)
			return
		}
		fmt.Printf("Rebuilt in %s\n", time.Since(start).Round(time.Millisecond))

		select {
		case reloadChan <- struct{}{}:
		default:
		}
	}

	// File watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("creating watcher: %w", err)
	}
	defer watcher.Close()

	watchDirs := []string{"content", "templates", "static"}
	for _, dir := range watchDirs {
		if err := watchRecursive(watcher, dir); err != nil {
			log.Printf("warning: could not watch %s: %v", dir, err)
		}
	}
	if err := watcher.Add(configPath); err != nil {
		log.Printf("warning: could not watch %s: %v", configPath, err)
	}

	go func() {
		var debounce *time.Timer
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) || event.Has(fsnotify.Remove) {
					if debounce != nil {
						debounce.Stop()
					}
					debounce = time.AfterFunc(100*time.Millisecond, rebuild)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Printf("watcher error: %v", err)
			}
		}
	}()

	// HTTP server
	mux := http.NewServeMux()

	// Serve dist directory
	fileServer := http.FileServer(http.Dir("dist"))
	mux.Handle("/", fileServer)

	// Long-poll reload endpoint
	mux.HandleFunc("/___reload", func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-reloadChan:
			w.WriteHeader(http.StatusOK)
		case <-r.Context().Done():
			w.WriteHeader(http.StatusRequestTimeout)
		case <-time.After(30 * time.Second):
			w.WriteHeader(http.StatusNoContent)
		}
	})

	addr := ":" + port
	fmt.Printf("Serving at http://localhost%s\n", addr)
	fmt.Println("Watching for changes... (Ctrl+C to stop)")
	return http.ListenAndServe(addr, injectReloadScript(mux))
}

func watchRecursive(watcher *fsnotify.Watcher, root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return watcher.Add(path)
		}
		return nil
	})
}

// injectReloadScript wraps the handler to inject a live-reload script into HTML responses.
func injectReloadScript(next http.Handler) http.Handler {
	reloadScript := []byte(`<script>
(function(){
    function poll(){
        fetch('/___reload').then(function(r){
            if(r.status===200){location.reload()}
            else{setTimeout(poll,500)}
        }).catch(function(){setTimeout(poll,2000)})
    }
    poll()
})()
</script></body>`)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/___reload" {
			next.ServeHTTP(w, r)
			return
		}

		rec := &responseRecorder{ResponseWriter: w, body: nil}
		next.ServeHTTP(rec, r)

		body := rec.body
		if len(body) > 0 && (rec.contentType == "text/html" || hasHTMLExtension(r.URL.Path)) {
			replaced := replaceLastOccurrence(body, []byte("</body>"), reloadScript)
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(replaced)))
			w.WriteHeader(rec.statusCode)
			w.Write(replaced)
			return
		}

		w.WriteHeader(rec.statusCode)
		w.Write(body)
	})
}

type responseRecorder struct {
	http.ResponseWriter
	statusCode  int
	body        []byte
	contentType string
}

func (r *responseRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.contentType = r.Header().Get("Content-Type")
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	if r.statusCode == 0 {
		r.statusCode = 200
	}
	r.body = append(r.body, b...)
	return len(b), nil
}

func hasHTMLExtension(path string) bool {
	return filepath.Ext(path) == ".html" || path == "/" || filepath.Ext(path) == ""
}

func replaceLastOccurrence(s, old, new []byte) []byte {
	i := bytes.LastIndex(s, old)
	if i == -1 {
		return s
	}
	result := make([]byte, 0, len(s)-len(old)+len(new))
	result = append(result, s[:i]...)
	result = append(result, new...)
	result = append(result, s[i+len(old):]...)
	return result
}
