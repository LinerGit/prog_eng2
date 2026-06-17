package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type User struct {
	ID    string `json:"user_id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type contextKey string

const UserKey contextKey = "user"

func Auth(authServiceURL string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			authHeader := r.Header.Get("Authorization")

			if authHeader == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			client := &http.Client{
				Timeout: 3 * time.Second,
			}

			req, err := http.NewRequest(
				http.MethodGet,
				authServiceURL+"/api/auth/validate-token",
				nil,
			)

			if err != nil {
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}

			req.Header.Set("Authorization", authHeader)

			resp, err := client.Do(req)

			if err != nil {
				http.Error(w, "auth service unavailable", http.StatusUnauthorized)
				return
			}

			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			var user User

			if err := json.NewDecoder(resp.Body).Decode(&user); err == nil {
				ctx := context.WithValue(
					r.Context(),
					UserKey,
					user,
				)

				r = r.WithContext(ctx)
			}

			next.ServeHTTP(w, r)
		})
	}
}
