# kubechat helm chart

Kubechat allows you to manage Kubernetes clusters. This Helm chart simplifies the installation and configuration of Kubechat.

## Installation

To install the kubechat chart using Helm, run the following command:

```bash
helm install kubechat oci://ghcr.io/pramodksahoo/charts/kubechat -n kubechat-system --create-namespace
```

### Notes:

- **Default Setup**: By default, Kubechat runs on port `8443` with self-signed certificates.
- **Namespace**: A new namespace `kubechat-system` will be created automatically if it doesn't exist.

### Using Custom TLS Certificates

To use your own TLS certificates instead of the default self-signed ones:

1. **Create a Kubernetes Secret**: Store your TLS certificate and key in a secret.

   ```bash
   kubectl create namespace kubechat-system
   kubectl -n kubechat-system create secret tls kubechat-tls-secret --cert=tls.crt --key=tls.key
   ```

2. **Install Kubechat with your certificates**:

   ```bash
   helm install kubechat oci://ghcr.io/pramodksahoo/charts/kubechat \
     -n kubechat-system --version v1.0.0 --create-namespace \
     --set tls.secretName=kubechat-tls-secret
   ```

### Using a Custom Service Account

By default, the chart creates a service account with `admin` RBAC permissions in the release namespace. If you'd like Kubechat to use an existing service account, you can disable the creation of a new one.

1. **Install Kubechat with an existing service account**:

   ```bash
   helm install kubechat oci://ghcr.io/pramodksahoo/charts/kubechat \
     -n kubechat-system --version v1.0.0 --create-namespace \
     --set serviceAccount.create=false \
     --set serviceAccount.name=<yourServiceAccountName>
   ```

## Upgrading the Chart

To upgrade to a newer version of the chart, run the following command:

```bash
helm upgrade kubechat oci://ghcr.io/pramodksahoo/charts/kubechat \
  -n kubechat-system --version v1.0.0
```

## Configuration Parameters

The following are some key configuration parameters you can customize when installing the chart:

| Parameter               | Description                                                                                       | Default  |
|-------------------------|---------------------------------------------------------------------------------------------------|----------|
| `tls.secretName`         | Kubernetes secret name containing your TLS certificate and key. Must be in the `kubechat-system` namespace. | `""`     |
| `service.port`           | The HTTPS port number Kubechat listens on.                                                        | `8443`   |
| `serviceAccount.create`  | Set to `false` if you want to use an existing service account.                                     | `true`   |
| `serviceAccount.name`    | Name of the service account to use (if `serviceAccount.create=false`).                            | `""`     |

For a complete list of configurable parameters, refer to the values file or documentation.
