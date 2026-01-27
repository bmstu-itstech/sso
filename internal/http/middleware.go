package http_server

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Middleware для проверки JWT токена
func (s *ServerGin) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			return
		}

		token := parts[1]

		tokenParse, err := s.auth.SignIn(c.Request.Context(), token, ssoAppID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}
		c.Set("userId", tokenParse.Uid)
		c.Set("jwtCleanToken", token)
		c.Set("isAdmin", tokenParse.IsAdmin)
		c.Set("appId", tokenParse.AppId)

		c.Next()
	}
}
