package middleware

import (
	"IM-system/internal/auth"
	"IM-system/internal/httpserver"
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
		authorization := c.GetHeader("authorizationHeader")

		if authorization == "" {
			//终止请求，并返回 HTTP 状态码和 JSON 响应
			c.AbortWithStatusJSON(http.StatusUnauthorized,
				httpserver.Fail("missing authorization header"),
			)
			return
		}

		if !strings.HasPrefix(authorization, bearerPrefix) {
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				httpserver.Fail("invalid authorization format"),
			)
			return
		}
		tokenString := strings.TrimPrefix(
			authorization, bearerPrefix,
		)
		userID, err := m.jwtService.ParseToken(tokenString)
		if err != nil{
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				httpserver.Fail("invalid token"),
			)
			return 
		}
		c.Set("userID", userID)

		c.Next()
	}
}
