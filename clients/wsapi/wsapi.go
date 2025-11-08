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
	"encoding/json"
	"fmt"
	"github.com/truenas/api_client_golang/truenas_api"
	"log"
	"os"
	"strings"
	"tnascert-deploy/clients"
	"tnascert-deploy/config"
)

const EndPoint = "/api/current"

// certificates list
var certsList = map[string]int64{}

// name of the certificate to be installed
var certName string

type TrueNASWebSocket struct {
	Url       string
	VerifySSL bool
	WSClient  WSClient
	Version   string
	Cfg       *config.Config
}

type WSClient interface {
	Login(username string, password string, apiKey string) error
	Call(method string, timeout int64, params interface{}) (json.RawMessage, error)
	CallWithJob(method string, params interface{}, callback func(progress float64, state string, desc string)) (*truenas_api.Job, error)
	Close() error
	SubscribeToJobs() error
}

type CertificateListResponse struct {
	JsonRPC string                   `json:"jsonrpc"`
	ID      int                      `json:"id"`
	Result  []map[string]interface{} `json:"result"`
}

func (c TrueNASWebSocket) Close() error {
	err := c.WSClient.Close()
	if err != nil {
		return fmt.Errorf("error closing the websocket client connection: %v", err)
	}
	return nil
}

func (c TrueNASWebSocket) Install() error {
	if c.Cfg.Debug {
		log.Println("running install tasks")
	}
	certName = c.Cfg.CertName()

	// import the certificate
	err := importCertificate(&c)
	if err != nil {
		return fmt.Errorf("could not import certificate: %v", err)
	}

	// collect a certificate list
	err = getCertificateList(&c)
	if err != nil {
		return fmt.Errorf("could not get certificate list: %v", err)
	}

	return nil
}

func (c TrueNASWebSocket) Login() error {
	// preferred login is with the API key
	if c.Cfg.ApiKey != "" {
		if c.Cfg.Debug {
			log.Printf("logging in to %s with the ApiKey", c.Cfg.ConnectHost)
		}
		err := c.WSClient.Login(c.Cfg.Username, c.Cfg.Password, c.Cfg.ApiKey)
		if err != nil {
			return fmt.Errorf("error logging in to %s with the ApiKey: %v", c.Cfg.ConnectHost, err)
		}
	} else if c.Cfg.Username != "" && c.Cfg.Password != "" {
		if c.Cfg.Debug {
			log.Printf("logging in to %s with the Username and Password\n", c.Cfg.ConnectHost)
		}
		err := c.WSClient.Login(c.Cfg.Username, c.Cfg.Password, "")
		if err != nil {
			return fmt.Errorf("error logging in to %s with the Username and Password: %v", c.Cfg.ConnectHost, err)
		}
	} else {
		return fmt.Errorf("you need to specify a valid ApiKey or Username and Password")
	}
	return nil
}

func NewClient(cfg *config.Config) (clients.Client, error) {
	serverURL := strings.TrimRight(cfg.ServerURL(), "/") + EndPoint
	cl, err := truenas_api.NewClient(serverURL, cfg.TlsSkipVerify)
	if err != nil {
		return nil, fmt.Errorf("error: %v", err)
	}

	websocket_client := TrueNASWebSocket{
		Url:       serverURL,
		VerifySSL: cfg.TlsSkipVerify,
		WSClient:  cl,
		Cfg:       cfg,
	}

	return websocket_client, nil
}

func (c TrueNASWebSocket) PostInstall() error {
	var activated bool = false
	if c.Cfg.Debug {
		log.Println("running post install tasks")
	}
	err := getSystemInfo(&c)
	if err != nil {
		return fmt.Errorf("could not get system info: %v", err)
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
					err := addAsAppCertificate(&c, strings.TrimSpace(app))
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
			err := deleteCertificates(c)
			if err != nil {
				log.Printf("error deleting old certificates: %v", err)
			}
		}

		// restart the UI
		err := restartUI(&c)
		if err != nil {
			return fmt.Errorf("failed to restart the UI")
		}
	}

	return nil
}

func (c TrueNASWebSocket) PreInstall() error {
	if c.Cfg.Debug {
		log.Printf("running preinstall tasks")
	}

	err := getSystemInfo(&c)
	if err != nil {
		return fmt.Errorf("could not get system info: %v", err)
	}

	err = clients.VerifyCertificateKeyPair(c.Cfg.FullChainPath, c.Cfg.PrivateKeyPath)
	if err != nil {
		return fmt.Errorf("failed certificate verification: %v", err)
	}

	return nil
}

func addAsAppCertificate(client *TrueNASWebSocket, appName string) error {
	var args []interface{}
	var response map[string]interface{}
	args = []interface{}{appName}
	log.Printf("processing certificate update for the '%s' application\n", appName)

	resp, err := client.WSClient.Call("app.config", client.Cfg.TimeoutSeconds, args)
	if err != nil {
		log.Printf("error retrieving the app config for %s: %v", appName, err)
		return nil
	}
	err = json.Unmarshal(resp, &response)
	if err != nil {
		return fmt.Errorf("error decoding the response for %s: %v", appName, err)
	}

	_, ok := response["error"].(map[string]interface{})
	if ok {
		log.Printf("the '%s' application does not exist or the query failed", appName)
		return nil
	}
	rstMap, ok := response["result"].(map[string]interface{})
	if ok {
		ntwkMap, found := rstMap["network"].(map[string]interface{})
		if found {
			_, exists := ntwkMap["certificate_id"]
			if !exists {
				log.Printf("the '%s' application is currently not using a certificate, will not add one", appName)
				return nil
			}

			// update the certificate id
			ntwkMap["certificate_id"] = certsList[certName]
			updateMap := map[string]map[string]interface{}{
				"values": {
					"network": ntwkMap,
				},
			}
			if client.Cfg.Debug {
				jsonData, err := json.Marshal(updateMap)
				if err != nil {
					log.Printf("error marshaling the update map for '%s' app: %v\n", appName, err)
				}
				log.Printf("app update message for '%s': %s\n", appName, string(jsonData))
			}
			params := [2]interface{}{appName, updateMap}
			job, err := client.WSClient.CallWithJob("app.update", params, func(progress float64, state string, desc string) {
				if client.Cfg.Debug {
					log.Printf("job progress: %.2f%%, state: %s, description: %s", progress, state, desc)
				}
			})
			if err != nil {
				return fmt.Errorf("failed to update the app certificate, %v", err)
			}
			log.Printf("started the app update job with ID: %d", job.ID)

			// Monitor the progress of the job.
			for !job.Finished {
				select {
				case progress := <-job.ProgressCh:
					if client.Cfg.Debug {
						log.Printf("job progress: %.2f%%", progress)
					}
				case err := <-job.DoneCh:
					if err != "" {
						return fmt.Errorf("job failed: %v", err)
					} else {
						log.Println("job completed successfully!")
						break
					}
				}
			}
		}
	}

	log.Printf("updated the certificate for app: %s to use: %s, id: %v", appName, certName, certsList[certName])

	return nil
}

func addAsFTPCertificate(client TrueNASWebSocket) error {
	var certName = client.Cfg.CertName()
	ID, ok := certsList[certName]
	if !ok {
		return fmt.Errorf("certificate %s was not found in the certificates list", certName)
	}
	pmap := map[string]int64{
		"ssltls_certificate": ID,
	}
	args := []interface{}{pmap}
	_, err := client.WSClient.Call("ftp.update", client.Cfg.TimeoutSeconds, args)
	if err != nil {
		return fmt.Errorf("updating the FTP service certificate failed, %v", err)
	} else {
		log.Printf("the FTP service certificate updated successfully to %s", certName)
	}

	return nil
}

func addAsUICertificate(client TrueNASWebSocket) error {
	var certName = client.Cfg.CertName()
	ID, ok := certsList[certName]
	if !ok {
		return fmt.Errorf("certificate %s was not found in the certificates list", certName)
	}
	pmap := map[string]int64{
		"ui_certificate": ID,
	}
	args := []interface{}{pmap}
	_, err := client.WSClient.Call("system.general.update", client.Cfg.TimeoutSeconds, args)
	if err != nil {
		return fmt.Errorf("system.general.update of ui_certificate failed, %v", err)
	}
	return nil
}

func deleteCertificates(client TrueNASWebSocket) error {
	var certName = client.Cfg.CertName()
	_, ok := certsList[certName]
	if !ok {
		return fmt.Errorf("certificate %s was not found in the certificates list", certName)
	}

	for k, v := range certsList {
		if strings.Compare(k, certName) == 0 {
			if client.Cfg.Debug {
				log.Printf("skipping deletion of certificate %v", k)
			}
			continue
		}

		arg := []int64{v}
		job, err := client.WSClient.CallWithJob("certificate.delete", arg, func(progress float64, state string, desc string) {
			if client.Cfg.Debug {
				log.Printf("job progress: %.2f%%, state: %s, description: %s", progress, state, desc)
			}
		})
		if err != nil {
			return fmt.Errorf("certificate deletion failed, %v", err)
		}
		if client.Cfg.Debug {
			log.Printf("deleting old certificate, job info: %v, ", job)
		}
		log.Printf("deleting old certificate %v, with job ID: %d", k, job.ID)

		// Monitor the progress of the job.
		for !job.Finished {
			select {
			case progress := <-job.ProgressCh:
				if client.Cfg.Debug {
					log.Printf("job progress: %.2f%%", progress)
				}
			case err := <-job.DoneCh:
				if err != "" {
					return fmt.Errorf("job failed: %v", err)
				} else {
					log.Printf("job completed successfully, certificate %v was deleted", k)
					break
				}
			}
		}
	}
	return nil
}

func getCertificateList(client *TrueNASWebSocket) error {
	var found = false
	args := []interface{}{}
	resp, err := client.WSClient.Call("app.certificate_choices", client.Cfg.TimeoutSeconds, args)
	if err != nil {
		return fmt.Errorf("certificate list request failed: %v", err)
	}
	if client.Cfg.Debug {
		log.Printf("received certificate list request response: %v", string(resp))
	}
	var response CertificateListResponse
	err = json.Unmarshal(resp, &response)
	if err != nil {
		return fmt.Errorf("error unmarshalling certificate list response: %v", err)
	}

	// range over the list obtained from the server and build up a local
	// certificate list
	for _, v := range response.Result {
		var cert = v
		_, ok := certsList[cert["name"].(string)]
		if client.Cfg.Debug {
			log.Printf("certslist, cert: %s", cert["name"].(string))
		}
		// add certificate to the certificate list if not already there
		// and skipping those that do not match the certificate basename
		if !ok {
			var name = cert["name"].(string)
			idValue := cert["id"].(float64)
			id := int64(idValue)
			// only add certs that match the Cert_basename to the list
			if strings.HasPrefix(name, client.Cfg.CertBasename) {
				certsList[name] = id
				if client.Cfg.Debug {
					log.Printf("cert list, name: %v, id: %d", cert["name"], id)
				}
			}
		}
		if id, ok := certsList[certName]; ok == true {
			log.Printf("found the new certificate, %v, id: %d", cert["name"], id)
			found = true
		}
	}
	if !found {
		return fmt.Errorf("certificate search failed, certificate %s was not deployed", certName)
	} else {
		log.Printf("certificate %s deployed successfully", certName)
	}
	return nil
}

func getSystemInfo(client *TrueNASWebSocket) error {

	res, err := client.WSClient.Call("system.info", 10, []interface{}{})
	if err != nil {
		log.Fatalf("failed to call system.info: %v", err)
	}

	var respData interface{}
	err = json.Unmarshal(res, &respData)
	if err != nil {
		return fmt.Errorf("getSystemInfo(): could not decode response: %v", err)
	}
	respMap, ok := respData.(map[string]interface{})
	if ok {
		resultMap := respMap["result"]
		version := resultMap.(map[string]interface{})["version"]
		client.Version = fmt.Sprintf("TrueNAS-SCALE-%s", version)
		log.Printf("%s is running version '%s'", client.Cfg.ConnectHost, client.Version)
	} else {
		log.Printf("unable to get the version of TrueNAS for '%s'", client.Cfg.ConnectHost)
	}
	return nil
}

func importCertificate(client *TrueNASWebSocket) error {
	log.Printf("importing the %s certificate", client.Cfg.CertName())
	certPem, err := os.ReadFile(client.Cfg.FullChainPath)
	if err != nil {
		return fmt.Errorf("error reading the certificate file: %v", err)
	}
	keyPem, err := os.ReadFile(client.Cfg.PrivateKeyPath)
	if err != nil {
		return fmt.Errorf("error reading the private key file: %v", err)
	}
	if err = client.WSClient.SubscribeToJobs(); err != nil {
		return fmt.Errorf("error subscribing to job notifications: %v", err)
	}

	params := map[string]string{
		"name":        certName,
		"certificate": string(certPem),
		"privatekey":  string(keyPem),
		"create_type": "CERTIFICATE_CREATE_IMPORTED",
	}
	args := []interface{}{params}

	// call the api to create and deploy the certificate
	job, err := client.WSClient.CallWithJob("certificate.create", args, func(progress float64, state string, desc string) {
		if client.Cfg.Debug {
			log.Printf("job progress: %.2f%%, state: %s, description: %s", progress, state, desc)
		}
	})
	if err != nil {
		return fmt.Errorf("failed to create the certificate job,  %v", err)
	}

	if job.ID > 0 {
		log.Printf("started the certificate creation job with ID: %d", job.ID)
	}

	// Monitor the progress of the job.
	for !job.Finished {
		select {
		case progress := <-job.ProgressCh:
			if client.Cfg.Debug {
				log.Printf("job progress: %.2f%%", progress)
			}
		case err := <-job.DoneCh:
			if err != "" {
				return fmt.Errorf("job failed: %v", err)
			} else {
				log.Println("job completed successfully!")
				break
			}
		}
	}

	return nil
}

func restartUI(client *TrueNASWebSocket) error {
	args := []interface{}{}
	_, err := client.WSClient.Call("system.general.ui_restart", client.Cfg.TimeoutSeconds, args)
	if err != nil {
		return fmt.Errorf("failed to restart the  UI: %v", err)
	} else {
		log.Printf("restarted the UI")
	}
	return nil
}
