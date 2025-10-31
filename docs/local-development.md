# Local Development & Helm Deployment Guide

This guide walks through running Kubechat locally, building the release image using `.goreleaser.Dockerfile`, and deploying the result to a Kubernetes cluster with the bundled Helm chart.

## Prerequisites

Install the following toolchain before continuing:

- Go 1.24+
- Node.js 20+ with `pnpm`
- Docker (or another OCI-compatible builder)
- Helm 3.15+
- `kubectl` and access to a Kubernetes cluster (Kind, Minikube, or an existing environment)

All commands are issued from the repository root unless noted.

## 1. Build UI assets

```bash
cd client
pnpm install
pnpm run build            # Outputs to client/dist
cd ..

# Copy the built UI so Go's embed picks it up during the backend build
rm -rf backend/routes/static
mkdir -p backend/routes/static
cp -R client/dist/. backend/routes/static/
```

`backend/routes/routes.go` embeds everything beneath `backend/routes/static` via `//go:embed`, so keep that directory in sync with the latest front-end build.

## 2. Run the backend locally (optional)

If you simply want to bootstrap the Go server directly on your workstation:

```bash
go run ./backend/main.go --listen localhost:7080 --no-open-browser
```

The server reads kubeconfigs from:

- `~/.kube/*.yaml`
- `~/.kubechat/kubeconfigs/*`

Upload additional configs through the UI if necessary. You can point your local UI (Vite dev server) to the API via the proxy in `client/vite.config.ts`.

## 3. Build the release binary and Docker image

`.goreleaser.Dockerfile` is the container pipeline used by GoReleaser. It:

1. Compiles the backend binary.
2. Downloads architecture-specific `kubectl` and `kubelogin`.
3. Produces a final distroless image containing the `kubechat` binary plus the two CLI tools (needed for cluster interactions inside the pod).

To build the same image manually:

```bash
# Ensure the UI assets are embedded (step 1)
# Build a Linux binary that the Dockerfile will copy (run inside backend/)
cd backend
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ../kubechat .
cd ..

# Build the container image (adjust TARGETARCH as needed)
docker build \
  --build-arg TARGETARCH=amd64 \
  -f .goreleaser.Dockerfile \
  -t kubechat:dev .
```

If you are using Kind, load the image into the cluster directly:

```bash
kind load docker-image kubechat:dev
```

For other clusters (Minikube, remote, etc.) push the image to a registry your cluster can reach.

## 4. Helm deployment (local cluster)

The `charts/kubechat` chart deploys the API, provisions a TLS secret (self-signed by default), and mounts a PVC for user data.

### 4.1 Install/upgrade

```bash
helm upgrade --install kubechat ./charts/kubechat \
  -n kubechat-system --create-namespace \
  --set image.repository=ghcr.io/pramodksahoo/kubechat \
  --set image.tag=dev \
  --set image.pullPolicy=IfNotPresent \
  --set service.listen=":7080" \
  -f charts/kubechat/values-dev.yaml
```

```bash
helm upgrade --install kubechat ./charts/kubechat \
  -n kubechat-system --create-namespace \
  -f charts/kubechat/values-dev.yaml \
  --set service.listen=":7080"
```
Notes:
- If you pushed the image to a registry, set `image.repository` and `image.tag` appropriately.
- Leaving `tls.secretName` blank instructs the chart to create a self-signed TLS secret (hooked via `templates/tls-secret.yaml`). Provide your own secret name to use custom certificates.
- The chart forces `--no-open-browser` inside the container, so it is safe for headless deployments.
- `charts/kubechat/values-dev.yaml` enables a Traefik ingress and sets the required annotations so Traefik talks HTTPS to the pod (`service.serversscheme: https`, `service.serverstransport`, and `service.servername`). It also switches the route to the `websecure` entrypoint and turns on `router.tls`, which means Traefik terminates TLS for browser traffic. If you build your own values file, either carry these annotations forward or expose a plain HTTP listener instead.

### 4.2 Access the UI

Depending on your environment, either expose the service or port-forward:

```bash
kubectl -n kubechat-system port-forward svc/kubechat 7080:7080
```

Visit `https://localhost:7080`. With the self-signed certificate you may need to accept the browser warning.

If you prefer to use the Traefik ingress:

1. Add an `/etc/hosts` entry pointing `kubechat.local` at your cluster (for local setups: `127.0.0.1 kubechat.local`).
2. Browse to `https://kubechat.local` or run  
   `curl -k -H "Host: kubechat.local" https://<traefik-ip>:443/healthz`
   (replace `<traefik-ip>` with the Traefik service endpoint if you’re not port-forwarding it).

### 4.3 Supplying kubeconfigs

The deployment mounts a PVC at `/.kubechat`. Use the UI to upload kubeconfigs, or mount secrets/configmaps by extending the chart (see `values.yaml` for volume patterns).

## 5. Workflow summary

1. `pnpm run build` → copy assets to `backend/routes/static`.
2. `go run ./backend/main.go` for direct development, or build the Linux binary with `go build -C backend -o ../kubechat .`.
3. `docker build -f .goreleaser.Dockerfile -t <image>` for the release image.
4. `helm upgrade --install` with your custom image reference.

Once these steps are scripted you can iterate quickly: rebuild UI, rebuild binary/image, helm upgrade.

## 6. Publishing to GHCR

To publish artefacts the same way the project CI does:

1. Log in to GHCR:  
   `echo "${GHCR_TOKEN}" | docker login ghcr.io -u pramodksahoo --password-stdin`
2. Run GoReleaser (for a tagged release):  
   `goreleaser release --clean`
   - This builds multi-arch images and pushes `ghcr.io/pramodksahoo/kubechat:<tag>`.
   - The `before.hooks` step rebuilds the UI and embeds it automatically.
3. Package & push the Helm chart:
   ```bash
   helm package charts/kubechat
   helm push kubechat-<version>.tgz oci://ghcr.io/pramodksahoo/charts
   ```
4. Consumers can then install with  
   `helm install kubechat oci://ghcr.io/pramodksahoo/charts/kubechat --version <version>`.

--- 

**Why `.goreleaser.Dockerfile` exists:** GoReleaser uses it to produce reproducible, multi-architecture images that include the CLI tooling (`kubectl`, `kubelogin`) required by Kubechat. You can reuse it locally for consistent builds or adapt it if you need additional tooling.
