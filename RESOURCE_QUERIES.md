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
# Total CPU usage in cores per deployment
k8s_deployment_cpu_usage_cores

# CPU usage in millicores
k8s_deployment_cpu_usage_cores * 1000

# CPU usage as percentage of request
k8s_deployment_cpu_usage_percent
```

### Sort by Highest CPU Usage
```promql
# Top 10 deployments by CPU usage (cores)
topk(10, k8s_deployment_cpu_usage_cores)

# Top 20 deployments by CPU usage
topk(20, k8s_deployment_cpu_usage_cores)

# All deployments sorted by CPU (highest first)
sort_desc(k8s_deployment_cpu_usage_cores)

# Bottom 10 (lowest CPU usage)
bottomk(10, k8s_deployment_cpu_usage_cores)
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
# CPU requests per deployment
k8s_deployment_cpu_request_cores

# CPU limits per deployment
k8s_deployment_cpu_limit_cores

# Deployments with no CPU limits set
k8s_deployment_cpu_limit_cores == 0

# CPU headroom (limit - usage)
k8s_deployment_cpu_limit_cores - k8s_deployment_cpu_usage_cores

# Deployments near CPU limit (>90% of limit)
(k8s_deployment_cpu_usage_cores / k8s_deployment_cpu_limit_cores) * 100 > 90
```

---

## Memory Usage Metrics

### Current Memory Usage
```promql
# Total memory usage in bytes
k8s_deployment_memory_usage_bytes

# Memory usage in MB
k8s_deployment_memory_usage_bytes / 1024 / 1024

# Memory usage in GB
k8s_deployment_memory_usage_bytes / 1024 / 1024 / 1024

# Memory usage as percentage of request
k8s_deployment_memory_usage_percent
```

### Sort by Highest Memory Usage
```promql
# Top 10 deployments by memory usage (bytes)
topk(10, k8s_deployment_memory_usage_bytes)

# Top 10 by memory usage (GB)
topk(10, k8s_deployment_memory_usage_bytes / 1024 / 1024 / 1024)

# All deployments sorted by memory (highest first)
sort_desc(k8s_deployment_memory_usage_bytes)

# Top 20 memory consumers
topk(20, k8s_deployment_memory_usage_bytes)
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
# Memory requests per deployment (in GB)
k8s_deployment_memory_request_bytes / 1024 / 1024 / 1024

# Memory limits per deployment (in GB)
k8s_deployment_memory_limit_bytes / 1024 / 1024 / 1024

# Deployments with no memory limits set
k8s_deployment_memory_limit_bytes == 0

# Memory headroom (limit - usage) in GB
(k8s_deployment_memory_limit_bytes - k8s_deployment_memory_usage_bytes) / 1024 / 1024 / 1024

# Deployments near memory limit (>90% of limit)
(k8s_deployment_memory_usage_bytes / k8s_deployment_memory_limit_bytes) * 100 > 90
```

---

## Combined Resource Queries

### Top Resource Consumers (Multi-Metric)
```promql
# Top 10 by CPU usage
topk(10, k8s_deployment_cpu_usage_cores)

# Top 10 by memory usage
topk(10, k8s_deployment_memory_usage_bytes / 1024 / 1024 / 1024)

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
# Total CPU usage per namespace
sum by (namespace) (k8s_deployment_cpu_usage_cores)

# Total memory usage per namespace (GB)
sum by (namespace) (k8s_deployment_memory_usage_bytes) / 1024 / 1024 / 1024

# Average CPU usage percent per namespace
avg by (namespace) (k8s_deployment_cpu_usage_percent)

# Average memory usage percent per namespace
avg by (namespace) (k8s_deployment_memory_usage_percent)

# Top 5 namespaces by CPU usage
topk(5, sum by (namespace) (k8s_deployment_cpu_usage_cores))

# Top 5 namespaces by memory usage
topk(5, sum by (namespace) (k8s_deployment_memory_usage_bytes))
```

---

## Complete Resource Ranking

### Single Metric Sorts

```promql
# ============================================
# TOP 10 BY CPU USAGE (ABSOLUTE)
# ============================================
topk(10, k8s_deployment_cpu_usage_cores)

# ============================================
# TOP 10 BY MEMORY USAGE (ABSOLUTE)
# ============================================
topk(10, k8s_deployment_memory_usage_bytes)

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
# Shows: namespace, deployment, CPU usage, memory usage, availability

k8s_deployment_cpu_usage_cores
or
k8s_deployment_memory_usage_bytes / 1024 / 1024 / 1024
or
k8s_deployment_cpu_usage_percent
or
k8s_deployment_memory_usage_percent
or
k8s_deployment_availability_ratio
```

### Top Consumers Table
Use these as separate queries in a table panel:

1. **CPU Usage**: `topk(20, k8s_deployment_cpu_usage_cores)`
2. **Memory Usage**: `topk(20, k8s_deployment_memory_usage_bytes / 1024 / 1024)`
3. **CPU %**: `topk(20, k8s_deployment_cpu_usage_percent)`
4. **Memory %**: `topk(20, k8s_deployment_memory_usage_percent)`

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
        expr: k8s_deployment_cpu_usage_cores >= k8s_deployment_cpu_limit_cores
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "CPU throttling in {{ $labels.namespace }}/{{ $labels.deployment }}"
          description: "Deployment is hitting CPU limits"

      - alert: DeploymentOOMRisk
        expr: k8s_deployment_memory_usage_bytes >= k8s_deployment_memory_limit_bytes * 0.95
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "OOM risk in {{ $labels.namespace }}/{{ $labels.deployment }}"
          description: "Memory usage is {{ $value }}% of limit - OOM kill imminent"

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
      - record: top10:deployment:cpu_usage_cores
        expr: topk(10, k8s_deployment_cpu_usage_cores)

      - record: top10:deployment:memory_usage_gb
        expr: topk(10, k8s_deployment_memory_usage_bytes / 1024 / 1024 / 1024)

      - record: top10:deployment:cpu_usage_percent
        expr: topk(10, k8s_deployment_cpu_usage_percent)

      - record: top10:deployment:memory_usage_percent
        expr: topk(10, k8s_deployment_memory_usage_percent)

      # Namespace aggregations
      - record: namespace:cpu_usage:sum
        expr: sum by (namespace) (k8s_deployment_cpu_usage_cores)

      - record: namespace:memory_usage:sum
        expr: sum by (namespace) (k8s_deployment_memory_usage_bytes)

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

---

For the complete list of metrics and basic queries, see [QUERY_EXAMPLES.md](QUERY_EXAMPLES.md)
