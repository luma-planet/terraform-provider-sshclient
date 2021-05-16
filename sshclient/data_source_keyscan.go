package sshclient

import (
	"context"
	"fmt"
	"net"

	"golang.org/x/crypto/ssh"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceKeyscan() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceKeyscanRead,

		Schema: map[string]*schema.Schema{
			"host_json": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"authorized_key": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func keyscanCallback(ch chan ssh.PublicKey) ssh.HostKeyCallback {
	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		ch <- key
		return nil
	}
}

func dataSourceKeyscanRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	hostJson := d.Get("host_json").(string)
	h, err := UnmarshalHost(hostJson)
	if err != nil {
		return diag.FromErr(err)
	}

	err = h.ValidateHostInfo()
	if err != nil {
		return diag.FromErr(err)
	}

	if !h.InsecureIgnoreHostKey {
		return diag.Errorf("To scan host key, insecure_ignore_host_key should be explicitly set.")
	}

	ch := make(chan ssh.PublicKey, 1)
	config := &ssh.ClientConfig{
		User:            h.Username,
		Auth:            []ssh.AuthMethod{},
		HostKeyCallback: keyscanCallback(ch),
	}

	ssh.Dial("tcp", fmt.Sprintf("%s:%d", h.Hostname, h.Port), config)

	pub := <-ch
	d.Set("authorized_key", string(ssh.MarshalAuthorizedKey(pub)))

	id := uuid.New().String()
	d.SetId(id)

	return diags
}
