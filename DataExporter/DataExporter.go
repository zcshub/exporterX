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
	Init()
	DoExport(n int, tool string, filePath string, outDir string, dataDef *DataDefine) (string, error)
	SetCpuNum(int)
	AfterExport()
}

type OptionalConf struct {
	CpuNum     int
	SrcDir     string
	OutDir     string
	ExportList []string
}

func NewExcelExporter(parser *ConfigParser, exporter *DataExporter, confPath string, opt *OptionalConf) *ExcelExporter {
	return &ExcelExporter{
		parser:     *parser,
		exporter:   *exporter,
		confPath:   confPath,
		cpuNum:     opt.CpuNum,
		srcDir:     opt.SrcDir,
		outDir:     opt.OutDir,
		exportList: opt.ExportList,
	}
}

type ExcelExporter struct {
	parser     ConfigParser
	exporter   DataExporter
	confPath   string
	srcDir     string
	outDir     string
	tool       string
	cpuNum     int
	dataDef    []DataDefine
	exportList []string
	workPool   *workpool.WorkPool
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

func (e *ExcelExporter) PrepareExport() error {
	var err error
	var configData = &ExportConf{}
	err = e.parser.UnmarshalAll(e.confPath, &configData)
	if err != nil {
		log.Panicf("ParseConfigFile %s got error: %s", e.confPath, err.Error())
	}

	e.tool = configData.Tool
	if e.srcDir == "" {
		e.srcDir = configData.SrcDir
	}
	if e.outDir == "" {
		e.outDir = configData.OutDir
	}
	if e.cpuNum == 0 {
		e.cpuNum = configData.CpuNum
	}
	e.exporter.SetCpuNum(e.cpuNum)

	if e.exportList != nil && len(e.exportList) > 0 {
		e.dataDef = make([]DataDefine, 0, len(e.exportList))
		for _, dataDef := range configData.DataDef {
			for _, name := range e.exportList {
				if dataDef.Name == name {
					e.dataDef = append(e.dataDef, dataDef)
				}
			}
		}
	} else {
		e.dataDef = configData.DataDef
	}
	e.exportList = nil

	if exist, err := pathExists(configData.SrcDir); !exist {
		if err != nil {
			log.Panicf("Find src_dir %s got error: %s", configData.SrcDir, err.Error())
		} else {
			log.Panicf("Find src_dir %s got failed", configData.SrcDir)
		}
	}

	if err = makePathExists(configData.OutDir); err != nil {
		log.Panicf("Make out_dir got error: %s", err.Error())
	}

	fmt.Printf("cpu: %v\n", e.cpuNum)
	fmt.Printf("src: %v\n", e.srcDir)
	fmt.Printf("out: %v\n", e.outDir)

	return err
}

func (e *ExcelExporter) BeforeExportData() {
	e.exporter.Init()
}

func (e *ExcelExporter) DoExport() {
	tasks := make([]workpool.Task, 0, 16)
	ignores := make([]string, 0, 2)
	for _, dataDef := range e.dataDef {
		filePath := path.Join(e.srcDir, dataDef.Excel)
		exist, _ := pathExists(filePath)
		if !exist {
			ignores = append(ignores, fmt.Sprintf("%s, %s not found", dataDef.Name, dataDef.Excel))
			continue
		}
		dataDefCp := dataDef
		tasks = append(tasks, workpool.Task{
			Id: dataDef.Name,
			F:  func(i int) (string, error) { return e.exporter.DoExport(i, e.tool, filePath, e.outDir, &dataDefCp) },
		})
	}
	e.workPool = workpool.NewWorkPool(tasks, e.cpuNum)
	e.workPool.Start()
	results := e.workPool.Results()

	for _, ignore := range ignores {
		log.Printf(ignore)
	}

	for _, dataDef := range e.dataDef {
		if _, ok := results[dataDef.Name]; ok {
			delete(results, dataDef.Name)
		} else {
			log.Printf("%s export failed", dataDef.Name)
		}
	}
}

func (e *ExcelExporter) AfterExportData() {
	log.Println("==================================")
	log.Println("Next is programers's Data analysis")
	e.exporter.AfterExport()
}

func (e *ExcelExporter) Run() {
	e.PrepareExport()

	e.BeforeExportData()

	e.DoExport()

	e.AfterExportData()
}
