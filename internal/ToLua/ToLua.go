package tolua

import (
	"bytes"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
)

var luaFilePrefix = `local _M =
`
var luaFileSuffix1 = `local meta = {
    __index = function(t, k)
        if key[k] == nil {
            return nil
        }
        return t[key[k]]
    end
}
for _, v in pairs(_M) do
    setmetatable(v, meta)
end

return _M`

var luaFileSuffix2 = `return _M`

type ToLua struct {
	logger        *log.Logger
	DataName      string
	OutPath       string
	OneRowOneFile bool
}

func NewToLua(dataName string, outPath string, oneRowOneFile bool) *ToLua {
	logger := log.New(os.Stdout, "["+dataName+"]: ", log.Lshortfile)
	if _, err := os.Stat(path.Join(outPath, dataName+".lua")); err == nil {
		err := os.Remove(path.Join(outPath, dataName+".lua"))
		_ = os.Remove(path.Join(outPath, dataName+".lua.meta"))
		if err != nil {
			logger.Panicf("%s is oneRowOneFile, delete %s got error %s", dataName, path.Join(outPath, dataName+".lua"), err.Error())
		}
	}
	if oneRowOneFile {
		outPath = path.Join(outPath, dataName)
		err := os.RemoveAll(outPath)
		if err != nil {
			logger.Panicf("Prepare directory %s got error: %s", outPath, err.Error())
		}
		if _, err := os.Stat(outPath); errors.Is(err, os.ErrNotExist) {
			err := os.Mkdir(outPath, os.ModePerm)
			if err != nil {
				logger.Panicf("Mkdir %s got error: %s", outPath, err.Error())
			}
		}
	}
	return &ToLua{logger, dataName, outPath, oneRowOneFile}
}

func (t *ToLua) writeLuaFile(filePath string, keysStr string, content string) {
	var buffer bytes.Buffer
	buffer.WriteString(keysStr)
	buffer.WriteString(luaFilePrefix)
	buffer.WriteString(content)
	buffer.WriteString("\n")
	buffer.WriteString(luaFileSuffix1)
	ioutil.WriteFile(filePath, buffer.Bytes(), 0644)
}

func (t *ToLua) writeLuaMapFile(filePath string, content string) {
	var buffer bytes.Buffer
	buffer.WriteString(luaFilePrefix)
	buffer.WriteString(content)
	buffer.WriteString("\n")
	buffer.WriteString(luaFileSuffix2)
	ioutil.WriteFile(filePath, buffer.Bytes(), 0644)
}

func (t *ToLua) optimizeDataForLuaUse(data map[string]interface{}) (map[string]int, map[string]interface{}) {
	newdata := make(map[string]interface{})
	keys := make(map[string]int)
	index := 0
	for _, row := range data {
		for key := range row.(map[string]interface{}) {
			index += 1
			keys[key] = index
		}
		break
	}

	for k, row := range data {
		rowSlice := make([]interface{}, index)
		for key, value := range row.(map[string]interface{}) {
			rowSlice[keys[key]-1] = value
		}
		newdata[k] = rowSlice
	}
	return keys, newdata
}

func (t *ToLua) convertKeysToLuaUse(data map[string]int) string {
	sortedArr := make([]string, len(data))
	for key, index := range data {
		sortedArr[index-1] = key
	}
	var buffer bytes.Buffer
	buffer.WriteString("local key = {")
	for index, key := range sortedArr {
		buffer.WriteString(key)
		buffer.WriteString("=")
		buffer.WriteString(strconv.Itoa(index + 1))
		buffer.WriteString(", ")
	}
	buffer.WriteString("}\n")
	return buffer.String()
}

func (t *ToLua) WriteData(data map[string]interface{}) {
	if t.OneRowOneFile {
		for id, row := range data {
			filePath := path.Join(t.OutPath, id+".lua")
			content := t.convertData(row, 0)
			ioutil.WriteFile(filePath, []byte(content), 0644)
		}
	} else {
		filePath := path.Join(t.OutPath, t.DataName+".lua")
		keys, data := t.optimizeDataForLuaUse(data)
		keysStr := t.convertKeysToLuaUse(keys)
		content := t.convertData(data, 0)
		t.writeLuaFile(filePath, keysStr, content)
	}
}

func (t *ToLua) WriteMapData(data map[string]interface{}) {
	if t.OneRowOneFile {
		for id, row := range data {
			filePath := path.Join(t.OutPath, id+".lua")
			content := t.convertData(map[string]interface{}{id: row}, 0)
			ioutil.WriteFile(filePath, []byte(content), 0644)
		}
	} else {
		filePath := path.Join(t.OutPath, t.DataName+".lua")
		content := t.convertData(data, 0)
		t.writeLuaMapFile(filePath, content)
	}
}

func (t *ToLua) convertData(a interface{}, layer int) string {
	switch a.(type) {
	case int:
		return t.convertInt(a.(int))
	case float64:
		return t.convertFloat(a.(float64))
	case bool:
		return t.convertBool(a.(bool))
	case string:
		return t.convertStr(a.(string))
	case []interface{}:
		return t.convertList(a.([]interface{}), layer)
	case map[string]interface{}:
		return t.convertDict(a.(map[string]interface{}), layer)
	default:
		t.logger.Panicf("%T %v cannot be convert to lua", a, a)
	}
	return ""
}

func (t *ToLua) convertInt(a int) string {
	return strconv.Itoa(a)
}

func (t *ToLua) convertFloat(a float64) string {
	return strconv.FormatFloat(a, 'f', -1, 64)
}

func (t *ToLua) convertBool(a bool) string {
	if a {
		return "true"
	} else {
		return "false"
	}
}

func (t *ToLua) convertStr(a string) string {
	return "\"" + a + "\""
}

func (t *ToLua) convertList(a []interface{}, layer int) string {
	var buffer bytes.Buffer
	buffer.WriteString("{")
	for _, elem := range a {
		buffer.WriteString(t.convertData(elem, layer))
		buffer.WriteString(", ")
	}
	buffer.WriteString("}")
	return buffer.String()
}

func luaTableLayer(layer int) string {
	var buffer bytes.Buffer
	for i := 0; i < layer; i++ {
		buffer.WriteString("\t")
	}
	return buffer.String()
}

func (t *ToLua) convertDict(a map[string]interface{}, layer int) string {
	var buffer bytes.Buffer
	buffer.WriteString("{")
	if layer == 0 {
		buffer.WriteString("\n")
	}
	for k, v := range a {
		// buffer.WriteString(luaTableLayer(layer + 1))
		buffer.WriteString("[")
		buffer.WriteString(t.convertData(k, 0))
		buffer.WriteString("]")
		if layer == 0 {
			buffer.WriteString("\t")
		}
		buffer.WriteString(" = ")
		buffer.WriteString(t.convertData(v, layer+1))
		buffer.WriteString(",")
		if layer == 0 {
			buffer.WriteString("\n")
		}
	}
	// buffer.WriteString(luaTableLayer(layer))
	buffer.WriteString("}")
	return buffer.String()
}
