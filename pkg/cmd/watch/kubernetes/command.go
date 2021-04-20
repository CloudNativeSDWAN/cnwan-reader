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
	"fmt"
	"os"
	"path/filepath"

	corev1 "k8s.io/api/core/v1"

	"github.com/CloudNativeSDWAN/cnwan-reader/pkg/internal/utils"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"k8s.io/client-go/util/homedir"
)

var (
	log zerolog.Logger
)

func init() {
	output := zerolog.ConsoleWriter{Out: os.Stdout}
	log = zerolog.New(output).With().Timestamp().Logger().Level(zerolog.InfoLevel)
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
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(opts.annotationKeys) == 0 {
				cmd.Help()
				return fmt.Errorf("no annotation keys provided")
			}

			adaptorEndpoint, err := utils.GetAdaptorEndpointFromFlags(cmd)
			if err != nil {
				cmd.Help()
				return err
			}
			opts.adaptorEndpoint = adaptorEndpoint

			k8s := &k8sWatcher{opts: opts, store: map[string]*corev1.Service{}}
			return k8s.main()
		},
	}

	// Flags
	cmd.Flags().StringVar(&opts.kubeconfigPath, "kubeconfig", func() string {
		if home := homedir.HomeDir(); len(home) > 0 {
			return filepath.Join(home, ".kube", "config")
		}

		return ""
	}(), "path to the kubeconfig file to use")
	cmd.Flags().StringSliceVar(&opts.annotationKeys, "annotation-keys", []string{}, "the annotations keys to look for")
	cmd.Flags().StringSliceVar(&opts.annotationKeys, "metadata-keys", []string{}, "alias for --annotation-keys")

	return cmd
}
