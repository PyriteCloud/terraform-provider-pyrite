provider "pyrite" {
  # Your Pyrite Cloud team ID
  # team_id = "your-team-id"

  # Your Pyrite Cloud Team API Key
  # token = "your-api-token"
}

resource "pyrite_service_environment" "example_service_environment_with_image" {
  # Your Pyrite Cloud project ID
  # project_id = "your-project-id"

  # Service name
  name = "example-service"

  # Deployment environment
  environment = "dev"

  # Service type: web, pod, worker or postgres
  type = "web"

  docker_config {
    # Deployment source
    source_type = "image"

    # Container runtime
    runtime = "docker"

    # Compute plan
    # Available plans: sNano, sMicro, sSmall, sMedium, sLarge
    plan = "sNano"

    # Service access configuration
    is_private      = false
    is_privileged   = false
    with_project_env = false

    # Docker image to deploy
    image {
      ref = "nginx:latest"

      # Optional: use a private registry
      # registry_id = "your-registry-id"
    }

    # Optional container command and arguments
    # args    = null
    # command = null

    # Environment variables.
    # env = ""

    # Network ports exposed by the service
    ports = [
      {
        port     = 80
        protocol = "http1"
        path     = "/"
      }
    ]

    # Deployment regions and replica configuration
    regions = [
      {
        region       = "eu-central-1"
        min_replicas = 1
        max_replicas = 2
      }
    ]

    # Optional files to mount into the container
    # files {
    #   path        = "/app/config.yaml"
    #   content     = "key: value"
    #   permissions = "0644"
    # }

    # Optional persistent volumes
    # volumes {
    #   mount_path     = "/data"
    #   team_volume_id = "your-volume-id"
    # }

    # Optional health checks
    # health_checks {
    #   port          = 80
    #   protocol      = "http1"
    #   path          = "/"
    #   initial_delay = 10
    #   interval      = 30
    #   timeout       = 5
    #   max_failures  = 3
    # }
  }
}

resource "pyrite_service_environment" "example_service_env_with_git" {
  # Your Pyrite Cloud project ID
  project_id = "your-project-id"

  # Service name
  name = "example-git-service"

  # Deployment environment
  environment = "dev"

  # Service type: web, pod, or worker
  type = "web"

  docker_config {
    # Deployment source
    source_type = "git"

    # Container runtime
    runtime = "docker"

    # Compute plan
    plan = "sNano"

    # Service access configuration
    is_private       = false
    is_privileged    = false
    with_project_env = false

    git {
      # Git repository URL
      url = "https://github.com/example/my-app.git"

      # Git branch to deploy
      branch = "main"

      # Optional commit SHA
      # sha = "a1b2c3d4..."

      # Enable Docker image build from the repository
      with_build = true

      build {
        # Build system
        # Defaults to "buildkit"
        builder = "buildkit"

        # Docker build context
        # Defaults to "."
        context = "."

        # Dockerfile path
        # Defaults to "Dockerfile"
        dockerfile_path = "Dockerfile"
      }
    }

    # Optional container command and arguments
    # args    = null
    # command = null

    # Environment variables
    # env = ""

    # Ports exposed by the service
    ports = [
      {
        port     = 80
        protocol = "http1"
        path     = "/"
      }
    ]

    # Deployment regions
    regions = [
      {
        region       = "eu-central-1"
        min_replicas = 1
        max_replicas = 2
      }
    ]

    # Optional files
    # files {
    #   path        = "/app/config.yaml"
    #   content     = "key: value"
    #   permissions = "0644"
    # }

    # Optional persistent volumes
    # volumes {
    #   mount_path     = "/data"
    #   team_volume_id = "your-volume-id"
    # }

    # Optional health checks
    # health_checks {
    #   port          = 80
    #   protocol      = "http1"
    #   path          = "/"
    #   initial_delay = 10
    #   interval      = 30
    #   timeout       = 5
    #   max_failures  = 3
    # }
  }
}

resource "pyrite_service_environment" "example_service_env_with_postgres" {
  # Your Pyrite Cloud project ID
  project_id = "your-project-id"

  # Service name
  name = "example-postgres"

  # Deployment environment
  environment = "dev"

  # PostgreSQL service
  type = "postgres"

  postgres_config {
    # PostgreSQL version
    version = "16"

    # PostgreSQL compute plan
    # Available plans:
    # sSmallDev, sSmallProd,
    # sMediumDev, sMediumProd,
    # sLargeProd
    plan = "sSmallDev"

    # Deployment region
    region = "eu-central-1"

    # Storage size
    size = 20

    # PostgreSQL password
    password = "your-secure-password"
  }
}