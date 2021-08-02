package validation

import (
	"fmt"

	"github.com/DataDog/datadog-agent/pkg/autodiscovery/integration"
	"github.com/DataDog/datadog-agent/pkg/autodiscovery/providers"
	"github.com/DataDog/datadog-agent/pkg/collector/corechecks"
)

func dataValidateJSONConfig(checkName string, data []byte) error {
	config, err := providers.ParseJSONConfig(data)
	if err != nil {
		return err
	}
	config.Name = checkName
	return ValidateConfig(config)
}

func ValidateConfig(config *integration.Config) error {
	factory := corechecks.GetCheckFactory(config.Name)
	if factory == nil {
		return fmt.Errorf("unable to get %s check", config.Name)
	}

	for _, instance := range config.Instances {
		if err := factory().Configure(instance, config.InitConfig, config.Source); err != nil {
			return err
		}
	}
	return nil
}
