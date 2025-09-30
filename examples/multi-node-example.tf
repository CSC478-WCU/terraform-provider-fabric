provider "fabric" {
  token    = "<your_fabric_token>"
  endpoint = "https://orchestrator.fabric-testbed.net"
  ssh_key  = "<your_ssh_key>"
}

resource "fabric_slice" "multi-node" {
  name           = "multi-node-slice"
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
      },
      {
        name          = "node2"
        site          = "NCSA"
        type          = "VM"
        image_ref     = "ubuntu-20.04"
        instance_type = "m1.medium"
        cores         = 2
        ram           = 4  # GB
        disk          = 10 # GB
      }
    ]

    links = [
      {
        name   = "link1"
        source = "node1"
        target = "node2"
      }
    ]
  }
}
