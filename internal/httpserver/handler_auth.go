package httpserver

import (
	"IM-system/internal/auth"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest,
			Fail("incvalid request"),
		)
		return
	}

	result, err := h.authService.Login(
		c.Request.Context(),
		auth.LoginInput{
			Username: req.Username,
			Password: req.Password,
		},
	)

	if err != nil {
		c.JSON(
			http.StatusUnauthorized,
			Fail(err.Error()),
		)
		return
	}

	c.JSON(http.StatusOK, OK(result))
}

func (h *Handler) Register(c *gin.Context){
	var req RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil{
		c.JSON(http.StatusBadRequest,
		Fail("invalid request"),
	)
		return
	}
	

	err := h.authService.Register(
		c.Request.Context(),
		auth.RegisterInput{
			Username: req.Username,
			Password: req.Password,
			Nickname: req.Nickname,
		},
	)
	if err != nil{
		c.JSON(http.StatusBadRequest,
		Fail(err.Error()),
	)
		return
	}
	

	c.JSON(
		http.StatusOK,
		OK(nil),
	)
}
