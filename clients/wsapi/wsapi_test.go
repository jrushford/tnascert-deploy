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

package wsapi

import (
	"fmt"
	"testing"
	"tnascert-deploy/config"
)

func getConfig() (*config.Config, error) {
	configFile := "test_files/tnas-cert.ini"
	cfgList, err := config.LoadConfig(configFile)
	if err != nil {
		return nil, fmt.Errorf("error loading config file '%s'", configFile)
	}
	cfg, ok := cfgList["deploy_default"]
	if !ok {
		return nil, fmt.Errorf("invalid section 'deploy_default'")
	}
	return cfg, nil
}

func TestAddAsAppCertificate(t *testing.T) {
	cfg, err := getConfig()
	if err != nil {
		t.Errorf("error loading config: %v", err)
	}

	client, err := NewMockWebSocketClient(cfg)
	if err != nil {

		t.Fatalf("error creating the mock websocket client: %v", err)
	}

	client.Version = "TrueNAS-SCALE-25.0.0.0"
	err = addAsAppCertificate(client, "grafana")
	if err != nil {
		t.Errorf("error adding app certificate: %v", err)
	}
}

func TestAddAsFTPCertificate(t *testing.T) {
	cfg, err := getConfig()
	if err != nil {
		t.Errorf("error loading config: %v", err)
	}

	client, err := NewMockWebSocketClient(cfg)
	if err != nil {

		t.Fatalf("error creating the mock websocket client: %v", err)
	}
	certsList[cfg.CertName()] = 102
	err = addAsFTPCertificate(*client)
	if err != nil {
		t.Errorf("error adding app certificate: %v", err)
	}
}

func TestAddAsUICertificate(t *testing.T) {
	cfg, err := getConfig()
	if err != nil {
		t.Errorf("error loading config: %v", err)
	}

	client, err := NewMockWebSocketClient(cfg)
	if err != nil {

		t.Fatalf("error creating the mock websocket client: %v", err)
	}
	certsList[cfg.CertName()] = 102
	err = addAsUICertificate(*client)
	if err != nil {
		t.Errorf("error adding app certificate: %v", err)
	}
}

func TestDeleteCertificates(t *testing.T) {
	cfg, err := getConfig()
	if err != nil {
		t.Errorf("error loading config: %v", err)
	}

	client, err := NewMockWebSocketClient(cfg)
	if err != nil {

		t.Fatalf("error creating the mock websocket client: %v", err)
	}
	certsList[cfg.CertName()] = 102
	certsList["tnas-cert-deploy-2024-01-01-08080808"] = 101
	certsList["tnas-cert-deploy-2024-02-01-09090909"] = 100
	err = deleteCertificates(*client)
	if err != nil {
		t.Errorf("error adding app certificate: %v", err)
	}
}

func TestInstall(t *testing.T) {
	cfg, err := getConfig()
	if err != nil {
		t.Fatalf("error loading config: %v", err)
	}

	client, err := NewMockWebSocketClient(cfg)
	if err != nil {
		t.Fatalf("error creating the mock websocket client: %v", err)
	}

	err = client.Install()
	if err != nil {
		t.Errorf("Install() test failed: %v", err)
	}
}

func TestLogin(t *testing.T) {
	cfg, err := getConfig()
	if err != nil {
		t.Fatalf("error loading config: %v", err)
	}

	client, err := NewMockWebSocketClient(cfg)
	if err != nil {
		t.Fatalf("error creating the mock websocket client: %v", err)
	}

	// test with a valid test ApiKey
	client.Cfg.ApiKey = "test"
	err = client.Login()
	if err != nil {
		t.Fatalf("Login() test failed: %v", err)
	}

	client.Cfg.ApiKey = ""
	client.Cfg.Username = "admin"
	client.Cfg.Password = "admin"
	// test with valid test username and password
	err = client.Login()
	if err != nil {
		t.Fatalf("Login() test failed: %v", err)
	}

	// test with empty credentials
	client.Cfg.Username = ""
	client.Cfg.Password = ""
	err = client.Login()
	if err == nil {
		t.Fatalf("expected Login() to fail due to no valid credentials")
	}
}

func TestPostInstall(t *testing.T) {
	cfg, err := getConfig()
	if err != nil {
		t.Fatalf("error loading config: %v", err)
	}

	client, err := NewMockWebSocketClient(cfg)
	if err != nil {
		t.Fatalf("error creating the mock websocket client: %v", err)
	}

	certsList[cfg.CertName()] = 101
	err = client.PostInstall()
	if err != nil {
		t.Errorf("PostInstall() test failed: %v", err)
	}
}

func TestPreInstall(t *testing.T) {
	cfg, err := getConfig()
	if err != nil {
		t.Fatalf("error loading config: %v", err)
	}

	client, err := NewMockWebSocketClient(cfg)
	if err != nil {
		t.Fatalf("error creating the mock websocket client: %v", err)
	}

	err = client.PreInstall()
	if err != nil {
		t.Errorf("PreInstall() test failed: %v", err)
	}
}

func TestRestartUI(t *testing.T) {
	cfg, err := getConfig()
	if err != nil {
		t.Errorf("error loading config: %v", err)
	}
	client, err := NewMockWebSocketClient(cfg)
	if err != nil {
		t.Fatalf("error creating the mock websocket client: %v", err)
	}

	err = restartUI(client)
	if err != nil {
		t.Errorf("error testing app restart: %v", err)
	}
}
