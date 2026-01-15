// Copyright 2026 MongoDB Inc
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
	"strconv"

	"go.mongodb.org/atlas-sdk/v20250312012/admin"

	"github.com/mongodb/mongodbatlas-cloudformation-resources/util"
)

const (
	Hourly  = "hourly"
	Daily   = "daily"
	Weekly  = "weekly"
	Monthly = "monthly"
	Yearly  = "yearly"
)

func boolToStringPtr(value bool) *string {
	return util.StringPtr(strconv.FormatBool(value))
}

func stringPtrToBool(value *string) bool {
	if value == nil {
		return false
	}
	result, _ := strconv.ParseBool(*value)
	return result
}

func stringPtrToInt(value *string, defaultValue int) int {
	result := util.StrPtrToIntPtr(value)
	if result == nil {
		return defaultValue
	}
	return *result
}

func isOnDemandPolicyItemEmpty(item *OnDemandPolicyItem) bool {
	return item == nil || (item.FrequencyInterval == nil && item.RetentionUnit == nil && item.RetentionValue == nil)
}

func isScheduledPolicyItemEmpty(item *ScheduledPolicyItem) bool {
	return item == nil || (item.FrequencyInterval == nil && item.RetentionUnit == nil && item.RetentionValue == nil)
}

func setBackupCompliancePolicyData(currentModel *Model, policy *admin.DataProtectionSettings20231001) {
	if policy == nil {
		return
	}

	if policy.ProjectId != nil {
		currentModel.ProjectId = policy.ProjectId
	}

	authorizedEmail := policy.GetAuthorizedEmail()
	currentModel.AuthorizedEmail = &authorizedEmail
	authorizedUserFirstName := policy.GetAuthorizedUserFirstName()
	currentModel.AuthorizedUserFirstName = &authorizedUserFirstName
	authorizedUserLastName := policy.GetAuthorizedUserLastName()
	currentModel.AuthorizedUserLastName = &authorizedUserLastName

	if policy.CopyProtectionEnabled != nil {
		currentModel.CopyProtectionEnabled = boolToStringPtr(*policy.CopyProtectionEnabled)
	}
	if policy.EncryptionAtRestEnabled != nil {
		currentModel.EncryptionAtRestEnabled = boolToStringPtr(*policy.EncryptionAtRestEnabled)
	}
	if policy.PitEnabled != nil {
		currentModel.PitEnabled = boolToStringPtr(*policy.PitEnabled)
	}
	if policy.RestoreWindowDays != nil {
		currentModel.RestoreWindowDays = util.IntPtrToStrPtr(policy.RestoreWindowDays)
	}

	if policy.OnDemandPolicyItem != nil {
		currentModel.OnDemandPolicyItem = getOnDemandPolicyItem(policy.OnDemandPolicyItem)
	}

	if policy.ScheduledPolicyItems != nil {
		currentModel.PolicyItemHourly = nil
		currentModel.PolicyItemDaily = nil
		currentModel.PolicyItemWeekly = []ScheduledPolicyItem{}
		currentModel.PolicyItemMonthly = []ScheduledPolicyItem{}
		currentModel.PolicyItemYearly = []ScheduledPolicyItem{}

		for _, item := range *policy.ScheduledPolicyItems {
			frequencyType := item.GetFrequencyType()
			scheduledItem := getScheduledPolicyItem(&item)
			switch frequencyType {
			case Hourly:
				currentModel.PolicyItemHourly = scheduledItem
			case Daily:
				currentModel.PolicyItemDaily = scheduledItem
			case Weekly:
				currentModel.PolicyItemWeekly = append(currentModel.PolicyItemWeekly, *scheduledItem)
			case Monthly:
				currentModel.PolicyItemMonthly = append(currentModel.PolicyItemMonthly, *scheduledItem)
			case Yearly:
				currentModel.PolicyItemYearly = append(currentModel.PolicyItemYearly, *scheduledItem)
			}
		}
	}

	state := policy.GetState()
	currentModel.State = &state
	if policy.UpdatedDate != nil {
		updatedDateStr := util.TimeToString(*policy.UpdatedDate)
		currentModel.UpdatedDate = &updatedDateStr
	}
	updatedUser := policy.GetUpdatedUser()
	currentModel.UpdatedUser = &updatedUser
}

func getOnDemandPolicyItem(item *admin.BackupComplianceOnDemandPolicyItem) *OnDemandPolicyItem {
	if item == nil {
		return nil
	}
	frequencyType := item.GetFrequencyType()
	retentionUnit := item.GetRetentionUnit()
	frequencyInterval := item.GetFrequencyInterval()
	retentionValue := item.GetRetentionValue()
	return &OnDemandPolicyItem{
		Id:                item.Id,
		FrequencyInterval: util.IntPtrToStrPtr(&frequencyInterval),
		FrequencyType:     &frequencyType,
		RetentionUnit:     &retentionUnit,
		RetentionValue:    util.IntPtrToStrPtr(&retentionValue),
	}
}

func getScheduledPolicyItem(item *admin.BackupComplianceScheduledPolicyItem) *ScheduledPolicyItem {
	if item == nil {
		return nil
	}
	frequencyType := item.GetFrequencyType()
	retentionUnit := item.GetRetentionUnit()
	frequencyInterval := item.GetFrequencyInterval()
	retentionValue := item.GetRetentionValue()
	return &ScheduledPolicyItem{
		Id:                item.Id,
		FrequencyType:     &frequencyType,
		FrequencyInterval: util.IntPtrToStrPtr(&frequencyInterval),
		RetentionUnit:     &retentionUnit,
		RetentionValue:    util.IntPtrToStrPtr(&retentionValue),
	}
}

func expandDataProtectionSettings(model *Model, projectID string) *admin.DataProtectionSettings20231001 {
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

	copyProtectionEnabled := stringPtrToBool(model.CopyProtectionEnabled)
	encryptionAtRestEnabled := stringPtrToBool(model.EncryptionAtRestEnabled)
	pitEnabled := stringPtrToBool(model.PitEnabled)
	restoreWindowDays := stringPtrToInt(model.RestoreWindowDays, 0)

	settings := &admin.DataProtectionSettings20231001{
		ProjectId:               &projectID,
		AuthorizedEmail:         authorizedEmail,
		AuthorizedUserFirstName: authorizedUserFirstName,
		AuthorizedUserLastName:  authorizedUserLastName,
		CopyProtectionEnabled:   &copyProtectionEnabled,
		EncryptionAtRestEnabled: &encryptionAtRestEnabled,
		PitEnabled:              &pitEnabled,
		RestoreWindowDays:       &restoreWindowDays,
	}

	if !isOnDemandPolicyItemEmpty(model.OnDemandPolicyItem) {
		settings.OnDemandPolicyItem = expandOnDemandPolicyItem(model.OnDemandPolicyItem)
	}

	var scheduledItems []admin.BackupComplianceScheduledPolicyItem

	if !isScheduledPolicyItemEmpty(model.PolicyItemHourly) {
		scheduledItems = append(scheduledItems, expandScheduledPolicyItem(model.PolicyItemHourly, Hourly))
	}
	if !isScheduledPolicyItemEmpty(model.PolicyItemDaily) {
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

func expandOnDemandPolicyItem(item *OnDemandPolicyItem) *admin.BackupComplianceOnDemandPolicyItem {
	if item == nil {
		return nil
	}
	frequencyInterval := stringPtrToInt(item.FrequencyInterval, 0)
	retentionValue := stringPtrToInt(item.RetentionValue, 0)
	retentionUnit := ""
	if item.RetentionUnit != nil {
		retentionUnit = *item.RetentionUnit
	}

	return &admin.BackupComplianceOnDemandPolicyItem{
		Id:                item.Id,
		FrequencyInterval: frequencyInterval,
		FrequencyType:     "ondemand",
		RetentionUnit:     retentionUnit,
		RetentionValue:    retentionValue,
	}
}

func expandScheduledPolicyItem(item *ScheduledPolicyItem, frequencyType string) admin.BackupComplianceScheduledPolicyItem {
	frequencyInterval := stringPtrToInt(item.FrequencyInterval, 0)
	retentionValue := stringPtrToInt(item.RetentionValue, 0)
	retentionUnit := ""
	if item.RetentionUnit != nil {
		retentionUnit = *item.RetentionUnit
	}

	return admin.BackupComplianceScheduledPolicyItem{
		FrequencyType:     frequencyType,
		FrequencyInterval: frequencyInterval,
		RetentionUnit:     retentionUnit,
		RetentionValue:    retentionValue,
	}
}
