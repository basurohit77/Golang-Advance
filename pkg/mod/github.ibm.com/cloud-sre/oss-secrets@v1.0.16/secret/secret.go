package secret

import (
	"io/ioutil"
	"log"
	"os"

	"path/filepath"
)

// The secret package includes functions that can be used to retrieve the value of secret from an in-memory volume or environment variable if not
// found in the in-memory volume. If the secret is only found in an environment variable a warning is logged as the use of environment variables
// for secrets is not allowed per the SEC048
// (https://github.ibm.com/cloud-docs-internal/service-framework/blob/master/SOURCE_CHEATSHEET/SEC/sec048_cheatsheet.md) cheetsheet.
//
// Note that the location of the in-memory volume is hardcoded to enforce consistency (see defaultMountPath for in-memory volume location).

const (
	// defaultMountPath is the default volume folder where secrets will be looked for. Note that this folder was picked because it matches the
	// mount path of the Vault Agent Sidecar Injector (https://www.vaultproject.io/docs/platform/k8s/injector) that we eventually want to move to.
	defaultMountPath = "/vault/secrets"
)

var (
	// mountPath is the volume folder where secrets files will be looked for (note that the mount path is not changable outside this package on
	// purpose at this time to enforce consistency; may be relaxed in the future if there is a need)
	mountPath = defaultMountPath
)

// Get returns the value of the provided secret name. If the secret is not found, an empty string is returned. If the secret can not be found in
// the in-memory volume, the environment variables are checked.
func Get(name string) string {
	value, found := Lookup(name)
	if !found {
		return ""
	}
	return value
}

// Lookup returns the value of the provided secret name and a boolean indicating whether the secret was found. If the secret can not be found in
// the in-memory volume, the environment variables are checked.
func Lookup(name string) (string, bool) {
	value, found := lookupWithPath(name, mountPath)
	if !found {
		// Not found in the in-memory volume, fallback and check if an environment variable is found:
		value, found = os.LookupEnv(name)
		if found {
			log.Printf("WARN: Secret '%s' was not found in the in-memory volume, but was found as an environment variable", name)
		}
	}
	return value, found
}

// lookupWithPath returns the value of the provided secret name at the provided path and a boolean indicating whether the secret was found. This
// function only looks in the path and does NOT check environment variables.
func lookupWithPath(name string, path string) (string, bool) {
	if name == "" {
		return "", false // no name to lookup
	}

	fullPath := filepath.Join(filepath.Clean(path), filepath.Clean(name))
	data, err := ioutil.ReadFile(filepath.Clean(fullPath))
	if err != nil {
		return "", false // not found
	}

	return string(data), true // success
}
