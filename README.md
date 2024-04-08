# RKE2 OIDC Discovery Tool

I put this together to provide a "proxy" to the two necessary endpoints for OIDC configuration with Rancher's RKE2 (although it should work with most K8s distros). This tool:

* Enables you to expose OIDC endpoints (`/.well-known/openid-configuration` and `/openid/v1/jwks`) without having to enable anonymous authentication nor expose your Kubernetes API directly to the outside world.
* Securely calls the Kubernetes API service with certificate validation and serviceAccount authentication within Kubernetes.
* Returns those values via a RESTful API, much like Kubernetes does.

## Installation

You can install this tool using [Helm](https://helm.sh/docs/intro/install/). Check out the [chart README](./chart/README.md) and [values.yaml](./chart/values.yaml) for further details.

### Add the Helm repo

```bash
helm repo add oidc-discovery https://ranchergovernment-sandbox.github.io/rke2-oidc-discovery/
helm repo update
```

### Ingress/Cert-Manager TLS Example (No Pod-Level TLS)

```bash
# No custom values necessary. Install
helm install -n oidc-discovery --create-namespace oidc-discovery oidc-discovery/rke2-oidc-discovery
```

### Pod-Level TLS Enabled Example 

```bash
# Override default values
cat <<EOT > /tmp/values.yaml
ingress:
  enabled: true
  host: oidc.example.com
  annotations:
    nginx.ingress.kubernetes.io/backend-protocol: "HTTPS"
    nginx.ingress.kubernetes.io/ssl-passthrough: "true"
    nginx.ingress.kubernetes.io/ssl-redirect: "true"

config:
  tls:
    enabled: true
    host: oidc.example.com
EOT

# Install with values
helm install -n oidc-discovery -f /tmp/values.yaml --create-namespace oidc-discovery oidc-discovery/rke2-oidc-discovery
```

## TLS Options

When serving up an OIDC endpoint, it is a requirement that it is served via HTTPS. With that, you have a couple of options:

* Use TLS with your ingress controller & cert-manager.
* Configure your ingress controller for TLS pass-through and allow the chart and application create the certificate for you.
    * To do this, set `config.tls.enabled=true` and modify the other settings under that block in your Helm installation.

## Use Cases

This is incredible useful if you want to do things like [AWS Pod-Level IAM for Service Accounts](https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html), but has other use-cases as well.

## Fleet Example

If you're leveraging Rancher & Fleet, check out the [Fleet directory](./fleet) to an example, fully-fleet-ified rollout of this application against AWS OIDC for pod-level IAM roles.

**NOTE**: You need to have the istio-operator preinstalled. Download `istioctl` and run `istioctl operator init`.

Components:

* **Cert Manager**: Creates self-signed certificate for OIDC endpoints.
* **Istio**: Provides secure ingress point for OIDC endpoints exposed by this service.
* **EKS IAM Webhook**: Operator for managing STS assumed roles at the pod level
* **RKE2 OIDC Discovery**: This tool itself
* **Demo-App**: Demo app to showcase what is going on. **NOTE**: Check out the [terraform](./fleet/demo-app/terraform) directory for an example on how to get our OIDC provider and roles configured.