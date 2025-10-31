#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")"/.. && pwd)"

cd "$ROOT_DIR"

echo "[1/7] Building frontend with pnpm"
pnpm config set recursive-install-mode parallel
pnpm --dir client install --frozen-lockfile --prefer-offline
pnpm --dir client run build

echo "[2/7] Embedding latest frontend assets"
rm -rf backend/routes/static
mkdir -p backend/routes/static
cp -R client/dist/. backend/routes/static/

echo "[3/7] Running backend unit tests"
(cd backend && go test ./...)

echo "[4/7] Building linux/amd64 binary"
(cd backend && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ../kubechat .)

echo "[5/7] Building local docker image kubechat:dev (no cache)"
docker build --no-cache --build-arg TARGETARCH=amd64 -f .goreleaser.Dockerfile -t kubechat:dev .

echo "[6/7] Deploying helm chart with dev values"
helm upgrade --install kubechat ./charts/kubechat \
  -n kubechat-system --create-namespace \
  --set image.repository=kubechat \
  --set image.tag=dev \
  --set image.pullPolicy=IfNotPresent \
  --set service.listen=":7080" \
  -f charts/kubechat/values-dev.yaml

echo "[7/7] Waiting for rollout"
kubectl -n kubechat-system rollout status deploy/kubechat

echo "Done. kubechat should now be serving via Traefik at https://kubechat.local (see docs/local-development.md)."
