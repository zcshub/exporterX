package snowExporter

import (
	conf "exporterX/DataExporter"
	factory "exporterX/DataExporter/Factory"
	"log"

	"github.com/xuri/excelize/v2"
)

const (
	SkipRow = "skiprow"
)

func init() {
	factory.RegisterDataExporter(&SnowExporter{
		header:       make([]*HeadType, 0, 4),
		defaultValue: make([]interface{}, 0, 4),
	})
}

type SnowExporter struct {
	header       []*HeadType
	defaultValue []interface{}
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

	return nil
}

func (s *SnowExporter) ReadType(row []string) {
	var header *HeadType
	var defaultValue interface{}
	for _, v := range row {
		header, defaultValue = ParseType(v)
		s.header = append(s.header, header)
		s.defaultValue = append(s.defaultValue, defaultValue)
	}
}

func (s *SnowExporter) ReadRange(row []string) {

}

func (s *SnowExporter) ReadHeader(row []string) {

}

func (s *SnowExporter) WriteData() error {
	return nil
}
