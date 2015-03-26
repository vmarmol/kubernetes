/*
Copyright 2014 Google Inc. All rights reserved.

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

package kubelet

import (
	"sync"

	"github.com/GoogleCloudPlatform/kubernetes/pkg/api"
)

// ContainerRefManager manages the references for the containers.
// The references are used for reporting events such as creation,
// failure, etc.
type ContainerRefManager struct {
	sync.RWMutex
	// TODO(yifan): Change type from string to types.UID.
	containerIDToRef map[string]*api.ObjectReference
}

// newContainerRefManager creates and returns a container reference manager
// with empty contents.
func newContainerRefManager() *ContainerRefManager {
	c := ContainerRefManager{}
	c.containerIDToRef = make(map[string]*api.ObjectReference)
	return &c
}

// SetRef stores a reference to a pod's container, associating it with the given container id.
func (c *ContainerRefManager) SetRef(id string, ref *api.ObjectReference) {
	c.Lock()
	defer c.Unlock()
	c.containerIDToRef[id] = ref
}

// ClearRef forgets the given container id and its associated container reference.
func (c *ContainerRefManager) ClearRef(id string) {
	c.Lock()
	defer c.Unlock()
	delete(c.containerIDToRef, id)
}

// GetRef returns the container reference of the given id, or (nil, false) if none is stored.
func (c *ContainerRefManager) GetRef(id string) (ref *api.ObjectReference, ok bool) {
	c.RLock()
	defer c.RUnlock()
	ref, ok = c.containerIDToRef[id]
	return ref, ok
}
