package options

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.ibm.com/cloud-sre/osscatalog/debug"
	"gopkg.in/yaml.v2"
)

// Functions for reading command-line flags from a config file

// ReadConfigFile reads the specified configuration file and initializes all the command-line flags from its contents
func ReadConfigFile(fname string) error {
	file, err := os.Open(fname) // #nosec G304
	if err != nil {
		return debug.WrapError(err, "Cannot open parameters config file %s", fname)
	}
	defer file.Close() // #nosec G307
	rawData, err := ioutil.ReadAll(file)
	if err != nil {
		return debug.WrapError(err, "Error reading parameters config file %s", fname)
	}
	config := make(map[string]string)
	err = yaml.Unmarshal(rawData, config)
	if err != nil {
		return debug.WrapError(err, "Error parsing parameters config file %s", fname)
	}

	debug.Debug(debug.Options, "Parsed config file: %#v", config)

	issues := strings.Builder{}

	for key, val := range config {
		if key == "rw" {
			return fmt.Errorf(`"-rw" flag is not allowed in parameters config file %s`, fname)
		}
		f := flag.Lookup(key)
		if f == nil {
			issues.WriteString(fmt.Sprintf(`Unknown parameter in config file: "%s"="%s"\n`, key, val))
			continue
		}
		if val == "" {
			if getter, ok := f.Value.(flag.Getter); ok {
				if _, ok := getter.Get().(bool); ok {
					val = "true"
				} else {
					issues.WriteString(fmt.Sprintf(`Empty parameter value in config file: "%s"="%s" does not appear to be for a boolean flag (actual type: %T)\n`, key, val, getter.Get()))
					continue
				}
			} else {
				issues.WriteString(fmt.Sprintf(`Empty parameter value in config file: "%s"="%s" and cannot determine if type is boolean (actual value: %#v)\n`, key, val, f.Value))
				continue
			}
		}
		if f.Value.String() != f.DefValue {
			issues.WriteString(fmt.Sprintf(`Parameter in config file "%s"="%s" is already set to a non-default value "%s" (from command line?)\n`, key, val, f.Value.String()))
			continue
		}
		debug.Debug(debug.Options, `Setting parameter "%s"="%s" from config file`, key, val)
		err := f.Value.Set(val)
		if err != nil {
			issues.WriteString(fmt.Sprintf(`Error setting parameter "%s"="%s" from config file: %v\n`, key, val, err))
			continue
		}
	}

	if issues.Len() > 0 {
		return fmt.Errorf(`Errors encountered while processing parameters config file %s: \n%s`, fname, issues.String())
	}

	return nil
}
