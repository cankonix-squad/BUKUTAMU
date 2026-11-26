#!/usr/bin/env bash
set -Eeuo pipefail
umask 077
BASE=/root/cankonix-node/apps/bukutamu
SHA=${1:?release SHA required}
[[ "$SHA" =~ ^[0-9a-f]{40}$ ]] || exit 2
exec 9>"$BASE/deploy.lock"
flock -w 900 9
RELEASE="$BASE/releases/$SHA"
export RELEASE_SHA="$SHA"
compose() { docker compose --env-file "$BASE/shared/.env" -f "$RELEASE/docker-compose.prod.yml" "$@"; }
compose config --quiet
previous=$(cat "$BASE/current-sha" 2>/dev/null || true)
rollback() {
  status=$?
  trap - ERR
  compose ps || true
  compose logs --tail 60 backend frontend || true
  if [[ "$previous" =~ ^[0-9a-f]{40}$ ]] && [[ "$previous" != "$SHA" ]]; then
    echo "Restoring application release $previous (database is not restored automatically)"
    RELEASE="$BASE/releases/$previous"
    export RELEASE_SHA="$previous"
    compose up -d --wait --wait-timeout 180 || true
  fi
  exit "$status"
}
trap rollback ERR
docker load -i "$RELEASE/images.tar.gz"
# Back up before automatic application migrations. Never remove persistent volumes.
if [[ -n "$previous" ]]; then
  mkdir -p "$BASE/backups"
  compose exec -T db pg_dump -U bukutamu -d bukutamu -Fc > "$BASE/backups/$(date -u +%Y%m%dT%H%M%SZ)-$previous.dump"
fi
compose up -d --wait --wait-timeout 180
curl --fail --silent --show-error --retry 12 --retry-delay 5 --retry-all-errors https://bukutamu.cankonix.com/api/health
curl --fail --silent --show-error https://bukutamu.cankonix.com/ -o /dev/null
printf '%s\n' "$SHA" > "$BASE/current-sha.tmp"
mv "$BASE/current-sha.tmp" "$BASE/current-sha"
rm -f "$RELEASE/images.tar.gz"
echo "Production active: $SHA"
