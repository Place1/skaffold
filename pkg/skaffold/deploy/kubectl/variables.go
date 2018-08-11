package kubectl

import (
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/util"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

func (l *ManifestList) SubstituteVariables(values map[string]string) (ManifestList, error) {
	for i, value := range values {
		tmpl, err := util.ParseEnvTemplate(value)
		if err != nil {
			return nil, errors.Wrap(err, "parsing manifest values template")
		}

		rendered, err := util.ExecuteEnvTemplate(tmpl, values)
		if err != nil {
			return nil, errors.Wrap(err, "substituting variables in manifest values template")
		}

		values[i] = rendered
	}

	var updatedManifests ManifestList

	for _, manifest := range *l {
		// replace vars
		tmpl, err := util.ParseEnvTemplate(string(manifest[:]))
		if err != nil {
			return nil, errors.Wrap(err, "parsing manifest template")
		}

		rendered, err := util.ExecuteEnvTemplate(tmpl, values)
		if err != nil {
			return nil, errors.Wrap(err, "substituting variables in manifest template")
		}

		m := make(map[interface{}]interface{})
		if err := yaml.Unmarshal([]byte(rendered), &m); err != nil {
			return nil, errors.Wrap(err, "reading kubernetes YAML")
		}

		if len(m) == 0 {
			continue
		}

		updatedManifest, err := yaml.Marshal(m)
		if err != nil {
			return nil, errors.Wrap(err, "marshalling yaml")
		}

		updatedManifests = append(updatedManifests, updatedManifest)
	}
	return updatedManifests, nil
}
