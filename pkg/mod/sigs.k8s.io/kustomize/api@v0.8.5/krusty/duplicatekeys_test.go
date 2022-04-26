package krusty_test

import (
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

func TestDuplicateKeys(t *testing.T) {
	th := kusttest_test.MakeHarness(t)
	th.WriteK(".", `
resources:
- resources.yaml
`)
	th.WriteF("resources.yaml", `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: podinfo
spec:
  selector:
    matchLabels:
      app: podinfo
  template:
    spec:
      containers:
      - name: podinfod
        image: ghcr.io/stefanprodan/podinfo:5.0.3 
        command:
        - ./podinfo
        env:
        - name: PODINFO_UI_COLOR
          value: "#34577c"
        env:
          - name: PODINFO_UI_COLOR
            value: "#34577c"
`)
	th.RunWithErr(".", th.MakeDefaultOptions())
}
