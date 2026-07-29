-- Seed Data — System Configs

INSERT INTO system_configs (config_key, config_value, description) VALUES
    ('otp.expiry_seconds', '120', 'مدت اعتبار OTP به ثانیه'),
    ('otp.max_attempts', '3', 'حداکثر تلاش ناموفق OTP'),
    ('otp.rate_limit.per_10min', '5', 'حداکثر درخواست OTP در ۱۰ دقیقه به ازای هر شماره'),
    ('security.max_failed_login', '5', 'حداکثر تلاش ناموفق ورود پیش از قفل'),
    ('security.lock_duration_minutes', '30', 'مدت قفل حساب پس از تلاش ناموفق'),
    ('integration.retry.max_attempts', '3', 'حداکثر تلاش مجدد ارسال به سازمان'),
    ('integration.retry.backoff_seconds', '[1,2,4,8,16]', 'فواصل تلاش مجدد به ثانیه'),
    ('deadline.check_interval_hours', '24', 'فاصله زمانی بررسی مهلت‌ها'),
    ('storage.presigned_url_expiry_seconds', '300', 'مدت اعتبار Pre-signed URL فایل‌ها');

-- Default tariffs (sample — adjust based on actual organization rates)
INSERT INTO tariffs (service_type, max_amount, non_refundable, effective_from, version, description) VALUES
    ('map', 5000000, 500000, CURRENT_DATE, '1.0.0', 'تعرفه تهیه نقشه ثبتی'),
    ('claim', 3000000, 300000, CURRENT_DATE, '1.0.0', 'تعرفه درج ادعا'),
    ('cert', 2000000, 200000, CURRENT_DATE, '1.0.0', 'تعرفه درج گواهی اقدام'),
    ('legal_advice', 1000000, 0, CURRENT_DATE, '1.0.0', 'تعرفه راهنمایی ثبتی');
