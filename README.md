# SIP Routing Flow

**A cloud-deployed, database-driven SIP routing decision API with a separate Python analytics service.**

SIP Routing Flow is a backend engineering project inspired by telecommunications call-routing systems. A Go REST API selects a configured SIP destination based on a requested region, applies a default fallback when needed, and records routing decisions in PostgreSQL. A separate Python FastAPI service queries those records to provide operational analytics.

Both applications are containerized with Docker and have been deployed together as an AWS ECS Fargate task. Their HTTP health endpoints, analytics/database connectivity, and ECS container health checks have been verified in AWS.

> **Scope:** This is a routing **decision** service, not a SIP proxy or PBX. It returns SIP destination URIs; it does not exchange SIP signaling, establish voice sessions, or confirm call completion.

## At a glance

| Component | Technology | Responsibility |
| --- | --- | --- |
| Routing API | Go, `net/http`, `pgx/v5` | Validate requests, select SIP destinations, persist decisions |
| Analytics API | Python, FastAPI | Summarize routing history from PostgreSQL |
| Database | Neon PostgreSQL | Store routing configuration and call records |
| Local environment | Docker, Docker Compose | Run both services together |
| Cloud deployment | AWS ECS Fargate (Linux/ARM64), Amazon ECR | Run container images in AWS |
| Operations | AWS Secrets Manager, CloudWatch Logs, IAM, ECS Exec | Manage credentials, inspect logs, securely troubleshoot |

## Features

- **Database-driven routing:** Select SIP destinations from PostgreSQL rather than hardcoded configuration.
- **Region matching and fallback:** Match regions case-insensitively and use an enabled `DEFAULT` route when an exact match is unavailable.
- **Configurable route availability:** Enable or disable routing entries in the database.
- **Persistent routing history:** Record call identifiers, caller and destination, region, selected SIP URI, status, and timestamp.
- **Analytics service:** Retrieve aggregate summaries, region-level statistics, and recent routing activity through a separate Python API.
- **Operational checks:** Expose HTTP health endpoints for both services; configure ECS container health checks and CloudWatch logging.
- **Cloud secrets management:** Inject the PostgreSQL connection string into Fargate containers using AWS Secrets Manager rather than embedding it in container images.
- **Automated tests:** Include Go and Python tests for application behavior.

## Architecture

```text
                        API clients / demo requests
                                  |
                 +----------------+----------------+
                 |                                 |
          POST /route                        GET /analytics/*
          GET /health                         GET /health
                 |                                 |
         +-------v--------+                +-------v---------+
         | Go Routing API |                | Python FastAPI  |
         |     :8080      |                | Analytics :8000 |
         +-------+--------+                +--------+--------+
                 |                                  |
                 +---------------+------------------+
                                 |
                         TLS PostgreSQL
                                 |
                       +---------v---------+
                       |  Neon PostgreSQL  |
                       |    sip_routes     |
                       |    call_records   |
                       +-------------------+

      AWS ECS Fargate task: both application containers (Linux/ARM64)
      Images: Amazon ECR | Secrets: AWS Secrets Manager
      Logs: CloudWatch   | Diagnostics: ECS Exec
```

The two APIs are independently implemented services sharing the same database. In AWS they run as two containers within one Fargate task. The deployment is currently intended for controlled demonstrations and troubleshooting, **not as a publicly exposed production service**.

### Routing flow

1. A client submits a JSON request to `POST /route`.
2. The Go API validates the request and looks up an enabled region-specific route.
3. If no matching region route is found, it attempts the configured `DEFAULT` route.
4. The selected route and request metadata are persisted to `call_records`.
5. The API returns the routing decision as JSON.
6. The Python analytics API reads persisted records to report routing activity.

## API reference

### Go routing API — port 8080

| Method | Endpoint | Purpose |
| --- | --- | --- |
| GET | `/health` | HTTP service health |
| POST | `/route` | Select a SIP destination and record the decision |

**Example routing request**

```bash
curl -X POST http://localhost:8080/route \
  -H 'Content-Type: application/json' \
  -d '{
    "call_id": "CALL-1001",
    "caller": "+14045550123",
    "destination": "+18005550100",
    "region": "US"
  }'
```

**Example response** (depends on configured routes)

```json
{
  "call_id": "CALL-1001",
  "route": "sip:us-gateway@pbx.internal",
  "status": "ROUTE_FOUND"
}
```

The Go API distinguishes invalid requests, unavailable routes, and database/persistence failures with appropriate HTTP errors. Its `/health` endpoint checks HTTP availability, not end-to-end PostgreSQL health.

### Python analytics API — port 8000

| Method | Endpoint | Purpose |
| --- | --- | --- |
| GET | `/health` | HTTP service health |
| GET | `/analytics/summary` | Aggregate routing activity |
| GET | `/analytics/regions` | Routing activity grouped by region |
| GET | `/analytics/recent` | Recent routing decisions |

```bash
curl http://localhost:8000/health
curl http://localhost:8000/analytics/summary
curl http://localhost:8000/analytics/regions
curl http://localhost:8000/analytics/recent
```

Analytics responses reflect the records currently stored in PostgreSQL.

## Database design

Two tables support routing and reporting:

- **`sip_routes`:** `id`, `region`, `sip_uri`, `priority`, `enabled`, `created_at`. Region is currently unique, so the `priority` field is reserved for potential multi-gateway support.
- **`call_records`:** `id`, `call_id`, `caller`, `destination`, `region`, `selected_route`, `routing_status`, `created_at`.

**Illustrative route configuration:**

```sql
INSERT INTO sip_routes (region, sip_uri, enabled)
VALUES
  ('US', 'sip:us-gateway@pbx.internal', TRUE),
  ('EU', 'sip:eu-gateway@pbx.internal', TRUE),
  ('DEFAULT', 'sip:default@pbx.internal', TRUE)
ON CONFLICT (region) DO NOTHING;
```

The `.internal` SIP URIs are demonstration values, not live gateways.

A route can be updated without rebuilding the API:

```sql
UPDATE sip_routes
SET sip_uri = 'sip:us-backup@pbx.internal'
WHERE region = 'US';
```

## Run locally

### Prerequisites

- Docker with Docker Compose
- A PostgreSQL database with the `sip_routes` and `call_records` tables configured
- A valid database connection string

### Docker Compose (both APIs)

```bash
git clone https://github.com/Shannon-Reed/SIP-Routing-Flow.git
cd SIP-Routing-Flow
```

Create a local `.env` file (do not commit it):

```dotenv
DATABASE_URL=postgresql://USERNAME:PASSWORD@HOST/DATABASE?sslmode=require
```

Start the services:

```bash
docker compose up --build -d
docker compose ps
```

Verify both APIs:

```bash
curl http://localhost:8080/health
curl http://localhost:8000/health
curl http://localhost:8000/analytics/summary
```

Stop the local containers:

```bash
docker compose down
```

### Run the Go API without Docker

With Go installed and `DATABASE_URL` configured:

```bash
go mod download
go run .
```

### Tests

```bash
go test ./... -v
```

The repository also includes Python tests; run them from the analytics environment using its configured test dependencies.

## AWS deployment (implemented and verified)

The current AWS deployment uses:

- **Amazon ECR:** Separate repositories for the Go routing and Python analytics container images.
- **Amazon ECS Fargate:** One Linux/ARM64 task definition with two essential application containers.
- **AWS Secrets Manager:** Stores the Neon `DATABASE_URL`, injected into each container at startup.
- **IAM:** Separate task execution and task roles, with scoped permissions for image pulls, secrets, logging, and ECS Exec.
- **Amazon CloudWatch Logs:** Dedicated log groups for both services.
- **ECS container health checks:** Each container periodically requests its local `/health` endpoint; both have reported `HEALTHY`.
- **ECS Exec:** Used to securely access the containers and verify API endpoints and database-backed analytics without enabling public inbound access.

**Deployment status:** Images have been pushed to ECR, both applications have started on Fargate, logs are available in CloudWatch, health checks report healthy, and the API endpoints have been tested from within the ECS task.

This is currently a **manually launched standalone ECS task**. It does not yet use an ECS service, public HTTPS load balancer, or automated GitHub Actions deployment. A standalone task is not automatically replaced when it fails. The task can be stopped between demonstrations to limit compute costs.

## Engineering decisions

- **Go for the routing path:** A small HTTP service with explicit validation and database access.
- **Python for analytics:** A separate FastAPI application for reporting-oriented endpoints.
- **Database-backed configuration:** Routing destinations can change independently of application builds.
- **Parameterized queries and connection pooling:** Support safer, reusable PostgreSQL access.
- **Separation of application concerns:** Routing logic, database access, and analytics live in distinct components.
- **Managed secrets:** Database credentials are supplied at runtime, not committed to Git or baked into images.
- **Private-first deployment:** API validation uses ECS Exec rather than exposing unauthenticated endpoints publicly.

## Limitations and potential enhancements

The project does **not** currently include:

- Live SIP signaling (`INVITE`, `REGISTER`, etc.), RTP/media, or call establishment.
- Real-time SIP gateway probing or availability-based failover.
- Multiple gateways per region with priority-based selection.
- Public API authentication, rate limiting, or a production HTTPS endpoint.
- An ECS service for automatic task replacement or a CI/CD deployment pipeline.
- Comprehensive integration testing against all failure conditions.

Possible future improvements include GitHub Actions CI/CD, ECS service management, HTTPS ingress, richer routing policies, and gateway availability monitoring.

## Project purpose

This project brings together telecommunications routing concepts and modern backend engineering: Go service development, Python APIs, relational data modeling, containerization, cloud deployment, access control, and operational diagnostics.

**Author:** Shannon Reed  
**Repository:** https://github.com/Shannon-Reed/SIP-Routing-Flow
