// Dark/light mode toggle
(function () {
    function setTheme(dark) {
        document.documentElement.classList.toggle('dark', dark);
        localStorage.setItem('theme', dark ? 'dark' : 'light');
    }

    function toggleTheme() {
        setTheme(!document.documentElement.classList.contains('dark'));
    }

    document.querySelectorAll('#theme-toggle, #theme-toggle-mobile').forEach(function (btn) {
        btn.addEventListener('click', toggleTheme);
    });

    // Mobile menu toggle
    var menuToggle = document.getElementById('menu-toggle');
    var mobileMenu = document.getElementById('mobile-menu');
    if (menuToggle && mobileMenu) {
        menuToggle.addEventListener('click', function () {
            mobileMenu.classList.toggle('hidden');
        });
    }
})();

// ASCII hero mouse-repel effect
(function () {
    var hero = document.getElementById('ascii-hero');
    if (!hero) return;

    var GLYPHS = ['▓','░','▒','╳','◈','◉','⬡','⬢','✦','⊕','⊗'];
    var RADIUS = 30;
    var RESTORE_MAX_DELAY = 300;

    // Wrap each character in a span, preserve newlines
    var text = hero.textContent;
    hero.textContent = '';
    var spans = [];
    for (var i = 0; i < text.length; i++) {
        var ch = text[i];
        if (ch === '\n') {
            hero.appendChild(document.createTextNode('\n'));
            continue;
        }
        var span = document.createElement('span');
        span.textContent = ch;
        span.setAttribute('data-char', ch);
        hero.appendChild(span);
        spans.push(span);
    }

    // Cache positions
    var positions = [];
    function cachePositions() {
        var heroRect = hero.getBoundingClientRect();
        positions = [];
        for (var i = 0; i < spans.length; i++) {
            var r = spans[i].getBoundingClientRect();
            positions.push({
                x: r.left + r.width / 2 - heroRect.left,
                y: r.top + r.height / 2 - heroRect.top
            });
        }
    }
    cachePositions();

    var restoreTimers = new Array(spans.length);
    var displaced = new Array(spans.length);
    var rafPending = false;

    function onMove(e) {
        if (rafPending) return;
        rafPending = true;
        requestAnimationFrame(function () {
            rafPending = false;
            var heroRect = hero.getBoundingClientRect();
            var mx = e.clientX - heroRect.left;
            var my = e.clientY - heroRect.top;

            for (var i = 0; i < spans.length; i++) {
                var dx = positions[i].x - mx;
                var dy = positions[i].y - my;
                var dist = Math.sqrt(dx * dx + dy * dy);

                if (dist < RADIUS) {
                    // Scatter
                    if (restoreTimers[i]) {
                        clearTimeout(restoreTimers[i]);
                        restoreTimers[i] = null;
                    }
                    if (!displaced[i]) {
                        spans[i].textContent = GLYPHS[Math.floor(Math.random() * GLYPHS.length)];
                        spans[i].style.opacity = '0.4';
                        displaced[i] = true;
                    }
                } else if (displaced[i] && !restoreTimers[i]) {
                    // Schedule restore
                    restoreTimers[i] = scheduleRestore(i);
                }
            }
        });
    }

    function scheduleRestore(idx) {
        var delay = Math.random() * RESTORE_MAX_DELAY;
        return setTimeout(function () {
            spans[idx].textContent = spans[idx].getAttribute('data-char');
            spans[idx].style.opacity = '';
            displaced[idx] = false;
            restoreTimers[idx] = null;
        }, delay);
    }

    function onLeave() {
        for (var i = 0; i < spans.length; i++) {
            if (displaced[i]) {
                if (restoreTimers[i]) clearTimeout(restoreTimers[i]);
                restoreTimers[i] = scheduleRestore(i);
            }
        }
    }

    hero.parentElement.addEventListener('mousemove', onMove);
    hero.parentElement.addEventListener('mouseleave', onLeave);

    // Recache positions on resize (debounced)
    var resizeTimer;
    window.addEventListener('resize', function () {
        clearTimeout(resizeTimer);
        resizeTimer = setTimeout(cachePositions, 200);
    });
})();

// Client-side search (only on blog page)
(function () {
    var searchInput = document.getElementById('search-input');
    if (!searchInput) return;

    var postList = document.getElementById('post-list');
    var noResults = document.getElementById('no-results');
    var posts = postList ? postList.querySelectorAll('.post-item') : [];
    var debounceTimer;

    searchInput.addEventListener('input', function () {
        clearTimeout(debounceTimer);
        debounceTimer = setTimeout(filterPosts, 150);
    });

    function filterPosts() {
        var query = searchInput.value.toLowerCase().trim();
        if (!query) {
            posts.forEach(function (el) { el.style.display = ''; });
            noResults.classList.add('hidden');
            return;
        }

        var visible = 0;
        posts.forEach(function (el) {
            var title = (el.dataset.title || '').toLowerCase();
            var tags = (el.dataset.tags || '').toLowerCase();
            var excerpt = (el.dataset.excerpt || '').toLowerCase();
            var match = title.includes(query) || tags.includes(query) || excerpt.includes(query);
            el.style.display = match ? '' : 'none';
            if (match) visible++;
        });

        noResults.classList.toggle('hidden', visible > 0);
    }
})();
