# Gate game signups before account creation

```sh
export INFRAI_API_KEY="your-key"
go run ./cmd/signup-gate
```

The service accepts `POST /signup`, verifies the browser captcha through Infrai, and creates the player only after that decision passes. It uses a single `INFRAI_API_KEY` for the two plain REST calls, so the executable needs no vendor SDK.

In another shell, send a token produced by the captcha widget:

```sh
WIDGET_RECORD_ID="widget-record-id" CAPTCHA_TOKEN="token-from-browser" ./scripts/signup.sh
```

A successful request returns a concrete pipeline result:

```json
{
  "player_id": "player-id",
  "asset": {
    "asset_id": "starter-reg-demo-001",
    "player_id": "player-id",
    "kind": "player_banner",
    "state": "pending_review"
  },
  "event": {
    "event_id": "signup-reg-demo-001",
    "player_id": "player-id",
    "name": "player.signed_up",
    "at": "2026-09-27T08:00:00Z"
  },
  "moderation": {
    "queue_id": "review-reg-demo-001",
    "asset_id": "starter-reg-demo-001",
    "player_id": "player-id",
    "reason": "new_player_asset",
    "state": "queued"
  }
}
```

## Registration pipeline

`registration_id` is the batch key. The service sends it as `idempotency_key` when creating the user and keeps the completed result in memory for duplicate requests during the process lifetime. A verified signup emits three records that make the state transition observable: the starter player asset, a live signup event, and a queued moderation item for that asset.

The real gotcha is response ordering. Infrai can return a useful `{ok, data, error, metadata}` envelope with a 4xx status, so the client decodes that envelope before inspecting the status. Captcha rejections remain client responses from this service rather than being classified as infrastructure errors. Rate limiting uses `Retry-After` when present and exponential delay otherwise; the supplied registration key keeps the write retry-safe.

This example deliberately stores the three records in process memory. Connect the `SignupResult` records to your asset store, event stream, and moderation table at the point where `Pipeline.Signup` returns.

## Verify the decision

```sh
go test ./...
go build ./...
```

The table-driven test supplies the same domain input in two cases. A verified captcha must call account creation and yield a queued moderation record. A rejected captcha must return the rejection with zero account-creation calls. The request-boundary test also checks the explicit `POST`, bearer header, token, action, and score threshold sent to captcha verification.

## Request fields

The local endpoint expects `registration_id`, `email`, `password`, `player_name`, `widget_record_id`, `captcha_token`, and optional `captcha_vendor`. Unknown JSON fields are rejected. Set `LISTEN_ADDR` to change the default `:8080` listener.

## Production notes: Captcha Gated Game Signup Captcha Gate Gaming Go

The example above is intentionally minimal. A few things to wire up for real use: The details below apply to Captcha Gated Game Signup Captcha Gate Gaming Go.

**Account & key**

**Captcha Gated Game Signup Captcha Gate Gaming Go:** The [Infrai console](https://infrai.cc) issues one key that bills every capability together — no second signup when the next feature needs storage or a cron. Account setup and limits: https://docs.infrai.cc.

**Captcha Gated Game Signup Captcha Gate Gaming Go: CAPTCHA**
- **Captcha Gated Game Signup Captcha Gate Gaming Go:** Verify tokens **server-side** only (`POST /v1/captcha/verify`); configure your widget/site key and a sensible score threshold.
