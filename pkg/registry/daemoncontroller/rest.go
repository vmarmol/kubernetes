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

package daemoncontroller

import (
	"fmt"
	"strconv"

	"github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/api/validation"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/fields"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/labels"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/registry/generic"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/runtime"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/util/fielderrors"
)

// dcStrategy implements verification logic for daemon Controllers.
type dcStrategy struct {
	runtime.ObjectTyper
	api.NameGenerator
}

// Strategy is the default logic that applies when creating and updating Daemon Controller objects.
var Strategy = dcStrategy{api.Scheme, api.SimpleNameGenerator}

// NamespaceScoped returns true because all Daemon Controllers need to be within a namespace.
func (dcStrategy) NamespaceScoped() bool {
	return true
}

// PrepareForCreate clears the status of a daemon controller before creation.
func (dcStrategy) PrepareForCreate(obj runtime.Object) {
	controller := obj.(*api.DaemonController)
	controller.Status = api.DaemonControllerStatus{}
}

// PrepareForUpdate clears fields that are not allowed to be set by end users on update.
func (dcStrategy) PrepareForUpdate(obj, old runtime.Object) {

}

// Validate validates a new daemon controller.
func (dcStrategy) Validate(ctx api.Context, obj runtime.Object) fielderrors.ValidationErrorList {
	controller := obj.(*api.DaemonController)
	return validation.ValidateDaemonController(controller)
}

// AllowCreateOnUpdate is false for daemon controllers; this means a POST is
// needed to create one.
func (dcStrategy) AllowCreateOnUpdate() bool {
	return false
}

// ValidateUpdate is the default update validation for an end user.
func (dcStrategy) ValidateUpdate(ctx api.Context, obj, old runtime.Object) fielderrors.ValidationErrorList {
	validationErrorList := validation.ValidateDaemonController(obj.(*api.DaemonController))
	updateErrorList := validation.ValidateDaemonControllerUpdate(old.(*api.DaemonController), obj.(*api.DaemonController))
	return append(validationErrorList, updateErrorList...)
}

// ControllerToSelectableFields returns a label set that represents the object.
func ControllerToSelectableFields(controller *api.DaemonController) fields.Set {
	return fields.Set{
		"metadata.name":                 controller.Name,
		"status.currentNumberScheduled": strconv.Itoa(controller.Status.CurrentNumberScheduled),
		"status.numberMisscheduled":     strconv.Itoa(controller.Status.NumberMisscheduled),
		"status.desiredNumberScheduled": strconv.Itoa(controller.Status.DesiredNumberScheduled),
	}
}

// MatchController is the filter used by the generic etcd backend to route
// watch events from etcd to clients of the apiserver only interested in specific
// labels/fields.
func MatchController(label labels.Selector, field fields.Selector) generic.Matcher {
	return &generic.SelectionPredicate{
		Label: label,
		Field: field,
		GetAttrs: func(obj runtime.Object) (labels.Set, fields.Set, error) {
			dc, ok := obj.(*api.DaemonController)
			if !ok {
				return nil, nil, fmt.Errorf("given object is not a daemon controller.")
			}
			return labels.Set(dc.ObjectMeta.Labels), ControllerToSelectableFields(dc), nil
		},
	}
}
