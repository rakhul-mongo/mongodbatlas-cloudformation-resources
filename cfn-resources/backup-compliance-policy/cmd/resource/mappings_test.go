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
	"testing"
	"time"

	"github.com/mongodb/mongodbatlas-cloudformation-resources/util"
	"github.com/stretchr/testify/assert"
	admin20250312010 "go.mongodb.org/atlas-sdk/v20250312010/admin"
)

// Helper function to create a test model
func createTestModel() *Model {
	projectID := "507f1f77bcf86cd799439011"
	authorizedEmail := "test@example.com"
	authorizedUserFirstName := "John"
	authorizedUserLastName := "Doe"

	return &Model{
		ProjectId:               &projectID,
		AuthorizedEmail:         &authorizedEmail,
		AuthorizedUserFirstName: &authorizedUserFirstName,
		AuthorizedUserLastName:  &authorizedUserLastName,
	}
}

// Helper function to create a test API policy response
func createTestPolicy() *admin20250312010.DataProtectionSettings20231001 {
	projectID := "507f1f77bcf86cd799439011"
	authorizedEmail := "test@example.com"
	authorizedUserFirstName := "John"
	authorizedUserLastName := "Doe"
	state := "ACTIVE"
	updatedUser := "user@example.com"
	copyProtectionEnabled := false
	encryptionAtRestEnabled := true
	pitEnabled := false
	restoreWindowDays := 7
	updatedDate := time.Now()

	onDemandId := "507f1f77bcf86cd799439020"
	onDemandItem := admin20250312010.BackupComplianceOnDemandPolicyItem{
		Id:                &onDemandId,
		FrequencyInterval: 1,
		FrequencyType:     "ondemand",
		RetentionUnit:     "days",
		RetentionValue:    30,
	}

	hourlyId := "507f1f77bcf86cd799439021"
	hourlyItem := admin20250312010.BackupComplianceScheduledPolicyItem{
		Id:                &hourlyId,
		FrequencyType:     "hourly",
		FrequencyInterval: 6,
		RetentionUnit:     "days",
		RetentionValue:    7,
	}

	dailyId := "507f1f77bcf86cd799439022"
	dailyItem := admin20250312010.BackupComplianceScheduledPolicyItem{
		Id:                &dailyId,
		FrequencyType:     "daily",
		FrequencyInterval: 1,
		RetentionUnit:     "days",
		RetentionValue:    30,
	}

	weeklyId1 := "507f1f77bcf86cd799439023"
	weeklyItem1 := admin20250312010.BackupComplianceScheduledPolicyItem{
		Id:                &weeklyId1,
		FrequencyType:     "weekly",
		FrequencyInterval: 1,
		RetentionUnit:     "weeks",
		RetentionValue:    4,
	}

	weeklyId2 := "507f1f77bcf86cd799439024"
	weeklyItem2 := admin20250312010.BackupComplianceScheduledPolicyItem{
		Id:                &weeklyId2,
		FrequencyType:     "weekly",
		FrequencyInterval: 2,
		RetentionUnit:     "weeks",
		RetentionValue:    8,
	}

	scheduledItems := []admin20250312010.BackupComplianceScheduledPolicyItem{
		hourlyItem,
		dailyItem,
		weeklyItem1,
		weeklyItem2,
	}

	policy := &admin20250312010.DataProtectionSettings20231001{
		ProjectId:               &projectID,
		AuthorizedEmail:         authorizedEmail,
		AuthorizedUserFirstName: authorizedUserFirstName,
		AuthorizedUserLastName:  authorizedUserLastName,
		State:                   &state,
		UpdatedUser:             &updatedUser,
		UpdatedDate:             &updatedDate,
		CopyProtectionEnabled:   &copyProtectionEnabled,
		EncryptionAtRestEnabled: &encryptionAtRestEnabled,
		PitEnabled:              &pitEnabled,
		RestoreWindowDays:       &restoreWindowDays,
		OnDemandPolicyItem:      &onDemandItem,
		ScheduledPolicyItems:    &scheduledItems,
	}
	return policy
}

func TestFlattenOnDemandPolicyItem(t *testing.T) {
	testCases := map[string]struct {
		item           *admin20250312010.BackupComplianceOnDemandPolicyItem
		expectedResult *OnDemandPolicyItem
	}{
		"withAllFields": {
			item: &admin20250312010.BackupComplianceOnDemandPolicyItem{
				Id:                func() *string { s := "507f1f77bcf86cd799439020"; return &s }(),
				FrequencyInterval: 1,
				FrequencyType:     "ondemand",
				RetentionUnit:     "days",
				RetentionValue:    30,
			},
			expectedResult: &OnDemandPolicyItem{
				Id:                func() *string { s := "507f1f77bcf86cd799439020"; return &s }(),
				FrequencyInterval: func() *int { i := 1; return &i }(),
				FrequencyType:     func() *string { s := "ondemand"; return &s }(),
				RetentionUnit:     func() *string { s := "days"; return &s }(),
				RetentionValue:    func() *int { i := 30; return &i }(),
			},
		},
		"nilItem": {
			item:           nil,
			expectedResult: nil,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			result := flattenOnDemandPolicyItem(tc.item)
			if tc.expectedResult == nil {
				assert.Nil(t, result)
			} else {
				assert.Equal(t, tc.expectedResult.Id, result.Id)
				assert.Equal(t, tc.expectedResult.FrequencyInterval, result.FrequencyInterval)
				assert.Equal(t, tc.expectedResult.FrequencyType, result.FrequencyType)
				assert.Equal(t, tc.expectedResult.RetentionUnit, result.RetentionUnit)
				assert.Equal(t, tc.expectedResult.RetentionValue, result.RetentionValue)
			}
		})
	}
}

func TestFlattenScheduledPolicyItem(t *testing.T) {
	testCases := map[string]struct {
		items          []admin20250312010.BackupComplianceScheduledPolicyItem
		frequencyType  string
		expectedResult *ScheduledPolicyItem
	}{
		"hourlyItem": {
			items: []admin20250312010.BackupComplianceScheduledPolicyItem{
				{
					Id:                func() *string { s := "507f1f77bcf86cd799439021"; return &s }(),
					FrequencyType:     "hourly",
					FrequencyInterval: 6,
					RetentionUnit:     "days",
					RetentionValue:    7,
				},
			},
			frequencyType: "hourly",
			expectedResult: &ScheduledPolicyItem{
				Id:                func() *string { s := "507f1f77bcf86cd799439021"; return &s }(),
				FrequencyType:     func() *string { s := "hourly"; return &s }(),
				FrequencyInterval: func() *int { i := 6; return &i }(),
				RetentionUnit:     func() *string { s := "days"; return &s }(),
				RetentionValue:    func() *int { i := 7; return &i }(),
			},
		},
		"notFound": {
			items: []admin20250312010.BackupComplianceScheduledPolicyItem{
				{
					FrequencyType:     "hourly",
					FrequencyInterval: 6,
				},
			},
			frequencyType:  "daily",
			expectedResult: nil,
		},
		"emptyItems": {
			items:          []admin20250312010.BackupComplianceScheduledPolicyItem{},
			frequencyType:  "hourly",
			expectedResult: nil,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			result := flattenScheduledPolicyItem(tc.items, tc.frequencyType)
			if tc.expectedResult == nil {
				assert.Nil(t, result)
			} else {
				assert.Equal(t, tc.expectedResult.Id, result.Id)
				assert.Equal(t, tc.expectedResult.FrequencyType, result.FrequencyType)
				assert.Equal(t, tc.expectedResult.FrequencyInterval, result.FrequencyInterval)
				assert.Equal(t, tc.expectedResult.RetentionUnit, result.RetentionUnit)
				assert.Equal(t, tc.expectedResult.RetentionValue, result.RetentionValue)
			}
		})
	}
}

func TestFlattenScheduledPolicyItems(t *testing.T) {
	testCases := map[string]struct {
		items         []admin20250312010.BackupComplianceScheduledPolicyItem
		frequencyType string
		expectedLen   int
	}{
		"weeklyItems": {
			items: []admin20250312010.BackupComplianceScheduledPolicyItem{
				{
					Id:                func() *string { s := "507f1f77bcf86cd799439023"; return &s }(),
					FrequencyType:     "weekly",
					FrequencyInterval: 1,
					RetentionUnit:     "weeks",
					RetentionValue:    4,
				},
				{
					Id:                func() *string { s := "507f1f77bcf86cd799439024"; return &s }(),
					FrequencyType:     "weekly",
					FrequencyInterval: 2,
					RetentionUnit:     "weeks",
					RetentionValue:    8,
				},
				{
					Id:                func() *string { s := "507f1f77bcf86cd799439025"; return &s }(),
					FrequencyType:     "monthly",
					FrequencyInterval: 1,
					RetentionUnit:     "months",
					RetentionValue:    6,
				},
			},
			frequencyType: "weekly",
			expectedLen:   2,
		},
		"noMatchingItems": {
			items: []admin20250312010.BackupComplianceScheduledPolicyItem{
				{
					FrequencyType: "hourly",
				},
				{
					FrequencyType: "daily",
				},
			},
			frequencyType: "weekly",
			expectedLen:   0,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			result := flattenScheduledPolicyItems(tc.items, tc.frequencyType)
			assert.Equal(t, tc.expectedLen, len(result))
		})
	}
}

func TestExpandOnDemandPolicyItem(t *testing.T) {
	testCases := map[string]struct {
		item           *OnDemandPolicyItem
		expectedResult *admin20250312010.BackupComplianceOnDemandPolicyItem
	}{
		"withAllFields": {
			item: &OnDemandPolicyItem{
				Id:                func() *string { s := "507f1f77bcf86cd799439020"; return &s }(),
				FrequencyInterval: func() *int { i := 1; return &i }(),
				RetentionUnit:     func() *string { s := "days"; return &s }(),
				RetentionValue:    func() *int { i := 30; return &i }(),
			},
			expectedResult: &admin20250312010.BackupComplianceOnDemandPolicyItem{
				Id:                func() *string { s := "507f1f77bcf86cd799439020"; return &s }(),
				FrequencyInterval: 1,
				FrequencyType:     "ondemand",
				RetentionUnit:     "days",
				RetentionValue:    30,
			},
		},
		"nilItem": {
			item:           nil,
			expectedResult: nil,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			result := expandOnDemandPolicyItem(tc.item)
			if tc.expectedResult == nil {
				assert.Nil(t, result)
			} else {
				assert.Equal(t, tc.expectedResult.Id, result.Id)
				assert.Equal(t, tc.expectedResult.FrequencyInterval, result.FrequencyInterval)
				assert.Equal(t, tc.expectedResult.FrequencyType, result.FrequencyType)
				assert.Equal(t, tc.expectedResult.RetentionUnit, result.RetentionUnit)
				assert.Equal(t, tc.expectedResult.RetentionValue, result.RetentionValue)
			}
		})
	}
}

func TestExpandScheduledPolicyItem(t *testing.T) {
	testCases := map[string]struct {
		item           *ScheduledPolicyItem
		frequencyType  string
		expectedResult admin20250312010.BackupComplianceScheduledPolicyItem
	}{
		"hourlyItem": {
			item: &ScheduledPolicyItem{
				FrequencyInterval: func() *int { i := 6; return &i }(),
				RetentionUnit:     func() *string { s := "days"; return &s }(),
				RetentionValue:    func() *int { i := 7; return &i }(),
			},
			frequencyType: "hourly",
			expectedResult: admin20250312010.BackupComplianceScheduledPolicyItem{
				FrequencyType:     "hourly",
				FrequencyInterval: 6,
				RetentionUnit:     "days",
				RetentionValue:    7,
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			result := expandScheduledPolicyItem(tc.item, tc.frequencyType)
			assert.Equal(t, tc.expectedResult.FrequencyType, result.FrequencyType)
			assert.Equal(t, tc.expectedResult.FrequencyInterval, result.FrequencyInterval)
			assert.Equal(t, tc.expectedResult.RetentionUnit, result.RetentionUnit)
			assert.Equal(t, tc.expectedResult.RetentionValue, result.RetentionValue)
		})
	}
}

func TestGetBackupCompliancePolicyModel(t *testing.T) {
	testCases := map[string]struct {
		policy       *admin20250312010.DataProtectionSettings20231001
		currentModel *Model
		validateFunc func(t *testing.T, model *Model)
	}{
		"completePolicy": {
			policy:       createTestPolicy(),
			currentModel: createTestModel(),
			validateFunc: func(t *testing.T, model *Model) {
				assert.Equal(t, "507f1f77bcf86cd799439011", util.SafeString(model.ProjectId))
				assert.Equal(t, "test@example.com", util.SafeString(model.AuthorizedEmail))
				assert.Equal(t, "John", util.SafeString(model.AuthorizedUserFirstName))
				assert.Equal(t, "Doe", util.SafeString(model.AuthorizedUserLastName))
				assert.Equal(t, "ACTIVE", util.SafeString(model.State))
				assert.Equal(t, "user@example.com", util.SafeString(model.UpdatedUser))
				assert.NotNil(t, model.UpdatedDate)
				assert.NotNil(t, model.CopyProtectionEnabled)
				assert.NotNil(t, model.EncryptionAtRestEnabled)
				assert.NotNil(t, model.PitEnabled)
				assert.NotNil(t, model.RestoreWindowDays)
				assert.NotNil(t, model.OnDemandPolicyItem)
				assert.NotNil(t, model.PolicyItemHourly)
				assert.NotNil(t, model.PolicyItemDaily)
				assert.Equal(t, 2, len(model.PolicyItemWeekly))
				assert.Equal(t, 0, len(model.PolicyItemMonthly))
				assert.Equal(t, 0, len(model.PolicyItemYearly))
			},
		},
		"nilPolicy": {
			policy:       nil,
			currentModel: createTestModel(),
			validateFunc: func(t *testing.T, model *Model) {
				// Should preserve currentModel fields
				assert.Equal(t, "507f1f77bcf86cd799439011", util.SafeString(model.ProjectId))
			},
		},
		"nilScheduledItems": {
			policy: func() *admin20250312010.DataProtectionSettings20231001 {
				p := createTestPolicy()
				p.ScheduledPolicyItems = nil
				return p
			}(),
			currentModel: createTestModel(),
			validateFunc: func(t *testing.T, model *Model) {
				// GetScheduledPolicyItems() returns empty slice when nil, so all policy items should be nil or empty
				assert.Nil(t, model.PolicyItemHourly)
				assert.Nil(t, model.PolicyItemDaily)
				assert.Equal(t, 0, len(model.PolicyItemWeekly))
			},
		},
		"nilCurrentModel": {
			policy:       createTestPolicy(),
			currentModel: nil,
			validateFunc: func(t *testing.T, model *Model) {
				assert.NotNil(t, model)
				assert.Equal(t, "507f1f77bcf86cd799439011", util.SafeString(model.ProjectId))
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			result := GetBackupCompliancePolicyModel(tc.policy, tc.currentModel)
			if tc.validateFunc != nil {
				tc.validateFunc(t, result)
			}
		})
	}
}

func TestExpandDataProtectionSettings(t *testing.T) {
	testCases := map[string]struct {
		model        *Model
		projectID    string
		validateFunc func(t *testing.T, settings *admin20250312010.DataProtectionSettings20231001)
	}{
		"completeModel": {
			model: func() *Model {
				m := createTestModel()
				copyProtection := true
				encryptionAtRest := true
				pitEnabled := true
				restoreWindowDays := 7
				m.CopyProtectionEnabled = &copyProtection
				m.EncryptionAtRestEnabled = &encryptionAtRest
				m.PitEnabled = &pitEnabled
				m.RestoreWindowDays = &restoreWindowDays

				onDemandId := "507f1f77bcf86cd799439020"
				freqInterval1 := 1
				retentionVal30 := 30
				retentionUnitDays := "days"
				m.OnDemandPolicyItem = &OnDemandPolicyItem{
					Id:                &onDemandId,
					FrequencyInterval: &freqInterval1,
					RetentionUnit:     &retentionUnitDays,
					RetentionValue:    &retentionVal30,
				}

				freqInterval6 := 6
				retentionVal7 := 7
				m.PolicyItemHourly = &ScheduledPolicyItem{
					FrequencyInterval: &freqInterval6,
					RetentionUnit:     &retentionUnitDays,
					RetentionValue:    &retentionVal7,
				}

				m.PolicyItemDaily = &ScheduledPolicyItem{
					FrequencyInterval: &freqInterval1,
					RetentionUnit:     &retentionUnitDays,
					RetentionValue:    &retentionVal30,
				}

				retentionUnitWeeks := "weeks"
				retentionVal4 := 4
				m.PolicyItemWeekly = []ScheduledPolicyItem{
					{
						FrequencyInterval: &freqInterval1,
						RetentionUnit:     &retentionUnitWeeks,
						RetentionValue:    &retentionVal4,
					},
				}

				return m
			}(),
			projectID: "507f1f77bcf86cd799439011",
			validateFunc: func(t *testing.T, settings *admin20250312010.DataProtectionSettings20231001) {
				assert.Equal(t, "507f1f77bcf86cd799439011", util.SafeString(settings.ProjectId))
				assert.Equal(t, "test@example.com", settings.AuthorizedEmail)
				assert.Equal(t, "John", settings.AuthorizedUserFirstName)
				assert.Equal(t, "Doe", settings.AuthorizedUserLastName)
				assert.NotNil(t, settings.CopyProtectionEnabled)
				assert.True(t, *settings.CopyProtectionEnabled)
				assert.NotNil(t, settings.EncryptionAtRestEnabled)
				assert.True(t, *settings.EncryptionAtRestEnabled)
				assert.NotNil(t, settings.PitEnabled)
				assert.True(t, *settings.PitEnabled)
				assert.NotNil(t, settings.RestoreWindowDays)
				assert.Equal(t, 7, *settings.RestoreWindowDays)
				assert.NotNil(t, settings.OnDemandPolicyItem)
				assert.NotNil(t, settings.ScheduledPolicyItems)
				assert.Equal(t, 3, len(*settings.ScheduledPolicyItems))
			},
		},
		"withDefaults": {
			model:     createTestModel(),
			projectID: "507f1f77bcf86cd799439011",
			validateFunc: func(t *testing.T, settings *admin20250312010.DataProtectionSettings20231001) {
				// Boolean fields should default to false
				assert.NotNil(t, settings.CopyProtectionEnabled)
				assert.False(t, *settings.CopyProtectionEnabled)
				assert.NotNil(t, settings.EncryptionAtRestEnabled)
				assert.False(t, *settings.EncryptionAtRestEnabled)
				assert.NotNil(t, settings.PitEnabled)
				assert.False(t, *settings.PitEnabled)
				// RestoreWindowDays should default to 0
				assert.NotNil(t, settings.RestoreWindowDays)
				assert.Equal(t, 0, *settings.RestoreWindowDays)
			},
		},
		"withAllPolicyItems": {
			model: func() *Model {
				m := createTestModel()
				freqInterval6 := 6
				freqInterval1 := 1
				retentionUnitDays := "days"
				retentionVal7 := 7
				retentionVal30 := 30
				retentionUnitWeeks := "weeks"
				retentionVal4 := 4
				retentionUnitMonths := "months"
				retentionVal6 := 6
				retentionUnitYears := "years"
				retentionVal1 := 1
				m.PolicyItemHourly = &ScheduledPolicyItem{
					FrequencyInterval: &freqInterval6,
					RetentionUnit:     &retentionUnitDays,
					RetentionValue:    &retentionVal7,
				}
				m.PolicyItemDaily = &ScheduledPolicyItem{
					FrequencyInterval: &freqInterval1,
					RetentionUnit:     &retentionUnitDays,
					RetentionValue:    &retentionVal30,
				}
				m.PolicyItemWeekly = []ScheduledPolicyItem{
					{
						FrequencyInterval: &freqInterval1,
						RetentionUnit:     &retentionUnitWeeks,
						RetentionValue:    &retentionVal4,
					},
				}
				m.PolicyItemMonthly = []ScheduledPolicyItem{
					{
						FrequencyInterval: &freqInterval1,
						RetentionUnit:     &retentionUnitMonths,
						RetentionValue:    &retentionVal6,
					},
				}
				m.PolicyItemYearly = []ScheduledPolicyItem{
					{
						FrequencyInterval: &freqInterval1,
						RetentionUnit:     &retentionUnitYears,
						RetentionValue:    &retentionVal1,
					},
				}
				return m
			}(),
			projectID: "507f1f77bcf86cd799439011",
			validateFunc: func(t *testing.T, settings *admin20250312010.DataProtectionSettings20231001) {
				assert.NotNil(t, settings.ScheduledPolicyItems)
				assert.Equal(t, 5, len(*settings.ScheduledPolicyItems))
			},
		},
		"withNoPolicyItems": {
			model:     createTestModel(),
			projectID: "507f1f77bcf86cd799439011",
			validateFunc: func(t *testing.T, settings *admin20250312010.DataProtectionSettings20231001) {
				// When no policy items are set, ScheduledPolicyItems should be nil
				assert.Nil(t, settings.ScheduledPolicyItems)
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			result := expandDataProtectionSettings(tc.model, tc.projectID)
			if tc.validateFunc != nil {
				tc.validateFunc(t, result)
			}
		})
	}
}
