-- Map Service + Photos

CREATE TABLE map_services (
    id                      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    case_id                 UUID              NOT NULL REFERENCES cases(id) ON DELETE CASCADE,
    status                  map_service_status NOT NULL DEFAULT 'pending_expert_assignment',

    property_type           VARCHAR(50),
    approx_area_sqm         DECIMAL(10,2),
    land_use                VARCHAR(50),
    ownership_type          VARCHAR(50),
    has_building            BOOLEAN       DEFAULT FALSE,
    annex_count             INT           DEFAULT 0,

    geo_latitude            DECIMAL(10,7),
    geo_longitude           DECIMAL(10,7),
    geo_location            GEOGRAPHY(POINT, 4326),

    grant_access_to_others  BOOLEAN       DEFAULT FALSE,
    access_granted_to       JSONB,

    consent_otp_id          UUID          REFERENCES otp_sessions(id),
    consent_granted_at      TIMESTAMPTZ,

    fieldwork_started_at    TIMESTAMPTZ,
    fieldwork_completed_at  TIMESTAMPTZ,

    map_file_path           VARCHAR(500),
    map_format              VARCHAR(20),
    descriptive_table       JSONB,
    submitted_to_org_at     TIMESTAMPTZ,

    tracking_code           VARCHAR(50),
    org_response_at         TIMESTAMPTZ,
    org_rejection_reason    TEXT,
    org_response_raw        JSONB,

    created_at              TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_map_services_case ON map_services (case_id);
CREATE INDEX idx_map_services_status ON map_services (status);
CREATE INDEX idx_map_services_tracking ON map_services (tracking_code);
CREATE INDEX idx_map_services_geo ON map_services USING GIST (geo_location);

CREATE TABLE map_photos (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    map_service_id  UUID         NOT NULL REFERENCES map_services(id) ON DELETE CASCADE,
    file_path       VARCHAR(500) NOT NULL,
    side            VARCHAR(20)  NOT NULL CHECK (side IN ('north', 'south', 'east', 'west')),
    photo_latitude  DECIMAL(10,7),
    photo_longitude DECIMAL(10,7),
    photo_location  GEOGRAPHY(POINT, 4326),
    photo_taken_at  TIMESTAMPTZ,
    exif_valid      BOOLEAN      DEFAULT FALSE,
    exif_validation_note TEXT,
    file_size       BIGINT,
    checksum_sha256 VARCHAR(64),
    uploaded_by     UUID         REFERENCES users(id),
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_map_photos_service ON map_photos (map_service_id);
CREATE INDEX idx_map_photos_geo ON map_photos USING GIST (photo_location);
