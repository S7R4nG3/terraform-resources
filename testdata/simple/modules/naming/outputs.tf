output "full_name" {
    value = "${var.app_name}-${var.microservice}"
}

output "tags" {
    value = var.tags
}