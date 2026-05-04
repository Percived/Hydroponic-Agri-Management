# Manual API Acceptance (curl)

This document provides module-by-module manual acceptance steps.

## Prerequisites
1. Start dependencies:
```bash
docker compose up -d
```
2. Init DB:
```bash
docker compose exec -T mysql mysql -uroot -proot hydroponic < migrations/0001_init.up.sql
docker compose exec -T mysql mysql -uroot -proot hydroponic < migrations/0002_seed_auth.up.sql
```
3. Start backend:
```bash
go run cmd/api/main.go
```
4. Set base URL:
```bash
BASE="http://127.0.0.1:8080"
```

## 1) Auth Module
### Login success
```bash
curl -s -X POST "$BASE/api/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```
Expected: `code=0` and token exists.

### Save token
```bash
TOKEN="<paste-token-here>"
```

### Unauthorized check
```bash
curl -s "$BASE/api/users"
```
Expected: unauthorized.

## 2) Device Module
### Create device
```bash
curl -s -X POST "$BASE/api/devices" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"device_code":"DEV-001","name":"Temp Sensor","type":"SENSOR","category":"TEMP","protocol":"MQTT","sampling_interval_sec":60}'
```

### Query devices
```bash
curl -s "$BASE/api/devices" -H "Authorization: Bearer $TOKEN"
```

### Device health
```bash
curl -s "$BASE/api/devices/1/health" -H "Authorization: Bearer $TOKEN"
```

## 3) Telemetry Module
### Seed metric (if missing)
```bash
docker compose exec -T mysql mysql -uroot -proot hydroponic -e "INSERT INTO metrics(code,name,unit,min_value,max_value) VALUES('TEMP','Temperature','C',-10,60) ON DUPLICATE KEY UPDATE name=VALUES(name);"
```

### Ingest telemetry
```bash
curl -s -X POST "$BASE/api/telemetry" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"device_code":"DEV-001","metrics":[{"code":"TEMP","value":26.3,"unit":"C"}]}'
```

### Query latest
```bash
curl -s "$BASE/api/telemetry/latest?device_id=1&metrics=TEMP" -H "Authorization: Bearer $TOKEN"
```

### Query history
```bash
START=$(date -u -v-1H +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -d '1 hour ago' +"%Y-%m-%dT%H:%M:%SZ")
END=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
curl -s "$BASE/api/telemetry/history?device_id=1&metric_code=TEMP&start_time=$START&end_time=$END&page=1&page_size=20" -H "Authorization: Bearer $TOKEN"
```

### Query stats
```bash
curl -s "$BASE/api/telemetry/stats?device_id=1&metric_code=TEMP&start_time=$START&end_time=$END" -H "Authorization: Bearer $TOKEN"
```

## 4) Control Module
### Create rule
```bash
curl -s -X POST "$BASE/api/controls/rules" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"High Temp Fan On","metric_code":"TEMP","operator":">","threshold":25,"action":{"command_type":"SWITCH","payload":{"state":"ON"}},"target_device_id":1,"enabled":true}'
```

### Trigger rule by telemetry
```bash
curl -s -X POST "$BASE/api/telemetry" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"device_code":"DEV-001","metrics":[{"code":"TEMP","value":30.1,"unit":"C"}]}'
```

### Check commands
```bash
curl -s "$BASE/api/controls/commands?page=1&page_size=20" -H "Authorization: Bearer $TOKEN"
```

## 5) Alert Module
### List alerts
```bash
curl -s "$BASE/api/alerts?page=1&page_size=20" -H "Authorization: Bearer $TOKEN"
```

### Alert stats
```bash
curl -s "$BASE/api/alerts/stats" -H "Authorization: Bearer $TOKEN"
```

### Update alert status
```bash
curl -s -X PATCH "$BASE/api/alerts/1/status" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"status":"ACK","comment":"checked"}'
```

## 6) Audit Module
### Query audit logs
```bash
curl -s "$BASE/api/audit-logs?page=1&page_size=20" -H "Authorization: Bearer $TOKEN"
```

### Query by action/time
```bash
curl -s "$BASE/api/audit-logs?action=CONTROL_COMMAND&start_time=$START&end_time=$END&page=1&page_size=20" -H "Authorization: Bearer $TOKEN"
```

## 7) Overview Module
### Dashboard
```bash
curl -s "$BASE/api/overview/dashboard" -H "Authorization: Bearer $TOKEN"
```
Expected keys: `devices_online`, `devices_offline`, `alerts_open`.

## Pass Criteria
- Protected APIs require token and role constraints work.
- Telemetry ingestion persists and can be queried.
- Rule trigger produces command and alert records.
- Audit logs are queryable with filters.
- Dashboard returns aggregated real values.
