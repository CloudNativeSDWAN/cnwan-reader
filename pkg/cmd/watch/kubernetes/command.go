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
LoadBalancer services that have the required annotation keys defined by
the --annotation-keys flag and create events that will be later sent to the
CN-WAN Adaptor for processing.

In order to work, a valid kubeconfig file is needed and its path must be
provided with the --kubeconfig flag. If empty, the default one will be
used, which is usually ~/.kube/config on Unix-based systems.

Additionally, a context can be used via --context: if you want the program
to monitor a specific Kubernetes cluster you can use this as --context value.
If empty, CN-WAN Reader will use the same context that kubectl is using.

Make sure the context you are using has permissions to read, watch and list
services on the Kubernetes cluster you chose, which involves creating
appropriate ClusterRole and ClusterRoleBinding resources. Please check CN-WAN
Reader's documentation and Kubernetes documentation to learn more.`,
		Example: "cnwan-reader watch kubernetes --kubeconfig /path/to/another/kubeconfig --context admin@my-gke-cluster --annotation-keys my-key",
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
	cmd.Flags().StringVar(&opts.currentContext, "context", "", "the context to use. If empty, the default one in kubeconfig will be used")

	return cmd
}
