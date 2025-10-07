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

package deploy

import (
	"fmt"
	"testing"
	"tnascert-deploy/config"
)

func TestClientAPIKeyLogin(t *testing.T) {
	configFile := "test_files/tnas-cert.ini"
	cfgList, err := config.LoadConfig(configFile)
	if err != nil {
		t.Errorf("loading the test config failed with error: %v", err)
	}
	cfg, ok := cfgList["deploy_default"]
	if !ok {
		t.Errorf("invalid section 'deploy_default'")
	}

	serverURL := cfg.ServerURL("api/current")
	client, err := NewClient(serverURL, cfg.TlsSkipVerify)
	if err != nil {
		t.Errorf("creating a client failed with error: %v", err)
	}

	// login with username and password
	err = clientLogin(client, cfg)
	if err != nil {
		t.Errorf("client login failed with error: %v", err)
	}
}

func TestClientBadAPIKey(t *testing.T) {
	configFile := "test_files/tnas-cert.ini"
	cfgList, err := config.LoadConfig(configFile)
	if err != nil {
		t.Errorf("loading the test config failed with error: %v", err)
	}
	cfg, ok := cfgList["bad-api-key"]
	if !ok {
		t.Errorf("invalid section 'bad-api-key'")
	}

	serverURL := cfg.ServerURL("api/current")
	client, err := NewClient(serverURL, cfg.TlsSkipVerify)
	if err != nil {
		t.Errorf("creating a client failed with error: %v", err)
	}

	// login with username and password
	err = clientLogin(client, cfg)
	if err == nil {
		t.Errorf("client login should fail with bad api key: %v", err)
	}
}

func TestClientUserNamePasswordLogin(t *testing.T) {
	configFile := "test_files/tnas-cert.ini"
	cfgList, err := config.LoadConfig(configFile)
	if err != nil {
		t.Errorf("loading the test config failed with error: %v", err)
	}
	cfg, ok := cfgList["username-password"]
	if !ok {
		t.Errorf("invalid section 'username-password'")
	}

	serverURL := cfg.ServerURL("api/current")
	client, err := NewClient(serverURL, cfg.TlsSkipVerify)
	if err != nil {
		t.Errorf("creating a client failed with error: %v", err)
	}

	// login with username and password
	err = clientLogin(client, cfg)
	if err != nil {
		t.Errorf("client login failed with error: %v", err)
	}
}

func TestClientBadUserName(t *testing.T) {
	configFile := "test_files/tnas-cert.ini"
	cfgList, err := config.LoadConfig(configFile)
	if err != nil {
		t.Errorf("loading the test config failed with error: %v", err)
	}
	cfg, ok := cfgList["bad-username"]
	if !ok {
		t.Errorf("invalid section 'bad-username'")
	}

	serverURL := cfg.ServerURL("api/current")
	client, err := NewClient(serverURL, cfg.TlsSkipVerify)
	if err != nil {
		t.Errorf("creating a client failed with error: %v", err)
	}

	// login with username and password
	err = clientLogin(client, cfg)
	if err == nil {
		t.Errorf("client login should faile with bad username error: %v", err)
	}
}

func TestClientBadPassword(t *testing.T) {
	configFile := "test_files/tnas-cert.ini"
	cfgList, err := config.LoadConfig(configFile)
	if err != nil {
		t.Errorf("loading the test config failed with error: %v", err)
	}
	cfg, ok := cfgList["bad-password"]
	if !ok {
		t.Errorf("invalid section 'bad-password'")
	}

	serverURL := cfg.ServerURL("api/current")
	client, err := NewClient(serverURL, cfg.TlsSkipVerify)
	if err != nil {
		t.Errorf("creating a client failed with error: %v", err)
	}

	// login with username and password
	err = clientLogin(client, cfg)
	if err == nil {
		t.Errorf("client login should fail with bad password error: %v", err)
	}
}

func TestVerifyCertificate(t *testing.T) {
	configFile := "test_files/tnas-cert.ini"

	// test vefifying a self signed cert
	cfgList, err := config.LoadConfig(configFile)
	if err != nil {
		t.Errorf("error loading the test config: %v", err)
	}
	cfg, ok := cfgList["deploy_default"]
	if !ok {
		t.Errorf("invalid section 'deploy_default'")
	}
	err = verifyCertificateKeyPair(cfg.FullChainPath, cfg.Private_key_path)
	if err != nil {
		t.Errorf("error verifing a self signed certificate: %v", err)
	}

	// test verifying an expired self signed cert.
	cfgList, err = config.LoadConfig(configFile)
	if err != nil {
		t.Errorf("error loading the test config: %v", err)
	}
	cfg, ok = cfgList["expired-cert"]
	if !ok {
		t.Errorf("invalid section 'expired-cert'")
	}
	err = verifyCertificateKeyPair(cfg.FullChainPath, cfg.Private_key_path)
	if err == nil {
		t.Errorf("failed to detect an expired certificate:")
	}
}

func TestDeployPkg(t *testing.T) {
	configFile := "test_files/tnas-cert.ini"

	cfgList, err := config.LoadConfig(configFile)
	if err != nil {
		t.Errorf("error loading the test config: %v", err)
	}
	cfg, ok := cfgList["deploy_default"]
	if !ok {
		t.Errorf("invalid section 'deploy_default'")
	}
	certName := cfg.CertName()
	fmt.Printf("certName: %s\n", certName)

	serverURL := cfg.ServerURL("api/current")
	client, err := NewClient(serverURL, cfg.TlsSkipVerify)
	if err != nil {
		t.Errorf("error creating a client: %v", err)
	}
	client.SetConfig(cfg)

	err = createCertificate(client, cfg)
	if err != nil {
		t.Errorf("create certificate failed with error: %v", err)
	}

	err = loadCertificateList(client, cfg)
	if err != nil {
		t.Errorf("load the certificate list failed with error: %v", err)
	}

	err = addAsFTPCertificate(client, cfg)
	if err != nil {
		t.Errorf("adding an FTP certificate failed with error: %v", err)
	}

	result, err := addAsUICertificate(client, cfg)
	if err != nil && result != true {
		t.Errorf("adding a UI certificate failed with error: %v", err)
	}

	err = addAsAppCertificate(client, cfg)
	if err != nil {
		t.Errorf("add an APP certificate failed with error: %v", err)
	}

	err = deleteCertificates(client, cfg)
	if err != nil {
		t.Errorf("delete certificates failed with error: %v", err)
	}

	err = InstallCertificate(client, cfg)
	if err != nil {
		t.Errorf("install certificate failed with error: %v", err)
	}
}
