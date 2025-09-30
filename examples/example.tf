provider "fabric" {
  token    = "<your_fabric_token>"
  endpoint = "https://orchestrator.fabric-testbed.net"
  ssh_key  = "<your_ssh_key>"
}

resource "fabric_slice" "example" {
  name           = "example-slice"
  lease_end_time = "2025-10-01T00:00:00Z"
  ssh_keys       = ["<your_ssh_public_key>"]

  topology {
    nodes = [
      {
        name          = "node1"
        site          = "CLEM"
        type          = "VM"
        image_ref     = "ubuntu-20.04"
        instance_type = "m1.medium"
        cores         = 2
        ram           = 4  # GB
        disk          = 10 # GB
      }
    ]
  }
}
