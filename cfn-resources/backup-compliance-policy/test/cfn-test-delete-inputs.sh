#!/usr/bin/env bash
# cfn-test-delete-inputs.sh
#
# This tool deletes the mongodb resources used for `cfn test` as inputs.
# For Backup Compliance Policy, we must delete clusters first before the policy can be deleted.
#

set -o errexit
set -o nounset
set -o pipefail

function usage {
	echo "usage:$0 "
}

projectId=$(jq -r '.ProjectId' ./inputs/inputs_1_create.json)

if [ -z "$projectId" ] || [ "$projectId" == "null" ]; then
	echo "ERROR: Could not read ProjectId from inputs/inputs_1_create.json"
	exit 1
fi

echo "Deleting resources for project: $projectId"

# First, delete all clusters in the project (required before deleting Backup Compliance Policy)
# Backup Compliance Policy cannot be disabled if clusters exist
echo "Checking for clusters in project..."
clusterList=$(atlas clusters list --projectId "${projectId}" --output json 2>/dev/null | jq -r '.results[]?.name // empty' 2>/dev/null || echo "")

if [ -n "$clusterList" ]; then
	echo "Found clusters, deleting them first (required before deleting Backup Compliance Policy)..."
	while IFS= read -r clusterName; do
		if [ -n "$clusterName" ] && [ "$clusterName" != "null" ] && [ "$clusterName" != "" ]; then
			echo "Deleting cluster: $clusterName"
			if atlas clusters delete "${clusterName}" --projectId "${projectId}" --force 2>/dev/null; then
				echo "Cluster deletion initiated for: $clusterName"
				# Wait for cluster to be deleted
				echo "Waiting for cluster $clusterName to be deleted..."
				while atlas clusters describe "${clusterName}" --projectId "${projectId}" 2>/dev/null; do
					sleep 10
					echo "Still waiting for cluster $clusterName to be deleted..."
				done
				echo "Cluster $clusterName deleted successfully"
			else
				echo "Warning: Failed to delete cluster $clusterName (may already be deleted)"
			fi
		fi
	done <<< "$clusterList"
else
	echo "No clusters found in project"
fi

# Now delete the Backup Compliance Policy (if it exists)
# The CFN DELETE handler will be called, but we've removed the blocker (clusters)
echo "Clusters deleted. Backup Compliance Policy can now be deleted by CFN DELETE handler."

# Finally, delete the project
echo "Deleting project: $projectId"
if mongocli iam projects delete "$projectId" --force 2>/dev/null; then
	echo "$projectId project deletion OK"
else
	echo "Warning: Failed to delete project $projectId (may already be deleted or have dependencies)"
	# Don't exit with error - project deletion is best-effort cleanup
fi
