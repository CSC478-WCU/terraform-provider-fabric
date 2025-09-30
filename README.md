```hcl
terraform {
  required_providers {
    fabric = {
      source  = "csc478-wcu/fabric"
      version = ">= 0.1.0"
    }
  }
}

provider "fabric" {
  token    = var.fabric_token
  endpoint = var.fabric_endpoint
  ssh_key  = var.fabric_ssh_key
}

data "fabric_sites" "all" {}

resource "fabric_slice" "sample" {
  name = var.slice_name

  topology = {
    nodes = [
      {
        name          = "primary"
        site          = var.primary_site
        type          = "VM"
        image_ref     = var.image_ref
        instance_type = var.instance_type
        cores         = var.node_cores
        ram           = var.node_ram
        disk          = var.node_disk
      },
      {
        name          = "peer"
        site          = var.secondary_site == "" ? var.primary_site : var.secondary_site
        type          = "VM"
        image_ref     = var.image_ref
        instance_type = var.instance_type
        cores         = var.node_cores
        ram           = var.node_ram
        disk          = var.node_disk
      }
    ]

    links = [
      {
        name   = "primary-peer"
        source = "primary"
        target = "peer"
      }
    ]
  }

  ssh_keys = var.fabric_ssh_key == "" ? [] : [var.fabric_ssh_key]
}

output "available_sites" {
  description = "Sample of discovered sites for reference"
  value       = data.fabric_sites.all.sites
}

output "slice_state" {
  value = fabric_slice.sample.state
}
```
```
variable "fabric_token" {
  description = "FABRIC API token. Can also be set via FABRIC_TOKEN environment variable."
  type        = string
}

variable "fabric_endpoint" {
  description = "FABRIC orchestrator API endpoint."
  type        = string
  default     = "https://orchestrator.fabric-testbed.net"
}

variable "fabric_ssh_key" {
  description = "SSH public key to inject into the slice. Leave empty to rely on provider default."
  type        = string
  default     = ""
}

variable "slice_name" {
  description = "Name for the sample slice."
  type        = string
  default     = "fabric-sample"
}

variable "primary_site" {
  description = "Primary FABRIC site for the first node."
  type        = string
  default     = "CLEM"
}

variable "secondary_site" {
  description = "Optional second site for the peer node. Leave empty to reuse the primary site."
  type        = string
  default     = ""
}

variable "image_ref" {
  description = "Image reference to use for nodes."
  type        = string
  default     = "default_rocky_8,qcow2"
}

variable "instance_type" {
  description = "Instance type to use for nodes."
  type        = string
  default     = "fabric.c2.m2.d10"
}

variable "node_cores" {
  description = "CPU cores for each node."
  type        = number
  default     = 2
}

variable "node_ram" {
  description = "RAM (GB) for each node."
  type        = number
  default     = 2
}

variable "node_disk" {
  description = "Disk size (GB) for each node."
  type        = number
  default     = 10
}
```
```hcl
# Copy this file to terraform.tfvars (gitignored) and populate with real values
# or set the equivalent environment variables before running Terraform.

fabric_endpoint  = "https://orchestrator.fabric-testbed.net"
slice_name       = "fabric-dev-slice"
primary_site     = "CLEM"
secondary_site   = "NCSA"
image_ref        = "default_rocky_8,qcow2"
instance_type    = "fabric.c2.m2.d10"
node_cores       = 2
node_ram         = 2
fabric_token    = ""
ssh_key         = ""
```
