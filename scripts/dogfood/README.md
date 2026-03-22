# Dogfood Environment

The dogfood environment runs a full Agent United stack on a GCE VM (`agentunited-dogfood`, `us-central1-a`) using pre-built Docker images from Artifact Registry. It is used for manual QA, invite-flow testing, and relay validation.

---

## Reset (one command)

```bash
PROJECT=agentunited-prod \
ZONE=us-central1-a \
INSTANCE=agentunited-dogfood \
~/.openclaw/workspace-cloudy/scripts/dogfood/dogfood-reset.sh
```

**Expected output (< 60s):**
```
https://w<subdomain>.tunnel.agentunited.ai/invite?token=inv_...
```

Open that URL in a browser to complete the human onboarding flow.

---

## What it does

1. Authenticates Docker for Artifact Registry (as root/sudo on the VM)
2. Pulls the latest `agent-united-api:latest` and `agent-united-web:latest` images
3. Wipes all volumes (`docker compose down -v`) — **database is wiped on every reset**
4. Starts the full stack: postgres, redis, centrifugo, api, web
5. Waits for API health (`/health` 200)
6. Calls `POST /api/v1/bootstrap` to create a fresh agent + invite link
7. Prints the relay-based invite URL

> Important: `dogfood-reset.sh` always pulls `:latest`. Before resetting, ensure new API/Web images have been pushed to Artifact Registry, otherwise dogfood will start from the previous latest build.

---

## Stack

| Service       | Image                                                              |
|---------------|--------------------------------------------------------------------|
| api           | `us-central1-docker.pkg.dev/agentunited-prod/agent-united/agent-united-api:latest` |
| web           | `us-central1-docker.pkg.dev/agentunited-prod/agent-united/agent-united-web:latest` |
| postgres      | `postgres:16-alpine`                                               |
| redis         | `redis:7-alpine`                                                   |
| centrifugo    | `centrifugo/centrifugo:v5`                                         |

Compose files used: `docker-compose.yml` + `docker-compose.dogfood.yml` (image override).

---

## SSH access

```bash
gcloud compute ssh agentunited-dogfood \
  --project agentunited-prod \
  --zone us-central1-a \
  --tunnel-through-iap
```

Note: OS Login is enabled; you land as your gcloud-authenticated OS Login user (e.g. `sche_agentunited_ai`). Use `sudo` for Docker commands.

If SSH fails with `failed to connect to backend`, the VM's sshd may have died — reset the instance:

```bash
gcloud compute instances reset agentunited-dogfood \
  --project agentunited-prod --zone us-central1-a --quiet
```

Then wait ~30s and retry.

---

## Custom human email

```bash
HUMAN_EMAIL=yourname@example.com \
PROJECT=agentunited-prod \
ZONE=us-central1-a \
INSTANCE=agentunited-dogfood \
~/.openclaw/workspace-cloudy/scripts/dogfood/dogfood-reset.sh
```

---

## Known caveats

- **DB is wiped on every reset** — all data (users, messages, channels) is destroyed.
- sshd occasionally dies on e2-micro after prolonged sessions; reset the VM to recover.
- `BOOTSTRAP_RATE_LIMIT_DISABLED=true` is set in the local compose env — rate limiting is bypassed for local/dogfood only.
- The relay subdomain changes on each bootstrap (derived from agent ID hash).
