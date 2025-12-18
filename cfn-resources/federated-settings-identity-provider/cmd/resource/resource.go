package resource

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/handler"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	admin20250312010 "go.mongodb.org/atlas-sdk/v20250312010/admin"

	"github.com/mongodb/mongodbatlas-cloudformation-resources/util"
	"github.com/mongodb/mongodbatlas-cloudformation-resources/util/constants"
	"github.com/mongodb/mongodbatlas-cloudformation-resources/util/progressevent"
	"github.com/mongodb/mongodbatlas-cloudformation-resources/util/validator"
)

var (
	CreateRequiredFields = []string{constants.FederationSettingsID, constants.Name, constants.IssuerURI}
	ReadRequiredFields   = []string{constants.FederationSettingsID, constants.IdpID}
	UpdateRequiredFields = []string{constants.FederationSettingsID, constants.IdpID}
	DeleteRequiredFields = []string{constants.FederationSettingsID, constants.IdpID}
	ListRequiredFields   = []string{constants.FederationSettingsID}

	// initEnvWithLatestClient is a variable that can be reassigned in tests for mocking
	initEnvWithLatestClient = initEnvWithLatestClientImpl

	// newAtlasClient is a seam for unit tests to avoid real AWS/SecretsManager access.
	newAtlasClient = util.NewAtlasClient
)

func safeBool(b *bool) bool {
	return b != nil && *b
}

func initEnvWithLatestClientImpl(req handler.Request, currentModel *Model, requiredFields []string) (*admin20250312010.APIClient, *handler.ProgressEvent) {
	util.SetupLogger("mongodb-atlas-federated-settings-identity-provider")

	// Default profile if not provided
	util.SetDefaultProfileIfNotDefined(&currentModel.Profile)

	// Backward compatibility: allow OktaIdpId to satisfy IdpId identifier
	normalizeIdForBackCompat(currentModel)

	if errEvent := validator.ValidateModel(requiredFields, currentModel); errEvent != nil {
		return nil, errEvent
	}

	client, pe := newAtlasClient(&req, currentModel.Profile)
	if pe != nil {
		return nil, pe
	}
	return client.AtlasSDK, nil
}

// Create handles the Create event from the Cloudformation service.
func Create(req handler.Request, prevModel *Model, currentModel *Model) (handler.ProgressEvent, error) {
	atlasV2, peErr := initEnvWithLatestClient(req, currentModel, CreateRequiredFields)
	if peErr != nil {
		return *peErr, nil
	}

	// Match Terraform behavior: Create is only supported for OIDC. SAML must be imported/read.
	if currentModel.Protocol == nil || *currentModel.Protocol != ProtocolOIDC {
		return progressevent.GetFailedEventByCode(
			fmt.Sprintf("create is only supported by %s, %s must be imported", ProtocolOIDC, ProtocolSAML),
			string(types.HandlerErrorCodeInvalidRequest),
		), nil
	}

	federationSettingsID := util.SafeString(currentModel.FederationSettingsId)

	createRequest := expandOIDCCreateRequest(currentModel)
	created, res, err := atlasV2.FederatedAuthenticationApi.CreateIdentityProvider(context.Background(), federationSettingsID, createRequest).Execute()
	if err != nil {
		return progressevent.GetFailedEventByResponse(
			fmt.Sprintf("error creating federation settings identity provider (%s): %s", federationSettingsID, err.Error()),
			res,
		), nil
	}

	// Match Terraform: set ID then read back the full resource
	createdID := created.GetId()
	currentModel.IdpId = &createdID
	return Read(req, prevModel, currentModel)
}

// Read handles the Read event from the Cloudformation service.
func Read(req handler.Request, prevModel *Model, currentModel *Model) (handler.ProgressEvent, error) {
	atlasV2, peErr := initEnvWithLatestClient(req, currentModel, ReadRequiredFields)
	if peErr != nil {
		return *peErr, nil
	}

	federationSettingsID := util.SafeString(currentModel.FederationSettingsId)
	idpID := util.SafeString(currentModel.IdpId)

	idp, res, err := atlasV2.FederatedAuthenticationApi.GetIdentityProvider(context.Background(), federationSettingsID, idpID).Execute()
	if err != nil {
		if res != nil && res.StatusCode == http.StatusNotFound {
			return progressevent.GetFailedEventByCode("Resource not found", string(types.HandlerErrorCodeNotFound)), nil
		}
		return progressevent.GetFailedEventByResponse(
			fmt.Sprintf("error getting federated settings identity provider: %s", err.Error()),
			res,
		), nil
	}

	model := GetFederatedSettingsIdentityProviderModel(idp, currentModel)
	return handler.ProgressEvent{
		OperationStatus: handler.Success,
		Message:         "Read Complete",
		ResourceModel:   model,
	}, nil
}

// Update handles the Update event from the Cloudformation service.
func Update(req handler.Request, prevModel *Model, currentModel *Model) (handler.ProgressEvent, error) {
	atlasV2, peErr := initEnvWithLatestClient(req, currentModel, UpdateRequiredFields)
	if peErr != nil {
		return *peErr, nil
	}

	federationSettingsID := util.SafeString(currentModel.FederationSettingsId)
	idpID := util.SafeString(currentModel.IdpId)

	existing, res, err := atlasV2.FederatedAuthenticationApi.GetIdentityProvider(context.Background(), federationSettingsID, idpID).Execute()
	if err != nil {
		return progressevent.GetFailedEventByResponse(
			fmt.Sprintf("error retreiving federation settings identity provider (%s): %s", federationSettingsID, err.Error()),
			res,
		), nil
	}

	updateReq := &admin20250312010.FederationIdentityProviderUpdate{
		AssociatedDomains:          existing.AssociatedDomains,
		Audience:                   existing.Audience,
		AuthorizationType:          existing.AuthorizationType,
		ClientId:                   existing.ClientId,
		Description:                existing.Description,
		DisplayName:                existing.DisplayName,
		GroupsClaim:                existing.GroupsClaim,
		IdpType:                    existing.IdpType,
		IssuerUri:                  existing.IssuerUri,
		Protocol:                   existing.Protocol,
		PemFileInfo:                nil,
		RequestBinding:             existing.RequestBinding,
		RequestedScopes:            existing.RequestedScopes,
		ResponseSignatureAlgorithm: existing.ResponseSignatureAlgorithm,
		SsoDebugEnabled:            existing.SsoDebugEnabled,
		SsoUrl:                     existing.SsoUrl,
		Status:                     existing.Status,
		UserClaim:                  existing.UserClaim,
	}

	if prevModel == nil || util.SafeString(prevModel.Protocol) != util.SafeString(currentModel.Protocol) {
		if currentModel.Protocol != nil {
			updateReq.Protocol = currentModel.Protocol
		}
	}
	if prevModel == nil || safeBool(prevModel.SsoDebugEnabled) != safeBool(currentModel.SsoDebugEnabled) {
		updateReq.SsoDebugEnabled = currentModel.SsoDebugEnabled
	}
	if prevModel == nil || !slicesEqual(prevModel.AssociatedDomains, currentModel.AssociatedDomains) {
		domains := currentModel.AssociatedDomains
		if domains == nil {
			domains = []string{}
		}
		updateReq.AssociatedDomains = &domains
	}
	if prevModel == nil || util.SafeString(prevModel.Name) != util.SafeString(currentModel.Name) {
		updateReq.DisplayName = currentModel.Name
	}
	if prevModel == nil || util.SafeString(prevModel.Status) != util.SafeString(currentModel.Status) {
		updateReq.Status = currentModel.Status
	}
	if prevModel == nil || util.SafeString(prevModel.IssuerUri) != util.SafeString(currentModel.IssuerUri) {
		updateReq.IssuerUri = currentModel.IssuerUri
	}
	if prevModel == nil || util.SafeString(prevModel.RequestBinding) != util.SafeString(currentModel.RequestBinding) {
		updateReq.RequestBinding = currentModel.RequestBinding
	}
	if prevModel == nil || util.SafeString(prevModel.ResponseSignatureAlgorithm) != util.SafeString(currentModel.ResponseSignatureAlgorithm) {
		updateReq.ResponseSignatureAlgorithm = currentModel.ResponseSignatureAlgorithm
	}
	if prevModel == nil || util.SafeString(prevModel.SsoUrl) != util.SafeString(currentModel.SsoUrl) {
		updateReq.SsoUrl = currentModel.SsoUrl
	}
	if prevModel == nil || util.SafeString(prevModel.Audience) != util.SafeString(currentModel.Audience) {
		updateReq.Audience = currentModel.Audience
	}
	if prevModel == nil || util.SafeString(prevModel.ClientId) != util.SafeString(currentModel.ClientId) {
		updateReq.ClientId = currentModel.ClientId
	}
	if prevModel == nil || util.SafeString(prevModel.Description) != util.SafeString(currentModel.Description) {
		updateReq.Description = currentModel.Description
	}
	if prevModel == nil || util.SafeString(prevModel.AuthorizationType) != util.SafeString(currentModel.AuthorizationType) {
		updateReq.AuthorizationType = currentModel.AuthorizationType
	}
	if prevModel == nil || util.SafeString(prevModel.IdpType) != util.SafeString(currentModel.IdpType) {
		updateReq.IdpType = currentModel.IdpType
	}
	if prevModel == nil || util.SafeString(prevModel.GroupsClaim) != util.SafeString(currentModel.GroupsClaim) {
		if currentModel.GroupsClaim != nil && *currentModel.GroupsClaim == "" {
			updateReq.GroupsClaim = nil
		} else {
			updateReq.GroupsClaim = currentModel.GroupsClaim
		}
	}
	if prevModel == nil || !slicesEqual(prevModel.RequestedScopes, currentModel.RequestedScopes) {
		scopes := currentModel.RequestedScopes
		if scopes == nil {
			scopes = []string{}
		}
		updateReq.RequestedScopes = &scopes
	}
	if prevModel == nil || util.SafeString(prevModel.UserClaim) != util.SafeString(currentModel.UserClaim) {
		updateReq.UserClaim = currentModel.UserClaim
	}

	updated, updRes, err := atlasV2.FederatedAuthenticationApi.UpdateIdentityProvider(context.Background(), federationSettingsID, idpID, updateReq).Execute()
	if err != nil {
		return progressevent.GetFailedEventByResponse(
			fmt.Sprintf("error updating federation settings identity provider (%s): %s", federationSettingsID, err.Error()),
			updRes,
		), nil
	}

	model := GetFederatedSettingsIdentityProviderModel(updated, currentModel)
	return handler.ProgressEvent{
		OperationStatus: handler.Success,
		Message:         "Update Complete",
		ResourceModel:   model,
	}, nil
}

// Delete handles the Delete event from the Cloudformation service.
func Delete(req handler.Request, prevModel *Model, currentModel *Model) (handler.ProgressEvent, error) {
	atlasV2, peErr := initEnvWithLatestClient(req, currentModel, DeleteRequiredFields)
	if peErr != nil {
		return *peErr, nil
	}

	federationSettingsID := util.SafeString(currentModel.FederationSettingsId)
	idpID := util.SafeString(currentModel.IdpId)

	res, err := atlasV2.FederatedAuthenticationApi.DeleteIdentityProvider(context.Background(), federationSettingsID, idpID).Execute()
	if err != nil {
		if res != nil && res.StatusCode == http.StatusNotFound {
			return handler.ProgressEvent{OperationStatus: handler.Success, Message: "Delete Complete"}, nil
		}
		return progressevent.GetFailedEventByResponse(
			fmt.Sprintf("error deleting federation settings identity provider (%s): %s, error: %s", federationSettingsID, idpID, err.Error()),
			res,
		), nil
	}

	return handler.ProgressEvent{
		OperationStatus: handler.Success,
		Message:         "Delete Complete",
	}, nil
}

// List handles the List event from the Cloudformation service.
func List(req handler.Request, prevModel *Model, currentModel *Model) (handler.ProgressEvent, error) {
	atlasV2, peErr := initEnvWithLatestClient(req, currentModel, ListRequiredFields)
	if peErr != nil {
		return *peErr, nil
	}

	federationSettingsID := util.SafeString(currentModel.FederationSettingsId)

	params := &admin20250312010.ListIdentityProvidersApiParams{
		FederationSettingsId: federationSettingsID,
	}
	providers, res, err := atlasV2.FederatedAuthenticationApi.ListIdentityProvidersWithParams(context.Background(), params).Execute()
	if err != nil {
		return progressevent.GetFailedEventByResponse(
			fmt.Sprintf("error getting federatedSettings Identity Providers assigned (%s): %s", federationSettingsID, err.Error()),
			res,
		), nil
	}

	results := providers.GetResults()
	models := make([]any, 0, len(results))
	for i := range results {
		m := &Model{
			Profile:              currentModel.Profile,
			FederationSettingsId: currentModel.FederationSettingsId,
		}
		models = append(models, GetFederatedSettingsIdentityProviderModel(&results[i], m))
	}

	return handler.ProgressEvent{
		OperationStatus: handler.Success,
		Message:         "List Complete",
		ResourceModels:  models,
	}, nil
}
