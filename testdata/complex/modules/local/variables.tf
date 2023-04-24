variable "name" {
    description = "A friendly name for this bucket"
    type = string
}

variable "tags" {
    description = "A map of tags to attach to this bucket"
    type = map(string)
}