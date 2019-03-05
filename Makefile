.PHONY: lambda-sync validate

LAMBDA_S3_BUCKET := buildkite-aws-codebuild-lox
LAMBDA_S3_BUCKET_PATH := /
ZIPS := handle-buildkite-webhook.zip buildkite-agent-codebuild.zip

build: $(ZIPS)

clean:
	-rm $(ZIPS)
	-rm lambdas/handle-buildkite-webhook/handler

handle-buildkite-webhook.zip: lambdas/handle-buildkite-webhook/handler
	zip -9 -v -j $@ "$<"

buildkite-agent-codebuild.zip: templates/codebuild/buildspec.yaml
	zip -9 -v -j $@ "$<"

lambdas/handle-buildkite-webhook/handler: lambdas/handle-buildkite-webhook/main.go
	docker run \
		--volume go-module-cache:/go/pkg/mod \
		--volume $(PWD):/code \
		--workdir /code \
		--rm golang:1.11 \
		go build -ldflags="$(FLAGS)" -o ./lambdas/handle-buildkite-webhook/handler ./lambdas/handle-buildkite-webhook/
	chmod +x lambdas/handle-buildkite-webhook/handler

publish: $(ZIPS)
	aws s3 sync \
		--acl public-read \
		--exclude '*' --include '*.zip' \
		. s3://$(LAMBDA_S3_BUCKET)$(LAMBDA_S3_BUCKET_PATH)

lambda-version:
	aws s3api head-object \
		--bucket ${LAMBDA_S3_BUCKET} \
		--key handle-buildkite-webhook.zip --query "VersionId" --output text

validate:
	aws cloudformation validate-template \
		--template-body file://templates/cloudformation/template.yml

