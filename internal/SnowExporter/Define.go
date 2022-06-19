package snowExporter

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
)

type HeadType string

const (
	Nil   = HeadType("Nil")
	Int   = HeadType("Int")
	Float = HeadType("Float")
	Str   = HeadType("Str")
	Bool  = HeadType("Bool")
	List  = HeadType("List")
	Dict  = HeadType("Dict")
)

func parseType(v string) (HeadType, interface{}) {
	reg := regexp.MustCompile(`\s*([\w\(\)]+)\s*=?\s*(\w*)\s*`)
	if reg == nil {
		log.Panicf("regexp.MustCompile failed")
	}
	result := reg.FindStringSubmatch(v)
	fmt.Println(result, len(result))
	if len(result) < 3 {
		return HeadType(Nil), nil
	}
	switch HeadType(result[1]) {
	case Int:
		return Int, toInt(result)
	case Float:
		return Float, toFloat(result)

	}
	return HeadType(result[0]), result[1]
}

func toInt(r []string) int {
	if r[2] != "" {
		defaultValue, err := strconv.Atoi(r[2])
		if err != nil {
			log.Panicf("Cannot parse %v", r)
		}
		return defaultValue
	}
	return 0
}

func toFloat(r []string) float64 {
	if r[2] != "" {
		defaultValue, err := strconv.ParseFloat(r[2], 64)
		if err != nil {
			log.Panicf("Cannot parse %v", r)
		}
		return defaultValue
	}
	return 0.0
}

func toStr(r []string) string {
	if r[2] != "" {
		return r[2]
	}
	return ""
}

func toBool(r []string) bool {
	if r[2] != "" {
		if r[2] == "1" {
			return true
		}
		if r[2] == "0" {
			return false
		}
	}
	return false
}

func toList(r []string) []interface{} {
}

func toDict(r []string) map[string]interface{} {

}
