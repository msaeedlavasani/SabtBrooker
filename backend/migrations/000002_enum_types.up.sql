-- All ENUM types

CREATE TYPE case_status AS ENUM (
    'draft',
    'map_in_progress',
    'map_completed',
    'claim_in_progress',
    'claim_completed',
    'cert_in_progress',
    'cert_completed',
    'expired',
    'rejected',
    'cancelled'
);

CREATE TYPE map_service_status AS ENUM (
    'pending_expert_assignment',
    'expert_assigned',
    'fieldwork_in_progress',
    'fieldwork_done',
    'submitted_to_org',
    'approved',
    'rejected'
);

CREATE TYPE claim_service_status AS ENUM (
    'pending_expert_assignment',
    'expert_assigned',
    'documents_verified',
    'submitted_to_org',
    'approved',
    'rejected'
);

CREATE TYPE cert_service_status AS ENUM (
    'pending_data',
    'submitted_to_org',
    'approved',
    'rejected'
);

CREATE TYPE user_role AS ENUM (
    'applicant',
    'legal_expert',
    'survey_expert',
    'admin',
    'auditor'
);

CREATE TYPE applicant_capacity AS ENUM (
    'principal',
    'legal_rep_natural',
    'legal_rep_legal'
);

CREATE TYPE claim_type AS ENUM (
    'ownership',
    'easement',
    'usufruct',
    'benefits_ownership'
);

CREATE TYPE ownership_type AS ENUM (
    'land',
    'building',
    'land_and_building'
);

CREATE TYPE document_type AS ENUM (
    'sales_agreement',
    'settlement',
    'promissory_note',
    'partition',
    'gift',
    'court_ruling',
    'cultivation_right',
    'witness_testimony',
    'inheritance_cert',
    'other'
);

CREATE TYPE action_reference AS ENUM (
    'registration_org',
    'judiciary',
    'organization_board',
    'determination_board'
);

CREATE TYPE action_type AS ENUM (
    'lawsuit',
    'initial_registration',
    'organization_request',
    'determination_request'
);

CREATE TYPE payment_status AS ENUM (
    'pending',
    'paid',
    'partially_refunded',
    'fully_refunded',
    'failed'
);

CREATE TYPE payment_type AS ENUM (
    'advance',
    'final_settlement',
    'refund'
);

CREATE TYPE notification_channel AS ENUM (
    'sms',
    'in_app',
    'email'
);

CREATE TYPE notification_status AS ENUM (
    'pending',
    'sent',
    'delivered',
    'failed'
);

CREATE TYPE audit_event_type AS ENUM (
    'case.created',
    'case.status_changed',
    'case.expired',
    'case.cancelled',
    'map.expert_assigned',
    'map.fieldwork_started',
    'map.fieldwork_completed',
    'map.submitted_to_org',
    'map.approved',
    'map.rejected',
    'claim.expert_assigned',
    'claim.documents_verified',
    'claim.submitted_to_org',
    'claim.approved',
    'claim.rejected',
    'cert.submitted_to_org',
    'cert.approved',
    'cert.rejected',
    'payment.created',
    'payment.paid',
    'payment.refunded',
    'expert.created',
    'expert.deactivated',
    'user.login',
    'user.login_failed',
    'otp.sent',
    'otp.verified',
    'otp.failed',
    'ai_advice.requested',
    'document.uploaded',
    'document.verified'
);
