
#################################################################
# S3 BUCKETS
#################################################################
module "s3_extract" {
  source = "github.com/gsa-tts/terraform-cloudgov//s3?ref=v0.9.1"
  cf_org_name      = var.cf_org
  cf_space_name    = var.cf_space
  name             = "s3_extract"
  s3_plan_name     = "basic"
  tags             = ["s3"]
}

module "s3_fetch" {
  source = "github.com/gsa-tts/terraform-cloudgov//s3?ref=v0.9.1"
  cf_org_name      = var.cf_org
  cf_space_name    = var.cf_space
  name             = "s3_fetch"
  s3_plan_name     = "basic"
  tags             = ["s3"]
}

module "s3_meta" {
  source = "github.com/gsa-tts/terraform-cloudgov//s3?ref=v0.9.1"
  cf_org_name      = var.cf_org
  cf_space_name    = var.cf_space
  name             = "s3_meta"
  s3_plan_name     = "basic"
  tags             = ["s3"]
}

module "s3_backup" {
  source = "github.com/gsa-tts/terraform-cloudgov//s3?ref=v0.9.1"
  cf_org_name      = var.cf_org
  cf_space_name    = var.cf_space
  name             = "s3_backup"
  s3_plan_name     = "basic"
  tags             = ["s3"]
}


