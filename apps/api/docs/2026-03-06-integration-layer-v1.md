# 2026-03-06 — Platform Integration Layer (Phase 8)

Implemented backend integration v1 primitives:

## Added integration adapter layer
- `pkg/integrations/adapter.go`
  - `IntegrationAdapter` interface
  - platform constants: `openclaw`, `langgraph`, `autogen`, `custom`
  - canonical `Event` + `InboundMessage`
- `pkg/integrations/openclaw.go`
  - reference OpenClaw adapter implementation
- `pkg/integrations/registry.go`
  - adapter registry
- `pkg/integrations/manager.go`
  - `IntegrationManager` skeleton for typed event routing

## Persistence
- Migration `012_integrations.sql` adds:
  - `integrations(id, workspace_id, platform, api_key, webhook_url, event_subscriptions jsonb, active, created_at)`

## API endpoints
Added authenticated integration management routes:
- `GET /api/v1/integrations`
- `POST /api/v1/integrations`
- `DELETE /api/v1/integrations/{id}`

### QA fixes (Moon pass)
- `POST /api/v1/integrations` now includes `name` in the returned integration object.
- `event_subscriptions` now reliably persists and returns as an array (never null when values are provided).
- Backward compatibility: create now also accepts `events` as an alias for `event_subscriptions`.

Same routes are mounted on `/v1` compatibility path.

## Event routing + outbound signing
- `message.created` now routes into integration service from `message_handler.go`.
- Integration service routes typed events to active adapters by subscription and platform.
- Outbound payloads are HMAC-SHA256 signed and delivered with headers:
  - `X-AgentUnited-Event`
  - `X-AgentUnited-Integration-ID`
  - `X-AgentUnited-Signature`
- Delivery/retry transport path reuses webhook delivery infrastructure via `WebhookService.DispatchSigned`.
