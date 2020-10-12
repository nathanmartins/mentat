# mentat - inter-node latency exporter
---

This is a go application that uses the Kubernetes API  to get a list of all the nodes in the current kubernetes cluster, then pings them and exports it as prometheus endpoint.

