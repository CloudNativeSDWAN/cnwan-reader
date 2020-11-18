// Copyright © 2020 Cisco
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// All rights reserved.

package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/CloudNativeSDWAN/cnwan-reader/pkg/openapi"
	"github.com/rs/zerolog/log"
)

// Handler is in charge of handling services, i.e. sending them to endpoints
// specified by CN-WAN Reader OpenAPI's specification.
type Handler interface {
	// Send these events to an external handler.
	Send(services []openapi.Event) error
}

type servicesHandler struct {
	mainCtx context.Context
	client  *openapi.APIClient
}

// NewHandler returns a services handler that uses the endpoints defined in
// the openAPI specification to send service events.
func NewHandler(ctx context.Context, endpoint string) (Handler, error) {
	if len(endpoint) == 0 {
		return nil, errors.New("endpoint is empty")
	}

	// Get the client
	cfg := openapi.NewConfiguration()
	apiClient := openapi.NewAPIClient(cfg)
	if endpoint != "localhost/cnwan" {
		apiClient.ChangeBasePath(strings.Replace(cfg.BasePath, "localhost/cnwan", endpoint, 1))
	}

	return &servicesHandler{
		client:  apiClient,
		mainCtx: ctx,
	}, nil
}

// Send these events to an external handler.
func (s *servicesHandler) Send(events []openapi.Event) error {
	l := log.With().Str("func", "services.servicesHandler.Send").Logger()
	timeOut := time.Duration(20 * time.Second)
	ctx, canc := context.WithTimeout(s.mainCtx, timeOut)
	defer canc()

	l.Debug().Msg("sending events....")
	resp, httpResp, err := s.client.EventsApi.SendEvents(ctx, events)
	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("%v seconds timeout expired", timeOut.Seconds())
	}

	if httpResp == nil {
		if err != nil {
			l.Err(err).Msg("error while getting response")
			return err
		}

		l.Info().Msg("no response to parse")
		return err
	}

	if httpResp.StatusCode >= 500 {
		newErr := err.(openapi.GenericOpenAPIError)
		if newErr.Model() != nil {
			resp = newErr.Model().(openapi.Response)
		} else {
			body := newErr.Body()
			if len(body) > 0 {
				if err := json.Unmarshal(body, &resp); err != nil {
					l.Warn().AnErr("error", err).Msg("error while trying to parse the response")
					resp.Title = "BODY"
					resp.Description = string(body)
				}
			}
		}
	}

	s.logResponseError(resp, httpResp.StatusCode)

	return err
}

func (s *servicesHandler) logResponseError(resp openapi.Response, statusCode int) {
	l := log.With().Str("func", "services.servicesHandler.logResponseError").Logger()

	responseMsg := "<>"
	if len(resp.Title) > 0 && len(resp.Description) > 0 {
		responseMsg = fmt.Sprintf("%d - %s: %s", statusCode, resp.Title, resp.Description)
	}

	l = l.With().Int("status-code", statusCode).Logger()
	if statusCode == 207 {
		l.Info().Str("response", responseMsg).Msg("received response from the adaptor")

		if len(resp.Errors) == 0 {
			// For 207 we actually *need* the response body, because we need
			// to provide feedback on what went wrong. So if no body is there,
			// we consider this another error.
			l.Err(fmt.Errorf("%s", "returned response is 207 but no content is returned")).Msg("returned response is invalid")
		}

		// Log the errors in the response
		for _, evErr := range resp.Errors {
			e := fmt.Errorf("Resource '%s': %d %s  %s", evErr.Resource, evErr.Status, evErr.Title, evErr.Description)
			l.Warn().AnErr("error", e).Msg("adaptor error occurred on resource")
		}

		return
	}

	switch {
	case statusCode >= 200 && statusCode < 300:
		l.Info().Str("response", responseMsg).Msg("received response from the adaptor")
	case statusCode >= 300 && statusCode < 400:
		l.Warn().Str("response", responseMsg).Msg("received response from the adaptor")
	case statusCode >= 400 && statusCode < 600:
		l.Error().AnErr("error", fmt.Errorf(responseMsg)).Msg("received response from the adaptor")
	}
}
