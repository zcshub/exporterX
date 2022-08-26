package snowExporter

import (
	"bufio"
	"errors"
	"log"
	"math"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"

	lua "github.com/yuin/gopher-lua"
	"github.com/yuin/gopher-lua/parse"
)

type LuaHookManager struct {
	logger   *log.Logger
	luaState *lua.LState
	lock     sync.RWMutex
	hooks    map[string]*lua.FunctionProto
}

func NewLuaHookManager() *LuaHookManager {
	logger := log.New(os.Stdout, "LuaHookManager:", log.Lshortfile)
	return &LuaHookManager{
		logger:   logger,
		luaState: lua.NewState(),
		hooks:    make(map[string]*lua.FunctionProto),
	}
}

func (m *LuaHookManager) CompileLuaFile(filePath string) (*lua.FunctionProto, error) {
	file, err := os.Open(filePath)
	defer file.Close()
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(file)
	chunk, err := parse.Parse(reader, filePath)
	if err != nil {
		return nil, err
	}
	proto, err := lua.Compile(chunk, filePath)
	if err != nil {
		return nil, err
	}
	return proto, nil
}

func (m *LuaHookManager) CompileLuaString(source string) (*lua.FunctionProto, error) {
	reader := strings.NewReader(source)
	chunk, err := parse.Parse(reader, "<string>")
	if err != nil {
		return nil, err
	}
	proto, err := lua.Compile(chunk, "<string>")
	if err != nil {
		return nil, err
	}
	return proto, nil
}

func (m *LuaHookManager) ConvertLuaValue(functionName string, text string, value lua.LValue) interface{} {
	switch value.(type) {
	case lua.LNumber:
		lnum := lua.LVAsNumber(value)
		num, err := strconv.ParseFloat(lnum.String(), 64)
		if err != nil {
			m.logger.Panicf("%s convert lua number to float64 failed, text: %s error: %s", functionName, text, err)
		}
		if num == math.Floor(num) {
			return int(num)
		} else {
			return num
		}
	case lua.LString:
		return lua.LVAsString(value)
	case lua.LBool:
		return lua.LVAsBool(value)
	case *lua.LTable:
		mapData := make(map[string]interface{})
		listData := make([]interface{}, 0)
		isMap := false
		value.(*lua.LTable).ForEach(func(k lua.LValue, v lua.LValue) {
			if lua.LVAsNumber(k) == lua.LNumber(0) {
				mapData[lua.LVAsString(k)] = m.ConvertLuaValue(functionName, text, v)
				isMap = true
			} else {
				listData = append(listData, m.ConvertLuaValue(functionName, text, v))
			}
		})
		if isMap {
			return mapData
		} else {
			return listData
		}
	default:
		m.logger.Panicf("lua %s return %T %v, cannot understand", functionName, value, value)
		return nil
	}
}

func (m *LuaHookManager) GetHookHandler(dataName string, keyName string) func(text string) interface{} {
	if _, ok := m.hooks[dataName]; !ok {
		path := path.Join("./hook/", dataName+".lua")
		if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
			return nil
		}

		proto, err := m.CompileLuaFile(path)
		if err != nil {
			m.logger.Panicf("%s compile got error: %s", path, err)
		}

		m.luaState.Push(m.luaState.NewFunctionFromProto(proto))
		err = m.luaState.PCall(0, lua.MultRet, nil)
		if err != nil {
			m.logger.Panicf("%s pcall got error: %s", path, err)
		}

		m.lock.Lock()
		m.hooks[dataName] = proto
		m.lock.Unlock()
	}

	functionName := dataName + "_" + keyName
	luaFunc := m.luaState.GetGlobal(functionName)
	if luaFunc == lua.LNil {
		return nil
	}

	return func(text string) interface{} {
		err := m.luaState.CallByParam(lua.P{
			Fn:      luaFunc,
			NRet:    1,
			Protect: true,
		}, lua.LString(text))
		if err != nil {
			m.logger.Panicf("call lua %s with %s got error: %s", functionName, text, err)
		}
		ret := m.luaState.Get(-1)
		m.luaState.Pop(1)

		return m.ConvertLuaValue(functionName, text, ret)
	}
}
