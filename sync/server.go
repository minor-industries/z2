//go:build !wasm

package sync

import (
	"fmt"
	"github.com/chrispappas/golang-generics-set/set"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/minor-industries/rtgraph/database/sqlite"
	"github.com/minor-industries/z2/cfg"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/tinylib/msgp/msgp"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

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

func insertMarkersBatchWithTransaction(
	db *gorm.DB,
	markers []Marker,
) (int, error) {
	count := 0

	err := db.Transaction(func(tx *gorm.DB) error {
		for _, marker := range markers {
			row := sqlite.Marker{
				ID:        marker.ID,
				Type:      marker.Type,
				Ref:       marker.Ref,
				Timestamp: marker.Timestamp,
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

func SetupRoutes(r *gin.Engine, env *cfg.SyncServerConfig) error {
	dbs := lo.Associate(env.Databases, func(name string) (string, *sqlite.Backend) {
		dbFile := filepath.Join(os.ExpandEnv(env.DBPath), name+".db")
		fmt.Println("opening db", dbFile)
		dst, _ := sqlite.Get(dbFile)
		return name, dst
	})

	if lo.Contains(lo.Values(dbs), nil) {
		return errors.New("couldn't open one or more dbs")
	}

	r.Use(cors.New(cors.Config{
		AllowOrigins:  []string{"*"},
		AllowMethods:  []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:  []string{"*"},
		ExposeHeaders: []string{"Content-Length"},
		MaxAge:        12 * time.Hour,
	}))

	r.POST("/sync/:db/series", func(c *gin.Context) {
		dbParam := c.Param("db")
		db, ok := dbs[dbParam]
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "unknown db"})
			return
		}

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

	r.POST("/sync/:db/markers", func(c *gin.Context) {
		dbParam := c.Param("db")
		db, ok := dbs[dbParam]
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "unknown db"})
			return
		}

		var markers Markers
		if err := msgp.Decode(c.Request.Body, &markers); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		count, err := insertMarkersBatchWithTransaction(db.GetORM(), markers)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		resp := SyncResponse{
			ExistingItems: len(markers) - count,
			NewItems:      count,
		}
		c.JSON(http.StatusOK, resp)
	})

	return nil
}

func RunServer(dbNames []string) error {
	r := gin.Default()

	if err := SetupRoutes(r, &cfg.SyncServerConfig{
		Enable:    true,
		Databases: dbNames,
		DBPath:    ".",
	}); err != nil {
		return errors.Wrap(err, "setup routes")
	}

	return r.Run()
}
