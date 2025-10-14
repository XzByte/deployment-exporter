package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
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
)

type DeploymentTracker struct {
	clientset     *kubernetes.Clientset
	downtimeStart map[string]time.Time
	namespace     string
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

	tracker := &DeploymentTracker{
		clientset:     clientset,
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
