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

package clients

import (
	"encoding/json"
)

// Job represents a long-running job in TrueNAS.
type Job struct {
	ID         int64                                             // Job ID
	Method     string                                            // Method associated with the job
	State      string                                            // Current state of the job (e.g., "PENDING", "SUCCESS")
	Result     interface{}                                       // Result of the job once it finishes
	Progress   float64                                           // Progress of the job (0.0 to 100.0)
	Finished   bool                                              // Indicates if the job is finished
	ProgressCh chan float64                                      // Channel to report progress updates
	DoneCh     chan string                                       // Channel to signal when the job is done
	Callback   func(progress float64, state string, desc string) // Callback function to report progress and state
}

// Client interface
type Client interface {
	Login(username string, password string, apiKey string) error
	Call(method string, timeout int64, params interface{}) (json.RawMessage, error)
	CallWithJob(method string, params interface{}, callback func(progress float64, state string, desc string)) (*Job, error)
	Close() error
	SubscribeToJobs() error
}
