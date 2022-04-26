# OCI Registry As Storage

[![GitHub Actions status](https://github.com/deislabs/oras/workflows/build/badge.svg)](https://github.com/deislabs/oras/actions?query=workflow%3Abuild)
[![Go Report Card](https://goreportcard.com/badge/github.com/deislabs/oras)](https://goreportcard.com/report/github.com/deislabs/oras)
[![GoDoc](https://godoc.org/github.com/deislabs/oras?status.svg)](https://godoc.org/github.com/deislabs/oras)

![ORAS](./oras.png)

[Registries are evolving as Cloud Native Artifact Stores](https://stevelasker.blog/2019/01/25/cloud-native-artifact-stores-evolve-from-container-registries/). To enable this goal, Microsoft has donated ORAS as a means to enable various client libraries with a way to push [OCI Artifacts][artifacts] to [OCI Conformant](https://github.com/opencontainers/oci-conformance) registries.

ORAS is both a [CLI](#oras-cli) for initial testing and a [Go Module](#oras-go-module) to be included with your CLI, enabling a native experience: `myclient push artifacts.azurecr.io/myartifact:1.0 ./mything.thang`

## Table of Contents

- [ORAS Background](#oras-background)
- [Supported Registries](./implementors.md#registries-supporting-artifacts)
- [Artifacts Implementing ORAS](./implementors.md#artifact-types-using-oras)
- [Getting Started](#getting-started)
- [ORAS CLI](#oras-cli)
- [ORAS Go Module](#oras-go-module)
- [Contributing](#contributing)
- [Maintainers](./MAINTAINERS)

## ORAS Background

- [OCI Image Support Comes to Open Source Docker Registry](https://opencontainers.org/posts/blog/2018-10-11-oci-image-support-comes-to-open-source-docker-registry/)
- [Registries Are Evolving as Cloud Native Artifact Stores](https://stevelasker.blog/2019/01/25/cloud-native-artifact-stores-evolve-from-container-registries/)
- [OCI Adopts Artifacts Project](https://opencontainers.org/posts/blog/2019-09-10-new-oci-artifacts-project/)
- [GitHub: OCI Artifacts Project](https://github.com/opencontainers/artifacts)

## Getting Started

[Select from one the registries that support OCI Artifacts](./implementors.md). Each registry identifies how they support authentication.

## ORAS CLI

ORAS is both a [CLI](#oras-cli) for initial testing and a [Go Module](#oras-go-module) to be included with your CLI, enabling a native experience: `myclient push artifacts.azurecr.io/myartifact:1.0 ./mything.thang`

### CLI Installation

- Install `oras` using [GoFish](https://gofi.sh/):

  ```sh
  gofish install oras
  ==> Installing oras...
  🐠  oras 0.11.1: installed in 65.131245ms
  ```

- Install from the latest [release artifacts](https://github.com/deislabs/oras/releases):

  - Linux

    ```sh
    curl -LO https://github.com/deislabs/oras/releases/download/v0.11.1/oras_0.11.1_linux_amd64.tar.gz
    mkdir -p oras-install/
    tar -zxf oras_0.11.1_*.tar.gz -C oras-install/
    mv oras-install/oras /usr/local/bin/
    rm -rf oras_0.11.1_*.tar.gz oras-install/
    ```

  - macOS

    ```sh
    curl -LO https://github.com/deislabs/oras/releases/download/v0.11.1/oras_0.11.1_darwin_amd64.tar.gz
    mkdir -p oras-install/
    tar -zxf oras_0.11.1_*.tar.gz -C oras-install/
    mv oras-install/oras /usr/local/bin/
    rm -rf oras_0.11.1_*.tar.gz oras-install/
    ```

  - Windows

    Add `%USERPROFILE%\bin\` to your `PATH` environment variable so that `oras.exe` can be found.

    ```sh
    curl.exe -sLO  https://github.com/deislabs/oras/releases/download/v0.11.1/oras_0.11.1_windows_amd64.tar.gz
    tar.exe -xvzf oras_0.11.1_windows_amd64.tar.gz
    mkdir -p %USERPROFILE%\bin\
    copy oras.exe %USERPROFILE%\bin\
    set PATH=%USERPROFILE%\bin\;%PATH%
    ```

  - Docker Image

    A public Docker image containing the CLI is available on [GitHub Container Registry](https://github.com/orgs/deislabs/packages/container/package/oras):

    ```sh
    docker run -it --rm -v $(pwd):/workspace ghcr.io/deislabs/oras:v0.11.1 help
    ```

    > Note: the default WORKDIR  in the image is `/workspace`.

### ORAS Authentication

Run `oras login` in advance for any private registries. By default, this will store credentials in `~/.docker/config.json` *(same file used by the docker client)*. If you have previously authenticated to a registry using `docker login`, the credentials will be reused.

Use the `-c`/`--config` option to specify an alternate location.

> While ORAS leverages the local docker client config store, ORAS does NOT have a dependency on Docker Desktop running or being installed. ORAS can be used independently of a local docker daemon.

`oras` also accepts explicit credentials via options, for example,

```sh
oras pull -u username -p password myregistry.io/myimage:latest
```

See [Supported Registries](./implementors.md) for registry specific authentication usage.

### Pushing Artifacts with Single Files

Pushing single files involves referencing the unique artifact type and at least one file.
Defining an Artifact uses the `config.mediaType` as the unique artifact type. If a config object is provided, the `mediaType` extension defines the config filetype. If a `null` config is passed, the config extension must be removed.

See: [Defining a Unique Artifact Type](https://github.com/opencontainers/artifacts/blob/master/artifact-authors.md#defining-a-unique-artifact-type)

The following sample defines a new Artifact Type of **Acme Rocket**, using `application/vnd.acme.rocket.config` as the `manifest.config.mediaType`.

- Create a sample file to push/pull as an artifact

  ```sh
  echo "hello world" > artifact.txt
  ```

- Push the sample file to the registry:

  ```sh
  oras push localhost:5000/hello-artifact:v1 \
  --manifest-config /dev/null:application/vnd.acme.rocket.config \
  ./artifact.txt
  ```

- Pull the file from the registry:

  ```sh
  rm -f artifact.txt # first delete the file
  oras pull localhost:5000/hello-artifact:v1
  cat artifact.txt  # should print "hello world"
  ```

- Push the sample file, with a layer `mediaType`, using the format `filename[:type]`:

  ```sh
  oras push localhost:5000/hello-artifact:v2 \
  --manifest-config /dev/null:application/vnd.acme.rocket.config \
    artifact.txt:text/plain
  ```

### Pushing Artifacts with Config Files

The [OCI distribution-spec][distribution-spec] provides for storing optional config objects. These can be used by the artifact to determine how or where to process and/or route the blobs. When providing a config object, the version and file type is required.

- Create a config file

  ```sh
  echo "{\"name\":\"foo\",\"value\":\"bar\"}" > config.json
  ```

- Push an the artifact, with the `config.json` file

  ```sh
  oras push localhost:5000/hello-artifact:v2 \
  --manifest-config config.json:application/vnd.acme.rocket.config.v1+json \
    artifact.txt:text/plain
  ```

### Pushing Artifacts with Multiple Files

Just as container images support multiple "layers" represented as blobs, ORAS supports pushing multiple layers. The layer type is up to the artifact author. You may push `.tar` representing a collection of files, individual files like `.yaml`, `.txt` or whatever your artifact should be represented as. Each layer type should have a `mediaType` representing the type of blob content.
In this example, we'll push a collection of files.

- A single file (`artifact.txt`) that represents overview content that might be displayed as a repository overview
- A collection of files (`docs/*`) that represents detailed content. When specifying a directory, ORAS will automatically tar the contents.

See [OCI Artifacts][artifacts] for more details.

- Create additional blobs

  ```sh
  mkdir docs
  echo "Docs on this artifact" > ./docs/readme.md
  echo "More content for this artifact" > ./docs/readme2.md
  ```

- Create a config file, referencing the entry doc file

  ```sh
  echo "{\"doc\":\"readme.md\"}" > config.json
  ```

- Push multiple files with different `mediaTypes`:

  ```sh
  oras push localhost:5000/hello-artifact:v2 \
    --manifest-config config.json:application/vnd.acme.rocket.config.v1+json \
    artifact.txt:text/plain \
    ./docs/:application/vnd.acme.rocket.docs.layer.v1+tar
  ```

- The push would generate the following manifest:

  ```json
  {
    "schemaVersion": 2,
    "config": {
      "mediaType": "application/vnd.acme.rocket.config.v1+json",
      "digest": "sha256:7aa5d0dee9a3a73c81db4356cf7aa5666e175d96e68ee763eeb977bd7ba59ee5",
      "size": 20
    },
    "layers": [
      {
        "mediaType": "text/plain",
        "digest": "sha256:a948904f2f0f479b8f8197694b30184b0d2ed1c1cd2a1ec0fb85d299a192a447",
        "size": 12,
        "annotations": {
          "org.opencontainers.image.title": "artifact.txt"
        }
      },
      {
        "mediaType": "application/vnd.acme.rocket.docs.layer.v1+tar",
        "digest": "sha256:20ae7d51e2365405e6942439140d897548e1d4610db60354aef8a5ce1f1699a7",
        "size": 196,
        "annotations": {
          "io.deis.oras.content.digest": "sha256:4329ea6c620ca4e9cedc5f5e8040432114cb5d64fc53107ea870db149e3d2b9e",
          "io.deis.oras.content.unpack": "true",
          "org.opencontainers.image.title": "docs"
        }
      }
    ]
  }
  ```

### Pulling Artifacts

Pulling artifacts involves specifying the content addressable artifact, along with the type of artifact.
> See: [Issue 130](https://github.com/deislabs/oras/issues/130) for eliminating `-a` and `--media-type`

```sh
oras pull localhost:5000/hello-artifact:v2 -a
```

## ORAS Go Module

While the ORAS CLI provides a great way to get started, and test registry support for [OCI Artifacts][artifacts], the primary experience enables a native experience for your artifact of choice. Using the ORAS Go Module, you can develop your own push/pull experience: `myclient push artifacts.azurecr.io/myartifact:1.0 ./mything.thang`

The package `github.com/deislabs/oras/pkg/oras` can quickly be imported in other Go-based tools that
wish to benefit from the ability to store arbitrary content in container registries.

### ORAS Go Module Example

[Source](examples/simple_push_pull.go)

```go
package main

import (
	"context"
	"fmt"

	"github.com/deislabs/oras/pkg/content"
	"github.com/deislabs/oras/pkg/oras"

	"github.com/containerd/containerd/remotes/docker"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	ref := "localhost:5000/oras:test"
	fileName := "hello.txt"
	fileContent := []byte("Hello World!\n")
	customMediaType := "my.custom.media.type"

	ctx := context.Background()
	resolver := docker.NewResolver(docker.ResolverOptions{})

	// Push file(s) w custom mediatype to registry
	memoryStore := content.NewMemoryStore()
	desc := memoryStore.Add(fileName, customMediaType, fileContent)
	pushContents := []ocispec.Descriptor{desc}
	fmt.Printf("Pushing %s to %s...\n", fileName, ref)
	desc, err := oras.Push(ctx, resolver, ref, memoryStore, pushContents)
	check(err)
	fmt.Printf("Pushed to %s with digest %s\n", ref, desc.Digest)

	// Pull file(s) from registry and save to disk
	fmt.Printf("Pulling from %s and saving to %s...\n", ref, fileName)
	fileStore := content.NewFileStore("")
	defer fileStore.Close()
	allowedMediaTypes := []string{customMediaType}
	desc, _, err = oras.Pull(ctx, resolver, ref, fileStore, oras.WithAllowedMediaTypes(allowedMediaTypes))
	check(err)
	fmt.Printf("Pulled from %s with digest %s\n", ref, desc.Digest)
	fmt.Printf("Try running 'cat %s'\n", fileName)
}
```

## Contributing

Want to reach the ORAS community and developers?
We're very interested in feedback and contributions for other artifacts.

[Join us](https://slack.cncf.io/) at [CNCF Slack](https://cloud-native.slack.com) under the **#oras** channel

[artifacts]:            https://github.com/opencontainers/artifacts
[distribution-spec]:    https://github.com/opencontainers/distribution-spec/
