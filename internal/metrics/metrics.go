/*
Copyright © 2020 Chris Kim

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package metrics

import (
	"log"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	metricsclientset "k8s.io/metrics/pkg/client/clientset/versioned"
)

type Resource string

const (
	POD  Resource = "PODS"
	NODE Resource = "NODES"
)

// MetricsValues is an object containing the cpu/memory resources for
// a pod/node that is used to populate termui widgets
type MetricsValues struct {
	Name       string
	CPUPercent float64
	MemPercent float64
	CPUCores   resource.Quantity
	MemCores   resource.Quantity
	CPULimit   resource.Quantity
	MemLimit   resource.Quantity

	Namespace string
	Node      string
	Status    string
	Age       string
	Restarts  int
	Ready     int
	Total     int
}

type MetricsClient struct {
	k     *kubernetes.Clientset
	m     *metricsclientset.Clientset
	flags *genericclioptions.ConfigFlags

	showManagedFields bool
}

func New(flags *genericclioptions.ConfigFlags, showManagedFields bool) MetricsClient {
	k, m, err := getClients(flags)
	if err != nil {
		log.Fatal(err)
	}
	return MetricsClient{
		k:     k,
		m:     m,
		flags: flags,

		showManagedFields: showManagedFields,
	}
}

func getClients(flags *genericclioptions.ConfigFlags) (*kubernetes.Clientset, *metricsclientset.Clientset, error) {
	clientSet, metricsClient, err := clientSets(flags)
	return clientSet, metricsClient, err
}

func getNamespace(flags *genericclioptions.ConfigFlags) (string, error) {
	matchVersionKubeConfigFlags := cmdutil.NewMatchVersionFlags(flags)
	f := cmdutil.NewFactory(matchVersionKubeConfigFlags)

	namespace, _, err := f.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return "", err
	}
	return namespace, err
}

func clientSets(flags *genericclioptions.ConfigFlags) (*kubernetes.Clientset, *metricsclientset.Clientset, error) {
	matchVersionKubeConfigFlags := cmdutil.NewMatchVersionFlags(flags)
	f := cmdutil.NewFactory(matchVersionKubeConfigFlags)

	var err error
	config, err := f.ToRESTConfig()
	flags.ToRESTConfig()
	if err != nil {
		return nil, nil, err
	}
	clientSet, err := f.KubernetesClientSet()
	if err != nil {
		return nil, nil, err
	}
	metricsClient, err := metricsclientset.NewForConfig(config)
	return clientSet, metricsClient, err
}
