package UserWorkspace

import (
	"encoding/binary"
	"fmt"
	"os"
	Types "proyecto1/types"
	Utilities "proyecto1/utils"
	"strings"
)

func InodeSearch(path string, file *os.File, superblock Types.SuperBlock) int32 {
	fmt.Println("======Start INITSEARCH======")
	fmt.Println("path:", path)

	stepsPath := strings.Split(path, "/")[1:] // Directly split and exclude the empty first element
	fmt.Println("StepsPath:", stepsPath, "len(StepsPath):", len(stepsPath))

	for _, step := range stepsPath {
		fmt.Println("step:", step)
	}

	var rootInode Types.Inode
	if err := Utilities.ReadObject(file, &rootInode, int64(superblock.S_inode_start)); err != nil {
		fmt.Println("Error reading root inode:", err)
		return -1
	}

	fmt.Println("======End INITSEARCH======")
	return SearchInodeByPath(stepsPath, rootInode, file, superblock)
}

func SearchInodeByPath(stepsPath []string, currentInode Types.Inode, file *os.File, superblock Types.SuperBlock) int32 {
	fmt.Println("======Start SEARCHINODEBYPATH======")

	if len(stepsPath) == 0 {
		fmt.Println("No path specified")
		return -1
	}

	searchedName := strings.TrimSpace(stepsPath[0]) // Use the first path step and trim spaces
	fmt.Println("SearchedName:", searchedName)

	for _, block := range currentInode.I_block[:12] { // Direct blocks only
		if block != -1 { // Valid block
			var folderBlock Types.DirectoryBlock
			if err := Utilities.ReadObject(file, &folderBlock, int64(superblock.S_block_start+block*int32(binary.Size(Types.DirectoryBlock{})))); err != nil {
				fmt.Println("Error reading folder block:", err)
				return -1
			}

			for _, folder := range folderBlock.B_content {
				fmt.Println("Folder === Name:", strings.TrimSpace(string(folder.B_name[:])), "B_inodo", folder.B_inodo)
				if strings.TrimSpace(string(folder.B_name[:])) == searchedName {
					if len(stepsPath) == 1 { // Final path element found
						fmt.Println("Folder found======")
						return folder.B_inodo
					}

					// Continue search with next path step
					var nextInode Types.Inode
					if err := Utilities.ReadObject(file, &nextInode, int64(superblock.S_inode_start+int32(folder.B_inodo)*int32(binary.Size(Types.Inode{})))); err != nil {
						fmt.Println("Error reading next inode:", err)
						return -1
					}
					return SearchInodeByPath(stepsPath[1:], nextInode, file, superblock)
				}
			}
		}
	}

	fmt.Println("======End SEARCHINODEBYPATH======")
	return 0 // Inode not found
}

func GetInodeFileData(inode Types.Inode, file *os.File, superblock Types.SuperBlock) string {
	fmt.Println("======Start GETINODEFILEDATA======")
	var contentBuilder strings.Builder

	// Handle direct blocks
	for i, block := range inode.I_block[:12] { // Assuming the first 12 are direct blocks
		if block == -1 {
			continue // Skip invalid blocks
		}

		var fileBlock Types.FileBlock
		blockPosition := superblock.S_block_start + block*int32(binary.Size(Types.FileBlock{}))
		if err := Utilities.ReadObject(file, &fileBlock, int64(blockPosition)); err != nil {
			fmt.Printf("Error reading block %d: %s\n", i, err)
			break // Stop processing further blocks on error
		}

		contentBuilder.Write(fileBlock.B_content[:])
	}

	// Handle indirect blocks if necessary
	if inode.I_block[12] != -1 {
		// Example for handling a single level of indirect blocks
		handleIndirectBlocks(inode.I_block[12], file, superblock, &contentBuilder)
	}

	fmt.Println("======End GETINODEFILEDATA======")
	return contentBuilder.String()
}

// handleIndirectBlocks demonstrates how to handle indirect blocks. This is a placeholder and should
// be implemented based on the actual filesystem structure for indirect blocks.
func handleIndirectBlocks(indirectBlock int32, file *os.File, superblock Types.SuperBlock, contentBuilder *strings.Builder) {
	// This function needs to be implemented based on how indirect addressing is structured in your filesystem.
	// Generally, it would involve reading the block of pointers, then iterating over each to read the file blocks they point to.
}
