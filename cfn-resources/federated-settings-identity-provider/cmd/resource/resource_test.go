// Copyright 2025 MongoDB Inc
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
	"github.com/mongodb/mongodbatlas-cloudformation-resources/util/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	admin20250312010 "go.mongodb.org/atlas-sdk/v20250312010/admin"
	"go.mongodb.org/atlas-sdk/v20250312010/mockadmin"
)

func createOIDCModel() *Model {
	fedID := "507f1f77bcf86cd799439011"
	protocol := ProtocolOIDC
	name := "oidc-idp"
	issuer := "https://issuer.example.com"
	return &Model{
		Profile:              util.StringPtr("default"),
		FederationSettingsId: &fedID,
		Protocol:             &protocol,
		Name:                 &name,
		IssuerUri:            &issuer,
	}
}

func createSAMLModel() *Model {
	fedID := "507f1f77bcf86cd799439011"
	protocol := ProtocolSAML
	name := "saml-idp"
	issuer := "https://issuer.example.com"
	return &Model{
		Profile:              util.StringPtr("default"),
		FederationSettingsId: &fedID,
		Protocol:             &protocol,
		Name:                 &name,
		IssuerUri:            &issuer,
	}
}

func createAPIIdentityProvider(protocol string) *admin20250312010.FederationIdentityProvider {
	p := protocol
	displayName := "name"
	issuerURI := "https://issuer.example.com"
	return &admin20250312010.FederationIdentityProvider{
		Id:          "idp-1",
		OktaIdpId:   "okta-1",
		Protocol:    &p,
		DisplayName: &displayName,
		IssuerUri:   &issuerURI,
	}
}

func stubInitEnvValidationOnly() func(handler.Request, *Model, []string) (*admin20250312010.APIClient, *handler.ProgressEvent) {
	return func(req handler.Request, currentModel *Model, requiredFields []string) (*admin20250312010.APIClient, *handler.ProgressEvent) {
		normalizeIdForBackCompat(currentModel)
		if pe := validator.ValidateModel(requiredFields, currentModel); pe != nil {
			return nil, pe
		}
		return &admin20250312010.APIClient{}, nil
	}
}

func TestInitEnvWithLatestClientImpl(t *testing.T) {
	orig := newAtlasClient
	defer func() { newAtlasClient = orig }()

	t.Run("validationFails", func(t *testing.T) {
		newAtlasClient = func(req *handler.Request, profileName *string) (*util.MongoDBClient, *handler.ProgressEvent) {
			t.Fatalf("newAtlasClient should not be called on validation failure")
			return nil, nil
		}

		// Missing required FederationSettingsId
		m := &Model{
			Name:      util.StringPtr("n"),
			IssuerUri: util.StringPtr("i"),
			Protocol:  util.StringPtr(ProtocolOIDC),
		}
		client, pe := initEnvWithLatestClientImpl(handler.Request{}, m, CreateRequiredFields)
		assert.Nil(t, client)
		require.NotNil(t, pe)
		assert.Equal(t, handler.Failed, pe.OperationStatus)
		assert.Contains(t, pe.Message, "required")
	})

	t.Run("newAtlasClientReturnsProgressEvent", func(t *testing.T) {
		newAtlasClient = func(req *handler.Request, profileName *string) (*util.MongoDBClient, *handler.ProgressEvent) {
			return nil, &handler.ProgressEvent{OperationStatus: handler.Failed, Message: "boom"}
		}

		m := createOIDCModel()
		client, pe := initEnvWithLatestClientImpl(handler.Request{}, m, CreateRequiredFields)
		assert.Nil(t, client)
		require.NotNil(t, pe)
		assert.Equal(t, handler.Failed, pe.OperationStatus)
		assert.Contains(t, pe.Message, "boom")
	})

	t.Run("successReturnsAtlasSDKClient", func(t *testing.T) {
		newAtlasClient = func(req *handler.Request, profileName *string) (*util.MongoDBClient, *handler.ProgressEvent) {
			return &util.MongoDBClient{AtlasSDK: &admin20250312010.APIClient{}}, nil
		}

		m := createOIDCModel()
		client, pe := initEnvWithLatestClientImpl(handler.Request{}, m, CreateRequiredFields)
		require.Nil(t, pe)
		require.NotNil(t, client)
	})
}

func TestCreateValidationErrors(t *testing.T) {
	originalInitEnv := initEnvWithLatestClient
	defer func() { initEnvWithLatestClient = originalInitEnv }()
	initEnvWithLatestClient = stubInitEnvValidationOnly()

	testCases := map[string]struct {
		currentModel *Model
	}{
		"missingFederationSettingsId": {currentModel: &Model{Name: util.StringPtr("n"), IssuerUri: util.StringPtr("i"), Protocol: util.StringPtr(ProtocolOIDC)}},
		"missingName":                 {currentModel: &Model{FederationSettingsId: util.StringPtr("507f1f77bcf86cd799439011"), IssuerUri: util.StringPtr("i"), Protocol: util.StringPtr(ProtocolOIDC)}},
		"missingIssuerUri":            {currentModel: &Model{FederationSettingsId: util.StringPtr("507f1f77bcf86cd799439011"), Name: util.StringPtr("n"), Protocol: util.StringPtr(ProtocolOIDC)}},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			event, err := Create(handler.Request{}, nil, tc.currentModel)
			require.NoError(t, err)
			assert.Equal(t, handler.Failed, event.OperationStatus)
			assert.Contains(t, event.Message, "required")
		})
	}
}

func TestCreateRejectsSAML(t *testing.T) {
	originalInitEnv := initEnvWithLatestClient
	defer func() { initEnvWithLatestClient = originalInitEnv }()
	initEnvWithLatestClient = stubInitEnvValidationOnly()

	event, err := Create(handler.Request{}, nil, createSAMLModel())
	require.NoError(t, err)
	assert.Equal(t, handler.Failed, event.OperationStatus)
	assert.Contains(t, event.Message, "create is only supported by")
}

func TestHandlersReturnInitProgressEvent(t *testing.T) {
	originalInitEnv := initEnvWithLatestClient
	defer func() { initEnvWithLatestClient = originalInitEnv }()

	initEnvWithLatestClient = func(req handler.Request, currentModel *Model, requiredFields []string) (*admin20250312010.APIClient, *handler.ProgressEvent) {
		return nil, &handler.ProgressEvent{OperationStatus: handler.Failed, Message: "init failed"}
	}

	// Create calls init before protocol check, so this should fail via init.
	evt, err := Create(handler.Request{}, nil, createOIDCModel())
	require.NoError(t, err)
	assert.Equal(t, handler.Failed, evt.OperationStatus)
	assert.Contains(t, evt.Message, "init failed")

	m := createOIDCModel()
	id := "idp-1"
	m.IdpId = &id

	evt, err = Read(handler.Request{}, nil, m)
	require.NoError(t, err)
	assert.Equal(t, handler.Failed, evt.OperationStatus)
	assert.Contains(t, evt.Message, "init failed")

	evt, err = Update(handler.Request{}, createOIDCModel(), m)
	require.NoError(t, err)
	assert.Equal(t, handler.Failed, evt.OperationStatus)
	assert.Contains(t, evt.Message, "init failed")

	evt, err = Delete(handler.Request{}, nil, m)
	require.NoError(t, err)
	assert.Equal(t, handler.Failed, evt.OperationStatus)
	assert.Contains(t, evt.Message, "init failed")

	evt, err = List(handler.Request{}, nil, createOIDCModel())
	require.NoError(t, err)
	assert.Equal(t, handler.Failed, evt.OperationStatus)
	assert.Contains(t, evt.Message, "init failed")
}

func TestCreateReadUpdateDeleteListWithMocks(t *testing.T) {
	originalInitEnv := initEnvWithLatestClient
	defer func() { initEnvWithLatestClient = originalInitEnv }()

	t.Run("createSuccessOIDC", func(t *testing.T) {
		mockApi := mockadmin.NewFederatedAuthenticationApi(t)

		// Create
		createReq := admin20250312010.CreateIdentityProviderApiRequest{ApiService: mockApi}
		mockApi.EXPECT().CreateIdentityProvider(mock.Anything, mock.Anything, mock.Anything).Return(createReq)
		created := &admin20250312010.FederationOidcIdentityProvider{Id: "idp-1"}
		mockApi.EXPECT().CreateIdentityProviderExecute(mock.Anything).Return(created, &http.Response{StatusCode: 200}, nil)

		// Read after create
		readReq := admin20250312010.GetIdentityProviderApiRequest{ApiService: mockApi}
		mockApi.EXPECT().GetIdentityProvider(mock.Anything, mock.Anything, mock.Anything).Return(readReq)
		mockApi.EXPECT().GetIdentityProviderExecute(mock.Anything).Return(createAPIIdentityProvider(ProtocolOIDC), &http.Response{StatusCode: 200}, nil)

		mockClient := &admin20250312010.APIClient{}
		mockClient.FederatedAuthenticationApi = mockApi
		initEnvWithLatestClient = func(req handler.Request, currentModel *Model, requiredFields []string) (*admin20250312010.APIClient, *handler.ProgressEvent) {
			return mockClient, nil
		}

		event, err := Create(handler.Request{}, nil, createOIDCModel())
		require.NoError(t, err)
		assert.Equal(t, handler.Success, event.OperationStatus)
	})

	t.Run("readNotFound", func(t *testing.T) {
		mockApi := mockadmin.NewFederatedAuthenticationApi(t)
		readReq := admin20250312010.GetIdentityProviderApiRequest{ApiService: mockApi}
		mockApi.EXPECT().GetIdentityProvider(mock.Anything, mock.Anything, mock.Anything).Return(readReq)
		mockApi.EXPECT().GetIdentityProviderExecute(mock.Anything).Return(nil, &http.Response{StatusCode: 404}, fmt.Errorf("not found"))

		mockClient := &admin20250312010.APIClient{}
		mockClient.FederatedAuthenticationApi = mockApi
		initEnvWithLatestClient = func(req handler.Request, currentModel *Model, requiredFields []string) (*admin20250312010.APIClient, *handler.ProgressEvent) {
			return mockClient, nil
		}

		m := createOIDCModel()
		id := "idp-1"
		m.IdpId = &id
		event, err := Read(handler.Request{}, nil, m)
		require.NoError(t, err)
		assert.Equal(t, handler.Failed, event.OperationStatus)
	})

	t.Run("updateSuccessAndClearGroupsClaim", func(t *testing.T) {
		mockApi := mockadmin.NewFederatedAuthenticationApi(t)

		// Get existing
		getReq := admin20250312010.GetIdentityProviderApiRequest{ApiService: mockApi}
		mockApi.EXPECT().GetIdentityProvider(mock.Anything, mock.Anything, mock.Anything).Return(getReq)
		existing := createAPIIdentityProvider(ProtocolOIDC)
		mockApi.EXPECT().GetIdentityProviderExecute(mock.Anything).Return(existing, &http.Response{StatusCode: 200}, nil)

		// Update
		updReq := admin20250312010.UpdateIdentityProviderApiRequest{ApiService: mockApi}
		mockApi.EXPECT().UpdateIdentityProvider(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(updReq)
		mockApi.EXPECT().UpdateIdentityProviderExecute(mock.Anything).Return(existing, &http.Response{StatusCode: 200}, nil)

		mockClient := &admin20250312010.APIClient{}
		mockClient.FederatedAuthenticationApi = mockApi
		initEnvWithLatestClient = func(req handler.Request, currentModel *Model, requiredFields []string) (*admin20250312010.APIClient, *handler.ProgressEvent) {
			return mockClient, nil
		}

		prev := createOIDCModel()
		curr := createOIDCModel()
		id := "idp-1"
		curr.IdpId = &id
		empty := ""
		curr.GroupsClaim = &empty // should clear to nil in request

		event, err := Update(handler.Request{}, prev, curr)
		require.NoError(t, err)
		assert.Equal(t, handler.Success, event.OperationStatus)
	})

	t.Run("updateCoversAllDiffBranches", func(t *testing.T) {
		mockApi := mockadmin.NewFederatedAuthenticationApi(t)

		// Get existing
		getReq := admin20250312010.GetIdentityProviderApiRequest{ApiService: mockApi}
		mockApi.EXPECT().GetIdentityProvider(mock.Anything, mock.Anything, mock.Anything).Return(getReq)
		existing := createAPIIdentityProvider(ProtocolOIDC)
		mockApi.EXPECT().GetIdentityProviderExecute(mock.Anything).Return(existing, &http.Response{StatusCode: 200}, nil)

		// Update (assert request contains our updated fields)
		updReq := admin20250312010.UpdateIdentityProviderApiRequest{ApiService: mockApi}
		mockApi.EXPECT().
			UpdateIdentityProvider(mock.Anything, mock.Anything, mock.Anything, mock.MatchedBy(func(u *admin20250312010.FederationIdentityProviderUpdate) bool {
				if u == nil {
					return false
				}
				// Spot-check several fields to ensure branches executed
				if u.Protocol == nil || *u.Protocol != ProtocolOIDC {
					return false
				}
				if u.SsoDebugEnabled == nil || *u.SsoDebugEnabled != true {
					return false
				}
				if u.AssociatedDomains == nil || len(*u.AssociatedDomains) != 0 {
					return false
				}
				if u.RequestedScopes == nil || len(*u.RequestedScopes) != 0 {
					return false
				}
				if u.DisplayName == nil || *u.DisplayName != "new-name" {
					return false
				}
				if u.Status == nil || *u.Status != "ACTIVE" {
					return false
				}
				if u.IssuerUri == nil || *u.IssuerUri != "new-issuer" {
					return false
				}
				if u.RequestBinding == nil || *u.RequestBinding != "HTTP-POST" {
					return false
				}
				if u.ResponseSignatureAlgorithm == nil || *u.ResponseSignatureAlgorithm != "RSA-SHA256" {
					return false
				}
				if u.SsoUrl == nil || *u.SsoUrl != "https://sso" {
					return false
				}
				if u.Audience == nil || *u.Audience != "aud" {
					return false
				}
				if u.ClientId == nil || *u.ClientId != "client" {
					return false
				}
				if u.Description == nil || *u.Description != "desc" {
					return false
				}
				if u.AuthorizationType == nil || *u.AuthorizationType != "authz" {
					return false
				}
				if u.IdpType == nil || *u.IdpType != "WORKFORCE" {
					return false
				}
				if u.UserClaim == nil || *u.UserClaim != "sub" {
					return false
				}
				return true
			})).
			Return(updReq)
		mockApi.EXPECT().UpdateIdentityProviderExecute(mock.Anything).Return(existing, &http.Response{StatusCode: 200}, nil)

		mockClient := &admin20250312010.APIClient{}
		mockClient.FederatedAuthenticationApi = mockApi
		initEnvWithLatestClient = func(req handler.Request, currentModel *Model, requiredFields []string) (*admin20250312010.APIClient, *handler.ProgressEvent) {
			return mockClient, nil
		}

		prev := createOIDCModel()
		prev.AssociatedDomains = []string{"old.example.com"}
		prev.RequestedScopes = []string{"openid"}
		curr := createOIDCModel()
		id := "idp-1"
		curr.IdpId = &id

		// Toggle and set many fields to traverse Update branches
		b := true
		curr.SsoDebugEnabled = &b
		curr.AssociatedDomains = nil // should become empty slice
		curr.RequestedScopes = nil   // should become empty slice
		curr.Name = util.StringPtr("new-name")
		curr.Status = util.StringPtr("ACTIVE")
		curr.IssuerUri = util.StringPtr("new-issuer")
		curr.RequestBinding = util.StringPtr("HTTP-POST")
		curr.ResponseSignatureAlgorithm = util.StringPtr("RSA-SHA256")
		curr.SsoUrl = util.StringPtr("https://sso")
		curr.Audience = util.StringPtr("aud")
		curr.ClientId = util.StringPtr("client")
		curr.Description = util.StringPtr("desc")
		curr.AuthorizationType = util.StringPtr("authz")
		curr.IdpType = util.StringPtr("WORKFORCE")
		curr.UserClaim = util.StringPtr("sub")

		event, err := Update(handler.Request{}, prev, curr)
		require.NoError(t, err)
		assert.Equal(t, handler.Success, event.OperationStatus)
	})

	t.Run("deleteNotFoundIsSuccess", func(t *testing.T) {
		mockApi := mockadmin.NewFederatedAuthenticationApi(t)
		delReq := admin20250312010.DeleteIdentityProviderApiRequest{ApiService: mockApi}
		mockApi.EXPECT().DeleteIdentityProvider(mock.Anything, mock.Anything, mock.Anything).Return(delReq)
		mockApi.EXPECT().DeleteIdentityProviderExecute(mock.Anything).Return(&http.Response{StatusCode: 404}, fmt.Errorf("not found"))

		mockClient := &admin20250312010.APIClient{}
		mockClient.FederatedAuthenticationApi = mockApi
		initEnvWithLatestClient = func(req handler.Request, currentModel *Model, requiredFields []string) (*admin20250312010.APIClient, *handler.ProgressEvent) {
			return mockClient, nil
		}

		m := createOIDCModel()
		id := "idp-1"
		m.IdpId = &id
		event, err := Delete(handler.Request{}, nil, m)
		require.NoError(t, err)
		assert.Equal(t, handler.Success, event.OperationStatus)
	})

	t.Run("listSuccess", func(t *testing.T) {
		mockApi := mockadmin.NewFederatedAuthenticationApi(t)
		listReq := admin20250312010.ListIdentityProvidersApiRequest{ApiService: mockApi}
		mockApi.EXPECT().ListIdentityProvidersWithParams(mock.Anything, mock.Anything).Return(listReq)
		results := []admin20250312010.FederationIdentityProvider{*createAPIIdentityProvider(ProtocolOIDC)}
		paginated := &admin20250312010.PaginatedFederationIdentityProvider{Results: &results}
		mockApi.EXPECT().ListIdentityProvidersExecute(mock.Anything).Return(paginated, &http.Response{StatusCode: 200}, nil)

		mockClient := &admin20250312010.APIClient{}
		mockClient.FederatedAuthenticationApi = mockApi
		initEnvWithLatestClient = func(req handler.Request, currentModel *Model, requiredFields []string) (*admin20250312010.APIClient, *handler.ProgressEvent) {
			return mockClient, nil
		}

		event, err := List(handler.Request{}, nil, createOIDCModel())
		require.NoError(t, err)
		assert.Equal(t, handler.Success, event.OperationStatus)
		assert.Greater(t, len(event.ResourceModels), 0)
	})
}
