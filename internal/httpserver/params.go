package httpserver

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func getRoomParam(c *gin.Context) (int64, bool) {
	room := c.Param("room")

	roomID, err := strconv.ParseInt(room, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, Fail("invalid room id"))
		return 0, false
	}
	return roomID, true
}

func getUserID(c *gin.Context) (int64, bool) {

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
