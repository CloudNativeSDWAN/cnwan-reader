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
	"testing"
	"time"

	assert "github.com/stretchr/testify/assert"
)

type fakeData struct {
	count int
}

func (f *fakeData) call() {
	f.count++
}

func TestPoll(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	d := &fakeData{}
	p := New(ctx, 2)
	p.SetPollFunction(d.call)
	p.Start()

	time.Sleep(5 * time.Second)
	cancel()

	// At this point, the registered function should havbe been executed
	// 3 times: once at Start(), and twice during these 5 seconds
	if d.count != 3 {
		assert.Fail(t, "polled more than twice or 3 times")
	}
}
