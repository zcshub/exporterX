package factory

import (
	exporter "exporterX/DataExporter"
)

var (
	configParser *exporter.ConfigParser
	dataExporter *exporter.DataExporter
)

func RegisterConfigParser(parser exporter.ConfigParser) {
	if parser == nil {
		panic("factory: Register parser is nil")
	}

	configParser = &parser
}

func GetConfigParser() *exporter.ConfigParser {
	return configParser
}

func RegisterDataExporter(exporter exporter.DataExporter) {
	if exporter == nil {
		panic("factory: Register exporter is nil")
	}

	dataExporter = &exporter
}

func GetDataExporter() *exporter.DataExporter {
	return dataExporter
}
