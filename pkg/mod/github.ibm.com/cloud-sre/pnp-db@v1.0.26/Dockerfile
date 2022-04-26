FROM alpine:3.7
RUN apk --no-cache add ca-certificates
COPY carootcert.crt caintermediatecert.crt /usr/local/share/ca-certificates/
RUN update-ca-certificates
RUN ln -s /dev/stdout pnpdb.log &&  ln -s /dev/null pnpdb_lm.log
COPY main servicemanifest.json /
ENTRYPOINT ["/main"]