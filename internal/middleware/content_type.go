package middleware

import (
	"net/http"
	"strings"
)

// ContentType ensures the request Content-Type is valid (JSON, form-data, or urlencoded).
func ContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
			ct := r.Header.Get("Content-Type")
			if ct != "" {
				isJSON := strings.HasPrefix(ct, "application/json")
				isForm := strings.HasPrefix(ct, "multipart/form-data")
				isURLEncoded := strings.HasPrefix(ct, "application/x-www-form-urlencoded")
				if !isJSON && !isForm && !isURLEncoded {
					http.Error(w, `{"success":false,"message":"Content-Type tidak didukung"}`, http.StatusUnsupportedMediaType)
					return
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}
