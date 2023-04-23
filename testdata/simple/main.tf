resource "aws_s3_bucket" "default" {
    name = "my-test-bucket"
}

resource "aws_s3_object" "obj" {
    bucket = aws_s3_bucket.default.id
    key = "a/magic/object/key"
    source = "${path.module}/outputs.tf"
}