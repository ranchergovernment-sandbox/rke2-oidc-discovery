# Demo App Terraform

This is example Terraform to create the appropriate role/policy that has the Trust Provider added. You can use this as a jumping-off point.

## Usage Example

```bash
# Set your URL
URL=oidc.kube.lol

# Get the thumbprint (https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_providers_create_oidc_verify-thumbprint.html)
THUMBPRINT=$(echo | openssl s_client -showcerts -servername $URL -connect $URL:443 2>&1 | sed -ne '/-BEGIN CERTIFICATE-/,/-END CERTIFICATE-/p' | openssl x509 -fingerprint -sha1 -noout | sed 's/SHA1 Fingerprint=//g; s/://g;')

# Create a file for your custom configs
cat <<EOT > /tmp/example.tf
thumbprint_list = ["$THUMBPRINT"]

audience_list = ["rke2"]

oidc_url = "oidc.kube.lol"

service_account_subjects = ["system:serviceaccount:demo:demo-sa"]
EOT

# Initialize Terraform
terraform init

# Plan
terraform plan -var-file=/tmp/example.tf

# Apply
terraform apply -var-file=/tmp/example.tf
```

## Updating the ServiceAccount annotations

Check out the [ServiceAccount](../resources/sa.yaml) and notice the `annotations` section. You'll need to take the output of your executed Terraform and update those annotations to match the role/audience.