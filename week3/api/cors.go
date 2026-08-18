package main

import "net/http"

func corsMiddleware(next http.Handler) http.Handler {
	allowedOrigins := map[string]bool{
		"http://localhost:8081": true, // Swagger UI
		"http://localhost:5173": true, // Vite UI
	}

	return http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		origin := r.Header.Get("Origin")

		if allowedOrigins[origin] {
			w.Header().Set(
				"Access-Control-Allow-Origin",
				origin,
			)
			w.Header().Set(
				"Access-Control-Allow-Methods",
				"GET, POST, PATCH, DELETE, OPTIONS",
			)
			w.Header().Set(
				"Access-Control-Allow-Headers",
				"Authorization, Content-Type, Accept",
			)
			w.Header().Add("Vary", "Origin")
		}

		// OPTIONS 预检请求。
		if r.Method == http.MethodOptions {
			if origin != "" && !allowedOrigins[origin] {
				http.Error(
					w,
					"origin not allowed",
					http.StatusForbidden,
				)
				return
			}

			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
