package sshclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"golang.org/x/crypto/ssh"
)

func dataSourceHost() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHostRead,
		Schema: map[string]*schema.Schema{
			"extends_host_json": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"hostname": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"password": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"port": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "If no port specified, 22 is used as default port.",
			},
			"username": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"client_private_key_pem": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Client private key in PEM format.",
				Sensitive:   true,
			},
			"host_publickey_authorized_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Host public key trusted in authorized_keys (sshd(8)) format.",
			},
			"insecure_ignore_host_key": {
				Type:        schema.TypeBool,
				Default:     false,
				Optional:    true,
				Description: "Insecurely trust the host public key. This may potentially cause Man-In-The-Middle attack.",
			},
			"json": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

const (
	tcpPortMin int = 1
	tcpPortMax int = 65535
)

type host struct {
	Hostname                   string `json:"hostname"`
	Port                       int    `json:"port"`
	Username                   string `json:"username"`
	Password                   string `json:"password"`
	ClientPrivateKeyPem        string `json:"client_private_key_pem"`
	HostPublickeyAuthorizedKey string `json:"host_publickey_authorized_key"`
	InsecureIgnoreHostKey      bool   `json:"insecure_ignore_host_key"`
}

func (h *host) String() string {
	hostPart := fmt.Sprintf("%s@%s:%d", h.Username, h.Hostname, h.Port)
	auth := h.stringAuthMethod()
	if auth == "" {
		return hostPart
	}
	return fmt.Sprintf("%s (Auth with %s)", hostPart, auth)
}

func (h *host) validateHostInfo() error {
	if h.Hostname == "" {
		return fmt.Errorf("hostname is not provided")
	}

	if h.Port < tcpPortMin || h.Port > tcpPortMax {
		return fmt.Errorf("port number out of range. %d", h.Port)
	}

	if h.Username == "" {
		return fmt.Errorf("username is not provided")
	}

	if (h.HostPublickeyAuthorizedKey == "") == !h.InsecureIgnoreHostKey {
		return fmt.Errorf("exactly one of host_publickey_authorized_key and insecure_ignore_host_key is needed")
	}

	return nil
}

func (h *host) validateAuthInfo() error {
	if (h.Password == "") == (h.ClientPrivateKeyPem == "") {
		return fmt.Errorf("exactly one of password and client_private_key_pem is needed")
	}

	return nil
}

func (h *host) stringAuthMethod() string {
	if h.Password != "" {
		return "password"
	} else if h.ClientPrivateKeyPem != "" {
		return "private key"
	}
	return ""
}

func (h *host) authMethod() ([]ssh.AuthMethod, error) {
	var auth []ssh.AuthMethod
	if h.Password != "" {
		auth = append(auth, ssh.Password(h.Password))
	} else if h.ClientPrivateKeyPem != "" {
		key, err := ssh.ParsePrivateKey([]byte(h.ClientPrivateKeyPem))
		if err != nil {
			return nil, err
		}

		auth = append(auth, ssh.PublicKeys(key))
	}

	return auth, nil
}

func (h *host) ClientConfig() (*ssh.ClientConfig, error) {
	auth, err := h.authMethod()
	if err != nil {
		return nil, err
	}

	var cb ssh.HostKeyCallback
	if h.InsecureIgnoreHostKey {
		cb = ssh.InsecureIgnoreHostKey()
	} else {
		key, _, _, _, err := ssh.ParseAuthorizedKey([]byte(h.HostPublickeyAuthorizedKey))
		cb = ssh.FixedHostKey(key)
		if err != nil {
			return nil, err
		}
	}

	return &ssh.ClientConfig{
		User:            h.Username,
		Auth:            auth,
		HostKeyCallback: cb,
	}, nil
}

func (h *host) RunCommand(command string, stdout io.Writer, stderr io.Writer) error {
	config, err := h.ClientConfig()
	if err != nil {
		return err
	}

	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", h.Hostname, h.Port), config)
	if err != nil {
		return err
	}
	defer conn.Close()

	session, err := conn.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	session.Stdout = stdout
	session.Stderr = stderr
	if err := session.Run(command); err != nil {
		return err
	}

	return nil
}

func MarshalHost(h *host) (string, error) {
	bytes, err := json.Marshal(h)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func UnmarshalHost(str string) (*host, error) {
	h := &host{}
	err := json.Unmarshal([]byte(str), h)
	if err != nil {
		return nil, err
	}

	return h, nil
}

func dataSourceHostRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	h := &host{
		Port: 22,
	}

	if j, ok := d.GetOk("extends_host_json"); ok {
		j := j.(string)
		unmarshaled, err := UnmarshalHost(j)
		if err != nil {
			return diag.FromErr(err)
		}
		h = unmarshaled
	}

	if hn, ok := d.GetOk("hostname"); ok {
		h.Hostname = hn.(string)
	} else {
		d.Set("hostname", h.Hostname)
	}

	if p, ok := d.GetOk("port"); ok {
		h.Port = p.(int)
	} else {
		d.Set("port", h.Port)
	}

	if un, ok := d.GetOk("username"); ok {
		h.Username = un.(string)
	} else {
		d.Set("username", h.Username)
	}

	if pw, ok := d.GetOk("password"); ok {
		h.Password = pw.(string)
	} else {
		d.Set("password", h.Password)
	}

	if key, ok := d.GetOk("client_private_key_pem"); ok {
		h.ClientPrivateKeyPem = key.(string)
	} else {
		d.Set("client_private_key_pem", h.ClientPrivateKeyPem)
	}

	if key, ok := d.GetOk("host_publickey_authorized_key"); ok {
		h.HostPublickeyAuthorizedKey = key.(string)
	} else {
		d.Set("host_publickey_authorized_key", h.HostPublickeyAuthorizedKey)
	}

	h.InsecureIgnoreHostKey = d.Get("insecure_ignore_host_key").(bool)

	j, err := MarshalHost(h)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("json", j); err != nil {
		return diag.FromErr(err)
	}
	id := uuid.New().String()
	d.SetId(id)

	var diags diag.Diagnostics

	return diags
}
