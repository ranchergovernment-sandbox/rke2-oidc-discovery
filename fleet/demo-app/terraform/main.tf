# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# Terraform configuration

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "5.6.1"
    }
  }
}

provider "aws" {
  region = "us-east-1"
}

resource "aws_iam_openid_connect_provider" "oidc_provider" {
  url             = "https://${var.oidc_url}"
  client_id_list  = var.audience_list
  thumbprint_list = var.thumbprint_list
}

resource "aws_iam_role" "s3_role" {
  name = "oidc_s3_role"

  assume_role_policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        "Effect" : "Allow",
        "Principal" : {
          "Federated" : aws_iam_openid_connect_provider.oidc_provider.arn
        },
        "Action" : "sts:AssumeRoleWithWebIdentity",
        "Condition" : {
          "StringEquals" : {
            "${var.oidc_url}:sub" : var.service_account_subjects,
            "${var.oidc_url}:aud" : var.audience_list
          }
        }
      }
    ]
  })

  managed_policy_arns = [aws_iam_policy.s3_policy.arn]
}

resource "aws_iam_policy" "s3_policy" {
  name = "oidc_s3_policy"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action   = ["s3:ListAllMyBuckets", "s3:ListBucket", "s3:HeadBucket"]
        Effect   = "Allow"
        Resource = var.s3_bucket_regex
      },
    ]
  })
}