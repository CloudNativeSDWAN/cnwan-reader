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

package etcd

import (
	"fmt"
	"testing"

	opsr "github.com/CloudNativeSDWAN/cnwan-operator/pkg/servregistry"
	"github.com/CloudNativeSDWAN/cnwan-reader/pkg/openapi"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestSanitizeLocalhost(t *testing.T) {
	a := assert.New(t)

	cases := []struct {
		host   string
		mode   string
		expRes string
		expErr error
	}{
		{
			host:   "http://",
			expErr: fmt.Errorf("whatever"),
		},
		{
			host:   "https://",
			expErr: fmt.Errorf("whatever"),
		},
		{
			host:   "non-localhost",
			expRes: "non-localhost",
		},
		{
			host:   "localhost/whatever",
			expRes: "localhost/whatever",
		},
		{
			host:   "localhost/whatever",
			mode:   "docker",
			expRes: "host.docker.internal/whatever",
		},
		{
			host:   "localhost/whatever",
			mode:   "whatever",
			expRes: "localhost/whatever",
		},
	}

	for i, currCase := range cases {
		res, err := sanitizeLocalhost(currCase.host, currCase.mode)

		if !a.Equal(currCase.expRes, res) {
			a.FailNow(fmt.Sprintf("case %d failed", i))
		}

		if currCase.expErr != nil && !a.Error(err) {
			a.FailNow(fmt.Sprintf("case %d failed", i))
		}
	}
}

func TestParseEndpointsFromFlags(t *testing.T) {
	a := assert.New(t)

	cases := []struct {
		endpoints []string
		expRes    []Endpoint
	}{
		{
			endpoints: []string{},
		},
		{
			endpoints: []string{
				"localhost",
				"localhost:4455",
				"example.com:444:333:2212",
				"https://",
				"http://",
				"localhost:4455",
			},
			expRes: []Endpoint{
				{
					Host: "localhost",
					Port: defaultPort,
				},
				{
					Host: "localhost",
					Port: 4455,
				},
			},
		},
	}

	for i, currCase := range cases {
		res := parseEndpointsFromFlags(currCase.endpoints)

		if !a.Equal(currCase.expRes, res) {
			a.FailNow("case failed", "case", i)
		}
	}
}

func TestParseFlags(t *testing.T) {
	a := assert.New(t)

	cases := []struct {
		cmd    *cobra.Command
		expRes *Options
		expErr error
	}{
		{
			cmd: func() *cobra.Command {
				c := GetEtcdCommand()
				c.SetArgs([]string{"--metadata-keys=whatever,whatever2"})
				c.PreRun = func(*cobra.Command, []string) {}
				c.Run = func(*cobra.Command, []string) {}
				c.Execute()
				return c
			}(),
			expRes: &Options{Endpoints: []Endpoint{{Host: defaultHost, Port: defaultPort}}, Prefix: "/", targetKeys: []string{"whatever"}},
		},
		{
			cmd: func() *cobra.Command {
				c := GetEtcdCommand()
				c.SetArgs([]string{"--metadata-keys=whatever", "--username=whatever"})
				c.PreRun = func(*cobra.Command, []string) {}
				c.Run = func(*cobra.Command, []string) {}
				c.Execute()
				return c
			}(),
			expErr: fmt.Errorf("username set but no password provided"),
		},
		{
			cmd: func() *cobra.Command {
				c := GetEtcdCommand()
				c.SetArgs([]string{"--metadata-keys=whatever", "--password=whatever"})
				c.PreRun = func(*cobra.Command, []string) {}
				c.Run = func(*cobra.Command, []string) {}
				c.Execute()
				return c
			}(),
			expErr: fmt.Errorf("password set but no username provided"),
		},
		{
			cmd: func() *cobra.Command {
				c := GetEtcdCommand()
				c.SetArgs([]string{
					"--metadata-keys=whatever",
					"--username=whatever",
					"--password=whatever",
					"--prefix=/service-registry/",
					"--endpoints=localhost:3344,example.com:5544",
				})
				c.PreRun = func(*cobra.Command, []string) {}
				c.Run = func(*cobra.Command, []string) {}
				c.Execute()
				return c
			}(),
			expRes: &Options{
				Endpoints: []Endpoint{
					{Host: "localhost", Port: 3344},
					{Host: "example.com", Port: 5544},
				},
				Prefix: "/service-registry/",
				Credentials: &Credentials{
					Username: "whatever", Password: "whatever",
				},
				targetKeys: []string{"whatever"},
			},
		},
	}

	failed := func(i int) {
		a.FailNow("case failed", fmt.Sprintf("case %d", i))
	}

	for i, currCase := range cases {
		res, err := parseFlags(currCase.cmd)
		er := currCase.expRes
		if !a.Equal(currCase.expErr, err) {
			failed(i)
		}

		if !a.Equal(er, res) {
			failed(i)
		}
	}
}

func TestParsePrefix(t *testing.T) {
	a := assert.New(t)
	empty := ""
	onlySlash := "/"
	multipleSlashes := "///test///"

	cases := []struct {
		prefix string
		expRes string
	}{
		{
			expRes: "/",
		},
		{
			prefix: empty,
			expRes: "/",
		},
		{
			prefix: onlySlash,
			expRes: "/",
		},
		{
			prefix: multipleSlashes,
			expRes: "/test/",
		},
	}

	for i, currCase := range cases {
		res := parsePrefix(currCase.prefix)
		errRes := a.Equal(currCase.expRes, res)
		if !errRes {
			a.FailNow(fmt.Sprintf("case %d failed", i))
		}
	}
}

func TestValidateEndpoint(t *testing.T) {
	a := assert.New(t)
	anyErr := fmt.Errorf("any")

	cases := []struct {
		arg    []byte
		expRes *opsr.Endpoint
		expErr error
	}{
		{
			expErr: fmt.Errorf("no value provided"),
		},
		{
			arg:    []byte("invalid"),
			expErr: anyErr,
		},
		{
			arg: func() []byte {
				ep := &opsr.Endpoint{
					Name: "ep",
				}
				epval, _ := yaml.Marshal(ep)
				return epval
			}(),
			expErr: opsr.ErrNsNameNotProvided,
		},
		{
			arg: func() []byte {
				ep := &opsr.Endpoint{
					Name:     "ep",
					ServName: "srv",
					NsName:   "ns",
				}
				epval, _ := yaml.Marshal(ep)
				return epval
			}(),
			expErr: fmt.Errorf("endpoint has no address"),
		},
		{
			arg: func() []byte {
				ep := &opsr.Endpoint{
					Name:     "ep",
					ServName: "srv",
					NsName:   "ns",
					Address:  "10.10.10.10",
					Metadata: map[string]string{"protocol": "tcp"},
				}
				epval, _ := yaml.Marshal(ep)
				return epval
			}(),
			expRes: &opsr.Endpoint{
				Name:     "ep",
				ServName: "srv",
				NsName:   "ns",
				Address:  "10.10.10.10",
				Port:     80,
				Metadata: map[string]string{"protocol": "tcp"},
			},
		},
		{
			arg: func() []byte {
				ep := &opsr.Endpoint{
					Name:     "ep",
					ServName: "srv",
					NsName:   "ns",
					Address:  "10.10.10.10",
					Port:     9596,
					Metadata: map[string]string{"protocol": "tcp"},
				}
				epval, _ := yaml.Marshal(ep)
				return epval
			}(),
			expRes: &opsr.Endpoint{
				Name:     "ep",
				ServName: "srv",
				NsName:   "ns",
				Address:  "10.10.10.10",
				Port:     9596,
				Metadata: map[string]string{"protocol": "tcp"},
			},
		},
	}

	fail := func(i int) {
		a.FailNow("case failed", fmt.Sprintf("case %d", i))
	}
	for i, currCase := range cases {
		res, err := validateEndpointFromEtcd(currCase.arg)

		if !a.Equal(currCase.expRes, res) {
			fail(i)
		}

		if currCase.expErr == anyErr {
			if !a.Error(err) {
				fail(i)
			}

		} else {
			if !a.Equal(currCase.expErr, err) {
				fail(i)
			}
		}
	}
}

func TestCreateOpenapiEvent(t *testing.T) {
	a := assert.New(t)

	endp := &opsr.Endpoint{
		Name:     "endp",
		ServName: "srv",
		NsName:   "ns",
		Address:  "10.10.10.10",
		Port:     9696,
		Metadata: map[string]string{
			"test":    "test",
			"another": "another",
		},
	}
	srv := &opsr.Service{
		Name:   "srv",
		NsName: "ns",
		Metadata: map[string]string{
			"name": "srv",
			"srv":  "yes",
		},
	}

	expRes := &openapi.Event{
		Event: "create",
		Service: openapi.Service{
			Name:    endp.Name,
			Address: endp.Address,
			Port:    endp.Port,
			Metadata: []openapi.Metadata{
				{Key: "name", Value: "srv"},
				{Key: "srv", Value: "yes"},
			},
		},
	}

	res := createOpenapiEvent(endp, srv, "create")
	a.Equal(expRes, res)
}

func TestMapContainsKeys(t *testing.T) {
	a := assert.New(t)
	m := map[string]string{
		"key1": "val1",
		"key2": "val2",
		"key3": "val3",
	}
	cases := []struct {
		targets []string
		expRes  bool
	}{
		{
			targets: []string{"key1"},
			expRes:  true,
		},
		{
			targets: []string{"key1", "key2", "key4"},
			expRes:  false,
		},
		{
			targets: []string{"key1", "key2", "key3", "key4"},
			expRes:  false,
		},
	}

	fail := func(i int) {
		a.FailNow("case failed", fmt.Sprintf("case %d", i))
	}
	for i, currCase := range cases {
		res := mapContainsKeys(m, currCase.targets)
		if !a.Equal(currCase.expRes, res) {
			fail(i)
		}
	}
}

func TestMapValuesChanged(t *testing.T) {
	a := assert.New(t)
	prev := map[string]string{
		"key1": "val1",
		"key2": "val2",
		"key3": "val3",
	}
	cases := []struct {
		now     map[string]string
		targets []string
		expRes  bool
	}{
		{
			now:     prev,
			targets: []string{"key1", "key2"},
			expRes:  false,
		},
		{
			now:     map[string]string{"key1": "new"},
			targets: []string{"key1", "key2"},
			expRes:  true,
		},
		{
			now:     map[string]string{"key5": "val5"},
			targets: []string{"key5"},
			expRes:  false,
		},
		{
			now:     map[string]string{"key5": "val5"},
			targets: []string{"key1"},
			expRes:  false,
		},
	}

	fail := func(i int) {
		a.FailNow("case failed", fmt.Sprintf("case %d", i))
	}
	for i, currCase := range cases {
		res := mapValuesChanged(currCase.now, prev, currCase.targets)
		if !a.Equal(currCase.expRes, res) {
			fail(i)
		}
	}
}
