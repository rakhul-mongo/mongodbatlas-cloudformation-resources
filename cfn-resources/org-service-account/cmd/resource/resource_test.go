// Copyright 2026 MongoDB Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package resource_test

import (
	"testing"

	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/handler"
	"github.com/mongodb/mongodbatlas-cloudformation-resources/org-service-account/cmd/resource"
	"github.com/stretchr/testify/assert"
)

func TestConstants(t *testing.T) {
	assert.Equal(t, []string{"OrgId", "Name", "Description", "Roles"}, resource.CreateRequiredFields)
	assert.Equal(t, []string{"OrgId", "ClientId"}, resource.ReadRequiredFields)
	assert.Equal(t, []string{"OrgId", "ClientId"}, resource.UpdateRequiredFields)
	assert.Equal(t, []string{"OrgId", "ClientId"}, resource.DeleteRequiredFields)
	assert.Equal(t, []string{"OrgId"}, resource.ListRequiredFields)
}

func TestValidationErrors(t *testing.T) {
	tests := []struct {
		name         string
		operation    func(handler.Request, *resource.Model, *resource.Model) (handler.ProgressEvent, error)
		currentModel *resource.Model
	}{
		{"Create_missingOrgId", resource.Create, &resource.Model{}},
		{"Read_missingOrgId", resource.Read, &resource.Model{}},
		{"Update_missingOrgId", resource.Update, &resource.Model{}},
		{"Delete_missingOrgId", resource.Delete, &resource.Model{}},
		{"List_missingOrgId", resource.List, &resource.Model{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := tt.operation(handler.Request{}, nil, tt.currentModel)
			assert.NoError(t, err)
			assert.Equal(t, handler.Failed, event.OperationStatus)
		})
	}
}
