package middleware

import (
	"net/http"
	"time"

	"github.com/kayumovtd/url-shortener/internal/service"
)

const cookieName = "auth_token"

func AuthMiddleware(auth *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var userID string

			cookie, err := r.Cookie(cookieName)
			if err == nil {
				userID, err = auth.ParseAndVerify(cookie.Value)
			}

			if err != nil {
				// По заданию:
				// Выдавать пользователю симметрично подписанную куку, содержащую уникальный идентификатор пользователя,
				// если такой куки не существует или она не проходит проверку подлинности.
				var token string
				userID, token, err = auth.GenerateNewToken()
				if err != nil {
					http.Error(w, "failed to generate token", http.StatusInternalServerError)
					return
				}

				http.SetCookie(w, &http.Cookie{
					Name:     cookieName,
					Value:    token,
					Path:     "/",
					HttpOnly: true,
					Secure:   false, // true — если HTTPS
					Expires:  time.Now().Add(365 * 24 * time.Hour),
				})
			}

			ctx := auth.WithUserID(r.Context(), userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
