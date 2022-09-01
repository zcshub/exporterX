package main

import (
	"flag"
	"log"
	"math"
	"runtime"

	factory "exporterX/DataExporter/Factory"
	_ "exporterX/internal/JsonParser"
	_ "exporterX/internal/SnowExporter"

	app "exporterX/DataExporter"
)

type arrayFlags []string

func (a *arrayFlags) String() string {
	return "arrayFlags string representation"
}

func (a *arrayFlags) Set(value string) error {
	*a = append(*a, value)
	return nil
}

var confPath = flag.String("conf", "conf.json", "配置文件")
var cpuNum = flag.Int("cpu", 0, "使用几核运行")
var srcDir = flag.String("src", "", "数值表路径")
var outDir = flag.String("out", "", "导出路径")

var exportList arrayFlags

func init() {
	useCPU := math.Floor(float64(runtime.NumCPU()+1) / 2)
	runtime.GOMAXPROCS(int(useCPU))
}

func main() {
	// f, _ := os.Create("cpuprofile")
	// defer f.Close()
	// pprof.StartCPUProfile(f)
	flag.Var(&exportList, "name", "DataName to export, access for multi name")
	flag.Parse()
	log.Printf("conf: %v\n", *confPath)

	optionalConf := &app.OptionalConf{
		CpuNum:     *cpuNum,
		SrcDir:     *srcDir,
		OutDir:     *outDir,
		ExportList: exportList,
	}
	excelExporter := app.NewExcelExporter(factory.GetConfigParser(), factory.GetDataExporter(), *confPath, optionalConf)
	excelExporter.PrintExporterInfo()
	excelExporter.Run()
	// pprof.StopCPUProfile()
}
