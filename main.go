package main

import (
	"embed"
	"fmt"
	"os"

	"github.com/igor/my-go-site/internal/builder"
	"github.com/igor/my-go-site/internal/server"
)

//go:embed templates/*
var templateFS embed.FS

//go:embed static/*
var staticFS embed.FS

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "build":
		if err := builder.Build("site.yaml", templateFS, staticFS); err != nil {
			fmt.Fprintf(os.Stderr, "build failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Site built successfully → dist/")
	case "serve":
		port := "1313"
		if len(os.Args) > 2 {
			port = os.Args[2]
		}
		if err := server.Serve("site.yaml", port, templateFS, staticFS); err != nil {
			fmt.Fprintf(os.Stderr, "server error: %v\n", err)
			os.Exit(1)
		}
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: my-go-site <command>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  build          Build the static site to dist/")
	fmt.Println("  serve [port]   Start dev server with live reload (default: 1313)")
}
