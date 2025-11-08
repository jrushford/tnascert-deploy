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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strings"
	"time"
	"tnascert-deploy/clients"
	"tnascert-deploy/config"
)

const EndPoint = "/api/v2.0"

// certificates list
var certsList = map[string]int64{}

// name of the certificate to be installed
var certName string

type AuthRoundTripper struct {
	Transport http.RoundTripper
	AuthToken string
}

// adds necessary headers to each request
func (art *AuthRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone the request to avoid modifying the original request object
	newReq := req.Clone(req.Context())

	// Add the Authorization header
	newReq.Header.Set("Authorization", art.AuthToken)

	newReq.Header.Set("Content-Type", "application/json")

	// Delegate the actual request execution to the underlying transport
	return art.Transport.RoundTrip(newReq)
}

type TrueNASRest struct {
	Url        string
	VerifySSL  bool
	HttpClient *http.Client
	Version    string
	Cfg        *config.Config
}

// noop for truenasrest
func (c *TrueNASRest) Close() error {
	if c.Cfg.Debug {
		log.Printf("close the client connection, %v", c.Url)
	}
	return nil
}

func (c *TrueNASRest) Install() error {
	if c.Cfg.Debug {
		log.Println("running install tasks")
	}
	certName = c.Cfg.CertName()

	// import the certificate
	err := importCertificate(c)
	if err != nil {
		return fmt.Errorf("could not import certificate: %v", err)
	}

	// collect a certificate list
	err = getCertificateList(c)
	if err != nil {
		return fmt.Errorf("could not get certificate list: %v", err)
	}
	return nil
}

func (c *TrueNASRest) Login() error {
	if c.Cfg.Debug {
		log.Printf("running login task")
	}

	r, err := http.NewRequest(http.MethodGet, c.Url+"/core/ping", nil)
	res, err := c.HttpClient.Do(r)
	if err != nil {
		return fmt.Errorf("login error %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Errorf("login error: %v", res.Status)
	}
	return nil
}

// constructor
func NewClient(cfg *config.Config) (clients.Client, error) {
	var authToken string
	serverURL := strings.TrimRight(cfg.ServerURL(), "/") + EndPoint

	if cfg.ApiKey != "" {
		authToken = "Bearer " + cfg.ApiKey
	} else if cfg.Username != "" && cfg.Password != "" {
		plainText := cfg.Username + ":" + cfg.Password
		authToken = "Basic " + base64.StdEncoding.EncodeToString([]byte(plainText))
	} else {
		return nil, fmt.Errorf("no valid credentials have been supplied")
	}

	authTransport := &AuthRoundTripper{
		Transport: http.DefaultTransport,
		AuthToken: authToken,
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("creating the cookie jar failed %v", err)
	}

	httpClient := &http.Client{
		Transport: authTransport,
		Jar:       jar,
	}

	rest_client := TrueNASRest{
		Url:        serverURL,
		VerifySSL:  cfg.TlsSkipVerify,
		HttpClient: httpClient,
		Cfg:        cfg,
	}
	return &rest_client, nil
}

func (c *TrueNASRest) PostInstall() error {
	var activated bool = false
	if c.Cfg.Debug {
		log.Println("running post install tasks")
	}

	// update the UI to use the newly
	// imported certificate
	if c.Cfg.AddAsUiCertificate {
		err := addAsUICertificate(c)
		if err != nil {
			return fmt.Errorf("failed to set %s as the UI certificate: %v", certName, err)
		}
		activated = true
	}

	// update the FTP service to use the newly
	// imported certificate
	if c.Cfg.AddAsFTPCertificate {
		err := addAsFTPCertificate(c)
		if err != nil {
			return fmt.Errorf("failed to set %s as the FTP certificate: %v", certName, err)
		}
	}

	if c.Cfg.AddAsAppCertificate {
		if c.Cfg.AppList == "" {
			log.Printf("the AppList config is empty, no apps to check")
			return nil
		} else {
			if strings.HasPrefix(c.Version, "TrueNAS-SCALE") {
				appList := strings.Split(c.Cfg.AppList, ",")
				for _, app := range appList {
					err := c.addAsAppCertificate(app)
					if err != nil {
						log.Printf("failed to add the '%s' certificate to the '%s' app: %v", certName, app, err)
					}
				}
			} else {
				log.Printf("will not process any apps as the system is not running TrueNAS-SCALE")
			}
		}
	}

	if activated {
		if c.Cfg.DeleteOldCerts {
			// give a wait of 5 seconds before deleting old certificates.
			// to insure app updates have completed.
			time.Sleep(5 * time.Second)
			err := deleteCertificates(c)
			if err != nil {
				return fmt.Errorf("error deleting old certificates: %v", err)
			} else {
				log.Println("successfully deleted old certificates")
			}
		}

		// restart the UI
		err := restartUI(c)
		if err != nil {
			return fmt.Errorf("failed to restart the UI")
		} else {
			log.Println("successfully restarted the UI")
		}
	}
	return nil
}

func (c *TrueNASRest) PreInstall() error {
	if c.Cfg.Debug {
		log.Println("running preinstall tasks")
	}

	err := getSystemInfo(c)
	if err != nil {
		return fmt.Errorf("could not get system info: %v", err)
	}

	err = clients.VerifyCertificateKeyPair(c.Cfg.FullChainPath, c.Cfg.PrivateKeyPath)
	if err != nil {
		return fmt.Errorf("failed certificate verification: %v", err)
	}

	return nil
}

func (c *TrueNASRest) addAsAppCertificate(appName string) error {
	log.Printf("adding %s with ID %d to the %s app", certName, certsList[certName], appName)

	// get the app configuration
	app := []byte("\"" + appName + "\"")
	req, err := http.NewRequest(http.MethodPost, c.Url+"/app/config", bytes.NewBuffer(app))
	if err != nil {
		return fmt.Errorf("error creating application configuration request for '%s': %v", appName, err)
	}
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error executing the application configuration request for '%s': %v", appName, err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("application configuration request for '%s' failed: %v", appName, resp.Status)
	}
	var respData interface{}
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		return fmt.Errorf("error decoding the application config info: %v", err)
	}

	// check the App for an existing certificate.  If it's not currently using one, we are not going
	// to add one to the App
	cMap, ok := respData.(map[string]interface{})
	nMap, ok := cMap["network"]
	if ok {
		cfg := nMap.(map[string]interface{})
		v, found := cfg["certificate_id"]
		if v == nil || !found {
			log.Printf("the '%s' application is currently not using a certificate, will not add one", appName)
			return nil
		}
	}
	resp.Body.Close()

	if m, ok := nMap.(map[string]interface{}); ok {
		certId := certsList[certName]
		m["certificate_id"] = certId
		uMap := map[string]map[string]interface{}{
			"values": {
				"network": m,
			},
		}
		jsonUpdate, err := json.Marshal(uMap)
		if err != nil {
			return fmt.Errorf("error marshaling an update message for the '%s' app: %v", appName, err)
		}
		if c.Cfg.Debug {
			log.Printf("update message for '%s' app: %s\n", appName, string(jsonUpdate))
		}
		body := bytes.NewBuffer([]byte(jsonUpdate))
		req, err = http.NewRequest(http.MethodPut, c.Url+"/app/id/"+appName, body)
		if err != nil {
			return fmt.Errorf("error creating application configuration update for '%s': %v", appName, err)
		}
		resp, err = c.HttpClient.Do(req)
		if err != nil {
			return fmt.Errorf("error executing the application update request for '%s': %v", appName, err)
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("the application update request for '%s' failed: %v", appName, resp.Status)
		}
		defer resp.Body.Close()

		time.Sleep(5 * time.Second)
		log.Printf("updated the  certificate for application '%s' to use %s", appName, certName)
	} else {
		return fmt.Errorf("error obtaining the network configuration for '%s'\n", appName)
	}

	return nil
}

func addAsFTPCertificate(c *TrueNASRest) error {
	if id, ok := certsList[certName]; ok {
		data := struct {
			CertId int64 `json:"ssltls_certificate"`
		}{
			CertId: id,
		}
		jsonData, err := json.Marshal(&data)
		if err != nil {
			return fmt.Errorf("could not marshal ftp update message: %v", err)
		}
		// update active FTP certificate
		req, err := http.NewRequest(http.MethodPut, c.Url+"/ftp", bytes.NewBuffer(jsonData))
		if err != nil {
			return fmt.Errorf("error creating FTP update request: %v", err)
		}
		resp, err := c.HttpClient.Do(req)
		if err != nil {
			return fmt.Errorf("error executing the active FTP update request: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("FTP update request failed: %v", resp.Status)
		} else {
			// wait 5 seconds for the imported certifcate to become available
			time.Sleep(5 * time.Second)
			log.Printf("updated the active FTP certificate to use %s", certName)
		}
		defer resp.Body.Close()
	} else {
		return fmt.Errorf("%s was not found, cannot add it as FTP certificate", certName)
	}
	return nil
}

func addAsUICertificate(client *TrueNASRest) error {
	if id, ok := certsList[certName]; ok {
		data := struct {
			CertId int64 `json:"ui_certificate"`
		}{
			CertId: id,
		}
		jsonData, err := json.Marshal(&data)
		if err != nil {
			return fmt.Errorf("could not marshal ui update message: %v", err)
		}
		// update active UI certificate
		req, err := http.NewRequest(http.MethodPut, client.Url+"/system/general", bytes.NewBuffer(jsonData))
		if err != nil {
			return fmt.Errorf("error creating UI update request: %v", err)
		}
		resp, err := client.HttpClient.Do(req)
		if err != nil {
			return fmt.Errorf("error executing the active UI update request: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("UI update request failed: %v", resp.Status)
		} else {
			// wait 5 seconds for the imported certifcate to become available
			time.Sleep(5 * time.Second)
			log.Printf("updated the active UI certificate to use %s", certName)
		}
		defer resp.Body.Close()
	} else {
		return fmt.Errorf("%s was not found, cannot add it as UI certificate", certName)
	}
	return nil
}

func deleteCertificates(client *TrueNASRest) error {
	log.Printf("deleting old certificates with prefix '%s'", client.Cfg.CertBasename)
	for k, v := range certsList {
		if strings.Compare(k, certName) == 0 {
			log.Printf("skip the deletion of the active UI certificate %s", certName)
			continue
		}
		if strings.HasPrefix(k, client.Cfg.CertBasename) {
			URL := fmt.Sprintf("%s/certificate/id/%d", client.Url, v)
			r, err := http.NewRequest(http.MethodDelete, URL, nil)
			resp, err := client.HttpClient.Do(r)
			if err != nil {
				return fmt.Errorf("error executing certificate deletion: %v", err)
			}
			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("error deleting certificate %s: %v", certName, resp.Status)
			} else {
				log.Printf("deleted certificate %s", k)
			}
			defer resp.Body.Close()
		}
	}

	return nil
}

func getCertificateList(client *TrueNASRest) error {
	// fetch the list a list of certificates
	var respData interface{}
	// certificate list get request
	req, err := http.NewRequest(http.MethodGet, client.Url+"/certificate?limit=0", nil)
	if err != nil {
		return fmt.Errorf("error creating certificate list request: %v", err)
	}
	resp, err := client.HttpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error executing certificate list request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("certificate list request failed: %v", resp.Status)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		return fmt.Errorf("error decoding the certificate list: %v", err)
	}
	// parse the response and build the certificates list
	for _, v := range respData.([]interface{}) {
		if t, ok := v.(map[string]interface{}); ok {
			name := t["name"].(string)
			idi := t["id"]
			id := int64(idi.(float64))
			certsList[name] = id
		}
	}
	if len(certsList) == 0 {
		return fmt.Errorf("no certificates were found in the certificate list")
	}
	if _, ok := certsList[certName]; !ok {
		return fmt.Errorf("certificate %s was not found in certificate list", certName)
	}

	return nil
}

func getSystemInfo(client *TrueNASRest) error {
	// fetch the system information.
	var respData interface{}
	// system info request
	req, err := http.NewRequest(http.MethodGet, client.Url+"/system/info", nil)
	if err != nil {
		return fmt.Errorf("error creating system info request: %v", err)
	}
	resp, err := client.HttpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error executing system info request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("system info request failed: %v", resp.Status)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		return fmt.Errorf("error decoding the system info: %v", err)
	}
	vmap, ok := respData.(map[string]interface{})
	if ok {
		version := vmap["version"]
		client.Version = version.(string)
		log.Printf("%s is running version '%s'", client.Cfg.ConnectHost, client.Version)
	} else {
		log.Printf("%s unable to get the version of TrueNAS", client.Cfg.ConnectHost)
	}
	return nil

}

func importCertificate(client *TrueNASRest) error {
	log.Printf("importing the %s certificate", client.Cfg.CertName())
	certPem, err := os.ReadFile(client.Cfg.FullChainPath)
	if err != nil {
		return fmt.Errorf("error reading the certificate file: %v", err)
	}
	keyPem, err := os.ReadFile(client.Cfg.PrivateKeyPath)
	if err != nil {
		return fmt.Errorf("error reading the private key file: %v", err)
	}

	data := struct {
		Name        string `json:"name"`
		CreateType  string `json:"create_type"`
		Certificate string `json:"certificate"`
		PrivateKey  string `json:"privatekey"`
	}{
		Name:        certName,
		CreateType:  "CERTIFICATE_CREATE_IMPORTED",
		Certificate: string(certPem),
		PrivateKey:  string(keyPem),
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("error marshaling the certificate import message: %v", err)
	}

	// certificate import post request
	req, err := http.NewRequest(http.MethodPost, client.Url+"/certificate", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating certificate import request: %v", err)
	}
	resp, err := client.HttpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error executing the import request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("certificate import request failed: %v", resp.Status)
	} else {
		// wait 5 seconds for the imported certifcate to become available
		time.Sleep(5 * time.Second)
		log.Printf("successfully imported the %s certificate", certName)
	}
	defer resp.Body.Close()

	return nil
}

func restartUI(client *TrueNASRest) error {
	// restart the UI request
	req, err := http.NewRequest(http.MethodGet, client.Url+"/system/general/ui_restart", nil)
	if err != nil {
		return fmt.Errorf("error creating the UI restart request: %v", err)
	}
	resp, err := client.HttpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error executing the UI restart request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to restart the UI: %v", resp.Status)
	}
	defer resp.Body.Close()

	return nil
}
