package dataExporter

import (
	"fmt"
	"log"
	"path"

	workpool "exporterX/DataExporter/WorkerPool"
)

type ConfigParser interface {
	Version() string
	ParseConfigFile(path string) error
	UnmarshalSection(section string, v interface{}) error
	UnmarshalAll(path string, v interface{}) error
}

type DataExporter interface {
	Version() string
	DoExport(filePath string, outDir string, dataDef DataDefine) error
	WriteData() error
}

func NewExcelExporter(parser *ConfigParser, exporter *DataExporter, confPath string) *ExcelExporter {
	return &ExcelExporter{
		parser:   *parser,
		exporter: *exporter,
		confPath: confPath,
	}
}

type ExcelExporter struct {
	parser   ConfigParser
	exporter DataExporter
	confPath string
	srcDir   string
	outDir   string
	dataDef  []DataDefine
	workPool *workpool.WorkPool
}

func (e *ExcelExporter) PrintExporterInfo() {
	fmt.Println("------------------------------------------------------")
	if e.parser == nil {
		panic("No Parser tool registered")
	}
	fmt.Printf("Parser tool is [%s].\n", e.parser.Version())
	if e.exporter == nil {
		panic("No Exporter tool registered")
	}
	fmt.Printf("Exporter tool is [%s].\n", e.exporter.Version())
	fmt.Println("------------------------------------------------------")
}

func (e *ExcelExporter) BeforeExportData() {

}

func (e *ExcelExporter) AfterExportData() {

}

func (e *ExcelExporter) PrepareExport() error {
	var err error
	var configData = &ExportConf{}
	err = e.parser.UnmarshalAll(e.confPath, &configData)
	if err != nil {
		log.Panicf("ParseConfigFile %s got error: %s", e.confPath, err.Error())
	}

	if exist, err := pathExists(configData.SrcDir); !exist {
		if err != nil {
			log.Panicf("Find src_dir %s got error: %s", configData.SrcDir, err.Error())
		} else {
			log.Panicf("Find src_dir %s got failed", configData.SrcDir)
		}
	}
	e.srcDir = configData.SrcDir

	if err = makePathExists(configData.OutDir); err != nil {
		log.Panicf("Make out_dir got error: %s", err.Error())
	}
	e.outDir = configData.OutDir
	e.dataDef = configData.DataDef

	return err
}

func (e *ExcelExporter) DoExport() {
	tasks := make([]workpool.Task, 0, 16)
	for _, dataDef := range e.dataDef {
		filePath := path.Join(e.srcDir, dataDef.Excel)
		exist, _ := pathExists(filePath)
		if !exist {
			log.Printf("Ignore %s, %s not exist", dataDef.Name, dataDef.Excel)
			continue
		}
		tasks = append(tasks, workpool.Task{
			Id: dataDef.Name,
			F:  func() error { return e.exporter.DoExport(filePath, e.outDir, dataDef) },
		})
	}
	e.workPool = workpool.NewWorkPool(tasks, 4)
	e.workPool.Start()
	e.workPool.Results()
}

func (e *ExcelExporter) Run() {
	e.PrepareExport()

	e.BeforeExportData()

	e.DoExport()

	e.AfterExportData()
}
