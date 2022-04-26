# oss-secrets

This repository is a library written in Golang that can be used to lookup the values of secrets at runtime. A Get and Lookup API are provided. See the /secret/secret.go file for details on the APIs provided.

To use this library for a service running in Kube, ensure that the Deployment Kube chart configuration of your service contains a volume mount with a /vault/secrets path and an associated volume with more or more secrets like the following:

```
      ...
      containers:
          ...
          volumeMounts:
          - name: secrets
            mountPath: "/vault/secrets"
            readOnly: true
      ...
      volumes:
      - name: secrets
        projected:
          sources:
          - secret:
              name: "{{ .Values.kdep.secrets.SECRET1 }}"
              items:
              - key: value
                path: SECRET1
          - secret:
              name: "{{ .Values.kdep.secrets.SECRET2 }}"
              items:
              - key: value
                path: SECRET2
          ...
```

In the above example, the `path` attribute is the name of the secret that you would call the Get or Lookup API with.
