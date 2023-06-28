# RKE2 OIDC Discovery Tool

I put this together to provide a "proxy" to the 2 necessary endpoints for OIDC configuration with Rancher's RKE2 (although it should work with most K8s distros). This tool:

* Enables you to expose OIDC endpoints (`/.well-known/openid-configuration` and `/openid/v1/jwks`) without having to enable anonymous authentication nor expose your Kubernetes API directly to the outside world.
* Securely calls the Kubernetes API service with certificate validation and serviceAccount authentication within Kubernetes.
* Returns those values via a RESTful API, much like Kubernetes does.

## Installation

You can install this tool using [Helm](https://helm.sh/docs/intro/install/). Check out the [chart README](./chart/README.md) and [values.yaml](./chart/values.yaml) for further details.

An example installation would be:
```bash
# Clone repo
git clone https://github.com/atoy3731/rke2-oidc-discovery.git

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

# Now install
helm install -n oidc-discovery --create-namespace -f /tmp/values.yaml oidc-discovery rke2-oidc-discovery/chart/
```

## TLS Options

When serving up an OIDC endpoint, it is a requirement that it is served via HTTPS. With that, you have a couple of options:

* Use TLS with your ingress controller & cert-manager.
* Configure your ingress controller for TLS pass-through and allow the chart and application create the certificate for you.
    * To do this, set `config.tls.enabled=true` and modify the other settings under that block in your Helm installation.

## Use Cases

This is incredible useful if you want to do things like [AWS Pod-Level IAM for Service Accounts](https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html), but has other use-cases as well.