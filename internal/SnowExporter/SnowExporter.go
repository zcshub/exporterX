package snowExporter

import (
	conf "exporterX/DataExporter"
	factory "exporterX/DataExporter/Factory"
	"fmt"
	"log"

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
	headType     []*HeadType
	defaultValue []interface{}
	header       []*Header
	data         []interface{}
}

func (s *SnowExporter) Version() string {
	return "internal/SnowExporter/SnowExporter"
}

func (s *SnowExporter) DoExport(filePath string, outDir string, dataDef conf.DataDefine) error {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		log.Printf("excelize open %s got error: %s", filePath, err.Error())
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Printf("excelize close %s got error: %s", filePath, err.Error())
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

	fmt.Println(s.data)

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
		s.header = append(s.header, NewHeader(v, i, s.headType[i], s.defaultValue[i]))
	}
}

func (s *SnowExporter) ReadData(row []string) {
	var header *Header
	var v interface{}
	var rowData []interface{}
	for index, text := range row {
		header = s.header[index]
		v = header.ParseData(text)
		rowData = append(rowData, v)
	}
	if len(rowData) > 0 && rowData[0] != nil {
		s.data = append(s.data, rowData)
	}
}

func (s *SnowExporter) WriteData() error {

	return nil
}
