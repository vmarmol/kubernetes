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

package client

import (
	"github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/fields"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/labels"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/watch"
)

// DaemonControllersNamespacer has methods to work with DaemonController resources in a namespace
type DaemonControllersNamespacer interface {
	DaemonControllers(namespace string) DaemonControllerInterface
}

type DaemonControllerInterface interface {
	List(selector labels.Selector) (*api.DaemonControllerList, error)
	Get(name string) (*api.DaemonController, error)
	Create(ctrl *api.DaemonController) (*api.DaemonController, error)
	Update(ctrl *api.DaemonController) (*api.DaemonController, error)
	Delete(name string) error
	Watch(label labels.Selector, field fields.Selector, resourceVersion string) (watch.Interface, error)
}

// daemonControllers implements DaemonControllersNamespacer interface
type daemonControllers struct {
	r  *Client
	ns string
}

func newDaemonControllers(c *Client, namespace string) *daemonControllers {
	return &daemonControllers{c, namespace}
}

func (c *daemonControllers) List(selector labels.Selector) (result *api.DaemonControllerList, err error) {
	result = &api.DaemonControllerList{}
	err = c.r.Get().Namespace(c.ns).Resource("daemonControllers").LabelsSelectorParam(selector).Do().Into(result)
	return
}

// Get returns information about a particular daemon controller.
func (c *daemonControllers) Get(name string) (result *api.DaemonController, err error) {
	result = &api.DaemonController{}
	err = c.r.Get().Namespace(c.ns).Resource("daemonControllers").Name(name).Do().Into(result)
	return
}

// Create creates a new daemon controller.
func (c *daemonControllers) Create(controller *api.DaemonController) (result *api.DaemonController, err error) {
	result = &api.DaemonController{}
	err = c.r.Post().Namespace(c.ns).Resource("daemonControllers").Body(controller).Do().Into(result)
	return
}

// Update updates an existing daemon controller.
func (c *daemonControllers) Update(controller *api.DaemonController) (result *api.DaemonController, err error) {
	result = &api.DaemonController{}
	err = c.r.Put().Namespace(c.ns).Resource("daemonControllers").Name(controller.Name).Body(controller).Do().Into(result)
	return
}

// Delete deletes an existing daemon controller.
func (c *daemonControllers) Delete(name string) error {
	return c.r.Delete().Namespace(c.ns).Resource("daemonControllers").Name(name).Do().Error()
}

// Watch returns a watch.Interface that watches the requested controllers.
func (c *daemonControllers) Watch(label labels.Selector, field fields.Selector, resourceVersion string) (watch.Interface, error) {
	return c.r.Get().
		Prefix("watch").
		Namespace(c.ns).
		Resource("daemonControllers").
		Param("resourceVersion", resourceVersion).
		LabelsSelectorParam(label).
		FieldsSelectorParam(field).
		Watch()
}
