package sync

import (
	"github.com/gin-gonic/gin"
	"github.com/minor-industries/rtgraph/database/sqlite"
	"github.com/tinylib/msgp/msgp"
	"net/http"
)

type SyncResponse struct {
	ExistingItems int `json:"existing_items"`
	NewItems      int `json:"new_items"`
}

func RunServer(db *sqlite.Backend) error {
	r := gin.Default()
	r.POST("/sync", func(c *gin.Context) {
		var series NamedSeries
		if err := msgp.Decode(c.Request.Body, &series); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		count, err := InsertSeriesBatchWithTransaction(db.GetORM(), series)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		resp := SyncResponse{
			ExistingItems: len(series.Timestamps) - count,
			NewItems:      count,
		}
		c.JSON(http.StatusOK, resp)
	})
	return r.Run()
}
