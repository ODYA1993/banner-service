package config

import (
	"banner-service/internal/models/user"
	"github.com/golang-jwt/jwt"
	"net/http"
	"net/url"
	"time"
)

func jwtMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Получаем куку с именем "token" из запроса
		cookie, err := r.Cookie("token")
		if err != nil {
			http.Error(w, "missing token cookie", http.StatusUnauthorized)
			return
		}

		var claims user.CustomClaims

		// Декодируем токен из куки с помощью секретного ключа
		tokenString, err := url.QueryUnescape(cookie.Value)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		_, err = jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
			return []byte("mysecretkey"), nil
		})
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Проверяем, не истек ли срок действия токена
		if time.Now().After(claims.ExpiresAt.Time) {
			http.Error(w, "token expired", http.StatusUnauthorized)
			return
		}

		// Проверяем, включен ли баннер для обычных пользователей
		if !claims.IsAdmin {
			http.Error(w, "user does not have access", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	}
}
