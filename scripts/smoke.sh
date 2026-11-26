#!/usr/bin/env bash
# End-to-end API smoke test: docker compose up bo'lganda yoki backend mustaqil ishlaganda ishlatish.
set -euo pipefail
API="${API:-http://localhost:8080/api}"
PASS=0
FAIL=0

ok()   { echo "PASS  $1"; PASS=$((PASS+1)); }
bad()  { echo "FAIL  $1"; FAIL=$((FAIL+1)); }
jget() { python3 -c 'import json,sys; print(json.load(sys.stdin).get(sys.argv[1],""))' "$1"; }

echo "API: $API"

# 1) Admin login
ADMIN_TOKEN=$(curl -sf -X POST "$API/auth/login" -H 'Content-Type: application/json' \
  -d '{"email":"admin@tourchef.uz","password":"admin123"}' | jget token)
[ -n "$ADMIN_TOKEN" ] && ok "admin login" || bad "admin login"

# 2) Chef registration -> admin panelda ko'rinishi kerak
EMAIL="chef-smoke-$(date +%s)@example.com"
REG=$(curl -sf -X POST "$API/auth/register" -H 'Content-Type: application/json' \
  -d "{\"email\":\"$EMAIL\",\"password\":\"test12345\",\"full_name\":\"Smoke Chef\",\"role\":\"chef\"}")
[ -n "$(echo "$REG" | jget token)" ] && ok "chef register" || bad "chef register"
curl -sf "$API/admin/chefs" -H "Authorization: Bearer $ADMIN_TOKEN" | grep -q "Smoke Chef" \
  && ok "new chef visible in /admin/chefs" || bad "new chef visible in /admin/chefs"

# 3) Admin taom qo'shadi -> katalogda ko'rinishi kerak
DISH_NAME="Smoke Dish $(date +%s)"
DISH=$(curl -sf -X POST "$API/dishes" -H "Authorization: Bearer $ADMIN_TOKEN" -H 'Content-Type: application/json' \
  -d "{\"name\":\"$DISH_NAME\",\"category\":\"Soup\",\"region\":\"Tashkent\",\"description\":\"Smoke test dish\",\"ingredients\":\"water, salt\",\"image_url\":\"\",\"status\":\"active\"}")
[ -n "$(echo "$DISH" | jget id)" ] && ok "admin create dish" || bad "admin create dish"
curl -sf "$API/dishes?q=Smoke%20Dish" | grep -q "$DISH_NAME" \
  && ok "new dish visible in /dishes" || bad "new dish visible in /dishes"

# 4) Masterclass filterlari
MC=$(curl -sf "$API/masterclasses?city=Tashkent&sort=price_asc&limit=5")
[ -n "$MC" ] && ok "masterclass filters" || bad "masterclass filters"

# 5) Client login va buyurtma
CLIENT_TOKEN=$(curl -sf -X POST "$API/auth/login" -H 'Content-Type: application/json' \
  -d '{"email":"client@tourchef.uz","password":"client123"}' | jget token)
[ -n "$CLIENT_TOKEN" ] && ok "client login" || bad "client login"
SID=$(curl -sf "$API/masterclasses" | python3 -c 'import json,sys; d=json.load(sys.stdin); print((d.get("items") or [{}])[0].get("slug",""))' )
if [ -n "$SID" ] && [ -n "$CLIENT_TOKEN" ]; then
  SESSION_ID=$(curl -sf "$API/masterclasses/$SID/sessions" | python3 -c 'import json,sys; d=json.load(sys.stdin); print((d[0] if isinstance(d,list) else (d.get("items") or [{}])[0]).get("id",""))')
  if [ -n "$SESSION_ID" ]; then
    curl -sf -X POST "$API/bookings" -H "Authorization: Bearer $CLIENT_TOKEN" -H 'Content-Type: application/json' \
      -d "{\"session_id\":\"$SESSION_ID\",\"guests\":2}" >/dev/null \
      && ok "client booking" || bad "client booking"
  fi
fi

echo "---"
echo "PASS=$PASS FAIL=$FAIL"
[ "$FAIL" -eq 0 ]
