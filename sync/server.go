package sync

import (
	"github.com/gin-gonic/gin"
	"github.com/tinylib/msgp/msgp"
	"net/http"
)

type SyncResponse struct {
	ExistingItems int `json:"existing_items"`
	NewItems      int `json:"new_items"`
}

func RunServer() error {
	r := gin.Default()
	r.POST("/sync", func(c *gin.Context) {
		var series NamedSeries
		if err := msgp.Decode(c.Request.Body, &series); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		existing, newItems := 0, 0
		resp := SyncResponse{
			ExistingItems: existing,
			NewItems:      newItems,
		}
		c.JSON(http.StatusOK, resp)
	})
	return r.Run()
}
