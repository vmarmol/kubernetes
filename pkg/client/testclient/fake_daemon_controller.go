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

package testclient

import (
	"github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/fields"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/labels"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/watch"
)

// FakeDaemonControllers implements DaemonControllerInterface. Meant to be embedded into a struct to get a default
// implementation. This makes faking out just the method you want to test easier.
type FakeDaemonControllers struct {
	Fake      *Fake
	Namespace string
}

const (
	GetDaemonControllerAction    = "get-daemonController"
	UpdateDaemonControllerAction = "update-daemonController"
	WatchDaemonControllerAction  = "watch-daemonController"
	DeleteDaemonControllerAction = "delete-daemonController"
	ListDaemonControllerAction   = "list-daemonControllers"
	CreateDaemonControllerAction = "create-daemonController"
)

func (c *FakeDaemonControllers) List(selector labels.Selector) (*api.DaemonControllerList, error) {
	obj, err := c.Fake.Invokes(FakeAction{Action: ListDaemonControllerAction}, &api.DaemonControllerList{})
	return obj.(*api.DaemonControllerList), err
}

func (c *FakeDaemonControllers) Get(name string) (*api.DaemonController, error) {
	obj, err := c.Fake.Invokes(FakeAction{Action: GetDaemonControllerAction, Value: name}, &api.DaemonController{})
	return obj.(*api.DaemonController), err
}

func (c *FakeDaemonControllers) Create(controller *api.DaemonController) (*api.DaemonController, error) {
	obj, err := c.Fake.Invokes(FakeAction{Action: CreateDaemonControllerAction, Value: controller}, &api.DaemonController{})
	return obj.(*api.DaemonController), err
}

func (c *FakeDaemonControllers) Update(controller *api.DaemonController) (*api.DaemonController, error) {
	obj, err := c.Fake.Invokes(FakeAction{Action: UpdateDaemonControllerAction, Value: controller}, &api.DaemonController{})
	return obj.(*api.DaemonController), err
}

func (c *FakeDaemonControllers) Delete(name string) error {
	_, err := c.Fake.Invokes(FakeAction{Action: DeleteDaemonControllerAction, Value: name}, &api.DaemonController{})
	return err
}

func (c *FakeDaemonControllers) Watch(label labels.Selector, field fields.Selector, resourceVersion string) (watch.Interface, error) {
	c.Fake.Actions = append(c.Fake.Actions, FakeAction{Action: WatchDaemonControllerAction, Value: resourceVersion})
	return c.Fake.Watch, nil
}
