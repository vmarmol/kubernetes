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

package client

import (
	"testing"

	"github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/api/testapi"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/labels"
)

func getDCResourceName() string {
	return "daemonControllers"
}

func TestListDaemonControllers(t *testing.T) {
	ns := api.NamespaceAll
	c := &testClient{
		Request: testRequest{
			Method: "GET",
			Path:   testapi.ResourcePath(getDCResourceName(), ns, ""),
		},
		Response: Response{StatusCode: 200,
			Body: &api.DaemonControllerList{
				Items: []api.DaemonController{
					{
						ObjectMeta: api.ObjectMeta{
							Name: "foo",
							Labels: map[string]string{
								"foo":  "bar",
								"name": "baz",
							},
						},
						Spec: api.DaemonControllerSpec{
							Template: &api.PodTemplateSpec{},
						},
					},
				},
			},
		},
	}
	receivedControllerList, err := c.Setup().DaemonControllers(ns).List(labels.Everything())
	c.Validate(t, receivedControllerList, err)

}

func TestGetDaemonController(t *testing.T) {
	ns := api.NamespaceDefault
	c := &testClient{
		Request: testRequest{Method: "GET", Path: testapi.ResourcePath(getDCResourceName(), ns, "foo"), Query: buildQueryValues(ns, nil)},
		Response: Response{
			StatusCode: 200,
			Body: &api.DaemonController{
				ObjectMeta: api.ObjectMeta{
					Name: "foo",
					Labels: map[string]string{
						"foo":  "bar",
						"name": "baz",
					},
				},
				Spec: api.DaemonControllerSpec{
					Template: &api.PodTemplateSpec{},
				},
			},
		},
	}
	receivedController, err := c.Setup().DaemonControllers(ns).Get("foo")
	c.Validate(t, receivedController, err)
}

func TestGetDaemonControllerWithNoName(t *testing.T) {
	ns := api.NamespaceDefault
	c := &testClient{Error: true}
	receivedPod, err := c.Setup().DaemonControllers(ns).Get("")
	if (err != nil) && (err.Error() != nameRequiredError) {
		t.Errorf("Expected error: %v, but got %v", nameRequiredError, err)
	}

	c.Validate(t, receivedPod, err)
}

func TestUpdateDaemonController(t *testing.T) {
	ns := api.NamespaceDefault
	requestController := &api.DaemonController{
		ObjectMeta: api.ObjectMeta{Name: "foo", ResourceVersion: "1"},
	}
	c := &testClient{
		Request: testRequest{Method: "PUT", Path: testapi.ResourcePath(getDCResourceName(), ns, "foo"), Query: buildQueryValues(ns, nil)},
		Response: Response{
			StatusCode: 200,
			Body: &api.DaemonController{
				ObjectMeta: api.ObjectMeta{
					Name: "foo",
					Labels: map[string]string{
						"foo":  "bar",
						"name": "baz",
					},
				},
				Spec: api.DaemonControllerSpec{
					Template: &api.PodTemplateSpec{},
				},
			},
		},
	}
	receivedController, err := c.Setup().DaemonControllers(ns).Update(requestController)
	c.Validate(t, receivedController, err)
}

func TestDeleteDaemonController(t *testing.T) {
	ns := api.NamespaceDefault
	c := &testClient{
		Request:  testRequest{Method: "DELETE", Path: testapi.ResourcePath(getDCResourceName(), ns, "foo"), Query: buildQueryValues(ns, nil)},
		Response: Response{StatusCode: 200},
	}
	err := c.Setup().DaemonControllers(ns).Delete("foo")
	c.Validate(t, nil, err)
}

func TestCreateDaemonController(t *testing.T) {
	ns := api.NamespaceDefault
	requestController := &api.DaemonController{
		ObjectMeta: api.ObjectMeta{Name: "foo"},
	}
	c := &testClient{
		Request: testRequest{Method: "POST", Path: testapi.ResourcePath(getDCResourceName(), ns, ""), Body: requestController, Query: buildQueryValues(ns, nil)},
		Response: Response{
			StatusCode: 200,
			Body: &api.DaemonController{
				ObjectMeta: api.ObjectMeta{
					Name: "foo",
					Labels: map[string]string{
						"foo":  "bar",
						"name": "baz",
					},
				},
				Spec: api.DaemonControllerSpec{
					Template: &api.PodTemplateSpec{},
				},
			},
		},
	}
	receivedController, err := c.Setup().DaemonControllers(ns).Create(requestController)
	c.Validate(t, receivedController, err)
}
