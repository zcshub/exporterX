package dataExporter

type DataDefine struct {
	Name      string `json:"name"`
	Excel     string `json:"excel"`
	Sheet     string `json:"sheet"`
	IsMapData bool   `json:"is_map"`
}

type ExportConf struct {
	SrcDir  string       `json:"src_dir"`
	OutDir  string       `json:"out_dir"`
	DataDef []DataDefine `json:"data_def"`
}
