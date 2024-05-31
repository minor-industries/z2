package variables

import "sync"

type Variable struct {
	Name    string
	Value   float64
	Present bool
}

type Cache struct {
	lock sync.Mutex
	vars map[string]float64
}

func NewCache() *Cache {
	return &Cache{
		vars: map[string]float64{
			"bike_target_speed": 42.0, // TODO: this should be in DB
		},
	}
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
		} else {
			delete(c.vars, v.Name)
		}
	}
}
