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
	"strings"

	"go.mongodb.org/atlas-sdk/v20250312010/admin"

	"github.com/aws-cloudformation/cloudformation-cli-go-plugin/cfn/handler"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"

	"github.com/mongodb/mongodbatlas-cloudformation-resources/util"
	"github.com/mongodb/mongodbatlas-cloudformation-resources/util/constants"
	"github.com/mongodb/mongodbatlas-cloudformation-resources/util/logger"
	progress_events "github.com/mongodb/mongodbatlas-cloudformation-resources/util/progressevent"
	"github.com/mongodb/mongodbatlas-cloudformation-resources/util/validator"
)

const (
	defaultCallbackDelaySeconds = 10
	maxDeleteRetries            = 30
	maxCreateRetries            = 30
	callbackKeyDelete           = "callbackBackupCompliancePolicyDelete"
	callbackKeyCreate           = "callbackBackupCompliancePolicyCreate"
	policyStateActive           = "ACTIVE"
	policyStateInactive         = "INACTIVE"
)

var (
	CreateRequiredFields    = []string{constants.ProjectID, constants.AuthorizedEmail, constants.AuthorizedUserFirstName, constants.AuthorizedUserLastName}
	ReadRequiredFields      = []string{constants.ProjectID}
	UpdateRequiredFields    = []string{constants.ProjectID}
	DeleteRequiredFields    = []string{constants.ProjectID}
	ListRequiredFields      = []string{constants.ProjectID}
	InitEnvWithLatestClient = initEnvWithLatestClientImpl
)

// initEnvWithLatestClient is kept for backward compatibility, points to exported version
var initEnvWithLatestClient = InitEnvWithLatestClient

func initEnvWithLatestClientImpl(req handler.Request, currentModel *Model, requiredFields []string) (*admin.APIClient, *handler.ProgressEvent) {
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
	// Check if this is a callback (waiting for policy to become ACTIVE)
	if req.CallbackContext != nil {
		if _, ok := req.CallbackContext[callbackKeyCreate]; ok {
			conn, peErr := InitEnvWithLatestClient(req, currentModel, CreateRequiredFields)
			if peErr != nil {
				return *peErr, nil
			}
			ctx := context.Background()
			projectID := *currentModel.ProjectId
			retryCount := 0
			if rc, ok := req.CallbackContext["createRetryCount"].(float64); ok {
				retryCount = int(rc)
			}
			return handleCreateCallback(ctx, conn, currentModel, projectID, retryCount)
		}
	}

	conn, peErr := InitEnvWithLatestClient(req, currentModel, CreateRequiredFields)
	if peErr != nil {
		return *peErr, nil
	}

	ctx := context.Background()

	projectID := *currentModel.ProjectId

	// Check if backup compliance policy already exists (required by CFN contract tests)
	// If it exists and is in ACTIVE state, return AlreadyExists error
	existingPolicy, apiResp, err := conn.CloudBackupsApi.GetCompliancePolicy(ctx, projectID).Execute()
	if err != nil {
		// If policy doesn't exist (404), proceed with creation
		if apiResp != nil && apiResp.StatusCode == http.StatusNotFound {
			// Policy doesn't exist, proceed with creation
		} else {
			// Error other than Not Found - return it
			return handleError(apiResp, constants.CREATE, err)
		}
	} else if existingPolicy != nil {
		// Policy exists - check if it's in ACTIVE state
		state := existingPolicy.GetState()
		if state == policyStateActive {
			return handler.ProgressEvent{
				OperationStatus:  handler.Failed,
				Message:          "Backup Compliance Policy already exists for project: " + projectID,
				HandlerErrorCode: string(types.HandlerErrorCodeAlreadyExists),
			}, nil
		}
		// If state is not ACTIVE (e.g., INACTIVE), we can proceed with creation
	}
	// Proceed with creation (policy doesn't exist or is not ACTIVE)

	// Expand CFN model to Atlas API request
	dataProtectionSettings := ExpandDataProtectionSettings(currentModel, projectID)

	// Call Atlas API to create/update the backup compliance policy
	params := admin.UpdateCompliancePolicyApiParams{
		GroupId:                        projectID,
		DataProtectionSettings20231001: dataProtectionSettings,
		OverwriteBackupPolicies:        util.Pointer(false),
	}

	_, apiResp, err = conn.CloudBackupsApi.UpdateCompliancePolicyWithParams(ctx, &params).Execute()
	if err != nil {
		// Check if error is due to pending action
		errStr := err.Error()
		if strings.Contains(errStr, "CANNOT_UPDATE_BACKUP_COMPLIANCE_POLICY_SETTINGS_WITH_PENDING_ACTION") ||
			strings.Contains(errStr, "pending action") {
			// Policy has pending actions, return InProgress to retry via callback
			return handler.ProgressEvent{
				OperationStatus:      handler.InProgress,
				Message:              "Waiting for pending actions to complete before creation",
				ResourceModel:        currentModel,
				CallbackDelaySeconds: defaultCallbackDelaySeconds,
				CallbackContext: map[string]any{
					callbackKeyCreate:  true,
					"projectID":        projectID,
					"createRetryCount": 1,
				},
			}, nil
		}
		return handleError(apiResp, constants.CREATE, err)
	}

	// Read back the policy to check its state
	policy, apiResp, err := conn.CloudBackupsApi.GetCompliancePolicy(ctx, projectID).Execute()
	if err != nil {
		return handleError(apiResp, constants.CREATE, err)
	}

	state := policy.GetState()
	if state == policyStateActive {
		// Policy is ACTIVE, creation is complete
		model := GetBackupCompliancePolicyModel(policy, currentModel)
		return handler.ProgressEvent{
			OperationStatus: handler.Success,
			Message:         "Create Completed",
			ResourceModel:   model,
		}, nil
	}

	// Policy is not yet ACTIVE, wait via callback
	return handler.ProgressEvent{
		OperationStatus:      handler.InProgress,
		Message:              fmt.Sprintf("Waiting for policy to become ACTIVE (current state: %s)", state),
		ResourceModel:        currentModel,
		CallbackDelaySeconds: defaultCallbackDelaySeconds,
		CallbackContext: map[string]any{
			callbackKeyCreate:  true,
			"projectID":        projectID,
			"createRetryCount": 1,
		},
	}, nil
}

// handleCreateCallback handles the callback retry for create operation
func handleCreateCallback(ctx context.Context, conn *admin.APIClient, currentModel *Model, projectID string, retryCount int) (handler.ProgressEvent, error) {
	// Check if we've exceeded max retries
	if retryCount >= maxCreateRetries {
		return handler.ProgressEvent{
			OperationStatus:  handler.Failed,
			Message:          "Create failed: Backup Compliance Policy did not become ACTIVE within timeout period",
			HandlerErrorCode: string(types.HandlerErrorCodeInternalFailure),
		}, nil
	}

	// Read the policy to check its current state
	policy, apiResp, err := conn.CloudBackupsApi.GetCompliancePolicy(ctx, projectID).Execute()
	if err != nil {
		if apiResp != nil && apiResp.StatusCode == http.StatusNotFound {
			// Policy was deleted, return error
			return handler.ProgressEvent{
				OperationStatus:  handler.Failed,
				Message:          "Backup Compliance Policy was deleted during creation",
				HandlerErrorCode: string(types.HandlerErrorCodeNotFound),
			}, nil
		}
		return handleError(apiResp, constants.CREATE, err)
	}

	state := policy.GetState()
	if state == policyStateActive {
		// Policy is ACTIVE, creation is complete
		model := GetBackupCompliancePolicyModel(policy, currentModel)
		return handler.ProgressEvent{
			OperationStatus: handler.Success,
			Message:         "Create Completed",
			ResourceModel:   model,
		}, nil
	}

	// Policy is still not ACTIVE, continue waiting via callback
	return handler.ProgressEvent{
		OperationStatus:      handler.InProgress,
		Message:              fmt.Sprintf("Waiting for policy to become ACTIVE (current state: %s, attempt %d/%d)", state, retryCount+1, maxCreateRetries),
		ResourceModel:        currentModel,
		CallbackDelaySeconds: defaultCallbackDelaySeconds,
		CallbackContext: map[string]any{
			callbackKeyCreate:  true,
			"projectID":        projectID,
			"createRetryCount": retryCount + 1,
		},
	}, nil
}

// Read handles the Read event from the Cloudformation service.
func Read(req handler.Request, prevModel *Model, currentModel *Model) (handler.ProgressEvent, error) {
	conn, peErr := InitEnvWithLatestClient(req, currentModel, ReadRequiredFields)
	if peErr != nil {
		return *peErr, nil
	}

	ctx := context.Background()

	projectID := *currentModel.ProjectId

	// Call Atlas API to get the backup compliance policy
	policy, apiResp, err := conn.CloudBackupsApi.GetCompliancePolicy(ctx, projectID).Execute()
	if err != nil {
		if apiResp != nil && apiResp.StatusCode == http.StatusNotFound {
			return progress_events.GetFailedEventByCode(
				"Backup Compliance Policy not found for project: "+projectID,
				string(types.HandlerErrorCodeNotFound)), nil
		}
		return handleError(apiResp, constants.READ, err)
	}

	// Map API response to CFN model, preserving currentModel fields (especially primary identifier)
	model := GetBackupCompliancePolicyModel(policy, currentModel)

	return handler.ProgressEvent{
		OperationStatus: handler.Success,
		Message:         "Read Completed",
		ResourceModel:   model,
	}, nil
}

// Update handles the Update event from the Cloudformation service.
func Update(req handler.Request, prevModel *Model, currentModel *Model) (handler.ProgressEvent, error) {
	conn, peErr := InitEnvWithLatestClient(req, currentModel, UpdateRequiredFields)
	if peErr != nil {
		return *peErr, nil
	}

	ctx := context.Background()

	projectID := *currentModel.ProjectId

	// Check if policy exists first (required by CFN contract tests)
	_, apiResp, err := conn.CloudBackupsApi.GetCompliancePolicy(ctx, projectID).Execute()
	if err != nil {
		// If policy doesn't exist, return NotFound error (required by CFN contract tests)
		if apiResp != nil && apiResp.StatusCode == http.StatusNotFound {
			return progress_events.GetFailedEventByCode(
				"Backup Compliance Policy not found for project: "+projectID,
				string(types.HandlerErrorCodeNotFound)), nil
		}
		return handleError(apiResp, constants.UPDATE, err)
	}

	// Expand CFN model to Atlas API request
	dataProtectionSettings := ExpandDataProtectionSettings(currentModel, projectID)

	// Call Atlas API to update the backup compliance policy
	params := admin.UpdateCompliancePolicyApiParams{
		GroupId:                        projectID,
		DataProtectionSettings20231001: dataProtectionSettings,
		OverwriteBackupPolicies:        util.Pointer(false),
	}

	_, apiResp, err = conn.CloudBackupsApi.UpdateCompliancePolicyWithParams(ctx, &params).Execute()
	if err != nil {
		return handleError(apiResp, constants.UPDATE, err)
	}

	// Read back the policy to get complete state (matching Terraform behavior)
	// This ensures all computed fields (State, UpdatedDate, UpdatedUser, etc.) are properly set
	policy, apiResp, err := conn.CloudBackupsApi.GetCompliancePolicy(ctx, projectID).Execute()
	if err != nil {
		return handleError(apiResp, constants.UPDATE, err)
	}

	// Map API response back to CFN model
	model := GetBackupCompliancePolicyModel(policy, currentModel)

	return handler.ProgressEvent{
		OperationStatus: handler.Success,
		Message:         "Update Completed",
		ResourceModel:   model,
	}, nil
}

// Delete handles the Delete event from the Cloudformation service.
func Delete(req handler.Request, prevModel *Model, currentModel *Model) (handler.ProgressEvent, error) {
	// Check if this is a callback (retry after pending action)
	if req.CallbackContext != nil {
		if _, ok := req.CallbackContext[callbackKeyDelete]; ok {
			conn, peErr := InitEnvWithLatestClient(req, currentModel, DeleteRequiredFields)
			if peErr != nil {
				return *peErr, nil
			}
			ctx := context.Background()
			projectID := *currentModel.ProjectId
			retryCount := 0
			if rc, ok := req.CallbackContext["deleteRetryCount"].(float64); ok {
				retryCount = int(rc)
			}
			return handleDeleteCallback(ctx, conn, currentModel, projectID, retryCount)
		}
	}

	conn, peErr := InitEnvWithLatestClient(req, currentModel, DeleteRequiredFields)
	if peErr != nil {
		return *peErr, nil
	}

	ctx := context.Background()
	projectID := *currentModel.ProjectId

	// Initial delete attempt
	apiResp, err := conn.CloudBackupsApi.DisableCompliancePolicy(ctx, projectID).Execute()
	if err == nil {
		// Successfully disabled
		_, _ = logger.Debugf("Successfully disabled backup compliance policy for project: %s", projectID)
		return handler.ProgressEvent{
			OperationStatus: handler.Success,
			Message:         "Delete Completed",
		}, nil
	}

	// Log the error for debugging
	errStr := err.Error()
	_, _ = logger.Debugf("Delete error for project %s: %s", projectID, errStr)
	if apiResp != nil {
		_, _ = logger.Debugf("Response status code: %d", apiResp.StatusCode)
	}

	// Check if error is due to permanent restriction (cannot disable via API)
	if strings.Contains(errStr, "CANNOT_DISABLE_BACKUP_COMPLIANCE_POLICY") {
		// Policy cannot be disabled via API - requires MongoDB support or removing clusters/snapshots
		_, _ = logger.Warnf("Cannot disable backup compliance policy for project %s: Policy requires MongoDB support or removing all clusters & retained snapshots", projectID)
		return handler.ProgressEvent{
			OperationStatus:  handler.Failed,
			Message:          fmt.Sprintf("Cannot disable Backup Compliance Policy for project %s. Disabling BCP requires MongoDB support or removing all clusters & retained snapshots. This can only be done by filing a support ticket.", projectID),
			HandlerErrorCode: string(types.HandlerErrorCodeInvalidRequest),
		}, nil
	}

	// Check if error is due to pending action
	if strings.Contains(errStr, "CANNOT_UPDATE_BACKUP_COMPLIANCE_POLICY_SETTINGS_WITH_PENDING_ACTION") ||
		strings.Contains(errStr, "pending action") {
		// Policy is still being activated, return InProgress to retry via callback
		_, _ = logger.Debugf("Policy has pending actions, will retry via callback for project: %s", projectID)
		// Matching stream-privatelink-endpoint pattern
		return handler.ProgressEvent{
			OperationStatus:      handler.InProgress,
			Message:              "Waiting for pending actions to complete before deletion",
			ResourceModel:        currentModel,
			CallbackDelaySeconds: defaultCallbackDelaySeconds,
			CallbackContext: map[string]any{
				callbackKeyDelete:  true,
				"projectID":        projectID,
				"deleteRetryCount": 1,
			},
		}, nil
	}

	// Check if policy doesn't exist (404) - deletion is already complete (idempotent)
	if apiResp != nil && apiResp.StatusCode == http.StatusNotFound {
		_, _ = logger.Debugf("Policy not found for project %s, deletion already complete", projectID)
		return handler.ProgressEvent{
			OperationStatus: handler.Success,
			Message:         "Delete Completed (policy not found)",
		}, nil
	}

	// Other errors - return them
	_, _ = logger.Warnf("Unexpected delete error for project %s: %s", projectID, errStr)
	return handleError(apiResp, constants.DELETE, err)
}

// handleDeleteCallback handles the callback retry for delete operation
func handleDeleteCallback(ctx context.Context, conn *admin.APIClient, currentModel *Model, projectID string, retryCount int) (handler.ProgressEvent, error) {
	// Check if we've exceeded max retries
	if retryCount >= maxDeleteRetries {
		return handler.ProgressEvent{
			OperationStatus:  handler.Failed,
			Message:          "Delete failed: Backup Compliance Policy has pending actions that did not complete within timeout period",
			HandlerErrorCode: string(types.HandlerErrorCodeInternalFailure),
		}, nil
	}

	// Try to disable the policy again
	apiResp, err := conn.CloudBackupsApi.DisableCompliancePolicy(ctx, projectID).Execute()
	if err == nil {
		// Successfully disabled
		_, _ = logger.Debugf("Successfully disabled backup compliance policy for project: %s (retry %d)", projectID, retryCount)
		return handler.ProgressEvent{
			OperationStatus: handler.Success,
			Message:         "Delete Completed",
		}, nil
	}

	// Log the error for debugging
	errStr := err.Error()
	_, _ = logger.Debugf("Delete callback error for project %s (retry %d): %s", projectID, retryCount, errStr)

	// Check if error is due to permanent restriction (cannot disable via API)
	if strings.Contains(errStr, "CANNOT_DISABLE_BACKUP_COMPLIANCE_POLICY") {
		// Policy cannot be disabled via API - requires MongoDB support or removing clusters/snapshots
		_, _ = logger.Warnf("Cannot disable backup compliance policy for project %s: Policy requires MongoDB support or removing all clusters & retained snapshots", projectID)
		return handler.ProgressEvent{
			OperationStatus:  handler.Failed,
			Message:          fmt.Sprintf("Cannot disable Backup Compliance Policy for project %s. Disabling BCP requires MongoDB support or removing all clusters & retained snapshots. This can only be done by filing a support ticket.", projectID),
			HandlerErrorCode: string(types.HandlerErrorCodeInvalidRequest),
		}, nil
	}

	// Check if error is still due to pending action
	if strings.Contains(errStr, "CANNOT_UPDATE_BACKUP_COMPLIANCE_POLICY_SETTINGS_WITH_PENDING_ACTION") ||
		strings.Contains(errStr, "pending action") {
		// Still pending, continue retrying via callback
		_, _ = logger.Debugf("Policy still has pending actions, continuing retry (attempt %d/%d) for project: %s", retryCount+1, maxDeleteRetries, projectID)
		// Matching stream-privatelink-endpoint pattern
		return handler.ProgressEvent{
			OperationStatus:      handler.InProgress,
			Message:              fmt.Sprintf("Waiting for pending actions to complete (attempt %d/%d)", retryCount+1, maxDeleteRetries),
			ResourceModel:        currentModel,
			CallbackDelaySeconds: defaultCallbackDelaySeconds,
			CallbackContext: map[string]any{
				callbackKeyDelete:  true,
				"projectID":        projectID,
				"deleteRetryCount": retryCount + 1,
			},
		}, nil
	}

	// Check if policy doesn't exist (404) - deletion is already complete (idempotent)
	if apiResp != nil && apiResp.StatusCode == http.StatusNotFound {
		_, _ = logger.Debugf("Policy not found for project %s, deletion already complete", projectID)
		return handler.ProgressEvent{
			OperationStatus: handler.Success,
			Message:         "Delete Completed (policy not found)",
		}, nil
	}

	// Other errors - return them
	_, _ = logger.Warnf("Unexpected delete callback error for project %s (retry %d): %s", projectID, retryCount, errStr)
	return handleError(apiResp, constants.DELETE, err)
}

// List handles the List event from the Cloudformation service.
func List(req handler.Request, prevModel *Model, currentModel *Model) (handler.ProgressEvent, error) {
	conn, peErr := InitEnvWithLatestClient(req, currentModel, ListRequiredFields)
	if peErr != nil {
		return *peErr, nil
	}

	ctx := context.Background()

	projectID := *currentModel.ProjectId

	// Call Atlas API to get the backup compliance policy
	policy, apiResp, err := conn.CloudBackupsApi.GetCompliancePolicy(ctx, projectID).Execute()
	if err != nil {
		// If not found, return empty list (not an error)
		if apiResp != nil && apiResp.StatusCode == http.StatusNotFound {
			return handler.ProgressEvent{
				OperationStatus: handler.Success,
				Message:         "List Completed",
				ResourceModels:  []interface{}{},
			}, nil
		}
		return handleError(apiResp, constants.LIST, err)
	}

	// Map API response to CFN model
	// For List, we create a new model with ProjectId set
	listModel := &Model{ProjectId: &projectID}
	model := GetBackupCompliancePolicyModel(policy, listModel)

	return handler.ProgressEvent{
		OperationStatus: handler.Success,
		Message:         "List Completed",
		ResourceModels:  []interface{}{model},
	}, nil
}

func handleError(response *http.Response, method constants.CfnFunctions, err error) (handler.ProgressEvent, error) {
	errMsg := fmt.Sprintf("%s error:%s", method, err.Error())
	return progress_events.GetFailedEventByResponse(errMsg, response), nil
}
