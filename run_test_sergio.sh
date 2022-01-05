
export APP_NAME="aaapythonhva"
export DOCKER_REGISTRY="986794016656.dkr.ecr.eu-west-2.amazonaws.com"
export DOCKER_REGISTRY_ORG="ytree-montrose"
export CREATE_ECR_LIFECYCLE_POLICY='false'
export CREATE_ECR_REPOSITORY_POLICY='true'
export ECR_REPOSITORY_POLICY='{"Version":"2008-10-17","Statement":[{"Sid":"CrossAccountPull","Effect":"Allow","Principal":{"AWS":["arn:aws:iam::792438178112:role/poc20210203202039067800000008","arn:aws:iam::161405424480:role/jersey20200514141536577100000008","arn:aws:iam::161405424480:role/jersey220211116181957844700000009"]},"Action":["ecr:BatchCheckLayerAvailability","ecr:BatchGetImage","ecr:GetDownloadUrlForLayer"]}]}'

export JX_LOG_LEVEL=debug

build/jx-registry create