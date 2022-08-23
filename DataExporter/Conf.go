package dataExporter

type DataDefine struct {
	Name      string `json:"name"`
	Excel     string `json:"excel"`
	Sheet     string `json:"sheet"`
	RowFile   bool   `json:"rowFile"`
	IsMapData bool   `json:"isMap"`
}

type ExportConf struct {
	Tool    string       `json:"tool"`
	CpuNum  int          `json:"cpu_num"`
	SrcDir  string       `json:"src_dir"`
	OutDir  string       `json:"out_dir"`
	DataDef []DataDefine `json:"data_def"`
}

const (
	Tool_To_Json = "to_json"
	Tool_To_Lua  = "to_lua"
)
