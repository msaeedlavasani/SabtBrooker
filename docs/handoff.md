# SabtBrooker — Project Handoff

> Last Updated: 2026-07-31  
> Status: Production Demo Ready  
> Repository Status: Stable  
> Deployment Status: Live

---

# Overview

SabtBrooker is a land registration platform built with a modern microservice-oriented architecture.

## Tech Stack

- Frontend: Next.js (PWA)
- Backend: Go (Echo Framework)
- Database: PostgreSQL + PostGIS
- Cache: Redis
- Object Storage: MinIO
- Event Bus: NATS JetStream
- Reverse Proxy: Nginx
- Deployment: Docker Compose

---

# Production Environment

### Domain

https://sabt.saeedlavasani.ir

### Server

Ubuntu VPS

### Project Location

```
/opt/sabtbrooker
```

---

# Running Services

| Service | Status |
|----------|--------|
| Frontend | ✅ Running |
| Backend | ✅ Running |
| PostgreSQL | ✅ Running |
| Redis | ✅ Running |
| MinIO | ✅ Running |
| NATS | ✅ Running |
| Nginx | ✅ Running |

All services are deployed using Docker Compose.

---

# Deployment Notes

The project is deployed using Docker Compose.

Main commands:

```bash
docker compose down
docker compose build --no-cache
docker compose up -d
```

Frontend only:

```bash
docker compose build --no-cache frontend
docker compose up -d frontend
```

Backend only:

```bash
docker compose build backend
docker compose up -d backend
```

---

# Important Production Fixes

## 1. Frontend API URL

One of the major deployment issues was that the frontend was still calling

```
http://localhost:8081
```

instead of

```
https://sabt.saeedlavasani.ir/api
```

### Cause

Next.js embeds every `NEXT_PUBLIC_*` variable during **build time**.

Changing `.env.production` alone does **not** update the frontend bundle.

### Final Solution

The Dockerfile now passes the API URL as a build argument.

Example:

```dockerfile
ARG NEXT_PUBLIC_API_URL
ENV NEXT_PUBLIC_API_URL=$NEXT_PUBLIC_API_URL
```

docker-compose:

```yaml
build:
  context: ./frontend
  dockerfile: Dockerfile
  args:
    NEXT_PUBLIC_API_URL: ${NEXT_PUBLIC_API_URL}
```

Then rebuild without cache.

Verification:

```bash
docker exec -it sabtbrooker-frontend sh

grep -R "localhost:8081" /app/.next
```

Expected:

```
NOT_FOUND
```

---

## 2. Environment Variables

The frontend container now loads

```
.env.production
```

using

```yaml
env_file:
  - .env.production
```

Current production value:

```
NEXT_PUBLIC_API_URL=https://sabt.saeedlavasani.ir/api
```

---

## 3. Reverse Proxy

Nginx routes

```
/
```

to

```
Frontend
```

and

```
/api
```

to

```
Backend
```

SSL is configured and active.

---

# Authentication

Current authentication is intentionally simplified for demonstration purposes.

Current flow:

1. User enters mobile number.
2. Frontend directly calls:

```
POST /v1/auth/otp/verify
```

using

```
OTP = 1234
```

3. Backend skips OTP validation.
4. JWT Access Token and Refresh Token are returned.

No SMS verification is currently required.

---

# Current Functional State

Implemented:

- JWT Authentication
- Automatic User Creation
- OTP Demo Mode
- Case Workflow Engine
- Map Service
- Claim Service
- Certificate Service
- Notification Infrastructure
- File Upload Infrastructure
- Docker Deployment
- Production Deployment

---

# Current Limitations

The backend contains infrastructure for:

- Admin
- Auditor
- Legal Expert
- Survey Expert

However, the frontend currently contains only the public user portal.

There is currently **no**:

- Admin Dashboard
- Expert Dashboard
- Reviewer Panel
- User Management UI

These APIs already exist (or are partially implemented) but have not yet been connected to frontend pages.

---

# Known Technical Decisions

## Demo Authentication

OTP verification is bypassed intentionally.

Reason:

To simplify client demonstrations and avoid dependency on SMS gateways.

To restore production authentication:

- Restore OTP validation inside:

```
backend/internal/handler/auth_handler.go
```

- Restore the two-step login UI inside:

```
frontend/src/app/page.tsx
```

---

# Useful Commands

## Open frontend container

```bash
docker exec -it sabtbrooker-frontend sh
```

## Open backend container

```bash
docker exec -it sabtbrooker-backend sh
```

## Backend logs

```bash
docker compose logs -f backend
```

## Frontend logs

```bash
docker compose logs -f frontend
```

## Check API URL

```bash
printenv | grep NEXT_PUBLIC_API_URL
```

## Verify frontend bundle

```bash
grep -R "localhost:8081" /app/.next
```

---

# Suggested Next Steps

Highest priority:

1. Restore real OTP authentication.
2. Build Admin Dashboard.
3. Build Expert Dashboard.
4. Implement Case Review UI.
5. Connect expert APIs to frontend.
6. Add role-based routing.
7. Add CI/CD pipeline.
8. Perform end-to-end workflow testing.

---

# Repository Status

Infrastructure: ✅ Complete

Dockerization: ✅ Stable

Production Deployment: ✅ Complete

Authentication: ⚠️ Demo Mode

Backend APIs: ✅ Mostly Implemented

Frontend User Portal: ✅ Operational

Admin / Expert Interfaces: 🚧 Pending

Overall project status:

The platform is successfully deployed and stable for demonstration purposes. Remaining work is primarily focused on administrative interfaces, expert workflows, and restoring production-grade OTP authentication.
