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
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/signal"

	"github.com/CloudNativeSDWAN/cnwan-reader/pkg/openapi"
	"github.com/CloudNativeSDWAN/cnwan-reader/pkg/queue"
	"github.com/CloudNativeSDWAN/cnwan-reader/pkg/services"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	corev1 "k8s.io/api/core/v1"

	ktypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type k8sOptions struct {
	kubeconfigPath  string
	adaptorEndpoint string
	annotationKeys  []string
}

type k8sWatcher struct {
	opts      k8sOptions
	sendQueue queue.Queue
	store     map[string]*corev1.Service
}

func (k *k8sWatcher) main() error {
	// --------------------------------------
	// Set ups
	// --------------------------------------

	config, err := clientcmd.BuildConfigFromFlags("", k.opts.kubeconfigPath)
	if err != nil {
		log.Err(err).Msg("error while connecting to kubernetes cluster: exiting...")
		return err
	}

	cli, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Err(err).Msg("error while connecting to kubernetes cluster: exiting...")
		return err
	}

	ctx, canc := context.WithCancel(context.Background())
	exitChan := make(chan struct{})
	w, err := cli.CoreV1().Services("").Watch(ctx, metav1.ListOptions{})
	if err != nil {
		canc()
		log.Err(err).Msg("error while watching for services on the kubernetes cluster: exiting...")
		return err
	}

	// --------------------------------------
	// Set the helpers
	// --------------------------------------

	servsHandler, err := services.NewHandler(ctx, k.opts.adaptorEndpoint)
	if err != nil {
		canc()
		return err
	}
	k.sendQueue = queue.New(ctx, servsHandler)

	// --------------------------------------
	// Watch
	// --------------------------------------

	go func() {
		k.watch(ctx, w)
		close(exitChan)
	}()

	// --------------------------------------
	// Graceful shutdown
	// --------------------------------------

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	<-sig
	fmt.Println()
	log.Info().Msg("exit requested")

	canc()
	w.Stop()
	<-exitChan

	log.Info().Msg("good bye!")
	return nil
}

func (k *k8sWatcher) watch(ctx context.Context, w watch.Interface) {
	wchan := w.ResultChan()

	for {
		var ev watch.Event

		select {
		case data := <-wchan:
			ev = data
		case <-ctx.Done():
			return
		}

		if serv, success := ev.Object.(*corev1.Service); success {
			k.processNextService(serv, ev.Type)
		}
	}
}

func (k *k8sWatcher) processNextService(serv *corev1.Service, evtype watch.EventType) {
	namespacedName := ktypes.NamespacedName{Namespace: serv.Namespace, Name: serv.Name}
	l := log.With().Str("service", namespacedName.String()).Logger()

	// --------------------------------------
	// Parse event
	// --------------------------------------

	// TODO: on future versions this will be changed with the new openapi
	oaevents := map[string]*openapi.Event{}
	switch evtype {

	case watch.Deleted:
		if serv, exists := k.store[namespacedName.String()]; exists {
			oaservs, _ := getDataFromK8sService(serv, k.opts.annotationKeys)

			for _, serv := range oaservs {
				oaevents[serv.Name] = &openapi.Event{
					Event:   "delete",
					Service: *serv,
				}
			}

			delete(k.store, namespacedName.String())
		}

	case watch.Added:
		oaservs, err := getDataFromK8sService(serv, k.opts.annotationKeys)
		if err == nil {
			for _, serv := range oaservs {
				oaevents[serv.Name] = &openapi.Event{
					Event:   "create",
					Service: *serv,
				}
			}

			k.store[namespacedName.String()] = serv
		}

	case watch.Modified:
		prev, prevExists := k.store[namespacedName.String()]
		prevServs := []*openapi.Service{}
		if prev != nil {
			prevServs, _ = getDataFromK8sService(prev, k.opts.annotationKeys)
		}

		currServs, currErr := getDataFromK8sService(serv, k.opts.annotationKeys)
		if currErr != nil {
			if prevExists {
				for _, serv := range prevServs {
					oaevents[serv.Name] = &openapi.Event{
						Event:   "delete",
						Service: *serv,
					}
				}
				delete(k.store, namespacedName.String())
			}
		} else {
			if prevExists {
				_oaevents := getServChanges(prevServs, currServs)
				for _, oae := range _oaevents {
					oaevents[oae.Service.Name] = &openapi.Event{
						Event:   "update",
						Service: oae.Service,
					}
				}
			} else {
				for _, serv := range currServs {
					oaevents[serv.Name] = &openapi.Event{
						Event:   "create",
						Service: *serv,
					}
				}
			}

			k.store[namespacedName.String()] = serv
		}
	}

	// --------------------------------------
	// Send events
	// --------------------------------------

	if len(oaevents) > 0 {
		go k.sendQueue.Enqueue(oaevents)
		l.Info().Int("events", len(oaevents)).Msg("sending events...")
	}
}

// TODO: this needs change with next version of openapi
// TODO: write this on a new package of the operator (since it uses the same code)?
func getDataFromK8sService(serv *corev1.Service, annKeys []string) ([]*openapi.Service, error) {
	// -- Is this a LoadBalancer
	if serv.Spec.Type != corev1.ServiceTypeLoadBalancer {
		return nil, fmt.Errorf("service is not of type LoadBalancer")
	}

	// -- Has an IP?
	if len(serv.Status.LoadBalancer.Ingress) == 0 {
		return nil, fmt.Errorf("service has no LoadBalancer IPs")
	}

	// -- Get the annotations
	filteredAnnotations := map[string]string{}
	filteredMet := []openapi.Metadata{}
	for _, key := range annKeys {
		val, exists := serv.Annotations[key]
		if !exists {
			return nil, fmt.Errorf("service does not have required annotations")
		}

		filteredAnnotations[key] = val
		filteredMet = append(filteredMet, openapi.Metadata{Key: key, Value: val})
	}

	// -- Prepare the services
	services := []*openapi.Service{}
	for _, ingr := range serv.Status.LoadBalancer.Ingress {
		for _, port := range serv.Spec.Ports {
			h := sha256.New()
			h.Write([]byte(fmt.Sprintf("%s:%d", ingr.IP, port.Port)))
			name := fmt.Sprintf("%s/%s-%s", serv.Namespace, serv.Name, hex.EncodeToString(h.Sum(nil))[:10])

			services = append(services, &openapi.Service{
				Name:     name,
				Address:  ingr.IP,
				Port:     port.Port,
				Metadata: filteredMet,
			})
		}
	}

	return services, nil
}

// TODO: this needs change with next version of openapi
func getServChanges(old, curr []*openapi.Service) []*openapi.Event {
	currcopy := make([]*openapi.Service, len(curr))
	copy(currcopy, curr)

	evs := []*openapi.Event{}
	for _, o := range old {
		found := -1

		for i, c := range currcopy {
			if o.Name == c.Name {
				if c.Address != o.Address || c.Port != o.Port || !areMetadataEqual(o.Metadata, c.Metadata) {
					evs = append(evs, &openapi.Event{
						Event:   "update",
						Service: *c,
					})
				}

				found = i
				break
			}
		}

		if found > -1 {
			currcopy[found] = currcopy[len(currcopy)-1]
			currcopy = currcopy[0 : len(currcopy)-1]
		} else {
			evs = append(evs, &openapi.Event{
				Event:   "delete",
				Service: *o,
			})
		}
	}

	for _, c := range currcopy {
		evs = append(evs, &openapi.Event{
			Event:   "create",
			Service: *c,
		})
	}

	return evs
}

func areMetadataEqual(old, curr []openapi.Metadata) bool {
	if len(old) != len(curr) {
		return false
	}

	for _, o := range old {
		found := false
		for _, c := range curr {
			if c.Key == o.Key {
				if c.Value != o.Value {
					return false
				}

				found = true
				break
			}
		}

		if !found {
			return false
		}
	}

	return true
}
