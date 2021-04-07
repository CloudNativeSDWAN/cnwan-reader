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

package kubernetes

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

var (
	log zerolog.Logger
)

func init() {
	output := zerolog.ConsoleWriter{Out: os.Stdout}
	log = zerolog.New(output).With().Timestamp().Logger().Level(zerolog.InfoLevel)
}

type k8sOptions struct {
	kubeconfigPath string
}

// GetK8sCommand returns the kubernetes command.
func GetK8sCommand() *cobra.Command {
	opts := k8sOptions{}

	cmd := &cobra.Command{
		Use:   "kubernetes [OPTIONS]",
		Short: "watch for the state of the services on the kubernetes cluster",
		Long: `Connect to the kubernetes cluster to watch for services.

Once the connection is established successfully, CN-WAN Reader will observe all
LoadBalancer services and create events that will be later sent to the CN-WAN Adaptor processing`,
		Example: "cnwan-reader watch kubernetes --kubeconfig /path/to/another/kubeconfig",
		Run: func(cmd *cobra.Command, args []string) {
			// TODO: implement me
		},
	}

	// Flags
	cmd.Flags().StringVar(&opts.kubeconfigPath, "kubeconfig", "~/.kube/config", "path to the kubeconfig file to use")

	return cmd
}
