package middleware

import (
	"net/http"
	"strings"

	"github.com/lstama/cardtalogue-backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// OptionalJWTAuth middleware checks for JWT token but doesn't require it
// If token is valid, sets isAuthenticated=true, otherwise sets isAuthenticated=false
func OptionalJWTAuth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		// Default to unauthenticated
		c.Set("isAuthenticated", false)

		if authHeader == "" {
			c.Next()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			// No Bearer prefix, continue as unauthenticated
			c.Next()
			return
		}

		// Create a temporary auth service to validate token
		authService := &authServiceValidator{jwtSecret: jwtSecret}
		token, err := authService.ValidateToken(tokenString)
		if err != nil || !token.Valid {
			// Invalid token, continue as unauthenticated
			c.Next()
			return
		}

		claims, ok := token.Claims.(*service.JWTClaims)
		if !ok {
			// Invalid claims, continue as unauthenticated
			c.Next()
			return
		}

		// Token is valid, set as authenticated
		c.Set("isAuthenticated", true)
		c.Set("userID", claims.UserID)
		c.Set("userEmail", claims.Email)
		c.Set("userRole", claims.Role)
		c.Set("user", gin.H{
			"id":    claims.UserID,
			"email": claims.Email,
			"role":  claims.Role,
		})

		c.Next()
	}
}

func JWTAuth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}

		// Create a temporary auth service to validate token
		authService := &authServiceValidator{jwtSecret: jwtSecret}
		token, err := authService.ValidateToken(tokenString)
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(*service.JWTClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("userEmail", claims.Email)
		c.Set("userRole", claims.Role)
		c.Set("user", gin.H{
			"id":    claims.UserID,
			"email": claims.Email,
			"role":  claims.Role,
		})

		c.Next()
	}
}

func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("userRole")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user role not found"})
			c.Abort()
			return
		}

		role := userRole.(string)
		for _, requiredRole := range roles {
			if role == requiredRole {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		c.Abort()
	}
}

// Helper struct for token validation
type authServiceValidator struct {
	jwtSecret string
}

func (a *authServiceValidator) ValidateToken(tokenString string) (*jwt.Token, error) {
	return jwt.ParseWithClaims(tokenString, &service.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(a.jwtSecret), nil
	})
}
