variable "image_tag" {
  type    = string
  default = "latest"
}

variable "repository" {
  type    = string
}


job "auth-service-go" {
  datacenters = ["dc1"]
  type        = "service"

  update {
    max_parallel     = 1
    min_healthy_time = "15s"
    auto_revert      = true
  }

  group "auth-service-go" {
    count = 2

    disconnect {
      lost_after = "40s"
    }

    service {
      name = "dashboard-auth-service"
      tags = [
          "traefik.enable=true",
          "traefik.http.routers.dashboard-auth-service.rule=Host(`api.nightcore.space`) && PathPrefix(`/auth`)",
          "traefik.http.routers.dashboard-auth-service.priority=20",
          "traefik.http.routers.dashboard-auth-service.entrypoints=websecure",
          "traefik.http.routers.dashboard-auth-service.service=dashboard-auth-service",
          "traefik.http.services.dashboard-auth-service.loadbalancer.server.port=5001",
          "traefik.http.routers.dashboard-auth-service.tls=true",
          "traefik.http.middlewares.auth-ratelimit.ratelimit.average=2",
          "traefik.http.middlewares.auth-ratelimit.ratelimit.period=1s",
          "traefik.http.middlewares.auth-ratelimit.ratelimit.burst=2",
          "traefik.http.routers.dashboard-auth-service.middlewares=auth-ratelimit",
      ]
    }

    task "auth-service-go" {
      driver = "docker"

      vault {
        role = "runner-nightcore-auth-service"
      }

      identity {
        name = "vault_default"
        aud  = ["vault.io"]
        ttl  = "1h"
      }

      template {
        data = <<EOT
{{ with secret "secret/data/ci/github-registry" }}
REGISTRY_USERNAME={{ .Data.data.username }}
REGISTRY_TOKEN={{ .Data.data.token }}
{{ end }}
EOT
        destination = "secrets/registry.env"
        env         = true
        change_mode = "restart"
      }

      resources {
        cpu    = 250
        memory = 150
      }

      config {
        image = "ghcr.io/${var.repository}:${var.image_tag}"

        network_mode = "host"

        auth {
          username       = "${REGISTRY_USERNAME}"
          password       = "${REGISTRY_TOKEN}"
        }
      }

      template {
        data = <<EOT
{{ with secret "secret/data/ci/repos/nightcore-auth-service" }}
API_PORT={{ .Data.data.API_PORT }}
API_HOST={{ .Data.data.API_HOST }}
API_DOMAIN={{ .Data.data.API_DOMAIN }}
DASHBOARD_FRONTEND_URI={{ .Data.data.DASHBOARD_FRONTEND_URI }}
JWT_PUBLIC_KEY={{ .Data.data.JWT_PUBLIC_KEY }}
JWT_PRIVATE_KEY={{ .Data.data.JWT_PRIVATE_KEY }}
JWT_ALGORITHM={{ .Data.data.JWT_ALGORITHM }}
DISCORD_AUTH_CLIENT_ID={{ .Data.data.DISCORD_AUTH_CLIENT_ID }}
DISCORD_AUTH_CLIENT_SECRET={{ .Data.data.DISCORD_AUTH_CLIENT_SECRET }}
DISCORD_AUTH_REDIRECT_URI={{ .Data.data.DISCORD_AUTH_REDIRECT_URI }}
{{ end }}
EOT
        destination = "secrets/auth-service.env"
        env         = true
      }

      template {
        data = <<EOT
{{ with secret "secret/data/keydb" }}
REDIS_PASSWORD={{ .Data.data.password }}
REDIS_HOST={{ .Data.data.host }}
{{ end }}
EOT
        destination = "secrets/keydb.env"
        env         = true
      }

      logs {
        max_files     = 3
        max_file_size = 10
      }

    }
  }
}