/*
 * This file is part of the KubeVirt project
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Copyright The KubeVirt Authors.
 *
 */

package libpod

import (
	"context"
	"fmt"
	"sort"

	v1 "kubevirt.io/api/core/v1"

	"kubevirt.io/kubevirt/pkg/controller"
	"kubevirt.io/kubevirt/tests/framework/kubevirt"

	k8sv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetRunningPodByLabel(label, labelType, namespace, node string) (*k8sv1.Pod, error) {
	virtCli := kubevirt.Client()
	labelSelector := fmt.Sprintf("%s=%s", labelType, label)
	var fieldSelector string
	if node != "" {
		fieldSelector = fmt.Sprintf("status.phase==%s,spec.nodeName==%s", k8sv1.PodRunning, node)
	} else {
		fieldSelector = fmt.Sprintf("status.phase==%s", k8sv1.PodRunning)
	}
	pods, err := virtCli.CoreV1().Pods(namespace).List(context.Background(),
		metav1.ListOptions{LabelSelector: labelSelector, FieldSelector: fieldSelector},
	)
	if err != nil {
		return nil, err
	}

	switch len(pods.Items) {
	case 0:
		return nil, fmt.Errorf("failed to find pod with the label %s", label)
	default:
		// There can be more than one running pod in case of migration
		// therefore 	return the latest running pod
		sort.Sort(sort.Reverse(PodsByCreationTimestamp(pods.Items)))
		return &pods.Items[0], nil
	}
}

func GetPodByVirtualMachineInstance(vmi *v1.VirtualMachineInstance, namespace string) (*k8sv1.Pod, error) {
	virtCli := kubevirt.Client()

	pods, err := virtCli.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var controlledPod *k8sv1.Pod
	for podIndex := range pods.Items {
		pod := &pods.Items[podIndex]
		if controller.IsControlledBy(pod, vmi) {
			controlledPod = pod
			break
		}
	}

	if controlledPod == nil {
		return nil, fmt.Errorf("no controlled pod was found for VMI")
	}

	return controlledPod, nil
}
