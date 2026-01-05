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
	"go.mongodb.org/atlas-sdk/v20250312010/admin"

	"github.com/mongodb/mongodbatlas-cloudformation-resources/util"
)

const (
	Hourly  = "hourly"
	Daily   = "daily"
	Weekly  = "weekly"
	Monthly = "monthly"
	Yearly  = "yearly"
)

// GetBackupCompliancePolicyModel converts an Atlas API response to a CFN Model
// Matching pattern from GetStreamProcessorModel: copy currentModel to preserve input values
func GetBackupCompliancePolicyModel(policy *admin.DataProtectionSettings20231001, currentModel *Model) *Model {
	model := &Model{}

	// Copy all fields from currentModel to preserve input values (matching GetStreamProcessorModel pattern)
	if currentModel != nil {
		*model = *currentModel
	}

	if policy == nil {
		return model
	}

	// Preserve primary identifier (ProjectId) from currentModel if present
	// Only set from API if currentModel doesn't have it (e.g., for List operations)
	if model.ProjectId == nil && policy.ProjectId != nil {
		model.ProjectId = policy.ProjectId
	}

	// Always set required fields from API (these are always returned)
	authorizedEmail := policy.GetAuthorizedEmail()
	model.AuthorizedEmail = &authorizedEmail
	authorizedUserFirstName := policy.GetAuthorizedUserFirstName()
	model.AuthorizedUserFirstName = &authorizedUserFirstName
	authorizedUserLastName := policy.GetAuthorizedUserLastName()
	model.AuthorizedUserLastName = &authorizedUserLastName

	// Set computed/read-only fields (always set from API)
	state := policy.GetState()
	model.State = &state
	if policy.UpdatedDate != nil {
		updatedDateStr := util.TimeToString(*policy.UpdatedDate)
		model.UpdatedDate = &updatedDateStr
	}
	updatedUser := policy.GetUpdatedUser()
	model.UpdatedUser = &updatedUser

	// DO NOT set optional fields from API if they weren't in the input
	// This ensures CREATE operations return the exact input values (CFN contract requirement)
	// Since we copied currentModel above, all input fields are already preserved
	// Only computed/read-only fields (State, UpdatedDate, UpdatedUser) are set from API

	return model
}

// FlattenOnDemandPolicyItem converts Atlas API on-demand policy item to CFN model
func FlattenOnDemandPolicyItem(item *admin.BackupComplianceOnDemandPolicyItem) *OnDemandPolicyItem {
	if item == nil {
		return nil
	}

	freqInterval := int(item.GetFrequencyInterval())
	frequencyType := item.GetFrequencyType()
	retentionUnit := item.GetRetentionUnit()
	retentionVal := int(item.GetRetentionValue())

	return &OnDemandPolicyItem{
		Id:                item.Id,
		FrequencyInterval: &freqInterval,
		FrequencyType:     &frequencyType,
		RetentionUnit:     &retentionUnit,
		RetentionValue:    &retentionVal,
	}
}

// FlattenScheduledPolicyItem converts Atlas API scheduled policy item to CFN model (single item)
func FlattenScheduledPolicyItem(items []admin.BackupComplianceScheduledPolicyItem, frequencyType string) *ScheduledPolicyItem {
	for i := range items {
		item := &items[i]
		// Use direct field access for comparison (matching Terraform behavior)
		if item.FrequencyType == frequencyType {
			freqInterval := int(item.GetFrequencyInterval())
			freqType := item.GetFrequencyType()
			retentionUnit := item.GetRetentionUnit()
			retentionVal := int(item.GetRetentionValue())
			return &ScheduledPolicyItem{
				Id:                item.Id,
				FrequencyType:     &freqType,
				FrequencyInterval: &freqInterval,
				RetentionUnit:     &retentionUnit,
				RetentionValue:    &retentionVal,
			}
		}
	}
	return nil
}

// FlattenScheduledPolicyItems converts Atlas API scheduled policy items to CFN model (multiple items)
func FlattenScheduledPolicyItems(items []admin.BackupComplianceScheduledPolicyItem, frequencyType string) []ScheduledPolicyItem {
	policyItems := make([]ScheduledPolicyItem, 0)
	for i := range items {
		item := &items[i]
		// Use direct field access for comparison (matching Terraform behavior)
		if item.FrequencyType == frequencyType {
			freqInterval := int(item.GetFrequencyInterval())
			freqType := item.GetFrequencyType()
			retentionUnit := item.GetRetentionUnit()
			retentionVal := int(item.GetRetentionValue())
			policyItems = append(policyItems, ScheduledPolicyItem{
				Id:                item.Id,
				FrequencyType:     &freqType,
				FrequencyInterval: &freqInterval,
				RetentionUnit:     &retentionUnit,
				RetentionValue:    &retentionVal,
			})
		}
	}
	return policyItems
}

// ExpandDataProtectionSettings converts CFN Model to Atlas API request
func ExpandDataProtectionSettings(model *Model, projectID string) *admin.DataProtectionSettings20231001 {
	var authorizedEmail string
	if model.AuthorizedEmail != nil {
		authorizedEmail = *model.AuthorizedEmail
	}
	var authorizedUserFirstName string
	if model.AuthorizedUserFirstName != nil {
		authorizedUserFirstName = *model.AuthorizedUserFirstName
	}
	var authorizedUserLastName string
	if model.AuthorizedUserLastName != nil {
		authorizedUserLastName = *model.AuthorizedUserLastName
	}

	settings := &admin.DataProtectionSettings20231001{
		ProjectId:               &projectID,
		AuthorizedEmail:         authorizedEmail,
		AuthorizedUserFirstName: authorizedUserFirstName,
		AuthorizedUserLastName:  authorizedUserLastName,
	}

	// Set optional boolean fields with defaults (matching Terraform behavior)
	// Terraform always sets these fields, defaulting to false if not provided
	copyProtectionEnabled := false
	if model.CopyProtectionEnabled != nil {
		copyProtectionEnabled = *model.CopyProtectionEnabled
	}
	settings.CopyProtectionEnabled = &copyProtectionEnabled

	encryptionAtRestEnabled := false
	if model.EncryptionAtRestEnabled != nil {
		encryptionAtRestEnabled = *model.EncryptionAtRestEnabled
	}
	settings.EncryptionAtRestEnabled = &encryptionAtRestEnabled

	pitEnabled := false
	if model.PitEnabled != nil {
		pitEnabled = *model.PitEnabled
	}
	settings.PitEnabled = &pitEnabled

	// RestoreWindowDays: Terraform always sets this (even if 0) using cast.ToInt
	restoreWindowDays := 0
	if model.RestoreWindowDays != nil {
		restoreWindowDays = *model.RestoreWindowDays
	}
	settings.RestoreWindowDays = &restoreWindowDays

	// Expand on-demand policy item
	if model.OnDemandPolicyItem != nil {
		settings.OnDemandPolicyItem = ExpandOnDemandPolicyItem(model.OnDemandPolicyItem)
	}

	// Expand scheduled policy items
	var scheduledItems []admin.BackupComplianceScheduledPolicyItem

	if model.PolicyItemHourly != nil {
		scheduledItems = append(scheduledItems, ExpandScheduledPolicyItem(model.PolicyItemHourly, Hourly))
	}
	if model.PolicyItemDaily != nil {
		scheduledItems = append(scheduledItems, ExpandScheduledPolicyItem(model.PolicyItemDaily, Daily))
	}
	if len(model.PolicyItemWeekly) > 0 {
		for _, item := range model.PolicyItemWeekly {
			scheduledItems = append(scheduledItems, ExpandScheduledPolicyItem(&item, Weekly))
		}
	}
	if len(model.PolicyItemMonthly) > 0 {
		for _, item := range model.PolicyItemMonthly {
			scheduledItems = append(scheduledItems, ExpandScheduledPolicyItem(&item, Monthly))
		}
	}
	if len(model.PolicyItemYearly) > 0 {
		for _, item := range model.PolicyItemYearly {
			scheduledItems = append(scheduledItems, ExpandScheduledPolicyItem(&item, Yearly))
		}
	}

	if len(scheduledItems) > 0 {
		settings.ScheduledPolicyItems = &scheduledItems
	}

	return settings
}

// ExpandOnDemandPolicyItem converts CFN on-demand policy item to Atlas API
func ExpandOnDemandPolicyItem(item *OnDemandPolicyItem) *admin.BackupComplianceOnDemandPolicyItem {
	if item == nil {
		return nil
	}

	var freqInterval int
	if item.FrequencyInterval != nil {
		freqInterval = *item.FrequencyInterval
	}
	var retentionVal int
	if item.RetentionValue != nil {
		retentionVal = *item.RetentionValue
	}
	var retentionUnit string
	if item.RetentionUnit != nil {
		retentionUnit = *item.RetentionUnit
	}

	return &admin.BackupComplianceOnDemandPolicyItem{
		Id:                item.Id,
		FrequencyInterval: freqInterval,
		FrequencyType:     "ondemand",
		RetentionUnit:     retentionUnit,
		RetentionValue:    retentionVal,
	}
}

// ExpandScheduledPolicyItem converts CFN scheduled policy item to Atlas API
func ExpandScheduledPolicyItem(item *ScheduledPolicyItem, frequencyType string) admin.BackupComplianceScheduledPolicyItem {
	var freqInterval int
	if item.FrequencyInterval != nil {
		freqInterval = *item.FrequencyInterval
	}
	var retentionVal int
	if item.RetentionValue != nil {
		retentionVal = *item.RetentionValue
	}
	var retentionUnit string
	if item.RetentionUnit != nil {
		retentionUnit = *item.RetentionUnit
	}

	return admin.BackupComplianceScheduledPolicyItem{
		FrequencyType:     frequencyType,
		FrequencyInterval: freqInterval,
		RetentionUnit:     retentionUnit,
		RetentionValue:    retentionVal,
	}
}
