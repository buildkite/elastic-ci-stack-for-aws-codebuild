#!/bin/bash
set -euo pipefail

# Used in development of the stack.

LAMBDA_S3_BUCKET=buildkite-aws-codebuild-lox

printf -- '\n--- Updating zips\n'

make build publish

lambda_object_version=$(aws s3api head-object \
		--bucket ${LAMBDA_S3_BUCKET} \
		--key handle-buildkite-webhook.zip --query "VersionId" --output text)

printf -- '\n--- Updating stack\n'

parfait update-stack \
  -t templates/cloudformation/template.yml \
  "buildkite-codebuild" \
  "BuildkiteQueue=default" \
  "LambdaBucket=${LAMBDA_S3_BUCKET}" \
  "LambdaObjectVersion=${lambda_object_version}"
