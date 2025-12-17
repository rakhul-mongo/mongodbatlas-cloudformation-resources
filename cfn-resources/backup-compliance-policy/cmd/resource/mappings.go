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
	admin20250312010 "go.mongodb.org/atlas-sdk/v20250312010/admin"

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
func GetBackupCompliancePolicyModel(policy *admin20250312010.DataProtectionSettings20231001, currentModel *Model) *Model {
	var model *Model
	if currentModel != nil {
		model = currentModel // Preserve all fields including primary identifier
	} else {
		model = &Model{}
	}

	if policy == nil {
		return model
	}

	// Set required fields (always set from API, matching Terraform behavior)
	if policy.ProjectId != nil {
		model.ProjectId = policy.ProjectId
	}
	authorizedEmail := policy.GetAuthorizedEmail()
	model.AuthorizedEmail = &authorizedEmail
	authorizedUserFirstName := policy.GetAuthorizedUserFirstName()
	model.AuthorizedUserFirstName = &authorizedUserFirstName
	authorizedUserLastName := policy.GetAuthorizedUserLastName()
	model.AuthorizedUserLastName = &authorizedUserLastName

	// Set optional fields (always set from API, matching Terraform behavior)
	if policy.CopyProtectionEnabled != nil {
		model.CopyProtectionEnabled = policy.CopyProtectionEnabled
	}
	if policy.EncryptionAtRestEnabled != nil {
		model.EncryptionAtRestEnabled = policy.EncryptionAtRestEnabled
	}
	// RestoreWindowDays: Terraform always sets this from API (even if 0)
	restoreWindowDays := policy.GetRestoreWindowDays()
	model.RestoreWindowDays = &restoreWindowDays
	if policy.PitEnabled != nil {
		model.PitEnabled = policy.PitEnabled
	}

	// Set computed/read-only fields (always set from API, matching Terraform behavior)
	state := policy.GetState()
	model.State = &state
	if policy.UpdatedDate != nil {
		updatedDateStr := util.TimeToString(*policy.UpdatedDate)
		model.UpdatedDate = &updatedDateStr
	}
	updatedUser := policy.GetUpdatedUser()
	model.UpdatedUser = &updatedUser

	// Map on-demand policy item
	if policy.OnDemandPolicyItem != nil {
		model.OnDemandPolicyItem = flattenOnDemandPolicyItem(policy.OnDemandPolicyItem)
	}

	// Map scheduled policy items
	// Terraform uses GetScheduledPolicyItems() which always returns a slice (empty if nil)
	// This ensures policy items are always set (even if empty), matching Terraform behavior
	items := policy.GetScheduledPolicyItems()
	model.PolicyItemHourly = flattenScheduledPolicyItem(items, Hourly)
	model.PolicyItemDaily = flattenScheduledPolicyItem(items, Daily)
	model.PolicyItemWeekly = flattenScheduledPolicyItems(items, Weekly)
	model.PolicyItemMonthly = flattenScheduledPolicyItems(items, Monthly)
	model.PolicyItemYearly = flattenScheduledPolicyItems(items, Yearly)

	return model
}

// flattenOnDemandPolicyItem converts Atlas API on-demand policy item to CFN model
func flattenOnDemandPolicyItem(item *admin20250312010.BackupComplianceOnDemandPolicyItem) *OnDemandPolicyItem {
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

// flattenScheduledPolicyItem converts Atlas API scheduled policy item to CFN model (single item)
func flattenScheduledPolicyItem(items []admin20250312010.BackupComplianceScheduledPolicyItem, frequencyType string) *ScheduledPolicyItem {
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

// flattenScheduledPolicyItems converts Atlas API scheduled policy items to CFN model (multiple items)
func flattenScheduledPolicyItems(items []admin20250312010.BackupComplianceScheduledPolicyItem, frequencyType string) []ScheduledPolicyItem {
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

// expandDataProtectionSettings converts CFN Model to Atlas API request
func expandDataProtectionSettings(model *Model, projectID string) *admin20250312010.DataProtectionSettings20231001 {
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

	settings := &admin20250312010.DataProtectionSettings20231001{
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
		settings.OnDemandPolicyItem = expandOnDemandPolicyItem(model.OnDemandPolicyItem)
	}

	// Expand scheduled policy items
	var scheduledItems []admin20250312010.BackupComplianceScheduledPolicyItem

	if model.PolicyItemHourly != nil {
		scheduledItems = append(scheduledItems, expandScheduledPolicyItem(model.PolicyItemHourly, Hourly))
	}
	if model.PolicyItemDaily != nil {
		scheduledItems = append(scheduledItems, expandScheduledPolicyItem(model.PolicyItemDaily, Daily))
	}
	if len(model.PolicyItemWeekly) > 0 {
		for _, item := range model.PolicyItemWeekly {
			scheduledItems = append(scheduledItems, expandScheduledPolicyItem(&item, Weekly))
		}
	}
	if len(model.PolicyItemMonthly) > 0 {
		for _, item := range model.PolicyItemMonthly {
			scheduledItems = append(scheduledItems, expandScheduledPolicyItem(&item, Monthly))
		}
	}
	if len(model.PolicyItemYearly) > 0 {
		for _, item := range model.PolicyItemYearly {
			scheduledItems = append(scheduledItems, expandScheduledPolicyItem(&item, Yearly))
		}
	}

	if len(scheduledItems) > 0 {
		settings.ScheduledPolicyItems = &scheduledItems
	}

	return settings
}

// expandOnDemandPolicyItem converts CFN on-demand policy item to Atlas API
func expandOnDemandPolicyItem(item *OnDemandPolicyItem) *admin20250312010.BackupComplianceOnDemandPolicyItem {
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

	return &admin20250312010.BackupComplianceOnDemandPolicyItem{
		Id:                item.Id,
		FrequencyInterval: freqInterval,
		FrequencyType:     "ondemand",
		RetentionUnit:     retentionUnit,
		RetentionValue:    retentionVal,
	}
}

// expandScheduledPolicyItem converts CFN scheduled policy item to Atlas API
func expandScheduledPolicyItem(item *ScheduledPolicyItem, frequencyType string) admin20250312010.BackupComplianceScheduledPolicyItem {
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

	return admin20250312010.BackupComplianceScheduledPolicyItem{
		FrequencyType:     frequencyType,
		FrequencyInterval: freqInterval,
		RetentionUnit:     retentionUnit,
		RetentionValue:    retentionVal,
	}
}
