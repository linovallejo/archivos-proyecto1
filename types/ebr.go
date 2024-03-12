package types

type EBR struct {
	PartMount [1]byte
	PartFit   [1]byte
	PartStart int32
	PartSize  int32
	PartNext  *EBR
	PartName  [16]byte
}
