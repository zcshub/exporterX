package jsonParser

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"

	factory "exporterX/DataExporter/Factory"
)

func init() {
	factory.RegisterConfigParser(&JsonParser{
		jsonData: make(map[string]json.RawMessage),
	})
}

type JsonParser struct {
	jsonData map[string]json.RawMessage
}

func (p *JsonParser) Version() string {
	return "internal/ConfParser/ConfParser"
}

func (p *JsonParser) ParseConfigFile(path string) error {
	confData, err := ioutil.ReadFile(path)
	if err != nil {
		return errors.New(fmt.Sprintf("Read file error: %s", err.Error()))
	}

	if err := json.Unmarshal(confData, &p.jsonData); err != nil {
		return errors.New(fmt.Sprintf("No json format error: %s", err.Error()))
	}

	return nil
}

func (p *JsonParser) UnmarshalSection(section string, v interface{}) error {
	sectionConf, ok := p.jsonData[section]
	if !ok {
		return errors.New(fmt.Sprintf("Config file has no section %s", section))
	}

	return json.Unmarshal(sectionConf, v)
}

func (p *JsonParser) UnmarshalAll(path string, v interface{}) error {
	confData, err := ioutil.ReadFile(path)
	if err != nil {
		return errors.New(fmt.Sprintf("Read file error: %s", err.Error()))
	}
	if err := json.Unmarshal(confData, v); err != nil {
		return errors.New(fmt.Sprintf("No json format error: %s", err.Error()))
	}

	return nil
}
