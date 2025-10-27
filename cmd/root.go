package cmd

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/nathanmartins/mentat/internal"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use: "mentat",
	Run: func(cmd *cobra.Command, args []string) {
		hostname, err := internal.GetNodeName()
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

				hosts, err := internal.GetNodeList()

				if err != nil {
					log.Fatalf("failed getting node list: %s", err)
				}

				if len(hosts) == 0 {
					log.Fatal("getNodes returned 0 nodes")
				}

				for _, host := range hosts {

					rtt, err := internal.PingHost(host)
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

	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.mentat.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
