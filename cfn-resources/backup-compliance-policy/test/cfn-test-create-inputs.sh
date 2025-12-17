#!/usr/bin/env bash
# cfn-test-create-inputs.sh
#
# This tool generates json files in the inputs/ for `cfn test`.
#

set -o errexit
set -o nounset
set -o pipefail

function usage {
	echo "Creates test inputs for Backup Compliance Policy"
}

if [ "$#" -ne 2 ]; then usage; fi
if [[ "$*" == help ]]; then usage; fi

rm -rf inputs
mkdir inputs

#set profile
profile="default"
if [ ${MONGODB_ATLAS_PROFILE+x} ]; then
	echo "profile set to ${MONGODB_ATLAS_PROFILE}"
	profile=${MONGODB_ATLAS_PROFILE}
fi

projectName="${1}"
projectId=$(atlas projects list --output json | jq --arg NAME "${projectName}" -r '.results[] | select(.name==$NAME) | .id')
if [ -z "$projectId" ]; then
	projectId=$(atlas projects create "${projectName}" --output=json | jq -r '.id')

	echo -e "Created project \"${projectName}\" with id: ${projectId}\n"
else
	echo -e "FOUND project \"${projectName}\" with id: ${projectId}\n"
fi
echo -e "=====\nrun this command to clean up\n=====\nmongocli iam projects delete ${projectId} --force\n====="

# Get the current user email for authorized email
authorizedEmail="${2}"
if [ -z "$authorizedEmail" ]; then
	# Try to get from Atlas CLI config
	authorizedEmail=$(atlas config get --output json | jq -r '.publicApiKey // "test@example.com"')
	if [ "$authorizedEmail" == "null" ] || [ -z "$authorizedEmail" ]; then
		authorizedEmail="test@example.com"
	fi
fi

jq --arg projectId "$projectId" \
	--arg authorizedEmail "$authorizedEmail" \
	'.ProjectId?|=$projectId |.AuthorizedEmail?|=$authorizedEmail' \
	"$(dirname "$0")/inputs_1_create.json" >"inputs/inputs_1_create.json"

jq --arg projectId "$projectId" \
	--arg authorizedEmail "$authorizedEmail" \
	'.ProjectId?|=$projectId |.AuthorizedEmail?|=$authorizedEmail' \
	"$(dirname "$0")/inputs_1_update.json" >"inputs/inputs_1_update.json"

jq --arg projectId "$projectId" \
	--arg authorizedEmail "$authorizedEmail" \
	'.ProjectId?|=$projectId |.AuthorizedEmail?|=$authorizedEmail' \
	"$(dirname "$0")/inputs_2_create.json" >"inputs/inputs_2_create.json"

jq --arg projectId "$projectId" \
	--arg authorizedEmail "$authorizedEmail" \
	'.ProjectId?|=$projectId |.AuthorizedEmail?|=$authorizedEmail' \
	"$(dirname "$0")/inputs_2_update.json" >"inputs/inputs_2_update.json"

ls -l inputs
echo "mongocli iam projects delete ${projectId} --force"
