
resource "aws_s3_bucket" "default" {
    bucket = var.name
    tags = var.tags
}

