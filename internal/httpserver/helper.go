package httpserver

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func getRoomParam(c *gin.Context) (string, bool) {
	room := c.Param("room")
	if room == "" {
		c.JSON(http.StatusBadRequest, Fail("invalid room parameter"))
		return "", false
	}
	return room, true
}
