package main

import (
	"net/http"
)

func postMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "This route only accepts POST request", 400)
			return
		}

		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Content-Type not set to application/json", 400)
			return
		}

		if r.Body == nil {
			http.Error(w, "Request body is missing", 400)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

func originMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		referer := r.Header.Get("Referer")
		if origin != "foo.bar" && referer != "foo.bar" {
			http.Error(w, "Origin nor Referer headers set properly", 400)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func cookieMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		
		if !verifyAccessToken(r.Header.Get("Cookie") {
			http.Error(w, "Access token is unauthorized", 400)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}
