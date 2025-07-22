/*
 * Copyright (C) 2025 by John J. Rushford jrushford@apache.org
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package config

import (
	"fmt"
	"github.com/ncruces/go-strftime"
	"gopkg.in/ini.v1"
	"time"
)

const (
	WS                      = "ws"
	WSS                     = "wss"
	Config_file             = "tnas-cert.ini"
	Default_base_cert_name  = "tnas-cert-deploy"
	Default_section         = "deploy_default"
	Default_port            = 443
	Default_protocol        = WSS
	Default_timeout_seconds = 10
	endpoint                = "api/current"
)

type Config struct {
	Api_key             string `ini:"api_key"`                // TrueNAS 64 byte API Key
	CertBasename        string `ini:"cert_basename"`          // basename for cert naming in TrueNAS
	ConnectHost         string `ini:"connect_host"`           // TrueNAS hostname
	DeleteOldCerts      bool   `ini:"delete_old_certs"`       // whether to remove old certificates
	FullChainPath       string `ini:"full_chain_path"`        // path to full_chain.pem
	Port                uint64 `ini:"port"`                   // TrueNAS API endpoint port
	Protocol            string `ini:"protocol"`               // websocket protocol 'ws' or 'wss' 'wss' is default
	Private_key_path    string `ini:"private_key_path"`       // path to private_key.pem
	TlsSkipVerify       bool   `ini:"tls_skip_verify"`        // strict SSL cert verification of the endpoint
	AddAsUiCertificate  bool   `ini:"add_as_ui_certificate"`  // Install as the active UI certificate if true
	AddAsFTPCertificate bool   `ini:"add_as_ftp_certificate"` // Install as the active FTP service certificate if true
	AddAsAppCertificate bool   `ini:"add_as_app_certificate"` // Install as the active APP service certificate if true
	TimeoutSeconds      int64  `ini:"timeoutSeconds"`         // the number of seconds after which the truenas client calls fail
	Debug               bool   `ini:"debug"`                  // debug logging if true
	Username            string `ini:"username"`               // an admin user name
	Password            string `ini:"password"`               // admin users password
	certName            string // instance generated certificate name
	serverURL           string // instance generated server URL
}

func LoadConfig(config_file string) (map[string]*Config, error) {
	var cfg_list = make(map[string]*Config)

	// load the config file
	f, err := ini.Load(config_file)
	if err != nil {
		return nil, err
	}

	for _, section := range f.Sections() {
		name := section.Name()
		if name == "DEFAULT" {
			continue
		}
		var c = Config{}
		err = f.Section(name).MapTo(&c)
		if err != nil {
			return nil, err
		}
		err = c.checkConfig()
		if err != nil {
			return nil, fmt.Errorf("error in section '%s': %v", name, err)
		}
		cfg_list[name] = &c
	}

	return cfg_list, nil
}

func (c *Config) CertName() string {
	if c.certName == "" {
		c.certName = c.CertBasename + strftime.Format("-%Y-%m-%d-%s", time.Now())
	}
	return c.certName
}

func (c *Config) ServerURL() string {
	if c.serverURL == "" {
		c.serverURL = fmt.Sprintf("%s://%s:%d/%s", c.Protocol, c.ConnectHost, c.Port, endpoint)
	}
	return c.serverURL
}

func (c *Config) checkConfig() error {
	// if not the cert_basename is not defined use the default
	if c.CertBasename == "" {
		c.CertBasename = Default_base_cert_name
	}
	if c.ConnectHost == "" {
		return fmt.Errorf("connect_host is not defined")
	}
	if c.FullChainPath == "" {
		return fmt.Errorf("fullchain_path is not defined")
	}
	// if port is not defined, use the default
	if c.Port == 0 {
		c.Port = Default_port
	}
	// if the protocol is not defined, use the default
	if c.Protocol == "" {
		c.Protocol = Default_protocol
	} else {
		if c.Protocol != WS && c.Protocol != WSS {
			return fmt.Errorf("invalid protocol")
		}
	}
	if c.Private_key_path == "" {
		return fmt.Errorf("private_key_path is not defined")
	}
	if c.TimeoutSeconds <= 0 {
		c.TimeoutSeconds = Default_timeout_seconds
	}

	return nil
}
