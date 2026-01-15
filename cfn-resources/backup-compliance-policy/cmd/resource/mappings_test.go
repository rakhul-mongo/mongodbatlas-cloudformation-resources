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
	"testing"
	"time"

	"github.com/mongodb/mongodbatlas-cloudformation-resources/util"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/atlas-sdk/v20250312012/admin"
)

func TestHelperFunctions(t *testing.T) {
	t.Run("boolToStringPtr", func(t *testing.T) {
		assert.Equal(t, "true", *boolToStringPtr(true))
		assert.Equal(t, "false", *boolToStringPtr(false))
	})

	t.Run("stringPtrToBool", func(t *testing.T) {
		assert.True(t, stringPtrToBool(util.StringPtr("true")))
		assert.False(t, stringPtrToBool(util.StringPtr("false")))
		assert.False(t, stringPtrToBool(nil))
	})

	t.Run("stringPtrToInt", func(t *testing.T) {
		assert.Equal(t, 42, stringPtrToInt(util.StringPtr("42"), 0))
		assert.Equal(t, 10, stringPtrToInt(nil, 10))
		assert.Equal(t, 20, stringPtrToInt(util.StringPtr("invalid"), 20))
	})
}

func TestPolicyItemConversions(t *testing.T) {
	t.Run("getOnDemandPolicyItem", func(t *testing.T) {
		assert.Nil(t, getOnDemandPolicyItem(nil))

		item := &admin.BackupComplianceOnDemandPolicyItem{
			Id:                util.StringPtr("id-1"),
			FrequencyType:     "ondemand",
			FrequencyInterval: 1,
			RetentionUnit:     "days",
			RetentionValue:    7,
		}
		result := getOnDemandPolicyItem(item)
		assert.Equal(t, "id-1", *result.Id)
		assert.Equal(t, "1", *result.FrequencyInterval)
		assert.Equal(t, "7", *result.RetentionValue)
	})

	t.Run("expandOnDemandPolicyItem", func(t *testing.T) {
		assert.Nil(t, expandOnDemandPolicyItem(nil))

		model := &OnDemandPolicyItem{
			FrequencyInterval: util.StringPtr("2"),
			RetentionUnit:     util.StringPtr("weeks"),
			RetentionValue:    util.StringPtr("4"),
		}
		result := expandOnDemandPolicyItem(model)
		assert.Equal(t, "ondemand", result.FrequencyType)
		assert.Equal(t, 2, result.FrequencyInterval)
		assert.Equal(t, 4, result.RetentionValue)
	})

	t.Run("getScheduledPolicyItem", func(t *testing.T) {
		assert.Nil(t, getScheduledPolicyItem(nil))

		item := &admin.BackupComplianceScheduledPolicyItem{
			FrequencyType:     "hourly",
			FrequencyInterval: 6,
			RetentionUnit:     "days",
			RetentionValue:    3,
		}
		result := getScheduledPolicyItem(item)
		assert.Equal(t, "hourly", *result.FrequencyType)
		assert.Equal(t, "6", *result.FrequencyInterval)
	})

	t.Run("expandScheduledPolicyItem", func(t *testing.T) {
		model := &ScheduledPolicyItem{
			FrequencyInterval: util.StringPtr("1"),
			RetentionUnit:     util.StringPtr("weeks"),
			RetentionValue:    util.StringPtr("4"),
		}
		result := expandScheduledPolicyItem(model, "daily")
		assert.Equal(t, "daily", result.FrequencyType)
		assert.Equal(t, 1, result.FrequencyInterval)
	})
}

func TestSetBackupCompliancePolicyData(t *testing.T) {
	tests := map[string]struct {
		policy *admin.DataProtectionSettings20231001
		check  func(*testing.T, *Model)
	}{
		"nil policy": {
			policy: nil,
			check:  func(t *testing.T, m *Model) { assert.Nil(t, m.ProjectId) },
		},
		"basic fields": {
			policy: &admin.DataProtectionSettings20231001{
				ProjectId:             util.StringPtr("proj-123"),
				AuthorizedEmail:       "admin@example.com",
				CopyProtectionEnabled: util.Pointer(true),
				RestoreWindowDays:     util.IntPtr(7),
				State:                 util.StringPtr("ACTIVE"),
			},
			check: func(t *testing.T, m *Model) {
				assert.Equal(t, "proj-123", *m.ProjectId)
				assert.Equal(t, "admin@example.com", *m.AuthorizedEmail)
				assert.Equal(t, "true", *m.CopyProtectionEnabled)
				assert.Equal(t, "7", *m.RestoreWindowDays)
				assert.Equal(t, "ACTIVE", *m.State)
			},
		},
		"with scheduled items": {
			policy: &admin.DataProtectionSettings20231001{
				ProjectId:       util.StringPtr("proj-456"),
				AuthorizedEmail: "test@example.com",
				ScheduledPolicyItems: &[]admin.BackupComplianceScheduledPolicyItem{
					{FrequencyType: "hourly", FrequencyInterval: 6, RetentionUnit: "days", RetentionValue: 3},
					{FrequencyType: "daily", FrequencyInterval: 1, RetentionUnit: "weeks", RetentionValue: 1},
					{FrequencyType: "weekly", FrequencyInterval: 1, RetentionUnit: "months", RetentionValue: 1},
				},
				State: util.StringPtr("ACTIVE"),
			},
			check: func(t *testing.T, m *Model) {
				assert.NotNil(t, m.PolicyItemHourly)
				assert.NotNil(t, m.PolicyItemDaily)
				assert.Len(t, m.PolicyItemWeekly, 1)
				assert.Equal(t, "hourly", *m.PolicyItemHourly.FrequencyType)
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			model := &Model{}
			setBackupCompliancePolicyData(model, tc.policy)
			tc.check(t, model)
		})
	}
}

func TestExpandDataProtectionSettings(t *testing.T) {
	tests := map[string]struct {
		model *Model
		check func(*testing.T, *admin.DataProtectionSettings20231001)
	}{
		"with empty policy items": {
			model: &Model{
				AuthorizedEmail:    util.StringPtr("admin@example.com"),
				OnDemandPolicyItem: &OnDemandPolicyItem{},
				PolicyItemHourly:   &ScheduledPolicyItem{},
				PolicyItemDaily:    &ScheduledPolicyItem{},
			},
			check: func(t *testing.T, s *admin.DataProtectionSettings20231001) {
				assert.Nil(t, s.OnDemandPolicyItem, "empty OnDemandPolicyItem should be omitted")
				assert.Nil(t, s.ScheduledPolicyItems, "empty policy items should not create scheduled items")
			},
		},
		"minimal model": {
			model: &Model{
				AuthorizedEmail:         util.StringPtr("admin@example.com"),
				AuthorizedUserFirstName: util.StringPtr("Jane"),
				AuthorizedUserLastName:  util.StringPtr("Smith"),
			},
			check: func(t *testing.T, s *admin.DataProtectionSettings20231001) {
				assert.Equal(t, "admin@example.com", s.AuthorizedEmail)
				assert.Equal(t, "Jane", s.AuthorizedUserFirstName)
				assert.False(t, *s.CopyProtectionEnabled)
			},
		},
		"with booleans and integers": {
			model: &Model{
				AuthorizedEmail:         util.StringPtr("test@example.com"),
				CopyProtectionEnabled:   util.StringPtr("true"),
				EncryptionAtRestEnabled: util.StringPtr("true"),
				PitEnabled:              util.StringPtr("false"),
				RestoreWindowDays:       util.StringPtr("14"),
			},
			check: func(t *testing.T, s *admin.DataProtectionSettings20231001) {
				assert.True(t, *s.CopyProtectionEnabled)
				assert.True(t, *s.EncryptionAtRestEnabled)
				assert.False(t, *s.PitEnabled)
				assert.Equal(t, 14, *s.RestoreWindowDays)
			},
		},
		"with all policy items": {
			model: &Model{
				AuthorizedEmail: util.StringPtr("admin@example.com"),
				PolicyItemHourly: &ScheduledPolicyItem{
					FrequencyInterval: util.StringPtr("6"),
					RetentionUnit:     util.StringPtr("days"),
					RetentionValue:    util.StringPtr("3"),
				},
				PolicyItemDaily: &ScheduledPolicyItem{
					FrequencyInterval: util.StringPtr("1"),
					RetentionUnit:     util.StringPtr("weeks"),
					RetentionValue:    util.StringPtr("1"),
				},
				PolicyItemWeekly: []ScheduledPolicyItem{
					{
						FrequencyInterval: util.StringPtr("1"),
						RetentionUnit:     util.StringPtr("months"),
						RetentionValue:    util.StringPtr("2"),
					},
				},
			},
			check: func(t *testing.T, s *admin.DataProtectionSettings20231001) {
				assert.NotNil(t, s.ScheduledPolicyItems)
				items := *s.ScheduledPolicyItems
				assert.Len(t, items, 3)
				assert.Equal(t, "hourly", items[0].FrequencyType)
				assert.Equal(t, "daily", items[1].FrequencyType)
				assert.Equal(t, "weekly", items[2].FrequencyType)
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := expandDataProtectionSettings(tc.model, "test-project")
			assert.Equal(t, "test-project", *result.ProjectId)
			tc.check(t, result)
		})
	}
}

func TestRoundTripConversion(t *testing.T) {
	t.Run("policy survives round trip", func(t *testing.T) {
		original := &admin.DataProtectionSettings20231001{
			ProjectId:               util.StringPtr("proj-999"),
			AuthorizedEmail:         "security@example.com",
			AuthorizedUserFirstName: "Test",
			AuthorizedUserLastName:  "User",
			CopyProtectionEnabled:   util.Pointer(true),
			PitEnabled:              util.Pointer(false),
			RestoreWindowDays:       util.IntPtr(7),
			State:                   util.StringPtr("ACTIVE"),
			UpdatedDate:             &time.Time{},
			UpdatedUser:             util.StringPtr("admin"),
		}

		model := &Model{}
		setBackupCompliancePolicyData(model, original)

		result := expandDataProtectionSettings(model, "proj-999")

		assert.Equal(t, *original.ProjectId, *result.ProjectId)
		assert.Equal(t, original.AuthorizedEmail, result.AuthorizedEmail)
		assert.Equal(t, *original.CopyProtectionEnabled, *result.CopyProtectionEnabled)
		assert.Equal(t, *original.RestoreWindowDays, *result.RestoreWindowDays)
	})
}
