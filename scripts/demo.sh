#!/usr/bin/env bash
# Usage: scripts/demo.sh   -> prints success, unavailable, and inventory-down evidence
set -u
cd "$(dirname "$0")/.."
mkdir -p evidence
go build -o bin/ ./cmd/...
order() {
  echo "\$ curl -X POST localhost:8080/orders -d '$1'"
  curl -s -w 'HTTP %{http_code}\n' -X POST localhost:8080/orders -H 'Content-Type: application/json' -d "$1"
}
pids=()
cleanup() { kill "${pids[@]}" 2>/dev/null; }
trap cleanup EXIT

echo "== 1. both services running (success) =="
bin/inventory 2>evidence/inventory.log & pids+=($!)
bin/order 2>evidence/order.log & pids+=($!)
sleep 1
order '{"item_id":"widget","quantity":2}'
echo "\$ curl localhost:8080/orders/ord-1"
curl -s -w 'HTTP %{http_code}\n' localhost:8080/orders/ord-1

echo "== 2. unavailable cases =="
order '{"item_id":"gadget","quantity":1}'
order '{"item_id":"nope","quantity":1}'

echo "== 3. inventory service stopped =="
kill "${pids[0]}"; wait "${pids[0]}" 2>/dev/null
echo "inventory service PID ${pids[0]} stopped"
order '{"item_id":"widget","quantity":2}'

echo "== 4. order service still running =="
kill -0 "${pids[1]}" && echo "order service PID ${pids[1]} alive"
order '{"item_id":"widget","quantity":0}'

echo "== 5. inventory restarted, orders succeed again =="
bin/inventory 2>>evidence/inventory.log & pids[0]=$!
sleep 1
order '{"item_id":"widget","quantity":2}'

echo "== order service log =="
cat evidence/order.log
