# How is the osscatpublisher built and deployed?

The osscatpublisher is built and deployed using the common OSS CI/CD pipline.

When a PR is merged into the [osscatalog](https://github.ibm.com/cloud-sre/osscatalog) master branch, the CI/CD pipeline will create a PR in the staging branch of the [oss-charts](https://github.ibm.com/cloud-sre/oss-charts) repository to update details such as the imageTag in the [api-osscatalog charts](https://github.ibm.com/cloud-sre/oss-charts/tree/staging/api-osscatalog). This PR will be automatically merged in by the CI/CD pipeline and the osscatpublisher cronjob will automatically be updated in the us-east staging cluster.

To see the osscatpublisher cronjob in the us-east staging cluster, execute the following command:
- kubectl get cronjob -n api
