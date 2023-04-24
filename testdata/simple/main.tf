module "naming" {
    source = "./modules/naming"
    app_name = "example"
    microservice = "tester"
    tags = {
        Tag = "value"
    }
}

resource "aws_s3_bucket" "default" {
    bucket = "${module.naming.full_name}-test"
}

resource "aws_s3_object" "obj" {
    bucket = aws_s3_bucket.default.id
    key = "a/magic/object/key"
    source = "${path.module}/outputs.tf"
}