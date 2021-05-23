terraform {
  required_providers {
    sshclient = {
      version = "1.0"
      source  = "github.com/luma-planet/sshclient"
    }
  }
}

provider "sshclient" {
}

data "sshclient_host" "myhost_keyscan" {
  hostname                 = var.hostname
  port                     = var.port
  username                 = "keyscan"
  insecure_ignore_host_key = true
}

data "sshclient_keyscan" "myhost" {
  host_json = data.sshclient_host.myhost_keyscan.json
}

data "sshclient_host" "myhost_main" {
  extends_host_json = data.sshclient_host.myhost_keyscan.json
  username          = var.username
  password          = var.password
  client_private_key_pem = (
    var.client_private_key_pem_path != null
    ? file(var.client_private_key_pem_path)
    : null
  )
  host_publickey_authorized_key = data.sshclient_keyscan.myhost.authorized_key
}

resource "sshclient_run" "myhost_whoami" {
  host_json       = data.sshclient_host.myhost_main.json
  command         = "sleep 0; whoami > whoami.txt && cat whoami.txt; echo err! >&2"
  expect          = var.username
  destroy_command = "rm -f whoami.txt && echo destroy_ok"
  destroy_expect  = "destroy_ok"
  timeouts {
    create = "5s"
    update = "5s"
    delete = "5s"
  }
}

resource "sshclient_scp_put" "myhost__some_dat" {
  host_json   = data.sshclient_host.myhost_main.json
  data_base64 = filebase64("./some.dat")
  remote_path = "some.dat"
  permissions = "644"
}

resource "sshclient_scp_put" "myhost__checksum_sh" {
  host_json   = data.sshclient_host.myhost_main.json
  permissions = "774"
  data        = file("./checksum.sh")
  remote_path = "checksum.sh"
}

resource "sshclient_run" "myhost_checksum" {
  host_json = data.sshclient_host.myhost_main.json
  command   = "./checksum.sh ${filesha256("./checksum.sh")}"
  expect    = "784955"
  depends_on = [
    sshclient_scp_put.myhost__checksum_sh,
    sshclient_scp_put.myhost__some_dat,
  ]
}
