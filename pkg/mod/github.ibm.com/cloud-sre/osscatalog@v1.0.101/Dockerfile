FROM wcp-cto-oss-docker-virtual.artifactory.swg-devops.com/ubi8/ubi-minimal:8.5

RUN microdnf update && microdnf install procps ;

COPY osscatimporter /
COPY osscatpublisher /

# this ENTRYPOINT is overriden by the helm charts when running in Kubernetes pod:
ENTRYPOINT ["echo" , "Entry point needs to be overridden to call osscatimporter or osscatpublisher binary. Container will exit."]
