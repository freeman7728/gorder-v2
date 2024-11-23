package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

var StripeCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Help: "Count",
	}, []string{"stripe_time"})

var Reg = prometheus.NewRegistry()

func init() {
	prometheus.WrapRegistererWith(prometheus.Labels{"serviceName": "statistic_stripe_time"}, reg).MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		StripeCounter,
	)
}
