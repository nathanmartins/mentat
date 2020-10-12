package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/digineo/go-ping"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"
)

type NodeList struct {
	Items []struct {
		Metadata struct {
			Name string `json:"name"`
			UID  string `json:"uid"`
		} `json:"metadata"`
	} `json:"items"`
}

func getNodeList() (NodeList, error) {

	var nodes NodeList

	// Disabling https checks
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	addr := "http://localhost:8001/api/v1/nodes?limit=500"

	req, err := http.NewRequest("GET", addr, nil)
	if err != nil {
		return nodes, err
	}

	data, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
	if err != nil {
		return nodes, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", string(data)))

	resp, err := client.Do(req)
	if err != nil {
		return nodes, err
	}

	// this might be over-written by other errors later on but it doesn't matter.
	defer func() {
		err = resp.Body.Close()
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nodes, err
	}

	err = json.Unmarshal(body, &nodes)
	if err != nil {
		return nodes, err
	}

	return nodes, err
}

func pingHost(destination string) (time.Duration, error) {
	bind := "0.0.0.0"

	var remoteAddr *net.IPAddr
	var pinger *ping.Pinger
	var rtt time.Duration

	if r, err := net.ResolveIPAddr("ip4", destination); err != nil {
		return rtt, err
	} else {
		remoteAddr = r
	}

	if p, err := ping.New(bind, ""); err != nil {
		return rtt, err
	} else {
		pinger = p
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
