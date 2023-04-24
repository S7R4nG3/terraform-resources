data "aws_caller_identity" "this" {}
data "aws_region" "this" {}

locals {
    account_id = data.aws_caller_identity.this.account_id
    region     = data.aws_region.this.name
    tags = {
        foo = "bar"
        tag = "value"
    }
}

// Local module
module "my_bucket" {
    source = "./modules/local"
    name = "a-fun-bucket"
    tags = local.tags
}

// TF Registry module
module "iam_role" {
  source  = "terraform-aws-modules/iam/aws//modules/iam-assumable-role"
  version = "~> 5.0"

  trusted_role_arns = [
    "arn:aws:iam::${local.account_id}:root",
  ]

  create_role = true

  role_name         = "custom"
  role_requires_mfa = true

  custom_role_policy_arns = [
    "arn:aws:iam::aws:policy/AmazonCognitoReadOnly",
    aws_iam_policy.policy.arn
  ]
  number_of_custom_role_policy_arns = 2
}

data "aws_iam_policy_document" "policy" {
    statement {
        effect = "Allow"
        actions = [
            "s3:ListBucket"
        ]
        resources = [
            module.my_bucket.bucket.arn,
        ]
    }
}

resource "aws_iam_policy" "policy" {
    name = "a-fancy-iam-policy"
    policy = data.aws_iam_policy_document.policy.json
}