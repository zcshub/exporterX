package snowExporter

import (
	"log"
	"strconv"
	"strings"
)

type Header struct {
	name         string
	index        int
	headType     *HeadType
	defaultValue interface{}
}

func NewHeader(name string, index int, headType *HeadType, defaultValue interface{}) *Header {
	key := strings.Replace(name, " ", "", -1)
	return &Header{
		name:         key,
		index:        index,
		headType:     headType,
		defaultValue: defaultValue,
	}
}

func (h *Header) GetType() *HeadType {
	return h.headType
}

func (h *Header) ParseData(text string) interface{} {
	if h.name == "" {
		return nil
	}
	text = strings.Replace(text, " ", "", -1)
	return parseByHeadType(text, h.headType, h.defaultValue)
}

func parseByHeadType(text string, headType *HeadType, defaultValue interface{}) interface{} {
	switch headType.MetaType {
	case Nil:
		return nil
	case Int:
		return parseInt(text, defaultValue)
	case Float:
		return parseFloat(text, defaultValue)
	case Str:
		return parseStr(text, defaultValue)
	case Bool:
		return parseBool(text, defaultValue)
	case ListPrefix:
		return parseList(text, headType)
	case DictPrefix:
		return parseDict(text, headType)
	default:
		log.Panicf("Cannot understand metaType %s", headType.MetaType)
	}
	return nil
}

func parseInt(text string, defaultValue interface{}) int {
	if text == "" && defaultValue != nil {
		return defaultValue.(int)
	}
	value, err := strconv.Atoi(text)
	if err != nil {
		log.Panicf("Cannot convert %s to Int, %s", text, err.Error())
	}
	return value
}

func parseFloat(text string, defaultValue interface{}) float64 {
	if text == "" && defaultValue != nil {
		return defaultValue.(float64)
	}
	value, err := strconv.ParseFloat(text, 64)
	if err != nil {
		log.Panicf("Cannot convert %s to Float, %s", text, err.Error())
	}
	return value
}

func parseStr(text string, defaultValue interface{}) string {
	if text == "" && defaultValue != nil {
		return defaultValue.(string)
	}
	return text
}

func parseBool(text string, defaultValue interface{}) bool {
	if text == "" && defaultValue != nil {
		return defaultValue.(bool)
	}
	value, err := strconv.ParseBool(text)
	if err != nil {
		log.Panicf("Cannot convert %s to Bool, %s", text, err.Error())
	}
	return value
}

func parseList(text string, headType *HeadType) []interface{} {
	list := make([]interface{}, 0, 2)
	if text[0] != '[' {
		substrings := strings.FieldsFunc(text, func(r rune) bool { return r == ',' || r == ';' })
		for _, substring := range substrings {
			list = append(list, parseByHeadType(substring, headType.ListIn, nil))
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
		list = append(list, parseByHeadType(substring, headType.ListIn, nil))
	}
	return list
}

func parseDict(text string, headType *HeadType) map[string]interface{} {
	dict := make(map[string]interface{})

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
		dict[k] = parseByHeadType(v, headType.DictIn[k], nil)
	}

	return dict
}
