package sshclient

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceRun() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRunCreate,
		ReadContext:   resourceRunRead,
		UpdateContext: resourceRunUpdate,
		DeleteContext: resourceRunDelete,
		Schema: map[string]*schema.Schema{
			"host_json": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"command": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Command run on creations and updates. This should be idempotent so that it can be executed any amount of times. This will also be run for reverting deletion failures.",
			},
			"expect": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The value that stdout is expected to be with trimming space characters. The output does not match, creations and updates will fail.",
			},
			"stdout": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"stdout_base64": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"stderr": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"stderr_base64": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"destroy_command": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Command run on deletions. This should be idempotent so that it can be executed any amount of times. If it fails, command for creation will be run.",
			},
			"destroy_expect": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Same as expect, but for destroy command.",
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Second),
			Update: schema.DefaultTimeout(10 * time.Second),
			Delete: schema.DefaultTimeout(10 * time.Second),
		},
	}
}

func resourceRunCommon(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
	h *Host,
	cmd, exp, keyOut, keyOut64, keyErr, keyErr64 string,
	timeout time.Duration,
) error {
	if err := h.ValidateHostInfo(); err != nil {
		return err
	}

	if err := h.ValidateAuthInfo(); err != nil {
		return err
	}

	var command string

	if c, ok := d.GetOk(cmd); ok {
		command = c.(string)
	} else {
		return nil
	}

	var stdout, stderr bytes.Buffer
	errCh := make(chan error, 2)
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		var err error
		err = h.RunCommand(command, &stdout, &stderr)
		if err != nil {
			errCh <- err
			return
		}
	}()
	go func() {
		defer wg.Done()
		time.Sleep(timeout)
		errCh <- fmt.Errorf("Timeout limit exceeded. Timeout is %s.", timeout)
	}()

	wg.Wait()
	close(errCh)

	for err := range errCh {
		return fmt.Errorf(`Error occured while running.
%s

stdout:
%s

stderr:
%s
`, err.Error(), stdout.String(), stderr.String())
	}

	if ex, ok := d.GetOk(exp); ok {
		ex := []byte(ex.(string))
		ex = bytes.TrimSpace(ex)
		ac := bytes.TrimSpace(stdout.Bytes())

		if !bytes.Equal(ex, ac) {
			return fmt.Errorf(`The output for destroy command is not the same as expectd.
	Expected: %s
	Actual  : %s`, string(ex), string(ac))
		}
	}

	if keyOut != "" {
		d.Set(keyOut, string(stdout.Bytes()))
	}
	if keyOut64 != "" {
		d.Set(keyOut64, base64.StdEncoding.EncodeToString(stdout.Bytes()))
	}
	if keyErr != "" {
		d.Set(keyErr, string(stdout.Bytes()))
	}
	if keyErr64 != "" {
		d.Set(keyErr64, base64.StdEncoding.EncodeToString(stdout.Bytes()))
	}

	return nil
}

func resourceRunCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	j := d.Get("host_json").(string)

	h, err := UnmarshalHost(j)
	if err != nil {
		return diag.FromErr(err)
	}

	err = resourceRunCommon(
		ctx,
		d,
		m,
		h,
		"command",
		"expect",
		"stdout",
		"stdout_base64",
		"stderr",
		"stderr_base64",
		d.Timeout(schema.TimeoutCreate),
	)

	if err != nil {
		return diag.Errorf("%s: %s", h, err.Error())
	}

	id := uuid.New().String()
	d.SetId(id)

	var diags diag.Diagnostics
	return diags
}

func resourceRunRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

func resourceRunUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	j := d.Get("host_json").(string)

	h, err := UnmarshalHost(j)
	if err != nil {
		return diag.FromErr(err)
	}

	err = resourceRunCommon(
		ctx,
		d,
		m,
		h,
		"command",
		"expect",
		"stdout",
		"stdout_base64",
		"stderr",
		"stderr_base64",
		d.Timeout(schema.TimeoutUpdate),
	)

	if err != nil {
		return diag.FromErr(err)
	}

	var diags diag.Diagnostics
	return diags
}

func resourceRunDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	j := d.Get("host_json").(string)

	h, err := UnmarshalHost(j)
	if err != nil {
		return diag.FromErr(err)
	}

	err = resourceRunCommon(
		ctx,
		d,
		m,
		h,
		"destroy_command",
		"destroy_expect",
		"",
		"",
		"",
		"",
		d.Timeout(schema.TimeoutDelete),
	)

	if err != nil {
		diags := diag.FromErr(err)
		revErr := resourceRunCommon(
			ctx,
			d,
			m,
			h,
			"command",
			"expect",
			"stdout",
			"stdout_base64",
			"stderr",
			"stderr_base64",
			d.Timeout(schema.TimeoutCreate),
		)
		if revErr != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "Error while revert deletion.",
				Detail:   fmt.Sprintf("Error while revert deletion. %s", revErr.Error()),
			})
		}
		return diags
	}

	var diags diag.Diagnostics
	return diags
}
