package sshclient

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"regexp"
	"time"

	"github.com/bramvdbogaerde/go-scp"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	permPatStr = `^[0-7][0-7][0-7]$`
)

var (
	permPat = regexp.MustCompile(permPatStr)
)

func resourceScpPut() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScpPutCreate,
		ReadContext:   resourceScpPutRead,
		UpdateContext: resourceScpPutUpdate,
		DeleteContext: resourceScpPutDelete,
		Schema: map[string]*schema.Schema{
			"host_json": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"data": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"data_base64": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"remote_path": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Remote path to place the file. Take care to avoid including spaces or other special characters. These may be troublesome when being interpreted by remote shell.",
			},
			"permissions": {
				Type:     schema.TypeString,
				Required: true,
				Description: fmt.Sprintf(
					"Permission information in %s form that each block represents user, group and others access in order, and each bits in blocks represents read, write and execute permissions. This is compatible with the stat(1) command `stat -c %%a`. For example, you can use 777 to grant all full access, or use can use 644 for restricted access.",
					permPatStr,
				),
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Second),
			Update: schema.DefaultTimeout(10 * time.Second),
		},
	}
}

func resourceScpPutCreateUpdate(ctx context.Context, d *schema.ResourceData, m interface{}, timeout time.Duration) diag.Diagnostics {
	data, okData := d.GetOk("data")
	data64, okData64 := d.GetOk("data_base64")

	if okData == okData64 {
		return diag.Errorf("Exactly one of data and data_base64 should be specified.")
	}

	var b []byte
	var err error
	if okData {
		b = []byte(data.(string))
	} else {
		b, err = base64.StdEncoding.DecodeString(data64.(string))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	j := d.Get("host_json").(string)

	h, err := UnmarshalHost(j)
	if err != nil {
		return diag.FromErr(err)
	}

	err = func() error {
		if err := h.ValidateHostInfo(); err != nil {
			return err
		}

		if err := h.ValidateAuthInfo(); err != nil {
			return err
		}

		client, err := h.ClientConfig()
		if err != nil {
			return err
		}

		c := scp.NewClientWithTimeout(fmt.Sprintf("%s:%d", h.Hostname, h.Port), client, timeout)
		if err = c.Connect(); err != nil {
			return nil
		}
		defer c.Close()

		remotePath := d.Get("remote_path").(string)
		perm, err := parsePermStr(d.Get("permissions").(string))
		if err != nil {
			return err
		}

		err = c.CopyFile(bytes.NewReader(b), remotePath, perm)
		if err != nil {
			return err
		}

		return nil
	}()

	if err != nil {
		return diag.Errorf("%s: %s", h.String(), err.Error())
	}

	var diags diag.Diagnostics
	return diags
}

func resourceScpPutCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	diags := resourceScpPutCreateUpdate(ctx, d, m, d.Timeout(schema.TimeoutCreate))
	if diags.HasError() {
		return diags
	}

	id := uuid.New().String()
	d.SetId(id)

	return diags
}

func resourceScpPutRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	h, err := UnmarshalHost(d.Get("host_json").(string))

	if err != nil {
		return diag.FromErr(err)
	}

	if err := h.ValidateHostInfo(); err != nil {
		return diag.FromErr(err)
	}

	if err := h.ValidateAuthInfo(); err != nil {
		return diag.FromErr(err)
	}

	var diags diag.Diagnostics
	return diags
}

func resourceScpPutUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceScpPutCreateUpdate(ctx, d, m, d.Timeout(schema.TimeoutUpdate))
}

func resourceScpPutDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceScpPutRead(ctx, d, m)
}

func parsePermStr(perms string) (string, error) {
	match := permPat.Match([]byte(perms))
	if !match {
		return "", fmt.Errorf("Permissions string must be in form of %s .", permPatStr)
	}

	return fmt.Sprintf("0%s", perms), nil
}
