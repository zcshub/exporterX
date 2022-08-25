package snowExporter

import (
	conf "exporterX/DataExporter"
	factory "exporterX/DataExporter/Factory"
	tojson "exporterX/internal/ToJson"
	tolua "exporterX/internal/ToLua"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

const (
	SkipRow = "skiprow"
)

var LuaHooker *LuaHookManager

func init() {
	factory.RegisterDataExporter(&SnowExporter{
		logger: log.New(os.Stdout, "[SnowExporter]: ", log.Lshortfile),
	})
	LuaHooker = NewLuaHookManager()
}

type SnowExporter struct {
	logger *log.Logger
}

func (s *SnowExporter) Version() string {
	return "internal/SnowExporter/SnowExporter"
}

func (s *SnowExporter) DoExport(tool string, filePath string, outDir string, dataDef *conf.DataDefine) error {
	if tool != conf.Tool_To_Lua && tool != conf.Tool_To_Json {
		panic("Cannot use tool: " + tool)
	}
	sse := &SnowSingleExporter{
		logger:       log.New(os.Stdout, "["+dataDef.Name+"] ", log.Lshortfile),
		tool:         tool,
		dataDef:      dataDef,
		headType:     make([]*HeadType, 0, 4),
		defaultValue: make([]interface{}, 0, 4),
		header:       make([]*Header, 0, 4),
		data:         make([][]interface{}, 0, 4),
		mapdata:      make(map[string]interface{}),
	}
	sse.DoExport(filePath, outDir)
	return nil
}

type SnowSingleExporter struct {
	logger       *log.Logger
	tool         string
	dataDef      *conf.DataDefine
	headType     []*HeadType
	defaultValue []interface{}
	header       []*Header
	data         [][]interface{}
	mapdata      map[string]interface{}
}

func (s *SnowSingleExporter) DoExport(filePath string, outDir string) error {
	s.logger.Printf("DoExport [%s] from %s %s", s.dataDef.Name, s.dataDef.Excel, s.dataDef.Sheet)
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		s.logger.Printf("excelize open %s got error: %s", filePath, err.Error())
	}
	defer func() {
		if err := f.Close(); err != nil {
			s.logger.Printf("excelize close %s got error: %s", filePath, err.Error())
		}
	}()

	rows, err := f.GetRows(s.dataDef.Sheet)
	if len(rows) <= 3 {
		return nil
	}
	line := 0
	for {
		if rows[line][0] != SkipRow {
			break
		}
		line++
	}

	if s.dataDef.IsMapData {
		// 跳过Map数据的第一行标签
		line++

		for line < len(rows) {
			s.ReadMapping(rows[line])
			line++
		}

		s.WriteMapData(outDir)

	} else {

		s.ReadType(rows[line])
		line++
		s.ReadRange(rows[line])
		line++
		s.ReadHeader(rows[line])
		line++

		for line < len(rows) {
			s.ReadData(rows[line])
			line++
		}

		s.WriteData(outDir)
	}

	return nil
}

func (s *SnowSingleExporter) ReadMapping(row []string) {
	// 每一行数据是 Key  Value  Type的形式
	if len(row) < 3 {
		return
	}
	key := strings.Replace(row[0], " ", "", -1)
	text := strings.Replace(row[1], " ", "", -1)
	keyType, defaultValue := ParseType(row[2])
	header := NewHeader(s.dataDef.Name, key, 1, keyType, defaultValue)
	value := header.ParseData(text)
	s.mapdata[key] = value
}

func (s *SnowSingleExporter) ReadType(row []string) {
	var headtype *HeadType
	var defaultValue interface{}
	for _, v := range row {
		headtype, defaultValue = ParseType(v)
		s.headType = append(s.headType, headtype)
		s.defaultValue = append(s.defaultValue, defaultValue)
	}
	// res1, _ := json.Marshal(s.headType)
	// fmt.Println(string(res1), len(s.headType))
	// res2, _ := json.Marshal(s.defaultValue)
	// fmt.Println(string(res2), len(s.defaultValue))
}

func (s *SnowSingleExporter) ReadRange(row []string) {

}

func (s *SnowSingleExporter) ReadHeader(row []string) {
	for i, v := range row {
		s.header = append(s.header, NewHeader(s.dataDef.Name, v, i, s.headType[i], s.defaultValue[i]))
	}
}

func (s *SnowSingleExporter) ReadData(row []string) {
	var header *Header
	var v interface{}
	var rowData []interface{}
	for i := 0; i < len(s.header); i++ {
		header = s.header[i]
		if i >= len(row) {
			v = header.ParseData("")
		} else {
			v = header.ParseData(row[i])
		}
		if header.IsExportFlag() && v.(bool) == false {
			// 如果设置了导表标签并且该行不导表直接跳过这行数据
			return
		}
		if len(row) == 0 {
			// 该行没有数据直接跳过
			return
		}
		rowData = append(rowData, v)
	}
	// for index, text := range row {
	// 	header = s.header[index]
	// 	v = header.ParseData(text)
	// 	if header.IsExportFlag() && v.(bool) == false {
	// 		// 如果设置了导表标签并且该行不导表直接跳过这行数据
	// 		return
	// 	}
	// 	rowData = append(rowData, v)
	// }
	if len(rowData) > 0 && rowData[0] != nil {
		s.data = append(s.data, rowData)
	}
}

func (s *SnowSingleExporter) WriteMapData(outDir string) error {
	if s.tool == conf.Tool_To_Json {
		toolMan := tojson.NewToJson(s.dataDef.Name, outDir, s.dataDef.RowFile)
		toolMan.WriteMapData(s.mapdata)
	} else if s.tool == conf.Tool_To_Lua {
		toolMan := tolua.NewToLua(s.dataDef.Name, outDir, s.dataDef.RowFile)
		toolMan.WriteMapData(s.mapdata)
	}
	return nil
}

func (s *SnowSingleExporter) WriteData(outDir string) error {
	outputIndexes := make([]int, 0, len(s.header))
	// 先确认导表列
	for index, header := range s.header {
		if header.Needed() && !header.IsExportFlag() {
			outputIndexes = append(outputIndexes, index)
		}
	}

	mapData := make(map[string]interface{})
	for _, row := range s.data {
		rowMap := make(map[string]interface{})
		for _, index := range outputIndexes {
			if index < len(row) {
				rowMap[s.header[index].Key()] = row[index]
			}
		}
		switch row[0].(type) {
		case string:
			mapData[row[0].(string)] = rowMap
		case int:
			mapData[strconv.FormatInt(int64(row[0].(int)), 10)] = rowMap
		default:
			s.logger.Panicf("first column is %T %v, cannot be received.", row[0], row[0])
		}

	}

	if s.tool == conf.Tool_To_Json {
		toolMan := tojson.NewToJson(s.dataDef.Name, outDir, s.dataDef.RowFile)
		toolMan.WriteData(mapData)
	} else if s.tool == conf.Tool_To_Lua {
		toolMan := tolua.NewToLua(s.dataDef.Name, outDir, s.dataDef.RowFile)
		toolMan.WriteData(mapData)
	}
	return nil
}
