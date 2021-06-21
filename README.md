# jx registry

[![Documentation](https://godoc.org/github.com/jenkins-x-plugins/jx-registry?status.svg)](https://pkg.go.dev/mod/github.com/jenkins-x-plugins/jx-registry)
[![Go Report Card](https://goreportcard.com/badge/github.com/jenkins-x-plugins/jx-registry)](https://goreportcard.com/report/github.com/jenkins-x-plugins/jx-registry)
[![Releases](https://img.shields.io/github/release-pre/jenkins-x/helmboot.svg)](https://github.com/jenkins-x-plugins/jx-registry/releases)
[![LICENSE](https://img.shields.io/github/license/jenkins-x/helmboot.svg)](https://github.com/jenkins-x-plugins/jx-registry/blob/master/LICENSE)
[![Slack Status](https://img.shields.io/badge/slack-join_chat-white.svg?logo=slack&style=social)](https://slack.k8s.io/)

`jx-registry` is a simple command line tool for working with container registries.

The main use case is initially to support lazy creation of AWS ECR registries on demand. Most other registries allow a registry to be created and used for different images.


## Getting Started

Download the [jx-registry binary](https://github.com/jenkins-x-plugins/jx-registry/releases) for your operating system and add it to your `$PATH`.

## Enabling Cache images

If you wish to also create a cache image in addition to the ECR image for your repository enable the `CACHE_SUFFIX` environment variable.

e.g. in your local `.lighthouse/jenkins-x/release.yaml` file you could do something like:

```yaml
apiVersion: tekton.dev/v1beta1
kind: PipelineRun
metadata:
  creationTimestamp: null
  name: release
spec:
  pipelineSpec:
    tasks:
    - name: from-build-pack
      resources: {}
      taskSpec:
        metadata: {}
        stepTemplate:
          image: uses:jenkins-x/jx3-pipeline-catalog/tasks/javascript/release.yaml@versionStream
        steps:
        - image: uses:jenkins-x/jx3-pipeline-catalog/tasks/git-clone/git-clone.yaml@versionStream
          name: ""
          resources: {}
        - name: next-version
          resources: {}
        - name: jx-variables
          resources: {}
        - name: build-npm-install
          resources: {}
        - name: build-npm-test
          resources: {}
        - name: check-registry
          env:
          - name: CACHE_SUFFIX
            value: "/cache"
          resources: {}
        - name: build-container-build
          resources: {}
        - name: promote-changelog
          resources: {}
        - name: promote-helm-release
          resources: {}
        - name: promote-jx-promote
          resources: {}
```

## Providing an ECR Lifecycle Policy

By default a policy to make images with a tag prefix of 0.0.0- expire after 14 days will be put in place. This prefix is
the default for pull request builds. If a policy exist and the default policy isn't overridden no policy will be put. To
choose another policy change the `check-registry` step to add the `ECR_LIFECYCLE_POLICY` environment variable. See the
[AWS documentation](https://docs.aws.amazon.com/AmazonECR/latest/userguide/LifecyclePolicies.html) for how to write policy.

Below is an example actually showing how to put the default policy in place:

```yaml
        - name: check-registry
          env:
          - name: ECR_LIFECYCLE_POLICY
            value: |-
              {
                "rules": [
                  {
                    "rulePriority": 1,
                    "description": "Expire images older than 14 days",
                    "selection": {
                      "tagStatus": "tagged",
                      "countType": "sinceImagePushed",
                      "tagPrefixList": ["0.0.0-"],
                      "countUnit": "days",
                      "countNumber": 14
                    },
                    "action": {
                      "type": "expire"
                    }
                  }
                ]
              }
          resources: {}
```

If you don't want any policy to be put on the repository you rewrite the step to:

```yaml
        - name: check-registry
          env:
          - name: CREATE_ECR_LIFECYCLE_POLICY
            value: "false"
          resources: {}
```

## Commands

See the [jx-registry command reference](https://github.com/jenkins-x-plugins/jx-registry/blob/master/docs/cmd/jx-registry.md#jx-registry)

