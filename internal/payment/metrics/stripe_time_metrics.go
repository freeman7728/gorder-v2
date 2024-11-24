package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

var StripeCounter = prometheus.NewSummaryVec(
	prometheus.SummaryOpts{
		Name: "stripe_request_duration_seconds",
		Help: "Count",
		Objectives: map[float64]float64{
			0.5:  0.05,
			0.9:  0.01,
			0.99: 0.001,
		},
	}, []string{"stripe_time"})

var Reg = prometheus.NewRegistry()

func init() {
	prometheus.WrapRegistererWith(prometheus.Labels{"serviceName": "statistic_stripe_time"}, Reg).MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		StripeCounter,
	)
}
