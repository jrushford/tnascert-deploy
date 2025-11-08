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

package restapi

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"tnascert-deploy/config"
)

// used for mock data responses.
type MockRoundTripper struct {
	Response *http.Response
	Err      error
}

// returns a mock data response.
func (m *MockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Response, nil
}

// loads the test configuration.
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

// builds and returns a new mock client with a round tripper to provide mock data responses
func NewClientWithMockRoundTripper(cfg *config.Config, rt *MockRoundTripper) (*TrueNASRest, error) {
	serverURL := strings.TrimRight(cfg.ServerURL(), "/") + EndPoint

	httpClient := &http.Client{
		Transport: rt,
	}

	rest_client := TrueNASRest{
		Url:        serverURL,
		VerifySSL:  cfg.TlsSkipVerify,
		HttpClient: httpClient,
		Cfg:        cfg,
	}
	return &rest_client, nil
}

// builds a new mock round tripper
func NewMockRoundTripper(statusCode int, body string) *MockRoundTripper {
	return &MockRoundTripper{
		Response: &http.Response{
			StatusCode: statusCode,
			Body:       io.NopCloser(bytes.NewBufferString(body)),
			Header:     make(http.Header),
		},
	}
}

func TestAddAsUICertificate(t *testing.T) {
	cfg, err := getConfig()
	if err != nil {
		t.Errorf("loading the test config file failed: %v", err)
	}
	cfg.Debug = true

	certsList["tnas-cert-deploy-2021-10-28-1761686579"] = 1
	certsList["tnas-cert-deploy-2020-10-28-1777168992"] = 2
	if certName == "" {
		certName = cfg.CertName()
	}
	certsList[certName] = 100

	mockRT := NewMockRoundTripper(http.StatusOK, "")
	mockClient, err := NewClientWithMockRoundTripper(cfg, mockRT)
	if err != nil {
		t.Errorf("creating the mock client failed: %v", err)
	}
	err = addAsUICertificate(mockClient)
	if err != nil {
		t.Errorf("addAsUICertificate() test failed: %v", err)
	}
}

func TestAddAsFTPCertificate(t *testing.T) {
	cfg, err := getConfig()
	if err != nil {
		t.Errorf("loading the test config file failed: %v", err)
	}
	cfg.Debug = true

	mockRT := NewMockRoundTripper(http.StatusOK, "")
	mockClient, err := NewClientWithMockRoundTripper(cfg, mockRT)
	if err != nil {
		t.Errorf("creating the mock client failed: %v", err)
	}
	err = addAsFTPCertificate(mockClient)
	if err != nil {
		t.Errorf("addAsFTPCertificate() failed: %v", err)
	}
}

func TestClientLogin(t *testing.T) {
	cfg, err := getConfig()
	if err != nil {
		t.Errorf("loading the test config file failed: %v", err)
	}
	cfg.Debug = true

	// 200 ok response
	mockRT := NewMockRoundTripper(http.StatusOK, `{"pong"}`)
	mockClient, err := NewClientWithMockRoundTripper(cfg, mockRT)

	// 200 response
	err = mockClient.Login()
	if err != nil {
		t.Errorf("login() with test failed: %v", err)
	}

	// 401 response
	mockRT = NewMockRoundTripper(http.StatusUnauthorized, `{"unauthorized"}`)
	mockClient, err = NewClientWithMockRoundTripper(cfg, mockRT)
	err = mockClient.Login()
	if err == nil {
		t.Errorf("expected a login failure: %v", err)
	}
}

func TestClose(t *testing.T) {
	cfg, err := getConfig()
	if err != nil {
		t.Errorf("loading the test config file failed: %v", err)
	}
	cfg.Debug = true

	mockRT := NewMockRoundTripper(http.StatusOK, "")
	mockClient, err := NewClientWithMockRoundTripper(cfg, mockRT)
	if err != nil {
		t.Errorf("creating the mock client failed: %v", err)
	}
	err = mockClient.Close()
	if err != nil {
		t.Errorf("PostInstall() Close() test failed: %v", err)
	}
}

func TestDeleteCertificate(t *testing.T) {
	cfg, err := getConfig()
	if err != nil {
		t.Errorf("loading the test config file failed: %v", err)
	}
	cfg.Debug = true

	if certName == "" {
		certName = cfg.CertName()
	}

	certsList["tnas-cert-deploy-2021-10-28-1761686579"] = 1
	certsList["tnas-cert-deploy-2020-10-28-1777168992"] = 2
	_, ok := certsList[certName]
	if !ok {
		certsList[certName] = 100
	}
	mockRT := NewMockRoundTripper(http.StatusOK, "")
	mockClient, err := NewClientWithMockRoundTripper(cfg, mockRT)
	if err != nil {
		t.Errorf("creating the mock client failed: %v", err)
	}
	err = deleteCertificates(mockClient)
	if err != nil {
		t.Errorf("deleteCertificate() test failed: %v", err)
	}
}

func TestGetCertificateList(t *testing.T) {
	cfg, err := getConfig()
	if err != nil {
		t.Errorf("loading the test config file failed: %v", err)
	}
	cfg.Debug = true

	certs := `
		[
			{"id": 1, "name": "tnas-cert-deploy-2021-10-28-1761686579"},
			{"id": 2, "name": "tnas-cert-deploy-2020-10-28-1777168992"}
		]
	`

	mockRT := NewMockRoundTripper(http.StatusOK, certs)
	mockClient, err := NewClientWithMockRoundTripper(cfg, mockRT)
	if err != nil {
		t.Errorf("creating the mock client failed: %v", err)
	}
	err = getCertificateList(mockClient)
	if err != nil {
		t.Errorf("expected certificate to not be found in certificates list: %v", err)
	}

	mockRT = NewMockRoundTripper(http.StatusOK, certs)
	mockClient, err = NewClientWithMockRoundTripper(cfg, mockRT)
	certsList[certName] = 100
	err = getCertificateList(mockClient)
	if err != nil {
		t.Errorf("getCertificateList() test failed: %v", err)
	}
}

func TestImportCertificate(t *testing.T) {
	cfg, err := getConfig()
	if err != nil {
		t.Errorf("loading the test config file failed: %v", err)
	}
	cfg.Debug = true

	certs := fmt.Sprintf("[{\"id\": 1, \"name\": \"%s\"},{\"id\": 2, \"name\": \"tnas-cert-deploy\"}]", cfg.CertName())
	mockRT := NewMockRoundTripper(http.StatusOK, certs)
	mockClient, err := NewClientWithMockRoundTripper(cfg, mockRT)
	if err != nil {
		t.Errorf("creating the mock client failed: %v", err)
	}
	err = importCertificate(mockClient)
	if err != nil {
		t.Errorf("importCertificate() test  failed: %v", err)
	}
}

func TestInstall(t *testing.T) {
	cfg, err := getConfig()
	if err != nil {
		t.Errorf("loading the test config file failed: %v", err)
	}
	cfg.Debug = true
	certList := fmt.Sprintf("[{\"id\": 1, \"name\": \"%s\"},{\"id\": 2, \"name\": \"tnas-cert-deploy\"}]", cfg.CertName())

	mockRT := NewMockRoundTripper(http.StatusOK, certList)
	mockClient, err := NewClientWithMockRoundTripper(cfg, mockRT)
	if err != nil {
		t.Errorf("creating the mock client failed: %v", err)
	}
	err = mockClient.Install()
	if err != nil {
		t.Errorf("Install() test failed: %v", err)
	}
}

func TestNewClient(t *testing.T) {
	cfg, err := getConfig()
	if err != nil {
		t.Errorf("loading the test config file failed: %v", err)
	}
	cfg.Debug = true

	// with ApiKey set from the config file.
	_, err = NewClient(cfg)
	if err != nil {
		t.Errorf("NewClient() test with ApiKey failed: %v", err)
	}

	// with an empty ApiKey and UserName and Password are used
	cfg.ApiKey = ""
	cfg.Username = "root"
	cfg.Password = "password"
	_, err = NewClient(cfg)
	if err != nil {
		t.Errorf("NewClient() test with UserName and Password failed: %v", err)
	}

	// with no credentials at all
	cfg.ApiKey = ""
	cfg.Username = ""
	cfg.Password = ""
	_, err = NewClient(cfg)
	if err == nil {
		t.Errorf("NewClient() test with no credentials should have failed.")
	}
}

func TestPostInstall(t *testing.T) {
	cfg, err := getConfig()
	if err != nil {
		t.Errorf("loading the test config file failed: %v", err)
	}

	// only run the addASUICertificate()
	cfg.AddAsUiCertificate = true
	cfg.AddAsFTPCertificate = false
	cfg.AddAsAppCertificate = false

	certsList[certName] = 100
	mockRT := NewMockRoundTripper(http.StatusOK, "")
	mockClient, err := NewClientWithMockRoundTripper(cfg, mockRT)
	if err != nil {
		t.Errorf("creating the mock client failed: %v", err)
	}
	err = mockClient.PostInstall()
	if err != nil {
		t.Errorf("PostInstall() addAsUICertificate() test failed: %v", err)
	}

	// run only the AddAsFTPCertificate()
	cfg, err = getConfig()
	if err != nil {
		t.Errorf("loading the test config file failed: %v", err)
	}

	// only run the addASUICertificate()
	cfg.AddAsUiCertificate = false
	cfg.AddAsFTPCertificate = true
	cfg.AddAsAppCertificate = false

	certsList[certName] = 100
	mockRT = NewMockRoundTripper(http.StatusOK, "")
	mockClient, err = NewClientWithMockRoundTripper(cfg, mockRT)
	if err != nil {
		t.Errorf("creating the mock client failed: %v", err)
	}
	err = mockClient.PostInstall()
	if err != nil {
		t.Errorf("PostInstall() addAsFTPCertificate() test failed: %v", err)
	}

	mockResp := `
		{
			"network": {
				"certificate_id": 5,
				"external_ssh_port": 22
			}
		}
	`

	// only run the addASAppCertificate() with empty AppList
	cfg.AddAsUiCertificate = false
	cfg.AddAsFTPCertificate = false
	cfg.AddAsAppCertificate = true

	certsList[certName] = 100
	mockRT = NewMockRoundTripper(http.StatusOK, mockResp)
	mockClient, err = NewClientWithMockRoundTripper(cfg, mockRT)
	if err != nil {
		t.Errorf("creating the mock client failed: %v", err)
	}
	mockClient.Version = "TrueNAS-SCALE-24.10.2.4"

	err = mockClient.PostInstall()
	if err != nil {
		t.Errorf("PostInstall() addAsAppCertificate() (empty AppLIst) test failed: %v", err)
	}

	// now add an app
	cfg.AppList = "gitea"
	err = mockClient.PostInstall()
	if err != nil {
		t.Errorf("PostInstall() addAsAppCertificate() (AppList) test failed: %v", err)
	}

	// app was not configured with a certificate
	noCertResp := `
		{
			"network": {
				"external_ssh_port": 22
			}
		}
	`
	mockRT = NewMockRoundTripper(http.StatusOK, noCertResp)
	mockClient, err = NewClientWithMockRoundTripper(cfg, mockRT)
	if err != nil {
		t.Errorf("creating the mock client failed: %v", err)
	}
	err = mockClient.PostInstall()
	if err != nil {
		t.Errorf("PostInstall() addAsAppCertificate() app with no cert configured test failed: %v", err)
	}

	// version is not TrueNAS-SCALE
	mockClient.Version = "TrueNAS-CORE-8.1"
	err = mockClient.PostInstall()
	if err != nil {
		t.Errorf("PostInstall() addAsAppCertificate() (AppList) test failed: %v", err)
	}
}

func TestPreInstall(t *testing.T) {
	versionBody := `{
 		"version": "TrueNAS-SCALE-24.10.2.4",
 		"hostname": "nas01.mydomain.com",
 		"timezone": "America/Los_Angeles",
 		"system_manufacturer": "QEMU",
 		"ecc_memory": true
	}`

	cfg, err := getConfig()
	if err != nil {
		t.Errorf("loading the test config file failed: %v", err)
	}
	cfg.Debug = true

	mockRT := NewMockRoundTripper(http.StatusOK, versionBody)
	mockClient, err := NewClientWithMockRoundTripper(cfg, mockRT)
	err = mockClient.PreInstall()
	if err != nil {
		t.Errorf("PreInstall() test failed: %v", err)
	}
}

func TestRestartUI(t *testing.T) {
	cfg, err := getConfig()
	if err != nil {
		t.Errorf("loading the test config file failed: %v", err)
	}

	mockRT := NewMockRoundTripper(http.StatusOK, "")
	mockClient, err := NewClientWithMockRoundTripper(cfg, mockRT)
	if err != nil {
		t.Errorf("creating the mock client failed: %v", err)
	}
	err = restartUI(mockClient)
	if err != nil {
		t.Errorf("restartUI() test failed: %v", err)
	}
}
