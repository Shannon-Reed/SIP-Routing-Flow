# SIP Routing Flow

**A Go-based telecommunications routing API with dynamic SIP gateway selection, cloud-hosted PostgreSQL configuration, and persistent routing history.**

SIP Routing Flow is a backend development project that simulates the routing decision layer of a Voice over IP (VoIP) telecommunications system. It provides a REST API for processing incoming call-routing requests, selecting configured SIP destinations, and recording routing decisions in a cloud-hosted PostgreSQL database.

The project explores backend service architecture, database-driven routing, connection pooling, error handling, automated testing, and cloud integration using Go.

> **Project scope:** This application is a SIP routing decision service, not a SIP proxy or PBX. It selects SIP destination URIs but does not currently process SIP signaling, establish voice sessions, or verify call completion.

## Technology Stack

| Technology | Purpose |
|---|---|
| Go | Core backend application and routing logic |
| Go `net/http` | REST API implementation |
| PostgreSQL | Routing configuration and persistent call records |
| Neon | Managed, cloud-hosted PostgreSQL |
| pgx/v5 | PostgreSQL driver and connection pooling |
| godotenv | Local environment configuration |
| Go testing package | Automated unit testing |
| Git / GitHub | Version control and source management |

## Features

### Dynamic SIP Gateway Routing

The application accepts call-routing requests and selects an appropriate SIP destination based on the requested geographic region.

Routing rules are stored in PostgreSQL rather than hardcoded in the application, allowing administrators to modify destinations without rebuilding or restarting the Go service.

Current routing behavior includes:

- Region-based SIP destination selection.
- Configurable fallback routing for unsupported regions.
- Ability to enable or disable routing entries.
- Case-insensitive region matching.
- Request validation and structured error handling.

### Cloud-Hosted PostgreSQL Integration

The application connects to a managed Neon PostgreSQL database using the `pgx` driver.

Database connectivity includes:

- Connection pooling to reuse PostgreSQL connections.
- TLS-enabled connections through the configured database URL.
- Startup connectivity verification.
- Context-based query timeouts.
- Parameterized SQL queries.

### Persistent Routing History

Each valid request that produces a routing decision is recorded in PostgreSQL.

Stored information includes:

- Call identifier.
- Originating caller number.
- Destination number.
- Requested region.
- Selected SIP gateway.
- Routing decision status.
- Database-generated timestamp.

These records provide a foundation for future operational reporting and analytics.

### Automated Testing

Go unit tests validate request handling and input requirements, including missing or whitespace-only call identifiers and destinations.

The application has also been manually validated against the live Neon PostgreSQL database to verify route retrieval and persistent call-record insertion.

## System Architecture

```text
                 HTTP Client
                      |
                      | POST /route
                      v
             +------------------+
             |    Go REST API   |
             |    net/http      |
             +------------------+
                      |
                      v
             +------------------+
             |  Routing Engine  |
             |                  |
             | Input Validation |
             | Region Matching  |
             | Fallback Routing |
             +------------------+
                      |
                      | SQL SELECT
                      v
             +------------------+
             | Neon PostgreSQL  |
             |    sip_routes    |
             +------------------+
                      |
                      | Selected SIP URI
                      v
             +------------------+
             | Call Record      |
             | Repository       |
             +------------------+
                      |
                      | SQL INSERT
                      v
             +------------------+
             | Neon PostgreSQL  |
             |   call_records   |
             +------------------+
                      |
                      v
                JSON Response
```

The Go API runs locally during development, while PostgreSQL is hosted in the cloud. The service is designed to support future containerization and deployment.

## Project Structure

```text
SIP-Routing-Flow/
├── main.go
├── go.mod
├── go.sum
├── README.md
├── .gitignore
├── .env.example
└── internal/
    ├── database/
    │   ├── postgres.go
    │   └── call_records.go
    └── routing/
        ├── router.go
        └── router_test.go
```

### Component Responsibilities

**main.go**

Initializes environment configuration, establishes database connectivity, registers HTTP endpoints, and coordinates routing requests and persistence.

**internal/database/postgres.go**

Creates and verifies the PostgreSQL connection pool.

**internal/database/call_records.go**

Implements the repository responsible for inserting routing decisions into PostgreSQL.

**internal/routing/router.go**

Contains request validation, region-based route selection, PostgreSQL lookup logic, and fallback routing behavior.

**internal/routing/router_test.go**

Contains automated tests for request validation and error conditions.

## API Reference

### GET /health

Returns the operational status of the Go HTTP service.

**Example request**

```bash
curl http://localhost:8080/health
```

**Example response**

```json
{
  "status": "healthy",
  "service": "sip-routing-flow"
}
```

This endpoint checks the application's HTTP availability. It does not continuously verify PostgreSQL connectivity.

### POST /route

Processes a routing request, retrieves a SIP destination from PostgreSQL, records the routing decision, and returns the result.

**Example request**

```bash
curl -X POST http://localhost:8080/route \
  -H "Content-Type: application/json" \
  -d '{
    "call_id": "CALL-1001",
    "caller": "+14045550123",
    "destination": "+18005550100",
    "region": "US"
  }'
```

**Example response**

```json
{
  "call_id": "CALL-1001",
  "route": "sip:us-gateway@pbx.internal",
  "status": "ROUTE_FOUND"
}
```

The returned route depends on the current database configuration.

### HTTP Response Codes

| Code | Meaning |
|---|---|
| 200 | Routing destination successfully selected |
| 400 | Invalid JSON or missing required fields |
| 404 | No enabled matching or fallback route |
| 405 | Unsupported HTTP method |
| 503 | Database routing or persistence operation unavailable |

## Database Schema

The application uses two PostgreSQL tables.

### sip_routes

Stores configured SIP routing destinations.

```sql
CREATE TABLE sip_routes (
    id BIGSERIAL PRIMARY KEY,
    region VARCHAR(20) NOT NULL UNIQUE,
    sip_uri VARCHAR(255) NOT NULL,
    priority INTEGER NOT NULL DEFAULT 1,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### call_records

Stores the results of routing decisions.

```sql
CREATE TABLE call_records (
    id BIGSERIAL PRIMARY KEY,
    call_id VARCHAR(100) NOT NULL,
    caller VARCHAR(50),
    destination VARCHAR(50) NOT NULL,
    region VARCHAR(20),
    selected_route VARCHAR(255),
    routing_status VARCHAR(30) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

**Example route configuration**

```sql
INSERT INTO sip_routes (region, sip_uri, enabled)
VALUES
    ('US', 'sip:us-gateway@pbx.internal', TRUE),
    ('EU', 'sip:eu-gateway@pbx.internal', TRUE),
    ('DEFAULT', 'sip:default@pbx.internal', TRUE)
ON CONFLICT (region) DO NOTHING;
```

The `.internal` addresses are demonstration placeholders and do not represent publicly reachable SIP gateways.

## Getting Started

### Prerequisites

- Go installed locally.
- Access to a PostgreSQL database (Neon recommended).
- Git.
- A terminal or IDE.

### 1. Clone the Repository

```bash
git clone https://github.com/Shannon-Reed/SIP-Routing-Flow.git
cd SIP-Routing-Flow
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Configure PostgreSQL

Create a Neon PostgreSQL project at https://console.neon.tech/ or use an existing compatible PostgreSQL instance.

Create the `sip_routes` and `call_records` tables using the SQL definitions above.

### 4. Configure Environment Variables

Create a `.env` file in the project root:

```dotenv
DATABASE_URL=postgresql://username:password@hostname/database?sslmode=require
PORT=8080
```

Replace the placeholder URL with your PostgreSQL connection string.

**Security:** Never commit your `.env` file or database credentials to source control.

### 5. Run the Application

```bash
go run .
```

The API listens on:

```text
http://localhost:8080
```

### 6. Run Automated Tests

```bash
go test ./... -v
```

Tests can be executed without a live PostgreSQL connection for the currently implemented input-validation scenarios.

## Verifying Dynamic Routing

One of the project's primary features is the ability to update routing configuration without restarting the application.

For example, execute the following SQL in Neon:

```sql
UPDATE sip_routes
SET sip_uri = 'sip:us-backup@pbx.internal'
WHERE region = 'US';
```

Send another request to `POST /route` with `"region": "US"`.

The API will return the updated SIP URI using the existing running Go service.

This demonstrates database-driven configuration and separation of application logic from operational routing data.

## Engineering Considerations

The project incorporates several backend engineering practices:

- **Separation of concerns:** HTTP request handling, routing logic, and database persistence are organized into distinct packages.
- **Connection management:** PostgreSQL connection pooling reduces the overhead of establishing new connections for every request.
- **Query safety:** Parameterized SQL queries keep input values separate from SQL statements.
- **Request cancellation:** Context propagation allows database operations to respect HTTP request deadlines.
- **Persistence reliability:** Routing decisions are written to PostgreSQL before successful API responses are returned.
- **Error handling:** The API distinguishes invalid requests, unavailable routes, and service failures.
- **Cloud configuration:** Database credentials are managed through environment variables rather than embedded in source code.

## Current Limitations

This project currently focuses on routing decisions and database persistence.

It does not yet implement:

- SIP INVITE processing or actual call establishment.
- Real-time SIP gateway health checks.
- Automatic failover based on live gateway availability.
- Authentication or rate limiting for public API access.
- Idempotent request handling.
- Full database integration-test coverage.
- Completed-call reporting or call-detail records.

The database schema includes a priority field for future multi-gateway routing, but the current implementation permits only one configured route per region.

## Planned Enhancements

### Python Analytics Microservice

Build a separate FastAPI service that connects to PostgreSQL and provides endpoints for:

- Total routing requests.
- Routing distribution by region.
- Successful versus unsuccessful route-selection attempts.
- Recent routing activity.
- Historical reporting.

### Containerization and Deployment

- Create Dockerfiles for backend services.
- Configure local multi-service development using Docker Compose.
- Deploy the Go API to a cloud hosting platform.
- Add application readiness and database health checks.

### Advanced Routing

- Support multiple SIP gateways per region.
- Introduce routing priority and failover.
- Add gateway availability monitoring.
- Implement configurable routing policies.
- Explore SIP signaling integration.

## Project Purpose

SIP Routing Flow was developed as a practical exploration of Go backend development and cloud-based telecommunications architecture.

The project applies concepts from VoIP and call-routing systems to a modern backend implementation, with an emphasis on modularity, maintainability, cloud data persistence, and extensibility.

It serves as a foundation for further development involving Python analytics, containerized services, and more advanced telecommunications routing behavior.


**Author:** Shannon Reed

**Repository:** https://github.com/Shannon-Reed/SIP-Routing-Flow
