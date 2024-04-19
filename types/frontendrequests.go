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
