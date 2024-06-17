package middleware

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

type ContextKey string

const UserClaimsKey ContextKey = "userClaims"

// ValidateJWT middleware to validate JWT and store claims in request context
func ValidateJWT(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header is required", http.StatusUnauthorized)
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenStr == authHeader {
			http.Error(w, "Bearer token is required", http.StatusUnauthorized)
			return
		}

		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(os.Getenv("JWT_SECRET_KEY")), nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Failed to parse token claims", http.StatusUnauthorized)
			return
		}

		// Store the claims in the request context
		ctx := context.WithValue(r.Context(), UserClaimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}


// package middleware

// import (
// 	"challenge/utils"
// 	"context"
// 	"os"
// 	"strings"

// 	"github.com/gofiber/fiber/v2"
// 	"github.com/golang-jwt/jwt/v4"
// )

// // ValidateJWT middleware function to validate JWT and store it in context
// func ValidateJWT() fiber.Handler {
// 	return func(c *fiber.Ctx) error {
// 		authHeader := c.Get("Authorization")
// 		if authHeader == "" {
// 			return c.Status(fiber.StatusUnauthorized).SendString("Authorization header is required")
// 		}

// 		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
// 		if tokenStr == authHeader {
// 			return c.Status(fiber.StatusUnauthorized).SendString("Bearer token is required")
// 		}

// 		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
// 			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
// 				return nil, fiber.NewError(fiber.StatusUnauthorized, "unexpected signing method")
// 			}
// 			return []byte(os.Getenv("JWT_SECRET_KEY")), nil
// 		})
// 		if err != nil || !token.Valid {
// 			return c.Status(fiber.StatusUnauthorized).SendString("Invalid token")
// 		}

// 		// Store the token in context
// 		ctx := context.WithValue(c.Context(), utils.UserContextKey, token)
// 		c.SetUserContext(ctx)

// 		return c.Next()
// 	}
// }
