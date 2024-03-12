package types

type LogicalPartitionInfo struct {
	ExtendedStart int32
	FirstEBR      *EBR
}
