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
	"encoding/json"
	"fmt"
	"github.com/ncruces/go-strftime"
	"log"
	"os"
	"reflect"
	"strings"
	"time"
	"tnascert-deploy/config"
	"tnascert-deploy/mock"
	"truenas_api/truenas_api"
)

type CertificateCreateResponse struct {
	JsonRPC string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Result  int    `json:"result"`
}

type CertificateListResponse struct {
	JsonRPC string                   `json:"jsonrpc"`
	ID      int                      `json:"id"`
	Result  []map[string]interface{} `json:"result"`
}

const endpoint = "api/current"

// certificate list obtained from TrueNAS client or the mock client
var certsList map[string]int64 = map[string]int64{}

// Client interface
type Client interface {
	Login(username string, password string, apiKey string) error
	Call(method string, timeout int64, params interface{}) (json.RawMessage, error)
}

// uses the configuration 'Environment' setting to get either a truenas_api.Client or a mock.Client used for testing.
func NewClient(serverURL string, cfg *config.Config) (Client, string, error) {

	if cfg.Environment == "production" {
		log.Println("Using production environment")
		certName := cfg.CertBasename + strftime.Format("-%Y-%m-%d-%s", time.Now())
		client, err := truenas_api.NewClient(serverURL, cfg.TlsSkipVerify)
		if err != nil {
			return client, certName, fmt.Errorf("NewClient(): %v", err)
		} else {
			return client, certName, nil
		}
	} else if cfg.Environment == "test" {
		log.Println("NewClient(): Using test environment")
		certName := mock.GetCertName(cfg)
		client, err := mock.NewClient(serverURL, cfg.TlsSkipVerify)
		if err != nil {
			return client, certName, fmt.Errorf("NewClient(): %v", err)
		} else {
			return client, certName, nil
		}
	}
	return nil, "", fmt.Errorf("NewClient(): Invalid environment")
}

// login with an API key
func clientLogin(client Client, cfg *config.Config) error {
	username, password := "", ""
	if cfg.Api_key == "" {
		return fmt.Errorf("NewClient(): No api key")
	}
	apikey := cfg.Api_key
	err := client.Login(username, password, apikey)
	if err == nil {
		log.Println("clientLogin(): successfully logged in")
		return nil
	}
	return fmt.Errorf("clientLogin: %v", err)
}

// create the certificate in TrueNAS
func createCertificate(client Client, certName string, cfg *config.Config) error {
	// read in the certificate data
	certPem, err := os.ReadFile(cfg.FullChainPath)
	if err != nil {
		return fmt.Errorf("createCertificate(): could not load the certificate, %v", err)
	}
	// read in the private key data
	keyPem, err := os.ReadFile(cfg.Private_key_path)
	if err != nil {
		return fmt.Errorf("createCertificate(): could not load the private key, %v", err)
	}

	if cfg.Debug {
		log.Printf("createCertificate(): install certificate: %s", certName)
	}

	params := map[string]string{"name": certName, "certificate": string(certPem),
		"privatekey": string(keyPem), "create_type": "CERTIFICATE_CREATE_IMPORTED"}
	args := []map[string]string{params}

	// call the api to create and deploy the certificate
	resp, err := client.Call("certificate.create", 10, args)
	if err != nil {
		return fmt.Errorf("createCertificate(): %v", err)
	} else {
		if cfg.Debug {
			log.Printf("createCertificate(): certificate.create response was: %s", string(resp))
		}
		var response CertificateCreateResponse
		err = json.Unmarshal(resp, &response)
		if err != nil {
			return fmt.Errorf("createCertificate(): %v", err)
		}
		log.Printf("createCertificate(): Job created id: %v", response.Result)
	}
	return nil
}

// poll and save all deployed certificates matching our Cert_basename
// including the newly created certificate
func loadCertificateList(client Client, certName string, cfg *config.Config) error {
	var inlist bool = false
	var count = 0

	// poll until the list shows up with the newly created certificate or give up after 10 seconds.
	for !inlist && count != 10 {
		count++
		args := []string{}
		resp, err := client.Call("app.certificate_choices", 10, args)
		if err != nil {
			return err
		}
		if cfg.Debug {
			log.Printf("resp: %v", string(resp))
		}

		var response CertificateListResponse
		err = json.Unmarshal(resp, &response)
		if err != nil {
			return err
		}

		for _, v := range response.Result {
			var cert = v
			_, ok := certsList[cert["name"].(string)]
			// add certificate to the cert_list if not already there
			// and skipping those that do not match the cert_basename
			if !ok {
				var name string = cert["name"].(string)
				idValue := cert["id"].(float64)
				id := int64(idValue)
				// only add certs that match the Cert_basename to the list
				if strings.HasPrefix(name, cfg.CertBasename) {
					certsList[name] = id
					if cfg.Debug {
						log.Printf("certificate name: %v, is: %d", cert["name"], id)
					}
				}
			}
			if id, ok := certsList[certName]; ok == true {
				log.Printf("found new certificate: %v, id: %d", cert["name"], id)
				inlist = true
			}
		}
		time.Sleep(1 * time.Second)
	}

	if !inlist {
		return fmt.Errorf("search timeout, certificate %s was not deployed", certName)
	} else {
		log.Printf("certificate %s deployed successfully", certName)
	}
	return nil
}

func InstallCertificate(cfg *config.Config) error {
	var serverURL = fmt.Sprintf("%s://%s:%d/%s", cfg.Protocol, cfg.ConnectHost, cfg.Port, endpoint)
	var certName = cfg.CertBasename + strftime.Format("-%Y-%m-%d-%s", time.Now())

	//
	client, certName, err := NewClient(serverURL, cfg)
	if cfg.Debug {
		log.Println("client is Type:", reflect.TypeOf(client))
	}
	log.Printf("installing certificate: %s", certName)

	if err != nil {
		return fmt.Errorf("NewClient() error: %v", err)
	}
	// login
	err = clientLogin(client, cfg)
	if err != nil {
		return err
	}
	// deploy the certificate
	err = createCertificate(client, certName, cfg)
	if err != nil {
		return err
	}

	err = loadCertificateList(client, certName, cfg)
	if err != nil {
		return err
	}

	if cfg.AddAsUiCertificate {
		pmap := make(map[string]int64)
		pmap["ui_certificate"] = certsList[certName]
		args := []map[string]int64{pmap}
		_, err := client.Call("system.general.update", 10, args)
		if err != nil {
			return fmt.Errorf("system.general.update of ui_certificate, %v", err)
		}
		arg := []map[string]interface{}{}
		_, err = client.Call("system.general.ui_restart", 10, arg)
	}

	// if configured to do so, delete old certificates matching the cert basename pattern
	if cfg.DeleteOldCerts {
		for k, v := range certsList {
			if strings.Compare(k, certName) == 0 {
				continue
			}
			arg := []int64{v}
			_, err := client.Call("certificate.delete", 10, arg)
			if err != nil {
				return fmt.Errorf("certificate.delete of certificate, %v failed", err)
			} else {
				log.Printf("certficate %v was deleted", k)
			}
		}
	}

	return nil
}
