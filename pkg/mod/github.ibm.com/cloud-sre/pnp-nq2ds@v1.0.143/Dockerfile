FROM wcp-cto-oss-docker-virtual.artifactory.swg-devops.com/ubi8/ubi-minimal:8.5

# Install ps utility
RUN microdnf update && microdnf install procps tzdata

COPY main /

ENTRYPOINT ["/main"]
