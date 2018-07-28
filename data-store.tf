resource "aws_s3_bucket" "metrics" {
  bucket = "${var.tags}-${var.provier_name}"
  acl    = "private"

  tags {
    Name        = "${var.tags}"
    Environment = "Production"
  }
}
