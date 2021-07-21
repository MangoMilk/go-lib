package dwarfloader

import (
	"encoding/json"
	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path"
)

func Load(configPath string, out interface{}) error {
	fileExt := path.Ext(configPath)
	//fmt.Println(fileExt)

	configByte, readErr := ioutil.ReadFile(configPath)
	if readErr != nil {
		return readErr
	}
	var unmarshalErr error
	switch fileExt {
	case ".yaml", ".yml":
		unmarshalErr = yaml.Unmarshal(configByte, out)
		break
	case ".toml":
		unmarshalErr = toml.Unmarshal(configByte, out)
		break
	case ".json":
		unmarshalErr = json.Unmarshal(configByte, out)
		break
	}

	if unmarshalErr != nil {
		return unmarshalErr
	}

	return nil
}
