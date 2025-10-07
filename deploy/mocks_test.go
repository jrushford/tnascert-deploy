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

/*
 * Unit tests mock client for the deploy package.
 */

package deploy

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"tnascert-deploy/clients"
	"tnascert-deploy/config"
)

// mock client for tests
type DeployClient struct {
	url           string // WebSocket server URL
	tlsSkipVerify bool   // WebSocket connection instance
	cfg           *config.Config
}

func NewClient(serverURL string, TlsSkipVerify bool) (*DeployClient, error) {
	client := &DeployClient{url: serverURL,
		tlsSkipVerify: TlsSkipVerify}
	return client, nil
}

func (c *DeployClient) Call(method string, timeout int64, params interface{}) (json.RawMessage, error) {
	if method == "app.config" {
		var resp json.RawMessage
		data := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      1,
			"result": map[string]interface{}{"ix_certificates": map[string]interface{}{
				"testcert": 100,
			}, "network": map[string]interface{}{}},
		}
		res, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("mock.Call(): Error marshalling response: %v", err)
		} else {
			resp = json.RawMessage(res)
			return resp, nil
		}
	} else if method == "app.query" {
		var resp json.RawMessage
		m := []map[string]interface{}{{"name": "testapp", "id": "testapp"}}
		data := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      1,
			"result":  m,
		}
		res, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("mock.Call(): Error marshalling response: %v", err)
		} else {
			resp = json.RawMessage(res)
			return resp, nil
		}
	} else if method == "certificate.create" {
		var resp json.RawMessage
		data := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      1,
			"result":  100,
		}
		res, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("mock.Call(): Error marshalling response: %v", err)
		} else {
			resp = json.RawMessage(res)
			return resp, nil
		}
	} else if method == "app.certificate_choices" {
		var resp json.RawMessage
		certs := []map[string]interface{}{
			{"id": 1, "name": "truenas_default"},
			{"id": 2, "name": "tnas-cert-deploy-2024-12-31-0801683628"},
			{"id": 3, "name": c.cfg.CertName()},
		}

		var args map[string]interface{} = make(map[string]interface{})
		args = map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      1,
			"result":  certs,
		}
		res, err := json.Marshal(args)
		if err != nil {
			return resp, fmt.Errorf("mock.Call(): Error marshalling response: %v", err)
		} else {
			resp = json.RawMessage(res)
			return resp, nil
		}
	} else if method == "ftp.update" {
		result := map[string]interface{}{
			"testresult": "ok",
		}
		args := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      1,
			"result":  result,
		}
		res, err := json.Marshal(args)
		if err != nil {
			return res, fmt.Errorf("mock.Call(): Error marshalling response: %v", err)
		} else {
			resp := json.RawMessage(res)
			return resp, nil
		}
	}
	return nil, nil
}

func jobRunner(job *clients.Job) {
	time.Sleep(2 * time.Second)
	job.ProgressCh <- 100
	job.DoneCh <- ""
	job.Finished = true
	close(job.DoneCh)
	close(job.ProgressCh)
}

func (c *DeployClient) CallWithJob(method string, params interface{}, callback func(progress float64, state string, desc string)) (*clients.Job, error) {
	var job clients.Job
	if method == "app.update" {
		job = clients.Job{
			ID:         100,
			Method:     "app.update",
			State:      "PENDING",
			ProgressCh: make(chan float64),
			DoneCh:     make(chan string),
		}
	} else if method == "certificate.create" {
		job = clients.Job{
			ID:         101,
			Method:     "certificate.create",
			State:      "PENDING",
			ProgressCh: make(chan float64),
			DoneCh:     make(chan string),
		}
	} else if method == "certificate.delete" {
		job = clients.Job{
			ID:         101,
			Method:     "certificate.create",
			State:      "PENDING",
			ProgressCh: make(chan float64),
			DoneCh:     make(chan string),
		}
	}

	go jobRunner(&job)

	return &job, nil
}

func (c *DeployClient) Close() error {
	return nil
}

func (c *DeployClient) Login(username string, password string, apiKey string) error {
	// apikey is preferred
	if apiKey != "" {
		if apiKey == "test" {
			return nil
		} else {
			return errors.New("mock.Client Login: invalid api key")
		}
	} else if username != "" && password != "" {
		if username == "admin" && password == "admin" {
			return nil
		} else {
			return errors.New("mock.Client Login: invalid username or password")
		}
	}
	return errors.New("mock.Client Login: failed")
}

func (c *DeployClient) SetConfig(cfg *config.Config) {
	c.cfg = cfg
}

func (c *DeployClient) SubscribeToJobs() error {
	return nil
}
