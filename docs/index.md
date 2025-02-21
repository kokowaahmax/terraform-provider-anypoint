---
layout: ""
page_title: "Provider: Anypoint"
subcategory: "Mulesoft"
description: |-
  The Anypoint provider provides resources to interact with a Mulesoft's Anypoint Platform API.
---

# Anypoint Provider

The Anypoint provider provides resources to interact with a Mulesoft's Anypoint Platform API.

It is not recommended to use your own account for management of your actions. A user specific to
Terraform is recommended. You can use a Connected App also. Two-factor authentication is not supported in the provider.

This project is maintained by a group of architects 👨‍⚕️ and consultants 🧑‍🔧 at Mulesoft. It implements the Anypoint Platform's terraform provider.

Any contribution is welcome. If you're interested in this project, please get in touch 📧

## Example Usage

```terraform
provider "anypoint" {
  # use either username/pwd or client id/secret to connect to the platform

  username = var.username               # optionally use ANYPOINT_USERNAME env var
  password = var.password               # optionally use ANYPOINT_USERNAME env var

  client_id = var.client_id             # optionally use ANYPOINT_CLIENT_ID env var
  client_secret = var.client_secret     # optionally use ANYPOINT_CLIENT_SECRET env var

  # You may need to change the anypoint control plane: use 'eu' or 'us'
  # by default the control plane is 'us'
  cplane= var.cplane                    # optionnaly use ANYPOINT_CPLANE env var
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- **client_id** (String, Sensitive) the connected app's id
- **client_secret** (String, Sensitive) the connected app's secret
- **cplane** (String) the anypoint control plane
- **password** (String, Sensitive) the user's password
- **username** (String, Sensitive) the user's username