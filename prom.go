package main

import (
	"github.com/minor-industries/codelab/cmd/bike/schema"
	"github.com/minor-industries/platform/common/broker"
	"github.com/minor-industries/platform/common/metrics"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

func publishPrometheusMetrics(
	errCh chan error,
	br *broker.Broker,
) {
	metricMap := map[string]*metrics.TimeoutGauge{}

	msgCh := br.Subscribe()

	for message := range msgCh {
		switch m := message.(type) {
		case *schema.Series:
			if _, ok := metricMap[m.SeriesName]; !ok {
				g := metrics.NewTimeoutGauge(15*time.Second, prometheus.GaugeOpts{
					Name: m.SeriesName,
				})
				metricMap[m.SeriesName] = g
				err := prometheus.Register(g.G)
				if err != nil {
					errCh <- errors.Wrap(err, "register prometheus metric")
				}
			}
			metricMap[m.SeriesName].Set(m.Value)
		}
	}
}
