package userworkspace

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	Fdisk "proyecto1/commands/fdisk"
	Types "proyecto1/types"
	Utilities "proyecto1/utils"
	"strconv"
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
	fmt.Println("======Start SEARCHINODEBYPATH======")
	fmt.Println(StepsPath)
	for _, step := range StepsPath {
		fmt.Println("step:", step)
	}
	fmt.Println("======End SEARCHINODEBYPATH======")

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

func GetFileBlockDataOriginal(Inode Types.Inode, file *os.File, tempSuperblock Types.SuperBlock) *Types.FileBlock {
	index := int32(0)
	// define content as a string

	// Iterate over i_blocks from Inode
	for _, block := range Inode.I_block {
		if block != -1 {
			if index < 13 {
				//CASO DIRECTO

				var crrFileBlock Types.FileBlock
				// Read object from bin file
				if err := Utilities.ReadObject(file, &crrFileBlock, int64(tempSuperblock.S_block_start+block*int32(binary.Size(Types.FileBlock{})))); err == nil {
					return &crrFileBlock
				} else {
					fmt.Printf("Error: %s\n", err)
				}

			} else {
				//CASO INDIRECTO
			}
		}
		index++
	}

	return nil
}

func ReturnFileContents(pathUsersFile string, partitionId string, diskFileName string) ([]string, error) {
	var fileContents []string = []string{}

	var TempMBR *Types.MBR
	// Leer el MBR existente
	TempMBR, err := Fdisk.ReadMBR(diskFileName)
	if err != nil {
		//fmt.Println("Error leyendo el MBR:", err)
		return nil, fmt.Errorf("Error leyendo el MBR")
	}

	var partitionStatusStr string = ""
	var index int = 0
	// Iterate over the partitions
	for i := 0; i < 4; i++ {
		// fmt.Println("Partition id:", string(TempMBR.Partitions[i].Id[:]))
		// fmt.Println("Partition name:", Utils.CleanPartitionName(TempMBR.Partitions[i].Name[:]))
		//fmt.Println("Partition size:", TempMBR.Partitions[i].Size)
		//fmt.Println("Partition start:", TempMBR.Partitions[i].Start)
		// fmt.Println("Partition status:", string(TempMBR.Partitions[i].Status[:]))
		partitionStatus := TempMBR.Partitions[i].Status[0]
		partitionStatusStr = strconv.Itoa(int(partitionStatus))
		// fmt.Println("Partition status:", partitionStatusStr)

		if TempMBR.Partitions[i].Size != 0 {
			if strings.Contains(string(TempMBR.Partitions[i].Id[:]), partitionId) {
				//fmt.Println("Partition found")
				if strings.Contains(partitionStatusStr, "1") {
					//fmt.Println("Partition is mounted")
					index = i
				} else {
					//fmt.Println("Partition is not mounted")
					return nil, fmt.Errorf("Partition is not mounted")
				}
				break
			}
		}
	}

	if index != -1 {
		//fmt.Println("Partition found")
	} else {
		//fmt.Println("Partition not found")
		return nil, fmt.Errorf("Partition not found")
	}

	file, err := Utilities.OpenFile(diskFileName)
	if err != nil {
		return nil, fmt.Errorf("Error abriendo el archivo")
	}

	var tempSuperblock Types.SuperBlock
	// Read object from bin file
	if err := Utilities.ReadObject(file, &tempSuperblock, int64(TempMBR.Partitions[index].Start)); err != nil {
		return nil, fmt.Errorf("Error leyendo el superbloque")
	}

	//fmt.Println("path:", path)
	// path = "/ruta/nueva"

	// split the path by /
	// TempStepsPath := strings.Split(pathFile, "/")
	// StepsPath := TempStepsPath[1:]

	fmt.Println("StepsPath:", pathUsersFile, "len(StepsPath):", len(pathUsersFile))
	fmt.Println("FilePathSeparator:", string(filepath.Separator))

	TempStepsPath := strings.Split(pathUsersFile, string(filepath.Separator))
	StepsPath := TempStepsPath[1:]

	// fmt.Println("StepsPath:", StepsPath, "len(StepsPath):", len(StepsPath))
	// for _, step := range StepsPath {
	// 	fmt.Println("step:", step)
	// }

	var Inode0 Types.Inode
	// Read object from bin file
	if err := Utilities.ReadObject(file, &Inode0, int64(tempSuperblock.S_inode_start)); err != nil {
		return nil, fmt.Errorf("unable to read")
	}

	indexInode := SarchInodeByPath(StepsPath, Inode0, file, tempSuperblock)

	if indexInode == -1 {
		//fmt.Println("User not found")
		return nil, fmt.Errorf("User not found")
	}

	var tempInode Types.Inode
	if err := Utilities.ReadObject(file, &tempInode, int64(tempSuperblock.S_inode_start+indexInode*int32(binary.Size(Types.Inode{})))); err != nil {
		return nil, fmt.Errorf("Error leyendo el inode")
	}
	//fmt.Printf("Inode: %+v\n", tempInode)

	if tempInode.I_block[0] == -1 {
		//fmt.Println("User not found")
		return nil, fmt.Errorf("User not found")
	}
	// else {
	// 	fmt.Println("User found")
	// }

	data := GetInodeFileDataOriginal(tempInode, file, tempSuperblock)

	lines := strings.Split(data, "\n")

	fileContents = lines

	return fileContents, nil
}
