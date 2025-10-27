# mentat — Inter‑node latency exporter for Kubernetes

mentat discovers Kubernetes nodes via the in‑cluster API, pings them using ICMP, and exposes round‑trip time (RTT) metrics on a Prometheus endpoint. Use it to observe network latency between nodes in your cluster.

## Features
- Discovers nodes using the Kubernetes API
- Prometheus metrics on `/metrics` (port `2112`)
- Lightweight container image, non‑root by default

## Metrics
The exporter records a Prometheus histogram named `node_latency` with the following labels:
- `origin_node`: the node where this pod is running
- `destination_node`: the node being pinged

Example queries:
- Average latency across all nodes:
  ```promql
  sum(node_latency_sum) / sum(node_latency_count)
  ```
- 99th percentile to a given destination:
  ```promql
  histogram_quantile(0.99, sum(rate(node_latency_bucket{destination_node="<node>"}[5m])) by (le))
  ```

## Build and Run Locally
Prerequisites: Go 1.22+

- Build:
  ```bash
  go build -o mentat .
  ```
- Run:
  ```bash
  ./mentat
  # metrics at http://localhost:2112/metrics
  ```

Note: local run requires access to a Kubernetes cluster environment and ICMP permissions, so it’s primarily intended for containerized/in‑cluster operation.

## Container Image
A multi‑stage `Dockerfile` is provided. It produces a small Alpine‑based image and runs as a non‑root user by default.

Build locally:
```bash
docker build -t ghcr.io/nathanmartins/mentat:local .
```

Run locally (requires raw socket capability for ICMP):
```bash
docker run --rm -p 2112:2112 \
  --cap-add=NET_RAW \
  ghcr.io/nathanmartins/mentat:local
```

Alternative: file capabilities
- The image includes `libcap-utils`. If you prefer to avoid `--cap-add=NET_RAW`, you can bake the capability into the binary by uncommenting the `setcap` line in the `Dockerfile`, then rebuild. Note: file capabilities may not survive certain storage drivers or copy operations.

## Kubernetes Deployment
mentat must run with permissions to list cluster nodes and must be able to send ICMP (NET_RAW). Below is a minimal setup using a dedicated `ServiceAccount`, `ClusterRole`, and `ClusterRoleBinding`, plus a `Deployment` and `Service`:

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: monitoring
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: mentat
  namespace: monitoring
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: mentat
rules:
  - apiGroups: [""]
    resources: ["nodes"]
    verbs: ["get", "list", "watch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: mentat
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: mentat
subjects:
  - kind: ServiceAccount
    name: mentat
    namespace: monitoring
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mentat
  namespace: monitoring
spec:
  replicas: 1
  selector:
    matchLabels: { app: mentat }
  template:
    metadata:
      labels: { app: mentat }
    spec:
      serviceAccountName: mentat
      containers:
        - name: mentat
          image: ghcr.io/nathanmartins/mentat:latest
          imagePullPolicy: IfNotPresent
          ports:
            - containerPort: 2112
              name: metrics
          env:
            - name: NODE_NAME
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
          securityContext:
            allowPrivilegeEscalation: false
            readOnlyRootFilesystem: true
            capabilities:
              drop: ["ALL"]
              add: ["NET_RAW"]
          resources:
            requests:
              cpu: 20m
              memory: 32Mi
            limits:
              cpu: 200m
              memory: 128Mi
---
apiVersion: v1
kind: Service
metadata:
  name: mentat
  namespace: monitoring
  labels:
    app: mentat
spec:
  selector:
    app: mentat
  ports:
    - name: http-metrics
      port: 2112
      targetPort: metrics
```

If you use Prometheus Operator, you can discover the endpoint with a `ServiceMonitor`:
```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: mentat
  namespace: monitoring
spec:
  selector:
    matchLabels:
      app: mentat
  namespaceSelector:
    matchNames: ["monitoring"]
  endpoints:
    - port: http-metrics
      interval: 15s
```

## Configuration
- `NODE_NAME`: Optional. When set, it labels metrics with the pod’s node. Defaults to the pod hostname if not provided; using the Downward API as shown above is recommended for accuracy.

## Security Notes
- The app needs to create raw sockets for ICMP. Grant either container capability `NET_RAW` (recommended) or use file capabilities via `setcap` during the image build.
- Runs as a non‑root user by default in the provided image.

## Troubleshooting
- No metrics scraped: verify the `Service` port and Prometheus/ServiceMonitor configuration.
- Permission denied on ping: confirm `NET_RAW` capability is present, or use file capabilities.
- Cannot list nodes: ensure the RBAC setup (ClusterRole/Binding) is applied and the pod uses the correct `ServiceAccount`.

## License
Apache License 2.0. See `LICENSE`.

