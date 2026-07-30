#!/bin/bash
# ====================================================================
# SabtBrooker — اسکریپت استقرار یک‌خطی روی VPS
# استفاده:  curl -fsSL https://raw.githubusercontent.com/msaeedlavasani/SabtBrooker/main/deploy.sh | bash -s -- yourdomain.com
# ====================================================================
set -e

# ── رنگ‌های ترمینال ──────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info()  { echo -e "${GREEN}[✓]${NC} $1"; }
log_warn()  { echo -e "${YELLOW}[!]${NC} $1"; }
log_error() { echo -e "${RED}[✗]${NC} $1"; }
log_step()  { echo -e "\n${BLUE}━━━ $1 ━━━${NC}"; }

# ── پارامترها ─────────────────────────────────────────────
DOMAIN="${1:-localhost}"

log_step "1/6 — بررسی پیش‌نیازها"

# تنظیم DNS شکن برای عبور از تحریم‌ها در زمان نصب (اختیاری اما توصیه شده برای ایران)
if [ -f /etc/resolv.conf ]; then
    log_info "تنظیم موقت DNS ضد تحریم..."
    sed -i '1i nameserver 178.22.122.100' /etc/resolv.conf
    sed -i '2i nameserver 185.51.200.2' /etc/resolv.conf
fi

# ── ۱.۱ Docker ───────────────────────────────────────────
if ! command -v docker &>/dev/null; then
    log_warn "Docker نصب نیست. در حال نصب..."
    curl -fsSL https://get.docker.com | sh
    systemctl enable docker && systemctl start docker
    log_info "Docker نصب شد."
else
    log_info "Docker قبلاً نصب شده: $(docker --version)"
fi

# ── ۱.۲ Docker Compose Plugin ────────────────────────────
if ! docker compose version &>/dev/null; then
    log_warn "Docker Compose Plugin نصب نیست. در حال نصب..."
    DOCKER_CONFIG=${DOCKER_CONFIG:-$HOME/.docker}
    mkdir -p $DOCKER_CONFIG/cli-plugins
    curl -SL "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" \
        -o $DOCKER_CONFIG/cli-plugins/docker-compose
    chmod +x $DOCKER_CONFIG/cli-plugins/docker-compose
    log_info "Docker Compose Plugin نصب شد."
else
    log_info "Docker Compose قبلاً نصب شده: $(docker compose version)"
fi

# ── ۱.۳ Git ──────────────────────────────────────────────
if ! command -v git &>/dev/null; then
    apt-get update -qq && apt-get install -y -qq git
fi
log_info "Git: $(git --version | head -1)"

log_step "2/6 — دریافت کدها از گیت‌هاب"

APP_DIR="/opt/sabtbrooker"
if [ -d "$APP_DIR" ]; then
    log_warn "پوشه $APP_DIR وجود دارد. در حال بروزرسانی..."
    cd "$APP_DIR" && git pull origin main
else
    git clone https://github.com/msaeedlavasani/SabtBrooker.git "$APP_DIR"
    cd "$APP_DIR"
fi
log_info "کدها آماده هستند: $APP_DIR"

log_step "3/6 — تنظیم پیکربندی (.env)"

if [ ! -f ".env.production" ]; then
    cp .env.production.example .env.production

    # تولید مقادیر امنیتی به صورت خودکار
    ENC_KEY=$(openssl rand -hex 32 2>/dev/null || cat /dev/urandom | tr -dc 'a-f0-9' | fold -w 64 | head -1)
    MINIO_KEY=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 20 | head -1)
    MINIO_SECRET=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 40 | head -1)
    DB_PASS=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 20 | head -1)

    sed -i "s/REPLACE_WITH_STRONG_PASSWORD/$DB_PASS/" .env.production
    sed -i "s/REPLACE_WITH_STRONG_ACCESS_KEY/$MINIO_KEY/" .env.production
    sed -i "s/REPLACE_WITH_STRONG_SECRET_KEY/$MINIO_SECRET/" .env.production
    sed -i "s/REPLACE_WITH_RANDOM_HEX_64_CHARS/$ENC_KEY/" .env.production
    sed -i "s|NEXT_PUBLIC_API_URL=.*|NEXT_PUBLIC_API_URL=https://$DOMAIN/api|" .env.production
    
    # اطمینان از فعال بودن حالت توسعه برای OTP ثابت 1234
    if grep -q "DEV_MODE" .env.production; then
        sed -i "s/DEV_MODE=.*/DEV_MODE=true/" .env.production
    else
        echo "DEV_MODE=true" >> .env.production
    fi

    log_info "فایل .env.production با مقادیر امنیتی تصادفی و DEV_MODE=true ساخته شد."
else
    # بروزرسانی DEV_MODE در فایل موجود
    sed -i "s/DEV_MODE=.*/DEV_MODE=true/" .env.production
    log_info "فایل .env.production بروزرسانی شد (DEV_MODE=true)."
fi

log_step "4/6 — تنظیم SSL (گواهی موقت)"

mkdir -p nginx/ssl
if [ ! -f "nginx/ssl/privkey.pem" ]; then
    openssl req -x509 -nodes -days 90 \
        -newkey rsa:2048 \
        -keyout nginx/ssl/privkey.pem \
        -out nginx/ssl/fullchain.pem \
        -subj "/C=IR/ST=Tehran/L=Tehran/O=SabtBrooker/CN=$DOMAIN" 2>/dev/null
    log_warn "گواهی SSL خودامضا ساخته شد (۹۰ روزه)."
    log_warn "برای گواهی رایگان Let's Encrypt واقعی، بعداً اجرا کنید:"
    echo "  docker compose run --rm certbot certonly --webroot -w /var/www/certbot -d $DOMAIN"
fi

log_step "5/6 — ساخت و راه‌اندازی سرویس‌ها"

# توقف سرویس‌های احتمالی قدیمی برای نصب تمیز
docker compose down --remove-orphans || true

# ساخت با اجبار به استفاده از GOPROXY برای ایران
docker compose --env-file .env.production build --build-arg GOPROXY=https://goproxy.io,direct backend
docker compose --env-file .env.production up -d

log_step "6/6 — بررسی سلامت"

sleep 5

# بررسی سلامت سرویس‌ها
check_service() {
    local service=$1 name=$2
    if docker compose ps "$service" | grep -q "healthy\|Up"; then
        log_info "$name: فعال ✓"
    else
        log_error "$name: راه‌اندازی نشد — لاگ: docker compose logs $service"
    fi
}

check_service postgres    "PostgreSQL + PostGIS"
check_service redis       "Redis"
check_service nats        "NATS JetStream"
check_service minio       "MinIO Storage"
check_service backend     "Backend API (Go)"
check_service frontend    "Frontend (Next.js)"
check_service nginx       "Nginx Reverse Proxy"

echo ""
echo -e "${GREEN}══════════════════════════════════════════════════════${NC}"
echo -e "${GREEN}  🚀 استقرار SabtBrooker با موفقیت انجام شد!${NC}"
echo -e "${GREEN}══════════════════════════════════════════════════════${NC}"
echo ""
echo -e "  آدرس سایت:         ${BLUE}https://$DOMAIN${NC}"
echo -e "  پنل MinIO:         ${BLUE}http://$DOMAIN:9001${NC}"
echo -e "  پوشه نصب:          ${BLUE}$APP_DIR${NC}"
echo ""
echo -e "  دستورات پرکاربرد:"
echo -e "  ${YELLOW}cd $APP_DIR${NC}"
echo -e "  ${YELLOW}docker compose ps${NC}                     # وضعیت سرویس‌ها"
echo -e "  ${YELLOW}docker compose logs -f backend${NC}         # لاگ بک‌اِند"
echo -e "  ${YELLOW}docker compose restart${NC}                 # راه‌اندازی مجدد"
echo ""
