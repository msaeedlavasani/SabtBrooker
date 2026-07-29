# Workflow Engine — طراحی تفصیلی
## موتور گردش‌کار + State Machine + Saga Orchestrator

> قلب تپنده سامانه — مسئول ردیابی وضعیت پرونده در کل زنجیره سه‌مرحله‌ای و تضمین صحت transitionها.

---

## ۱. معماری رویدادمحور (Event-Driven)

### اصل اساسی

هر تغییر وضعیت در سیستم یک **Domain Event** منتشر می‌کند. سرویس‌های دیگر با subscribe کردن روی این events، واکنش نشان می‌دهند — بدون dependency مستقیم.

```
┌──────────────────────────────────────────────────────────┐
│                     NATS JetStream                        │
│                                                           │
│  Stream: workflow.events                                  │
│  Subjects:                                                │
│    case.{case_id}.status.changed                         │
│    map.{map_id}.status.changed                           │
│    claim.{claim_id}.status.changed                       │
│    cert.{cert_id}.status.changed                         │
│    payment.{payment_id}.completed                         │
│    deadline.{case_id}.approaching                        │
│    deadline.{case_id}.expired                            │
└──────────────────────────────────────────────────────────┘
         ▲                ▲                ▲
         │                │                │
    ┌────────┐      ┌─────────┐      ┌─────────┐
    │  Case  │      │   Map   │      │  Claim  │
    │Service │      │ Service │      │ Service │
    └────────┘      └─────────┘      └─────────┘
         │                │                │
         ▼                ▼                ▼
    ┌─────────────────────────────────────────┐
    │        Workflow Saga Orchestrator        │
    │  (compensating transactions coordinator)│
    └─────────────────────────────────────────┘
```

---

## ۲. State Machine — تعریف کامل

### ۲.۱ وضعیت‌های Case

```
                         ┌──────────┐
                         │  DRAFT   │ ◄── شروع
                         └────┬─────┘
                              │
                    ┌─────────┼──────────┐
                    │ [create map]       │ [cancel]
                    ▼                    ▼
           ┌────────────────┐    ┌───────────┐
           │MAP_IN_PROGRESS │    │ CANCELLED │ ◄── پایان
           └───────┬────────┘    └───────────┘
                   │
                   │ [map approved by org]
                   ▼
           ┌────────────────┐
           │ MAP_COMPLETED  │
           └───────┬────────┘
                   │
        ┌──────────┼──────────┐
        │ [create claim]      │ [cancel]
        ▼                     ▼
  ┌──────────────────┐  ┌───────────┐
  │CLAIM_IN_PROGRESS │  │ CANCELLED │
  └────────┬─────────┘  └───────────┘
           │
           │ [claim approved by org]
           ▼
  ┌──────────────────┐
  │ CLAIM_COMPLETED  │────── (deadline_2years = now + 2y)
  └────────┬─────────┘
           │              │
     ┌─────┼──────┐       │ [deadline exceeded]
     │ [create cert]     ▼
     ▼             ┌───────────┐
┌─────────────────┐│  EXPIRED  │ ◄── پایان
│CERT_IN_PROGRESS │└───────────┘
└────────┬────────┘
         │
         │ [cert approved by org]
         ▼
  ┌──────────────────┐
  │ CERT_COMPLETED   │ ◄── پایان زنجیره (موفق)
  └──────────────────┘
```

### ۲.۲ Transition Table کامل

| # | از | به | Guard Condition | Side Effect |
|---|---|---|---|---|
| T1 | `draft` | `map_in_progress` | applicant identity verified | ایجاد MapService, تخصیص خودکار کارشناسان, event: `case.map.started` |
| T2 | `draft` | `cancelled` | (none) | event: `case.cancelled` |
| T3 | `map_in_progress` | `map_completed` | map_service.status = `approved` | ذخیره map_tracking_code در case, event: `case.map.completed` |
| T4 | `map_in_progress` | `cancelled` | (none) | لغو MapService مرتبط, event: `case.cancelled` |
| T5 | `map_completed` | `claim_in_progress` | applicant capacity verified by expert | ایجاد ClaimService, event: `case.claim.started` |
| T6 | `map_completed` | `cancelled` | (none) | event: `case.cancelled` |
| T7 | `claim_in_progress` | `claim_completed` | claim_service.status = `approved` | ذخیره claim_tracking_code, تنظیم deadline_2years, event: `case.claim.completed`, job: `deadline_2years` |
| T8 | `claim_in_progress` | `rejected` | claim_service.status = `rejected` (rejection غیرقابل اصلاح) | event: `case.claim.rejected` |
| T9 | `claim_completed` | `cert_in_progress` | claim age < 2 years | ایجاد CertService, event: `case.cert.started` |
| T10 | `claim_completed` | `expired` | now > deadline_2years (توسط scheduler) | event: `case.expired`, SMS به متقاضی |
| T11 | `cert_in_progress` | `cert_completed` | cert_service.status = `approved` | ذخیره cert_tracking_code, completed_at = now, event: `case.completed` |
| T12 | `cert_in_progress` | `expired` | (توسط scheduler) deadline exceeded | event: `case.expired`, SMS به متقاضی/وراث |

### ۲.۳ وضعیت‌های MapService

```
pending_expert_assignment ──→ expert_assigned ──→ fieldwork_in_progress
                                                       │
                                                       ▼
                                               fieldwork_done
                                                       │
                                                  ┌────┴────┐
                                                  ▼         ▼
                                            submitted    rejected
                                            _to_org      (برگشت به fieldwork)
                                                  │
                                             ┌────┴────┐
                                             ▼         ▼
                                         approved   rejected
```

### ۲.۴ وضعیت‌های ClaimService

```
pending_expert_assignment ──→ expert_assigned ──→ documents_verified
                                                       │
                                                  ┌────┴────┐
                                                  ▼         ▼
                                            submitted    rejected
                                            _to_org      (عدم احراز)
                                                  │
                                             ┌────┴────┐
                                             ▼         ▼
                                         approved   rejected
```

### ۲.۵ وضعیت‌های CertService

```
pending_data ──→ submitted_to_org ──→ approved
                          │
                          ▼
                       rejected
```

---

## ۳. Go Implementation — State Machine

### ۳.۱ Interface اصلی

```go
// workflow/state.go
package workflow

import "time"

// Status represents a state in the state machine
type Status string

// Transition defines a valid move from one state to another
type Transition struct {
    From   Status
    To     Status
    Guard  GuardFunc        // must return nil to allow transition
    Effect EffectFunc       // executed after transition
    Events []string         // NATS subjects to publish
}

// GuardFunc validates whether a transition is allowed
type GuardFunc func(ctx context.Context, resourceID uuid.UUID) error

// EffectFunc performs side effects after a successful transition
type EffectFunc func(ctx context.Context, resourceID uuid.UUID) error

// StateMachine manages transitions for a resource type
type StateMachine struct {
    resourceType string
    transitions  map[Status]map[Status]*Transition // from → to → transition
    currentState func(ctx context.Context, id uuid.UUID) (Status, error)
    applyState   func(ctx context.Context, id uuid.UUID, from, to Status) error
}
```

### ۳.۲ Case State Machine

```go
// workflow/case_statemachine.go
package workflow

func NewCaseStateMachine(
    db *sql.DB,
    eventBus *nats.Conn,
    scheduler *Scheduler,
) *StateMachine {
    sm := &StateMachine{
        resourceType: "case",
        transitions:  make(map[Status]map[Status]*Transition),
    }

    // T1: draft → map_in_progress
    sm.AddTransition(Transition{
        From:  "draft",
        To:    "map_in_progress",
        Guard: allOf(
            identityVerified,      // کد ملی + ثنا + شاهکار
            notDeceased,           // متقاضی زنده باشد
        ),
        Effect: sequence(
            autoAssignExperts,     // تخصیص خودکار کارشناس
            createMapService,      // ایجاد رکورد MapService
            notifyApplicant("فرآیند تهیه نقشه آغاز شد"),     // پیامک
            notifyExperts("پرونده جدید تخصیص یافت"),         // نوتیفیکیشن
        ),
        Events: []string{
            "case.{id}.status.changed",
            "case.{id}.map.started",
        },
    })

    // T3: map_in_progress → map_completed
    sm.AddTransition(Transition{
        From:  "map_in_progress",
        To:    "map_completed",
        Guard: allOf(
            mapServiceApproved,    // MapService.status == approved
        ),
        Effect: sequence(
            storeMapTrackingCode,  // ذخیره کد رهگیری نقشه
            notifyApplicant("نقشه ثبتی تایید شد — کد رهگیری: {code}"),
        ),
        Events: []string{
            "case.{id}.status.changed",
            "case.{id}.map.completed",
        },
    })

    // T5: map_completed → claim_in_progress
    sm.AddTransition(Transition{
        From:  "map_completed",
        To:    "claim_in_progress",
        Guard: allOf(
            capacityVerified,      // کارشناس حقوقی نمایندگی را تایید کرده
        ),
        Effect: sequence(
            createClaimService,
            notifyApplicant("فرآیند درج ادعا آغاز شد"),
        ),
        Events: []string{
            "case.{id}.status.changed",
            "case.{id}.claim.started",
        },
    })

    // T7: claim_in_progress → claim_completed
    sm.AddTransition(Transition{
        From:  "claim_in_progress",
        To:    "claim_completed",
        Guard: allOf(
            claimServiceApproved,
        ),
        Effect: sequence(
            storeClaimTrackingCode,
            setDeadline2Years,      // deadline = now + 2 years
            scheduler.Schedule(DeadlineJob{
                CaseID:    resourceID,
                Deadline:  time.Now().Add(2 * 365 * 24 * time.Hour),
                JobType:   "deadline_2years",
            }),
            notifyApplicant("ادعا ثبت شد — کد رهگیری: {code}"),
        ),
        Events: []string{
            "case.{id}.status.changed",
            "case.{id}.claim.completed",
        },
    })

    // T9: claim_completed → cert_in_progress
    sm.AddTransition(Transition{
        From:  "claim_completed",
        To:    "cert_in_progress",
        Guard: allOf(
            withinDeadline,         // now < deadline_2years
        ),
        Effect: sequence(
            createCertService,
            handleDeceasedApplicant, // اگر متوفی: تنظیم deadline_5months
            notifyApplicant("فرآیند درج گواهی اقدام آغاز شد"),
        ),
        Events: []string{
            "case.{id}.status.changed",
            "case.{id}.cert.started",
        },
    })

    // T11: cert_in_progress → cert_completed
    sm.AddTransition(Transition{
        From:  "cert_in_progress",
        To:    "cert_completed",
        Guard: allOf(
            certServiceApproved,
        ),
        Effect: sequence(
            storeCertTrackingCode,
            markCompleted,          // completed_at = now
            scheduler.CancelDeadline(caseID), // لغو jobهای deadline
            notifyApplicant("گواهی اقدام ثبت شد — فرآیند تکمیل گردید"),
            notifyAdmin("پرونده {id} تکمیل شد"),
        ),
        Events: []string{
            "case.{id}.status.changed",
            "case.{id}.completed",
        },
    })

    return sm
}

// Transition executes a state change with guard and effect
func (sm *StateMachine) Transition(
    ctx context.Context,
    resourceID uuid.UUID,
    to Status,
) error {
    current, err := sm.currentState(ctx, resourceID)
    if err != nil {
        return fmt.Errorf("failed to get current state: %w", err)
    }

    t, ok := sm.transitions[current][to]
    if !ok {
        return &InvalidTransitionError{
            ResourceType: sm.resourceType,
            ResourceID:   resourceID,
            From:         current,
            To:           to,
        }
    }

    // 1. Guard check
    if t.Guard != nil {
        if err := t.Guard(ctx, resourceID); err != nil {
            return &GuardFailedError{
                Transition: t,
                Reason:     err,
            }
        }
    }

    // 2. Apply state change (atomic — in transaction)
    if err := sm.applyState(ctx, resourceID, current, to); err != nil {
        return fmt.Errorf("failed to apply state: %w", err)
    }

    // 3. Side effects (async — failure here does NOT rollback state)
    if t.Effect != nil {
        go func() {
            if err := t.Effect(context.Background(), resourceID); err != nil {
                log.Error("side effect failed", "transition", t, "error", err)
                // Side effect failure is logged but does not block
            }
        }()
    }

    // 4. Publish events
    for _, subject := range t.Events {
        subj := strings.ReplaceAll(subject, "{id}", resourceID.String())
        sm.eventBus.Publish(subj, StateChangeEvent{
            ResourceType: sm.resourceType,
            ResourceID:   resourceID,
            From:         current,
            To:           to,
            Timestamp:    time.Now(),
        })
    }

    return nil
}
```

### ۳.۳ Guard Conditions (پیاده‌سازی)

```go
// workflow/guards.go
package workflow

// identityVerified: استعلام‌های هویتی انجام شده باشد
func identityVerified(ctx context.Context, caseID uuid.UUID) error {
    c, err := repo.GetCase(ctx, caseID)
    if err != nil {
        return err
    }
    applicant, err := repo.GetUser(ctx, c.ApplicantID)
    if err != nil {
        return err
    }

    var errs []string
    if !applicant.MobileVerified {
        errs = append(errs, "شماره موبایل تایید نشده")
    }
    if !applicant.NCRMobileMatch {
        errs = append(errs, "تطابق کد ملی و موبایل احراز نشده (شاهکار)")
    }
    if applicant.SanaStatus != "active" {
        errs = append(errs, "ثبت‌نام در سامانه ثنا تایید نشده")
    }
    if applicant.BirthDate != nil {
        age := time.Now().Year() - applicant.BirthDate.Year()
        if age < 18 {
            errs = append(errs, "سن کمتر از ۱۸ سال")
        }
    }

    if len(errs) > 0 {
        return &GuardError{Reasons: errs}
    }
    return nil
}

// notDeceased: متقاضی زنده باشد
func notDeceased(ctx context.Context, caseID uuid.UUID) error {
    c, _ := repo.GetCase(ctx, caseID)
    applicant, _ := repo.GetUser(ctx, c.ApplicantID)
    if !applicant.IsAlive {
        return errors.New("متقاضی در قید حیات نیست — نیاز به وراث")
    }
    return nil
}

// capacityVerified: کارشناس حقوقی نمایندگی را تایید کرده
func capacityVerified(ctx context.Context, caseID uuid.UUID) error {
    c, _ := repo.GetCase(ctx, caseID)
    if !c.ProxyVerified {
        return errors.New("احراز نمایندگی توسط کارشناس حقوقی تایید نشده")
    }
    return nil
}

// mapServiceApproved: سرویس نقشه توسط سازمان تایید شده
func mapServiceApproved(ctx context.Context, caseID uuid.UUID) error {
    ms, err := repo.GetMapServiceByCase(ctx, caseID)
    if err != nil || ms.Status != "approved" {
        return errors.New("نقشه ثبتی هنوز تایید نشده")
    }
    return nil
}

// withinDeadline: آیا هنوز در مهلت ۲ ساله هستیم
func withinDeadline(ctx context.Context, caseID uuid.UUID) error {
    c, _ := repo.GetCase(ctx, caseID)
    if c.Deadline2Years != nil && time.Now().After(*c.Deadline2Years) {
        return fmt.Errorf("مهلت ۲ ساله به پایان رسیده (%s)", c.Deadline2Years.Format("1385/01/02"))
    }
    return nil
}

// claimServiceApproved, certServiceApproved مشابه mapServiceApproved
```

---

## ۴. Saga Orchestrator — تراکنش‌های توزیع‌شده

### ۴.۱ چرا Saga؟

یک گردش‌کار کامل (مثلاً «ارسال نقشه به سازمان») شامل:

1. **MapService:** تغییر وضعیت به `submitted_to_org` (دیتابیس MapService)
2. **Integration:** ارسال به سازمان از طریق API (خارج از سیستم)
3. **Case:** در صورت تایید → تغییر به `map_completed` (دیتابیس Case)
4. **Notification:** ارسال پیامک به متقاضی
5. **AuditLog:** ثبت رویداد

این ۵ مرحله در ۳ سرویس مختلف اتفاق می‌افتد. Saga تضمین می‌کند که در صورت شکست هر مرحله، مراحل قبلی compensate شوند.

### ۴.۲ الگو: Choreography-based Saga (با NATS)

```
┌──────────────┐  event: map.submit.requested   ┌──────────────┐
│ MapService   │ ──────────────────────────────→│ Integration  │
│ (status:     │                                  │ Service      │
│  submitting) │←───────────────────────────────│ (ارسال به    │
└──────────────┘  event: map.submit.sent         │  سازمان)     │
                                                 └──────┬───────┘
                                                        │
                              event: map.org.response    │
                                                        │
                      ┌─────────────────────────────────┘
                      ▼
              ┌──────────────┐
              │ Workflow     │
              │ Saga Handler │
              └──────┬───────┘
                     │
         ┌───────────┼───────────┐
         ▼           ▼           ▼
   ┌──────────┐ ┌──────────┐ ┌──────────┐
   │  Update  │ │ Notify   │ │  Audit   │
   │  Status  │ │ Applicant│ │   Log    │
   └──────────┘ └──────────┘ └──────────┘
```

### ۴.۳ پیاده‌سازی Saga Step

```go
// workflow/saga.go
package workflow

type SagaDefinition struct {
    Name   string
    Steps  []SagaStep
}

type SagaStep struct {
    Name       string
    Action     func(ctx context.Context, data SagaData) error
    Compensate func(ctx context.Context, data SagaData) error // rollback
    Timeout    time.Duration
    MaxRetries int
}

type SagaData map[string]interface{}

type SagaOrchestrator struct {
    eventBus  *nats.Conn
    sagas     map[string]*SagaDefinition
    store     SagaStore // persistence for saga state
}

// Execute runs a saga and compensates on failure
func (so *SagaOrchestrator) Execute(
    ctx context.Context,
    sagaName string,
    data SagaData,
) error {
    saga, ok := so.sagas[sagaName]
    if !ok {
        return fmt.Errorf("saga %s not found", sagaName)
    }

    sagaID := uuid.New()
    completedSteps := []int{}

    for i, step := range saga.Steps {
        // Execute step with retry
        var err error
        for attempt := 0; attempt <= step.MaxRetries; attempt++ {
            ctx, cancel := context.WithTimeout(ctx, step.Timeout)
            err = step.Action(ctx, data)
            cancel()
            if err == nil {
                break
            }
            log.Warn("saga step failed, retrying",
                "saga", sagaName, "step", step.Name,
                "attempt", attempt+1, "error", err)
            time.Sleep(time.Duration(attempt+1) * time.Second) // linear backoff
        }

        if err != nil {
            log.Error("saga step failed after retries, starting compensation",
                "saga", sagaName, "step", step.Name, "error", err)

            // Compensate completed steps in reverse order
            for j := len(completedSteps) - 1; j >= 0; j-- {
                compStep := saga.Steps[completedSteps[j]]
                if compStep.Compensate != nil {
                    if compErr := compStep.Compensate(ctx, data); compErr != nil {
                        log.Error("compensation failed",
                            "saga", sagaName, "step", compStep.Name, "error", compErr)
                        // Compensation failure is critical — manual intervention needed
                        so.notifyAdmin(sagaID, sagaName, compStep.Name, compErr)
                    }
                }
            }
            return fmt.Errorf("saga %s failed at step %s: %w", sagaName, step.Name, err)
        }

        completedSteps = append(completedSteps, i)
    }

    return nil
}
```

### ۴.۴ تعریف Saga برای «ارسال نقشه به سازمان»

```go
// workflow/sagas/map_submission.go
func MapSubmissionSaga() *SagaDefinition {
    return &SagaDefinition{
        Name: "map_submission",
        Steps: []SagaStep{
            {
                Name: "validate_map_data",
                Action: func(ctx context.Context, data SagaData) error {
                    mapID := data["map_service_id"].(uuid.UUID)
                    return validateMapCompleteness(ctx, mapID)
                    // checks: نقشه آپلود شده؟ جدول توصیفی کامل؟ ۴ عکس با EXIF valid؟
                },
                // No compensation needed — read-only validation
            },
            {
                Name: "set_submitting_status",
                Action: func(ctx context.Context, data SagaData) error {
                    mapID := data["map_service_id"].(uuid.UUID)
                    return repo.UpdateMapStatus(ctx, mapID, "submitted_to_org")
                },
                Compensate: func(ctx context.Context, data SagaData) error {
                    mapID := data["map_service_id"].(uuid.UUID)
                    return repo.UpdateMapStatus(ctx, mapID, "fieldwork_done")
                },
            },
            {
                Name: "send_to_organization",
                Action: func(ctx context.Context, data SagaData) error {
                    mapID := data["map_service_id"].(uuid.UUID)
                    return integrationService.SendMapToOrg(ctx, mapID)
                },
                Compensate: func(ctx context.Context, data SagaData) error {
                    // اگر ارسال موفق بود و سازمان جواب داد، ولی مراحل بعد شکست خورد
                    // نمی‌توان ارسال را undo کرد — فقط لاگ می‌کنیم
                    log.Error("cannot undo org submission",
                        "map_id", data["map_service_id"])
                    return nil
                },
                MaxRetries: 3,
            },
            {
                Name: "log_integration",
                Action: func(ctx context.Context, data SagaData) error {
                    return auditLog.Record(ctx, AuditEvent{
                        Type:    "map.submitted_to_org",
                        Data:    data,
                    })
                },
                // No compensation — audit log is append-only
            },
        },
    }
}
```

---

## ۵. Scheduler — مدیریت مهلت‌های قانونی

### ۵.۱ معماری

```
┌─────────────────────────────────────────────┐
│              Scheduler Service               │
│                                              │
│  ┌──────────────┐    ┌────────────────────┐  │
│  │  Deadline     │    │  OTP Cleanup       │  │
│  │  Checker      │    │  (every 1 min)     │  │
│  │  (every 1 hr) │    │                    │  │
│  └──────┬───────┘    └────────────────────┘  │
│         │                                     │
│         ▼                                     │
│  ┌──────────────────────────────────────┐    │
│  │  scheduled_jobs table (PostgreSQL)    │    │
│  │  WHERE executed_at IS NULL            │    │
│  │  AND scheduled_at <= NOW()            │    │
│  │  ORDER BY scheduled_at                │    │
│  │  LIMIT 100 FOR UPDATE SKIP LOCKED     │    │
│  └──────────────────────────────────────┘    │
└─────────────────────────────────────────────┘
```

### ۵.۲ پیاده‌سازی Scheduler

```go
// scheduler/scheduler.go
package scheduler

type Scheduler struct {
    db       *sql.DB
    eventBus *nats.Conn
    jobs     map[string]JobHandler
}

type JobHandler func(ctx context.Context, job ScheduledJob) error

func New(db *sql.DB, eventBus *nats.Conn) *Scheduler {
    s := &Scheduler{
        db:       db,
        eventBus: eventBus,
        jobs:     make(map[string]JobHandler),
    }

    // Register job handlers
    s.Register("deadline_2years", handleDeadline2Years)
    s.Register("deadline_5months", handleDeadline5Months)
    s.Register("otp_cleanup", handleOTPCleanup)

    return s
}

// Poll checks for due jobs and executes them
func (s *Scheduler) Poll(ctx context.Context) error {
    tx, _ := s.db.BeginTx(ctx, nil)
    defer tx.Rollback()

    // FOR UPDATE SKIP LOCKED — prevents duplicate execution in multi-replica
    rows, err := tx.QueryContext(ctx, `
        SELECT id, job_type, target_case_id, scheduled_at
        FROM scheduled_jobs
        WHERE executed_at IS NULL
          AND scheduled_at <= NOW()
        ORDER BY scheduled_at
        LIMIT 100
        FOR UPDATE SKIP LOCKED
    `)
    if err != nil {
        return err
    }
    defer rows.Close()

    for rows.Next() {
        var job ScheduledJob
        rows.Scan(&job.ID, &job.JobType, &job.TargetCaseID, &job.ScheduledAt)

        handler, ok := s.jobs[job.JobType]
        if !ok {
            log.Warn("unknown job type", "type", job.JobType)
            s.markSkipped(ctx, tx, job.ID, "unknown job type")
            continue
        }

        if err := handler(ctx, job); err != nil {
            log.Error("job execution failed", "job", job.ID, "error", err)
            s.markError(ctx, tx, job.ID, err.Error())
        } else {
            s.markSuccess(ctx, tx, job.ID)
        }
    }

    return tx.Commit()
}

// Start runs the poll loop
func (s *Scheduler) Start(ctx context.Context, interval time.Duration) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            if err := s.Poll(ctx); err != nil {
                log.Error("scheduler poll failed", "error", err)
            }
        }
    }
}
```

### ۵.۳ Handlerهای کلیدی

```go
// scheduler/handlers.go

func handleDeadline2Years(ctx context.Context, job ScheduledJob) error {
    c, err := repo.GetCase(ctx, job.TargetCaseID)
    if err != nil {
        return err
    }

    // Check if already completed (race condition safety)
    if c.Status == "cert_completed" || c.Status == "expired" {
        return nil // skip — already done
    }

    // Transition: claim_completed → expired
    sm := GetCaseStateMachine()
    if err := sm.Transition(ctx, job.TargetCaseID, "expired"); err != nil {
        return err
    }

    // Send SMS
    notificationService.SendSMS(ctx, c.ApplicantID,
        "مهلت ۲ ساله درج گواهی اقدام برای پرونده شما به پایان رسیده است. پرونده منقضی شد.")

    return nil
}

func handleDeadline5Months(ctx context.Context, job ScheduledJob) error {
    c, err := repo.GetCase(ctx, job.TargetCaseID)
    if err != nil {
        return err
    }

    if !c.ApplicantDeceased {
        return nil // applicant not deceased — skip
    }
    if c.Status != "cert_in_progress" {
        return nil
    }

    // محاسبه deadline نهایی
    finalDeadline := maxDeadline(c.Deadline5Months, c.Deadline2Years)
    if time.Now().Before(*finalDeadline) {
        // هنوز مهلت دارد — reschedule
        scheduler.Reschedule(job, *finalDeadline)
        return nil
    }

    // مهلت تمام شده
    sm := GetCaseStateMachine()
    sm.Transition(ctx, job.TargetCaseID, "expired")
    notificationService.SendSMS(ctx, c.ApplicantID,
        "با توجه به فوت متقاضی و پایان مهلت قانونی، پرونده منقضی شد.")

    return nil
}

func handleOTPCleanup(ctx context.Context, job ScheduledJob) error {
    // حذف OTPهای منقضی شده
    _, err := db.ExecContext(ctx, `
        DELETE FROM otp_sessions
        WHERE expires_at < NOW() - INTERVAL '10 minutes'
    `)
    return err
}

func maxDeadline(a, b *time.Time) *time.Time {
    if a == nil {
        return b
    }
    if b == nil {
        return a
    }
    if a.After(*b) {
        return a
    }
    return b
}
```

---

## ۶. Event Subscription — واکنش به رویدادها

### ۶.۱ الگوی Consumer

```go
// events/consumer.go
package events

type Consumer struct {
    conn     *nats.Conn
    handlers map[string]MsgHandler
}

func (c *Consumer) Subscribe(subject string, handler MsgHandler) {
    c.conn.Subscribe(subject, func(msg *nats.Msg) {
        var event StateChangeEvent
        json.Unmarshal(msg.Data, &event)

        if err := handler(context.Background(), event); err != nil {
            log.Error("event handler failed",
                "subject", msg.Subject,
                "event", event,
                "error", err)
            // NAK with delay → retry later
            msg.NakWithDelay(5 * time.Second)
        } else {
            msg.Ack()
        }
    })
}
```

### ۶.۲ واکنش‌های کلیدی

```go
// events/handlers.go

func SetupEventHandlers(consumer *Consumer, services Services) {

    // وقتی سازمان نقشه را تایید کرد → تغییر وضعیت case
    consumer.Subscribe("map.{id}.org.approved", func(ctx context.Context, e StateChangeEvent) error {
        mapSvc, _ := repo.GetMapService(ctx, e.ResourceID)
        sm := GetCaseStateMachine()
        return sm.Transition(ctx, mapSvc.CaseID, "map_completed")
    })

    // وقتی ادعا تایید شد → تنظیم deadline ۲ ساله
    consumer.Subscribe("claim.{id}.org.approved", func(ctx context.Context, e StateChangeEvent) error {
        claimSvc, _ := repo.GetClaimService(ctx, e.ResourceID)
        sm := GetCaseStateMachine()
        return sm.Transition(ctx, claimSvc.CaseID, "claim_completed")
    })

    // وقتی گواهی اقدام تایید شد → تکمیل پرونده
    consumer.Subscribe("cert.{id}.org.approved", func(ctx context.Context, e StateChangeEvent) error {
        certSvc, _ := repo.GetCertService(ctx, e.ResourceID)
        sm := GetCaseStateMachine()
        return sm.Transition(ctx, certSvc.CaseID, "cert_completed")
    })

    // وقتی پرداخت کامل شد → اگر advance payment بود، به مرحله بعد برو
    consumer.Subscribe("payment.{id}.completed", func(ctx context.Context, e StateChangeEvent) error {
        payment, _ := repo.GetPayment(ctx, e.ResourceID)
        if payment.PaymentType == "advance" {
            // ارسال نوتیفیکیشن به متقاضی
            notificationService.Notify(ctx, payment.CaseID, "پرداخت با موفقیت انجام شد")
        }
    })
}
```

---

## ۷. خطاها و بازیابی (Error Recovery)

### ۷.۱ انواع خطا

| نوع خطا | رفتار |
|---|---|
| **Guard Failed** | برگرداندن خطای واضح به کاربر (فارسی) — بدون تغییر وضعیت |
| **Effect Failed** | خطا لاگ می‌شود — وضعیت تغییر کرده (committed). ادمین alert می‌شود. |
| **Saga Step Failed** | Retry با backoff → compensation در صورت شکست نهایی |
| **Organization API Down** | Circuit Breaker باز می‌شود → درخواست در DLQ می‌رود → ادمین دستی بررسی می‌کند |
| **Database Connection Lost** | K8s readiness probe fail → pod از service خارج می‌شود → replicaهای سالم پاسخ می‌دهند |

### ۷.۲ Dead Letter Queue Handler

```go
// integration/dlq_handler.go

func (h *DLQHandler) Process(ctx context.Context) {
    for {
        msg, err := h.dlq.Fetch(ctx, 10*time.Second)
        if err != nil {
            continue
        }

        // Log to admin dashboard
        h.notifyAdmin(AdminAlert{
            Severity: "warning",
            Title:    "درخواست ناموفق در صف انتظار",
            Body:     fmt.Sprintf("درخواست %s پس از %d بار تلاش ناموفق ماند", msg.ID, msg.RetryCount),
            Actions: []AdminAction{
                {Label: "تلاش مجدد", Endpoint: fmt.Sprintf("/api/integration/failed-requests/%s/retry", msg.ID)},
                {Label: "رد درخواست", Endpoint: fmt.Sprintf("/api/integration/failed-requests/%s/discard", msg.ID)},
            },
        })
    }
}
```

---

## ۸. جمع‌بندی — جریان یک پرونده کامل

```
زمان     رویداد                          وضعیت Case
─────    ──────                          ──────────
T+0      متقاضی ثبت‌نام + احراز هویت      (new user)
T+1      ایجاد پرونده + آدرس ملک          DRAFT
T+2      شروع سرویس نقشه                  MAP_IN_PROGRESS
         ├─ تخصیص خودکار کارشناسان
         ├─ کارشناس حقوقی: احراز نمایندگی
         └─ OTP رضایت
T+5      کارشناس نقشه‌بردار: عملیات میدانی
         ├─ عکس‌برداری ۴ ضلع
         ├─ ترسیم نقشه با AutoCAD
         └─ تکمیل جدول توصیفی (مانا)
T+8      ارسال به سازمان                 (MapService: submitted)
T+9      سازمان: تایید + کد رهگیری نقشه   MAP_COMPLETED
T+10     شروع سرویس درج ادعا              CLAIM_IN_PROGRESS
         ├─ اعتبارسنجی کد رهگیری نقشه
         ├─ پیامک هشدار ادعای واهی
         ├─ OTP رضایت
         ├─ ثبت پلاک + نوع ادعا
         └─ آپلود مستندات
T+14     ارسال به سازمان                 (ClaimService: submitted)
T+15     سازمان: تایید + کد رهگیری ادعا   CLAIM_COMPLETED
         └─ deadline_2years = T+15 + 2y
T+16     شروع سرویس گواهی اقدام           CERT_IN_PROGRESS
         ├─ اعتبارسنجی مهلت ۲ ساله
         ├─ OTP رضایت
         ├─ ثبت مرجع + نوع اقدام
         └─ آپلود تصویر گواهی
T+18     ارسال به سازمان                 (CertService: submitted)
T+19     سازمان: تایید + کد رهگیری نهایی   CERT_COMPLETED ✅
         └─ پیامک: «فرآیند تکمیل شد»
```

---

> **فایل مرتبط:** Schema دیتابیس در [database-schema.sql](database-schema.sql)، قرارداد API در [api-contract.yaml](api-contract.yaml)
