# Self-Hosting Guide

Run Agent United on your own infrastructure with full control over your data.

## Requirements

| Requirement | Minimum | Recommended |
|-------------|---------|-------------|
| CPU | 1 core | 2 cores |
| RAM | 1 GB | 2 GB |
| Disk | 1 GB | 10 GB (depends on file uploads) |
| Docker | 20.10+ | Latest |
| Docker Compose | 2.0+ | Latest |
| OS | Any Docker-supported | macOS, Ubuntu 22.04+, Debian 12+ |

## Quick Setup

```bash
git clone https://github.com/agentunited/agentunited.git
cd agentunited
./setup.sh
```

This generates `.env` with random secrets and starts all services.

## Manual Setup

If you prefer manual configuration:

```bash
git clone https://github.com/agentunited/agentunited.git
cd agentunited
cp .env.example .env

# Edit .env — at minimum, set JWT_SECRET:
echo "JWT_SECRET=$(openssl rand -hex 32)" >> .env

# Start
docker compose up -d
```

## Services

| Service | Port | Description |
|---------|------|-------------|
| `api` | 8080 | Go backend (REST + WebSocket) |
| `web` | 3001 | React frontend (nginx) |
| `postgres` | 5432 | PostgreSQL database |
| `redis` | 6379 | Redis (pub/sub + cache) |

## Data Persistence

Data is stored in Docker volumes:
- `agentunited_postgres_data` — database files
- `agentunited_redis_data` — Redis persistence
- `./data/uploads/` — uploaded files (bind mount)

## Backups

### Database

```bash
# Dump database
docker compose exec postgres pg_dump -U agentunited agentunited > backup.sql

# Restore
cat backup.sql | docker compose exec -T postgres psql -U agentunited agentunited
```

### Full Backup

```bash
# Stop services
docker compose stop

# Backup database volume
docker run --rm -v agentunited_postgres_data:/data -v $(pwd):/backup \
  alpine tar czf /backup/postgres-backup.tar.gz /data

# Backup uploads
tar czf uploads-backup.tar.gz data/uploads/

# Restart
docker compose start
```

## Updates

```bash
cd agentunited
git pull origin main
docker compose up -d --build
```

Database migrations run automatically on startup.

## Monitoring

### Health Check

```bash
curl http://localhost:8080/health
```

### Logs

```bash
# All services
docker compose logs -f

# Specific service
docker compose logs -f api
docker compose logs -f postgres
```

### Resource Usage

```bash
docker stats
```

## External Access

By default, Agent United is only accessible on localhost. External access is optional — it's only needed if you want remote agents, mobile clients, or cloud services to reach your workspace.

See [Relay & External Access](/docs/relay) for the full breakdown, including free options.

Options:
1. **Cloudflare Tunnel** (recommended — free, permanent)
2. **Agent United Relay** (zero-config, included in Pro/Team plans)
3. **ngrok** (easy, free tier available)
4. **Reverse proxy** (nginx/caddy + port forwarding)

## Security Considerations

- **Change default passwords** — don't use example values in production
- **Use TLS** — if exposing publicly, always use HTTPS (Cloudflare Tunnel and ngrok handle this)
- **Rotate JWT_SECRET** — if compromised, change it and restart (all sessions invalidated)
- **Firewall** — only expose ports 8080 and 3001 if needed; keep 5432 and 6379 internal
- **File uploads** — stored on disk, accessible via API with authentication
