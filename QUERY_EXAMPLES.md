# Prometheus Query Examples for K8s Deployment Exporter

This document provides comprehensive Prometheus query examples for monitoring Kubernetes deployments using the K8s Deployment Exporter.

## Table of Contents
- [Basic Queries](#basic-queries)
- [Deployment Status](#deployment-status)
- [Downtime & Recovery](#downtime--recovery)
- [Deployment Conditions](#deployment-conditions)
- [Replica Monitoring](#replica-monitoring)
- [Advanced Queries](#advanced-queries)
- [Alerting Rules](#alerting-rules)

---

## Basic Queries

### Current Deployment Status
```promql
# All deployments and their current status
k8s_deployment_status

# Only running deployments
k8s_deployment_status == 1

# Only down deployments
k8s_deployment_status == 0

# Status of specific deployment
k8s_deployment_status{namespace="production",deployment="my-app"}
```

### Heartbeat Monitoring
```promql
# Last heartbeat timestamp
k8s_deployment_heartbeat_timestamp_seconds

# Time since last heartbeat (in seconds)
time() - k8s_deployment_heartbeat_timestamp_seconds

# Deployments with stale heartbeat (> 2 minutes)
(time() - k8s_deployment_heartbeat_timestamp_seconds) > 120
```

---

## Deployment Status

### Count Deployments by Status
```promql
# Total number of running deployments
count(k8s_deployment_status == 1)

# Total number of down deployments
count(k8s_deployment_status == 0)

# Count by namespace
count by (namespace) (k8s_deployment_status)

# Percentage of healthy deployments
(count(k8s_deployment_status == 1) / count(k8s_deployment_status)) * 100
```

### Filter by Namespace
```promql
# All deployments in production namespace
k8s_deployment_status{namespace="production"}

# Down deployments in staging
k8s_deployment_status{namespace="staging"} == 0

# Multiple namespaces
k8s_deployment_status{namespace=~"production|staging"}
```

---

## Downtime & Recovery

### Recovery Time Metrics
```promql
# Last recovery time in milliseconds
k8s_deployment_recovery_time_milliseconds

# Recovery time in seconds
k8s_deployment_recovery_time_milliseconds / 1000

# Average recovery time (last hour)
avg_over_time(k8s_deployment_recovery_time_milliseconds[1h])

# Max recovery time (last hour)
max_over_time(k8s_deployment_recovery_time_milliseconds[1h])

# Deployments with slow recovery (> 30 seconds)
k8s_deployment_recovery_time_milliseconds > 30000
```

### Downtime Duration
```promql
# Current downtime duration in seconds
k8s_deployment_downtime_duration_seconds

# Deployments down for more than 5 minutes
k8s_deployment_downtime_duration_seconds > 300

# Average downtime in last 24 hours
avg_over_time(k8s_deployment_downtime_duration_seconds[24h])

# When deployment went down
k8s_deployment_downtime_start_timestamp_seconds
```

### Restart Tracking
```promql
# Total restart count
k8s_deployment_restart_total

# Restart rate (per minute)
rate(k8s_deployment_restart_total[5m]) * 60

# Restarts in last hour
increase(k8s_deployment_restart_total[1h])

# Restarts in last 24 hours
increase(k8s_deployment_restart_total[24h])

# Top 10 deployments by restart count
topk(10, k8s_deployment_restart_total)

# Deployments with frequent restarts (> 5 in last hour)
increase(k8s_deployment_restart_total[1h]) > 5
```

---

## Deployment Conditions

### Available Condition
```promql
# Deployments that are available
k8s_deployment_condition_status{condition="Available",status="True"} == 1

# Deployments that are NOT available
k8s_deployment_condition_status{condition="Available",status="False"} == 0

# Count unavailable deployments
count(k8s_deployment_condition_status{condition="Available",status="False"} == 0)
```

### Progressing Condition
```promql
# Deployments currently progressing (updating)
k8s_deployment_condition_status{condition="Progressing",status="True"} == 1

# Stuck deployments (not progressing when they should be)
k8s_deployment_condition_status{condition="Progressing",status="False"} == 0

# Count deployments being updated
count(k8s_deployment_condition_status{condition="Progressing",status="True"} == 1)
```

### ReplicaFailure Condition
```promql
# Deployments with replica failures
k8s_deployment_condition_status{condition="ReplicaFailure",status="True"} == 1

# List all deployments with failures
k8s_deployment_condition_status{condition="ReplicaFailure",status="True"}

# Count deployments with replica failures
count(k8s_deployment_condition_status{condition="ReplicaFailure",status="True"} == 1)
```

### Condition Status by Type
```promql
# All conditions for a specific deployment
k8s_deployment_condition_status{namespace="production",deployment="my-app"}

# Deployments with unknown condition status
k8s_deployment_condition_status == -1

# Count conditions by status
count by (condition, status) (k8s_deployment_condition_status)
```

---

## Replica Monitoring

### Replica Counts
```promql
# Desired replicas
k8s_deployment_replicas_desired

# Ready replicas
k8s_deployment_replicas_ready

# Available replicas
k8s_deployment_replicas_available

# Unavailable replicas
k8s_deployment_replicas_unavailable

# Updated replicas
k8s_deployment_replicas_updated
```

### Replica Comparison
```promql
# Deployments with replica mismatch (desired vs ready)
k8s_deployment_replicas_desired - k8s_deployment_replicas_ready != 0

# Number of missing replicas
k8s_deployment_replicas_desired - k8s_deployment_replicas_ready

# Deployments with unavailable replicas
k8s_deployment_replicas_unavailable > 0

# Replica readiness percentage
(k8s_deployment_replicas_ready / k8s_deployment_replicas_desired) * 100

# Deployments below 80% ready
(k8s_deployment_replicas_ready / k8s_deployment_replicas_desired) * 100 < 80
```

### Replica Scaling Issues
```promql
# Deployments not at desired replica count
k8s_deployment_replicas_ready != k8s_deployment_replicas_desired

# Deployments with 0 ready replicas but desired > 0
k8s_deployment_replicas_ready == 0 and k8s_deployment_replicas_desired > 0

# Total desired vs total ready (cluster-wide)
sum(k8s_deployment_replicas_desired) - sum(k8s_deployment_replicas_ready)
```

---

## Advanced Queries

### Deployment Health Score
```promql
# Calculate health score (0-100)
(
  (k8s_deployment_status * 40) +
  (k8s_deployment_condition_status{condition="Available",status="True"} * 30) +
  ((k8s_deployment_replicas_ready / k8s_deployment_replicas_desired) * 30)
)

# Deployments with health score below 80
(
  (k8s_deployment_status * 40) +
  (k8s_deployment_condition_status{condition="Available",status="True"} * 30) +
  ((k8s_deployment_replicas_ready / k8s_deployment_replicas_desired) * 30)
) < 80
```

### Generation Mismatch (Pending Updates)
```promql
# Deployments with pending configuration updates
k8s_deployment_metadata_generation != k8s_deployment_status_observed_generation

# Count deployments with pending updates
count(k8s_deployment_metadata_generation != k8s_deployment_status_observed_generation)
```

### Deployment Age
```promql
# Deployment age in seconds
time() - k8s_deployment_created_timestamp_seconds

# Deployment age in days
(time() - k8s_deployment_created_timestamp_seconds) / 86400

# Deployments older than 30 days
(time() - k8s_deployment_created_timestamp_seconds) / 86400 > 30

# Newest deployments (last 24 hours)
(time() - k8s_deployment_created_timestamp_seconds) < 86400
```

### Deployment Stability
```promql
# Deployments with no restarts in last 24 hours
increase(k8s_deployment_restart_total[24h]) == 0

# Stability score (inverse of restart rate)
1 / (rate(k8s_deployment_restart_total[1h]) * 3600 + 1)

# Most stable deployments
topk(10, 1 / (rate(k8s_deployment_restart_total[24h]) * 86400 + 1))
```

### Cross-Metric Correlation
```promql
# Deployments that are down AND have unavailable replicas
k8s_deployment_status == 0 and k8s_deployment_replicas_unavailable > 0

# Deployments progressing but not yet available
k8s_deployment_condition_status{condition="Progressing",status="True"} == 1
and
k8s_deployment_condition_status{condition="Available",status="False"} == 0

# Deployments with replica failures AND down
k8s_deployment_condition_status{condition="ReplicaFailure",status="True"} == 1
and
k8s_deployment_status == 0
```

---

## Alerting Rules

### Critical Alerts

```yaml
groups:
  - name: deployment-critical
    interval: 30s
    rules:
      - alert: DeploymentDown
        expr: k8s_deployment_status == 0
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "Deployment {{ $labels.namespace }}/{{ $labels.deployment }} is down"
          description: "Deployment has been down for more than 2 minutes"

      - alert: DeploymentAllReplicasDown
        expr: k8s_deployment_replicas_ready == 0 and k8s_deployment_replicas_desired > 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "All replicas down for {{ $labels.namespace }}/{{ $labels.deployment }}"
          description: "Zero ready replicas out of {{ $value }} desired"

      - alert: DeploymentReplicaFailure
        expr: k8s_deployment_condition_status{condition="ReplicaFailure",status="True"} == 1
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "Replica failure in {{ $labels.namespace }}/{{ $labels.deployment }}"
          description: "Deployment is experiencing replica failures"
```

### Warning Alerts

```yaml
      - alert: DeploymentSlowRecovery
        expr: k8s_deployment_recovery_time_milliseconds > 60000
        for: 1m
        labels:
          severity: warning
        annotations:
          summary: "Slow recovery for {{ $labels.namespace }}/{{ $labels.deployment }}"
          description: "Deployment took {{ $value }}ms to recover (> 60 seconds)"

      - alert: DeploymentFrequentRestarts
        expr: increase(k8s_deployment_restart_total[1h]) > 5
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Frequent restarts in {{ $labels.namespace }}/{{ $labels.deployment }}"
          description: "Deployment has restarted {{ $value }} times in the last hour"

      - alert: DeploymentReplicaMismatch
        expr: (k8s_deployment_replicas_desired - k8s_deployment_replicas_ready) > 0
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "Replica mismatch in {{ $labels.namespace }}/{{ $labels.deployment }}"
          description: "{{ $value }} replicas are not ready"

      - alert: DeploymentNotAvailable
        expr: k8s_deployment_condition_status{condition="Available",status="False"} == 0
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Deployment {{ $labels.namespace }}/{{ $labels.deployment }} not available"
          description: "Deployment has been unavailable for more than 5 minutes"

      - alert: DeploymentUpdateStuck
        expr: k8s_deployment_metadata_generation != k8s_deployment_status_observed_generation
        for: 15m
        labels:
          severity: warning
        annotations:
          summary: "Deployment update stuck for {{ $labels.namespace }}/{{ $labels.deployment }}"
          description: "Deployment configuration change not applied for 15 minutes"
```

### Info Alerts

```yaml
      - alert: DeploymentHeartbeatStale
        expr: (time() - k8s_deployment_heartbeat_timestamp_seconds) > 120
        for: 5m
        labels:
          severity: info
        annotations:
          summary: "Stale heartbeat for {{ $labels.namespace }}/{{ $labels.deployment }}"
          description: "Exporter heartbeat is {{ $value }}s old (> 120s)"

      - alert: DeploymentHighRestartRate
        expr: rate(k8s_deployment_restart_total[5m]) * 60 > 0.5
        for: 10m
        labels:
          severity: info
        annotations:
          summary: "High restart rate for {{ $labels.namespace }}/{{ $labels.deployment }}"
          description: "Deployment is restarting {{ $value }} times per minute"
```

---

## Grafana Dashboard Queries

### Single Stat Panels

```promql
# Total Deployments
count(k8s_deployment_status)

# Healthy Deployments
count(k8s_deployment_status == 1)

# Down Deployments
count(k8s_deployment_status == 0)

# Average Recovery Time
avg(k8s_deployment_recovery_time_milliseconds) / 1000

# Total Restarts (24h)
sum(increase(k8s_deployment_restart_total[24h]))
```

### Time Series Panels

```promql
# Deployment Status Over Time
k8s_deployment_status{namespace="production"}

# Recovery Time Trend
k8s_deployment_recovery_time_milliseconds

# Restart Rate
rate(k8s_deployment_restart_total[5m]) * 60

# Replica Readiness
k8s_deployment_replicas_ready / k8s_deployment_replicas_desired
```

### Table Panels

```promql
# Deployment Overview (use Format as: Table)
k8s_deployment_status
or
k8s_deployment_replicas_desired
or
k8s_deployment_replicas_ready
or
k8s_deployment_replicas_unavailable
or
k8s_deployment_condition_status{condition="Available"}
```

### Heatmap Panels

```promql
# Downtime Duration Distribution
k8s_deployment_downtime_duration_seconds

# Recovery Time Distribution
k8s_deployment_recovery_time_milliseconds / 1000
```

---

## Recording Rules

For frequently used queries, create recording rules to improve performance:

```yaml
groups:
  - name: deployment-recording-rules
    interval: 30s
    rules:
      - record: deployment:replicas:mismatch
        expr: k8s_deployment_replicas_desired - k8s_deployment_replicas_ready

      - record: deployment:health:percentage
        expr: (k8s_deployment_replicas_ready / k8s_deployment_replicas_desired) * 100

      - record: deployment:restart:rate_5m
        expr: rate(k8s_deployment_restart_total[5m]) * 60

      - record: deployment:downtime:total_24h
        expr: sum by (namespace, deployment) (increase(k8s_deployment_downtime_duration_seconds[24h]))

      - record: namespace:deployment:count
        expr: count by (namespace) (k8s_deployment_status)

      - record: namespace:deployment:down_count
        expr: count by (namespace) (k8s_deployment_status == 0)
```

---

## Tips & Best Practices

1. **Use rate() for counters**: Always use `rate()` or `increase()` with counter metrics like `k8s_deployment_restart_total`

2. **Avoid high cardinality**: Be careful with queries that generate many time series (use aggregations when possible)

3. **Use recording rules**: For complex queries used in multiple dashboards, create recording rules

4. **Set appropriate timeouts**: Some queries may need longer evaluation times, especially with many deployments

5. **Label filtering**: Always filter by namespace when possible to improve query performance

6. **Use instant queries for current state**: Use instant queries (without time ranges) when you only need the current value

7. **Time ranges for trends**: Use appropriate time ranges based on your metrics:
   - Heartbeat: 5-15 minutes
   - Restarts: 1-24 hours
   - Downtime: 1-24 hours
   - Recovery time: 1-24 hours

---

## Query Performance Optimization

```promql
# BAD: Queries all time series
k8s_deployment_status

# GOOD: Filter by namespace
k8s_deployment_status{namespace="production"}

# BETTER: Filter by multiple labels
k8s_deployment_status{namespace="production",deployment=~"api-.*"}

# BAD: Complex calculation on all deployments
(k8s_deployment_replicas_ready / k8s_deployment_replicas_desired) * 100

# GOOD: Use recording rule instead
deployment:health:percentage{namespace="production"}
```

---

## Common Use Cases

### 1. Deployment Health Check
```promql
# Quick health check for all deployments
k8s_deployment_status == 1
and
k8s_deployment_condition_status{condition="Available",status="True"} == 1
and
k8s_deployment_replicas_ready == k8s_deployment_replicas_desired
```

### 2. Troubleshooting Slow Deployments
```promql
# Find deployments with long recovery times
topk(10, k8s_deployment_recovery_time_milliseconds) > 30000
```

### 3. Capacity Planning
```promql
# Total desired replicas by namespace
sum by (namespace) (k8s_deployment_replicas_desired)

# Replica utilization
sum(k8s_deployment_replicas_ready) / sum(k8s_deployment_replicas_desired) * 100
```

### 4. SLA Monitoring
```promql
# Uptime percentage (last 30 days)
(1 - (sum(increase(k8s_deployment_downtime_duration_seconds[30d])) / (30 * 86400))) * 100
```

---

For more examples and updates, visit: https://github.com/XzByte/deployment-exporter
