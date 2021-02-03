// Copyright © 2021 Cisco
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

package cloudmap

import (
	"fmt"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestParseFlags(t *testing.T) {
	a := assert.New(t)

	cases := []struct {
		cmd *cobra.Command

		expRes *Options
		expErr error
	}{
		{
			cmd: func() *cobra.Command {
				c := GetCloudMapCommand()
				c.PreRun = func(*cobra.Command, []string) {}
				c.Run = func(*cobra.Command, []string) {}
				c.Execute()
				return c
			}(),
			expErr: fmt.Errorf("region not provided"),
		},
		{
			cmd: func() *cobra.Command {
				c := GetCloudMapCommand()
				c.SetArgs([]string{"--region=whatever"})
				c.PreRun = func(*cobra.Command, []string) {}
				c.Run = func(*cobra.Command, []string) {}
				c.Execute()
				return c
			}(),
			expRes: &Options{
				PollInterval: 5,
				Region:       "whatever",
			},
		},
		{
			cmd: func() *cobra.Command {
				c := GetCloudMapCommand()
				c.SetArgs([]string{"--region=whatever", "--credentials-path=/path/to/file"})
				c.PreRun = func(*cobra.Command, []string) {}
				c.Run = func(*cobra.Command, []string) {}
				c.Execute()
				return c
			}(),
			expRes: &Options{
				Region:          "whatever",
				PollInterval:    5,
				CredentialsPath: "/path/to/file",
			},
		},
		{
			cmd: func() *cobra.Command {
				c := GetCloudMapCommand()
				c.Flags().Int("poll-interval", 5, "")
				c.SetArgs([]string{"--region=whatever", "--credentials-path=/path/to/file", "--poll-interval=55"})
				c.PreRun = func(*cobra.Command, []string) {}
				c.Run = func(*cobra.Command, []string) {}
				c.Execute()
				return c
			}(),
			expRes: &Options{
				Region:          "whatever",
				PollInterval:    55,
				CredentialsPath: "/path/to/file",
			},
		},
	}

	failed := func(i int) {
		a.FailNow("case failed", fmt.Sprintf("case %d", i))
	}
	for i, currCase := range cases {
		res, err := parseFlags(currCase.cmd)
		if !a.Equal(currCase.expRes, res) || !a.Equal(currCase.expErr, err) {
			failed(i)
		}
	}
}
