package types

type MBR struct {
	MbrTamano        int32
	MbrFechaCreacion [20]byte
	MbrDiskSignature int32
	DskFit           [1]byte
	Partitions       [4]Partition
}
