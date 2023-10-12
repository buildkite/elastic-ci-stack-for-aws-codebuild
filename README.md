> [!IMPORTANT]
> This experiment is no longer under development, and is not maintained/supported.

# Elastic CI Stack for AWS Lambda and CodeBuild

This is an experimental stack to explore the use of AWS Lambda for terminating Buildkite webhooks and then AWS CodeBuild for execution of workloads.

The advantage of this is CodeBuild is well isolated and provides a docker daemon, so it's a good environment for builds, where as Fargate doesn't allow docker to be run.

https://docs.aws.amazon.com/codebuild/latest/userguide/limits.html

### Notes

- Lambda receives webhooks from BK for job on queue (say `default`)
- Lambda creates compute with an agent that will have a queue like `job-{uuid}`
- Lambda updates job to new queue `job-{uuid}`
- Compute starts up, agent starts up with `--disconnect-after-job`
- Build runs, finishes, compute finishes

### Questions?

* Is ~20s job start time acceptable?
* How do we protect the agent registration token in each environment?
