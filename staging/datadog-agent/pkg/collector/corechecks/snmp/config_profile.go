package snmp

type profileConfigMap map[string]profileConfig

type profileConfig struct {
	DefinitionFile string `json:"definition_file"`
}
