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

func getUserID(c *gin.Context)(int64, bool) {

	value, exists := c.Get("user_id")

	if !exists {
		return 0, false
	}

	userID, ok := value.(int64)
	
	if !ok {
		return 0, false
	}	

	return userID, true

}