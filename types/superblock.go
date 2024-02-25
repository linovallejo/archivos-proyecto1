package types

import "time"

type SuperBlock struct {
	s_filesystem_type   int
	s_inodes_count      int
	s_blocks_count      int
	s_free_blocks_count int
	s_free_inodes_count int
	s_mtime             time.Time
	s_umtime            time.Time
	s_mnt_count         int
	s_magic             int
	s_inode_size        int
	s_block_size        int
	s_first_ino         int
	s_first_blo         int
	s_bm_inode_start    int
	s_bm_block_start    int
	s_inode_start       int
	s_block_start       int
}
