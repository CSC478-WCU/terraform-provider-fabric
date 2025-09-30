<!-- markdownlint-disable first-line-h1 no-inline-html -->
<a href="https://terraform.io">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset=".github/terraform_logo_dark.svg">
    <source media="(prefers-color-scheme: light)" srcset=".github/terraform_logo_light.svg">
    <img src=".github/terraform_logo_light.svg" alt="Terraform logo" title="Terraform" align="right" height="50">
  </picture>
</a>

# Terraform Provider for FABRIC Testbed

This Terraform provider allows you to manage resources on the **FABRIC Testbed**. It supports creating, reading, updating, and deleting slices (multi-node experiments) as well as querying available resources and sites.

> **Disclaimer**: This provider is **not maintained** by the official FABRIC Testbed team. It is an open-source project developed at **West Chester University** to provide a convenient way to manage resources on the FABRIC Testbed using Terraform.

> **Note:** This is built off the Fabric Orchestrator Client here: https://github.com/CSC478-WCU/fabric-orchestrator-go-client

## Table of Contents

- [Terraform Provider for FABRIC Testbed](#terraform-provider-for-fabric-testbed)
  - [Table of Contents](#table-of-contents)
  - [Overview](#overview)
  - [Requirements](#requirements)
  - [Installation](#installation)
  - [Token Authentication](#token-authentication)
    - [Extracting the `id_token`](#extracting-the-id_token)
    - [Passing the `id_token` to the Provider](#passing-the-id_token-to-the-provider)
  - [Usage Examples](#usage-examples)
    - [Creating a Slice (VMs)](#creating-a-slice-vms)
  - [Data Sources](#data-sources)
    - [`fabric_resources`](#fabric_resources)
    - [`fabric_sites`](#fabric_sites)
  - [Roadmap](#roadmap)
  - [Contributing](#contributing)
  - [License](#license)

---

## Overview

This provides IaC (Infrastructure As Code) integration with the FABRIC Testbed, allowing you to automate the creation and management of resources on the testbed via Terraform. You can manage:

- **Slices**: Create, read, update, and delete testbed slices (multi-node topologies).
- **Resources**: List available resources such as compute instances, storage, etc.
- **Sites**: Query information about available FABRIC testbed sites.

---

## Requirements

- **Terraform**: v1.0 or higher
- **FABRIC API Token**: A valid token for accessing the FABRIC orchestrator.
- **FABRIC SSH Key**: An SSH key to access the nodes in the slices.

---

## Installation

To use this provider, define it in your Terraform configuration:

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
```

You can set the values for `FABRIC_TOKEN` and `FABRIC_SSH_KEY` via environment variables or in your `terraform.tfvars`.

---

## Token Authentication

To authenticate with the FABRIC testbed, you will need to use an `id_token` obtained from the [FABRIC Credentials Manager](https://cm.fabric-testbed.net/). After logging into the Credentials Manager, you can create a token by selecting the **id_token** option.

### Extracting the `id_token`

Once logged into the [FABRIC Credentials Manager](https://cm.fabric-testbed.net/), create a token by selecting the **id_token** option. You will be presented with a response similar to the following:

```json
{
  "comment": "Created via GUI",
  "created_at": "2025-09-30 17:59:51 +0000",
  "id_token": "eyJhbGciOiJSUzI1NiIsImtpZCI6Inl1ZmVrV...", // THIS VALUE HERE
  "refresh_token": "NB2HI4DTHIXS6Y3JNRXWO33OFZXXEZZPN5QXK5DIGIXTCODFGZQTQZBZGN...",
  "state": "Valid"
}
```

Here, the `id_token` is what you will use in the next step.

### Passing the `id_token` to the Provider

In your `terraform.tfvars` or environment variables, you can pass the `id_token` as follows:

```hcl
provider "fabric" {
  token    = var.fabric_token  # This should be the id_token value you obtained
  endpoint = var.fabric_endpoint
  ssh_key  = var.fabric_ssh_key
}
```

Alternatively, you can set the environment variable for `FABRIC_TOKEN` to the `id_token` value:

```bash
export FABRIC_TOKEN="eyJhbGciOiJSUzI1NiIsImtpZCI6Inl1ZmVrV..."
```

This will allow your Terraform provider to authenticate against the FABRIC testbed.

---

## Usage Examples

### Creating a Slice (VMs)

```hcl
resource "fabric_slice" "my_slice" {
  name          = "my-slice"
  lease_end_time = "2023-12-01T00:00:00Z"  # Optional, defaults to 24 hours from now

  topology {
    nodes = [
      {
        name          = "node1"
        site          = "CLEM"
        type          = "VM"
        image_ref     = "default-ubuntu"
        instance_type = "m1.small"
        cores         = 2
        ram           = 4 # GB
        disk          = 10 # GB
      },
      {
        name          = "node2"
        site          = "NCSA"
        type          = "VM"
        image_ref     = "default-ubuntu"
        instance_type = "m1.medium"
        cores         = 2
        ram           = 4 # GB
        disk          = 20 # GB
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
```

---

## Data Sources

### `fabric_resources`

List available resources in FABRIC:

```hcl
data "fabric_resources" "available_resources" {}
```

### `fabric_sites`

Get information about FABRIC testbed sites:

```hcl
data "fabric_sites" "all_sites" {}
```

---

## Roadmap

- [ ] Improve state handling (reduce reliance on `terraform apply` refresh).
- [ ] Add support for more FABRIC resources (storage, networking, etc.).
- [ ] Expand documentation and usage examples.

---

## Contributing

Feel free to open an issue or submit a pull request. All contributions are welcome.

---

## License

MIT License.
