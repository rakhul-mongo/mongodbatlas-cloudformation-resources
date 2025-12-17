// Copyright 2023 MongoDB Inc
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
	"net/http"

	admin20250312010 "go.mongodb.org/atlas-sdk/v20250312010/admin"

	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/handler"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"

	"github.com/mongodb/mongodbatlas-cloudformation-resources/util"
	"github.com/mongodb/mongodbatlas-cloudformation-resources/util/constants"
	"github.com/mongodb/mongodbatlas-cloudformation-resources/util/progressevent"
	"github.com/mongodb/mongodbatlas-cloudformation-resources/util/validator"
)

var (
	CreateRequiredFields    = []string{constants.ProjectID, constants.AuthorizedEmail, constants.AuthorizedUserFirstName, constants.AuthorizedUserLastName}
	ReadRequiredFields      = []string{constants.ProjectID}
	UpdateRequiredFields    = []string{constants.ProjectID}
	DeleteRequiredFields    = []string{constants.ProjectID}
	initEnvWithLatestClient = initEnvWithLatestClientImpl
)

func initEnvWithLatestClientImpl(req handler.Request, currentModel *Model, requiredFields []string) (*admin20250312010.APIClient, *handler.ProgressEvent) {
	util.SetupLogger("mongodb-atlas-backup-compliance-policy")

	// Profile is handled via request, not model for this resource
	var profile *string
	util.SetDefaultProfileIfNotDefined(&profile)

	if errEvent := validator.ValidateModel(requiredFields, currentModel); errEvent != nil {
		return nil, errEvent
	}

	client, peErr := util.NewAtlasClient(&req, profile)
	if peErr != nil {
		return nil, peErr
	}
	return client.AtlasSDK, nil
}

// Create handles the Create event from the Cloudformation service.
func Create(req handler.Request, prevModel *Model, currentModel *Model) (handler.ProgressEvent, error) {
	atlasV2, peErr := initEnvWithLatestClient(req, currentModel, CreateRequiredFields)
	if peErr != nil {
		return *peErr, nil
	}

	projectID := *currentModel.ProjectId

	// Expand CFN model to Atlas API request
	dataProtectionSettings := expandDataProtectionSettings(currentModel, projectID)

	// Call Atlas API to create/update the backup compliance policy
	params := admin20250312010.UpdateCompliancePolicyApiParams{
		GroupId:                        projectID,
		DataProtectionSettings20231001: dataProtectionSettings,
		OverwriteBackupPolicies:        util.Pointer(false),
	}

	policy, res, err := atlasV2.CloudBackupsApi.UpdateCompliancePolicyWithParams(context.Background(), &params).Execute()
	if err != nil {
		return progressevent.GetFailedEventByResponse(
			"Failed to create Backup Compliance Policy: "+err.Error(),
			res), nil
	}

	// Map API response back to CFN model
	model := GetBackupCompliancePolicyModel(policy, currentModel)

	return handler.ProgressEvent{
		OperationStatus: handler.Success,
		Message:         "Create Complete",
		ResourceModel:   model,
	}, nil
}

// Read handles the Read event from the Cloudformation service.
func Read(req handler.Request, prevModel *Model, currentModel *Model) (handler.ProgressEvent, error) {
	atlasV2, peErr := initEnvWithLatestClient(req, currentModel, ReadRequiredFields)
	if peErr != nil {
		return *peErr, nil
	}

	projectID := *currentModel.ProjectId

	// Call Atlas API to get the backup compliance policy
	policy, res, err := atlasV2.CloudBackupsApi.GetCompliancePolicy(context.Background(), projectID).Execute()
	if err != nil {
		if res != nil && res.StatusCode == http.StatusNotFound {
			return progressevent.GetFailedEventByCode(
				"Backup Compliance Policy not found for project: "+projectID,
				string(types.HandlerErrorCodeNotFound)), nil
		}
		return progressevent.GetFailedEventByResponse(
			"Failed to read Backup Compliance Policy: "+err.Error(),
			res), nil
	}

	// Map API response to CFN model, preserving currentModel fields (especially primary identifier)
	model := GetBackupCompliancePolicyModel(policy, currentModel)

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

	projectID := *currentModel.ProjectId

	// Expand CFN model to Atlas API request
	dataProtectionSettings := expandDataProtectionSettings(currentModel, projectID)

	// Call Atlas API to update the backup compliance policy
	params := admin20250312010.UpdateCompliancePolicyApiParams{
		GroupId:                        projectID,
		DataProtectionSettings20231001: dataProtectionSettings,
		OverwriteBackupPolicies:        util.Pointer(false),
	}

	policy, res, err := atlasV2.CloudBackupsApi.UpdateCompliancePolicyWithParams(context.Background(), &params).Execute()
	if err != nil {
		return progressevent.GetFailedEventByResponse(
			"Failed to update Backup Compliance Policy: "+err.Error(),
			res), nil
	}

	// Map API response back to CFN model
	model := GetBackupCompliancePolicyModel(policy, currentModel)

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

	projectID := *currentModel.ProjectId

	// Call Atlas API to disable the backup compliance policy
	// Note: Terraform returns an error on 404, but CFN could handle it gracefully for idempotency
	// For exact parity with Terraform, we return an error on 404
	res, err := atlasV2.CloudBackupsApi.DisableCompliancePolicy(context.Background(), projectID).Execute()
	if err != nil {
		return progressevent.GetFailedEventByResponse(
			"Failed to delete Backup Compliance Policy: "+err.Error(),
			res), nil
	}

	return handler.ProgressEvent{
		OperationStatus: handler.Success,
		Message:         "Delete Complete",
	}, nil
}

// List handles the List event from the Cloudformation service.
func List(req handler.Request, prevModel *Model, currentModel *Model) (handler.ProgressEvent, error) {
	atlasV2, peErr := initEnvWithLatestClient(req, currentModel, []string{constants.ProjectID})
	if peErr != nil {
		return *peErr, nil
	}

	projectID := *currentModel.ProjectId

	// Call Atlas API to get the backup compliance policy
	policy, res, err := atlasV2.CloudBackupsApi.GetCompliancePolicy(context.Background(), projectID).Execute()
	if err != nil {
		// If not found, return empty list (not an error)
		if res != nil && res.StatusCode == http.StatusNotFound {
			return handler.ProgressEvent{
				OperationStatus: handler.Success,
				Message:         "List Complete",
				ResourceModels:  []interface{}{},
			}, nil
		}
		return progressevent.GetFailedEventByResponse(
			"Failed to list Backup Compliance Policy: "+err.Error(),
			res), nil
	}

	// Map API response to CFN model
	// For List, we create a new model with ProjectId set
	listModel := &Model{ProjectId: &projectID}
	model := GetBackupCompliancePolicyModel(policy, listModel)

	return handler.ProgressEvent{
		OperationStatus: handler.Success,
		Message:         "List Complete",
		ResourceModels:  []interface{}{model},
	}, nil
}
