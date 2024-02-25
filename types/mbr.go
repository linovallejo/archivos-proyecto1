package types

type MBR struct {
	MbrTamano        int64
	MbrFechaCreacion [20]byte
	MbrDiskSignature int64
	DskFit           byte
	Partitions       [4]Partition
}
