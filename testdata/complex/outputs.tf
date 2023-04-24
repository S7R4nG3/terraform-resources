output "bucket" {
    value = module.my_bucket.bucket
}

output "role" {
    value = module.iam_role
}