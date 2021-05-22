package sshclient

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/joho/godotenv"
)

var testAccProviders map[string]*schema.Provider
var testAccProvider *schema.Provider
var testAccProviderFactories map[string]func() (*schema.Provider, error)

func init() {
	err := godotenv.Load("../local.env")
	if err != nil {
		fmt.Println(err)
	}
	err = godotenv.Load("../test.local.env")
	if err != nil {
		fmt.Println(err)
	}
	err = godotenv.Load("../test.env")
	if err != nil {
		fmt.Println(err)
	}

	testAccProvider = Provider()
	config := terraform.NewResourceConfigRaw(map[string]interface{}{})
	testAccProvider.Configure(context.Background(), config)
	testAccProviders = map[string]*schema.Provider{
		"sshclient": testAccProvider,
	}

	testAccProviderFactories = map[string]func() (*schema.Provider, error){
		"sshclient": func() (*schema.Provider, error) {
			return testAccProvider, nil
		},
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func getTestEnvs() map[string]struct{} {
	return map[string]struct{}{
		"TEST_PW_SSH_HOST":     {},
		"TEST_PW_SSH_PORT":     {},
		"TEST_PW_SSH_USER":     {},
		"TEST_PW_SSH_PASSWORD": {},

		"TEST_PUBKEY_SSH_HOST":        {},
		"TEST_PUBKEY_SSH_PORT":        {},
		"TEST_PUBKEY_SSH_USER":        {},
		"TEST_PUBKEY_SSH_PRIKEY_PATH": {},
	}
}

func testGetenv(t *testing.T, key string) string {
	_, ok := getTestEnvs()[key]
	if !ok {
		t.Fatalf("%s is not registered as testing env", key)
	}
	return os.Getenv(key)
}

func testAccPreCheck(t *testing.T) {
	for n := range getTestEnvs() {
		if v := os.Getenv(n); v == "" {
			t.Fatalf("%s must be set for acceptance tests", n)
		}
	}
}
