#!/bin/bash

# This was copied from Datadog's system-test repository
get_circleci_artifact() {

    SLUG=$1
    WORKFLOW_NAME=$2
    JOB_NAME=$3
    ARTIFACT_PATTERN=$4

    echo "CircleCI: https://app.circleci.com/pipelines/$SLUG?branch=master"
    PIPELINES=$(curl --silent https://circleci.com/api/v2/project/$SLUG/pipeline?branch=master -H "Circle-Token: $CIRCLECI_TOKEN")

    for i in {0..30}; do
        PIPELINE_ID=$(echo $PIPELINES| jq -r ".items[$i].id")
        PIPELINE_NUMBER=$(echo $PIPELINES | jq -r ".items[$i].number")

        echo "Trying pipeline #$i $PIPELINE_NUMBER/$PIPELINE_ID"
        WORKFLOWS=$(curl --silent https://circleci.com/api/v2/pipeline/$PIPELINE_ID/workflow -H "Circle-Token: $CIRCLECI_TOKEN")

        QUERY=".items[] | select(.name == \"$WORKFLOW_NAME\") | .id"
        WORKFLOW_IDS=$(echo $WORKFLOWS | jq -r "$QUERY")

        if [ ! -z "$WORKFLOW_IDS" ]; then

            for WORKFLOW_ID in $WORKFLOW_IDS; do
                echo "=> https://app.circleci.com/pipelines/$SLUG/$PIPELINE_NUMBER/workflows/$WORKFLOW_ID"

                JOBS=$(curl --silent https://circleci.com/api/v2/workflow/$WORKFLOW_ID/job -H "Circle-Token: $CIRCLECI_TOKEN")

                QUERY=".items[] | select(.name == \"$JOB_NAME\" and .status==\"success\")"
                JOB=$(echo $JOBS | jq "$QUERY")

                if [ ! -z "$JOB" ]; then
                    break
                fi
            done

            if [ ! -z "$JOB" ]; then
                break
            fi
        fi
    done

    if [ -z "$JOB" ]; then
        echo "Oooops, I did not found any successful pipeline"
        exit 1
    fi

    JOB_NUMBER=$(echo $JOB | jq -r ".job_number")
    JOB_ID=$(echo $JOB | jq -r ".id")

    echo "Job number/ID: $JOB_NUMBER/$JOB_ID"
    echo "Job URL: https://app.circleci.com/pipelines/$SLUG/$PIPELINE_NUMBER/workflows/$WORKFLOW_ID/jobs/$JOB_NUMBER"

    ARTICATS_URL="https://circleci.com/api/v2/project/$SLUG/$JOB_NUMBER/artifacts"
    echo "Artifacts URL: $ARTICATS_URL"

    ARTIFACTS=$(curl --silent $ARTICATS_URL -H "Circle-Token: $CIRCLECI_TOKEN")

    QUERY=".items[] | select(.path | test(\"$ARTIFACT_PATTERN\"))"
    ARTIFACT_URL=$(echo $ARTIFACTS | jq -r "$QUERY | .url")

    if [ -z "$ARTIFACT_URL" ]; then
        echo "Oooops, I did not found any artifact that satisfy this pattern: $ARTIFACT_PATTERN. Here is the list:"
        echo $ARTIFACTS | jq -r ".items[] | .path"
        exit 1
    fi

    ARTIFACT_NAME=$(echo $ARTIFACTS | jq -r "$QUERY | .path" | sed -E 's/libs\///')
    echo "Artifact URL: $ARTIFACT_URL"
    echo "Artifact name: $ARTIFACT_NAME"
    echo "Downloading artifact..."

    curl --silent -L $ARTIFACT_URL --output $ARTIFACT_NAME
}

get_circleci_artifact "gh/DataDog/dd-trace-php" "build_packages" "package extension" "datadog-setup.php"

BINARIES_TRACER_N=$(ls -1 dd-library-php*.tar.gz)
php ./datadog-setup.php --php-bin php --enable-profiling
