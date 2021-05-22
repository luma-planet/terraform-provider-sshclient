package sshclient

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccSshclientScpPut(t *testing.T) {
	t.Parallel()
	rText := acctest.RandString(30)
	rText64 := base64.StdEncoding.EncodeToString([]byte(rText))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSshclientScpPutBasic(t, "basic", rText, "", ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sshclient_run.cat__basic_txt", "stdout", rText),
				),
			},
			{
				Config: testAccSshclientScpPutBasic(t, "base64-770", "", rText64, "770"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sshclient_run.cat__basic_txt", "stdout", rText),
				),
			},
		},
	})
}

func testAccSshclientScpPutBasic(t *testing.T, prefix, dat, dat64, perm string) string {
	return fmt.Sprintf(`
		locals {
			path        = "/tmp/test-terraform-provider-sshclient/%s-basic.txt"
			data        = "%s"
			data_base64 = "%s"
			permissions = "%s"
		}
		%s
		resource "sshclient_scp_put" "put__basic_txt" {
			host_json   = data.sshclient_host.test_pubkey_insecure.json
			remote_path = local.path
			data        = local.data
			data_base64 = local.data_base64
			permissions = local.permissions
			depends_on  = [sshclient_run.caretake__basic_txt]
		}
		resource "sshclient_run" "stat__basic_txt" {
			host_json   = data.sshclient_host.test_pubkey_insecure.json
			command     = "stat -c %%a ${local.path}"
			expect      = local.permissions
			depends_on  = [sshclient_scp_put.put__basic_txt]
		}
		resource "sshclient_run" "cat__basic_txt" {
			host_json   = data.sshclient_host.test_pubkey_insecure.json
			command     = "cat ${local.path}"
			depends_on  = [sshclient_scp_put.put__basic_txt]
		}
		resource "sshclient_run" "caretake__basic_txt" {
			host_json       = data.sshclient_host.test_pubkey_insecure.json
			command         = "mkdir -p $(dirname ${local.path}) || true; test -d $(dirname ${local.path}); rm ${local.path} || true"
		}
		`,
		prefix,
		dat,
		dat64,
		perm,
		testAccSshclientHostPubkey(t),
	)
}

func TestParsePermStr(t *testing.T) {
	cases := []struct {
		input    string
		accepted bool
		expected string
	}{
		{
			input:    "000",
			accepted: true,
			expected: "0000",
		},
		{
			input:    "644",
			accepted: true,
			expected: "0644",
		},
		{
			input:    "777",
			accepted: true,
			expected: "0777",
		},
		{
			input: "987",
		},
		{
			input: "0123",
		},
	}
	for _, c := range cases {
		r, err := parsePermStr(c.input)
		if (err == nil) != c.accepted {
			t.Errorf(`Error status not match:
	Case:                 %#v
	Succeeded?:           %v
	Expected to succeed?: %v`, r, err == nil, c.accepted)
			continue
		}
		if err == nil && r != c.expected {
			t.Errorf(`Error output not match:
	Case:     %#v
	Actual:   %v
	Expected: %v`, r, r, c.expected)
			continue
		}
	}
}
