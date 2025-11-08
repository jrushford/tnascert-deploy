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
	"strings"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	configFile := "test_files/tnas-loadconfig.ini"

	cfg_list, err := LoadConfig(configFile)
	if err != nil {
		t.Errorf("loading the test config failed with error: %v", err)
	}
	if len(cfg_list) != 3 {
		t.Errorf("the config list size should be 3")
	}
	if cfg_list["deploy_default"].ConnectHost != "nas01.mydomain.com" {
		t.Errorf("connect_host should be nas01.mydomain.com")
	}
	if cfg_list["nas02"].ConnectHost != "nas02.mydomain.com" {
		t.Errorf("connect_host should be nas02.mydomain.com")
	}
	if cfg_list["nas03"].ConnectHost != "nas03.mydomain.com" {
		t.Errorf("connect_host should be nas03.mydomain.com")
	}
}

func TestReadConfigs(t *testing.T) {
	configFile := "test_files/tnas-cert.ini"

	// test loading the default config section
	cfgList, err := LoadConfig(configFile)
	if err != nil {
		t.Errorf("loading the test config failed with error: %v", err)
	}
	cfg, ok := cfgList["deploy_default"]
	if !ok {
		t.Fatalf("invalid section 'deploy_default'")
	}
	if cfg.ConnectHost != "nas01.mydomain.com" {
		t.Errorf("connect_host should be nas01.mydomain.com")
	}
	if cfg.PrivateKeyPath != "test_files/privkey.pem" {
		t.Errorf("private_key_path should be test_files/privkey.pem")
	}
	if cfg.FullChainPath != "test_files/fullchain.pem" {
		t.Errorf("fullchain_path should be test_files/fullchain.pem")
	}
	if cfg.Protocol != "wss" {
		t.Errorf("protocol should be wss")
	}
	if cfg.TlsSkipVerify != false {
		t.Errorf("tls_skip_verify should be false")
	}
	if cfg.DeleteOldCerts != true {
		t.Errorf("delete_old_certs should be true")
	}
	if cfg.AddAsUiCertificate != true {
		t.Errorf("add_as_ui_certificate should be true")
	}

	// test opening  non-existent file
	cfgList, err = LoadConfig("non_existent_file")
	if err == nil {
		t.Errorf("exepected an error opening a non-existent file: %v", err)
	}

	// test nas02 config section
	if cfgList, err = LoadConfig(configFile); err != nil {
		t.Errorf("loading the test config failed with error: %v", err)
	}
	cfg, ok = cfgList["nas02"]
	if cfg == nil && !ok {
		t.Errorf("section 'nas02' does not exist")
	}
	if cfg.ConnectHost != "nas02.mydomain.com" {
		t.Errorf("connect_host should be nas02.mydomain.com")
	}
	serverURL := cfg.ServerURL() + "/api/current"
	if serverURL != "wss://nas02.mydomain.com:443/api/current" {
		t.Errorf("ServerURL should be wss://nas02.mydomain.com:443/api/current")
	}

	certName := cfg.CertName()
	if strings.HasPrefix(certName, "letsencrypt-") == false {
		t.Errorf("certname prefix should be letsencrypt-")
	}

	// test checkConfig function
	if err = cfg.checkConfig(); err != nil {
		t.Errorf("checkConfig() failed with error: %v", err)
	}

	// test nas03 config section
	if cfgList, err = LoadConfig(configFile); err != nil {
		t.Errorf("loading the test config failed with error: %v", err)
	}
	cfg, ok = cfgList["nas03"]
	if !ok {
		t.Errorf("invalid section 'nas03'")
	}
	if cfg.ConnectHost != "nas03.mydomain.com" {
		t.Errorf("connect_host should be nas02.mydomain.com")
	}

	// load a config file with no cert_base_name defined
	cfg, ok = cfgList["no_cert_basename"]
	if !ok {
		t.Errorf("invalid section 'no_cert_basename'")
	}
	if cfg != nil && cfg.CertBasename != Default_base_cert_name {
		t.Errorf("cert_basename should be %s", Default_base_cert_name)
	}

	// load a config file with no protocol defined
	cfg, ok = cfgList["no_protocol"]
	if !ok {
		t.Errorf("invalid section 'no_protocol'")
	}
	if cfg != nil && cfg.Protocol != Default_protocol {
		t.Errorf("protocol should be the %s", Default_protocol)
	}

	// load a config file with no timeout seconds defined
	cfg, ok = cfgList["no_timeout_seconds"]
	if !ok {
		t.Errorf("invalid section 'no_timeout_seconds'")
	}
	if cfg != nil && cfg.TimeoutSeconds != Default_timeout_seconds {
		t.Errorf("timeout_seconds should be %d", Default_timeout_seconds)
	}
}
