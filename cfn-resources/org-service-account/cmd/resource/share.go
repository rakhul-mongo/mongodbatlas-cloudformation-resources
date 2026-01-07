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
	"context"
	"fmt"
	"net/http"

	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/handler"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/mongodb/mongodbatlas-cloudformation-resources/util"
	"github.com/mongodb/mongodbatlas-cloudformation-resources/util/constants"
	progress_events "github.com/mongodb/mongodbatlas-cloudformation-resources/util/progressevent"
)

func HandleCreate(req *handler.Request, client *util.MongoDBClient, model *Model) handler.ProgressEvent {
	ctx := context.Background()
	orgID := model.OrgId
	serviceAccountReq := NewOrgServiceAccountCreateReq(model)

	serviceAccountResp, apiResp, err := client.AtlasSDK.ServiceAccountsApi.CreateOrgServiceAccount(ctx, *orgID, serviceAccountReq).Execute()
	if err != nil {
		return handleError(apiResp, constants.CREATE, err)
	}

	resourceModel := GetOrgServiceAccountModel(serviceAccountResp, model)

	return handler.ProgressEvent{
		OperationStatus: handler.Success,
		Message:         constants.Complete,
		ResourceModel:   resourceModel,
	}
}

func HandleRead(req *handler.Request, client *util.MongoDBClient, model *Model) handler.ProgressEvent {
	ctx := context.Background()
	orgID := model.OrgId
	clientID := model.ClientId

	serviceAccount, apiResp, err := client.AtlasSDK.ServiceAccountsApi.GetOrgServiceAccount(ctx, *orgID, *clientID).Execute()
	if err != nil {
		return handleError(apiResp, constants.READ, err)
	}

	resourceModel := GetOrgServiceAccountModel(serviceAccount, model)
	// Mask secrets on read (writeOnly property)
	if resourceModel.Secrets != nil {
		for i := range resourceModel.Secrets {
			resourceModel.Secrets[i].Secret = nil
		}
	}

	return handler.ProgressEvent{
		OperationStatus: handler.Success,
		Message:         constants.ReadComplete,
		ResourceModel:   resourceModel,
	}
}

func HandleUpdate(req *handler.Request, client *util.MongoDBClient, model *Model) handler.ProgressEvent {
	ctx := context.Background()
	orgID := model.OrgId
	clientID := model.ClientId

	// Verify resource exists
	_, apiResp, err := client.AtlasSDK.ServiceAccountsApi.GetOrgServiceAccount(ctx, *orgID, *clientID).Execute()
	if err != nil {
		// Check if resource doesn't exist (404)
		if apiResp != nil && apiResp.StatusCode == http.StatusNotFound {
			return handler.ProgressEvent{
				OperationStatus:  handler.Failed,
				Message:          "Resource not found",
				HandlerErrorCode: string(types.HandlerErrorCodeNotFound),
			}
		}
		// Other errors
		return handleError(apiResp, constants.UPDATE, err)
	}

	serviceAccountReq := NewOrgServiceAccountUpdateReq(model)
	serviceAccountResp, apiResp, err := client.AtlasSDK.ServiceAccountsApi.UpdateOrgServiceAccount(ctx, *clientID, *orgID, serviceAccountReq).Execute()
	if err != nil {
		return handleError(apiResp, constants.UPDATE, err)
	}

	// GetOrgServiceAccountModel will preserve roles order from input model if available
	resourceModel := GetOrgServiceAccountModel(serviceAccountResp, model)
	// Mask secrets on update response (writeOnly property)
	if resourceModel.Secrets != nil {
		for i := range resourceModel.Secrets {
			resourceModel.Secrets[i].Secret = nil
		}
	}

	return handler.ProgressEvent{
		OperationStatus: handler.Success,
		Message:         constants.Complete,
		ResourceModel:   resourceModel,
	}
}

func HandleDelete(req *handler.Request, client *util.MongoDBClient, model *Model) handler.ProgressEvent {
	ctx := context.Background()
	orgID := model.OrgId
	clientID := model.ClientId

	// Check if resource exists before deleting (contract test requirement)
	_, resp, err := client.AtlasSDK.ServiceAccountsApi.GetOrgServiceAccount(ctx, *orgID, *clientID).Execute()
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			// Resource doesn't exist - return FAILED with NotFound (contract test requirement)
			return handler.ProgressEvent{
				OperationStatus:  handler.Failed,
				Message:          "Resource not found",
				HandlerErrorCode: string(types.HandlerErrorCodeNotFound),
			}
		}
		return handleError(resp, constants.DELETE, err)
	}

	// Resource exists, proceed with delete
	apiResp, err := client.AtlasSDK.ServiceAccountsApi.DeleteOrgServiceAccount(ctx, *clientID, *orgID).Execute()
	if err != nil {
		return handleError(apiResp, constants.DELETE, err)
	}

	return handler.ProgressEvent{
		OperationStatus: handler.Success,
		Message:         constants.Complete,
	}
}

func HandleList(req *handler.Request, client *util.MongoDBClient, model *Model) handler.ProgressEvent {
	ctx := context.Background()
	orgID := model.OrgId

	serviceAccounts, apiResp, err := client.AtlasSDK.ServiceAccountsApi.ListOrgServiceAccounts(ctx, *orgID).Execute()
	if err != nil {
		return handleError(apiResp, constants.LIST, err)
	}

	response := make([]interface{}, 0)
	if serviceAccounts != nil && serviceAccounts.Results != nil {
		for i := range *serviceAccounts.Results {
			itemModel := &Model{}
			resourceModel := GetOrgServiceAccountModel(&(*serviceAccounts.Results)[i], itemModel)
			resourceModel.OrgId = model.OrgId
			resourceModel.Profile = model.Profile
			// Mask secrets in list response (writeOnly property)
			if resourceModel.Secrets != nil {
				for j := range resourceModel.Secrets {
					resourceModel.Secrets[j].Secret = nil
				}
			}
			response = append(response, resourceModel)
		}
	}

	return handler.ProgressEvent{
		OperationStatus: handler.Success,
		Message:         constants.Complete,
		ResourceModels:  response,
	}
}

func handleError(response *http.Response, method constants.CfnFunctions, err error) handler.ProgressEvent {
	errMsg := fmt.Sprintf("%s error:%s", method, err.Error())
	return progress_events.GetFailedEventByResponse(errMsg, response)
}

