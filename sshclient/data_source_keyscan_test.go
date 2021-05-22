package sshclient

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccSshclientKeyscan(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSshclientKeyscanRead(t),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.sshclient_keyscan.pubkey_scan", "authorized_key"),
				),
			},
		},
	})
}

func testAccSshclientKeyscanRead(t *testing.T) string {
	return fmt.Sprintf(`
		%s
		data "sshclient_keyscan" "pubkey_scan" {
			host_json = data.sshclient_host.test_pubkey_insecure.json
		}
		`,
		testAccSshclientHostPubkey(t),
	)
}
