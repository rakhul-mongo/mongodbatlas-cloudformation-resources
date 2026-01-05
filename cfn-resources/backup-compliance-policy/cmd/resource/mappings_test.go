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

package resource_test

import (
	"testing"

	"go.mongodb.org/atlas-sdk/v20250312010/admin"

	"github.com/mongodb/mongodbatlas-cloudformation-resources/backup-compliance-policy/cmd/resource"
	"github.com/mongodb/mongodbatlas-cloudformation-resources/util"
	"github.com/stretchr/testify/assert"
)

func TestFlattenOnDemandPolicyItem(t *testing.T) {
	testCases := map[string]struct {
		item           *admin.BackupComplianceOnDemandPolicyItem
		expectedResult *resource.OnDemandPolicyItem
	}{
		"withAllFields": {
			item: &admin.BackupComplianceOnDemandPolicyItem{
				Id:                func() *string { s := "507f1f77bcf86cd799439020"; return &s }(),
				FrequencyInterval: 1,
				FrequencyType:     "ondemand",
				RetentionUnit:     "days",
				RetentionValue:    30,
			},
			expectedResult: &resource.OnDemandPolicyItem{
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
			result := resource.FlattenOnDemandPolicyItem(tc.item)
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
		items          []admin.BackupComplianceScheduledPolicyItem
		frequencyType  string
		expectedResult *resource.ScheduledPolicyItem
	}{
		"hourlyItem": {
			items: []admin.BackupComplianceScheduledPolicyItem{
				{
					Id:                func() *string { s := "507f1f77bcf86cd799439021"; return &s }(),
					FrequencyType:     "hourly",
					FrequencyInterval: 6,
					RetentionUnit:     "days",
					RetentionValue:    7,
				},
			},
			frequencyType: "hourly",
			expectedResult: &resource.ScheduledPolicyItem{
				Id:                func() *string { s := "507f1f77bcf86cd799439021"; return &s }(),
				FrequencyType:     func() *string { s := "hourly"; return &s }(),
				FrequencyInterval: func() *int { i := 6; return &i }(),
				RetentionUnit:     func() *string { s := "days"; return &s }(),
				RetentionValue:    func() *int { i := 7; return &i }(),
			},
		},
		"notFound": {
			items: []admin.BackupComplianceScheduledPolicyItem{
				{
					FrequencyType:     "hourly",
					FrequencyInterval: 6,
				},
			},
			frequencyType:  "daily",
			expectedResult: nil,
		},
		"emptyItems": {
			items:          []admin.BackupComplianceScheduledPolicyItem{},
			frequencyType:  "hourly",
			expectedResult: nil,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			result := resource.FlattenScheduledPolicyItem(tc.items, tc.frequencyType)
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
		items         []admin.BackupComplianceScheduledPolicyItem
		frequencyType string
		expectedLen   int
	}{
		"weeklyItems": {
			items: []admin.BackupComplianceScheduledPolicyItem{
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
			items: []admin.BackupComplianceScheduledPolicyItem{
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
			result := resource.FlattenScheduledPolicyItems(tc.items, tc.frequencyType)
			assert.Equal(t, tc.expectedLen, len(result))
		})
	}
}

func TestExpandOnDemandPolicyItem(t *testing.T) {
	testCases := map[string]struct {
		item           *resource.OnDemandPolicyItem
		expectedResult *admin.BackupComplianceOnDemandPolicyItem
	}{
		"withAllFields": {
			item: &resource.OnDemandPolicyItem{
				Id:                func() *string { s := "507f1f77bcf86cd799439020"; return &s }(),
				FrequencyInterval: func() *int { i := 1; return &i }(),
				RetentionUnit:     func() *string { s := "days"; return &s }(),
				RetentionValue:    func() *int { i := 30; return &i }(),
			},
			expectedResult: &admin.BackupComplianceOnDemandPolicyItem{
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
			result := resource.ExpandOnDemandPolicyItem(tc.item)
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
		item           *resource.ScheduledPolicyItem
		frequencyType  string
		expectedResult admin.BackupComplianceScheduledPolicyItem
	}{
		"hourlyItem": {
			item: &resource.ScheduledPolicyItem{
				FrequencyInterval: func() *int { i := 6; return &i }(),
				RetentionUnit:     func() *string { s := "days"; return &s }(),
				RetentionValue:    func() *int { i := 7; return &i }(),
			},
			frequencyType: "hourly",
			expectedResult: admin.BackupComplianceScheduledPolicyItem{
				FrequencyType:     "hourly",
				FrequencyInterval: 6,
				RetentionUnit:     "days",
				RetentionValue:    7,
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			result := resource.ExpandScheduledPolicyItem(tc.item, tc.frequencyType)
			assert.Equal(t, tc.expectedResult.FrequencyType, result.FrequencyType)
			assert.Equal(t, tc.expectedResult.FrequencyInterval, result.FrequencyInterval)
			assert.Equal(t, tc.expectedResult.RetentionUnit, result.RetentionUnit)
			assert.Equal(t, tc.expectedResult.RetentionValue, result.RetentionValue)
		})
	}
}

func TestGetBackupCompliancePolicyModel(t *testing.T) {
	testCases := map[string]struct {
		policy       *admin.DataProtectionSettings20231001
		currentModel *resource.Model
		validateFunc func(t *testing.T, model *resource.Model)
	}{
		"completePolicy": {
			policy:       createTestPolicy(),
			currentModel: createTestModel(),
			validateFunc: func(t *testing.T, model *resource.Model) {
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
			validateFunc: func(t *testing.T, model *resource.Model) {
				// Should preserve currentModel fields
				assert.Equal(t, "507f1f77bcf86cd799439011", util.SafeString(model.ProjectId))
			},
		},
		"nilScheduledItems": {
			policy: func() *admin.DataProtectionSettings20231001 {
				p := createTestPolicy()
				p.ScheduledPolicyItems = nil
				return p
			}(),
			currentModel: createTestModel(),
			validateFunc: func(t *testing.T, model *resource.Model) {
				// GetScheduledPolicyItems() returns empty slice when nil, so all policy items should be nil or empty
				assert.Nil(t, model.PolicyItemHourly)
				assert.Nil(t, model.PolicyItemDaily)
				assert.Equal(t, 0, len(model.PolicyItemWeekly))
			},
		},
		"nilCurrentModel": {
			policy:       createTestPolicy(),
			currentModel: nil,
			validateFunc: func(t *testing.T, model *resource.Model) {
				assert.NotNil(t, model)
				assert.Equal(t, "507f1f77bcf86cd799439011", util.SafeString(model.ProjectId))
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			result := resource.GetBackupCompliancePolicyModel(tc.policy, tc.currentModel)
			if tc.validateFunc != nil {
				tc.validateFunc(t, result)
			}
		})
	}
}

func TestExpandDataProtectionSettings(t *testing.T) {
	testCases := map[string]struct {
		model        *resource.Model
		projectID    string
		validateFunc func(t *testing.T, settings *admin.DataProtectionSettings20231001)
	}{
		"completeModel": {
			model: func() *resource.Model {
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
				m.OnDemandPolicyItem = &resource.OnDemandPolicyItem{
					Id:                &onDemandId,
					FrequencyInterval: &freqInterval1,
					RetentionUnit:     &retentionUnitDays,
					RetentionValue:    &retentionVal30,
				}

				freqInterval6 := 6
				retentionVal7 := 7
				m.PolicyItemHourly = &resource.ScheduledPolicyItem{
					FrequencyInterval: &freqInterval6,
					RetentionUnit:     &retentionUnitDays,
					RetentionValue:    &retentionVal7,
				}

				m.PolicyItemDaily = &resource.ScheduledPolicyItem{
					FrequencyInterval: &freqInterval1,
					RetentionUnit:     &retentionUnitDays,
					RetentionValue:    &retentionVal30,
				}

				retentionUnitWeeks := "weeks"
				retentionVal4 := 4
				m.PolicyItemWeekly = []resource.ScheduledPolicyItem{
					{
						FrequencyInterval: &freqInterval1,
						RetentionUnit:     &retentionUnitWeeks,
						RetentionValue:    &retentionVal4,
					},
				}

				return m
			}(),
			projectID: "507f1f77bcf86cd799439011",
			validateFunc: func(t *testing.T, settings *admin.DataProtectionSettings20231001) {
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
			validateFunc: func(t *testing.T, settings *admin.DataProtectionSettings20231001) {
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
			model: func() *resource.Model {
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
				m.PolicyItemHourly = &resource.ScheduledPolicyItem{
					FrequencyInterval: &freqInterval6,
					RetentionUnit:     &retentionUnitDays,
					RetentionValue:    &retentionVal7,
				}
				m.PolicyItemDaily = &resource.ScheduledPolicyItem{
					FrequencyInterval: &freqInterval1,
					RetentionUnit:     &retentionUnitDays,
					RetentionValue:    &retentionVal30,
				}
				m.PolicyItemWeekly = []resource.ScheduledPolicyItem{
					{
						FrequencyInterval: &freqInterval1,
						RetentionUnit:     &retentionUnitWeeks,
						RetentionValue:    &retentionVal4,
					},
				}
				m.PolicyItemMonthly = []resource.ScheduledPolicyItem{
					{
						FrequencyInterval: &freqInterval1,
						RetentionUnit:     &retentionUnitMonths,
						RetentionValue:    &retentionVal6,
					},
				}
				m.PolicyItemYearly = []resource.ScheduledPolicyItem{
					{
						FrequencyInterval: &freqInterval1,
						RetentionUnit:     &retentionUnitYears,
						RetentionValue:    &retentionVal1,
					},
				}
				return m
			}(),
			projectID: "507f1f77bcf86cd799439011",
			validateFunc: func(t *testing.T, settings *admin.DataProtectionSettings20231001) {
				assert.NotNil(t, settings.ScheduledPolicyItems)
				assert.Equal(t, 5, len(*settings.ScheduledPolicyItems))
			},
		},
		"withNoPolicyItems": {
			model:     createTestModel(),
			projectID: "507f1f77bcf86cd799439011",
			validateFunc: func(t *testing.T, settings *admin.DataProtectionSettings20231001) {
				// When no policy items are set, ScheduledPolicyItems should be nil
				assert.Nil(t, settings.ScheduledPolicyItems)
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			result := resource.ExpandDataProtectionSettings(tc.model, tc.projectID)
			if tc.validateFunc != nil {
				tc.validateFunc(t, result)
			}
		})
	}
}
