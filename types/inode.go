package types

import "time"

type Inode struct {
	i_uid   int
	i_gid   int
	i_size  int
	i_atime time.Time
	i_ctime time.Time
	i_mtime time.Time
	i_block [15]int
	i_type  byte
	i_perm  [3]byte
}
