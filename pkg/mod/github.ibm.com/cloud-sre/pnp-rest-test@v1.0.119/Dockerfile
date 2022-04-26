FROM registry.access.redhat.com/ubi8/ubi-minimal:8.4

COPY server.* /etc/pki/ca-trust/source/anchors/
RUN update-ca-trust force-enable
RUN update-ca-trust extract

# Install ps utility
RUN microdnf update && microdnf install procps bash wget

# Setup jq
RUN wget -qO jq https://github.com/stedolan/jq/releases/download/jq-1.5/jq-linux64 \
  && chmod +x ./jq \
  && cp jq /usr/bin

# Install Postgresql client
RUN wget -O pgdg-redhat-repo-latest.noarch.rpm  https://download.postgresql.org/pub/repos/yum/reporpms/EL-8-x86_64/pgdg-redhat-repo-latest.noarch.rpm
RUN rpm -ivh pgdg-redhat-repo-latest.noarch.rpm
RUN microdnf install postgresql11

COPY main /
ENTRYPOINT ["/main"]
