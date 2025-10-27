package internal

import (
	"context"
	"log"
	"net"
	"os"
	"time"

	"github.com/digineo/go-ping"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// GetNodeList returns a list of nodes in the cluster.
func GetNodeList() ([]string, error) {

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

// GetNodeName returns the name of the current node.
func GetNodeName() (string, error) {
	name := os.Getenv("NODE_NAME")

	var err error

	// This method is unreliable since it will return just the pod name.
	if name == "" {
		name, err = os.Hostname()
	}

	return name, err

}

// PingHost pings a host and returns the RTT.
func PingHost(destination string) (time.Duration, error) {
	bind := "0.0.0.0"

	var remoteAddr *net.IPAddr
	var pinger *ping.Pinger
	var rtt time.Duration
	var err error

	if _, err = net.ResolveIPAddr("ip4", destination); err != nil {
		return rtt, err
	}

	if pinger, err = ping.New(bind, ""); err != nil {
		return rtt, err
	}

	defer pinger.Close()

	timeout, _ := time.ParseDuration("30s")

	rtt, err = pinger.PingAttempts(remoteAddr, timeout, int(3))

	if err != nil {
		return rtt, err
	}

	log.Printf("ping %s (%s) rtt=%v\n", destination, remoteAddr, rtt)

	return rtt, nil

}
