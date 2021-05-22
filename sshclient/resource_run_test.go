package sshclient

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccSshclientRun(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSshclientRunRead(t),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sshclient_run.pw__echo_hi", "stdout", "hi"),
					resource.TestCheckResourceAttr("sshclient_run.pw__echo_hi", "stderr", "err"),

					resource.TestCheckResourceAttr("sshclient_run.pubkey__echo_hi", "stdout", "hi"),
					resource.TestCheckResourceAttr("sshclient_run.pubkey__echo_hi", "stderr", ""),
				),
			},
		},
	})
}

func testAccSshclientRunRead(t *testing.T) string {
	return fmt.Sprintf(`
		locals {
			path        = "/tmp/test-terraform-provider-sshclient/basic.txt"
		}
		%s
		%s
		resource "sshclient_run" "pw__echo_hi" {
			host_json = data.sshclient_host.test_pw_insecure.json
			command   = "echo -n hi; echo -n err >/dev/fd/2"
			expect    = "hi"
		}
		resource "sshclient_run" "pubkey__echo_hi" {
			host_json      = data.sshclient_host.test_pubkey_insecure.json
			command_base64 = "ZWNobyAtbiBoaQo="
			expect         = "hi"
		}
		`,
		testAccSshclientHostPw(t),
		testAccSshclientHostPubkey(t),
	)
}
