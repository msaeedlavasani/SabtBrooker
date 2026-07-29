DELETE FROM tariffs WHERE version = '1.0.0';
DELETE FROM system_configs WHERE config_key IN (
    'otp.expiry_seconds', 'otp.max_attempts', 'otp.rate_limit.per_10min',
    'security.max_failed_login', 'security.lock_duration_minutes',
    'integration.retry.max_attempts', 'integration.retry.backoff_seconds',
    'deadline.check_interval_hours', 'storage.presigned_url_expiry_seconds'
);
