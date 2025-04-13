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
	"strings"
	"time"
	"tnascert-deploy/config"
	"truenas_api/truenas_api"
)

const endpoint = "api/current"

var certs_list map[string]int64 = map[string]int64{}

// login with an API key
func clientLogin(client *truenas_api.Client, cfg *config.Config) error {
	username, password := "", ""
	apikey := cfg.Api_key
	return client.Login(username, password, apikey)
}

// deploy certificate
func deployCertificate(client *truenas_api.Client, cert_name string, cfg *config.Config) error {
	log.Println("Deploying certificate", cert_name)

	// read in the certificate data
	cert_pem, err := os.ReadFile(cfg.Fullchain_path)
	if err != nil {
		return fmt.Errorf("could not load the certificate, %v", err)
	}
	// read in the private key data
	key_pem, err := os.ReadFile(cfg.Private_key_path)
	if err != nil {
		return fmt.Errorf("could not load the private key, %v", err)
	}

	params := map[string]string{"name": cert_name, "certificate": string(cert_pem),
		"privatekey": string(key_pem), "create_type": "CERTIFICATE_CREATE_IMPORTED"}
	args := []map[string]string{params}

	// call the api to create and deploy the certificate
	job, err := client.Call("certificate.create", 10, args)
	if err != nil {
		return err
	} else {
		resp_map := make(map[string]interface{})
		err = json.Unmarshal(job, &resp_map)
		if err != nil {
			return err
		}
		log.Printf("Job created id: %v", resp_map["result"])
	}

	var inlist bool = false
	var count = 0

	// poll and save all deployed certificates matching our Cert_basename
	// including the newly created certificate
	for !inlist && count != 10 {
		count++
		args2 := []string{}
		resp, err := client.Call("app.certificate_choices", 10, args2)
		if err != nil {
			return err
		}
		resp_map := make(map[string]interface{})
		err = json.Unmarshal(resp, &resp_map)
		if err != nil {
			return err
		}
		list := resp_map["result"].([]interface{})
		for _, v := range list {
			var m = v.(map[string]interface{})
			_, ok := certs_list[m["name"].(string)]
			// add certificate to the cert_list if not already there
			// and skipping those that do not match the cert_basename
			if !ok {
				nm := m["name"].(string)
				idvalue := m["id"].(float64)
				id := int64(idvalue)
				// only add certs that match the Cert_basename to the list
				if strings.HasPrefix(nm, cfg.Cert_basename) {
					certs_list[nm] = id
					log.Printf("certificate name: %v, is: %d", m["name"], id)
				}
			}
			if id, ok := certs_list[cert_name]; ok == true {
				log.Printf("found new certificate: %v, id: %d", m["name"], id)
				inlist = true
			}
		}
		time.Sleep(1 * time.Second)
	}

	if !inlist {
		return fmt.Errorf("search timeout, certificate %s was not deployed", cert_name)
	} else {
		log.Printf("certificate %s deployed successfully", cert_name)
	}
	return nil
}

func InstallCertificate(cfg *config.Config) error {
	var serverURL = fmt.Sprintf("%s://%s:%d/%s", cfg.Protocol, cfg.Connect_host, cfg.Port, endpoint)
	var cert_name = cfg.Cert_basename + strftime.Format("-%Y-%m-%d-%s", time.Now())

	// connect to the truenas api endpoint
	client, err := truenas_api.NewClient(serverURL, false)
	if err != nil {
		return err
	}
	// login
	err = clientLogin(client, cfg)
	if err != nil {
		return err
	} else {
		log.Println("Successfully logged in")
	}
	// deploy the certificate
	err = deployCertificate(client, cert_name, cfg)
	if err != nil {
		return err
	}

	if cfg.Add_as_ui_certificate {
		pmap := make(map[string]int64)
		pmap["ui_certificate"] = certs_list[cert_name]
		args := []map[string]int64{pmap}
		_, err := client.Call("system.general.update", 10, args)
		if err != nil {
			return fmt.Errorf("system.general.update of ui_certificate, %v", err)
		}
		arg := []map[string]interface{}{}
		_, err = client.Call("system.general.ui_restart", 10, arg)
	}

	if cfg.Delete_old_certs {
		for k, v := range certs_list {
			if strings.Compare(k, cert_name) == 0 {
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
