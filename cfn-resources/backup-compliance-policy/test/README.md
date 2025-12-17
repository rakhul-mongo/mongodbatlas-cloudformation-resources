# Backup Compliance Policy

## Impact

The following components use this resource and are potentially impacted by any changes. They should also be validated to ensure the changes do not cause a regression.

- Backup Compliance Policy L1 CDK constructor

## Prerequisites

### Resources needed to run the manual QA

- Atlas Project

All resources are created as part of `cfn-testing-helper.sh`

## Manual QA

Please, follows the steps in [TESTING.md](../../../TESTING.md).

### Success criteria when testing the resource

- Backup Compliance Policy for the respective Project in Atlas should be correctly configured:
  - Navigate to the Atlas UI → Project → Backup → Compliance Policy
  - Verify that the policy settings match the CloudFormation template:
    - Authorized Email, First Name, Last Name
    - Copy Protection Enabled status
    - Encryption at Rest Enabled status
    - Point-in-Time Recovery (PIT) Enabled status
    - Restore Window Days
    - On-demand policy item (if configured)
    - Scheduled policy items (hourly, daily, weekly, monthly, yearly)
  - Verify that the policy state is ACTIVE after creation
  - Verify that updates to the policy are reflected correctly
  - Verify that deletion disables the policy

## Important Links

- [API Documentation](https://www.mongodb.com/docs/atlas/reference/api-resources-spec/v2/#tag/Cloud-Backups/operation/updateCompliancePolicy)
- [Resource Usage Documentation](https://www.mongodb.com/docs/atlas/backup/cloud-backup/backup-compliance-policy/)

## Local Testing

The local tests are integrated with the AWS `sam local` and `cfn invoke` tooling features:

```
sam local start-lambda --skip-pull-image
```

then in another shell:

```bash
repo_root=$(git rev-parse --show-toplevel)
source <(${repo_root}/quickstart-mongodb-atlas/scripts/export-mongocli-config.py)
cd ${repo_root}/cfn-resources/backup-compliance-policy
./test/cfn-test-create-inputs.sh YourProjectName your-email@example.com
cd test
cfn invoke CREATE inputs/inputs_1_create.json
cfn invoke READ inputs/inputs_1_create.json
cfn invoke UPDATE inputs/inputs_1_update.json
cfn invoke DELETE inputs/inputs_1_create.json
```

Both CREATE, READ, UPDATE & DELETE tests must pass.

## Test Input Files

### inputs_1_create.json

Basic backup compliance policy with minimal configuration:

- Required fields: ProjectId, AuthorizedEmail, AuthorizedUserFirstName, AuthorizedUserLastName
- Optional fields: CopyProtectionEnabled (false), EncryptionAtRestEnabled (false), PitEnabled (false)

### inputs_1_update.json

Update to inputs_1_create.json with:

- CopyProtectionEnabled set to true
- EncryptionAtRestEnabled set to true
- PitEnabled set to true
- RestoreWindowDays set to 7

### inputs_2_create.json

Comprehensive backup compliance policy with all policy items:

- All required fields
- All optional boolean fields set to true
- RestoreWindowDays set to 14
- OnDemandPolicyItem configured
- PolicyItemHourly configured
- PolicyItemDaily configured
- Multiple PolicyItemWeekly items
- PolicyItemMonthly items
- PolicyItemYearly items

### inputs_2_update.json

Update to inputs_2_create.json with:

- Boolean fields set to false
- RestoreWindowDays increased to 30
- Updated retention values for all policy items
- PolicyItemYearly removed (empty array)

## Notes

- The Backup Compliance Policy is a project-level resource (one per project)
- The primary identifier is ProjectId
- Policy items (Id, FrequencyType) are read-only and returned by the API
- When updating policy items, only FrequencyInterval, RetentionUnit, and RetentionValue can be modified
- The policy must be enabled before it can be used
- Deleting the resource disables the policy but does not delete the project
