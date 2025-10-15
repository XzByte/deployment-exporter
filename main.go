package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

var (
	// Deployment downtime duration in seconds
	deploymentDowntimeDuration = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "k8s_deployment_downtime_duration_seconds",
			Help: "Duration in seconds that a deployment was down (from not ready to ready)",
		},
		[]string{"namespace", "deployment"},
	)

	// Deployment restart count
	deploymentRestartCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "k8s_deployment_restart_total",
			Help: "Total number of deployment restarts",
		},
		[]string{"namespace", "deployment"},
	)

	// Deployment current status
	deploymentStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "k8s_deployment_status",
			Help: "Current deployment status (1=ready, 0=not ready)",
		},
		[]string{"namespace", "deployment"},
	)

	// Deployment heartbeat - updates every time status is checked
	deploymentHeartbeat = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "k8s_deployment_heartbeat_timestamp_seconds",
			Help: "Timestamp of last heartbeat check (Unix epoch)",
		},
		[]string{"namespace", "deployment"},
	)

	// Time to recovery in milliseconds
	deploymentRecoveryTimeMs = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "k8s_deployment_recovery_time_milliseconds",
			Help: "Time taken for deployment to recover from down state in milliseconds",
		},
		[]string{"namespace", "deployment"},
	)

	// Last downtime start timestamp
	deploymentDowntimeStart = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "k8s_deployment_downtime_start_timestamp_seconds",
			Help: "Unix timestamp when the deployment went down",
		},
		[]string{"namespace", "deployment"},
	)

	// Deployment condition status
	deploymentConditionStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "k8s_deployment_condition_status",
			Help: "Deployment condition status (1=true, 0=false, -1=unknown)",
		},
		[]string{"namespace", "deployment", "condition", "status"},
	)

	// Deployment replicas info
	deploymentReplicasDesired = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "k8s_deployment_replicas_desired",
			Help: "Number of desired replicas for deployment",
		},
		[]string{"namespace", "deployment"},
	)

	deploymentReplicasReady = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "k8s_deployment_replicas_ready",
			Help: "Number of ready replicas for deployment",
		},
		[]string{"namespace", "deployment"},
	)

	deploymentReplicasAvailable = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "k8s_deployment_replicas_available",
			Help: "Number of available replicas for deployment",
		},
		[]string{"namespace", "deployment"},
	)

	deploymentReplicasUnavailable = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "k8s_deployment_replicas_unavailable",
			Help: "Number of unavailable replicas for deployment",
		},
		[]string{"namespace", "deployment"},
	)

	deploymentReplicasUpdated = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "k8s_deployment_replicas_updated",
			Help: "Number of updated replicas for deployment",
		},
		[]string{"namespace", "deployment"},
	)

	// Deployment metadata
	deploymentCreationTime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "k8s_deployment_created_timestamp_seconds",
			Help: "Unix timestamp when the deployment was created",
		},
		[]string{"namespace", "deployment"},
	)

	deploymentGeneration = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "k8s_deployment_metadata_generation",
			Help: "Sequence number representing a specific generation of the desired state",
		},
		[]string{"namespace", "deployment"},
	)

	deploymentObservedGeneration = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "k8s_deployment_status_observed_generation",
			Help: "The generation observed by the deployment controller",
		},
		[]string{"namespace", "deployment"},
	)

	// Deployment availability ratio
	deploymentAvailabilityRatio = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "k8s_deployment_availability_ratio",
			Help: "Deployment availability ratio (ready/desired)",
		},
		[]string{"namespace", "deployment", "available", "desired"},
	)

	// Resource usage metrics
	deploymentCPUUsage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "k8s_deployment_cpu_usage_millicores",
			Help: "Total CPU usage in millicores for all pods in the deployment",
		},
		[]string{"namespace", "deployment"},
	)

	deploymentMemoryUsage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "k8s_deployment_memory_usage_mebibytes",
			Help: "Total memory usage in MiB for all pods in the deployment",
		},
		[]string{"namespace", "deployment"},
	)

	deploymentCPURequest = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "k8s_deployment_cpu_request_millicores",
			Help: "Total CPU requests in millicores for all pods in the deployment",
		},
		[]string{"namespace", "deployment"},
	)

	deploymentMemoryRequest = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "k8s_deployment_memory_request_mebibytes",
			Help: "Total memory requests in MiB for all pods in the deployment",
		},
		[]string{"namespace", "deployment"},
	)

	deploymentCPULimit = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "k8s_deployment_cpu_limit_millicores",
			Help: "Total CPU limits in millicores for all pods in the deployment",
		},
		[]string{"namespace", "deployment"},
	)

	deploymentMemoryLimit = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "k8s_deployment_memory_limit_mebibytes",
			Help: "Total memory limits in MiB for all pods in the deployment",
		},
		[]string{"namespace", "deployment"},
	)

	// Resource usage percentage
	deploymentCPUUsagePercent = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "k8s_deployment_cpu_usage_percent",
			Help: "CPU usage as percentage of request",
		},
		[]string{"namespace", "deployment"},
	)

	deploymentMemoryUsagePercent = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "k8s_deployment_memory_usage_percent",
			Help: "Memory usage as percentage of request",
		},
		[]string{"namespace", "deployment"},
	)
)

type DeploymentTracker struct {
	clientset      *kubernetes.Clientset
	metricsClient  *metricsv.Clientset
	downtimeStart  map[string]time.Time
	namespace      string
}

func init() {
	// Register metrics with Prometheus
	prometheus.MustRegister(deploymentDowntimeDuration)
	prometheus.MustRegister(deploymentRestartCount)
	prometheus.MustRegister(deploymentStatus)
	prometheus.MustRegister(deploymentHeartbeat)
	prometheus.MustRegister(deploymentRecoveryTimeMs)
	prometheus.MustRegister(deploymentDowntimeStart)
	prometheus.MustRegister(deploymentConditionStatus)
	prometheus.MustRegister(deploymentReplicasDesired)
	prometheus.MustRegister(deploymentReplicasReady)
	prometheus.MustRegister(deploymentReplicasAvailable)
	prometheus.MustRegister(deploymentReplicasUnavailable)
	prometheus.MustRegister(deploymentReplicasUpdated)
	prometheus.MustRegister(deploymentCreationTime)
	prometheus.MustRegister(deploymentGeneration)
	prometheus.MustRegister(deploymentObservedGeneration)
	prometheus.MustRegister(deploymentAvailabilityRatio)
	prometheus.MustRegister(deploymentCPUUsage)
	prometheus.MustRegister(deploymentMemoryUsage)
	prometheus.MustRegister(deploymentCPURequest)
	prometheus.MustRegister(deploymentMemoryRequest)
	prometheus.MustRegister(deploymentCPULimit)
	prometheus.MustRegister(deploymentMemoryLimit)
	prometheus.MustRegister(deploymentCPUUsagePercent)
	prometheus.MustRegister(deploymentMemoryUsagePercent)
}

func main() {
	var (
		kubeconfig     string
		namespace      string
		metricsAddr    string
		scrapeInterval int
	)

	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to kubeconfig file (optional, uses in-cluster config if not set)")
	flag.StringVar(&namespace, "namespace", "", "Namespace to monitor (empty = all namespaces)")
	flag.StringVar(&metricsAddr, "metrics-addr", ":9101", "Address to expose metrics on")
	flag.IntVar(&scrapeInterval, "scrape-interval", 15, "Scrape interval in seconds")
	flag.Parse()

	// Create Kubernetes client
	config, err := getKubeConfig(kubeconfig)
	if err != nil {
		log.Fatalf("Error creating kubernetes config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating kubernetes client: %v", err)
	}

	// Create metrics client
	metricsClient, err := metricsv.NewForConfig(config)
	if err != nil {
		log.Printf("Warning: Could not create metrics client: %v (resource metrics will not be available)", err)
	}

	tracker := &DeploymentTracker{
		clientset:     clientset,
		metricsClient: metricsClient,
		downtimeStart: make(map[string]time.Time),
		namespace:     namespace,
	}

	// Start watching deployments
	go tracker.watchDeployments()

	// Start periodic scraper for heartbeat
	go tracker.periodicScrape(time.Duration(scrapeInterval) * time.Second)

	// Expose metrics endpoint
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	log.Printf("Starting K8s Deployment Exporter on %s", metricsAddr)
	log.Printf("Monitoring namespace: %s (empty = all)", namespace)
	log.Fatal(http.ListenAndServe(metricsAddr, nil))
}

func getKubeConfig(kubeconfig string) (*rest.Config, error) {
	// Try in-cluster config first
	if kubeconfig == "" {
		config, err := rest.InClusterConfig()
		if err == nil {
			return config, nil
		}
		log.Printf("In-cluster config failed, trying kubeconfig file")
	}

	// Fall back to kubeconfig file
	if kubeconfig == "" {
		kubeconfig = os.Getenv("KUBECONFIG")
	}
	if kubeconfig == "" {
		homeDir, _ := os.UserHomeDir()
		kubeconfig = homeDir + "/.kube/config"
	}

	return clientcmd.BuildConfigFromFlags("", kubeconfig)
}

func (t *DeploymentTracker) watchDeployments() {
	for {
		watcher, err := t.clientset.AppsV1().Deployments(t.namespace).Watch(context.Background(), metav1.ListOptions{})
		if err != nil {
			log.Printf("Error creating watcher: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		log.Println("Started watching deployments...")

		for event := range watcher.ResultChan() {
			if event.Type == watch.Error {
				log.Printf("Watch error: %v", event.Object)
				break
			}

			deployment, ok := event.Object.(*appsv1.Deployment)
			if !ok {
				continue
			}

			t.processDeployment(deployment)
		}

		watcher.Stop()
		log.Println("Watcher stopped, restarting...")
		time.Sleep(5 * time.Second)
	}
}

func (t *DeploymentTracker) periodicScrape(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		deployments, err := t.clientset.AppsV1().Deployments(t.namespace).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			log.Printf("Error listing deployments: %v", err)
			continue
		}

		for _, deployment := range deployments.Items {
			t.processDeployment(&deployment)
		}
	}
}

func (t *DeploymentTracker) processDeployment(deployment *appsv1.Deployment) {
	ns := deployment.Namespace
	name := deployment.Name
	key := ns + "/" + name

	// Update heartbeat
	now := time.Now()
	deploymentHeartbeat.WithLabelValues(ns, name).Set(float64(now.Unix()))

	// Set metadata metrics
	deploymentCreationTime.WithLabelValues(ns, name).Set(float64(deployment.CreationTimestamp.Unix()))
	deploymentGeneration.WithLabelValues(ns, name).Set(float64(deployment.Generation))
	deploymentObservedGeneration.WithLabelValues(ns, name).Set(float64(deployment.Status.ObservedGeneration))

	// Set replica metrics
	if deployment.Spec.Replicas != nil {
		deploymentReplicasDesired.WithLabelValues(ns, name).Set(float64(*deployment.Spec.Replicas))
	}
	deploymentReplicasReady.WithLabelValues(ns, name).Set(float64(deployment.Status.ReadyReplicas))
	deploymentReplicasAvailable.WithLabelValues(ns, name).Set(float64(deployment.Status.AvailableReplicas))
	deploymentReplicasUnavailable.WithLabelValues(ns, name).Set(float64(deployment.Status.UnavailableReplicas))
	deploymentReplicasUpdated.WithLabelValues(ns, name).Set(float64(deployment.Status.UpdatedReplicas))

	// Set availability ratio with labels showing "X/Y" format
	if deployment.Spec.Replicas != nil {
		available := fmt.Sprintf("%d", deployment.Status.ReadyReplicas)
		desired := fmt.Sprintf("%d", *deployment.Spec.Replicas)
		ratio := float64(0)
		if *deployment.Spec.Replicas > 0 {
			ratio = float64(deployment.Status.ReadyReplicas) / float64(*deployment.Spec.Replicas)
		}
		deploymentAvailabilityRatio.WithLabelValues(ns, name, available, desired).Set(ratio)
	}

	// Collect resource usage metrics
	t.collectResourceMetrics(ns, name, deployment)

	// Process deployment conditions (Available, Progressing, ReplicaFailure)
	for _, condition := range deployment.Status.Conditions {
		conditionType := string(condition.Type)
		conditionStatus := string(condition.Status)
		
		var statusValue float64
		switch conditionStatus {
		case "True":
			statusValue = 1
		case "False":
			statusValue = 0
		default: // "Unknown"
			statusValue = -1
		}
		
		deploymentConditionStatus.WithLabelValues(ns, name, conditionType, conditionStatus).Set(statusValue)
	}

	// Check if deployment is ready
	desiredReplicas := int32(0)
	if deployment.Spec.Replicas != nil {
		desiredReplicas = *deployment.Spec.Replicas
	}
	isReady := deployment.Status.ReadyReplicas == desiredReplicas &&
		desiredReplicas > 0 &&
		deployment.Status.UnavailableReplicas == 0

	// Track status
	if isReady {
		deploymentStatus.WithLabelValues(ns, name).Set(1)

		// If we have a downtime start time, calculate recovery
		if startTime, exists := t.downtimeStart[key]; exists {
			downtime := now.Sub(startTime)
			downtimeSeconds := downtime.Seconds()
			downtimeMs := float64(downtime.Milliseconds())

			// Display time in WIB (UTC+7)
			wibTime := now.UTC().Add(7 * time.Hour).Format("2006/01/02 15:04:05")
			log.Printf("[%s WIB] Deployment %s/%s recovered after %.2fs (%.0fms)", wibTime, ns, name, downtimeSeconds, downtimeMs)

			deploymentDowntimeDuration.WithLabelValues(ns, name).Set(downtimeSeconds)
			deploymentRecoveryTimeMs.WithLabelValues(ns, name).Set(downtimeMs)
			deploymentRestartCount.WithLabelValues(ns, name).Inc()

			delete(t.downtimeStart, key)
		}
	} else {
		deploymentStatus.WithLabelValues(ns, name).Set(0)

		// If this is a new downtime, record start time
		if _, exists := t.downtimeStart[key]; !exists {
			t.downtimeStart[key] = now
			deploymentDowntimeStart.WithLabelValues(ns, name).Set(float64(now.Unix()))
			// Display time in WIB (UTC+7)
			wibTime := now.UTC().Add(7 * time.Hour).Format("2006/01/02 15:04:05")
			log.Printf("[%s WIB] Deployment %s/%s went down", wibTime, ns, name)
		}
	}
}

func (t *DeploymentTracker) collectResourceMetrics(namespace, deploymentName string, deployment *appsv1.Deployment) {
	// Get pods for this deployment
	labelSelector := metav1.FormatLabelSelector(deployment.Spec.Selector)
	pods, err := t.clientset.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		log.Printf("Error listing pods for deployment %s/%s: %v", namespace, deploymentName, err)
		return
	}

	// Calculate resource requests and limits
	var totalCPURequest, totalMemoryRequest resource.Quantity
	var totalCPULimit, totalMemoryLimit resource.Quantity

	for _, pod := range pods.Items {
		for _, container := range pod.Spec.Containers {
			if cpuReq := container.Resources.Requests[corev1.ResourceCPU]; !cpuReq.IsZero() {
				totalCPURequest.Add(cpuReq)
			}
			if memReq := container.Resources.Requests[corev1.ResourceMemory]; !memReq.IsZero() {
				totalMemoryRequest.Add(memReq)
			}
			if cpuLim := container.Resources.Limits[corev1.ResourceCPU]; !cpuLim.IsZero() {
				totalCPULimit.Add(cpuLim)
			}
			if memLim := container.Resources.Limits[corev1.ResourceMemory]; !memLim.IsZero() {
				totalMemoryLimit.Add(memLim)
			}
		}
	}

	// Set request and limit metrics (in millicores and MiB)
	deploymentCPURequest.WithLabelValues(namespace, deploymentName).Set(float64(totalCPURequest.MilliValue()))
	deploymentMemoryRequest.WithLabelValues(namespace, deploymentName).Set(float64(totalMemoryRequest.Value()) / 1024 / 1024)
	deploymentCPULimit.WithLabelValues(namespace, deploymentName).Set(float64(totalCPULimit.MilliValue()))
	deploymentMemoryLimit.WithLabelValues(namespace, deploymentName).Set(float64(totalMemoryLimit.Value()) / 1024 / 1024)

	// Try to get actual usage from metrics server
	if t.metricsClient != nil {
		podMetrics, err := t.metricsClient.MetricsV1beta1().PodMetricses(namespace).List(context.Background(), metav1.ListOptions{
			LabelSelector: labelSelector,
		})
		if err != nil {
			// Metrics server might not be available
			return
		}

		var totalCPUUsage, totalMemoryUsage int64
		for _, pm := range podMetrics.Items {
			for _, container := range pm.Containers {
				cpuUsage := container.Usage[corev1.ResourceCPU]
				memUsage := container.Usage[corev1.ResourceMemory]
				totalCPUUsage += cpuUsage.MilliValue()
				totalMemoryUsage += memUsage.Value()
			}
		}

		// Set usage metrics (millicores and MiB)
		deploymentCPUUsage.WithLabelValues(namespace, deploymentName).Set(float64(totalCPUUsage))
		deploymentMemoryUsage.WithLabelValues(namespace, deploymentName).Set(float64(totalMemoryUsage) / 1024 / 1024)

		// Calculate usage percentages
		if totalCPURequest.MilliValue() > 0 {
			cpuPercent := (float64(totalCPUUsage) / float64(totalCPURequest.MilliValue())) * 100
			deploymentCPUUsagePercent.WithLabelValues(namespace, deploymentName).Set(cpuPercent)
		}
		if totalMemoryRequest.Value() > 0 {
			memPercent := (float64(totalMemoryUsage) / float64(totalMemoryRequest.Value())) * 100
			deploymentMemoryUsagePercent.WithLabelValues(namespace, deploymentName).Set(memPercent)
		}
	}
}
