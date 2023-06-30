# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# Output variable definitions

output "role_arn" {
  description = "ARN of the created Role (required in ServiceAccount annotations)"
  value       = aws_iam_role.s3_role.arn
}

output "audience_list" {
  description = "List of audienced (required in ServiceAccount annotations)"
  value       = var.audience_list
}
