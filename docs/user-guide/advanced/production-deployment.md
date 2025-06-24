# Production Deployment: Deployment and Monitoring Guide

> **[Project Root](/) / [Documentation](/docs/) / [User Guide](/docs/user-guide/) / [Advanced Topics](/docs/user-guide/advanced/) / Production Deployment**

Comprehensive guide for deploying Go-LLMs applications to production, including infrastructure setup, deployment strategies, monitoring, scaling, and operational best practices.

## Deployment Overview

Production deployment involves:
- **Infrastructure Setup** - Cloud platforms, containers, orchestration
- **Deployment Strategies** - Blue-green, canary, rolling updates
- **Monitoring & Observability** - Metrics, logs, traces, alerts
- **Scaling & Performance** - Auto-scaling, load balancing
- **Security & Compliance** - Hardening, auditing, compliance

---

## Infrastructure Setup

### Container-Based Deployment

```dockerfile
# Multi-stage Dockerfile for Go-LLMs application
FROM golang:1.21-alpine AS builder

# Install dependencies
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1000 -S appgroup && \
    adduser -u 1000 -S appuser -G appgroup

# Copy binary from builder
COPY --from=builder /app/main /app/main

# Copy configuration files
COPY --from=builder /app/config /app/config

# Set ownership
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run application
ENTRYPOINT ["/app/main"]
```

### Kubernetes Deployment

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: gollms-app
  labels:
    app: gollms
spec:
  replicas: 3
  selector:
    matchLabels:
      app: gollms
  template:
    metadata:
      labels:
        app: gollms
    spec:
      containers:
      - name: gollms
        image: myregistry/gollms:v1.0.0
        ports:
        - containerPort: 8080
        env:
        - name: OPENAI_API_KEY
          valueFrom:
            secretKeyRef:
              name: llm-secrets
              key: openai-key
        - name: LOG_LEVEL
          value: "info"
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
        volumeMounts:
        - name: config
          mountPath: /app/config
          readOnly: true
      volumes:
      - name: config
        configMap:
          name: gollms-config

---
# service.yaml
apiVersion: v1
kind: Service
metadata:
  name: gollms-service
spec:
  selector:
    app: gollms
  ports:
    - protocol: TCP
      port: 80
      targetPort: 8080
  type: LoadBalancer

---
# hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: gollms-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: gollms-app
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

### Helm Chart Structure

```yaml
# Chart.yaml
apiVersion: v2
name: gollms
description: Go-LLMs Application Helm Chart
type: application
version: 1.0.0
appVersion: "1.0.0"

# values.yaml
replicaCount: 3

image:
  repository: myregistry/gollms
  pullPolicy: IfNotPresent
  tag: "v1.0.0"

service:
  type: LoadBalancer
  port: 80
  targetPort: 8080

ingress:
  enabled: true
  className: "nginx"
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
    nginx.ingress.kubernetes.io/rate-limit: "100"
  hosts:
    - host: api.example.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: gollms-tls
      hosts:
        - api.example.com

resources:
  limits:
    cpu: 1000m
    memory: 1Gi
  requests:
    cpu: 500m
    memory: 512Mi

autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70

secrets:
  openaiKey: ""
  anthropicKey: ""

config:
  logLevel: "info"
  providers:
    - name: openai
      enabled: true
    - name: anthropic
      enabled: true
```

---

## Deployment Strategies

### Blue-Green Deployment

```go
// Blue-Green deployment controller
type BlueGreenController struct {
    k8sClient kubernetes.Interface
    namespace string
}

func (bgc *BlueGreenController) Deploy(newVersion string) error {
    // Deploy to green environment
    greenDeployment := createDeployment("gollms-green", newVersion)
    if err := bgc.deployToEnvironment(greenDeployment); err != nil {
        return fmt.Errorf("green deployment failed: %w", err)
    }
    
    // Wait for green to be ready
    if err := bgc.waitForReady("gollms-green", 5*time.Minute); err != nil {
        return fmt.Errorf("green not ready: %w", err)
    }
    
    // Run smoke tests
    if err := bgc.runSmokeTests("gollms-green"); err != nil {
        bgc.rollback("gollms-green")
        return fmt.Errorf("smoke tests failed: %w", err)
    }
    
    // Switch traffic to green
    if err := bgc.switchTraffic("gollms-green"); err != nil {
        return fmt.Errorf("traffic switch failed: %w", err)
    }
    
    // Monitor for issues
    if err := bgc.monitorDeployment("gollms-green", 10*time.Minute); err != nil {
        bgc.switchTraffic("gollms-blue") // Rollback
        return fmt.Errorf("monitoring detected issues: %w", err)
    }
    
    // Cleanup blue environment
    return bgc.cleanup("gollms-blue")
}

func (bgc *BlueGreenController) switchTraffic(target string) error {
    service, err := bgc.k8sClient.CoreV1().Services(bgc.namespace).Get(
        context.Background(), "gollms-service", metav1.GetOptions{},
    )
    if err != nil {
        return err
    }
    
    service.Spec.Selector["version"] = target
    _, err = bgc.k8sClient.CoreV1().Services(bgc.namespace).Update(
        context.Background(), service, metav1.UpdateOptions{},
    )
    
    return err
}
```

### Canary Deployment

```yaml
# Canary deployment with Flagger
apiVersion: flagger.app/v1beta1
kind: Canary
metadata:
  name: gollms
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: gollms
  progressDeadlineSeconds: 60
  service:
    port: 80
    targetPort: 8080
    gateways:
    - public-gateway.istio-system.svc.cluster.local
    hosts:
    - api.example.com
  analysis:
    interval: 1m
    threshold: 5
    maxWeight: 50
    stepWeight: 10
    metrics:
    - name: request-success-rate
      threshold: 99
      interval: 1m
    - name: request-duration
      threshold: 500
      interval: 30s
    webhooks:
    - name: load-test
      url: http://flagger-loadtester.test/
      timeout: 5s
      metadata:
        cmd: "hey -z 1m -q 10 -c 2 http://gollms-canary.test:80/"
```

### Rolling Update with Health Checks

```go
// Custom rolling update controller
type RollingUpdateController struct {
    deploymentName string
    namespace      string
    k8sClient      kubernetes.Interface
    maxSurge       int
    maxUnavailable int
}

func (ruc *RollingUpdateController) Update(newImage string) error {
    deployment, err := ruc.getDeployment()
    if err != nil {
        return err
    }
    
    // Update image
    deployment.Spec.Template.Spec.Containers[0].Image = newImage
    
    // Configure rolling update strategy
    deployment.Spec.Strategy = appsv1.DeploymentStrategy{
        Type: appsv1.RollingUpdateDeploymentStrategyType,
        RollingUpdate: &appsv1.RollingUpdateDeployment{
            MaxSurge:       &intstr.IntOrString{IntVal: int32(ruc.maxSurge)},
            MaxUnavailable: &intstr.IntOrString{IntVal: int32(ruc.maxUnavailable)},
        },
    }
    
    // Apply update
    _, err = ruc.k8sClient.AppsV1().Deployments(ruc.namespace).Update(
        context.Background(), deployment, metav1.UpdateOptions{},
    )
    if err != nil {
        return err
    }
    
    // Monitor rollout
    return ruc.monitorRollout()
}

func (ruc *RollingUpdateController) monitorRollout() error {
    timeout := time.After(10 * time.Minute)
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-timeout:
            return errors.New("rollout timeout")
            
        case <-ticker.C:
            deployment, err := ruc.getDeployment()
            if err != nil {
                return err
            }
            
            if deployment.Status.UpdatedReplicas == *deployment.Spec.Replicas &&
               deployment.Status.AvailableReplicas == *deployment.Spec.Replicas {
                return nil // Rollout complete
            }
            
            // Check for failed pods
            if deployment.Status.UnavailableReplicas > 0 {
                pods, _ := ruc.getPodsForDeployment(deployment)
                for _, pod := range pods.Items {
                    if pod.Status.Phase == corev1.PodFailed {
                        return fmt.Errorf("pod %s failed: %s", 
                            pod.Name, pod.Status.Message)
                    }
                }
            }
        }
    }
}
```

---

## Monitoring and Observability

### Prometheus Metrics Setup

```go
// Comprehensive metrics instrumentation
type Metrics struct {
    requestDuration   *prometheus.HistogramVec
    requestsTotal     *prometheus.CounterVec
    requestsInFlight  *prometheus.GaugeVec
    tokensUsed        *prometheus.CounterVec
    providerErrors    *prometheus.CounterVec
    cacheHits         *prometheus.CounterVec
    cacheMisses       *prometheus.CounterVec
    activeUsers       prometheus.Gauge
    systemInfo        *prometheus.GaugeVec
}

func NewMetrics() *Metrics {
    m := &Metrics{
        requestDuration: prometheus.NewHistogramVec(
            prometheus.HistogramOpts{
                Name:    "gollms_request_duration_seconds",
                Help:    "Duration of requests in seconds",
                Buckets: []float64{0.1, 0.25, 0.5, 1, 2.5, 5, 10},
            },
            []string{"method", "endpoint", "status", "provider"},
        ),
        
        requestsTotal: prometheus.NewCounterVec(
            prometheus.CounterOpts{
                Name: "gollms_requests_total",
                Help: "Total number of requests",
            },
            []string{"method", "endpoint", "status"},
        ),
        
        tokensUsed: prometheus.NewCounterVec(
            prometheus.CounterOpts{
                Name: "gollms_tokens_used_total",
                Help: "Total tokens consumed",
            },
            []string{"provider", "model", "type"},
        ),
        
        providerErrors: prometheus.NewCounterVec(
            prometheus.CounterOpts{
                Name: "gollms_provider_errors_total",
                Help: "Provider errors by type",
            },
            []string{"provider", "error_type"},
        ),
    }
    
    // Register all metrics
    prometheus.MustRegister(
        m.requestDuration,
        m.requestsTotal,
        m.tokensUsed,
        m.providerErrors,
    )
    
    return m
}

// Middleware for automatic metrics collection
func (m *Metrics) Middleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.Request.URL.Path
        method := c.Request.Method
        
        // Track in-flight requests
        m.requestsInFlight.WithLabelValues(method, path).Inc()
        defer m.requestsInFlight.WithLabelValues(method, path).Dec()
        
        // Process request
        c.Next()
        
        // Record metrics
        duration := time.Since(start).Seconds()
        status := strconv.Itoa(c.Writer.Status())
        provider := c.GetString("provider")
        
        m.requestDuration.WithLabelValues(method, path, status, provider).Observe(duration)
        m.requestsTotal.WithLabelValues(method, path, status).Inc()
    }
}
```

### Grafana Dashboard Configuration

```json
{
  "dashboard": {
    "title": "Go-LLMs Production Metrics",
    "panels": [
      {
        "title": "Request Rate",
        "targets": [
          {
            "expr": "rate(gollms_requests_total[5m])",
            "legendFormat": "{{method}} {{endpoint}}"
          }
        ]
      },
      {
        "title": "Response Time (p95)",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(gollms_request_duration_seconds_bucket[5m]))",
            "legendFormat": "p95 - {{endpoint}}"
          }
        ]
      },
      {
        "title": "Token Usage",
        "targets": [
          {
            "expr": "rate(gollms_tokens_used_total[5m])",
            "legendFormat": "{{provider}} - {{model}}"
          }
        ]
      },
      {
        "title": "Error Rate",
        "targets": [
          {
            "expr": "rate(gollms_provider_errors_total[5m])",
            "legendFormat": "{{provider}} - {{error_type}}"
          }
        ]
      },
      {
        "title": "Cache Hit Rate",
        "targets": [
          {
            "expr": "rate(gollms_cache_hits[5m]) / (rate(gollms_cache_hits[5m]) + rate(gollms_cache_misses[5m]))",
            "legendFormat": "Cache Hit Rate"
          }
        ]
      }
    ]
  }
}
```

### Structured Logging

```go
// Production-ready logging configuration
func SetupLogging() *zap.Logger {
    config := zap.NewProductionConfig()
    
    // Configure based on environment
    if os.Getenv("ENVIRONMENT") == "development" {
        config = zap.NewDevelopmentConfig()
    }
    
    // Custom configuration
    config.OutputPaths = []string{"stdout", "/var/log/gollms/app.log"}
    config.ErrorOutputPaths = []string{"stderr", "/var/log/gollms/error.log"}
    config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
    
    // Add custom fields
    config.InitialFields = map[string]interface{}{
        "service": "gollms",
        "version": Version,
        "host":    os.Getenv("HOSTNAME"),
    }
    
    logger, _ := config.Build()
    
    // Replace global logger
    zap.ReplaceGlobals(logger)
    
    return logger
}

// Request logging with context
func LogRequest(ctx context.Context, logger *zap.Logger) {
    requestID := ctx.Value("request_id").(string)
    userID := ctx.Value("user_id").(string)
    
    logger.Info("Processing request",
        zap.String("request_id", requestID),
        zap.String("user_id", userID),
        zap.String("provider", ctx.Value("provider").(string)),
        zap.String("model", ctx.Value("model").(string)),
        zap.Int("input_tokens", ctx.Value("input_tokens").(int)),
    )
}
```

### Distributed Tracing

```go
// OpenTelemetry setup
func SetupTracing() (*trace.TracerProvider, error) {
    exporter, err := jaeger.New(
        jaeger.WithCollectorEndpoint(
            jaeger.WithEndpoint("http://jaeger:14268/api/traces"),
        ),
    )
    if err != nil {
        return nil, err
    }
    
    tp := trace.NewTracerProvider(
        trace.WithBatcher(exporter),
        trace.WithSampler(trace.AlwaysSample()),
        trace.WithResource(resource.NewWithAttributes(
            semconv.SchemaURL,
            semconv.ServiceNameKey.String("gollms"),
            semconv.ServiceVersionKey.String(Version),
            attribute.String("environment", os.Getenv("ENVIRONMENT")),
        )),
    )
    
    otel.SetTracerProvider(tp)
    otel.SetTextMapPropagator(propagation.TraceContext{})
    
    return tp, nil
}

// Trace LLM operations
func TraceLLMOperation(ctx context.Context, operation string) (context.Context, trace.Span) {
    tracer := otel.Tracer("gollms")
    ctx, span := tracer.Start(ctx, operation,
        trace.WithSpanKind(trace.SpanKindClient),
        trace.WithAttributes(
            attribute.String("llm.provider", ctx.Value("provider").(string)),
            attribute.String("llm.model", ctx.Value("model").(string)),
            attribute.String("llm.operation", operation),
        ),
    )
    
    return ctx, span
}
```

---

## Scaling Strategies

### Horizontal Pod Autoscaling

```yaml
# Advanced HPA configuration
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: gollms-advanced-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: gollms-app
  minReplicas: 3
  maxReplicas: 50
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
      - type: Percent
        value: 10
        periodSeconds: 60
      - type: Pods
        value: 2
        periodSeconds: 60
      selectPolicy: Min
    scaleUp:
      stabilizationWindowSeconds: 60
      policies:
      - type: Percent
        value: 100
        periodSeconds: 15
      - type: Pods
        value: 5
        periodSeconds: 15
      selectPolicy: Max
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
  - type: Pods
    pods:
      metric:
        name: gollms_requests_per_second
      target:
        type: AverageValue
        averageValue: "100"
  - type: External
    external:
      metric:
        name: queue_depth
        selector:
          matchLabels:
            queue: "llm-requests"
      target:
        type: Value
        value: "30"
```

### Load Balancing Configuration

```go
// Custom load balancer with health checks
type LoadBalancer struct {
    backends    []*Backend
    strategy    BalancingStrategy
    healthCheck *HealthChecker
}

type Backend struct {
    URL        string
    Weight     int
    Healthy    bool
    LastCheck  time.Time
    ErrorCount int
}

func (lb *LoadBalancer) SelectBackend() (*Backend, error) {
    healthyBackends := lb.getHealthyBackends()
    if len(healthyBackends) == 0 {
        return nil, errors.New("no healthy backends available")
    }
    
    switch lb.strategy {
    case RoundRobin:
        return lb.roundRobin(healthyBackends)
    case LeastConnections:
        return lb.leastConnections(healthyBackends)
    case WeightedRandom:
        return lb.weightedRandom(healthyBackends)
    default:
        return healthyBackends[0], nil
    }
}

// Health checking
func (lb *LoadBalancer) StartHealthChecks(ctx context.Context) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            for _, backend := range lb.backends {
                go lb.checkBackendHealth(backend)
            }
        case <-ctx.Done():
            return
        }
    }
}

func (lb *LoadBalancer) checkBackendHealth(backend *Backend) {
    client := &http.Client{Timeout: 5 * time.Second}
    resp, err := client.Get(backend.URL + "/health")
    
    backend.LastCheck = time.Now()
    
    if err != nil || resp.StatusCode != http.StatusOK {
        backend.ErrorCount++
        if backend.ErrorCount >= 3 {
            backend.Healthy = false
            log.Printf("Backend %s marked unhealthy", backend.URL)
        }
    } else {
        backend.Healthy = true
        backend.ErrorCount = 0
    }
}
```

### Queue-Based Scaling

```go
// Auto-scaling based on queue depth
type QueueScaler struct {
    k8sClient   kubernetes.Interface
    queueClient QueueClient
    deployment  string
    namespace   string
}

func (qs *QueueScaler) ScaleBasedOnQueue(ctx context.Context) error {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            depth, err := qs.queueClient.GetQueueDepth()
            if err != nil {
                log.Printf("Failed to get queue depth: %v", err)
                continue
            }
            
            desiredReplicas := qs.calculateDesiredReplicas(depth)
            if err := qs.scaleDeployment(desiredReplicas); err != nil {
                log.Printf("Failed to scale: %v", err)
            }
            
        case <-ctx.Done():
            return ctx.Err()
        }
    }
}

func (qs *QueueScaler) calculateDesiredReplicas(queueDepth int) int32 {
    // Scale based on queue depth
    // 1 replica per 100 messages
    replicas := int32(queueDepth/100) + 1
    
    // Apply bounds
    if replicas < 3 {
        replicas = 3
    } else if replicas > 50 {
        replicas = 50
    }
    
    return replicas
}
```

---

## Security Hardening

### Runtime Security

```yaml
# Security policies
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: gollms-psp
spec:
  privileged: false
  allowPrivilegeEscalation: false
  requiredDropCapabilities:
    - ALL
  volumes:
    - 'configMap'
    - 'emptyDir'
    - 'projected'
    - 'secret'
    - 'downwardAPI'
    - 'persistentVolumeClaim'
  hostNetwork: false
  hostIPC: false
  hostPID: false
  runAsUser:
    rule: 'MustRunAsNonRoot'
  seLinux:
    rule: 'RunAsAny'
  supplementalGroups:
    rule: 'RunAsAny'
  fsGroup:
    rule: 'RunAsAny'
  readOnlyRootFilesystem: true

---
# Network policies
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: gollms-netpol
spec:
  podSelector:
    matchLabels:
      app: gollms
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: ingress-nginx
    ports:
    - protocol: TCP
      port: 8080
  egress:
  - to:
    - namespaceSelector: {}
    ports:
    - protocol: TCP
      port: 443  # HTTPS for LLM APIs
    - protocol: TCP
      port: 53   # DNS
  - to:
    - podSelector:
        matchLabels:
          app: redis
    ports:
    - protocol: TCP
      port: 6379
```

### Secret Management

```go
// Vault integration for secrets
type VaultSecretManager struct {
    client *vault.Client
    path   string
}

func NewVaultSecretManager(addr, token string) (*VaultSecretManager, error) {
    config := vault.DefaultConfig()
    config.Address = addr
    
    client, err := vault.NewClient(config)
    if err != nil {
        return nil, err
    }
    
    client.SetToken(token)
    
    return &VaultSecretManager{
        client: client,
        path:   "secret/data/gollms",
    }, nil
}

func (vsm *VaultSecretManager) GetAPIKey(provider string) (string, error) {
    secret, err := vsm.client.Logical().Read(vsm.path)
    if err != nil {
        return "", err
    }
    
    if secret == nil || secret.Data == nil {
        return "", errors.New("secret not found")
    }
    
    data, ok := secret.Data["data"].(map[string]interface{})
    if !ok {
        return "", errors.New("invalid secret format")
    }
    
    key, ok := data[provider+"_api_key"].(string)
    if !ok {
        return "", fmt.Errorf("API key for %s not found", provider)
    }
    
    return key, nil
}

// Kubernetes secret rotation
func (vsm *VaultSecretManager) RotateSecrets(ctx context.Context) error {
    ticker := time.NewTicker(24 * time.Hour)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            // Generate new API keys
            newKeys, err := vsm.generateNewKeys()
            if err != nil {
                log.Printf("Failed to generate keys: %v", err)
                continue
            }
            
            // Update Vault
            if err := vsm.updateVault(newKeys); err != nil {
                log.Printf("Failed to update Vault: %v", err)
                continue
            }
            
            // Update Kubernetes secrets
            if err := vsm.updateK8sSecrets(newKeys); err != nil {
                log.Printf("Failed to update K8s: %v", err)
                continue
            }
            
            // Restart pods to pick up new secrets
            if err := vsm.restartPods(); err != nil {
                log.Printf("Failed to restart pods: %v", err)
            }
            
        case <-ctx.Done():
            return ctx.Err()
        }
    }
}
```

---

## Disaster Recovery

### Backup and Restore

```go
// Automated backup system
type BackupManager struct {
    storage     StorageBackend
    databases   []Database
    schedule    string
    retention   time.Duration
}

func (bm *BackupManager) ScheduleBackups(ctx context.Context) error {
    c := cron.New()
    
    _, err := c.AddFunc(bm.schedule, func() {
        if err := bm.performBackup(); err != nil {
            log.Printf("Backup failed: %v", err)
            // Send alert
            alerting.SendAlert("Backup Failed", err.Error())
        }
    })
    
    if err != nil {
        return err
    }
    
    c.Start()
    
    <-ctx.Done()
    c.Stop()
    
    return nil
}

func (bm *BackupManager) performBackup() error {
    timestamp := time.Now().Format("20060102-150405")
    
    for _, db := range bm.databases {
        // Create database dump
        dump, err := db.Dump()
        if err != nil {
            return fmt.Errorf("database dump failed: %w", err)
        }
        
        // Compress dump
        compressed, err := compress(dump)
        if err != nil {
            return fmt.Errorf("compression failed: %w", err)
        }
        
        // Encrypt backup
        encrypted, err := encrypt(compressed)
        if err != nil {
            return fmt.Errorf("encryption failed: %w", err)
        }
        
        // Store backup
        backupName := fmt.Sprintf("%s-backup-%s.enc", db.Name(), timestamp)
        if err := bm.storage.Store(backupName, encrypted); err != nil {
            return fmt.Errorf("storage failed: %w", err)
        }
    }
    
    // Clean old backups
    return bm.cleanOldBackups()
}

// Disaster recovery procedures
type DisasterRecovery struct {
    primary     *Region
    secondary   *Region
    replication *Replication
}

func (dr *DisasterRecovery) Failover() error {
    // Check primary health
    if dr.primary.IsHealthy() {
        return errors.New("primary is healthy, failover not needed")
    }
    
    // Verify secondary is ready
    if !dr.secondary.IsReady() {
        return errors.New("secondary not ready for failover")
    }
    
    // Stop replication
    if err := dr.replication.Stop(); err != nil {
        log.Printf("Failed to stop replication: %v", err)
    }
    
    // Promote secondary
    if err := dr.secondary.Promote(); err != nil {
        return fmt.Errorf("promotion failed: %w", err)
    }
    
    // Update DNS
    if err := dr.updateDNS(dr.secondary.Endpoint); err != nil {
        return fmt.Errorf("DNS update failed: %w", err)
    }
    
    // Notify operations team
    alerting.SendAlert("Failover Completed", 
        fmt.Sprintf("Failed over from %s to %s", 
            dr.primary.Name, dr.secondary.Name))
    
    return nil
}
```

---

## Operational Runbooks

### Deployment Runbook

```markdown
# Go-LLMs Deployment Runbook

## Pre-Deployment Checklist
- [ ] All tests passing in CI/CD
- [ ] Security scan completed
- [ ] Performance benchmarks acceptable
- [ ] Database migrations ready
- [ ] Rollback plan documented
- [ ] Team notified of deployment window

## Deployment Steps

1. **Prepare Environment**
   ```bash
   # Verify cluster access
   kubectl config current-context
   
   # Check current deployment status
   kubectl get deployments -n production
   
   # Backup current configuration
   kubectl get deployment gollms -o yaml > backup/gollms-$(date +%Y%m%d).yaml
   ```

2. **Deploy New Version**
   ```bash
   # Update image tag
   kubectl set image deployment/gollms gollms=myregistry/gollms:v1.2.0 -n production
   
   # Monitor rollout
   kubectl rollout status deployment/gollms -n production
   ```

3. **Verify Deployment**
   ```bash
   # Check pod status
   kubectl get pods -n production -l app=gollms
   
   # Check logs
   kubectl logs -n production -l app=gollms --tail=100
   
   # Run smoke tests
   ./scripts/smoke-tests.sh production
   ```

4. **Monitor Metrics**
   - Check Grafana dashboard: https://grafana.example.com/d/gollms
   - Verify error rates remain below 1%
   - Confirm response times < 500ms p95
   - Monitor token usage for anomalies

## Rollback Procedure

1. **Immediate Rollback**
   ```bash
   # Rollback to previous version
   kubectl rollout undo deployment/gollms -n production
   
   # Verify rollback
   kubectl rollout status deployment/gollms -n production
   ```

2. **Restore from Backup**
   ```bash
   # Apply backup configuration
   kubectl apply -f backup/gollms-20240115.yaml
   ```

## Post-Deployment
- [ ] Update deployment log
- [ ] Notify stakeholders
- [ ] Schedule retrospective if issues occurred
- [ ] Update documentation if needed
```

### Incident Response

```go
// Incident response automation
type IncidentResponder struct {
    pagerDuty   PagerDutyClient
    slack       SlackClient
    runbooks    map[string]Runbook
    metrics     MetricsClient
}

func (ir *IncidentResponder) HandleAlert(alert Alert) error {
    // Create incident
    incident := &Incident{
        Title:       alert.Title,
        Description: alert.Description,
        Severity:    ir.determineSeverity(alert),
        StartTime:   time.Now(),
    }
    
    // Page on-call if critical
    if incident.Severity == SeverityCritical {
        if err := ir.pagerDuty.CreateIncident(incident); err != nil {
            log.Printf("Failed to page: %v", err)
        }
    }
    
    // Post to Slack
    if err := ir.slack.PostToChannel("#incidents", incident.Format()); err != nil {
        log.Printf("Failed to post to Slack: %v", err)
    }
    
    // Execute automated remediation
    if runbook, ok := ir.runbooks[alert.Type]; ok {
        go ir.executeRunbook(runbook, incident)
    }
    
    // Collect diagnostics
    diagnostics := ir.collectDiagnostics(alert)
    incident.Diagnostics = diagnostics
    
    return nil
}

func (ir *IncidentResponder) collectDiagnostics(alert Alert) map[string]interface{} {
    return map[string]interface{}{
        "metrics":     ir.metrics.GetMetrics(alert.Labels, time.Hour),
        "logs":        ir.getLogs(alert.Labels, time.Hour),
        "traces":      ir.getTraces(alert.Labels, time.Hour),
        "pod_status":  ir.getPodStatus(alert.Labels),
        "recent_deployments": ir.getRecentDeployments(),
    }
}
```

---

## Production Checklist

### Pre-Production
- [ ] **Infrastructure**
  - [ ] Kubernetes cluster provisioned
  - [ ] Load balancers configured
  - [ ] SSL certificates installed
  - [ ] DNS records configured
  - [ ] CDN setup (if applicable)

- [ ] **Security**
  - [ ] API keys stored in Vault/Secrets Manager
  - [ ] Network policies configured
  - [ ] Pod security policies applied
  - [ ] RBAC roles defined
  - [ ] Audit logging enabled

- [ ] **Monitoring**
  - [ ] Prometheus configured
  - [ ] Grafana dashboards created
  - [ ] Alerts configured
  - [ ] Log aggregation setup
  - [ ] Tracing enabled

- [ ] **Backup & Recovery**
  - [ ] Backup strategy implemented
  - [ ] Restore procedures tested
  - [ ] Disaster recovery plan documented
  - [ ] RTO/RPO defined

### Go-Live
- [ ] **Deployment**
  - [ ] CI/CD pipeline tested
  - [ ] Blue-green deployment verified
  - [ ] Rollback procedures tested
  - [ ] Health checks configured
  - [ ] Smoke tests passing

- [ ] **Performance**
  - [ ] Load testing completed
  - [ ] Auto-scaling configured
  - [ ] Rate limiting enabled
  - [ ] Caching strategy implemented
  - [ ] CDN configured

- [ ] **Operations**
  - [ ] Runbooks created
  - [ ] On-call rotation setup
  - [ ] Incident response tested
  - [ ] Documentation complete
  - [ ] Team trained

### Post-Production
- [ ] **Monitoring**
  - [ ] All metrics collecting
  - [ ] Alerts firing correctly
  - [ ] Dashboards accessible
  - [ ] SLOs defined
  - [ ] Error budgets tracked

- [ ] **Optimization**
  - [ ] Performance baselines established
  - [ ] Cost optimization reviewed
  - [ ] Capacity planning updated
  - [ ] Security posture assessed
  - [ ] Lessons learned documented

---

## Next Steps

- **[Security Considerations](security-considerations.md)** - Security deep dive
- **[Performance Optimization](performance-optimization.md)** - Performance tuning
- **[Troubleshooting Guide](troubleshooting.md)** - Problem resolution
- **[Best Practices Checklist](/docs/user-guide/reference/best-practices-checklist.md)** - Complete checklist
- **[Configuration Reference](/docs/user-guide/reference/configuration-reference.md)** - Configuration details