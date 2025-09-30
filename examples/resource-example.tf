provider "fabric" {
  token    = "<your_fabric_token>"
  endpoint = "https://orchestrator.fabric-testbed.net"
  ssh_key  = "<your_ssh_key>"
}

data "fabric_resources" "available_resources" {}

output "available_resources" {
  value = data.fabric_resources.available_resources
}
