provider "pyrite" {
  # Your Pyrite Cloud team ID
  # team_id = "your-team-id"

  # Your Pyrite Cloud Team API Key
  # token = "your-api-token"
}

resource "pyrite_service" "example_service" {
  # Your Pyrite Cloud project ID
  # project_id = "your-project-id"

  # Service name
  name = "example-service"

  # Service type: web, pod, worker, or postgres
  type = "web"
}