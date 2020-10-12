package main

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {

	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalf("failed getting hostname: %s", err)
	}

	// Prometheus: Histogram to collect required metrics
	histogram := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "node_latency",
		Help:    "Time take ping other nodes",
		Buckets: []float64{1, 2, 5, 6, 10}, //defining small buckets as this app should not take more than 1 sec to respond
	}, []string{"origin_node", "destination_node"}) // this will be partitioned by nodes

	err = prometheus.Register(histogram)
	if err != nil {
		log.Fatalf("failed registering historgram: %s", err)
	}

	go func() {

		for {

			hosts, err := getNodeList()

			if err != nil {
				log.Fatalf("failed getting node list: %s", err)
			}

			if len(hosts) == 0 {
				log.Fatal("getNodes returned 0 nodes")
			}

			for _, host := range hosts {

				rtt, err := pingHost(host)
				if err != nil {
					log.Printf("failed pinging node '%s' : %s", host, err)
				} else {
					fmt.Printf("Time: %v\n", rtt.Seconds())
					histogram.WithLabelValues(hostname, host).Observe(rtt.Seconds())
				}

			}

			time.Sleep(10 * time.Second)
		}

	}()

	http.Handle("/metrics", promhttp.Handler())
	_ = http.ListenAndServe(":2112", nil)
}

// sum(node_latency_sum)/sum(node_latency_count)
// histogram_quantile(0.99, sum(rate(node_latency_bucket{destination_node="facebook.com", origin_node="odin"}[5m])) by (le))
