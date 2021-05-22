package sshclient

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccSshclientHost(t *testing.T) {
	t.Parallel()
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSshclientHostRead(t),
				Check: resource.ComposeTestCheckFunc(
					resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("data.sshclient_host.empty", "json"),
						resource.TestCheckResourceAttr("data.sshclient_host.empty", "hostname", ""),
						resource.TestCheckResourceAttr("data.sshclient_host.empty", "username", ""),
						resource.TestCheckResourceAttr("data.sshclient_host.empty", "insecure_ignore_host_key", "false"),
						resource.TestCheckResourceAttr("data.sshclient_host.empty", "port", "22"),
					),

					resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("data.sshclient_host.base", "json"),
						resource.TestCheckResourceAttr("data.sshclient_host.base", "hostname", "11.22.33.44"),
						resource.TestCheckResourceAttr("data.sshclient_host.base", "username", ""),
						resource.TestCheckResourceAttr("data.sshclient_host.base", "insecure_ignore_host_key", "false"),
						resource.TestCheckResourceAttr("data.sshclient_host.base", "port", "22"),
					),

					resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("data.sshclient_host.base_foobar", "json"),
						resource.TestCheckResourceAttr("data.sshclient_host.base_foobar", "hostname", "11.22.33.44"),
						resource.TestCheckResourceAttr("data.sshclient_host.base_foobar", "username", "foobar"),
					),

					resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("data.sshclient_host.base_foobar", "port", "2222"),

						resource.TestCheckResourceAttrSet("data.sshclient_host.base_foobar_pw", "json"),
						resource.TestCheckResourceAttr("data.sshclient_host.base_foobar_pw", "hostname", "11.22.33.44"),
						resource.TestCheckResourceAttr("data.sshclient_host.base_foobar_pw", "username", "foobar"),
						resource.TestCheckResourceAttr("data.sshclient_host.base_foobar_pw", "password", "supersecret_for_foobar"),
						resource.TestCheckResourceAttr("data.sshclient_host.base_foobar_pw", "insecure_ignore_host_key", "false"),
						resource.TestCheckResourceAttr("data.sshclient_host.base_foobar_pw", "port", "2222"),
					),
				),
			},
		},
	})
}

func testAccSshclientHostRead(t *testing.T) string {
	return `
	data "sshclient_host" "empty" {
	}
	data "sshclient_host" "base" {
		hostname = "11.22.33.44"
	}
	data "sshclient_host" "base_foobar" {
		extends_host_json = data.sshclient_host.base.json
		username = "foobar"
		insecure_ignore_host_key = true
		port = 2222
	}
	data "sshclient_host" "base_foobar_pw" {
		extends_host_json = data.sshclient_host.base_foobar.json
		password = "supersecret_for_foobar"
	}
	`
}

func testAccSshclientHostPw(t *testing.T) string {
	return fmt.Sprintf(`
		data "sshclient_host" "test_pw" {
			hostname = "%s"
			port     = %s
			username = "%s"
			password = "%s"
		}
		data "sshclient_host" "test_pw_insecure" {
			extends_host_json = data.sshclient_host.test_pw.json
			insecure_ignore_host_key = true
		}
		`,
		testGetenv(t, "TEST_PW_SSH_HOST"),
		testGetenv(t, "TEST_PW_SSH_PORT"),
		testGetenv(t, "TEST_PW_SSH_USER"),
		testGetenv(t, "TEST_PW_SSH_PASSWORD"),
	)
}

func testAccSshclientHostPubkey(t *testing.T) string {
	return fmt.Sprintf(`
		data "sshclient_host" "test_pubkey" {
			hostname = "%s"
			port     = %s
			username = "%s"
			client_private_key_pem = file("%s")
		}
		data "sshclient_host" "test_pubkey_insecure" {
			extends_host_json = data.sshclient_host.test_pubkey.json
			insecure_ignore_host_key = true
		}
		`,
		testGetenv(t, "TEST_PUBKEY_SSH_HOST"),
		testGetenv(t, "TEST_PUBKEY_SSH_PORT"),
		testGetenv(t, "TEST_PUBKEY_SSH_USER"),
		testGetenv(t, "TEST_PUBKEY_SSH_PRIKEY_PATH"),
	)
}
