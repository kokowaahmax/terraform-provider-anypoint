---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "anypoint_rolegroup Data Source - terraform-provider-anypoint"
subcategory: ""
description: |-
  Reads a specific `rolegroup` in your business group.
---

# anypoint_rolegroup (Data Source)

Reads a specific `rolegroup` in your business group.

## Example Usage

```terraform
data "anypoint_rolegroup" "result" {
  id = "ROLEGROUP_ID"
}

output "rg" {
  value = data.anypoint_rolegroup.result
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- **id** (String) The ID of this resource.
- **org_id** (String)

### Read-Only

- **created_at** (String)
- **description** (String)
- **editable** (Boolean)
- **external_names** (List of String)
- **name** (String)
- **updated_at** (String)


