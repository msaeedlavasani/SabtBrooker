# ── Build Stage ──────────────────────────────────────────
FROM node:20-alpine AS builder

# اضافه کردن آرگومان بیلد برای آدرس API (حیاتی برای Next.js)
ARG NEXT_PUBLIC_API_URL
ENV NEXT_PUBLIC_API_URL=$NEXT_PUBLIC_API_URL

WORKDIR /app

# Cache dependencies
COPY package.json package-lock.json* ./
RUN npm ci --legacy-peer-deps

# Build
COPY . .
RUN npm run build

# ── Runtime Stage ────────────────────────────────────────
FROM node:20-alpine AS runner

RUN apk add --no-cache tzdata curl
ENV TZ=Asia/Tehran

WORKDIR /app

# Create non-root user
RUN adduser -D -g '' appuser

COPY --from=builder /app/public ./public
COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./.next/static

EXPOSE 3000

HEALTHCHECK --interval=15s --timeout=3s --retries=3 \
  CMD wget -q --spider http://localhost:3000 || exit 1

USER appuser
ENV PORT=3000
ENV HOSTNAME="0.0.0.0"

CMD ["node", "server.js"]
