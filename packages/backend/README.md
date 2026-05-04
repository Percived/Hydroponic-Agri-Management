# Hydroponic Agri Management Backend

Go backend for hydroponic greenhouse management.

## Tech Stack
- Go (`gin`, `gorm`)
- MySQL (relational data)
- InfluxDB (time-series telemetry)
- EMQX (MQTT broker)

## Current Capabilities
- Auth and RBAC (`ADMIN` / `OPERATOR` / `VIEWER`)
- Device and device-group management
- Telemetry ingestion/query (`latest`, `history`, `stats`)
- Control commands / rules / templates
- Alert query and status workflow
- Audit-log query
- Dashboard aggregation (`devices_online`, `devices_offline`, `alerts_open`)

## Project Structure
- `cmd/api`: application entrypoint
- `internal/auth`: auth, JWT, user/role APIs
- `internal/device`: device and group APIs
- `internal/telemetry`: telemetry ingest/query and retention config
- `internal/control`: command/rule/template APIs
- `internal/alert`: alert APIs
- `internal/audit`: audit write/query
- `internal/overview`: dashboard aggregation API
- `internal/platform`: config, db clients, middleware, response envelope
- `migrations`: SQL migrations
- `scripts`: local utility scripts

## Quick Start
### 1. Start dependencies
```bash
docker compose up -d
```

### 2. Initialize database
```bash
docker compose exec -T mysql mysql -uroot -proot hydroponic < migrations/0001_init.up.sql
docker compose exec -T mysql mysql -uroot -proot hydroponic < migrations/0002_seed_auth.up.sql
docker compose exec -T mysql mysql -uroot -proot hydroponic < migrations/0003_seed_metrics.up.sql
```

### 3. Run backend
```bash
go run cmd/api/main.go
```

### 4. Login smoke test
```bash
scripts/dev-login-smoke.sh
```

## Default Local Config
From `configs/config.yaml`:
- API: `http://127.0.0.1:8080`
- MySQL: `127.0.0.1:3307` (container internal `3306`)
- InfluxDB: `http://127.0.0.1:8086`
- MQTT: `tcp://127.0.0.1:1883`

Default seeded admin account:
- username: `admin`
- password: `admin123`

## API Docs
- Main API spec: `docs/specs/API_SPEC.md`
- Device protocol: `docs/specs/DEVICE_PROTOCOL.md`
- Requirements: `docs/analysis/REQUIREMENTS.md`
- MVP scope: `docs/analysis/MVP.md`
- DFD and sequence: `docs/architecture/DFD-0-Context.md`, `docs/architecture/DFD-1-System.md`, `docs/architecture/DFD-2-Modules.md`, `docs/architecture/SEQUENCE.md`
- Manual acceptance: `docs/testing/MANUAL_API_ACCEPTANCE.md`

## Development Checks
```bash
GOCACHE=/tmp/gocache go test ./...
```

## Notes
- If local MySQL also runs on `3306`, this project uses host `3307` to avoid conflicts.
- Influx write failures are logged as warnings and do not block telemetry ingestion.
