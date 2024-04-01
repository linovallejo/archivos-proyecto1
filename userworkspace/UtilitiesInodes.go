package UserWorkspace

import (
	"encoding/binary"
	"os"
	Types "proyecto1/types"
	Utilities "proyecto1/utils"
	"strings"
)

// login -user=root -pass=123 -id=A119
func InitSearch(path string, file *os.File, tempSuperblock Types.SuperBlock) int32 {
	//fmt.Println("path:", path)
	// path = "/ruta/nueva"

	// split the path by /
	TempStepsPath := strings.Split(path, "/")
	StepsPath := TempStepsPath[1:]

	// fmt.Println("StepsPath:", StepsPath, "len(StepsPath):", len(StepsPath))
	// for _, step := range StepsPath {
	// 	fmt.Println("step:", step)
	// }

	var Inode0 Types.Inode
	// Read object from bin file
	if err := Utilities.ReadObject(file, &Inode0, int64(tempSuperblock.S_inode_start)); err != nil {
		return -1
	}

	return SarchInodeByPath(StepsPath, Inode0, file, tempSuperblock)
}

func pop(s *[]string) string {
	lastIndex := len(*s) - 1
	last := (*s)[lastIndex]
	*s = (*s)[:lastIndex]
	return last
}

// login -user=root -pass=123 -id=A119
func SarchInodeByPath(StepsPath []string, Inode Types.Inode, file *os.File, tempSuperblock Types.SuperBlock) int32 {
	index := int32(0)
	SearchedName := strings.Replace(pop(&StepsPath), " ", "", -1)

	//fmt.Println("========== SearchedName:", SearchedName)

	// Iterate over i_blocks from Inode
	for _, block := range Inode.I_block {
		if block != -1 {
			if index < 13 {
				//CASO DIRECTO

				var crrFolderBlock Types.DirectoryBlock
				// Read object from bin file
				if err := Utilities.ReadObject(file, &crrFolderBlock, int64(tempSuperblock.S_block_start+block*int32(binary.Size(Types.DirectoryBlock{})))); err != nil {
					return -1
				}

				for _, folder := range crrFolderBlock.B_content {
					// fmt.Println("Folder found======")
					//fmt.Println("Folder | Name:", string(folder.B_name[:]), "B_inodo", folder.B_inodo)

					// fmt.Println("SearchedName:", SearchedName)
					// fmt.Println("folder.B_name:", string(folder.B_name[:]))

					if strings.Contains(string(folder.B_name[:]), SearchedName) {

						//fmt.Println("len(StepsPath)", len(StepsPath), "StepsPath", StepsPath)
						if len(StepsPath) == 0 {
							//fmt.Println("Folder found======")
							return folder.B_inodo
						} else {
							//fmt.Println("NextInode======")
							var NextInode Types.Inode
							// Read object from bin file
							if err := Utilities.ReadObject(file, &NextInode, int64(tempSuperblock.S_inode_start+folder.B_inodo*int32(binary.Size(Types.Inode{})))); err != nil {
								return -1
							}
							return SarchInodeByPath(StepsPath, NextInode, file, tempSuperblock)
						}
					}
				}

			} else {
				//CASO INDIRECTO
			}
		}
		index++
	}

	return 0
}

func GetInodeFileDataOriginal(Inode Types.Inode, file *os.File, tempSuperblock Types.SuperBlock) string {
	index := int32(0)
	// define content as a string
	var content string

	// Iterate over i_blocks from Inode
	for _, block := range Inode.I_block {
		if block != -1 {
			if index < 13 {
				//CASO DIRECTO

				var crrFileBlock Types.FileBlock
				// Read object from bin file
				if err := Utilities.ReadObject(file, &crrFileBlock, int64(tempSuperblock.S_block_start+block*int32(binary.Size(Types.FileBlock{})))); err != nil {
					return ""
				}

				content += string(crrFileBlock.B_content[:])

			} else {
				//CASO INDIRECTO
			}
		}
		index++
	}

	return content
}
