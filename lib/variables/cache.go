package variables

import (
	"github.com/minor-industries/z2/lib/data"
	"github.com/pkg/errors"
	"sync"
)

type Storage interface {
	Load() ([]data.Variable, error)
	Save(value *data.Variable) error
}

type Variable struct {
	Name    string
	Value   float64
	Present bool
}

type Cache struct {
	lock sync.Mutex
	vars map[string]float64
	db   Storage
}

func NewCache(db Storage) (*Cache, error) {
	cache := &Cache{
		db: db,
		vars: map[string]float64{
			// TODO: this should be elsewhere
			"bike_target_speed":      41.5,
			"bike_allowed_error_pct": 1.0,
			"bike_max_drift_pct":     2.0,

			"rower_target_power":      115.0,
			"rower_allowed_error_pct": 2.0,
			"rower_max_drift_pct":     5.0,
		},
	}

	rows, err := db.Load()
	if err != nil {
		return nil, errors.Wrap(err, "load rows")
	}

	for _, row := range rows {
		cache.vars[row.Name] = row.Value
	}
	return cache, nil
}

func (c *Cache) Get(keys []string) []Variable {
	c.lock.Lock()
	defer c.lock.Unlock()

	var result []Variable
	for _, k := range keys {
		v, present := c.vars[k]
		result = append(result, Variable{
			Name:    k,
			Value:   v,
			Present: present,
		})
	}

	return result
}

func (c *Cache) Update(vars []Variable) {
	c.lock.Lock()
	defer c.lock.Unlock()

	for _, v := range vars {
		if v.Present {
			c.vars[v.Name] = v.Value

			// TODO: errors?
			_ = c.db.Save(&data.Variable{
				Name:  v.Name,
				Value: v.Value,
			})
		}
	}
}

func (c *Cache) GetOne(key string) (float64, bool) {
	get := c.Get([]string{key})
	return get[0].Value, get[0].Present
}
