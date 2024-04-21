package types

type DiskPartitionDto struct {
	Type  string
	Start int32
	Name  string
	Id    string
}

type LoginRequestDto struct {
	Username    string
	Password    string
	PartitionId string
}

type ReportDto struct {
	ReportFileName string `json:"reportFilename"`
	DotFileName    string `json:"dotFileName"`
}

type GetReportDto struct {
	DotFileName string `json:"dotFilename"`
}
