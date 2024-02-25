package types

type EBR struct {
	part_mount byte
	part_fit   byte
	part_start int
	part_size  int
	part_next  int
	part_name  [16]byte
}
