#!/bin/bash
set -euo pipefail

LAMBDA_BUCKET=buildkite-aws-codebuild-lox

parfait update-stack \
  -t templates/cloudformation/template.yml \
  "buildkite-codebuild" \
  "BuildkiteQueue=default" \
  "LambdaBucket=${LAMBDA_BUCKET}"
