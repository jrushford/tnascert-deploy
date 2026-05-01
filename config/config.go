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
	"os"
	"strconv"
	"time"
)

const (
	Config_file             = "tnas-cert.ini"
	Default_base_cert_name  = "tnas-cert-deploy"
	Default_section         = "deploy_default"
	Default_port            = 443
	Default_protocol        = "wss"
	Default_timeout_seconds = 10
)

type Config struct {
	ApiKey                 string `ini:"api_key"`                // TrueNAS 64 byte API Key
	CertBasename           string `ini:"cert_basename"`          // basename for cert naming in TrueNAS
	ClientApi              string `ini:"client_api"`             // Client type, 'wsapi' (default) or restapi
	ConnectHost            string `ini:"connect_host"`           // TrueNAS hostname
	DeleteOldCertsStr      string `ini:"delete_old_certs"`       // whether to remove old certificates, String value
	FullChainPath          string `ini:"full_chain_path"`        // path to full_chain.pem
	PortStr                string `ini:"port"`                   // TrueNAS API endpoint port, String value
	Protocol               string `ini:"protocol"`               // websocket protocol 'ws' or 'wss' 'wss' is default
	PrivateKeyPath         string `ini:"private_key_path"`       // path to private_key.pem
	TlsSkipVerifyStr       string `ini:"tls_skip_verify"`        // strict SSL cert verification of the endpoint, String value
	AddAsUiCertificateStr  string `ini:"add_as_ui_certificate"`  // Install as the active UI certificate if true, String value
	AddAsFTPCertificateStr string `ini:"add_as_ftp_certificate"` // Install as the active FTP service certificate if true, String value
	AddAsAppCertificateStr string `ini:"add_as_app_certificate"` // Install as the active APP service certificate if true, String value
	AppList                string `ini:"app_list"`               // comma separated list of Apps to deploy the certificate too.
	TimeoutSecondsStr      string `ini:"timeoutSeconds"`         // the number of seconds after which the truenas client calls fail, String value
	DebugStr               string `ini:"debug"`                  // debug logging if true, String value
	Username               string `ini:"username"`               // an admin user name
	Password               string `ini:"password"`               // admin users password
	DeleteOldCerts         bool   // whether to remove old certificates
	Port                   uint64 // TrueNAS API endpoint port
	TlsSkipVerify          bool   // strict SSL cert verification of the endpoint
	AddAsUiCertificate     bool   // Install as the active UI certificate if true.
	AddAsFTPCertificate    bool   // Install as the active FTP certificate if true.
	AddAsAppCertificate    bool   // Install as the active APP certificate if true.
	TimeoutSeconds         int64  // the number of seconds after which the truenas client calls fail
	Debug                  bool   // debug logging if true.
	certName               string // instance generated certificate name.
	serverURL              string // instance generated server URL
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
		c.serverURL = fmt.Sprintf("%s://%s:%d", c.Protocol, c.ConnectHost, c.Port)
	}
	return c.serverURL
}

func (c *Config) checkConfig() error {
	// lookup the ApiKey
	c.ApiKey = os.ExpandEnv(c.ApiKey)

	// lookup the Username
	c.Username = os.ExpandEnv(c.Username)

	// lookup the Password
	c.Password = os.ExpandEnv(c.Password)
	if c.ApiKey == "" && c.Username == "" {
		return fmt.Errorf("no authentication is defined, use an 'api_key' or the 'username' and 'password'")
	}

	// lookup the cert_basename
	c.CertBasename = os.ExpandEnv(c.CertBasename)
	if c.CertBasename == "" {
		c.CertBasename = Default_base_cert_name
	}

	// lookup the connect_host
	c.ConnectHost = os.ExpandEnv(c.ConnectHost)
	if c.ConnectHost == "" {
		return fmt.Errorf("the required 'connect_host' parameter is not defined")
	}

	// lookup the client_api
	c.ClientApi = os.ExpandEnv(c.ClientApi)
	if c.ClientApi == "" {
		c.ClientApi = Default_protocol
	} else if c.ClientApi != "restapi" && c.ClientApi != "wsapi" {
		return fmt.Errorf("invalid client_api '%s' use 'restapi' or 'wsapi'", c.ClientApi)
	}

	// lookup delete_old_certs
	c.DeleteOldCertsStr = os.ExpandEnv(c.DeleteOldCertsStr)
	if b, err := strconv.ParseBool(c.DeleteOldCertsStr); err == nil {
		c.DeleteOldCerts = b
	} else {
		return err
	}

	//lookup the full_chain_path
	c.FullChainPath = os.ExpandEnv(c.FullChainPath)
	if c.FullChainPath == "" {
		return fmt.Errorf("the required 'fullchain_path' is not defined")
	}

	// lookup the private_key_path
	c.PrivateKeyPath = os.ExpandEnv(c.PrivateKeyPath)
	if c.PrivateKeyPath == "" {
		return fmt.Errorf("the required 'private_key_path' is not defined")
	}

	// lookup the port
	c.PortStr = os.ExpandEnv(c.PortStr)
	if c.PortStr != "" {
		if i, err := strconv.ParseUint(c.PortStr, 10, 64); err == nil {
			c.Port = i
		} else {
			return err
		}
	}
	// if port is not defined, use the default
	if c.Port == 0 {
		c.Port = Default_port
	}

	// lookup the protocol
	c.Protocol = os.ExpandEnv(c.Protocol)
	// if the protocol is not defined, use the default
	if c.Protocol == "" {
		c.Protocol = Default_protocol
	} else if c.Protocol != "ws" && c.Protocol != "wss" && c.Protocol != "http" && c.Protocol != "https" {
		return fmt.Errorf("invalid protocol use 'ws' or 'wss' or 'http' or 'https")
	}

	// lookup tls_skip_verify
	c.TlsSkipVerifyStr = os.ExpandEnv(c.TlsSkipVerifyStr)
	if b, err := strconv.ParseBool(c.TlsSkipVerifyStr); err == nil {
		c.TlsSkipVerify = b
	} else {
		return err
	}

	// lookup the add_as_ui_certificate
	c.AddAsUiCertificateStr = os.ExpandEnv(c.AddAsUiCertificateStr)
	if b, err := strconv.ParseBool(c.AddAsUiCertificateStr); err == nil {
		c.AddAsUiCertificate = b
	} else {
		return err
	}

	// lookup the add_as_ftp_certificate
	c.AddAsFTPCertificateStr = os.ExpandEnv(c.AddAsFTPCertificateStr)
	if b, err := strconv.ParseBool(c.AddAsFTPCertificateStr); err == nil {
		c.AddAsFTPCertificate = b
	} else {
		return err
	}

	// lookup the add_as_app_certificate
	c.AddAsAppCertificateStr = os.ExpandEnv(c.AddAsAppCertificateStr)
	if b, err := strconv.ParseBool(c.AddAsAppCertificateStr); err == nil {
		c.AddAsAppCertificate = b
	} else {
		return err
	}

	// lookup the app_list
	c.AppList = os.ExpandEnv(c.AppList)

	// lookup the timeoutSeconds
	c.TimeoutSecondsStr = os.ExpandEnv(c.TimeoutSecondsStr)
	if c.TimeoutSecondsStr != "" {
		if i, err := strconv.ParseInt(c.TimeoutSecondsStr, 10, 64); err == nil {
			c.TimeoutSeconds = i
		} else {
			return err
		}
	}
	if c.TimeoutSeconds <= 0 {
		c.TimeoutSeconds = Default_timeout_seconds
	}

	// lookup the debug key
	c.DebugStr = os.ExpandEnv(c.DebugStr)
	if b, err := strconv.ParseBool(c.DebugStr); err == nil {
		c.Debug = b
	} else {
		return err
	}

	return nil
}
