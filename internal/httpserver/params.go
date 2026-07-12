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
func getUserParam(c *gin.Context)(string, bool) {
	user := c.Param("user")
	if user == "" {
		c.JSON(http.StatusBadRequest, Fail("invalid user parameter"))
		return "", false
	}
	return user, true
}