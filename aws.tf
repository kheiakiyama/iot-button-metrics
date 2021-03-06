variable "aws_access_key" {}
variable "aws_secret_key" {}
variable "provier_name" {}

variable "region" {
  default = "ap-northeast-1"
}

variable "tags" {
  default = "iot-button-metrics"
}

variable "button_count" {}
variable "lastmodified_key_prifix" {}
variable "metrics_key_prifix" {}
variable "button_prefix" {}
variable "timeout" {}
variable "slack_verified_token" {}

provider "aws" {
  access_key = "${var.aws_access_key}"
  secret_key = "${var.aws_secret_key}"
  region     = "${var.region}"
}
