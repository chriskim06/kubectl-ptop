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
	"context"
	"sort"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/kubectl/pkg/cmd/top"
	"k8s.io/kubectl/pkg/metricsutil"
	metricsapi "k8s.io/metrics/pkg/apis/metrics"
	metricsV1beta1api "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

func GetNodeMetrics(o *top.TopNodeOptions, flags *genericclioptions.ConfigFlags) ([]MetricsValues, error) {
	clientset, metricsClient, err := getClients(flags)
	if err != nil {
		return nil, err
	}
	o.MetricsClient = metricsClient
	o.NodeClient = clientset.CoreV1()
	o.Printer = metricsutil.NewTopCmdPrinter(o.Out)

	versionedMetrics := &metricsV1beta1api.NodeMetricsList{}
	mc := o.MetricsClient.MetricsV1beta1()
	nm := mc.NodeMetricses()

	// handle getting all or with resource name
	versionedMetrics, err = nm.List(context.TODO(), metav1.ListOptions{LabelSelector: labels.Everything().String()})
	if err != nil {
		return nil, err
	}
	metrics := &metricsapi.NodeMetricsList{}
	err = metricsV1beta1api.Convert_v1beta1_NodeMetricsList_To_metrics_NodeMetricsList(versionedMetrics, metrics, nil)
	if err != nil {
		return nil, err
	}

	nodeList, err := o.NodeClient.Nodes().List(context.TODO(), metav1.ListOptions{
		LabelSelector: labels.Everything().String(),
	})
	if err != nil {
		return nil, err
	}
	var nodes []v1.Node
	nodes = append(nodes, nodeList.Items...)
	allocatable := make(map[string]v1.ResourceList)
	for _, n := range nodes {
		allocatable[n.Name] = n.Status.Allocatable
	}

	values := []MetricsValues{}
	for _, m := range metrics.Items {
		cpuQuantity := m.Usage[v1.ResourceCPU]
		cpuAvailable := allocatable[m.Name][v1.ResourceCPU]
		cpuFraction := float64(cpuQuantity.MilliValue()) / float64(cpuAvailable.MilliValue()) * 100
		memQuantity := m.Usage[v1.ResourceMemory]
		memAvailable := allocatable[m.Name][v1.ResourceMemory]
		memFraction := float64(memQuantity.MilliValue()) / float64(memAvailable.MilliValue()) * 100
		values = append(values, MetricsValues{
			Name:       m.Name,
			CPUPercent: cpuFraction,
			MemPercent: memFraction,
			CPUCores:   int(cpuQuantity.MilliValue()),
			MemCores:   int(memQuantity.Value()),
		})
	}

	// Sort the metrics results somehow
	sort.Slice(values, func(i, j int) bool {
		return values[i].Name < values[j].Name
	})

	return values, nil
}