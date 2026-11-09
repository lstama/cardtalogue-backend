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
		if !Authenticate(c, jwtSecret, false) {
			c.Next()
			return
		}
		c.Next()
	}
}

func JWTAuth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !Authenticate(c, jwtSecret, true) {
			c.Abort()
			return
		}
		c.Next()
	}
}

// Authenticate verifies the Bearer token in the Authorization header. When
// require is true it aborts the chain with 401 on any failure; otherwise it
// leaves the request unauthenticated and returns false.
func Authenticate(c *gin.Context, jwtSecret string, require bool) bool {
	authHeader := c.GetHeader("Authorization")
	c.Set("isAuthenticated", false)

	if authHeader == "" {
		return false
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		return false
	}

	token, err := (&authServiceValidator{jwtSecret: jwtSecret}).ValidateToken(tokenString)
	if err != nil || !token.Valid {
		if require {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		}
		return false
	}

	claims, ok := token.Claims.(*service.JWTClaims)
	if !ok {
		if require {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
		}
		return false
	}

	c.Set("isAuthenticated", true)
	c.Set("userID", claims.UserID)
	c.Set("userEmail", claims.Email)
	c.Set("userRole", claims.Role)
	c.Set("user", gin.H{
		"id":    claims.UserID,
		"email": claims.Email,
		"role":  claims.Role,
	})
	return true
}

func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !HasRole(c, roles) {
			c.Abort()
			return
		}
		c.Next()
	}
}

// HasRole reports whether the role stored in the gin context is in roles.
// It writes the appropriate error response and aborts when it is not.
func HasRole(c *gin.Context, roles []string) bool {
	userRole, exists := c.Get("userRole")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user role not found"})
		return false
	}

	role := userRole.(string)
	for _, requiredRole := range roles {
		if role == requiredRole {
			return true
		}
	}

	c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
	return false
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
