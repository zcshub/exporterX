package tolua

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"path"
	"sort"
	"strconv"
)

var luaFilePrefix = `local _M =
`
var luaFileSuffix1 = `local meta = {
    __index = function(t, k)
        if key[k] == nil then
            return nil
        end
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
	if _, err := os.Stat(path.Join(outPath, dataName)); err == nil {
		err := os.RemoveAll(path.Join(outPath, dataName))
		if err != nil {
			logger.Panicf("delete %s got error %s", path.Join(outPath, dataName), err.Error())
		}
	}
	if _, err := os.Stat(path.Join(outPath, dataName+".lua")); err == nil {
		err := os.Remove(path.Join(outPath, dataName+".lua"))
		if err != nil {
			logger.Panicf("delete %s got error %s", path.Join(outPath, dataName+".lua"), err.Error())
		}
	}
	if oneRowOneFile {
		dirPath := path.Join(outPath, dataName)
		err := os.RemoveAll(dirPath)
		if err != nil {
			logger.Panicf("Prepare directory %s got error: %s", outPath, err.Error())
		}
		err = os.Mkdir(dirPath, os.ModePerm)
		if err != nil {
			logger.Panicf("Mkdir %s got error: %s", outPath, err.Error())
		}
	}
	return &ToLua{logger, dataName, outPath, oneRowOneFile}
}

func (t *ToLua) writeLuaFile(filePath string, elem ...string) {
	var buffer bytes.Buffer
	for _, e := range elem {
		buffer.WriteString(e)
	}
	err := ioutil.WriteFile(filePath, buffer.Bytes(), 0644)
	if err != nil {
		t.logger.Panicf(err.Error())
	}
}

// 规整数据，将map无序数据转换到打印成lua table的样式
func (t *ToLua) optimizeDataForLuaUse(data map[string]interface{}, keysOrder []string, rowsOrder []string) (map[string]int, []interface{}) {
	newdata := make([]interface{}, 0, len(data))
	keys := make(map[string]int)

	for i, key := range keysOrder {
		if _, exist := keys[key]; exist {
			t.logger.Panicf("Dunplicate key %s", key)
		}
		keys[key] = i + 1
	}

	for _, id := range rowsOrder {
		row := data[id].(map[string]interface{})
		rowSlice := make([]interface{}, len(keysOrder))
		for _, key := range keysOrder {
			rowSlice[keys[key]-1] = row[key]
		}
		newdata = append(newdata, rowSlice)
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

func (t *ToLua) convertIndexesToLuaUse(indexes []string) string {
	if len(indexes) == 0 {
		return ""
	}
	isIntKey := false
	if _, err := strconv.Atoi(indexes[0]); err == nil {
		isIntKey = true
	}

	var buffer bytes.Buffer
	buffer.WriteString("{")
	if isIntKey {
		intIndexes := make([]int, 0, len(indexes))
		for _, index := range indexes {
			intIndex, _ := strconv.Atoi(index)
			intIndexes = append(intIndexes, intIndex)
		}
		sort.Ints(intIndexes)
		for _, index := range intIndexes {
			buffer.WriteString(strconv.Itoa(index))
			buffer.WriteString(",")
		}
	} else {
		sort.Strings(indexes)
		for _, index := range indexes {
			buffer.WriteString("\"")
			buffer.WriteString(index)
			buffer.WriteString("\"")
			buffer.WriteString(",")
		}
	}

	buffer.WriteString("}")
	return buffer.String()
}

func (t *ToLua) WriteData(data map[string]interface{}, keysOrder []string, rowsOder []string, isMap bool) {
	filePath := path.Join(t.OutPath, t.DataName+".lua")
	if !isMap {
		if t.OneRowOneFile {
			indexes := make([]string, 0, len(data))
			for id, row := range data {
				var content bytes.Buffer
				indexes = append(indexes, id)
				filePath = path.Join(t.OutPath, t.DataName, id+".lua")
				content.WriteString(t.convertData(row))
				t.writeLuaFile(filePath, luaFilePrefix, content.String(), "\n", luaFileSuffix2)
			}
			filePath = path.Join(t.OutPath, t.DataName, "index.lua")
			indexesStr := t.convertIndexesToLuaUse(indexes)
			t.writeLuaFile(filePath, luaFilePrefix, indexesStr, "\n", luaFileSuffix2)
		} else {
			var content bytes.Buffer
			keys, formatdata := t.optimizeDataForLuaUse(data, keysOrder, rowsOder)
			content.WriteString("{\n")
			for _, row := range formatdata {
				content.WriteString("[")
				content.WriteString(t.convertData(row.([]interface{})[0]))
				content.WriteString("]\t=\t")
				content.WriteString(t.convertData(row))
				content.WriteString(",\n")
			}
			content.WriteString("}\n")
			keysStr := t.convertKeysToLuaUse(keys)
			t.writeLuaFile(filePath, keysStr, luaFilePrefix, content.String(), "\n", luaFileSuffix1)
		}
	} else {
		sortedKey := make([]string, 0, len(data))
		for k := range data {
			sortedKey = append(sortedKey, k)
		}
		sort.Strings(sortedKey)

		filePath := path.Join(t.OutPath, t.DataName+".lua")
		var content bytes.Buffer
		content.WriteString("{\n")
		for _, key := range sortedKey {
			content.WriteString(key)
			content.WriteString("\t=\t")
			content.WriteString(t.convertData(data[key]))
			content.WriteString(",\n")
		}
		content.WriteString("}\n")
		t.writeLuaFile(filePath, luaFilePrefix, content.String(), "\n", luaFileSuffix2)
	}
}

func (t *ToLua) convertData(a interface{}) string {
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
		return t.convertList(a.([]interface{}))
	case map[string]interface{}:
		return t.convertDict(a.(map[string]interface{}))
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

func (t *ToLua) convertList(a []interface{}) string {
	var buffer bytes.Buffer
	buffer.WriteString("{")
	for _, elem := range a {
		buffer.WriteString(t.convertData(elem))
		buffer.WriteString(", ")
	}
	buffer.WriteString("}")
	return buffer.String()
}

func (t *ToLua) convertDict(a map[string]interface{}) string {
	var buffer bytes.Buffer

	sortedKey := make([]string, 0, len(a))
	for k := range a {
		sortedKey = append(sortedKey, k)
	}
	sort.Strings(sortedKey)

	buffer.WriteString("{")
	for _, k := range sortedKey {
		v := a[k]
		buffer.WriteString("[")
		buffer.WriteString(t.convertData(k))
		buffer.WriteString("]")
		buffer.WriteString(" = ")
		buffer.WriteString(t.convertData(v))
		buffer.WriteString(",")
	}
	buffer.WriteString("}")

	return buffer.String()
}
