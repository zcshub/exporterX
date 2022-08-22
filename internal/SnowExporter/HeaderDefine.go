package snowExporter

import (
	"log"
	"os"
	"strconv"
	"strings"
)

type Header struct {
	logger       *log.Logger
	name         string
	index        int
	headType     *HeadType
	defaultValue interface{}
}

func NewHeader(dataName string, name string, index int, headType *HeadType, defaultValue interface{}) *Header {
	key := strings.Replace(name, " ", "", -1)
	return &Header{
		logger:       log.New(os.Stdout, "["+dataName+"]: ", log.Lshortfile),
		name:         key,
		index:        index,
		headType:     headType,
		defaultValue: defaultValue,
	}
}

func (h *Header) Key() string {
	return h.name
}

func (h *Header) Needed() bool {
	return h.name != ""
}

func (h *Header) IsExportFlag() bool {
	return h.name == "ExportTable"
}

func (h *Header) ParseData(text string) interface{} {
	if h.name == "" {
		return nil
	}
	text = strings.Replace(text, " ", "", -1)
	return h.parseByHeadType(text, h.headType, h.defaultValue)
}

func (h *Header) parseByHeadType(text string, headType *HeadType, defaultValue interface{}) interface{} {
	switch headType.MetaType {
	case Nil:
		return nil
	case Int:
		return h.parseInt(text, defaultValue)
	case Float:
		return h.parseFloat(text, defaultValue)
	case Str:
		return h.parseStr(text, defaultValue)
	case Bool:
		return h.parseBool(text, defaultValue)
	case ListPrefix:
		return h.parseList(text, headType)
	case DictPrefix:
		return h.parseDict(text, headType)
	default:
		h.logger.Panicf("Cannot understand metaType %s", headType.MetaType)
	}
	return nil
}

func (h *Header) parseInt(text string, defaultValue interface{}) int {
	if text == "" && defaultValue != nil {
		return defaultValue.(int)
	}
	value, err := strconv.Atoi(text)
	if err != nil {
		h.logger.Panicf("Cannot convert %s to Int, %s", text, err.Error())
	}
	return value
}

func (h *Header) parseFloat(text string, defaultValue interface{}) float64 {
	if text == "" && defaultValue != nil {
		return defaultValue.(float64)
	}
	value, err := strconv.ParseFloat(text, 64)
	if err != nil {
		h.logger.Panicf("Cannot convert %s to Float, %s", text, err.Error())
	}
	return value
}

func (h *Header) parseStr(text string, defaultValue interface{}) string {
	if text == "" && defaultValue != nil {
		return defaultValue.(string)
	}
	return text
}

func (h *Header) parseBool(text string, defaultValue interface{}) bool {
	if text == "" && defaultValue != nil {
		return defaultValue.(bool)
	}
	value, err := strconv.ParseBool(text)
	if err != nil {
		h.logger.Panicf("Cannot convert %s to Bool, %s", text, err.Error())
	}
	return value
}

func (h *Header) parseList(text string, headType *HeadType) []interface{} {
	list := make([]interface{}, 0, 2)
	if text == "" {
		return list
	}

	if text[0] != '[' {
		substrings := strings.FieldsFunc(text, func(r rune) bool { return r == ',' || r == ';' })
		for _, substring := range substrings {
			list = append(list, h.parseByHeadType(substring, headType.ListIn, nil))
		}
		return list
	}

	leftIndex := 0
	num := 0
	substrings := make([]string, 0, 2)
	for index, r := range text {
		if r == '[' {
			if num == 0 {
				leftIndex = index
			}
			num += 1
		} else if r == ']' {
			num -= 1
			if num == 0 {
				substrings = append(substrings, text[leftIndex+1:index])
			}
		}
	}
	for _, substring := range substrings {
		list = append(list, h.parseByHeadType(substring, headType.ListIn, nil))
	}
	return list
}

func (h *Header) parseDict(text string, headType *HeadType) map[string]interface{} {
	dict := make(map[string]interface{})
	if text == "" {
		return dict
	}

	keyIndexes := make([]int, 0, 2)
	for index, r := range text {
		if r == '=' {
			left := index
			for left > 0 {
				left--
				if text[left] == ',' {
					left++
					break
				}
			}
			keyIndexes = append(keyIndexes, left, index)
		}
	}

	kvMap := make(map[string]string)
	for i := 0; i < len(keyIndexes); i += 2 {
		key := text[keyIndexes[i]:keyIndexes[i+1]]
		value := ""
		if i+2 >= len(keyIndexes) {
			value = text[keyIndexes[i+1]+1:]
		} else {
			value = text[keyIndexes[i+1]+1 : keyIndexes[i+2]-1]
		}
		kvMap[key] = value
	}

	for k, v := range kvMap {
		if headType.DictIn[k] == nil {
			log.Panicf("key %s not exist in dict %s", k, headType.Meta)
		}
		dict[k] = h.parseByHeadType(v, headType.DictIn[k], nil)
	}

	return dict
}
