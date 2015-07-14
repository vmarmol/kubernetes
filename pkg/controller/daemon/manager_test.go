/*
Copyright 2015 The Kubernetes Authors All rights reserved.

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

package daemon

import (
	"fmt"
	"testing"

	"github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/api/testapi"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/client"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/client/cache"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/controller"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/securitycontext"
)

var (
	simpleDaemonLabel  = map[string]string{"name": "simple-daemon", "type": "production"}
	simpleDaemonLabel2 = map[string]string{"name": "simple-daemon", "type": "test"}
	simpleNodeLabel    = map[string]string{"color": "blue", "speed": "fast"}
	simpleNodeLabel2   = map[string]string{"color": "red", "speed": "fast"}
)

func init() {
	api.ForTesting_ReferencesAllowBlankSelfLinks = true
}

func newDaemon(name string) *api.DaemonController {
	return &api.DaemonController{
		TypeMeta: api.TypeMeta{APIVersion: testapi.Version()},
		ObjectMeta: api.ObjectMeta{
			Name:      name,
			Namespace: api.NamespaceDefault,
		},
		Spec: api.DaemonControllerSpec{
			Selector: simpleDaemonLabel,
			Template: &api.PodTemplateSpec{
				ObjectMeta: api.ObjectMeta{
					Labels: simpleDaemonLabel,
				},
				Spec: api.PodSpec{
					Containers: []api.Container{
						{
							Image: "foo/bar",
							TerminationMessagePath: api.TerminationMessagePathDefault,
							ImagePullPolicy:        api.PullIfNotPresent,
							SecurityContext:        securitycontext.ValidSecurityContextWithContainerDefaults(),
						},
					},
					DNSPolicy: api.DNSDefault,
				},
			},
		},
	}
}

func newNode(name string, label map[string]string) *api.Node {
	return &api.Node{
		TypeMeta: api.TypeMeta{APIVersion: testapi.Version()},
		ObjectMeta: api.ObjectMeta{
			Name:      name,
			Labels:    label,
			Namespace: api.NamespaceDefault,
		},
	}
}

func addNodes(nodeStore cache.Store, startIndex, numNodes int, label map[string]string) {
	for i := startIndex; i < startIndex+numNodes; i++ {
		nodeStore.Add(newNode(fmt.Sprintf("node-%d", i), label))
	}
}

func newPod(podName string, nodeName string, label map[string]string) *api.Pod {
	pod := &api.Pod{
		TypeMeta: api.TypeMeta{APIVersion: testapi.Version()},
		ObjectMeta: api.ObjectMeta{
			GenerateName: podName,
			Labels:       label,
			Namespace:    api.NamespaceDefault,
		},
		Spec: api.PodSpec{
			NodeName: nodeName,
			Containers: []api.Container{
				{
					Image: "foo/bar",
					TerminationMessagePath: api.TerminationMessagePathDefault,
					ImagePullPolicy:        api.PullIfNotPresent,
					SecurityContext:        securitycontext.ValidSecurityContextWithContainerDefaults(),
				},
			},
			DNSPolicy: api.DNSDefault,
		},
	}
	api.GenerateName(api.SimpleNameGenerator, &pod.ObjectMeta)
	return pod
}

func addPods(podStore cache.Store, nodeName string, label map[string]string, number int) {
	for i := 0; i < number; i++ {
		podStore.Add(newPod(fmt.Sprintf("%s-", nodeName), nodeName, label))
	}
}

func makeTestManager() (*DaemonManager, *controller.FakePodControl) {
	client := client.NewOrDie(&client.Config{Host: "", Version: testapi.Version()})
	manager := NewDaemonManager(client)
	podControl := &controller.FakePodControl{}
	manager.podControl = podControl
	return manager, podControl
}

func syncAndValidateDaemons(t *testing.T, manager *DaemonManager, daemon *api.DaemonController, podControl *controller.FakePodControl, expectedCreates, expectedDeletes int) {
	key, err := controller.KeyFunc(daemon)
	if err != nil {
		t.Errorf("Could not get key for daemon.")
	}
	manager.syncHandler(key)
	controller.ValidateSyncDaemons(t, podControl, expectedCreates, expectedDeletes)
}

// Daemon without node selectors should launch pods on every node.
func TestSimpleDaemonLaunchesPods(t *testing.T) {
	manager, podControl := makeTestManager()
	addNodes(manager.nodeStore.Store, 0, 5, nil)
	daemon := newDaemon("foo")
	manager.dcStore.Add(daemon)
	syncAndValidateDaemons(t, manager, daemon, podControl, 5, 0)
}

// Daemon without node selectors should launch pods on every node.
func TestNoNodesDoesNothing(t *testing.T) {
	manager, podControl := makeTestManager()
	daemon := newDaemon("foo")
	manager.dcStore.Add(daemon)
	syncAndValidateDaemons(t, manager, daemon, podControl, 0, 0)
}

// Daemon without node selectors should launch pods on every node.
func TestOneNodeDaemonLaunchesPod(t *testing.T) {
	manager, podControl := makeTestManager()
	manager.nodeStore.Add(newNode("only-node", nil))
	daemon := newDaemon("foo")
	manager.dcStore.Add(daemon)
	syncAndValidateDaemons(t, manager, daemon, podControl, 1, 0)
}

// Manager should not create pods on nodes which have daemon pods, and should remove excess pods from nodes that have extra pods.
func TestDealsWithExistingPods(t *testing.T) {
	manager, podControl := makeTestManager()
	addNodes(manager.nodeStore.Store, 0, 5, nil)
	addPods(manager.podStore.Store, "node-1", simpleDaemonLabel, 1)
	addPods(manager.podStore.Store, "node-2", simpleDaemonLabel, 2)
	addPods(manager.podStore.Store, "node-3", simpleDaemonLabel, 5)
	addPods(manager.podStore.Store, "node-4", simpleDaemonLabel2, 2)
	daemon := newDaemon("foo")
	manager.dcStore.Add(daemon)
	syncAndValidateDaemons(t, manager, daemon, podControl, 2, 5)
}

// Daemon with node selector should launch pods on nodes matching selector.
func TestSelectorDaemonLaunchesPods(t *testing.T) {
	manager, podControl := makeTestManager()
	addNodes(manager.nodeStore.Store, 0, 4, nil)
	addNodes(manager.nodeStore.Store, 4, 3, simpleNodeLabel)
	daemon := newDaemon("foo")
	daemon.Spec.Template.Spec.NodeSelector = simpleNodeLabel
	manager.dcStore.Add(daemon)
	syncAndValidateDaemons(t, manager, daemon, podControl, 3, 0)
}

// Daemon with node selector should delete pods from nodes that do not satisfy selector.
func TestSelectorDaemonDeletesUnselectedPods(t *testing.T) {
	manager, podControl := makeTestManager()
	addNodes(manager.nodeStore.Store, 0, 5, nil)
	addNodes(manager.nodeStore.Store, 5, 5, simpleNodeLabel)
	addPods(manager.podStore.Store, "node-0", simpleDaemonLabel2, 2)
	addPods(manager.podStore.Store, "node-1", simpleDaemonLabel, 3)
	addPods(manager.podStore.Store, "node-1", simpleDaemonLabel2, 1)
	addPods(manager.podStore.Store, "node-4", simpleDaemonLabel, 1)
	daemon := newDaemon("foo")
	daemon.Spec.Template.Spec.NodeSelector = simpleNodeLabel
	manager.dcStore.Add(daemon)
	syncAndValidateDaemons(t, manager, daemon, podControl, 5, 4)
}

// Daemon with node selector should launch pods on nodes matching selector, but also deal with existing pods on nodes.
func TestSelectorDaemonDealsWithExistingPods(t *testing.T) {
	manager, podControl := makeTestManager()
	addNodes(manager.nodeStore.Store, 0, 5, nil)
	addNodes(manager.nodeStore.Store, 5, 5, simpleNodeLabel)
	addPods(manager.podStore.Store, "node-0", simpleDaemonLabel, 1)
	addPods(manager.podStore.Store, "node-1", simpleDaemonLabel, 3)
	addPods(manager.podStore.Store, "node-1", simpleDaemonLabel2, 2)
	addPods(manager.podStore.Store, "node-2", simpleDaemonLabel, 4)
	addPods(manager.podStore.Store, "node-6", simpleDaemonLabel, 13)
	addPods(manager.podStore.Store, "node-7", simpleDaemonLabel2, 4)
	addPods(manager.podStore.Store, "node-9", simpleDaemonLabel, 1)
	addPods(manager.podStore.Store, "node-9", simpleDaemonLabel2, 1)
	daemon := newDaemon("foo")
	daemon.Spec.Template.Spec.NodeSelector = simpleNodeLabel
	manager.dcStore.Add(daemon)
	syncAndValidateDaemons(t, manager, daemon, podControl, 3, 20)
}

// Daemon with node selector which does not match any node labels should not launch pods.
func TestBadSelectorDaemonDoesNothing(t *testing.T) {
	manager, podControl := makeTestManager()
	addNodes(manager.nodeStore.Store, 0, 4, nil)
	addNodes(manager.nodeStore.Store, 4, 3, simpleNodeLabel)
	daemon := newDaemon("foo")
	daemon.Spec.Template.Spec.NodeSelector = simpleNodeLabel2
	manager.dcStore.Add(daemon)
	syncAndValidateDaemons(t, manager, daemon, podControl, 0, 0)
}

// Daemon with node name should launch pod on node with corresponding name.
func TestNameDaemonLaunchesPods(t *testing.T) {
	manager, podControl := makeTestManager()
	addNodes(manager.nodeStore.Store, 0, 5, nil)
	daemon := newDaemon("foo")
	daemon.Spec.Template.Spec.NodeName = "node-0"
	manager.dcStore.Add(daemon)
	syncAndValidateDaemons(t, manager, daemon, podControl, 1, 0)
}

// Daemon with node name that does not exist should not launch pods.
func TestBadNameDaemonDoesNothing(t *testing.T) {
	manager, podControl := makeTestManager()
	addNodes(manager.nodeStore.Store, 0, 5, nil)
	daemon := newDaemon("foo")
	daemon.Spec.Template.Spec.NodeName = "node-10"
	manager.dcStore.Add(daemon)
	syncAndValidateDaemons(t, manager, daemon, podControl, 0, 0)
}

// Daemon with node selector, and node name, matching a node, should launch a pod on the node.
func TestNameAndSelectorDaemonLaunchesPods(t *testing.T) {
	manager, podControl := makeTestManager()
	addNodes(manager.nodeStore.Store, 0, 4, nil)
	addNodes(manager.nodeStore.Store, 4, 3, simpleNodeLabel)
	daemon := newDaemon("foo")
	daemon.Spec.Template.Spec.NodeSelector = simpleNodeLabel
	daemon.Spec.Template.Spec.NodeName = "node-6"
	manager.dcStore.Add(daemon)
	syncAndValidateDaemons(t, manager, daemon, podControl, 1, 0)
}

// Daemon with node selector that matches some nodes, and node name that matches a different node, should do nothing.
func TestInconsistentNameSelectorDaemonDoesNothing(t *testing.T) {
	manager, podControl := makeTestManager()
	addNodes(manager.nodeStore.Store, 0, 4, nil)
	addNodes(manager.nodeStore.Store, 4, 3, simpleNodeLabel)
	daemon := newDaemon("foo")
	daemon.Spec.Template.Spec.NodeSelector = simpleNodeLabel
	daemon.Spec.Template.Spec.NodeName = "node-0"
	manager.dcStore.Add(daemon)
	syncAndValidateDaemons(t, manager, daemon, podControl, 0, 0)
}

func TestExperiment(t *testing.T) {
	p := controller.FakePodControl{}
	p.CreateReplica = func(namespace string, spec *api.ReplicationController) error {
		return nil
	}
}
