resource "cloudfoundry_route" "admin_route" {
  space    = var.app_space_id # data.cloudfoundry_space.app_space.id
  domain   = var.domain_id # data.cloudfoundry_domain.public.id
  hostname = "jemison-admin-${var.space_name}" #FIXME - ${spacename}
}

resource "cloudfoundry_app" "admin" {
  name                 = "admin"
  space                = var.app_space_id # data.cloudfoundry_space.app_space.id
  buildpacks            = ["https://github.com/cloudfoundry/apt-buildpack", "https://github.com/cloudfoundry/binary-buildpack.git"]
  path                 = "${path.module}/../app.tar.gz"
  source_code_hash     = filesha256("${path.module}/../app.tar.gz")
  disk_quota           = var.disk_quota
  memory               = var.memory
  instances            = var.instances
  strategy             = "rolling"
  timeout              = 200
  health_check_type    = "port"
  health_check_timeout = 180
  health_check_http_endpoint = "/api/heartbeat"

  service_binding {
    service_instance = var.db_queue_id
  }

  service_binding {
    service_instance = var.db_search_id
  }

  service_binding {
    service_instance = var.db_work_id
  }

  routes {
    route = cloudfoundry_route.admin_route.id # cloudfoundry_route.admin_route.id
  }
  
  # environment = {
  #   ENV = var.cf_space
  # }

  # environment = {
  #   ENV = "SANDBOX"
  #   API_KEY = "${var.api_key}"
  #   GIN_MODE = "${var.gin_debug_level}"
  # }
}
