variable "thumbprint_list" {
  description = "List of cert thumbprints for Identity Provider"
  type        = list(string)
  default     = []
}

variable "audience_list" {
  description = "List of audiences for Identity Provider"
  type        = list(string)
  default     = []
}

variable "oidc_url" {
  description = "URL for Identity Provider"
  type        = string
  default     = ""
}

variable "s3_bucket_regex" {
  description = "Regex for S3 policy"
  type        = string
  default     = "*"
}

variable "service_account_subjects" {
  description = "List of Service Account Subjects. Should be in format 'system:serviceaccount:example-namespace:example-service-account'"
  type        = list(string)
  default     = ["system:serviceaccount:demo:demo-sa"]
}