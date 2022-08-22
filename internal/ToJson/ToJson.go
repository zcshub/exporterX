package tojson

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path"
)

type ToJson struct {
	DataName      string
	OutPath       string
	OneRowOneFile bool
}

func NewToJson(dataName string, outPath string, oneRowOneFile bool) *ToJson {
	if _, err := os.Stat(path.Join(outPath, dataName+".json")); err == nil {
		err := os.Remove(path.Join(outPath, dataName+".json"))
		_ = os.Remove(path.Join(outPath, dataName+".json.meta"))
		if err != nil {
			log.Panicf("%s is oneRowOneFile, delete %s got error %s", dataName, path.Join(outPath, dataName+".json"), err.Error())
		}
	}
	if oneRowOneFile {
		outPath = path.Join(outPath, dataName)
		err := os.RemoveAll(outPath)
		if err != nil {
			log.Panicf("Prepare directory %s got error: %s", outPath, err.Error())
		}
		if _, err := os.Stat(outPath); errors.Is(err, os.ErrNotExist) {
			err := os.Mkdir(outPath, os.ModePerm)
			if err != nil {
				log.Panicf("Mkdir %s got error: %s", outPath, err.Error())
			}
		}
	}
	return &ToJson{dataName, outPath, oneRowOneFile}
}

func (t *ToJson) WriteData(data map[string]map[string]interface{}) {
	if t.OneRowOneFile {
		for id, row := range data {
			filePath := path.Join(t.OutPath, id+".json")
			content, _ := json.MarshalIndent(row, "", "\t")
			ioutil.WriteFile(filePath, content, 0644)
		}
	} else {
		filePath := path.Join(t.OutPath, t.DataName+".json")
		content, _ := json.MarshalIndent(data, "", "\t")
		ioutil.WriteFile(filePath, content, 0644)
	}
}
