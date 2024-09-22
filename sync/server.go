package sync

import (
	"github.com/chrispappas/golang-generics-set/set"
	"github.com/gin-gonic/gin"
	"github.com/minor-industries/rtgraph/database/sqlite"
	"github.com/pkg/errors"
	"github.com/tinylib/msgp/msgp"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"net/http"
)

type SyncResponse struct {
	ExistingItems int `json:"existing_items"`
	NewItems      int `json:"new_items"`
}

func insertSeriesBatchWithTransaction(
	db *gorm.DB,
	series NamedSeries,
) (int, error) {
	count := 0

	err := db.Transaction(func(tx *gorm.DB) error {
		seen := set.Set[string]{}

		for i := range series.Timestamps {
			id := sqlite.HashedID(series.Name)

			if err := addSeriesName(tx, seen, series.Name, id); err != nil {
				return errors.Wrap(err, "add series name")
			}

			row := sqlite.Sample{
				SeriesID:  id,
				Timestamp: series.Timestamps[i],
				Value:     series.Values[i],
			}

			res := tx.Clauses(clause.OnConflict{
				DoNothing: true,
			}).Create(&row)

			if res.Error != nil {
				return res.Error
			}

			if res.RowsAffected > 0 {
				count++
			}
		}
		return nil
	})

	if err != nil {
		return 0, err
	}

	return count, nil
}

func addSeriesName(tx *gorm.DB, seen set.Set[string], seriesName string, id []byte) error {
	if !seen.Has(seriesName) {
		res := tx.Clauses(clause.OnConflict{
			DoNothing: true,
		}).Create(&sqlite.Series{
			ID:   id,
			Name: seriesName,
		})

		if res.Error != nil {
			return errors.Wrap(res.Error, "create")
		}

		seen.Add(seriesName)
	}

	return nil
}

func RunServer(db *sqlite.Backend) error {
	r := gin.Default()

	r.POST("/sync/series", func(c *gin.Context) {
		var series NamedSeries
		if err := msgp.Decode(c.Request.Body, &series); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		count, err := insertSeriesBatchWithTransaction(db.GetORM(), series)
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
