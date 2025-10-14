# Alternative Approach: Fork Node Exporter

If you want to fork the official node_exporter instead of creating a standalone exporter, here's how to do it:

## Why Fork Node Exporter?

**Pros:**
- Leverage existing infrastructure and patterns
- Get all node-level metrics plus custom ones
- Benefit from ongoing node_exporter maintenance

**Cons:**
- Much larger binary and resource footprint
- More complex codebase to maintain
- Requires understanding node_exporter architecture
- Not ideal for Kubernetes-specific metrics

## Steps to Fork and Extend Node Exporter

### 1. Fork the Repository

```bash
# Clone node_exporter
git clone https://github.com/prometheus/node_exporter.git
cd node_exporter

# Create your fork on GitHub and add as remote
git remote add myfork https://github.com/YOUR_USERNAME/node_exporter.git
```

### 2. Create Custom Collector

Create a new file `collector/k8s_deployment.go`:

```go
// +build !nok8sdeployment

package collector

import (
    "context"
    "fmt"
    "sync"
    "time"

    "github.com/go-kit/log"
    "github.com/go-kit/log/level"
    "github.com/prometheus/client_golang/prometheus"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/rest"
)

const (
    k8sDeploymentSubsystem = "k8s_deployment"
)

type k8sDeploymentCollector struct {
    clientset     *kubernetes.Clientset
    downtimeStart map[string]time.Time
    mu            sync.RWMutex
    logger        log.Logger

    status          *prometheus.Desc
    downtimeDuration *prometheus.Desc
    recoveryTime     *prometheus.Desc
    restartTotal     *prometheus.Desc
    heartbeat        *prometheus.Desc
}

func init() {
    registerCollector("k8sdeployment", defaultEnabled, NewK8sDeploymentCollector)
}

// NewK8sDeploymentCollector returns a new Collector exposing K8s deployment stats
func NewK8sDeploymentCollector(logger log.Logger) (Collector, error) {
    config, err := rest.InClusterConfig()
    if err != nil {
        return nil, fmt.Errorf("failed to create in-cluster config: %w", err)
    }

    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
    }

    return &k8sDeploymentCollector{
        clientset:     clientset,
        downtimeStart: make(map[string]time.Time),
        logger:        logger,
        status: prometheus.NewDesc(
            prometheus.BuildFQName(namespace, k8sDeploymentSubsystem, "status"),
            "Current deployment status (1=ready, 0=not ready)",
            []string{"namespace", "deployment"}, nil,
        ),
        downtimeDuration: prometheus.NewDesc(
            prometheus.BuildFQName(namespace, k8sDeploymentSubsystem, "downtime_duration_seconds"),
            "Duration in seconds that a deployment was down",
            []string{"namespace", "deployment"}, nil,
        ),
        recoveryTime: prometheus.NewDesc(
            prometheus.BuildFQName(namespace, k8sDeploymentSubsystem, "recovery_time_milliseconds"),
            "Time taken to recover from down state in milliseconds",
            []string{"namespace", "deployment"}, nil,
        ),
        restartTotal: prometheus.NewDesc(
            prometheus.BuildFQName(namespace, k8sDeploymentSubsystem, "restart_total"),
            "Total number of deployment restarts",
            []string{"namespace", "deployment"}, nil,
        ),
        heartbeat: prometheus.NewDesc(
            prometheus.BuildFQName(namespace, k8sDeploymentSubsystem, "heartbeat_timestamp_seconds"),
            "Timestamp of last heartbeat check",
            []string{"namespace", "deployment"}, nil,
        ),
    }, nil
}

func (c *k8sDeploymentCollector) Update(ch chan<- prometheus.Metric) error {
    deployments, err := c.clientset.AppsV1().Deployments("").List(context.Background(), metav1.ListOptions{})
    if err != nil {
        level.Error(c.logger).Log("msg", "Failed to list deployments", "err", err)
        return err
    }

    now := time.Now()

    c.mu.Lock()
    defer c.mu.Unlock()

    for _, deployment := range deployments.Items {
        ns := deployment.Namespace
        name := deployment.Name
        key := ns + "/" + name

        // Heartbeat
        ch <- prometheus.MustNewConstMetric(
            c.heartbeat,
            prometheus.GaugeValue,
            float64(now.Unix()),
            ns, name,
        )

        // Check readiness
        isReady := deployment.Status.ReadyReplicas == deployment.Status.Replicas &&
            deployment.Status.Replicas > 0 &&
            deployment.Status.UnavailableReplicas == 0

        if isReady {
            ch <- prometheus.MustNewConstMetric(c.status, prometheus.GaugeValue, 1, ns, name)

            if startTime, exists := c.downtimeStart[key]; exists {
                downtime := now.Sub(startTime)
                ch <- prometheus.MustNewConstMetric(
                    c.downtimeDuration,
                    prometheus.GaugeValue,
                    downtime.Seconds(),
                    ns, name,
                )
                ch <- prometheus.MustNewConstMetric(
                    c.recoveryTime,
                    prometheus.GaugeValue,
                    float64(downtime.Milliseconds()),
                    ns, name,
                )
                delete(c.downtimeStart, key)
            }
        } else {
            ch <- prometheus.MustNewConstMetric(c.status, prometheus.GaugeValue, 0, ns, name)

            if _, exists := c.downtimeStart[key]; !exists {
                c.downtimeStart[key] = now
            }
        }
    }

    return nil
}
```

### 3. Build Configuration

Update `collector/collector.go` to include your collector in the list.

### 4. Build Custom Node Exporter

```bash
# Build with your custom collector
make build

# Build Docker image
docker build -t custom-node-exporter:latest .
```

### 5. Deploy to Kubernetes

Use a DaemonSet similar to standard node_exporter but with your custom image:

```yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: custom-node-exporter
  namespace: monitoring
spec:
  selector:
    matchLabels:
      app: custom-node-exporter
  template:
    metadata:
      labels:
        app: custom-node-exporter
    spec:
      serviceAccountName: node-exporter
      containers:
      - name: node-exporter
        image: custom-node-exporter:latest
        args:
          - --path.procfs=/host/proc
          - --path.sysfs=/host/sys
          - --path.rootfs=/host/root
          - --collector.filesystem.mount-points-exclude=^/(sys|proc|dev|host|etc)($$|/)
          - --collector.k8sdeployment  # Enable your custom collector
        ports:
        - name: metrics
          containerPort: 9100
        volumeMounts:
        - name: proc
          mountPath: /host/proc
          readOnly: true
        - name: sys
          mountPath: /host/sys
          readOnly: true
        - name: root
          mountPath: /host/root
          readOnly: true
      volumes:
      - name: proc
        hostPath:
          path: /proc
      - name: sys
        hostPath:
          path: /sys
      - name: root
        hostPath:
          path: /
```

## Recommendation

For your use case (Kubernetes deployment monitoring with millisecond precision), I **strongly recommend the standalone exporter** approach I created above because:

1. **Resource Friendly**: ~50-64Mi RAM vs ~100-150Mi for node_exporter
2. **Focused**: Only deployment metrics, no unnecessary node metrics
3. **Simpler**: Easier to maintain and customize
4. **Better Architecture**: Node exporter runs per-node; deployment monitoring should be cluster-wide
5. **Millisecond Precision**: Custom built for your exact requirements

The standalone exporter is already production-ready with all the features you need!
