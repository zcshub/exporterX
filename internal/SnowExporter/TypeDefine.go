package snowExporter

import (
	"log"
	"regexp"
	"strconv"
	"strings"
)

var HeadDefine *regexp.Regexp

var ListDefine *regexp.Regexp
var DictDefine *regexp.Regexp
var EnumDefine *regexp.Regexp
var FuncDefine *regexp.Regexp

func init() {
	// 所有head类型匹配规则
	HeadDefine = regexp.MustCompile(`([\w\(\):,]+)=?(.*)`)
	if HeadDefine == nil {
		log.Panicf("regexp.MustCompile HeadDefine failed")
	}
	// List类型匹配规则
	ListDefine = regexp.MustCompile(`^List\((.+)\)$`)
	if ListDefine == nil {
		log.Panicf("regexp.MustCompile ListDefine failed")
	}
	// Dict类型匹配规则
	DictDefine = regexp.MustCompile(`^Dict\((.+)\)$`)
	if DictDefine == nil {
		log.Panicf("regexp.MustCompile DictDefine failed")
	}
	// Enum类型匹配规则
	EnumDefine = regexp.MustCompile(`^Enum\((.+)\)$`)
	if EnumDefine == nil {
		log.Panicf("regexp.MustCompile EnumDefine failed")
	}
	// Func类型匹配规则
	FuncDefine = regexp.MustCompile(`^Func\((.+)\)$`)
	if FuncDefine == nil {
		log.Panicf("regexp.MustCompile FuncDefine failed")
	}
}

const (
	Nil   = "Nil"
	Int   = "Int"
	Float = "Float"
	Str   = "Str"
	Bool  = "Bool"
)

const (
	ListPrefix = "List"
	DictPrefix = "Dict"
	EnumPrefix = "Enum"
	FuncPrefix = "Func"
)

const (
	FuncFunc1  = "Func1" // 针对一元二次表达式
	FuncSwitch = "Switch"
	FuncAwaken = "Awaken"
)

type HeadType struct {
	Meta     string
	MetaType string
	ListIn   *HeadType
	DictIn   map[string]*HeadType
	EnumIn   map[string]int
}

func (h *HeadType) IsNil() bool {
	return h.Meta == Nil
}

func NewHeadType(name string, nameType string) *HeadType {
	return &HeadType{
		Meta:     name,
		MetaType: nameType,
	}
}

func ParseType(v string) (*HeadType, interface{}) {
	v = strings.Replace(v, " ", "", -1)
	v = strings.Replace(v, "\n", "", -1)
	result := HeadDefine.FindStringSubmatch(v)

	if len(result) < 3 {
		return NewHeadType(Nil, ""), nil
	}
	switch result[1] {
	case Int:
		return NewHeadType(Int, Int), toInt(result[2])
	case Float:
		return NewHeadType(Float, Float), toFloat(result[2])
	case Str:
		return NewHeadType(Str, Str), toStr(result[2])
	case Bool:
		return NewHeadType(Bool, Bool), toBool(result[2])
	default:
		return parseSecondType(result)
	}
}

func parseSecondType(r []string) (*HeadType, interface{}) {
	if len(r[1]) >= len(ListPrefix) && r[1][0:len(ListPrefix)] == ListPrefix {
		result := ListDefine.FindStringSubmatch(r[1])
		if len(result) < 2 {
			log.Panicf("cannot parse %v %v", r, result)
		}

		listType, _ := ParseType(result[1])
		ht := &HeadType{
			Meta:     r[1],
			MetaType: ListPrefix,
			ListIn:   listType,
		}
		return ht, nil
	}
	if len(r[1]) >= len(DictPrefix) && r[1][0:len(DictPrefix)] == DictPrefix {
		result := DictDefine.FindStringSubmatch(r[1])
		if len(result) < 2 {
			log.Panicf("cannot parse %v %v", r, result)
		}

		kvReg := regexp.MustCompile(`(\w+):([\w\(\)]+)`)
		kvList := kvReg.FindAllStringSubmatch(result[1], -1)
		// log.Println(kvList)
		dictIn := make(map[string]*HeadType)
		for _, kv := range kvList {
			dictIn[kv[1]], _ = ParseType(kv[2])
		}
		ht := &HeadType{
			Meta:     r[1],
			MetaType: DictPrefix,
			DictIn:   dictIn,
		}
		return ht, nil
	}
	if len(r[1]) >= len(EnumPrefix) && r[1][0:len(EnumPrefix)] == EnumPrefix {
		result := EnumDefine.FindStringSubmatch(r[1])
		if len(result) < 2 {
			log.Panicf("cannot parse %v %v", r, result)
		}

		kvReg := regexp.MustCompile(`(\w+):([\w\(\)]+)`)
		kvList := kvReg.FindAllStringSubmatch(result[1], -1)
		// log.Println(kvList)
		enumIn := make(map[string]int)
		for _, kv := range kvList {
			enumIn[kv[1]] = toInt(kv[2])
		}
		ht := &HeadType{
			Meta:     r[1],
			MetaType: EnumPrefix,
			EnumIn:   enumIn,
		}
		return ht, nil
	}
	if len(r[1]) >= len(FuncPrefix) && r[1][0:len(FuncPrefix)] == FuncPrefix {
		return &HeadType{
			Meta:     r[1],
			MetaType: FuncPrefix,
		}, toInt(r[2])
	}

	log.Panicf("Unknown type %s", r[1])
	return NewHeadType(Int, Int), 0
}

func toInt(r string) int {
	if r != "" {
		defaultValue, err := strconv.Atoi(r)
		if err != nil {
			log.Panicf("Cannot parse %v", r)
		}
		return defaultValue
	}
	return 0
}

func toFloat(r string) float64 {
	if r != "" {
		defaultValue, err := strconv.ParseFloat(r, 64)
		if err != nil {
			log.Panicf("Cannot parse %v", r)
		}
		return defaultValue
	}
	return 0.0
}

func toStr(r string) string {
	if r != "" {
		if len(r) != len(strings.Trim(r, "\"")) {
			log.Panicf("Str type no quites needed, got %v", r)
		}
		return r
	}
	return ""
}

func toBool(r string) bool {
	if r != "" {
		defaultValue, err := strconv.ParseBool(r)
		if err != nil {
			log.Panicf("Cannot parse %v", r)
		}
		return defaultValue
	}
	return false
}
