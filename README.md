# goshort

A URL shortener written in Go for personal use

### Why?
I wanted shorter URLs to garbage on my website.

### How?
Written entirely with the go standard library, I took some hints from Git. This shortener
generates a SHA-1 hash of the entire URL, and then uses the first 6 characters of the hash
to access the URLs.