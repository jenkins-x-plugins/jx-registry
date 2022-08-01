module github.com/jenkins-x-plugins/jx-registry

require (
	github.com/aws/aws-sdk-go-v2 v1.16.7
	github.com/aws/aws-sdk-go-v2/config v1.15.14
	github.com/aws/aws-sdk-go-v2/service/ecr v1.17.8
	github.com/aws/smithy-go v1.12.0
	github.com/cpuguy83/go-md2man v1.0.10
	github.com/jenkins-x-plugins/jx-gitops v0.7.27
	github.com/jenkins-x/jx-api/v4 v4.5.1
	github.com/jenkins-x/jx-helpers/v3 v3.4.0
	github.com/jenkins-x/jx-logging/v3 v3.0.9
	github.com/pkg/errors v0.9.1
	github.com/sethvargo/go-envconfig v0.3.5
	github.com/spf13/cobra v1.2.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	golang.org/x/sys v0.0.0-20220614162138-6c1b26c55098 // indirect
)

replace (
	// helm dependencies
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d
	github.com/docker/docker => github.com/moby/moby v17.12.0-ce-rc1.0.20200618181300-9dc6525e6118+incompatible

	k8s.io/api => k8s.io/api v0.20.2
	k8s.io/apimachinery => k8s.io/apimachinery v0.20.2
	k8s.io/client-go => k8s.io/client-go v0.20.2
)

go 1.15
