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
 *  This package is strictly for unit testing.  It is a mock client used to test the
 *  deploy package
 */

package mock

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/truenas/api_client_golang/truenas_api"
	"tnascert-deploy/config"
)

const defaultCertName = "tnas-cert-deploy-2025-01-01-0101683628"

// mock client for tests
type Client struct {
	url           string // WebSocket server URL
	tlsSkipVerify bool   // WebSocket connection instance
}

func NewClient(serverURL string, TlsSkipVerify bool) (*Client, error) {
	client := &Client{url: serverURL,
		tlsSkipVerify: TlsSkipVerify}
	return client, nil
}

func (c *Client) Call(method string, timeout int64, params interface{}) (json.RawMessage, error) {
	if method == "certificate.create" {
		data := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      1,
			"result":  100,
		}
		res, err := json.Marshal(data)
		var resp json.RawMessage
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
			{"id": 3, "name": defaultCertName},
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
	}

	return nil, nil
}

func (c *Client) CallWithJob(method string, params interface{}, callback func(progress float64, state string, desc string)) (*truenas_api.Job, error) {
	var job truenas_api.Job
	job.ID = 1

	return &job, nil
}

func (c *Client) Close() error {
	return nil
}

func (c *Client) Login(username string, password string, apiKey string) error {
	if apiKey == "test" {
		return nil
	}
	return errors.New("mock.Client Login: invalid api key")
}

func GetCertName(cfg *config.Config) string {
	return defaultCertName
}

func (c *Client) SubscribeToJobs() error {
	return nil
}
