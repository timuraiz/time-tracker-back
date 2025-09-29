package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
	Role  string `json:"role"`
	jwt.RegisteredClaims
}

// SupabaseAuth middleware - simple JWT validation
func SupabaseAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := tokenParts[1]

		// Simple approach: Use JWT secret for verification
		jwtSecret := os.Getenv("JWT_SECRET")
		if jwtSecret == "" {
			// Fallback to development mode - parse without verification
			token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &Claims{})
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
				c.Abort()
				return
			}

			claims, ok := token.Claims.(*Claims)
			if !ok {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
				c.Abort()
				return
			}

			userID, err := uuid.Parse(claims.Sub)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID in token"})
				c.Abort()
				return
			}

			c.Set("user_id", userID)
			c.Set("user_email", claims.Email)
			c.Set("user_role", claims.Role)
			c.Next()
			return
		}

		// Production: verify with JWT secret
		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(*Claims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		userID, err := uuid.Parse(claims.Sub)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID in token"})
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)

		c.Next()
	}
}

// OptionalAuth middleware - simple version
func OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.Next()
			return
		}

		tokenString := tokenParts[1]
		jwtSecret := os.Getenv("JWT_SECRET")

		if jwtSecret == "" {
			// Development mode: parse without verification
			token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &Claims{})
			if err == nil {
				if claims, ok := token.Claims.(*Claims); ok {
					if userID, err := uuid.Parse(claims.Sub); err == nil {
						c.Set("user_id", userID)
						c.Set("user_email", claims.Email)
						c.Set("user_role", claims.Role)
					}
				}
			}
		} else {
			// Production: verify with secret
			token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
				return []byte(jwtSecret), nil
			})

			if err == nil && token.Valid {
				if claims, ok := token.Claims.(*Claims); ok {
					if userID, err := uuid.Parse(claims.Sub); err == nil {
						c.Set("user_id", userID)
						c.Set("user_email", claims.Email)
						c.Set("user_role", claims.Role)
					}
				}
			}
		}

		c.Next()
	}
}
