package krusty_test

import (
	"fmt"
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestIssue3489Simplified(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK(".", `
namespace: kube-system
resources:
- aa
- bb
`)
	th.WriteK("aa", `
resources:
- ../base
`)
	th.WriteK("bb", `
resources:
- ../base
nameSuffix: -private
`)
	th.WriteK("base", `
resources:
- deployment.yaml
- serviceAccount.yaml
`)
	th.WriteF("base/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDep
spec:
  template:
    spec:
      serviceAccountName: mySvcAcct
      containers:
      - name: whatever
        image: k8s.gcr.io/governmentCheese
`)
	th.WriteF("base/serviceAccount.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: mySvcAcct
`)
	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDep
  namespace: kube-system
spec:
  template:
    spec:
      containers:
      - image: k8s.gcr.io/governmentCheese
        name: whatever
      serviceAccountName: mySvcAcct
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: mySvcAcct
  namespace: kube-system
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myDep-private
  namespace: kube-system
spec:
  template:
    spec:
      containers:
      - image: k8s.gcr.io/governmentCheese
        name: whatever
      serviceAccountName: mySvcAcct-private
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: mySvcAcct-private
  namespace: kube-system
`)
}

func TestIssue3489(t *testing.T) {
	const assets = `{
	"tenantId": "XXXXX-XXXXXX-XXXXX-XXXXXX-XXXXXX",
	"subscriptionId": "XXXXX-XXXXXX-XXXXX-XXXXXX-XXXXXX",
	"resourceGroup": "DNS-EUW-XXX-RG",
	"useManagedIdentityExtension": true,
	"userAssignedIdentityID": "XXXXX-XXXXXX-XXXXX-XXXXXX-XXXXXX"
}
`
	th := kusttest_test.MakeHarness(t)
	th.WriteK(".", `
namespace: kube-system
resources:
- external-dns
- external-dns-private
`)
	th.WriteK("external-dns", `
resources:
- ../base
commonLabels:
  app: external-dns
  instance: public
images:
- name: k8s.gcr.io/external-dns/external-dns
  newName: xxx.azurecr.io/external-dns
  newTag: v0.7.4_sylr.1
- name: quay.io/sylr/external-dns
  newName: xxx.azurecr.io/external-dns
  newTag: v0.7.4_sylr.1
secretGenerator:
- name: azure-config-file
  behavior: replace
  files:
  - assets/azure.json
patches:
- target:
    group: apps
    version: v1
    kind: Deployment
    name: external-dns
  patch: |-
    - op: replace
      path: /spec/template/spec/containers/0/args
      value:
      - --txt-owner-id="aks"
      - --txt-prefix=external-dns-
      - --source=service
      - --provider=azure
      - --registry=txt
      - --domain-filter=dev.company.com
`)

	th.WriteF("external-dns/assets/azure.json", assets)
	th.WriteK("external-dns-private", `
resources:
- ../base
nameSuffix: -private
commonLabels:
  app: external-dns
  instance: private
images:
- name: k8s.gcr.io/external-dns/external-dns
  newName: xxx.azurecr.io/external-dns
  newTag: v0.7.4_sylr.1
- name: quay.io/sylr/external-dns
  newName: xxx.azurecr.io/external-dns
  newTag: v0.7.4_sylr.1
secretGenerator:
- name: azure-config-file
  behavior: replace
  files:
  - assets/azure.json
patches:
- target:
    group: apps
    version: v1
    kind: Deployment
    name: external-dns
  patch: |-
    - op: replace
      path: /spec/template/spec/containers/0/args
      value:
      - --txt-owner-id="aks"
      - --txt-prefix=external-dns-private-
      - --source=service
      - --provider=azure-private-dns
      - --registry=txt
      - --domain-filter=static.company.az
`)
	th.WriteF("external-dns-private/assets/azure.json", assets)
	th.WriteK("base", `
resources:
- clusterrole.yaml
- clusterrolebinding.yaml
- deployment.yaml
- serviceaccount.yaml
commonLabels:
  app: external-dns
  instance: public
images:
- name: k8s.gcr.io/external-dns/external-dns
  newName: quay.io/sylr/external-dns
  newTag: v0.7.4-73-g00a9a0c7
secretGenerator:
- name: azure-config-file
  files:
  - assets/azure.json
`)
	th.WriteF("base/assets/azure.json", assets)
	th.WriteF("base/clusterrolebinding.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: external-dns-viewer
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: external-dns
subjects:
- kind: ServiceAccount
  name: external-dns
`)
	th.WriteF("base/clusterrole.yaml", `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: external-dns
rules:
- apiGroups: ['']
  resources: ['endpoints', 'pods', 'services', 'nodes']
  verbs: ['get', 'watch', 'list']
- apiGroups: ['extensions', 'networking.k8s.io']
  resources: ['ingresses']
  verbs: ['get', 'watch', 'list']
`)
	th.WriteF("base/deployment.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: external-dns
spec:
  strategy:
    type: Recreate
  selector:
    matchLabels: {}
  template:
    metadata: {}
    spec:
      serviceAccountName: external-dns
      containers:
      - name: external-dns
        image: k8s.gcr.io/external-dns/external-dns
        args:
        - --domain-filter=""
        - --txt-owner-id=""
        - --txt-prefix=external-dns-
        - --source=service
        - --provider=azure
        - --registry=txt
        resources: {}
        volumeMounts:
        - name: azure-config-file
          mountPath: /etc/kubernetes
          readOnly: true
      volumes:
      - name: azure-config-file
        secret:
          secretName: azure-config-file
`)
	th.WriteF("base/serviceaccount.yaml", `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: external-dns
`)
	opts := th.MakeDefaultOptions()
	m := th.Run(".", opts)
	expFmt := `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  labels:
    app: external-dns
    instance: public
  name: external-dns
rules:
- apiGroups:
  - ""
  resources:
  - endpoints
  - pods
  - services
  - nodes
  verbs:
  - get
  - watch
  - list
- apiGroups:
  - extensions
  - networking.k8s.io
  resources:
  - ingresses
  verbs:
  - get
  - watch
  - list
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  labels:
    app: external-dns
    instance: public
  name: external-dns-viewer
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: external-dns
subjects:
- kind: ServiceAccount
  name: external-dns
  namespace: kube-system
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: external-dns
    instance: public
  name: external-dns
  namespace: kube-system
spec:
  selector:
    matchLabels:
      app: external-dns
      instance: public
  strategy:
    type: Recreate
  template:
    metadata:
      labels:
        app: external-dns
        instance: public
    spec:
      containers:
      - args:
        - --txt-owner-id="aks"
        - --txt-prefix=external-dns-
        - --source=service
        - --provider=azure
        - --registry=txt
        - --domain-filter=dev.company.com
        image: xxx.azurecr.io/external-dns:v0.7.4_sylr.1
        name: external-dns
        resources: {}
        volumeMounts:
        - mountPath: /etc/kubernetes
          name: azure-config-file
          readOnly: true
      serviceAccountName: external-dns
      volumes:
      - name: azure-config-file
        secret:
          secretName: azure-config-file-%s
---
apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    app: external-dns
    instance: public
  name: external-dns
  namespace: kube-system
---
apiVersion: v1
data:
  azure.json: %s
kind: Secret
metadata:
  labels:
    app: external-dns
    instance: public
  name: azure-config-file-%s
  namespace: kube-system
type: Opaque
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  labels:
    app: external-dns
    instance: private
  name: external-dns-private
rules:
- apiGroups:
  - ""
  resources:
  - endpoints
  - pods
  - services
  - nodes
  verbs:
  - get
  - watch
  - list
- apiGroups:
  - extensions
  - networking.k8s.io
  resources:
  - ingresses
  verbs:
  - get
  - watch
  - list
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  labels:
    app: external-dns
    instance: private
  name: external-dns-viewer-private
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: external-dns-private
subjects:
- kind: ServiceAccount
  name: external-dns-private
  namespace: kube-system
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: external-dns
    instance: private
  name: external-dns-private
  namespace: kube-system
spec:
  selector:
    matchLabels:
      app: external-dns
      instance: private
  strategy:
    type: Recreate
  template:
    metadata:
      labels:
        app: external-dns
        instance: private
    spec:
      containers:
      - args:
        - --txt-owner-id="aks"
        - --txt-prefix=external-dns-private-
        - --source=service
        - --provider=azure-private-dns
        - --registry=txt
        - --domain-filter=static.company.az
        image: xxx.azurecr.io/external-dns:v0.7.4_sylr.1
        name: external-dns
        resources: {}
        volumeMounts:
        - mountPath: /etc/kubernetes
          name: azure-config-file
          readOnly: true
      serviceAccountName: external-dns-private
      volumes:
      - name: azure-config-file
        secret:
          secretName: azure-config-file-private-%s
---
apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    app: external-dns
    instance: private
  name: external-dns-private
  namespace: kube-system
---
apiVersion: v1
data:
  azure.json: %s
kind: Secret
metadata:
  labels:
    app: external-dns
    instance: private
  name: azure-config-file-private-%s
  namespace: kube-system
type: Opaque
`
	const (
		nameHashKyaml = "66cc4224mm"
		contentKyaml  = `|
    ewoJInRlbmFudElkIjogIlhYWFhYLVhYWFhYWC1YWFhYWC1YWFhYWFgtWFhYWFhYIiwKCS
    JzdWJzY3JpcHRpb25JZCI6ICJYWFhYWC1YWFhYWFgtWFhYWFgtWFhYWFhYLVhYWFhYWCIs
    CgkicmVzb3VyY2VHcm91cCI6ICJETlMtRVVXLVhYWC1SRyIsCgkidXNlTWFuYWdlZElkZW
    50aXR5RXh0ZW5zaW9uIjogdHJ1ZSwKCSJ1c2VyQXNzaWduZWRJZGVudGl0eUlEIjogIlhY
    WFhYLVhYWFhYWC1YWFhYWC1YWFhYWFgtWFhYWFhYIgp9Cg==`
		nameHashApiMach = "g2k4bkgt4d"
		// nolint: lll
		contentApiMach = `ewoJInRlbmFudElkIjogIlhYWFhYLVhYWFhYWC1YWFhYWC1YWFhYWFgtWFhYWFhYIiwKCSJzdWJzY3JpcHRpb25JZCI6ICJYWFhYWC1YWFhYWFgtWFhYWFgtWFhYWFhYLVhYWFhYWCIsCgkicmVzb3VyY2VHcm91cCI6ICJETlMtRVVXLVhYWC1SRyIsCgkidXNlTWFuYWdlZElkZW50aXR5RXh0ZW5zaW9uIjogdHJ1ZSwKCSJ1c2VyQXNzaWduZWRJZGVudGl0eUlEIjogIlhYWFhYLVhYWFhYWC1YWFhYWC1YWFhYWFgtWFhYWFhYIgp9Cg==`
	)
	th.AssertActualEqualsExpected(
		m,
		// TODO(#3304): DECISION - kyaml better; not a bug.
		opts.IfApiMachineryElseKyaml(
			fmt.Sprintf(expFmt,
				nameHashApiMach,
				contentApiMach, nameHashApiMach,
				nameHashApiMach,
				contentApiMach, nameHashApiMach),
			fmt.Sprintf(expFmt,
				nameHashKyaml,
				contentKyaml, nameHashKyaml,
				nameHashKyaml,
				contentKyaml, nameHashKyaml)))
}

func TestEmptyFieldSpecValue(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK(".", `
generators:
- generators.yaml
configurations:
- kustomizeconfig.yaml
`)
	th.WriteF("generators.yaml", `
apiVersion: builtin
kind: ConfigMapGenerator
metadata:
  name: secret-example
labels:
  app.kubernetes.io/name: secret-example
literals:
- this_is_a_secret_name=
`)
	th.WriteF("kustomizeconfig.yaml", `
nameReference:
- kind: Secret
  version: v1
  fieldSpecs:
  - path: data/this_is_a_secret_name
    kind: ConfigMap
`)
	m := th.Run(".", th.MakeDefaultOptions())
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
data:
  this_is_a_secret_name: ""
kind: ConfigMap
metadata:
  name: secret-example-7hf4fh868h
`)
}
