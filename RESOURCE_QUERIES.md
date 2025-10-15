# Resource Usage Query Examples

## Deployment Availability Ratio

### View Availability in "X/Y" Format
```promql
# Shows availability with labels: available="2", desired="4"
k8s_deployment_availability_ratio

# Example output interpretation:
# k8s_deployment_availability_ratio{namespace="production",deployment="api",available="2",desired="4"} 0.5
# This means 2 out of 4 pods are available (50%)
```

### Filter Deployments with Low Availability
```promql
# Deployments with less than 100% availability
k8s_deployment_availability_ratio < 1

# Deployments with less than 50% availability
k8s_deployment_availability_ratio < 0.5

# Deployments with zero availability
k8s_deployment_availability_ratio == 0
```

### Sort by Availability
```promql
# Worst availability (lowest first)
sort(k8s_deployment_availability_ratio)

# Best availability (highest first)
sort_desc(k8s_deployment_availability_ratio)

# Bottom 10 deployments by availability
topk(10, k8s_deployment_availability_ratio) by (namespace, deployment, available, desired)

# Deployments sorted by missing replicas
sort_desc(k8s_deployment_replicas_desired - k8s_deployment_replicas_ready)
```

---

## CPU Usage Metrics

### Current CPU Usage
```promql
# Total CPU usage in millicores per deployment
k8s_deployment_cpu_usage_millicores

# CPU usage in cores (divide by 1000)
k8s_deployment_cpu_usage_millicores / 1000

# CPU usage as percentage of request
k8s_deployment_cpu_usage_percent
```

### Sort by Highest CPU Usage
```promql
# Top 10 deployments by CPU usage (millicores)
topk(10, k8s_deployment_cpu_usage_millicores)

# Top 20 deployments by CPU usage
topk(20, k8s_deployment_cpu_usage_millicores)

# All deployments sorted by CPU (highest first)
sort_desc(k8s_deployment_cpu_usage_millicores)

# Bottom 10 (lowest CPU usage)
bottomk(10, k8s_deployment_cpu_usage_millicores)

# Top 10 by CPU in cores
topk(10, k8s_deployment_cpu_usage_millicores / 1000)
```

### CPU Usage Percentage
```promql
# Top 10 deployments by CPU usage percentage
topk(10, k8s_deployment_cpu_usage_percent)

# Deployments using more than 80% of CPU request
k8s_deployment_cpu_usage_percent > 80

# Deployments using more than 100% (exceeding request)
k8s_deployment_cpu_usage_percent > 100

# Sort by CPU usage percentage
sort_desc(k8s_deployment_cpu_usage_percent)
```

### CPU Requests and Limits
```promql
# CPU requests per deployment (in millicores)
k8s_deployment_cpu_request_millicores

# CPU requests in cores
k8s_deployment_cpu_request_millicores / 1000

# CPU limits per deployment (in millicores)
k8s_deployment_cpu_limit_millicores

# CPU limits in cores
k8s_deployment_cpu_limit_millicores / 1000

# Deployments with no CPU limits set
k8s_deployment_cpu_limit_millicores == 0

# CPU headroom (limit - usage) in millicores
k8s_deployment_cpu_limit_millicores - k8s_deployment_cpu_usage_millicores

# Deployments near CPU limit (>90% of limit)
(k8s_deployment_cpu_usage_millicores / k8s_deployment_cpu_limit_millicores) * 100 > 90
```

---

## Memory Usage Metrics

### Current Memory Usage
```promql
# Total memory usage in MiB
k8s_deployment_memory_usage_mebibytes

# Memory usage in GiB
k8s_deployment_memory_usage_mebibytes / 1024

# Memory usage in bytes
k8s_deployment_memory_usage_mebibytes * 1024 * 1024

# Memory usage as percentage of request
k8s_deployment_memory_usage_percent
```

### Sort by Highest Memory Usage
```promql
# Top 10 deployments by memory usage (MiB)
topk(10, k8s_deployment_memory_usage_mebibytes)

# Top 10 by memory usage (GiB)
topk(10, k8s_deployment_memory_usage_mebibytes / 1024)

# All deployments sorted by memory (highest first)
sort_desc(k8s_deployment_memory_usage_mebibytes)

# Top 20 memory consumers
topk(20, k8s_deployment_memory_usage_mebibytes)
```

### Memory Usage Percentage
```promql
# Top 10 deployments by memory usage percentage
topk(10, k8s_deployment_memory_usage_percent)

# Deployments using more than 80% of memory request
k8s_deployment_memory_usage_percent > 80

# Deployments using more than 100% (exceeding request)
k8s_deployment_memory_usage_percent > 100

# Sort by memory usage percentage
sort_desc(k8s_deployment_memory_usage_percent)
```

### Memory Requests and Limits
```promql
# Memory requests per deployment (in MiB)
k8s_deployment_memory_request_mebibytes

# Memory requests in GiB
k8s_deployment_memory_request_mebibytes / 1024

# Memory limits per deployment (in MiB)
k8s_deployment_memory_limit_mebibytes

# Memory limits in GiB
k8s_deployment_memory_limit_mebibytes / 1024

# Deployments with no memory limits set
k8s_deployment_memory_limit_mebibytes == 0

# Memory headroom (limit - usage) in MiB
k8s_deployment_memory_limit_mebibytes - k8s_deployment_memory_usage_mebibytes

# Memory headroom in GiB
(k8s_deployment_memory_limit_mebibytes - k8s_deployment_memory_usage_mebibytes) / 1024

# Deployments near memory limit (>90% of limit)
(k8s_deployment_memory_usage_mebibytes / k8s_deployment_memory_limit_mebibytes) * 100 > 90
```

---

## Combined Resource Queries

### Top Resource Consumers (Multi-Metric)
```promql
# Top 10 by CPU usage (millicores)
topk(10, k8s_deployment_cpu_usage_millicores)

# Top 10 by memory usage (MiB)
topk(10, k8s_deployment_memory_usage_mebibytes)

# Top 10 by memory usage (GiB)
topk(10, k8s_deployment_memory_usage_mebibytes / 1024)

# Top 10 by CPU percentage
topk(10, k8s_deployment_cpu_usage_percent)

# Top 10 by memory percentage  
topk(10, k8s_deployment_memory_usage_percent)
```

### Resource Efficiency
```promql
# Deployments with low CPU efficiency (<20% of request used)
k8s_deployment_cpu_usage_percent < 20

# Deployments with low memory efficiency (<20% of request used)
k8s_deployment_memory_usage_percent < 20

# Over-provisioned deployments (using <30% of requests)
(k8s_deployment_cpu_usage_percent < 30) and (k8s_deployment_memory_usage_percent < 30)

# Under-provisioned deployments (using >90% of requests)
(k8s_deployment_cpu_usage_percent > 90) or (k8s_deployment_memory_usage_percent > 90)
```

### Resource Utilization by Namespace
```promql
# Total CPU usage per namespace (in millicores)
sum by (namespace) (k8s_deployment_cpu_usage_millicores)

# Total CPU usage per namespace (in cores)
sum by (namespace) (k8s_deployment_cpu_usage_millicores) / 1000

# Total memory usage per namespace (MiB)
sum by (namespace) (k8s_deployment_memory_usage_mebibytes)

# Total memory usage per namespace (GiB)
sum by (namespace) (k8s_deployment_memory_usage_mebibytes) / 1024

# Average CPU usage percent per namespace
avg by (namespace) (k8s_deployment_cpu_usage_percent)

# Average memory usage percent per namespace
avg by (namespace) (k8s_deployment_memory_usage_percent)

# Top 5 namespaces by CPU usage
topk(5, sum by (namespace) (k8s_deployment_cpu_usage_millicores))

# Top 5 namespaces by memory usage
topk(5, sum by (namespace) (k8s_deployment_memory_usage_mebibytes))
```

---

## Complete Resource Ranking

### Single Metric Sorts

```promql
# ============================================
# TOP 10 BY CPU USAGE (MILLICORES)
# ============================================
topk(10, k8s_deployment_cpu_usage_millicores)

# ============================================
# TOP 10 BY CPU USAGE (CORES)
# ============================================
topk(10, k8s_deployment_cpu_usage_millicores / 1000)

# ============================================
# TOP 10 BY MEMORY USAGE (MiB)
# ============================================
topk(10, k8s_deployment_memory_usage_mebibytes)

# ============================================
# TOP 10 BY MEMORY USAGE (GiB)
# ============================================
topk(10, k8s_deployment_memory_usage_mebibytes / 1024)

# ============================================
# TOP 10 BY CPU USAGE (PERCENTAGE)
# ============================================
topk(10, k8s_deployment_cpu_usage_percent)

# ============================================
# TOP 10 BY MEMORY USAGE (PERCENTAGE)
# ============================================
topk(10, k8s_deployment_memory_usage_percent)
```

### Cross-Metric Analysis

```promql
# Deployments using high CPU AND high memory (both >80%)
(k8s_deployment_cpu_usage_percent > 80) and (k8s_deployment_memory_usage_percent > 80)

# Deployments using high CPU but low memory
(k8s_deployment_cpu_usage_percent > 80) and (k8s_deployment_memory_usage_percent < 30)

# Deployments using high memory but low CPU
(k8s_deployment_memory_usage_percent > 80) and (k8s_deployment_cpu_usage_percent < 30)

# Resource imbalance score (difference between CPU% and Memory%)
abs(k8s_deployment_cpu_usage_percent - k8s_deployment_memory_usage_percent)

# Most balanced resource usage
sort(abs(k8s_deployment_cpu_usage_percent - k8s_deployment_memory_usage_percent))
```

---

## Grafana Table Queries

### Resource Overview Table
```promql
# Use "Format as: Table" and "Instant" query
# Shows: namespace, deployment, CPU usage (millicores), memory usage (MiB), availability

k8s_deployment_cpu_usage_millicores
or
k8s_deployment_memory_usage_mebibytes
or
k8s_deployment_cpu_usage_percent
or
k8s_deployment_memory_usage_percent
or
k8s_deployment_availability_ratio
```

### Top Consumers Table
Use these as separate queries in a table panel:

1. **CPU Usage (millicores)**: `topk(20, k8s_deployment_cpu_usage_millicores)`
2. **CPU Usage (cores)**: `topk(20, k8s_deployment_cpu_usage_millicores / 1000)`
3. **Memory Usage (MiB)**: `topk(20, k8s_deployment_memory_usage_mebibytes)`
4. **Memory Usage (GiB)**: `topk(20, k8s_deployment_memory_usage_mebibytes / 1024)`
5. **CPU %**: `topk(20, k8s_deployment_cpu_usage_percent)`
6. **Memory %**: `topk(20, k8s_deployment_memory_usage_percent)`
```

---

## Alerting Rules for Resources

```yaml
groups:
  - name: deployment-resources
    interval: 30s
    rules:
      - alert: DeploymentHighCPUUsage
        expr: k8s_deployment_cpu_usage_percent > 90
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High CPU usage in {{ $labels.namespace }}/{{ $labels.deployment }}"
          description: "CPU usage is {{ $value }}% of request"

      - alert: DeploymentHighMemoryUsage
        expr: k8s_deployment_memory_usage_percent > 90
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High memory usage in {{ $labels.namespace }}/{{ $labels.deployment }}"
          description: "Memory usage is {{ $value }}% of request"

      - alert: DeploymentCPUThrottling
        expr: k8s_deployment_cpu_usage_millicores >= k8s_deployment_cpu_limit_millicores
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "CPU throttling in {{ $labels.namespace }}/{{ $labels.deployment }}"
          description: "Deployment is hitting CPU limits ({{ $labels.cpu_millicores }}m / {{ $labels.cpu_limit_millicores }}m)"

      - alert: DeploymentOOMRisk
        expr: k8s_deployment_memory_usage_mebibytes >= k8s_deployment_memory_limit_mebibytes * 0.95
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "OOM risk in {{ $labels.namespace }}/{{ $labels.deployment }}"
          description: "Memory usage is {{ $value }}% of limit - OOM kill imminent ({{ $labels.memory_mebibytes }}MiB / {{ $labels.memory_limit_mebibytes }}MiB)"

      - alert: DeploymentLowAvailability
        expr: k8s_deployment_availability_ratio < 0.5
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Low availability in {{ $labels.namespace }}/{{ $labels.deployment }}"
          description: "Only {{ $labels.available }}/{{ $labels.desired }} replicas available"

      - alert: DeploymentPartiallyUnavailable
        expr: k8s_deployment_availability_ratio < 1 and k8s_deployment_availability_ratio > 0
        for: 10m
        labels:
          severity: info
        annotations:
          summary: "Partial unavailability in {{ $labels.namespace }}/{{ $labels.deployment }}"
          description: "{{ $labels.available }}/{{ $labels.desired }} replicas available"
```

---

## Recording Rules for Performance

```yaml
groups:
  - name: deployment-resource-recording
    interval: 30s
    rules:
      # Top resource consumers
      - record: top10:deployment:cpu_usage_millicores
        expr: topk(10, k8s_deployment_cpu_usage_millicores)

      - record: top10:deployment:memory_usage_mebibytes
        expr: topk(10, k8s_deployment_memory_usage_mebibytes)

      - record: top10:deployment:cpu_usage_percent
        expr: topk(10, k8s_deployment_cpu_usage_percent)

      - record: top10:deployment:memory_usage_percent
        expr: topk(10, k8s_deployment_memory_usage_percent)

      # Namespace aggregations
      - record: namespace:cpu_usage_millicores:sum
        expr: sum by (namespace) (k8s_deployment_cpu_usage_millicores)

      - record: namespace:memory_usage_mebibytes:sum
        expr: sum by (namespace) (k8s_deployment_memory_usage_mebibytes)

      - record: namespace:deployments:count
        expr: count by (namespace) (k8s_deployment_status)

      # Resource efficiency
      - record: deployment:resource_efficiency:avg
        expr: (k8s_deployment_cpu_usage_percent + k8s_deployment_memory_usage_percent) / 2

      - record: deployment:low_availability:count
        expr: count(k8s_deployment_availability_ratio < 1)
```

---

## Tips for Resource Monitoring

1. **Metrics Server Required**: Resource usage metrics require Kubernetes Metrics Server to be installed
2. **Scrape Interval**: Resource metrics update based on your scrape interval (default 15s)
3. **Historical Data**: Use `rate()` or `increase()` for trending over time
4. **Combine with Availability**: Cross-reference resource usage with availability for better insights
5. **Set Appropriate Thresholds**: Adjust alert thresholds based on your workload characteristics

---

## Common Troubleshooting Scenarios

### Scenario 1: High Memory, Low Availability
```promql
# Find deployments with high memory usage and low availability
(k8s_deployment_memory_usage_percent > 80) 
and 
(k8s_deployment_availability_ratio < 1)
```

### Scenario 2: Find Over-Provisioned Deployments
```promql
# Deployments with low resource usage (candidates for downsizing)
(k8s_deployment_cpu_usage_percent < 20)
and
(k8s_deployment_memory_usage_percent < 20)
and
(k8s_deployment_availability_ratio == 1)
```

### Scenario 3: Find Under-Provisioned Deployments
```promql
# Deployments that might need more resources
(k8s_deployment_cpu_usage_percent > 85 or k8s_deployment_memory_usage_percent > 85)
and
(k8s_deployment_availability_ratio < 1)
```

### Scenario 4: No Resource Limits Set
```promql
# Deployments without CPU or memory limits (risk of noisy neighbor)
k8s_deployment_cpu_limit_millicores == 0
or
k8s_deployment_memory_limit_mebibytes == 0
```

### Scenario 5: Resource Waste Calculation
```promql
# Calculate wasted CPU resources (millicores)
sum by (namespace, deployment) (
  k8s_deployment_cpu_request_millicores - k8s_deployment_cpu_usage_millicores
)

# Calculate wasted memory resources (MiB)
sum by (namespace, deployment) (
  k8s_deployment_memory_request_mebibytes - k8s_deployment_memory_usage_mebibytes
)
```

---

For the complete list of metrics and basic queries, see [QUERY_EXAMPLES.md](QUERY_EXAMPLES.md)
