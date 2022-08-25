package main

import (
	"flag"
	"fmt"
	"math"
	"runtime"

	factory "exporterX/DataExporter/Factory"
	_ "exporterX/internal/JsonParser"
	_ "exporterX/internal/SnowExporter"

	app "exporterX/DataExporter"
)

var confPath = flag.String("conf", "conf.json", "配置文件")
var cpuNum = flag.Int("cpu", 8, "使用几核运行")
var srcDir = flag.String("src", "", "数值表路径")
var outDir = flag.String("out", "", "导出路径")

func init() {
	useCPU := math.Floor(float64(runtime.NumCPU()+1) / 2)
	runtime.GOMAXPROCS(int(useCPU))
}

func main() {
	flag.Parse()
	fmt.Printf("conf: %v\n", *confPath)

	optionalConf := &app.OptionalConf{
		CpuNum: *cpuNum,
		SrcDir: *srcDir,
		OutDir: *outDir,
	}
	excelExporter := app.NewExcelExporter(factory.GetConfigParser(), factory.GetDataExporter(), *confPath, optionalConf)
	excelExporter.PrintExporterInfo()
	excelExporter.Run()
}
