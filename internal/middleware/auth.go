package middleware

import (
	"IM-system/internal/auth"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

const (
	bearerPrefix        = "Bearer "
	authorizationHeader = "Authorization"
)

type AuthMiddleware struct {
	jwtService *auth.JWTService
}

func NewAuthMiddleware(jwtService *auth.JWTService) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService: jwtService,
	}
}

func (m *AuthMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		authorization := c.GetHeader(authorizationHeader)

		if authorization == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code": -1, "msg": "missing authorization header",
			})
			return
		}

		if !strings.HasPrefix(authorization, bearerPrefix) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code": -1, "msg": "invalid authorization format",
			})
			return
		}
		tokenString := strings.TrimPrefix(authorization, bearerPrefix)
		userID, err := m.jwtService.ParseToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code": -1, "msg": "invalid token",
			})
			return
		}
		c.Set("user_id", userID)

		c.Next()
	}
}
