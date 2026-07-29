# Integration Layer — طراحی تفصیلی
## Adapter Pattern + Circuit Breaker + Outbox + Retry

> لایه حیاتی اتصال به سازمان — جایی که یک خطا می‌تواند کل زنجیره را متوقف کند.

---

## ۱. معماری کلی

```
┌─────────────────────────────────────────────────────────────┐
│                     Integration Service                      │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐   │
│  │                   Adapter Registry                     │   │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────────────┐   │   │
│  │  │   Map    │  │  Claim   │  │   Certificate    │   │   │
│  │  │ Adapter  │  │ Adapter  │  │    Adapter       │   │   │
│  │  └────┬─────┘  └────┬─────┘  └────────┬─────────┘   │   │
│  └───────┼─────────────┼─────────────────┼──────────────┘   │
│          └─────────────┼─────────────────┘                   │
│                        │                                      │
│                        ▼                                      │
│  ┌──────────────────────────────────────────────────────┐   │
│  │              Organization HTTP Client                  │   │
│  │  ┌──────────┐ ┌──────────────┐ ┌─────────────────┐   │   │
│  │  │ Circuit  │ │   Retry      │ │   Timeout        │   │   │
│  │  │ Breaker  │ │   Policy     │ │   Manager        │   │   │
│  │  │ (3 fails │ │ (backoff)    │ │   (connect:5s   │   │   │
│  │  │ → open)  │ │              │ │    request:30s)  │   │   │
│  │  └──────────┘ └──────────────┘ └─────────────────┘   │   │
│  └──────────────────────────────────────────────────────┘   │
│                                                              │
│  ┌──────────────────────┐  ┌─────────────────────────┐     │
│  │   Outbox Publisher    │  │   Dead Letter Queue      │     │
│  │   (NATS JetStream)   │  │   Handler                │     │
│  └──────────────────────┘  └─────────────────────────┘     │
└─────────────────────────────────────────────────────────────┘
                             │
                             ▼
                  ┌──────────────────────┐
                  │  مرکز تبادل اطلاعات    │
                  │  ملی (سازمان ثبت)      │
                  └──────────────────────┘
```

---

## ۲. Adapter Interface

```go
// integration/adapter.go
package integration

import (
    "context"
    "time"
    "github.com/google/uuid"
)

// OrgRequest represents a standardized request to the organization
type OrgRequest struct {
    ID            uuid.UUID              `json:"id"`
    ServiceType   string                 `json:"service_type"`   // map | claim | cert
    Action        string                 `json:"action"`         // submit | status | cancel
    Payload       json.RawMessage        `json:"payload"`        // service-specific payload
    Timestamp     time.Time              `json:"timestamp"`
    CorrelationID uuid.UUID              `json:"correlation_id"`
    RetryCount    int                    `json:"retry_count"`
    Metadata      map[string]string      `json:"metadata"`
}

// OrgResponse represents a standardized response from the organization
type OrgResponse struct {
    RequestID     uuid.UUID              `json:"request_id"`
    Status        string                 `json:"status"`         // approved | rejected | processing
    TrackingCode  string                 `json:"tracking_code,omitempty"`
    RejectReason  string                 `json:"reject_reason,omitempty"`
    RawResponse   json.RawMessage        `json:"raw_response"`
    ProcessedAt   time.Time              `json:"processed_at"`
}

// OrganizationAdapter defines the contract for all org integrations
type OrganizationAdapter interface {
    // Name returns the adapter identifier
    Name() string

    // ServiceType returns which service this adapter handles
    ServiceType() string

    // Validate checks if the data is ready for submission
    Validate(ctx context.Context, resourceID uuid.UUID) error

    // BuildPayload constructs the organization-specific request payload
    BuildPayload(ctx context.Context, resourceID uuid.UUID) (json.RawMessage, error)

    // Submit sends the request to the organization
    Submit(ctx context.Context, request OrgRequest) (*OrgResponse, error)

    // ParseCallback parses an incoming callback/webhook from the organization
    ParseCallback(ctx context.Context, raw []byte) (*OrgResponse, error)

    // GetStatus checks the status of a previously submitted request
    GetStatus(ctx context.Context, trackingCode string) (*OrgResponse, error)
}
```

---

## ۳. پیاده‌سازی Adapter برای هر سرویس

### ۳.۱ Map Adapter

```go
// integration/adapters/map_adapter.go
package adapters

type MapAdapter struct {
    client  *OrganizationHTTPClient
    baseURL string
}

func (a *MapAdapter) Name() string        { return "map_service_adapter" }
func (a *MapAdapter) ServiceType() string { return "map" }

func (a *MapAdapter) Validate(ctx context.Context, resourceID uuid.UUID) error {
    ms, err := repo.GetMapService(ctx, resourceID)
    if err != nil {
        return fmt.Errorf("map service not found: %w", err)
    }

    var errs []string

    // نقشه آپلود شده؟
    if ms.MapFilePath == "" {
        errs = append(errs, "فایل نقشه آپلود نشده است")
    }

    // ۴ عکس با EXIF valid؟
    photos, _ := repo.GetMapPhotos(ctx, resourceID)
    if len(photos) != 4 {
        errs = append(errs, "هر ۴ عکس از اضلاع ملک الزامی است")
    }
    for _, p := range photos {
        if !p.ExifValid {
            errs = append(errs, fmt.Sprintf("اعتبارسنجی Geo-tag عکس %s ناموفق: %s", p.Side, p.ExifValidationNote))
        }
    }

    // جدول توصیفی کامل؟
    if ms.DescriptiveTable == nil {
        errs = append(errs, "جدول توصیفی (فرمت مانا) تکمیل نشده است")
    }

    // رضایت متقاضی؟
    if ms.ConsentGrantedAt == nil {
        errs = append(errs, "رضایت متقاضی ثبت نشده است")
    }

    if len(errs) > 0 {
        return &ValidationError{Reasons: errs}
    }
    return nil
}

func (a *MapAdapter) BuildPayload(ctx context.Context, resourceID uuid.UUID) (json.RawMessage, error) {
    ms, _ := repo.GetMapService(ctx, resourceID)
    c, _ := repo.GetCase(ctx, ms.CaseID)
    applicant, _ := repo.GetUser(ctx, c.ApplicantID)
    photos, _ := repo.GetMapPhotos(ctx, resourceID)

    payload := MapSubmissionPayload{
        // متقاضی
        ApplicantNationalID: applicant.NationalID,
        ApplicantName:       applicant.FirstName + " " + applicant.LastName,
        ApplicantMobile:     applicant.Mobile,
        ApplicantCapacity:   string(c.ApplicantCapacity),

        // ملک
        PropertyType:     ms.PropertyType,
        ApproxArea:       ms.ApproxAreaSqm,
        LandUse:          ms.LandUse,
        OwnershipType:    ms.OwnershipType,
        HasBuilding:      ms.HasBuilding,
        AnnexCount:       ms.AnnexCount,
        Province:         c.Province,
        City:             c.City,
        District:         c.District,
        Village:          c.Village,
        PostalCode:       c.PostalCode,
        Address:          c.AddressDetail,
        Latitude:         ms.GeoLatitude,
        Longitude:        ms.GeoLongitude,

        // نقشه
        MapFileURL:          ms.MapFilePath,     // سازمان از URL ما دانلود می‌کند
        MapFormat:           ms.MapFormat,
        DescriptiveTable:    ms.DescriptiveTable,

        // عکس‌ها
        Photos: make([]PhotoInfo, len(photos)),
    }

    for i, p := range photos {
        payload.Photos[i] = PhotoInfo{
            Side:       p.Side,
            FileURL:    p.FilePath,
            Latitude:   p.PhotoLatitude,
            Longitude:  p.PhotoLongitude,
            TakenAt:    p.PhotoTakenAt,
        }
    }

    return json.Marshal(payload)
}
```

### ۳.۲ Claim Adapter

```go
// integration/adapters/claim_adapter.go
package adapters

type ClaimAdapter struct {
    client  *OrganizationHTTPClient
    baseURL string
}

func (a *ClaimAdapter) Validate(ctx context.Context, resourceID uuid.UUID) error {
    cs, err := repo.GetClaimService(ctx, resourceID)
    if err != nil {
        return err
    }

    var errs []string

    // کد رهگیری نقشه معتبر؟
    if !cs.MapTrackingValid {
        errs = append(errs, "کد رهگیری نقشه نامعتبر است یا دسترسی وجود ندارد")
    }

    // رضایت + هشدار ادعای واهی ارسال شده؟
    if !cs.FalseClaimWarningSent {
        errs = append(errs, "هشدار ادعای واهی ارسال نشده است")
    }
    if cs.ConsentGrantedAt == nil {
        errs = append(errs, "رضایت متقاضی ثبت نشده است")
    }

    // اطلاعات ادعا کامل؟
    if cs.ClaimType == "" {
        errs = append(errs, "نوع ادعا مشخص نشده است")
    }
    if cs.MainPlateNumber == "" {
        errs = append(errs, "پلاک ثبتی اصلی ثبت نشده است")
    }

    // مستندات تایید شده توسط کارشناس؟
    docs, _ := repo.GetClaimDocuments(ctx, resourceID)
    verifiedCount := 0
    for _, d := range docs {
        if d.VerifiedAt != nil {
            verifiedCount++
        }
    }
    if len(docs) == 0 {
        errs = append(errs, "حداقل یک مستند الزامی است")
    }
    if verifiedCount != len(docs) {
        errs = append(errs, "تمام مستندات باید توسط کارشناس حقوقی تایید شوند")
    }

    // حقوق دولتی؟
    if cs.HasGovernmentRights && cs.TreasuryPaymentRef == "" {
        errs = append(errs, "شماره واریز حقوق دولتی به خزانه ثبت نشده است")
    }

    if len(errs) > 0 {
        return &ValidationError{Reasons: errs}
    }
    return nil
}
```

---

## ۴. Organization HTTP Client

```go
// integration/client.go
package integration

import (
    "net/http"
    "time"
    "github.com/sony/gobreaker"
)

type OrganizationHTTPClient struct {
    httpClient  *http.Client
    baseURL     string
    apiKey      string
    circuitBr   *gobreaker.CircuitBreaker
    retryPolicy RetryPolicy
}

type RetryPolicy struct {
    MaxAttempts  int
    BaseDelay    time.Duration   // 1 second
    MaxDelay     time.Duration   // 16 seconds
    Multiplier   float64         // 2.0 (exponential backoff)
}

func DefaultRetryPolicy() RetryPolicy {
    return RetryPolicy{
        MaxAttempts: 3,
        BaseDelay:   1 * time.Second,
        MaxDelay:    16 * time.Second,
        Multiplier:  2.0,
    }
}

func NewOrganizationHTTPClient(baseURL, apiKey string) *OrganizationHTTPClient {
    return &OrganizationHTTPClient{
        httpClient: &http.Client{
            Timeout: 30 * time.Second,
            Transport: &http.Transport{
                DialContext: (&net.Dialer{
                    Timeout:   5 * time.Second,   // connection timeout
                    KeepAlive: 30 * time.Second,
                }).DialContext,
                TLSHandshakeTimeout:   5 * time.Second,
                ResponseHeaderTimeout: 30 * time.Second,
                MaxIdleConns:          100,
                MaxIdleConnsPerHost:   10,
            },
        },
        baseURL: baseURL,
        apiKey:  apiKey,
        circuitBr: gobreaker.NewCircuitBreaker(gobreaker.Settings{
            Name:        "organization-api",
            MaxRequests: 3,                            // half-open state: allow 3 test requests
            Interval:    60 * time.Second,             // reset failure count after 60s
            Timeout:     30 * time.Second,             // open → half-open after 30s
            ReadyToTrip: func(counts gobreaker.Counts) bool {
                failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
                return counts.Requests >= 5 && failureRatio >= 0.6 // open after 5+ requests with 60% failure
            },
        }),
        retryPolicy: DefaultRetryPolicy(),
    }
}

func (c *OrganizationHTTPClient) Do(ctx context.Context, method, path string, body []byte) (*http.Response, error) {
    return c.circuitBr.Execute(func() (interface{}, error) {
        return c.doWithRetry(ctx, method, path, body)
    })
}

func (c *OrganizationHTTPClient) doWithRetry(ctx context.Context, method, path string, body []byte) (*http.Response, error) {
    var lastErr error

    for attempt := 0; attempt < c.retryPolicy.MaxAttempts; attempt++ {
        if attempt > 0 {
            delay := time.Duration(float64(c.retryPolicy.BaseDelay) *
                math.Pow(c.retryPolicy.Multiplier, float64(attempt-1)))
            if delay > c.retryPolicy.MaxDelay {
                delay = c.retryPolicy.MaxDelay
            }
            log.Info("retrying organization request",
                "path", path,
                "attempt", attempt+1,
                "delay_ms", delay.Milliseconds())

            select {
            case <-time.After(delay):
            case <-ctx.Done():
                return nil, ctx.Err()
            }
        }

        req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bytes.NewReader(body))
        if err != nil {
            return nil, err
        }

        req.Header.Set("Content-Type", "application/json")
        req.Header.Set("Authorization", "Bearer "+c.apiKey)
        req.Header.Set("X-Correlation-ID", uuid.New().String())
        req.Header.Set("X-Retry-Count", strconv.Itoa(attempt))

        resp, err := c.httpClient.Do(req)
        if err != nil {
            lastErr = err
            // Retry on connection errors
            if isRetryableError(err) {
                continue
            }
            return nil, err
        }

        // Retry on server errors (5xx) and 429 (rate limit)
        if resp.StatusCode >= 500 || resp.StatusCode == 429 {
            resp.Body.Close()
            lastErr = fmt.Errorf("server error: %d", resp.StatusCode)
            continue
        }

        return resp, nil
    }

    return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

func isRetryableError(err error) bool {
    if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
        return true
    }
    if errors.Is(err, syscall.ECONNREFUSED) ||
       errors.Is(err, syscall.ECONNRESET) ||
       errors.Is(err, syscall.EPIPE) {
        return true
    }
    return false
}
```

### ۴.۱ Circuit Breaker States

```
     ┌──────────┐   failure threshold reached   ┌──────────┐
     │  CLOSED  │ ──────────────────────────────→│   OPEN   │
     │ (normal) │                                 │ (reject) │
     └──────────┘                                 └────┬─────┘
          ▲                                            │
          │          timeout elapsed                   │
          │     (30 seconds)                           │
          │                                            ▼
          │                                   ┌──────────────┐
          └───────────────────────────────────│  HALF-OPEN   │
             success rate > threshold          │ (test with 3 │
                                               │  requests)   │
                                               └──────────────┘
```

---

## ۵. Outbox Pattern — اطمینان از ارسال

### ۵.۱ مشکل

اگر بعد از commit تراکنش دیتابیس، قبل از ارسال به سازمان، سرویس crash کند → درخواست گم می‌شود.

### ۵.۲ راه‌حل: Transactional Outbox

```go
// integration/outbox.go
package integration

type OutboxMessage struct {
    ID            uuid.UUID
    AggregateType string    // map_service | claim_service | cert_service
    AggregateID   uuid.UUID
    EventType     string    // submit_to_org | status_check
    Payload       json.RawMessage
    CreatedAt     time.Time
    ProcessedAt   *time.Time
    RetryCount    int
    Status        string    // pending | processing | sent | failed
}

// OutboxPublisher runs in a goroutine, polling for unprocessed messages
type OutboxPublisher struct {
    db        *sql.DB
    client    *OrganizationHTTPClient
    eventBus  *nats.Conn
}

func (p *OutboxPublisher) Start(ctx context.Context) {
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            p.processBatch(ctx)
        }
    }
}

func (p *OutboxPublisher) processBatch(ctx context.Context) {
    tx, _ := p.db.BeginTx(ctx, nil)
    defer tx.Rollback()

    // Fetch pending messages with FOR UPDATE SKIP LOCKED
    rows, _ := tx.QueryContext(ctx, `
        SELECT id, aggregate_type, aggregate_id, event_type, payload, retry_count
        FROM outbox_messages
        WHERE status = 'pending'
           OR (status = 'failed' AND retry_count < 3)
        ORDER BY created_at
        LIMIT 50
        FOR UPDATE SKIP LOCKED
    `)
    defer rows.Close()

    var messages []OutboxMessage
    for rows.Next() {
        var m OutboxMessage
        rows.Scan(&m.ID, &m.AggregateType, &m.AggregateID,
            &m.EventType, &m.Payload, &m.RetryCount)
        messages = append(messages, m)
    }

    for _, msg := range messages {
        err := p.publishMessage(ctx, msg)

        if err != nil {
            // Mark as failed — will retry later
            tx.ExecContext(ctx, `
                UPDATE outbox_messages
                SET status = 'failed',
                    retry_count = retry_count + 1,
                    last_error = $2
                WHERE id = $1
            `, msg.ID, err.Error())

            // If all retries exhausted → move to DLQ
            if msg.RetryCount >= 2 {
                p.moveToDLQ(ctx, msg)
            }
        } else {
            // Mark as sent
            tx.ExecContext(ctx, `
                UPDATE outbox_messages
                SET status = 'sent', processed_at = NOW()
                WHERE id = $1
            `, msg.ID)
        }
    }

    tx.Commit()
}
```

### ۵.۳ استفاده در Application Code

```go
// در سرویس MapService — وقتی کارشناس نقشه را submit می‌کند

func (s *MapService) SubmitToOrganization(ctx context.Context, mapID uuid.UUID) error {
    tx, _ := s.db.BeginTx(ctx, nil)
    defer tx.Rollback()

    // 1. تغییر وضعیت MapService
    tx.ExecContext(ctx, `
        UPDATE map_services
        SET status = 'submitted_to_org',
            submitted_to_org_at = NOW(),
            updated_at = NOW()
        WHERE id = $1
    `, mapID)

    // 2. ذخیره پیام در Outbox (در همان تراکنش!)
    payload, _ := s.adapter.BuildPayload(ctx, mapID)
    tx.ExecContext(ctx, `
        INSERT INTO outbox_messages (id, aggregate_type, aggregate_id, event_type, payload)
        VALUES ($1, 'map_service', $2, 'submit_to_org', $3)
    `, uuid.New(), mapID, payload)

    // 3. ثبت Audit Log (در همان تراکنش!)
    tx.ExecContext(ctx, `
        INSERT INTO audit_logs (event_type, actor_type, actor_id, resource_type, resource_id, changes)
        VALUES ('map.submitted_to_org', 'system', $1, 'map_service', $2, $3)
    `, actorID, mapID, changes)

    // 4. Commit — همه با هم
    return tx.Commit()
    // OutboxPublisher در background پیام را از outbox_messages می‌خواند و به سازمان می‌فرستد.
    // اگر crash شود: پیام در outbox_messages می‌ماند و بعد از restart پردازش می‌شود.
}
```

---

## ۶. Dead Letter Queue

```go
// integration/dlq.go
package integration

func (p *OutboxPublisher) moveToDLQ(ctx context.Context, msg OutboxMessage) {
    p.eventBus.Publish("integration.dlq", DLQMessage{
        OriginalMessage: msg,
        MovedAt:         time.Now(),
        Reason:          "max retries exhausted",
    })

    log.Error("message moved to DLQ",
        "aggregate_type", msg.AggregateType,
        "aggregate_id", msg.AggregateID,
        "retry_count", msg.RetryCount)

    // Alert admin
    p.alertService.Send(Alert{
        Severity: "critical",
        Title:    "ارسال به سازمان ناموفق — نیاز به مداخله دستی",
        Body:     fmt.Sprintf(
            "درخواست %s (شناسه: %s) پس از %d بار تلاش ناموفق ماند. لطفاً بررسی کنید.",
            msg.AggregateType, msg.AggregateID, msg.RetryCount,
        ),
        ActionURL: fmt.Sprintf("/admin/integration/failed/%s", msg.ID),
        Tags:      []string{"integration", "dlq", msg.AggregateType},
    })
}
```

---

## ۷. Callback/Webhook از سازمان

```go
// integration/webhook.go
package integration

type WebhookHandler struct {
    adapters   map[string]OrganizationAdapter
    eventBus   *nats.Conn
    allowedIPs []net.IPNet // فقط IPهای مجاز سازمان
}

func (h *WebhookHandler) Handle(w http.ResponseWriter, r *http.Request) {
    // 1. IP whitelist check
    clientIP := net.ParseIP(r.RemoteAddr)
    if !h.isAllowedIP(clientIP) {
        http.Error(w, "forbidden", http.StatusForbidden)
        return
    }

    // 2. Verify signature (if organization provides HMAC)
    signature := r.Header.Get("X-Org-Signature")
    body, _ := io.ReadAll(r.Body)
    if !h.verifySignature(body, signature) {
        http.Error(w, "invalid signature", http.StatusUnauthorized)
        return
    }

    // 3. Parse and route to appropriate adapter
    serviceType := r.URL.Query().Get("service") // map | claim | cert
    adapter, ok := h.adapters[serviceType]
    if !ok {
        http.Error(w, "unknown service", http.StatusBadRequest)
        return
    }

    response, err := adapter.ParseCallback(r.Context(), body)
    if err != nil {
        http.Error(w, "parse error", http.StatusBadRequest)
        return
    }

    // 4. Publish event to workflow
    h.eventBus.Publish(fmt.Sprintf("%s.%s.org.callback", serviceType, response.RequestID), response)

    // 5. Respond immediately — processing happens asynchronously
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"status": "received"})
}
```

---

## ۸. مانیتورینگ Integration

```go
// integration/metrics.go
package integration

var (
    orgRequestTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "org_requests_total",
            Help: "Total organization API requests",
        },
        []string{"service", "action", "status"}, // status: success | failure | timeout
    )

    orgRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "org_request_duration_seconds",
            Help:    "Organization API request duration",
            Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 30},
        },
        []string{"service", "action"},
    )

    circuitBreakerState = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "org_circuit_breaker_state",
            Help: "Circuit breaker state (0=closed, 1=half-open, 2=open)",
        },
        []string{"service"},
    )

    outboxQueueSize = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "outbox_queue_size",
            Help: "Number of pending outbox messages",
        },
    )

    dlqSize = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "dlq_size",
            Help: "Number of messages in Dead Letter Queue",
        },
    )
)
```

### Alert Rules (Prometheus)

```yaml
groups:
  - name: integration
    rules:
      - alert: OrgAPIDown
        expr: org_circuit_breaker_state{service="map"} == 2
        for: 2m
        annotations:
          summary: "اتصال به سازمان قطع است (Circuit Breaker Open)"
          description: "سرویس {{ $labels.service }} بیش از ۲ دقیقه در دسترس نیست."

      - alert: OutboxQueueGrowing
        expr: outbox_queue_size > 50
        for: 5m
        annotations:
          summary: "صف Outbox در حال رشد است"
          description: "{{ $value }} پیام در صف Outbox انباشته شده."

      - alert: DLQNotEmpty
        expr: dlq_size > 0
        for: 1m
        annotations:
          summary: "صف Dead Letter خالی نیست"
          description: "{{ $value }} پیام در DLQ — نیاز به بررسی دستی ادمین."

      - alert: HighFailureRate
        expr: rate(org_requests_total{status="failure"}[5m]) / rate(org_requests_total[5m]) > 0.1
        for: 5m
        annotations:
          summary: "نرخ شکست سازمان بالاست"
          description: "بیش از ۱۰٪ درخواست‌ها در ۵ دقیقه اخیر ناموفق بوده‌اند."
```

---

## ۹. خروجی جدول Outbox

```sql
CREATE TABLE outbox_messages (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    aggregate_type  VARCHAR(50) NOT NULL,
    aggregate_id    UUID NOT NULL,
    event_type      VARCHAR(100) NOT NULL,
    payload         JSONB NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'pending',
    retry_count     INT NOT NULL DEFAULT 0,
    last_error      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at    TIMESTAMPTZ,
    correlation_id  UUID
);

CREATE INDEX idx_outbox_pending ON outbox_messages (status, created_at)
    WHERE status IN ('pending', 'failed');
CREATE INDEX idx_outbox_aggregate ON outbox_messages (aggregate_type, aggregate_id);
```

---

> **فایل‌های مرتبط:** [api-contract.yaml](api-contract.yaml) — endpointهای callback سازمان، [database-schema.sql](database-schema.sql) — جدول integration_logs
