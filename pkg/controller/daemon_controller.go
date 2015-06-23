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
	"time"

	"github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/client"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/client/cache"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/client/record"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/controller/framework"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/fields"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/labels"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/runtime"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/util"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/util/workqueue"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/watch"
	"github.com/golang/glog"
)

var (
	dcKeyFunc = framework.DeletionHandlingMetaNamespaceKeyFunc
)

const (
	// Daemon Controllers don't need relisting.
	FullDaemonControllerResyncPeriod = 0
	// Nodes don't need relisting.
	FullNodeResyncPeriod = 0
)

type DaemonManager struct {
	kubeClient client.Interface
	podControl PodControlInterface

	// To allow injection of syncDaemonController for testing.
	syncHandler func(dcKey string) error
	// A store of pods, populated by the podController.
	dcStore cache.StoreToPodLister
	// Watches changes to all pods.
	dcController *framework.Controller
	// Watches changes to all nodes.
	nodeController *framework.Controller
	// Controllers that need to be updated.
	queue *workqueue.Type
}

func NewDaemonManager(kubeClient client.Interface) *DaemonManager {
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(kubeClient.Events(""))

	dm := &DaemonManager{
		kubeClient: kubeClient,
		podControl: RealPodControl{
			kubeClient: kubeClient,
			recorder:   eventBroadcaster.NewRecorder(api.EventSource{Component: "daemon-controller"}),
		},
		queue: workqueue.New(),
	}
	// Manage addition/update of daemon controllers.
	dm.dcStore.Store, dm.dcController = framework.NewInformer(
		&cache.ListWatch{
			ListFunc: func() (runtime.Object, error) {
				return dm.kubeClient.DaemonControllers(api.NamespaceAll).List(labels.Everything())
			},
			WatchFunc: func(rv string) (watch.Interface, error) {
				return dm.kubeClient.DaemonControllers(api.NamespaceAll).Watch(labels.Everything(), fields.Everything(), rv)
			},
		},
		&api.DaemonController{},
		FullDaemonControllerResyncPeriod,
		framework.ResourceEventHandlerFuncs{
			AddFunc:    dm.enqueueController,
			UpdateFunc: func(old, cur interface{}) {},
			DeleteFunc: dm.enqueueController,
		},
	)
	// Watch for new nodes or updates to nodes - daemons are typically launched on new nodes as well,
	_, dm.nodeController = framework.NewInformer(
		&cache.ListWatch{
			ListFunc: func() (runtime.Object, error) {
				return dm.kubeClient.Nodes().List(labels.Everything(), fields.Everything())
			},
			WatchFunc: func(rv string) (watch.Interface, error) {
				return dm.kubeClient.Nodes().Watch(labels.Everything(), fields.Everything(), rv)
			},
		},
		&api.Node{},
		FullNodeResyncPeriod,
		framework.ResourceEventHandlerFuncs{
			AddFunc:    dm.addNode,
			UpdateFunc: func(old, cur interface{}) {},
			DeleteFunc: dm.deleteNode,
		},
	)
	dm.syncHandler = dm.syncDaemonController
	return dm
}

func (dm *DaemonManager) Run(workers int, stopCh <-chan struct{}) {
	glog.Infoln("The DM is up!")
	go dm.dcController.Run(stopCh)
	go dm.nodeController.Run(stopCh)
	for i := 0; i < workers; i++ {
		go util.Until(dm.worker, time.Second, stopCh)
	}
	<-stopCh
	glog.Infof("Shutting down Daemon Controller Manager")
	dm.queue.ShutDown()
}

func (dm *DaemonManager) worker() {
	for {
		func() {
			key, quit := dm.queue.Get()
			if quit {
				return
			}
			defer dm.queue.Done(key)
			err := dm.syncHandler(key.(string))
			if err != nil {
				glog.Errorf("Error syncing daemon controller: %v", err)
			}
		}()
	}
}

func (dm *DaemonManager) enqueueController(obj interface{}) {
	key, err := dcKeyFunc(obj)
	if err != nil {
		glog.Errorf("Couldn't get key for object %+v: %v", obj, err)
		return
	}
	dm.queue.Add(key)
}

func (dm *DaemonManager) addNode(obj interface{}) {
	glog.Infoln("Node has been added")
}

func (dm *DaemonManager) deleteNode(obj interface{}) {
	glog.Infoln("Node has been removed")
}

func (dm *DaemonManager) syncDaemonController(key string) error {
	startTime := time.Now()
	defer func() {
		glog.V(4).Infof("Finished syncing daemon controller %q (%v)", key, time.Now().Sub(startTime))
	}()
	obj, exists, err := dm.dcStore.Store.GetByKey(key)
	if !exists {
		glog.Infof("Daemon Controller has been deleted %v", key)
		return nil
	}
	if err != nil {
		glog.Infof("Unable to retrieve dc %v from store: %v", key, err)
		dm.queue.Add(key)
		return err
	}
	dc := obj.(*api.DaemonController)

	nodeList, err := dm.kubeClient.Nodes().List(labels.Everything(), fields.Everything())
	if err != nil {
		glog.Errorf("Couldn't get list of nodes when adding daemon controller %+v: %v", obj, err)
		return err
	}
	nodeNameList := make([]string, len(nodeList.Items))
	for i := range nodeList.Items {
		nodeNameList[i] = nodeList.Items[i].Name
	}

	dm.podControl.createReplicaOnNodes(dc.Namespace, dc, nodeNameList)
	return nil
}
