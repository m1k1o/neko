package utils

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/render"
	"github.com/dgrijalva/jwt-go"
)

type key int

const (
    keyPrincipalID key = iota
)

func GetUserName(r *http.Request) interface{} {
	props, _ := r.Context().Value(keyPrincipalID).(jwt.MapClaims)
	return props["user_name"]
}

type HttpMiddleware = func(next http.Handler) http.Handler

func AuthMiddleware(next http.Handler, jwtSecrets ...[]byte) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := strings.Split(r.Header.Get("Authorization"), "Bearer ")
		if len(authHeader) != 2 {
			_ = render.Render(w, r, ErrMessage(401, "Malformed JWT token."))
			return
		}

		jwtToken := authHeader[1]
		var jwtVerified *jwt.Token
		var err error
		for _, jwtSecret := range jwtSecrets {
			jwtVerified, err = jwt.Parse(jwtToken, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
				}

				return jwtSecret, nil
			})

			if err == nil {
				break
			}
		}

		if err != nil {
			_ = render.Render(w, r, ErrMessage(401, "Invalid JWT token."))
			return
		}

		if claims, ok := jwtVerified.Claims.(jwt.MapClaims); ok && jwtVerified.Valid {
			ctx := context.WithValue(r.Context(), keyPrincipalID, claims)
			// Access context values in handlers like this
			// props, _ := r.Context().Value("props").(jwt.MapClaims)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			_ = render.Render(w, r, ErrMessage(401, "Unauthorized."))
		}
	})
}
