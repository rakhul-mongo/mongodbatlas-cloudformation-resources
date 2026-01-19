package resource

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/handler"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"go.mongodb.org/atlas-sdk/v20250312012/admin"

	"github.com/mongodb/mongodbatlas-cloudformation-resources/util"
	"github.com/mongodb/mongodbatlas-cloudformation-resources/util/progressevent"
)

func HandleCreate(client *util.MongoDBClient, currentModel *Model) handler.ProgressEvent {
	if currentModel.Protocol == nil || *currentModel.Protocol != ProtocolOIDC {
		return progressevent.GetFailedEventByCode(
			fmt.Sprintf("create is only supported by %s, %s must be imported", ProtocolOIDC, ProtocolSAML),
			string(types.HandlerErrorCodeInvalidRequest),
		)
	}

	federationSettingsID := util.SafeString(currentModel.FederationSettingsId)

	createRequest := ExpandOIDCCreateRequest(currentModel)
	created, res, err := client.AtlasSDK.FederatedAuthenticationApi.CreateIdentityProvider(context.Background(), federationSettingsID, createRequest).Execute()
	if err != nil {
		return progressevent.GetFailedEventByResponse(
			fmt.Sprintf("error creating federation settings identity provider (%s): %s", federationSettingsID, err.Error()),
			res,
		)
	}

	createdID := created.GetId()
	currentModel.IdpId = &createdID
	return HandleRead(client, currentModel)
}

func HandleRead(client *util.MongoDBClient, currentModel *Model) handler.ProgressEvent {
	federationSettingsID := util.SafeString(currentModel.FederationSettingsId)
	idpID := util.SafeString(currentModel.IdpId)

	idp, res, err := client.AtlasSDK.FederatedAuthenticationApi.GetIdentityProvider(context.Background(), federationSettingsID, idpID).Execute()
	if err != nil {
		if res != nil && res.StatusCode == http.StatusNotFound {
			return progressevent.GetFailedEventByCode("Resource not found", string(types.HandlerErrorCodeNotFound))
		}
		return progressevent.GetFailedEventByResponse(
			fmt.Sprintf("error getting federated settings identity provider: %s", err.Error()),
			res,
		)
	}

	model := GetFederatedSettingsIdentityProviderModel(idp, currentModel)
	return handler.ProgressEvent{
		OperationStatus: handler.Success,
		Message:         "Read Complete",
		ResourceModel:   model,
	}
}

func HandleUpdate(client *util.MongoDBClient, prevModel *Model, currentModel *Model) handler.ProgressEvent {
	federationSettingsID := util.SafeString(currentModel.FederationSettingsId)
	idpID := util.SafeString(currentModel.IdpId)

	associatedDomains := currentModel.AssociatedDomains
	if associatedDomains == nil {
		associatedDomains = []string{}
	}
	requestedScopes := currentModel.RequestedScopes
	if requestedScopes == nil {
		requestedScopes = []string{}
	}

	updateReq := &admin.FederationIdentityProviderUpdate{
		AssociatedDomains:          &associatedDomains,
		Audience:                   currentModel.Audience,
		AuthorizationType:          currentModel.AuthorizationType,
		ClientId:                   currentModel.ClientId,
		Description:                currentModel.Description,
		DisplayName:                currentModel.Name,
		GroupsClaim:                currentModel.GroupsClaim,
		IdpType:                    currentModel.IdpType,
		IssuerUri:                  currentModel.IssuerUri,
		Protocol:                   currentModel.Protocol,
		PemFileInfo:                nil,
		RequestBinding:             currentModel.RequestBinding,
		RequestedScopes:            &requestedScopes,
		ResponseSignatureAlgorithm: currentModel.ResponseSignatureAlgorithm,
		SsoDebugEnabled:            currentModel.SsoDebugEnabled,
		SsoUrl:                     currentModel.SsoUrl,
		Status:                     currentModel.Status,
		UserClaim:                  currentModel.UserClaim,
	}

	updated, updRes, err := client.AtlasSDK.FederatedAuthenticationApi.UpdateIdentityProvider(context.Background(), federationSettingsID, idpID, updateReq).Execute()
	if err != nil {
		return progressevent.GetFailedEventByResponse(
			fmt.Sprintf("error updating federation settings identity provider (%s): %s", federationSettingsID, err.Error()),
			updRes,
		)
	}

	model := GetFederatedSettingsIdentityProviderModel(updated, currentModel)
	return handler.ProgressEvent{
		OperationStatus: handler.Success,
		Message:         "Update Complete",
		ResourceModel:   model,
	}
}

func HandleDelete(client *util.MongoDBClient, currentModel *Model) handler.ProgressEvent {
	federationSettingsID := util.SafeString(currentModel.FederationSettingsId)
	idpID := util.SafeString(currentModel.IdpId)

	res, err := client.AtlasSDK.FederatedAuthenticationApi.DeleteIdentityProvider(context.Background(), federationSettingsID, idpID).Execute()
	if err != nil {
		return progressevent.GetFailedEventByResponse(
			fmt.Sprintf("error deleting federation settings identity provider (%s): %s, error: %s", federationSettingsID, idpID, err.Error()),
			res,
		)
	}

	return handler.ProgressEvent{
		OperationStatus: handler.Success,
		Message:         "Delete Complete",
	}
}

func HandleList(client *util.MongoDBClient, currentModel *Model) handler.ProgressEvent {
	federationSettingsID := util.SafeString(currentModel.FederationSettingsId)

	protocols := []string{"SAML", "OIDC"}
	idpTypes := []string{"WORKFORCE", "WORKLOAD"}
	params := &admin.ListIdentityProvidersApiParams{
		FederationSettingsId: federationSettingsID,
		Protocol:             &protocols,
		IdpType:              &idpTypes,
	}
	providers, res, err := client.AtlasSDK.FederatedAuthenticationApi.ListIdentityProvidersWithParams(context.Background(), params).Execute()
	if err != nil {
		return progressevent.GetFailedEventByResponse(
			fmt.Sprintf("error getting federatedSettings Identity Providers assigned (%s): %s", federationSettingsID, err.Error()),
			res,
		)
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
	}
}
