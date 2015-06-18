/*
Copyright 2014 The Kubernetes Authors All rights reserved.

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

package controller

import (
	"github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/client"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/client/cache"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/controller/framework"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/fields"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/labels"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/runtime"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/watch"
	"github.com/golang/glog"
)

type DaemonManager struct {
	kubeClient client.Interface
}

func NewDaemonManager(kubeClient client.Interface) *DaemonManager {
	dm := &DaemonManager{
		kubeClient: kubeClient,
	}
	framework.NewInformer(
		&cache.ListWatch{
			ListFunc: func() (runtime.Object, error) {
				return dm.kubeClient.Pods(api.NamespaceAll).List(labels.Everything(), fields.Everything())
			},
			WatchFunc: func(rv string) (watch.Interface, error) {
				return dm.kubeClient.Pods(api.NamespaceAll).Watch(labels.Everything(), fields.Everything(), rv)
			},
		},
		&api.Pod{},
		PodRelistPeriod,
		framework.ResourceEventHandlerFuncs{
			AddFunc: dm.addPod,
			// This invokes the rc for every pod change, eg: host assignment. Though this might seem like overkill
			// the most frequent pod update is status, and the associated rc will only list from local storage, so
			// it should be ok.
			UpdateFunc: dm.updatePod,
			DeleteFunc: dm.deletePod,
		},
	)
	return dm
}

func (dm *DaemonManager) Run(workers int, stopCh <-chan struct{}) {
	glog.Infoln("The DM is up!")
}

func (dm *DaemonManager) addPod(obj interface{}) {
	glog.Infoln("DM detected new pod")
}

func (dm *DaemonManager) updatePod(old, cur interface{}) {
	glog.Infoln("DM detected pod update")
}

func (dm *DaemonManager) deletePod(obj interface{}) {
	glog.Infoln("DM detected pod deletion")
}
