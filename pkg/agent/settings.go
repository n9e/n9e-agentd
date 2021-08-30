package agent

import (
	"github.com/n9e/n9e-agentd/pkg/api"
	"github.com/n9e/n9e-agentd/pkg/config/settings"
)

type settingsClient struct {
	*EnvSettings
}

func NewSettingsClient(env *EnvSettings) settings.Client {
	return &settingsClient{EnvSettings: env}
}

func (p *settingsClient) Get(key string) (interface{}, error) {
	var output interface{}
	input := &api.SettingInput{Setting: key}

	err := p.ApiCall("GET", "/api/v1/config/{setting}", input, nil, &output)
	return output, err
}

func (p *settingsClient) Set(key string, value string) (bool, error) {

	list, err := p.List()
	if err != nil {
		return false, err
	}

	if err := p.set(key, value); err != nil {
		return false, err
	}

	if setting, ok := list[key]; ok {
		return setting.Hidden, nil
	}

	return false, nil
}

func (p *settingsClient) set(key, value string) error {
	input := &api.SettingInput{
		Setting: key,
		Value:   value,
	}
	return p.ApiCall("POST", "/api/v1/config/{setting}", input, nil, nil)
}

func (p *settingsClient) List() (map[string]settings.RuntimeSettingResponse, error) {
	output := map[string]settings.RuntimeSettingResponse{}

	if err := p.ApiCall("GET", "/api/v1/config/list-runtime", nil, nil, &output); err != nil {
		return nil, err
	}

	return output, nil
}

func (p *settingsClient) FullConfig() (string, error) {
	var resp []byte
	err := p.ApiCall("GET", "/api/v1/config", nil, nil, &resp)
	return string(resp), err
}
