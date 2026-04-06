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
