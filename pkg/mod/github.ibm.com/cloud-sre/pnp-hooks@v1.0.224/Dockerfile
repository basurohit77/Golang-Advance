FROM wcp-cto-oss-docker-virtual.artifactory.swg-devops.com/ubi8/ubi-minimal:8.5

COPY carootcert.crt caintermediatecert.crt /etc/pki/ca-trust/source/anchors/
RUN update-ca-trust force-enable
RUN update-ca-trust extract

# Update Name Switch Service configuration so that the /etc/hosts file in the container
# is used first to resolve hosts to IPs (see https://github.com/golang/go/issues/22846 &
# https://linode.com/docs/networking/dns/using-your-systems-hosts-file/ for details):
# RUN echo "hosts: files dns" > /etc/nsswitch.conf

# Install ps utility
RUN microdnf update && microdnf install procps

COPY main /

ENTRYPOINT ["/main"]
