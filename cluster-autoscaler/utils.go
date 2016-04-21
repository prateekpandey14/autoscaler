/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package main

import (
	"time"

	kube_api "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/cache"
	kube_client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"
)

// UnscheduledPodLister list unscheduled pods
type UnscheduledPodLister struct {
	podLister *cache.StoreToPodLister
}

// List returns all unscheduled pods.
func (unscheduledPodLister *UnscheduledPodLister) List() ([]*kube_api.Pod, error) {
	//TODO: Extra filter based on pod condition.
	return unscheduledPodLister.podLister.List(labels.Everything())
}

// NewUnscheduledPodLister returns a lister providing pods that failed to be scheduled.
func NewUnscheduledPodLister(kubeClient *kube_client.Client) *UnscheduledPodLister {
	// watch unscheduled pods
	selector := fields.ParseSelectorOrDie("spec.nodeName==" + "" + ",status.phase!=" +
		string(kube_api.PodSucceeded) + ",status.phase!=" + string(kube_api.PodFailed))
	podListWatch := cache.NewListWatchFromClient(kubeClient, "pods", kube_api.NamespaceAll, selector)
	podLister := &cache.StoreToPodLister{Store: cache.NewStore(cache.MetaNamespaceKeyFunc)}
	podReflector := cache.NewReflector(podListWatch, &kube_api.Pod{}, podLister.Store, time.Hour)
	podReflector.Run()

	return &UnscheduledPodLister{
		podLister: podLister,
	}
}

// ReadyNodeLister lists ready nodes.
type ReadyNodeLister struct {
	nodeLister *cache.StoreToNodeLister
}

// List returns ready nodes.
func (readyNodeLister *ReadyNodeLister) List() ([]kube_api.Node, error) {
	nodes, err := readyNodeLister.nodeLister.List()
	if err != nil {
		return []kube_api.Node{}, err
	}
	readyNodes := make([]kube_api.Node, 0, len(nodes.Items))
	for _, node := range nodes.Items {
		for _, condition := range node.Status.Conditions {
			if condition.Type == kube_api.NodeReady && condition.Status == kube_api.ConditionTrue {
				readyNodes = append(readyNodes, node)
				break
			}
		}
	}
	return readyNodes, nil
}

// NewNodeLister builds a node lister.
func NewNodeLister(kubeClient *kube_client.Client) *ReadyNodeLister {
	listWatcher := cache.NewListWatchFromClient(kubeClient, "nodes", kube_api.NamespaceAll, fields.Everything())
	nodeLister := &cache.StoreToNodeLister{Store: cache.NewStore(cache.MetaNamespaceKeyFunc)}
	reflector := cache.NewReflector(listWatcher, &kube_api.Node{}, nodeLister.Store, time.Hour)
	reflector.Run()
	return &ReadyNodeLister{
		nodeLister: nodeLister,
	}
}

// GetNewestNode returns the newest node from the given list.
func GetNewestNode(nodes []kube_api.Node) *kube_api.Node {
	var result *kube_api.Node
	for i, node := range nodes {
		if result == nil || node.CreationTimestamp.After(result.CreationTimestamp.Time) {
			result = &(nodes[i])
		}
	}
	return result
}

// GetOldestFailedSchedulingTrail returns the oldest time when a pod from the given list failed to
// be scheduled.
func GetOldestFailedSchedulingTrail(pods []*kube_api.Pod) *time.Time {
	// Dummy implementation.
	//TODO: Implement once pod condition is there.
	now := time.Now()
	return &now
}
