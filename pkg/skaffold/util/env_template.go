/*
Copyright 2018 The Skaffold Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package util

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"strings"
	"text/template"

	"github.com/sirupsen/logrus"

	"github.com/pkg/errors"
)

var OSEnviron = os.Environ

// recovery will silently swallow all unexpected panics.
func recovery() {
	recover()
}

var SkaffoldFuncMap = template.FuncMap{
	"default": func(arg string, value interface{}) interface{} {
		defer recovery()

		v := reflect.ValueOf(value)
		if v.Kind() != reflect.String {
			return arg
		}

		if len(value.(string)) == 0 {
			return arg
		}

		return value
	},
	"required": func(value interface{}) (interface{}, error) {
		defer recovery()
		if value == nil {
			return nil, errors.New("missing required value")
		}
		return value, nil
	},
}

// ParseEnvTemplate is a simple wrapper to parse an env template
func ParseEnvTemplate(t string) (*template.Template, error) {
	tmpl, err := template.New("envTemplate").Funcs(SkaffoldFuncMap).Parse(t)
	return tmpl, err
}

// ExecuteEnvTemplate executes an envTemplate based on OS environment variables and a custom map
func ExecuteEnvTemplate(envTemplate *template.Template, customMap map[string]string) (string, error) {
	var buf bytes.Buffer
	envMap := map[string]string{}
	for _, env := range OSEnviron() {
		kvp := strings.SplitN(env, "=", 2)
		if len(kvp) != 2 {
			return "", fmt.Errorf("error parsing environment variables, %s does not contain an =", kvp)
		}
		envMap[kvp[0]] = kvp[1]
	}

	for k, v := range customMap {
		envMap[k] = v
	}

	logrus.Debugf("Executing template %v with environment %v", envTemplate, envMap)
	if err := envTemplate.Execute(&buf, envMap); err != nil {
		return "", errors.Wrap(err, "executing template")
	}
	return buf.String(), nil
}
