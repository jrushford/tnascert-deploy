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
 * Adapter to decouple and wrap the TrueNAS go API to conform to
 * clients.Client interface.
 */

package rpc_client

import (
	"encoding/json"
	"github.com/truenas/api_client_golang/truenas_api"
	"tnascert-deploy/clients"
)

const Endpoint = "api/current"

type TrueNAS_RPC_Client struct {
	apiClient *truenas_api.Client
}

func NewClient(serverURL string, verifySSL bool) (clients.Client, error) {
	cl, err := truenas_api.NewClient(serverURL, verifySSL)
	if err != nil {
		return nil, err
	}

	client := &TrueNAS_RPC_Client{
		apiClient: cl,
	}

	return client, nil
}

func (c *TrueNAS_RPC_Client) Login(user string, pass string, apiKey string) error {
	return c.apiClient.Login(user, pass, apiKey)
}

func (c *TrueNAS_RPC_Client) Call(method string, timeout int64, params interface{}) (json.RawMessage, error) {
	return c.apiClient.Call(method, timeout, params)
}

func (c *TrueNAS_RPC_Client) CallWithJob(method string, params interface{}, callback func(progress float64, state string, desc string)) (*clients.Job, error) {
	apiJob, err := c.apiClient.CallWithJob(method, params, callback)

	if err != nil {
		return nil, err
	}

	return (*clients.Job)(apiJob), nil
}

func (c *TrueNAS_RPC_Client) Close() error {
	return c.apiClient.Close()
}

func (c *TrueNAS_RPC_Client) SubscribeToJobs() error {
	return c.apiClient.SubscribeToJobs()
}
