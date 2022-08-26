package snowExporter

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

type Header struct {
	logger       *log.Logger
	name         string
	index        int
	headType     *HeadType
	defaultValue interface{}
	hooker       func(text string) interface{}
	luaState     *lua.LState
}

func NewHeader(n int, dataName string, name string, index int, headType *HeadType, defaultValue interface{}) *Header {
	key := strings.Replace(name, " ", "", -1)
	hooker := LuaHooker.GetHookHandler(dataName, name)
	return &Header{
		logger:       log.New(os.Stdout, "["+dataName+"]: ", log.Lshortfile),
		name:         key,
		index:        index,
		headType:     headType,
		defaultValue: defaultValue,
		hooker:       hooker,
		luaState:     LuaStates[n],
	}
}

func (h *Header) SetLoggerPrefix(p string) {
	h.logger.SetPrefix(p)
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
	text = strings.Replace(text, "\n", "", -1)
	return h.parseByHeadType(text, h.headType, h.defaultValue)
}

func (h *Header) parseByHeadType(text string, headType *HeadType, defaultValue interface{}) interface{} {
	if h.hooker != nil {
		return h.hooker(text)
	}
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
	case EnumPrefix:
		return h.parseEnum(text, headType)
	case FuncPrefix:
		return h.parseFunc(text, headType)
	default:
		h.logger.Panicf("Cannot understand metaType %s when meet %s", headType.MetaType, text)
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
		if _, ok := headType.DictIn[k]; !ok {
			h.logger.Panicf("key %s not exist in dict %s", k, headType.Meta)
		}
		dict[k] = h.parseByHeadType(v, headType.DictIn[k], nil)
	}

	return dict
}

func (h *Header) parseEnum(text string, headType *HeadType) int {
	if text == "" {
		return 0
	}
	if _, ok := headType.EnumIn[text]; ok {
		return headType.EnumIn[text]
	}
	h.logger.Panicf("%s not in Enum %v", text, headType.EnumIn)
	return 0
}

var CoefficientsOfUnaryQuadraticExpressionFormat = `
	local expression = function(x) return %s end
	local a, b, c = 0, 0, 0
	c = expression(0)
	local a_add_b = expression(1) - c
	local a_sub_b = expression(-1) - c
	return (a_add_b + a_sub_b)/2, (a_add_b - a_sub_b)/2, c
`

func (h *Header) parseFunc(text string, headType *HeadType) []interface{} {
	if text == "" {
		return []interface{}{1, h.defaultValue.(int)}
	}

	textLength := len(text)

	if textLength >= len(FuncSwitch) && text[:len(FuncSwitch)] == FuncSwitch {
		switchParam := h.parseList(text[len(FuncSwitch)+1:], &HeadType{
			Meta:     "List(List(Str))",
			MetaType: "List",
			ListIn: &HeadType{
				Meta:     "List(Str)",
				MetaType: "List",
				ListIn: &HeadType{
					Meta:     "Str",
					MetaType: "Str",
				},
			},
		})
		r := make([]interface{}, 0, len(switchParam))
		for _, group := range switchParam {
			group := group.([]interface{})
			first, err := strconv.ParseFloat(group[0].(string), 64)
			if err != nil {
				h.logger.Panicf("%s parse failed %s", text, err)
			}
			second, err := strconv.ParseFloat(group[1].(string), 64)
			if err != nil {
				h.logger.Panicf("%s parse failed %s", text, err)
			}
			expression := group[2].(string)
			err = h.luaState.DoString(fmt.Sprintf(CoefficientsOfUnaryQuadraticExpressionFormat, expression))
			if err != nil {
				h.logger.Panicf("%s run lua failed %s", text, err)
			}
			a, b, c := h.luaState.Get(-3), h.luaState.Get(-2), h.luaState.Get(-1)
			af, bf, cf := float64(lua.LVAsNumber(a)), float64(lua.LVAsNumber(b)), float64(lua.LVAsNumber(c))
			h.luaState.Pop(3)
			r = append(r, []interface{}{first, second, []interface{}{af, bf, cf}})
		}
		return []interface{}{
			2,
			r,
		}
	}

	if textLength >= len(FuncAwaken) && text[:len(FuncAwaken)] == FuncAwaken {
		return []interface{}{
			3,
			h.parseList(text[len(FuncAwaken)+1:], &HeadType{
				Meta:     "List(Float)",
				MetaType: "List",
				ListIn: &HeadType{
					Meta:     "Float",
					MetaType: "Float",
				},
			}),
		}
	}

	// 针对一元二次表达式的，必须是3个返回，a*x*x+b*x+c, 返回a,b,c
	if textLength >= len(FuncFunc1) && text[:len(FuncFunc1)] == FuncFunc1 {
		expression := text[len(FuncFunc1)+1:]
		err := h.luaState.DoString(fmt.Sprintf(CoefficientsOfUnaryQuadraticExpressionFormat, expression))
		if err != nil {
			h.logger.Panicf("%s run lua failed %s", text, err)
		}
		a, b, c := h.luaState.Get(-3), h.luaState.Get(-2), h.luaState.Get(-1)
		h.luaState.Pop(3)

		return []interface{}{
			4,
			[]interface{}{float64(lua.LVAsNumber(a)), float64(lua.LVAsNumber(b)), float64(lua.LVAsNumber(c))},
		}
	}

	value, err := strconv.ParseFloat(text, 64)
	if err != nil {
		h.logger.Panicf("%s parse to float64 failed %s", text, err)
	}
	return []interface{}{1, value}
}
