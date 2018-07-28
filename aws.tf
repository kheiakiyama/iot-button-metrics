variable "aws_access_key" {}
variable "aws_secret_key" {}
variable "provier_name" {}

variable "region" {
  default = "ap-northeast-1"
}

variable "tags" {
  default = "iot-button-metrics"
}

provider "aws" {
  access_key = "${var.aws_access_key}"
  secret_key = "${var.aws_secret_key}"
  region     = "${var.region}"
}
