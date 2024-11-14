
data "cloudfoundry_domain" "public" {
  name = "app.cloud.gov"
}

data "cloudfoundry_space" "app_space" {
  org_name = "sandbox-gsa"
  name     = "matthew.jadud"
}

#################################################################
# POSTGRES
#################################################################

module "database" {
  source = "github.com/gsa-tts/terraform-cloudgov//database?ref=v0.9.1"
  cf_org_name      = "sandbox-gsa"
  cf_space_name    = "matthew.jadud"
  name             = "jemison-queues-db"
  recursive_delete = false
  tags             = ["rds"]
  rds_plan_name    = "micro-psql"
}

module "database" {
  source = "github.com/gsa-tts/terraform-cloudgov//database?ref=v0.9.1"
  cf_org_name      = "sandbox-gsa"
  cf_space_name    = "matthew.jadud"
  name             = "jemison-db"
  recursive_delete = false
  tags             = ["rds"]
  rds_plan_name    = "micro-psql"
}

#################################################################
# S3 BUCKETS
#################################################################
module "s3-private-extract" {
  source = "github.com/gsa-tts/terraform-cloudgov//s3?ref=v0.9.1"
  cf_org_name      = "sandbox-gsa"
  cf_space_name    = "matthew.jadud"
  name             = "extract"
  s3_plan_name     = "basic"
  recursive_delete = false
  tags             = ["s3"]
}

module "s3-private-fetch" {
  source = "github.com/gsa-tts/terraform-cloudgov//s3?ref=v0.9.1"
  cf_org_name      = "sandbox-gsa"
  cf_space_name    = "matthew.jadud"
  name             = "fetch"
  s3_plan_name     = "basic"
  recursive_delete = false
  tags             = ["s3"]
}

module "s3-private-serve" {
  source = "github.com/gsa-tts/terraform-cloudgov//s3?ref=v0.9.1"
  cf_org_name      = "sandbox-gsa"
  cf_space_name    = "matthew.jadud"
  name             = "serve"
  s3_plan_name     = "basic"
  recursive_delete = false
  tags             = ["s3"]
}


#################################################################
# FETCH
#################################################################

resource "cloudfoundry_route" "fetch_route" {
  space    = data.cloudfoundry_space.app_space.id
  domain   = data.cloudfoundry_domain.public.id
  hostname = "jemison-fetch"
}
resource "cloudfoundry_app" "fetch" {
  name                 = "fetch"
  space                = data.cloudfoundry_space.app_space.id
  buildpacks            = ["https://github.com/cloudfoundry/apt-buildpack", "https://github.com/cloudfoundry/binary-buildpack.git"]
  path                 = "zips/fetch.zip"
  source_code_hash     = filesha256("zips/fetch.zip")
  disk_quota           = var.disk_quota_l
  memory               = var.service_fetch_ram
  instances            = 1
  strategy             = "rolling"
  timeout              = 200
  health_check_type    = "port"
  health_check_timeout = 180
  health_check_http_endpoint = "/heartbeat"
    service_binding {
    service_instance = module.s3-private-fetch.bucket_id
  }

  service_binding {
    service_instance = module.database.instance_id
  }

  routes {
    route = cloudfoundry_route.fetch_route.id
  }

  environment = {
    ENV = "SANDBOX"
    API_KEY = "${var.api_key}"
    DEBUG_LEVEL = "${var.zap_debug_level}"
    GIN_MODE = "${var.gin_debug_level}"
    # REQUESTS_CA_BUNDLE = "/etc/ssl/certs/ca-certificates.crt"
  }
}

#################################################################
# EXTRACT
#################################################################
resource "cloudfoundry_app" "extract" {
  name                 = "extract"
  space                = data.cloudfoundry_space.app_space.id
  buildpacks            = ["https://github.com/cloudfoundry/apt-buildpack", "https://github.com/cloudfoundry/binary-buildpack.git"]
  path                 = "zips/extract.zip"
  source_code_hash     = filesha256("zips/extract.zip")
  disk_quota           = var.disk_quota_l
  memory               = var.service_extract_ram
  instances            = 1
  strategy             = "rolling"
  timeout              = 200
  health_check_type    = "port"
  health_check_timeout = 180
  health_check_http_endpoint = "/heartbeat"
  service_binding {
    service_instance = module.s3-private-extract.bucket_id
  }
  service_binding {
    service_instance = module.s3-private-fetch.bucket_id
  }
  service_binding {
    service_instance = module.database.instance_id
  }

  # routes {
  #   route = cloudfoundry_route.serve_route.id
  # }

  environment = {
    ENV = "SANDBOX"
    API_KEY = "${var.api_key}"
    DEBUG_LEVEL = "${var.zap_debug_level}"
    GIN_MODE = "${var.gin_debug_level}"
    # REQUESTS_CA_BUNDLE = "/etc/ssl/certs/ca-certificates.crt"
  }
}

#################################################################
# PACK
#################################################################

resource "cloudfoundry_app" "pack" {
  name                 = "pack"
  space                = data.cloudfoundry_space.app_space.id
  buildpacks            = ["https://github.com/cloudfoundry/apt-buildpack", "https://github.com/cloudfoundry/binary-buildpack.git"]
  path                 = "zips/pack.zip"
  source_code_hash     = filesha256("zips/pack.zip")
  disk_quota           = var.disk_quota_l
  memory               = var.service_pack_ram
  instances            = 1
  strategy             = "rolling"
  timeout              = 200
  health_check_type    = "port"
  health_check_timeout = 180
  health_check_http_endpoint = "/heartbeat"
  service_binding {
    service_instance = module.s3-private-extract.bucket_id
  }
  service_binding {
    service_instance = module.s3-private-serve.bucket_id
  }

  service_binding {
    service_instance = module.database.instance_id
  }

  # routes {
  #   route = cloudfoundry_route.serve_route.id
  # }

  environment = {
    ENV = "SANDBOX"
    API_KEY = "${var.api_key}"
    DEBUG_LEVEL = "${var.zap_debug_level}"
    GIN_MODE = "${var.gin_debug_level}"
    # REQUESTS_CA_BUNDLE = "/etc/ssl/certs/ca-certificates.crt"
  }
}


#################################################################
# SERVE
#################################################################

resource "cloudfoundry_route" "serve_route" {
  space    = data.cloudfoundry_space.app_space.id
  domain   = data.cloudfoundry_domain.public.id
  hostname = "jemison"
}
resource "cloudfoundry_app" "serve" {
  name                 = "serve"
  space                = data.cloudfoundry_space.app_space.id
  buildpacks            = ["https://github.com/cloudfoundry/apt-buildpack", "https://github.com/cloudfoundry/binary-buildpack.git"]
  path                 = "zips/serve.zip"
  source_code_hash     = filesha256("zips/serve.zip")
  disk_quota           = var.disk_quota_l
  memory               = var.service_serve_ram
  instances            = 1
  strategy             = "rolling"
  timeout              = 200
  health_check_type    = "port"
  health_check_timeout = 180
  health_check_http_endpoint = "/heartbeat"
    service_binding {
    service_instance = module.s3-private-fetch.bucket_id
  }
  service_binding {
    service_instance = module.s3-private-serve.bucket_id
  }
  service_binding {
    service_instance = module.database.instance_id
  }

  routes {
    route = cloudfoundry_route.serve_route.id
  }

  environment = {
    ENV = "SANDBOX"
    API_KEY = "${var.api_key}"
    DEBUG_LEVEL = "${var.zap_debug_level}"
    GIN_MODE = "${var.gin_debug_level}"
    # REQUESTS_CA_BUNDLE = "/etc/ssl/certs/ca-certificates.crt"
  }
}

#################################################################
# WALK
#################################################################

resource "cloudfoundry_app" "walk" {
  name                 = "walk"
  space                = data.cloudfoundry_space.app_space.id
  buildpacks            = ["https://github.com/cloudfoundry/apt-buildpack", "https://github.com/cloudfoundry/binary-buildpack.git"]
  path                 = "zips/walk.zip"
  source_code_hash     = filesha256("zips/walk.zip")
  disk_quota           = var.disk_quota_m
  memory               = var.service_walk_ram
  instances            = 1
  strategy             = "rolling"
  timeout              = 200
  health_check_type    = "port"
  health_check_timeout = 180
  health_check_http_endpoint = "/heartbeat"
  service_binding {
    service_instance = module.s3-private-fetch.bucket_id
  }

  service_binding {
    service_instance = module.database.instance_id
  }

  # routes {
  #   route = cloudfoundry_route.serve_route.id
  # }

  environment = {
    ENV = "SANDBOX"
    API_KEY = "${var.api_key}"
    DEBUG_LEVEL = "${var.zap_debug_level}"
    GIN_MODE = "${var.gin_debug_level}"
    # REQUESTS_CA_BUNDLE = "/etc/ssl/certs/ca-certificates.crt"
  }
}