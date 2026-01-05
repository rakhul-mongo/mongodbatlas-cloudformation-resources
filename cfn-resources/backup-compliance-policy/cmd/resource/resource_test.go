// Copyright 2024 MongoDB Inc
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
	"fmt"
	"net/http"
	"testing"
	"time"

	"go.mongodb.org/atlas-sdk/v20250312010/admin"
	"go.mongodb.org/atlas-sdk/v20250312010/mockadmin"

	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/handler"
	"github.com/mongodb/mongodbatlas-cloudformation-resources/backup-compliance-policy/cmd/resource"
	"github.com/mongodb/mongodbatlas-cloudformation-resources/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Helper function to create a test model
func createTestModel() *resource.Model {
	projectID := "507f1f77bcf86cd799439011"
	authorizedEmail := "test@example.com"
	authorizedUserFirstName := "John"
	authorizedUserLastName := "Doe"

	return &resource.Model{
		ProjectId:               &projectID,
		AuthorizedEmail:         &authorizedEmail,
		AuthorizedUserFirstName: &authorizedUserFirstName,
		AuthorizedUserLastName:  &authorizedUserLastName,
	}
}

// Helper function to create a test API policy response
func createTestPolicy() *admin.DataProtectionSettings20231001 {
	projectID := "507f1f77bcf86cd799439011"
	authorizedEmail := "test@example.com"
	authorizedUserFirstName := "John"
	authorizedUserLastName := "Doe"
	state := "ACTIVE"
	updatedUser := "user@example.com"
	copyProtectionEnabled := false
	encryptionAtRestEnabled := true
	pitEnabled := false
	restoreWindowDays := 7
	updatedDate := time.Now()

	onDemandId := "507f1f77bcf86cd799439020"
	onDemandItem := admin.BackupComplianceOnDemandPolicyItem{
		Id:                &onDemandId,
		FrequencyInterval: 1,
		FrequencyType:     "ondemand",
		RetentionUnit:     "days",
		RetentionValue:    30,
	}

	hourlyId := "507f1f77bcf86cd799439021"
	hourlyItem := admin.BackupComplianceScheduledPolicyItem{
		Id:                &hourlyId,
		FrequencyType:     "hourly",
		FrequencyInterval: 6,
		RetentionUnit:     "days",
		RetentionValue:    7,
	}

	dailyId := "507f1f77bcf86cd799439022"
	dailyItem := admin.BackupComplianceScheduledPolicyItem{
		Id:                &dailyId,
		FrequencyType:     "daily",
		FrequencyInterval: 1,
		RetentionUnit:     "days",
		RetentionValue:    30,
	}

	weeklyId1 := "507f1f77bcf86cd799439023"
	weeklyItem1 := admin.BackupComplianceScheduledPolicyItem{
		Id:                &weeklyId1,
		FrequencyType:     "weekly",
		FrequencyInterval: 1,
		RetentionUnit:     "weeks",
		RetentionValue:    4,
	}

	weeklyId2 := "507f1f77bcf86cd799439024"
	weeklyItem2 := admin.BackupComplianceScheduledPolicyItem{
		Id:                &weeklyId2,
		FrequencyType:     "weekly",
		FrequencyInterval: 2,
		RetentionUnit:     "weeks",
		RetentionValue:    8,
	}

	scheduledItems := []admin.BackupComplianceScheduledPolicyItem{
		hourlyItem,
		dailyItem,
		weeklyItem1,
		weeklyItem2,
	}

	policy := &admin.DataProtectionSettings20231001{
		ProjectId:               &projectID,
		AuthorizedEmail:         authorizedEmail,
		AuthorizedUserFirstName: authorizedUserFirstName,
		AuthorizedUserLastName:  authorizedUserLastName,
		State:                   &state,
		UpdatedUser:             &updatedUser,
		UpdatedDate:             &updatedDate,
		CopyProtectionEnabled:   &copyProtectionEnabled,
		EncryptionAtRestEnabled: &encryptionAtRestEnabled,
		PitEnabled:              &pitEnabled,
		RestoreWindowDays:       &restoreWindowDays,
		OnDemandPolicyItem:      &onDemandItem,
		ScheduledPolicyItems:    &scheduledItems,
	}
	return policy
}

// Test validation errors
func TestCreateValidationErrors(t *testing.T) {
	testCases := map[string]struct {
		currentModel   *resource.Model
		expectedStatus handler.Status
		expectedMsg    string
	}{
		"missingProjectId": {
			currentModel: &resource.Model{
				AuthorizedEmail:         util.StringPtr("test@example.com"),
				AuthorizedUserFirstName: util.StringPtr("John"),
				AuthorizedUserLastName:  util.StringPtr("Doe"),
			},
			expectedStatus: handler.Failed,
			expectedMsg:    "required",
		},
		"missingAuthorizedEmail": {
			currentModel: &resource.Model{
				ProjectId:               func() *string { s := "507f1f77bcf86cd799439011"; return &s }(),
				AuthorizedUserFirstName: func() *string { s := "John"; return &s }(),
				AuthorizedUserLastName:  func() *string { s := "Doe"; return &s }(),
			},
			expectedStatus: handler.Failed,
			expectedMsg:    "required",
		},
		"missingAuthorizedUserFirstName": {
			currentModel: &resource.Model{
				ProjectId:              func() *string { s := "507f1f77bcf86cd799439011"; return &s }(),
				AuthorizedEmail:        func() *string { s := "test@example.com"; return &s }(),
				AuthorizedUserLastName: func() *string { s := "Doe"; return &s }(),
			},
			expectedStatus: handler.Failed,
			expectedMsg:    "required",
		},
		"missingAuthorizedUserLastName": {
			currentModel: &resource.Model{
				ProjectId:               func() *string { s := "507f1f77bcf86cd799439011"; return &s }(),
				AuthorizedEmail:         func() *string { s := "test@example.com"; return &s }(),
				AuthorizedUserFirstName: func() *string { s := "John"; return &s }(),
			},
			expectedStatus: handler.Failed,
			expectedMsg:    "required",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			req := handler.Request{
				RequestContext: handler.RequestContext{},
			}
			event, err := resource.Create(req, nil, tc.currentModel)

			require.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, event.OperationStatus)
			assert.Contains(t, event.Message, tc.expectedMsg)
		})
	}
}

func TestReadValidationErrors(t *testing.T) {
	testCases := map[string]struct {
		currentModel   *resource.Model
		expectedStatus handler.Status
		expectedMsg    string
	}{
		"missingProjectId": {
			currentModel:   &resource.Model{},
			expectedStatus: handler.Failed,
			expectedMsg:    "required",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			req := handler.Request{
				RequestContext: handler.RequestContext{},
			}
			event, err := resource.Read(req, nil, tc.currentModel)

			require.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, event.OperationStatus)
			assert.Contains(t, event.Message, tc.expectedMsg)
		})
	}
}

func TestUpdateValidationErrors(t *testing.T) {
	testCases := map[string]struct {
		currentModel   *resource.Model
		expectedStatus handler.Status
		expectedMsg    string
	}{
		"missingProjectId": {
			currentModel: &resource.Model{
				AuthorizedEmail:         util.StringPtr("test@example.com"),
				AuthorizedUserFirstName: util.StringPtr("John"),
				AuthorizedUserLastName:  util.StringPtr("Doe"),
			},
			expectedStatus: handler.Failed,
			expectedMsg:    "required",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			req := handler.Request{
				RequestContext: handler.RequestContext{},
			}
			event, err := resource.Update(req, nil, tc.currentModel)

			require.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, event.OperationStatus)
			assert.Contains(t, event.Message, tc.expectedMsg)
		})
	}
}

func TestDeleteValidationErrors(t *testing.T) {
	testCases := map[string]struct {
		currentModel   *resource.Model
		expectedStatus handler.Status
		expectedMsg    string
	}{
		"missingProjectId": {
			currentModel:   &resource.Model{},
			expectedStatus: handler.Failed,
			expectedMsg:    "required",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			req := handler.Request{
				RequestContext: handler.RequestContext{},
			}
			event, err := resource.Delete(req, nil, tc.currentModel)

			require.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, event.OperationStatus)
			assert.Contains(t, event.Message, tc.expectedMsg)
		})
	}
}

func TestListValidationErrors(t *testing.T) {
	testCases := map[string]struct {
		currentModel   *resource.Model
		expectedStatus handler.Status
		expectedMsg    string
	}{
		"missingProjectId": {
			currentModel:   &resource.Model{},
			expectedStatus: handler.Failed,
			expectedMsg:    "required",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			req := handler.Request{
				RequestContext: handler.RequestContext{},
			}
			event, err := resource.List(req, nil, tc.currentModel)

			require.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, event.OperationStatus)
			assert.Contains(t, event.Message, tc.expectedMsg)
		})
	}
}

// Test CRUD operations with mocks
func TestCreateWithMocks(t *testing.T) {
	// Save original function
	originalInitEnv := resource.InitEnvWithLatestClient
	defer func() {
		resource.InitEnvWithLatestClient = originalInitEnv
	}()

	testCases := map[string]struct {
		req            handler.Request
		currentModel   *resource.Model
		mockSetup      func(*mockadmin.CloudBackupsApi)
		expectedStatus handler.Status
	}{
		"successfulCreate": {
			req: handler.Request{
				RequestContext: handler.RequestContext{},
			},
			currentModel: createTestModel(),
			mockSetup: func(m *mockadmin.CloudBackupsApi) {
				req := admin.UpdateCompliancePolicyApiRequest{ApiService: m}
				m.EXPECT().UpdateCompliancePolicyWithParams(mock.Anything, mock.Anything).Return(req)
				policy := createTestPolicy()
				m.EXPECT().UpdateCompliancePolicyExecute(mock.Anything).Return(policy, &http.Response{StatusCode: 200}, nil)
			},
			expectedStatus: handler.Success,
		},
		"apiError": {
			req: handler.Request{
				RequestContext: handler.RequestContext{},
			},
			currentModel: createTestModel(),
			mockSetup: func(m *mockadmin.CloudBackupsApi) {
				req := admin.UpdateCompliancePolicyApiRequest{ApiService: m}
				m.EXPECT().UpdateCompliancePolicyWithParams(mock.Anything, mock.Anything).Return(req)
				m.EXPECT().UpdateCompliancePolicyExecute(mock.Anything).Return(nil, &http.Response{StatusCode: 500}, fmt.Errorf("API error"))
			},
			expectedStatus: handler.Failed,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			mockApi := mockadmin.NewCloudBackupsApi(t)
			tc.mockSetup(mockApi)

			mockClient := &admin.APIClient{}
			mockClient.CloudBackupsApi = mockApi

			resource.InitEnvWithLatestClient = func(req handler.Request, currentModel *resource.Model, requiredFields []string) (*admin.APIClient, *handler.ProgressEvent) {
				return mockClient, nil
			}

			event, err := resource.Create(tc.req, nil, tc.currentModel)

			require.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, event.OperationStatus)
		})
	}
}

func TestReadWithMocks(t *testing.T) {
	// Save original function
	originalInitEnv := resource.InitEnvWithLatestClient
	defer func() {
		resource.InitEnvWithLatestClient = originalInitEnv
	}()

	testCases := map[string]struct {
		req            handler.Request
		currentModel   *resource.Model
		mockSetup      func(*mockadmin.CloudBackupsApi)
		expectedStatus handler.Status
	}{
		"successfulRead": {
			req: handler.Request{
				RequestContext: handler.RequestContext{},
			},
			currentModel: createTestModel(),
			mockSetup: func(m *mockadmin.CloudBackupsApi) {
				req := admin.GetCompliancePolicyApiRequest{ApiService: m}
				m.EXPECT().GetCompliancePolicy(mock.Anything, mock.Anything).Return(req)
				policy := createTestPolicy()
				m.EXPECT().GetCompliancePolicyExecute(mock.Anything).Return(policy, &http.Response{StatusCode: 200}, nil)
			},
			expectedStatus: handler.Success,
		},
		"readNotFound": {
			req: handler.Request{
				RequestContext: handler.RequestContext{},
			},
			currentModel: createTestModel(),
			mockSetup: func(m *mockadmin.CloudBackupsApi) {
				req := admin.GetCompliancePolicyApiRequest{ApiService: m}
				m.EXPECT().GetCompliancePolicy(mock.Anything, mock.Anything).Return(req)
				m.EXPECT().GetCompliancePolicyExecute(mock.Anything).Return(nil, &http.Response{StatusCode: 404}, fmt.Errorf("not found"))
			},
			expectedStatus: handler.Failed,
		},
		"apiError": {
			req: handler.Request{
				RequestContext: handler.RequestContext{},
			},
			currentModel: createTestModel(),
			mockSetup: func(m *mockadmin.CloudBackupsApi) {
				req := admin.GetCompliancePolicyApiRequest{ApiService: m}
				m.EXPECT().GetCompliancePolicy(mock.Anything, mock.Anything).Return(req)
				m.EXPECT().GetCompliancePolicyExecute(mock.Anything).Return(nil, &http.Response{StatusCode: 500}, fmt.Errorf("server error"))
			},
			expectedStatus: handler.Failed,
		},
		"readErrorWithNilResponse": {
			req: handler.Request{
				RequestContext: handler.RequestContext{},
			},
			currentModel: createTestModel(),
			mockSetup: func(m *mockadmin.CloudBackupsApi) {
				req := admin.GetCompliancePolicyApiRequest{ApiService: m}
				m.EXPECT().GetCompliancePolicy(mock.Anything, mock.Anything).Return(req)
				m.EXPECT().GetCompliancePolicyExecute(mock.Anything).Return(nil, nil, fmt.Errorf("network error"))
			},
			expectedStatus: handler.Failed,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			mockApi := mockadmin.NewCloudBackupsApi(t)
			tc.mockSetup(mockApi)

			mockClient := &admin.APIClient{}
			mockClient.CloudBackupsApi = mockApi

			resource.InitEnvWithLatestClient = func(req handler.Request, currentModel *resource.Model, requiredFields []string) (*admin.APIClient, *handler.ProgressEvent) {
				return mockClient, nil
			}

			event, err := resource.Read(tc.req, nil, tc.currentModel)

			require.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, event.OperationStatus)
		})
	}
}

func TestUpdateWithMocks(t *testing.T) {
	// Save original function
	originalInitEnv := resource.InitEnvWithLatestClient
	defer func() {
		resource.InitEnvWithLatestClient = originalInitEnv
	}()

	testCases := map[string]struct {
		req            handler.Request
		prevModel      *resource.Model
		currentModel   *resource.Model
		mockSetup      func(*mockadmin.CloudBackupsApi)
		expectedStatus handler.Status
	}{
		"successfulUpdate": {
			req: handler.Request{
				RequestContext: handler.RequestContext{},
			},
			prevModel: createTestModel(),
			currentModel: func() *resource.Model {
				m := createTestModel()
				copyProtection := true
				m.CopyProtectionEnabled = &copyProtection
				return m
			}(),
			mockSetup: func(m *mockadmin.CloudBackupsApi) {
				// Mock GetCompliancePolicy (called first to check if policy exists)
				getReq := admin.GetCompliancePolicyApiRequest{ApiService: m}
				m.EXPECT().GetCompliancePolicy(mock.Anything, mock.Anything).Return(getReq)
				policy := createTestPolicy()
				m.EXPECT().GetCompliancePolicyExecute(mock.Anything).Return(policy, &http.Response{StatusCode: 200}, nil)
				// Mock UpdateCompliancePolicyWithParams
				updateReq := admin.UpdateCompliancePolicyApiRequest{ApiService: m}
				m.EXPECT().UpdateCompliancePolicyWithParams(mock.Anything, mock.Anything).Return(updateReq)
				m.EXPECT().UpdateCompliancePolicyExecute(mock.Anything).Return(nil, &http.Response{StatusCode: 200}, nil)
				// Mock GetCompliancePolicy again (called after update to get complete state)
				m.EXPECT().GetCompliancePolicy(mock.Anything, mock.Anything).Return(getReq)
				m.EXPECT().GetCompliancePolicyExecute(mock.Anything).Return(policy, &http.Response{StatusCode: 200}, nil)
			},
			expectedStatus: handler.Success,
		},
		"updateApiError": {
			req: handler.Request{
				RequestContext: handler.RequestContext{},
			},
			prevModel:    createTestModel(),
			currentModel: createTestModel(),
			mockSetup: func(m *mockadmin.CloudBackupsApi) {
				req := admin.UpdateCompliancePolicyApiRequest{ApiService: m}
				m.EXPECT().UpdateCompliancePolicyWithParams(mock.Anything, mock.Anything).Return(req)
				m.EXPECT().UpdateCompliancePolicyExecute(mock.Anything).Return(nil, &http.Response{StatusCode: 500}, fmt.Errorf("update failed"))
			},
			expectedStatus: handler.Failed,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			mockApi := mockadmin.NewCloudBackupsApi(t)
			tc.mockSetup(mockApi)

			mockClient := &admin.APIClient{}
			mockClient.CloudBackupsApi = mockApi

			resource.InitEnvWithLatestClient = func(req handler.Request, currentModel *resource.Model, requiredFields []string) (*admin.APIClient, *handler.ProgressEvent) {
				return mockClient, nil
			}

			event, err := resource.Update(tc.req, tc.prevModel, tc.currentModel)

			require.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, event.OperationStatus)
		})
	}
}

func TestDeleteWithMocks(t *testing.T) {
	// Save original function
	originalInitEnv := resource.InitEnvWithLatestClient
	defer func() {
		resource.InitEnvWithLatestClient = originalInitEnv
	}()

	testCases := map[string]struct {
		req            handler.Request
		currentModel   *resource.Model
		mockSetup      func(*mockadmin.CloudBackupsApi)
		expectedStatus handler.Status
	}{
		"successfulDelete": {
			req: handler.Request{
				RequestContext: handler.RequestContext{},
			},
			currentModel: createTestModel(),
			mockSetup: func(m *mockadmin.CloudBackupsApi) {
				req := admin.DisableCompliancePolicyApiRequest{ApiService: m}
				m.EXPECT().DisableCompliancePolicy(mock.Anything, mock.Anything).Return(req)
				m.EXPECT().DisableCompliancePolicyExecute(mock.Anything).Return(&http.Response{StatusCode: 200}, nil)
			},
			expectedStatus: handler.Success,
		},
		"deleteWithError": {
			req: handler.Request{
				RequestContext: handler.RequestContext{},
			},
			currentModel: createTestModel(),
			mockSetup: func(m *mockadmin.CloudBackupsApi) {
				req := admin.DisableCompliancePolicyApiRequest{ApiService: m}
				m.EXPECT().DisableCompliancePolicy(mock.Anything, mock.Anything).Return(req)
				m.EXPECT().DisableCompliancePolicyExecute(mock.Anything).Return(&http.Response{StatusCode: 500}, fmt.Errorf("delete failed"))
			},
			expectedStatus: handler.Failed,
		},
		"deleteNotFound": {
			req: handler.Request{
				RequestContext: handler.RequestContext{},
			},
			currentModel: createTestModel(),
			mockSetup: func(m *mockadmin.CloudBackupsApi) {
				req := admin.DisableCompliancePolicyApiRequest{ApiService: m}
				m.EXPECT().DisableCompliancePolicy(mock.Anything, mock.Anything).Return(req)
				m.EXPECT().DisableCompliancePolicyExecute(mock.Anything).Return(&http.Response{StatusCode: 404}, fmt.Errorf("not found"))
			},
			expectedStatus: handler.Failed,
		},
		"deleteErrorWithNilResponse": {
			req: handler.Request{
				RequestContext: handler.RequestContext{},
			},
			currentModel: createTestModel(),
			mockSetup: func(m *mockadmin.CloudBackupsApi) {
				req := admin.DisableCompliancePolicyApiRequest{ApiService: m}
				m.EXPECT().DisableCompliancePolicy(mock.Anything, mock.Anything).Return(req)
				m.EXPECT().DisableCompliancePolicyExecute(mock.Anything).Return(nil, fmt.Errorf("network error"))
			},
			expectedStatus: handler.Failed,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			mockApi := mockadmin.NewCloudBackupsApi(t)
			tc.mockSetup(mockApi)

			mockClient := &admin.APIClient{}
			mockClient.CloudBackupsApi = mockApi

			resource.InitEnvWithLatestClient = func(req handler.Request, currentModel *resource.Model, requiredFields []string) (*admin.APIClient, *handler.ProgressEvent) {
				return mockClient, nil
			}

			event, err := resource.Delete(tc.req, nil, tc.currentModel)

			require.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, event.OperationStatus)
		})
	}
}

func TestListWithMocks(t *testing.T) {
	// Save original function
	originalInitEnv := resource.InitEnvWithLatestClient
	defer func() {
		resource.InitEnvWithLatestClient = originalInitEnv
	}()

	testCases := map[string]struct {
		req            handler.Request
		currentModel   *resource.Model
		mockSetup      func(*mockadmin.CloudBackupsApi)
		expectedStatus handler.Status
		expectedCount  int
	}{
		"successfulList": {
			req: handler.Request{
				RequestContext: handler.RequestContext{},
			},
			currentModel: createTestModel(),
			mockSetup: func(m *mockadmin.CloudBackupsApi) {
				req := admin.GetCompliancePolicyApiRequest{ApiService: m}
				m.EXPECT().GetCompliancePolicy(mock.Anything, mock.Anything).Return(req)
				policy := createTestPolicy()
				m.EXPECT().GetCompliancePolicyExecute(mock.Anything).Return(policy, &http.Response{StatusCode: 200}, nil)
			},
			expectedStatus: handler.Success,
			expectedCount:  1,
		},
		"listWithEmptyResults": {
			req: handler.Request{
				RequestContext: handler.RequestContext{},
			},
			currentModel: createTestModel(),
			mockSetup: func(m *mockadmin.CloudBackupsApi) {
				req := admin.GetCompliancePolicyApiRequest{ApiService: m}
				m.EXPECT().GetCompliancePolicy(mock.Anything, mock.Anything).Return(req)
				m.EXPECT().GetCompliancePolicyExecute(mock.Anything).Return(nil, &http.Response{StatusCode: 404}, fmt.Errorf("not found"))
			},
			expectedStatus: handler.Success,
			expectedCount:  0,
		},
		"listApiError": {
			req: handler.Request{
				RequestContext: handler.RequestContext{},
			},
			currentModel: createTestModel(),
			mockSetup: func(m *mockadmin.CloudBackupsApi) {
				req := admin.GetCompliancePolicyApiRequest{ApiService: m}
				m.EXPECT().GetCompliancePolicy(mock.Anything, mock.Anything).Return(req)
				m.EXPECT().GetCompliancePolicyExecute(mock.Anything).Return(nil, &http.Response{StatusCode: 500}, fmt.Errorf("list failed"))
			},
			expectedStatus: handler.Failed,
			expectedCount:  0,
		},
		"listErrorWithNilResponse": {
			req: handler.Request{
				RequestContext: handler.RequestContext{},
			},
			currentModel: createTestModel(),
			mockSetup: func(m *mockadmin.CloudBackupsApi) {
				req := admin.GetCompliancePolicyApiRequest{ApiService: m}
				m.EXPECT().GetCompliancePolicy(mock.Anything, mock.Anything).Return(req)
				m.EXPECT().GetCompliancePolicyExecute(mock.Anything).Return(nil, nil, fmt.Errorf("network error"))
			},
			expectedStatus: handler.Failed,
			expectedCount:  0,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			mockApi := mockadmin.NewCloudBackupsApi(t)
			tc.mockSetup(mockApi)

			mockClient := &admin.APIClient{}
			mockClient.CloudBackupsApi = mockApi

			resource.InitEnvWithLatestClient = func(req handler.Request, currentModel *resource.Model, requiredFields []string) (*admin.APIClient, *handler.ProgressEvent) {
				return mockClient, nil
			}

			event, err := resource.List(tc.req, nil, tc.currentModel)

			require.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, event.OperationStatus)
			if tc.expectedStatus == handler.Success {
				if tc.expectedCount == 0 {
					assert.Equal(t, 0, len(event.ResourceModels))
				} else {
					assert.Greater(t, len(event.ResourceModels), 0)
				}
			}
		})
	}
}
