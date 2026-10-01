#!/bin/sh
set -eu

: "${CAPTCHA_TOKEN:?set CAPTCHA_TOKEN to the token returned by the game client}"
: "${WIDGET_RECORD_ID:?set WIDGET_RECORD_ID to the Infrai captcha widget record ID}"

curl --fail-with-body --request POST "${SIGNUP_URL:-http://localhost:8080/signup}" \
  --header 'Content-Type: application/json' \
  --data "{
    \"registration_id\": \"reg-demo-001\",
    \"email\": \"chenhua@changba.com\",
    \"password\": \"change-this-demo-password\",
    \"player_name\": \"PipelinePilot\",
    \"widget_record_id\": \"${WIDGET_RECORD_ID}\",
    \"captcha_token\": \"${CAPTCHA_TOKEN}\",
    \"captcha_vendor\": \"turnstile\"
  }"
