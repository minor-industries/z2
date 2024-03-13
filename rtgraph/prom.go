package rtgraph

import (
	"github.com/minor-industries/codelab/cmd/z2/rtgraph/schema"
	"github.com/minor-industries/platform/common/metrics"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

func (g *Graph) publishPrometheusMetrics() {
	metricMap := map[string]*metrics.TimeoutGauge{}

	msgCh := g.broker.Subscribe()

	for message := range msgCh {
		switch m := message.(type) {
		case *schema.Series:
			if _, ok := metricMap[m.SeriesName]; !ok {
				tg := metrics.NewTimeoutGauge(15*time.Second, prometheus.GaugeOpts{
					Name: m.SeriesName,
				})
				metricMap[m.SeriesName] = tg
				err := prometheus.Register(tg.G)
				if err != nil {
					g.errCh <- errors.Wrap(err, "register prometheus metric")
				}
			}
			metricMap[m.SeriesName].Set(m.Value)
		}
	}
}
