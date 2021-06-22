package main

import (
	"context"
	"github.com/digineo/go-ping"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"log"
	"net"
	"os"
	"time"
)

func getNodeList() ([]string, error) {

	var simpleList []string

	// Get configuration from within the cluster itself
	config, err := rest.InClusterConfig()
	if err != nil {
		return simpleList, err
	}

	c, err := kubernetes.NewForConfig(config)
	if err != nil {
		return simpleList, err
	}

	nodes, err := c.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{
		Limit: 500,
	})

	if err != nil {
		return simpleList, err
	}

	for _, items := range nodes.Items {
		simpleList = append(simpleList, items.Name)
	}

	return simpleList, nil
}

func getNodeName() (string, error) {
	name := os.Getenv("NODE_NAME")

	var err error

	// This method is unreliable since it will return just the pod name.
	if name == "" {
		name, err = os.Hostname()
	}

	return name, err

}

func pingHost(destination string) (time.Duration, error) {
	bind := "0.0.0.0"

	var remoteAddr *net.IPAddr
	var pinger *ping.Pinger
	var rtt time.Duration

	if remoteAddr, err := net.ResolveIPAddr("ip4", destination); err != nil {
		return rtt, err
	}

	if pinger, err := ping.New(bind, ""); err != nil {
		return rtt, err
	}

	defer pinger.Close()

	timeout, _ := time.ParseDuration("30s")

	rtt, err := pinger.PingAttempts(remoteAddr, timeout, int(3))

	if err != nil {
		return rtt, err
	}

	log.Printf("ping %s (%s) rtt=%v\n", destination, remoteAddr, rtt)

	return rtt, nil

}
