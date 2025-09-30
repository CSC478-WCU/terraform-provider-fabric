provider "fabric" {
  token    = "<your_fabric_token>"
  endpoint = "https://orchestrator.fabric-testbed.net"
  ssh_key  = "<your_ssh_key>"
}

data "fabric_sites" "all_sites" {}

output "all_sites" {
  value = data.fabric_sites.all_sites
}
