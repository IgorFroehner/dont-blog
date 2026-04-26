---
title: "Hello World"
date: 2026-04-04
tags: ["go", "blog"]
draft: true
excerpt: "Welcome to my new blog, built with a custom Go static site generator."
---

Welcome to my blog! This is my first post, written in Markdown and rendered by a custom static site generator built in Go.

## Why Build a Custom SSG?

There are great tools like Hugo and Jekyll, but building your own is a great way to learn. Plus, you get exactly the features you want — nothing more.

## Features

Here's what this generator supports:

- **Markdown** with full CommonMark + GFM extensions
- **Syntax highlighting** with Chroma
- **Dark/light mode** that respects your system preference
- **RSS feed** for subscribers
- **Client-side search** across all posts

### Code Example

Here's a simple Go HTTP server:

```go
package main

import (
    "fmt"
    "net/http"
)

func main() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "Hello, World!")
    })
    http.ListenAndServe(":8080", nil)
}
```

### A Table

| Feature | Status |
|---------|--------|
| Markdown | Done |
| Dark Mode | Done |
| RSS | Done |
| Search | Done |

## What's Next?

I'll be writing about Go, web development, and the tools I use. Stay tuned!
