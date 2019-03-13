package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/buildkite/elastic-ci-stack-for-aws-codebuild/buildkite"
)

const (
	agentQueryRuleTemplate = `queue=job-%s`
)

func handleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Printf("Processing request data for request %s.\n", request.RequestContext.RequestID)
	log.Printf("Body size = %d.\n", len(request.Body))
	log.Printf("Body = %s", request.Body)

	var event struct {
		Event string           `json:"event"`
		Build *json.RawMessage `json:"build"`
		Job   *json.RawMessage `json:"job"`
	}

	if err := json.Unmarshal([]byte(request.Body), &event); err != nil {
		log.Fatal(err)
	}

	switch event.Event {
	case `job.scheduled`:
		var eventJob struct {
			UUID string `json:"id"`
		}

		if err := json.Unmarshal(*event.Job, &eventJob); err != nil {
			log.Fatal(err)
		}

		codebuildProject := mustGetEnv(`CODEBUILD_PROJECT`)
		graphQLToken := mustGetEnv(`BUILDKITE_GRAPHQL_TOKEN`)

		if err := dispatchJob(codebuildProject, eventJob.UUID, graphQLToken); err != nil {
			log.Fatal(err)
		}
	default:
		log.Printf("Not sure how to handle event type %q", event.Event)
	}

	return events.APIGatewayProxyResponse{Body: request.Body, StatusCode: 200}, nil
}

func mustGetEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic(fmt.Sprintf("Must set %s in environment", key))
	}
	return val
}

func main() {
	if os.Getenv(`DEBUG`) == `1` {
		codebuildProject := mustGetEnv(`CODEBUILD_PROJECT`)
		jobUUID := mustGetEnv(`BUILDKITE_JOB_ID`)
		graphQLToken := mustGetEnv(`BUILDKITE_GRAPHQL_TOKEN`)

		if err := dispatchJob(codebuildProject, jobUUID, graphQLToken); err != nil {
			log.Fatal(err)
		}

	} else {
		lambda.Start(handleRequest)
	}
}

func dispatchJob(codebuildProject string, jobUUID string, graphQLToken string) error {
	log.Printf("CodeBuild Project: %s", codebuildProject)
	log.Printf("Buildkite Job UUID: %s", jobUUID)

	if err := requeueJob(jobUUID, graphQLToken); err != nil {
		return err
	}

	return startCodeBuildEnvironment(codebuildProject, jobUUID)
}

func requeueJob(jobUUID string, graphQLToken string) error {
	client, err := buildkite.NewClient(graphQLToken)
	if err != nil {
		return err
	}

	log.Printf("Resolving Job UUID %s to a GraphQL ID", jobUUID)

	jobID, err := client.GetJobID(jobUUID)
	if err != nil {
		return err
	}

	log.Printf("Found GraphQL ID of %s", jobID)

	query := fmt.Sprintf(agentQueryRuleTemplate, jobUUID)
	log.Printf("Rewriting job query rule to %s", query)

	return client.ChangeJobQueryRule(jobID, query)
}

func startCodeBuildEnvironment(project string, jobUUID string) error {
	sess := session.Must(session.NewSession(aws.NewConfig()))
	svc := codebuild.New(sess)

	input := &codebuild.StartBuildInput{
		ProjectName: aws.String(project),
		EnvironmentVariablesOverride: []*codebuild.EnvironmentVariable{
			{
				Name:  aws.String("BUILDKITE_JOB_ID"),
				Value: aws.String(jobUUID),
			},
		},
		ArtifactsOverride: &codebuild.ProjectArtifacts{
			Type: aws.String("NO_ARTIFACTS"),
		},
	}

	log.Printf("Creating a build environment for %s", project)

	startBuildOutput, err := svc.StartBuild(input)
	if err != nil {
		return err
	}

	log.Printf("Result: %#v", startBuildOutput)

	return nil
}
