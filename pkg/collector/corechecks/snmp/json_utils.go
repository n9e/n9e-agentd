package snmp

import (
	"encoding/json"
	"strconv"
)

func (a *StringArray) UnmarshalJSON(b []byte) error {
	var multi []string
	if err := json.Unmarshal(b, &multi); err != nil {
		var single string
		err := json.Unmarshal(b, &single)
		if err != nil {
			return err
		}
		*a = []string{single}
	} else {
		*a = multi
	}
	return nil
}

func (n *Number) UnmarshalJSON(b []byte) error {
	var integer int
	err := json.Unmarshal(b, &integer)
	if err != nil {
		var str string
		err := json.Unmarshal(b, &str)
		if err != nil {
			return err
		}
		num, err := strconv.Atoi(str)
		if err != nil {
			return err
		}
		*n = Number(num)
	} else {
		*n = Number(integer)
	}
	return nil
}

func (a *metricTagConfigList) UnmarshalJSON(b []byte) error {
	var multi []metricTagConfig
	err := json.Unmarshal(b, &multi)
	if err != nil {
		var tags []string
		err := json.Unmarshal(b, &tags)
		if err != nil {
			return err
		}
		multi = []metricTagConfig{}
		for _, tag := range tags {
			multi = append(multi, metricTagConfig{symbolTag: tag})
		}
	}
	*a = multi
	return nil
}
