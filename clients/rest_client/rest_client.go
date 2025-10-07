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

package rest_client

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"tnascert-deploy/clients"
)

const Endpoint = "api/v2.0"

type TrueNAS_DDP_Client struct {
	url       string
	verifySSL bool
	client    *http.Client
	isClosed  bool
	callID    int
}

type AuthRoundTripper struct {
	Transport http.RoundTripper
	AuthToken string
}

func (art *AuthRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone the request to avoid modifying the original request object
	newReq := req.Clone(req.Context())

	// Add the Authorization header
	newReq.Header.Set("Authorization", art.AuthToken)

	newReq.Header.Set("Content-Type", "application/json")

	// Delegate the actual request execution to the underlying transport
	return art.Transport.RoundTrip(newReq)
}

func NewClient(serverURL string, verifySSL bool) (clients.Client, error) {
	if serverURL == "" {
		return nil, fmt.Errorf("the serverURL is empty")
	}

	client := TrueNAS_DDP_Client{
		url:       serverURL,
		verifySSL: verifySSL,
	}

	return &client, nil
}

func (c *TrueNAS_DDP_Client) Call(method string, timeout int64, params interface{}) (json.RawMessage, error) {
	return nil, nil
}

func (c *TrueNAS_DDP_Client) CallWithJob(method string, params interface{},
	callback func(progress float64, state string, desc string)) (*clients.Job, error) {

	// create a certificate from an import
	if method == "certificate.create" {
		var cParams map[string]string

		if obj, ok := params.([]interface{}); ok {
			if myMap, ok := obj[0].(map[string]string); ok {
				cParams = myMap
			}
		}
		var URLEndPoint = "/certificate"
		job := &clients.Job{
			ID:       0,
			Method:   "certificate.create",
			State:    "PENDING",
			Progress: 100.0,
			Finished: true,
		}
		jsonData, err := json.Marshal(cParams)
		if err != nil {
			return nil, fmt.Errorf("error marshalling JSON: %v", err)
		}
		req, err := http.NewRequest("POST", c.url+URLEndPoint, bytes.NewBuffer(jsonData))
		if err != nil {
			return nil, fmt.Errorf("error creating request: %v", err)
		}
		resp, err := c.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("error sending request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("certificate import failed, response: %d", resp.Status)
		} else {
			log.Printf("certificate named %s has been imported", cParams["name"])
		}
		return job, nil
	}

	return nil, nil
}

func (c *TrueNAS_DDP_Client) Close() error {
	return nil
}

func (c *TrueNAS_DDP_Client) Login(username string, password string, apiKey string) error {
	var authToken string
	var URLEndPoint = "/core/ping"

	if apiKey != "" {
		authToken = "Bearer " + apiKey
	} else if username != "" && password != "" {
		plainText := username + ":" + password
		authToken = "Basic " + base64.StdEncoding.EncodeToString([]byte(plainText))
	} else {
		return fmt.Errorf("no valid credentials have been supplied")
	}

	authTransport := &AuthRoundTripper{
		Transport: http.DefaultTransport,
		AuthToken: authToken,
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return fmt.Errorf("creating the cookie jar failed %v", err)
	}

	c.client = &http.Client{
		Transport: authTransport,
		Jar:       jar,
	}

	r, err := http.NewRequest("GET", c.url+URLEndPoint, nil)
	res, err := c.client.Do(r)
	if err != nil {
		return fmt.Errorf("login error %v", err)
	}
	defer res.Body.Close()
	return nil
}

func (c *TrueNAS_DDP_Client) SubscribeToJobs() error {
	return nil
}
