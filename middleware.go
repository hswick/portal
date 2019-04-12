package main

import (
	"net/http"
	"fmt"
	"net/url"
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

//var localhost string = fmt.Sprintf("http://localhost%s/", config.Port)

//TODO: Review this origin policy, may still be insecure
//Probably need to check localhost
func badOrigin(origin string, referer string) bool {	
	return origin != config.Domain && referer != config.Domain && origin != "localhost" && referer != "localhost"
}

func originMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		
		origin := r.Header.Get("Origin")
		referer := r.Header.Get("Referer")
		
		if badOrigin(origin, referer) {
			refererUrl, _ := url.Parse(referer)
			errorMessage := fmt.Sprintf("Origin: %s nor Referer: %s are authorized", origin, refererUrl.Hostname())
			http.Error(w, errorMessage, 400)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

func cookieMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		
		if (!verifyAccessToken(r.Header.Get("Cookie"))) {
			http.Error(w, "Access token is unauthorized, yikes!", 400)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}
