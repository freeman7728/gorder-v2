package main

import (
	"bytes"
	"encoding/json"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"math/rand"
	"net/http"
	"time"
)

const (
	testAddr = "192.168.192.131:9123"
)

var httpStatusCodeCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_status_code_counter",
		Help: "Count http status code",
	}, []string{"status_code"})

func main() {
	go produceData()
	reg := prometheus.NewRegistry()
	prometheus.WrapRegistererWith(prometheus.Labels{"serviceName": "demo-service"}, reg).MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		httpStatusCodeCounter,
	)

	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	http.HandleFunc("/", sendMetricsHandler)
	_ = http.ListenAndServe(testAddr, nil)
}

func sendMetricsHandler(w http.ResponseWriter, r *http.Request) {
	var req request
	defer func() {
		httpStatusCodeCounter.WithLabelValues(req.StatusCode).Inc()
		log.Printf("%s += 1", req.StatusCode)
	}()
	_ = json.NewDecoder(r.Body).Decode(&req)
	log.Printf("receive req from %s,req: %v", testAddr, req)
	_, _ = w.Write([]byte(req.StatusCode))
}

type request struct {
	StatusCode string
}

func produceData() {
	codes := []string{"503", "403", "400", "200", "304"}
	for {
		body, _ := json.Marshal(request{
			StatusCode: codes[rand.Intn(len(codes))],
		})
		requestBody := bytes.NewBuffer(body)
		http.Post("http://"+testAddr, "application/json", requestBody)
		log.Printf("send statusCode=%v,to %s", requestBody, testAddr)
		time.Sleep(300 * time.Millisecond)
	}
}
