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
	"testing"
)

func TestNewConfig(t *testing.T) {
	configFile := "test_files/tnas-cert.ini"

	// test loading the default config section
	cfg, err := New(configFile, "default")
	if cfg == nil || err != nil {
		t.Errorf("New config failed with error: %v", err)
	}
	if cfg.ConnectHost != "nas01.mydomain.com" {
		t.Errorf("Connect_host should be nas01.mydomain.com")
	}
	if cfg.Private_key_path != "test_files/privkey.pem" {
		t.Errorf("Private_key_path should be test_files/privkey.pem")
	}
	if cfg.FullChainPath != "test_files/fullchain.pem" {
		t.Errorf("Fullchain_path should be test_files/fullchain.pem")
	}
	if cfg.Protocol != "wss" {
		t.Errorf("Protocol should be wss")
	}
	if cfg.TlsSkipVerify != false {
		t.Errorf("TLS_skip_verify should be false")
	}
	if cfg.DeleteOldCerts != true {
		t.Errorf("Delete_old_certs should be true")
	}
	if cfg.AddAsUiCertificate != true {
		t.Errorf("Add_as_ui_certificate should be true")
	}

	// test opening  non-existent file
	cfg, err = New("non_existent_file", "default")
	if err == nil && cfg != nil {
		t.Errorf("Exepected an error opening a non-existent file: %v", err)
	}

	// test nas02 config section
	if cfg, err = New(configFile, "nas02"); err != nil {
		t.Errorf("New config failed with error: %v", err)
	}
	if cfg.ConnectHost != "nas02.mydomain.com" {
		t.Errorf("Connect_host should be nas02.mydomain.com")
	}

	// test checkConfig function
	if err = cfg.checkConfig(); err != nil {
		t.Errorf("Check config failed with error: %v", err)
	}

	// test nas03 config section
	if cfg, err = New(configFile, "nas03"); err != nil {
		t.Errorf("New config failed with error: %v", err)
	}
	if cfg.ConnectHost != "nas03.mydomain.com" {
		t.Errorf("Connect_host should be nas02.mydomain.com")
	}

	// test loading a non-existent config section
	if cfg, err = New(configFile, "nas10"); err == nil {
		t.Errorf("New config failed with error: %v", err)
	}
}
