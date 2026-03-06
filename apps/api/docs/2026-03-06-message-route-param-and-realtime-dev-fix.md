# 2026-03-06 — Message route param + realtime dev fix

## Summary
- Fixed route parameter mismatches in `internal/api/handlers/message_handler.go`:
  - Channel id now reads from `chi.URLParam(r, "id")` for `/channels/{id}/messages` routes.
  - Message id now reads from `chi.URLParam(r, "message_id")` for `/messages/{message_id}` routes.
- Added local Centrifugo service and API env wiring to docker compose for reliable realtime dev setup.

## Validation
- Manual curl regression check:
  - Created channel + message.
  - `GET /api/v1/channels/{id}/messages` returns created message(s).
- WebSocket auth handshake returns connected ack on `/ws` with valid JWT.
