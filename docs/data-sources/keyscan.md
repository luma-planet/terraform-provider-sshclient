---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "sshclient_keyscan Data Source - terraform-provider-sshclient"
subcategory: ""
description: |-
  
---

# sshclient_keyscan (Data Source)

```
data "sshclient_keyscan" "myhost" {
  host_json = data.sshclient_host.myhost_keyscan.json
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- **host_json** (String, Sensitive)

### Optional

- **id** (String) The ID of this resource.

### Read-Only

- **authorized_key** (String)

