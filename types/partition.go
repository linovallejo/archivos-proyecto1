package types

type Partition struct {
	Status          byte
	Type            byte
	Fit             byte
	Start           int64
	Size            int64
	Name            [16]byte
	PartCorrelative int
	PartId          [4]byte
}
