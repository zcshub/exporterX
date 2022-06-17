package main

import (
	"math"
	"runtime"

	factory "exporterX/DataExporter/Factory"
	_ "exporterX/internal/JsonParser"
	_ "exporterX/internal/SnowExporter"

	app "exporterX/DataExporter"
)

func init() {
	useCPU := math.Floor(float64(runtime.NumCPU()+1) / 2)
	runtime.GOMAXPROCS(int(useCPU))
}

func main() {
	excelExporter := app.NewExcelExporter(factory.GetConfigParser(), factory.GetDataExporter(), "conf.json")
	excelExporter.PrintExporterInfo()
	excelExporter.Run()
}
