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

func TestDeployPkg(t *testing.T) {
	configFile := "test_files/tnas-cert.ini"

	cfg, err := config.New(configFile, "default")
	if cfg == nil || err != nil {
		t.Errorf("New config failed with error: %v", err)
	}
	certName := cfg.CertName()
	fmt.Printf("certName: %s\n", certName)

	serverURL := fmt.Sprintf("%s://%s:%d/%s", cfg.Protocol, cfg.ConnectHost, cfg.Port, endpoint)
	client, err := NewClient(serverURL, cfg)
	if err != nil {
		t.Errorf("New client failed with error: %v", err)
	}

	err = clientLogin(client, cfg)
	if err != nil {
		t.Errorf("client login failed with error: %v", err)
	}

	err = createCertificate(client, cfg)
	if err != nil {
		t.Errorf("create certificate failed with error: %v", err)
	}

	err = loadCertificateList(client, cfg)
	if err != nil {
		t.Errorf("load certificate list failed with error: %v", err)
	}

	err = addAsFTPCertificate(client, cfg)
	if err != nil {
		t.Errorf("addAsFTPCertificate failed with error: %v", err)
	}

	result, err := addAsUICertificate(client, cfg)
	if err != nil && result != true {
		t.Errorf("addAsUICertificate failed with error: %v", err)
	}

	err = InstallCertificate(cfg)
	if err != nil {
		t.Errorf("install certificate failed with error: %v", err)
	}
}
