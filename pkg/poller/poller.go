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

package poller

import (
	"context"
	"errors"
	"time"

	"github.com/rs/zerolog/log"
)

type fn func()

// Poller periodically executes a given function
type Poller interface {
	// Start the poller
	Start() error
	// SetPollFunction sets the function that must be called
	SetPollFunction(fn)
}

type funcPoller struct {
	interval time.Duration
	mainCtx  context.Context
	pollFunc fn
}

// New returns a new instance of a poller
func New(ctx context.Context, interval int) Poller {
	return &funcPoller{
		interval: time.Duration(interval) * time.Second,
		mainCtx:  ctx,
	}
}

func (p *funcPoller) SetPollFunction(function fn) {
	p.pollFunc = function
}

// Start starts the poller
func (p *funcPoller) Start() error {
	if p.pollFunc == nil {
		return errors.New("poll function is not set")
	}

	p.pollFunc()

	// Now poll on a timer
	go p.poll()

	return nil
}

func (p *funcPoller) poll() {
	l := log.With().Str("func", "poller.funcPoller.poll").Logger()
	ticker := time.NewTicker(p.interval)

	for {
		// Which one happens first?
		select {
		case <-ticker.C:
			go p.pollFunc()
		case <-p.mainCtx.Done():
			l.Info().Msg("stop requested")
			ticker.Stop()
			return
		}
	}
}
