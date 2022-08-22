package snowExporter

import (
	conf "exporterX/DataExporter"
	factory "exporterX/DataExporter/Factory"
	tojson "exporterX/internal/ToJson"
	"log"
	"os"
	"strconv"

	"github.com/xuri/excelize/v2"
)

const (
	SkipRow = "skiprow"
)

func init() {
	factory.RegisterDataExporter(&SnowExporter{
		headType:     make([]*HeadType, 0, 4),
		defaultValue: make([]interface{}, 0, 4),
		header:       make([]*Header, 0, 4),
	})
}

type SnowExporter struct {
	dataDef      *conf.DataDefine
	logger       *log.Logger
	headType     []*HeadType
	defaultValue []interface{}
	header       []*Header
	data         [][]interface{}
}

func (s *SnowExporter) Version() string {
	return "internal/SnowExporter/SnowExporter"
}

func (s *SnowExporter) DoExport(filePath string, outDir string, dataDef *conf.DataDefine) error {
	s.dataDef = dataDef
	s.logger = log.New(os.Stdout, "["+dataDef.Name+"]: ", log.Lshortfile)
	s.logger.Printf("DoExport [%s]", dataDef.Name)
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		s.logger.Printf("excelize open %s got error: %s", filePath, err.Error())
	}
	defer func() {
		if err := f.Close(); err != nil {
			s.logger.Printf("excelize close %s got error: %s", filePath, err.Error())
		}
	}()

	rows, err := f.GetRows(dataDef.Sheet)
	if len(rows) <= 4 {
		return nil
	}
	line := 0
	for {
		if rows[line][0] != SkipRow {
			break
		}
		line++
	}

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

	return nil
}

func (s *SnowExporter) ReadType(row []string) {
	var header *HeadType
	var defaultValue interface{}
	for _, v := range row {
		header, defaultValue = ParseType(v)
		s.headType = append(s.headType, header)
		s.defaultValue = append(s.defaultValue, defaultValue)
	}
	// res1, _ := json.Marshal(s.headType)
	// fmt.Println(string(res1), len(s.headType))
	// res2, _ := json.Marshal(s.defaultValue)
	// fmt.Println(string(res2), len(s.defaultValue))
}

func (s *SnowExporter) ReadRange(row []string) {

}

func (s *SnowExporter) ReadHeader(row []string) {
	for i, v := range row {
		s.header = append(s.header, NewHeader(s.dataDef.Name, v, i, s.headType[i], s.defaultValue[i]))
	}
}

func (s *SnowExporter) ReadData(row []string) {
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

func (s *SnowExporter) WriteData(outDir string) error {
	outputIndexes := make([]int, 0, len(s.header))
	// 先确认导表列
	for index, header := range s.header {
		if header.Needed() && !header.IsExportFlag() {
			outputIndexes = append(outputIndexes, index)
		}
	}

	mapData := make(map[string]map[string]interface{})
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

	toolMan := tojson.NewToJson(s.dataDef.Name, outDir, s.dataDef.RowFile)
	toolMan.WriteData(mapData)
	return nil
}
