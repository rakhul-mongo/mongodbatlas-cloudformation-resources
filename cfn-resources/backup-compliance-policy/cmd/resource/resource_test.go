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

package resource

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/handler"
	"github.com/mongodb/mongodbatlas-cloudformation-resources/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	admin20250312010 "go.mongodb.org/atlas-sdk/v20250312010/admin"
	"go.mongodb.org/atlas-sdk/v20250312010/mockadmin"
)

// Test validation errors
func TestCreateValidationErrors(t *testing.T) {
	testCases := map[string]struct {
		currentModel   *Model
		expectedStatus handler.Status
		expectedMsg    string
	}{
		"missingProjectId": {
			currentModel: &Model{
				AuthorizedEmail:         util.StringPtr("test@example.com"),
				AuthorizedUserFirstName: util.StringPtr("John"),
				AuthorizedUserLastName:  util.StringPtr("Doe"),
			},
			expectedStatus: handler.Failed,
			expectedMsg:    "required",
		},
		"missingAuthorizedEmail": {
			currentModel: &Model{
				ProjectId:               func() *string { s := "507f1f77bcf86cd799439011"; return &s }(),
				AuthorizedUserFirstName: func() *string { s := "John"; return &s }(),
				AuthorizedUserLastName:  func() *string { s := "Doe"; return &s }(),
			},
			expectedStatus: handler.Failed,
			expectedMsg:    "required",
		},
		"missingAuthorizedUserFirstName": {
			currentModel: &Model{
				ProjectId:              func() *string { s := "507f1f77bcf86cd799439011"; return &s }(),
				AuthorizedEmail:        func() *string { s := "test@example.com"; return &s }(),
				AuthorizedUserLastName: func() *string { s := "Doe"; return &s }(),
			},
			expectedStatus: handler.Failed,
			expectedMsg:    "required",
		},
		"missingAuthorizedUserLastName": {
			currentModel: &Model{
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
			event, err := Create(req, nil, tc.currentModel)

			require.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, event.OperationStatus)
			assert.Contains(t, event.Message, tc.expectedMsg)
		})
	}
}

func TestReadValidationErrors(t *testing.T) {
	testCases := map[string]struct {
		currentModel   *Model
		expectedStatus handler.Status
		expectedMsg    string
	}{
		"missingProjectId": {
			currentModel:   &Model{},
			expectedStatus: handler.Failed,
			expectedMsg:    "required",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			req := handler.Request{
				RequestContext: handler.RequestContext{},
			}
			event, err := Read(req, nil, tc.currentModel)

			require.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, event.OperationStatus)
			assert.Contains(t, event.Message, tc.expectedMsg)
		})
	}
}

func TestUpdateValidationErrors(t *testing.T) {
	testCases := map[string]struct {
		currentModel   *Model
		expectedStatus handler.Status
		expectedMsg    string
	}{
		"missingProjectId": {
			currentModel: &Model{
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
			event, err := Update(req, nil, tc.currentModel)

			require.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, event.OperationStatus)
			assert.Contains(t, event.Message, tc.expectedMsg)
		})
	}
}

func TestDeleteValidationErrors(t *testing.T) {
	testCases := map[string]struct {
		currentModel   *Model
		expectedStatus handler.Status
		expectedMsg    string
	}{
		"missingProjectId": {
			currentModel:   &Model{},
			expectedStatus: handler.Failed,
			expectedMsg:    "required",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			req := handler.Request{
				RequestContext: handler.RequestContext{},
			}
			event, err := Delete(req, nil, tc.currentModel)

			require.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, event.OperationStatus)
			assert.Contains(t, event.Message, tc.expectedMsg)
		})
	}
}

func TestListValidationErrors(t *testing.T) {
	testCases := map[string]struct {
		currentModel   *Model
		expectedStatus handler.Status
		expectedMsg    string
	}{
		"missingProjectId": {
			currentModel:   &Model{},
			expectedStatus: handler.Failed,
			expectedMsg:    "required",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			req := handler.Request{
				RequestContext: handler.RequestContext{},
			}
			event, err := List(req, nil, tc.currentModel)

			require.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, event.OperationStatus)
			assert.Contains(t, event.Message, tc.expectedMsg)
		})
	}
}

// Test CRUD operations with mocks
func TestCreateWithMocks(t *testing.T) {
	// Save original function
	originalInitEnv := initEnvWithLatestClient
	defer func() {
		initEnvWithLatestClient = originalInitEnv
	}()

	testCases := map[string]struct {
		req            handler.Request
		currentModel   *Model
		mockSetup      func(*mockadmin.CloudBackupsApi)
		expectedStatus handler.Status
	}{
		"successfulCreate": {
			req: handler.Request{
				RequestContext: handler.RequestContext{},
			},
			currentModel: createTestModel(),
			mockSetup: func(m *mockadmin.CloudBackupsApi) {
				req := admin20250312010.UpdateCompliancePolicyApiRequest{ApiService: m}
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
				req := admin20250312010.UpdateCompliancePolicyApiRequest{ApiService: m}
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

			mockClient := &admin20250312010.APIClient{}
			mockClient.CloudBackupsApi = mockApi

			initEnvWithLatestClient = func(req handler.Request, currentModel *Model, requiredFields []string) (*admin20250312010.APIClient, *handler.ProgressEvent) {
				return mockClient, nil
			}

			event, err := Create(tc.req, nil, tc.currentModel)

			require.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, event.OperationStatus)
		})
	}
}

func TestReadWithMocks(t *testing.T) {
	// Save original function
	originalInitEnv := initEnvWithLatestClient
	defer func() {
		initEnvWithLatestClient = originalInitEnv
	}()

	testCases := map[string]struct {
		req            handler.Request
		currentModel   *Model
		mockSetup      func(*mockadmin.CloudBackupsApi)
		expectedStatus handler.Status
	}{
		"successfulRead": {
			req: handler.Request{
				RequestContext: handler.RequestContext{},
			},
			currentModel: createTestModel(),
			mockSetup: func(m *mockadmin.CloudBackupsApi) {
				req := admin20250312010.GetCompliancePolicyApiRequest{ApiService: m}
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
				req := admin20250312010.GetCompliancePolicyApiRequest{ApiService: m}
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
				req := admin20250312010.GetCompliancePolicyApiRequest{ApiService: m}
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
				req := admin20250312010.GetCompliancePolicyApiRequest{ApiService: m}
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

			mockClient := &admin20250312010.APIClient{}
			mockClient.CloudBackupsApi = mockApi

			initEnvWithLatestClient = func(req handler.Request, currentModel *Model, requiredFields []string) (*admin20250312010.APIClient, *handler.ProgressEvent) {
				return mockClient, nil
			}

			event, err := Read(tc.req, nil, tc.currentModel)

			require.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, event.OperationStatus)
		})
	}
}

func TestUpdateWithMocks(t *testing.T) {
	// Save original function
	originalInitEnv := initEnvWithLatestClient
	defer func() {
		initEnvWithLatestClient = originalInitEnv
	}()

	testCases := map[string]struct {
		req            handler.Request
		prevModel      *Model
		currentModel   *Model
		mockSetup      func(*mockadmin.CloudBackupsApi)
		expectedStatus handler.Status
	}{
		"successfulUpdate": {
			req: handler.Request{
				RequestContext: handler.RequestContext{},
			},
			prevModel: createTestModel(),
			currentModel: func() *Model {
				m := createTestModel()
				copyProtection := true
				m.CopyProtectionEnabled = &copyProtection
				return m
			}(),
			mockSetup: func(m *mockadmin.CloudBackupsApi) {
				req := admin20250312010.UpdateCompliancePolicyApiRequest{ApiService: m}
				m.EXPECT().UpdateCompliancePolicyWithParams(mock.Anything, mock.Anything).Return(req)
				policy := createTestPolicy()
				m.EXPECT().UpdateCompliancePolicyExecute(mock.Anything).Return(policy, &http.Response{StatusCode: 200}, nil)
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
				req := admin20250312010.UpdateCompliancePolicyApiRequest{ApiService: m}
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

			mockClient := &admin20250312010.APIClient{}
			mockClient.CloudBackupsApi = mockApi

			initEnvWithLatestClient = func(req handler.Request, currentModel *Model, requiredFields []string) (*admin20250312010.APIClient, *handler.ProgressEvent) {
				return mockClient, nil
			}

			event, err := Update(tc.req, tc.prevModel, tc.currentModel)

			require.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, event.OperationStatus)
		})
	}
}

func TestDeleteWithMocks(t *testing.T) {
	// Save original function
	originalInitEnv := initEnvWithLatestClient
	defer func() {
		initEnvWithLatestClient = originalInitEnv
	}()

	testCases := map[string]struct {
		req            handler.Request
		currentModel   *Model
		mockSetup      func(*mockadmin.CloudBackupsApi)
		expectedStatus handler.Status
	}{
		"successfulDelete": {
			req: handler.Request{
				RequestContext: handler.RequestContext{},
			},
			currentModel: createTestModel(),
			mockSetup: func(m *mockadmin.CloudBackupsApi) {
				req := admin20250312010.DisableCompliancePolicyApiRequest{ApiService: m}
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
				req := admin20250312010.DisableCompliancePolicyApiRequest{ApiService: m}
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
				req := admin20250312010.DisableCompliancePolicyApiRequest{ApiService: m}
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
				req := admin20250312010.DisableCompliancePolicyApiRequest{ApiService: m}
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

			mockClient := &admin20250312010.APIClient{}
			mockClient.CloudBackupsApi = mockApi

			initEnvWithLatestClient = func(req handler.Request, currentModel *Model, requiredFields []string) (*admin20250312010.APIClient, *handler.ProgressEvent) {
				return mockClient, nil
			}

			event, err := Delete(tc.req, nil, tc.currentModel)

			require.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, event.OperationStatus)
		})
	}
}

func TestListWithMocks(t *testing.T) {
	// Save original function
	originalInitEnv := initEnvWithLatestClient
	defer func() {
		initEnvWithLatestClient = originalInitEnv
	}()

	testCases := map[string]struct {
		req            handler.Request
		currentModel   *Model
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
				req := admin20250312010.GetCompliancePolicyApiRequest{ApiService: m}
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
				req := admin20250312010.GetCompliancePolicyApiRequest{ApiService: m}
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
				req := admin20250312010.GetCompliancePolicyApiRequest{ApiService: m}
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
				req := admin20250312010.GetCompliancePolicyApiRequest{ApiService: m}
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

			mockClient := &admin20250312010.APIClient{}
			mockClient.CloudBackupsApi = mockApi

			initEnvWithLatestClient = func(req handler.Request, currentModel *Model, requiredFields []string) (*admin20250312010.APIClient, *handler.ProgressEvent) {
				return mockClient, nil
			}

			event, err := List(tc.req, nil, tc.currentModel)

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
