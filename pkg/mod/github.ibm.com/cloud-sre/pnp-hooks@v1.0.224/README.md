[![Build Status](https://wcp-cto-sre-jenkins.swg-devops.com/buildStatus/icon?job=Pipeline/api-pnp-hooks/master)](https://wcp-cto-sre-jenkins.swg-devops.com/job/Pipeline/job/api-pnp-hooks/job/master/)

# pnp-hooks

## Unit tests, coverage and scanning

- pnp-hooks requires rabbitmq to run unit tests.  You therefore
cannot use `gomake test` from the declarative-deployment-tools repo to run the tests locally.
- The local Makefile will start a rabbitmq container needed for the tests.
- Run `make test` to test.
- To view unit test coverage run `go tool cover -html=coverage.out` after running `make test`.
- You should aim to get at least 80% coverage for each package.
- Run `make scan` to run a security scan.
- Unit tests and scan should be successful before submitting a pull request to the master branch.
- You can find results of unit tests in unittest.out.

## Steps to build and push to artifactory

1. Clone the [declarative-deployment-tools](https://github.ibm.com/cloud-sre/declarative-deployment-tools) repository and add it to your system PATH
2. Close this repository to <SOME_PATH>/src/github.ibm.com/cloud-sre
3. cd to the <SOME_PATH>/src/github.ibm.com/cloud-sre/pnp-hooks folder
4. Use the gomake command in declarative-deployment-tools to invoke a target on the shared Makefile
    For example:
    - To get dependencies: gomake dep
    - To build: gomake
    - To build image: gomake image
    - To upload test image to artifactory: gomake deploy

## Steps to manually deploy to Armada

1. Clone the https://github.ibm.com/cloud-sre/oss-charts respository
2. Follow the 'Steps to build and push to artifactory' steps above to build the code and push the test image to artifactory
3. Update the `imageTag` value in /oss-charts/api-pnp-hooks/values.yaml file to be the image tag from step 2 above
4. Open a terminal
5. Add the declarative-deployment-tools to your system path
6. Access IKS dev clusters via Bastion.  See https://github.ibm.com/cloud-sre/ToolsPlatform/wiki/OSS-Bastion-User-Guide---Account-Migration#access-iks-clusters-via-bastion-1 and https://github.ibm.com/cloud-sre/ToolsPlatform/wiki/OSS-Bastion-User-Guide---Account-Migration#ticket-validation-1 for more information.
7.  cd /oss-charts/api-pnp-hooks
8.  Run kdep with the appropriate yaml file. For example: `kdep useast-development-values.yaml`

