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
	"github.com/truenas/api_client_golang/truenas_api"
	"log"
	"os"
	"reflect"
	"strings"
	"time"
	"tnascert-deploy/config"
	"tnascert-deploy/mock"
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
	CallWithJob(method string, params interface{}, callback func(progress float64, state string, desc string)) (*truenas_api.Job, error)
	Close() error
	SubscribeToJobs() error
}

// uses the configuration 'Environment' setting to get either a truenas_api.Client or a mock.Client used for testing.
func NewClient(serverURL string, cfg *config.Config) (Client, string, error) {

	if cfg.Environment == "production" {
		log.Println("using the production environment")
		certName := cfg.CertBasename + strftime.Format("-%Y-%m-%d-%s", time.Now())
		client, err := truenas_api.NewClient(serverURL, cfg.TlsSkipVerify)
		if err != nil {
			return client, certName, fmt.Errorf("error connecting to the server, %v", err)
		} else {
			return client, certName, nil
		}
	} else if cfg.Environment == "test" {
		log.Println("using test environment")
		certName := mock.GetCertName(cfg)
		client, err := mock.NewClient(serverURL, cfg.TlsSkipVerify)
		if err != nil {
			return client, certName, fmt.Errorf("NewClient(): %v", err)
		} else {
			return client, certName, nil
		}
	}
	return nil, "", fmt.Errorf("invalid environment")
}

// login with an API key
func clientLogin(client Client, cfg *config.Config) error {
	username, password := "", ""
	if cfg.Api_key == "" {
		return fmt.Errorf("login failure, o api key")
	}
	apikey := cfg.Api_key
	err := client.Login(username, password, apikey)
	if err == nil {
		log.Println("successfully logged in")
		return nil
	}
	return fmt.Errorf("login failed, %v", err)
}

// create the certificate in TrueNAS
func createCertificate(client Client, certName string, cfg *config.Config) error {
	// read in the certificate data
	certPem, err := os.ReadFile(cfg.FullChainPath)
	if err != nil {
		return fmt.Errorf("could not load the pem encoded certificate, %v", err)
	}
	// read in the private key data
	keyPem, err := os.ReadFile(cfg.Private_key_path)
	if err != nil {
		return fmt.Errorf("could not load the pem encoded private key, %v", err)
	}

	if cfg.Debug {
		log.Printf("install certificate: %s", certName)
	}

	if err = client.SubscribeToJobs(); err != nil {
		return fmt.Errorf("unable to subscribe to job notifications, %v", err)
	}

	params := map[string]string{"name": certName, "certificate": string(certPem),
		"privatekey": string(keyPem), "create_type": "CERTIFICATE_CREATE_IMPORTED"}
	args := []map[string]string{params}

	// call the api to create and deploy the certificate
	job, err := client.CallWithJob("certificate.create", args, func(progress float64, state string, desc string) {
		log.Printf("Job Progress: %.2f%%, State: %s, Description: %s", progress, state, desc)
	})
	if err != nil {
		return fmt.Errorf("failed to create the certificate job,  %v", err)
	}

	log.Printf("started the certificate creation job with ID: %d", job.ID)

	// Monitor the progress of the job.
	for !job.Finished {
		select {
		case progress := <-job.ProgressCh:
			log.Printf("Job progress: %.2f%%", progress)
		case err := <-job.DoneCh:
			if err != "" {
				return fmt.Errorf("Job failed: %v", err)
			} else {
				log.Println("Job completed successfully!")
			}
		}
	}

	return nil
}

// poll and save all deployed certificates matching our Cert_basename
// including the newly created certificate
func loadCertificateList(client Client, certName string, cfg *config.Config) error {
	var inlist bool = false

	args := []string{}
	resp, err := client.Call("app.certificate_choices", 10, args)
	if err != nil {
		return fmt.Errorf("failed to get a certifcate list from the server,  %v", err)
	}
	if cfg.Debug {
		log.Printf("certificate list response: %v", string(resp))
	}

	var response CertificateListResponse
	err = json.Unmarshal(resp, &response)
	if err != nil {
		return err
	}

	// range over the list obtained from the server and build up a local
	// certificate list
	for _, v := range response.Result {
		var cert = v
		_, ok := certsList[cert["name"].(string)]
		// add certificate to the certificate list if not already there
		// and skipping those that do not match the certificate basename
		if !ok {
			var name string = cert["name"].(string)
			idValue := cert["id"].(float64)
			id := int64(idValue)
			// only add certs that match the Cert_basename to the list
			if strings.HasPrefix(name, cfg.CertBasename) {
				certsList[name] = id
				if cfg.Debug {
					log.Printf("cert list, name: %v, id: %d", cert["name"], id)
				}
			}
		}
		if id, ok := certsList[certName]; ok == true {
			log.Printf("found new certificate, %v, id: %d", cert["name"], id)
			inlist = true
		}
	}

	if !inlist {
		return fmt.Errorf("certificate search failed, certificate %s was not deployed", certName)
	} else {
		log.Printf("certificate %s deployed successfully", certName)
	}
	return nil
}

func InstallCertificate(cfg *config.Config) error {
	var serverURL = fmt.Sprintf("%s://%s:%d/%s", cfg.Protocol, cfg.ConnectHost, cfg.Port, endpoint)
	var certName = cfg.CertBasename + strftime.Format("-%Y-%m-%d-%s", time.Now())
	var activated = false

	// connect to the server websocket endpoint
	client, certName, err := NewClient(serverURL, cfg)
	if cfg.Debug {
		log.Println("client is Type:", reflect.TypeOf(client))
	}
	log.Printf("installing certificate: %s", certName)
	defer client.Close()

	if err != nil {
		return fmt.Errorf("failed to connet to the server, %v", err)
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

	// load the certificates list
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
		activated = true
		log.Printf("%s is now the active UI certificate", certName)
	}

	if activated {
		// if configured to do so, delete old certificates matching the cert basename pattern
		if cfg.DeleteOldCerts && activated == true {
			for k, v := range certsList {
				if strings.Compare(k, certName) == 0 {
					continue
				}
				arg := []int64{v}
				_, err := client.Call("certificate.delete", 10, arg)
				if err != nil {
					return fmt.Errorf("certificate deletion failed, %v", err)
				} else {
					log.Printf("certficate %v was deleted", k)
				}
			}
		}
		// restart the UI
		arg := []map[string]interface{}{}
		_, err = client.Call("system.general.ui_restart", 10, arg)
		if err != nil {
			return fmt.Errorf("failed to restart the UI, %v", err)
		} else {
			log.Println("the UI has been restarted")
		}
	} else {
		log.Printf("%s was not activated as the UI certificate therefore no certificates will be deleted", certName)
	}

	return nil
}
