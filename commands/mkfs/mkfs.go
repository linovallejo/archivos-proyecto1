package Mkfs

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	Fdisk "proyecto1/commands/fdisk"
	Types "proyecto1/types"
	Utilities "proyecto1/utils"
	Utils "proyecto1/utils"
	"strings"
	"time"
)

func ExtractMkfsParams(params []string) (string, string, string, error) {
	var id string = ""
	var type_ string = ""
	var fs string = ""

	if len(params) == 0 {
		return "", "", "", fmt.Errorf("No se encontraron parámetros")
	}
	var parametrosObligatoriosOk bool = false
	idOk := false
	for _, param1 := range params {
		if strings.HasPrefix(param1, "-id=") {
			idOk = true
		}
	}

	parametrosObligatoriosOk = idOk

	if !parametrosObligatoriosOk {
		return "", "", "", fmt.Errorf("No se encontraron parámetros obligatorios")
	}

	for _, param := range params {
		switch {
		case strings.HasPrefix(param, "-id="):
			id = strings.TrimPrefix(param, "-id=")
			// Validar el id de la partición
			// TODO
		case strings.HasPrefix(param, "-type="):
			type_ = strings.TrimPrefix(param, "-type=")
		case strings.HasPrefix(param, "-fs="):
			fs = strings.TrimPrefix(param, "-fs=")
		}
	}

	// Unidad por defecto es Kilobytes
	if type_ == "" {
		type_ = "Full"
	}

	// Unidad por defecto es Kilobytes
	if fs == "" {
		fs = "2fs"
	}

	return id, type_, fs, nil
}

func MakeFileSystem(diskFileName string, id string, type_ string, fs_ string) error {
	fmt.Println("======Start MKFS======")
	fmt.Println("Id:", id)
	fmt.Println("Type:", type_)
	fmt.Println("Fs:", fs_)

	var err error
	var TempMBR *Types.MBR

	TempMBR, err = Fdisk.ReadMBR(diskFileName)
	if err != nil {
		//fmt.Println("Error leyendo el MBR:", err)
		return err
	}

	// Print object
	//PrintMBR(TempMBR)
	Utilities.PrintMBRv3(TempMBR)

	fmt.Println("-------------")

	var index int = -1
	// Iterate over the partitions
	for i := 0; i < 4; i++ {
		if TempMBR.Partitions[i].Size != 0 {
			fmt.Println("Partition id:", string(TempMBR.Partitions[i].Id[:]))
			fmt.Println("Partition name:", Utils.CleanPartitionName(TempMBR.Partitions[i].Name[:]))
			fmt.Println("Partition size:", TempMBR.Partitions[i].Size)
			//fmt.Println("Partition start:", TempMBR.Partitions[i].Start)
			fmt.Println("Partition status:", string(TempMBR.Partitions[i].Status[:]))
			if strings.Contains(string(TempMBR.Partitions[i].Id[:]), id) {
				fmt.Println("Partition found")
				if TempMBR.Partitions[i].Status[0] == 1 {
					fmt.Println("Partition is mounted")
					index = i
				} else {
					//fmt.Println("Partition is not mounted")
					return fmt.Errorf("Partition is not mounted")
				}
				break
			}
		}
	}

	if index != -1 {
		PrintPartition(TempMBR.Partitions[index])
	} else {
		//fmt.Println("Partition not found")
		return fmt.Errorf("Partition not found")
	}

	var today string = time.Now().Format("02/01/2006")

	numerador := int32(TempMBR.Partitions[index].Size - int32(binary.Size(Types.SuperBlock{})))
	denrominador_base := int32(4 + int32(binary.Size(Types.Inode{})) + 3*int32(binary.Size(Types.FileBlock{})))
	var temp int32 = 0
	if fs_ == "2fs" {
		temp = 0
	} else {
		temp = int32(binary.Size(Types.Journaling{}))
	}
	denrominador := denrominador_base + temp
	n := int32(numerador / denrominador)

	fmt.Println("N:", n)

	// var newMRB Types.MRB
	var newSuperblock Types.SuperBlock
	newSuperblock.S_inodes_count = 0
	newSuperblock.S_blocks_count = 0

	newSuperblock.S_free_blocks_count = 3 * n
	newSuperblock.S_free_inodes_count = n

	copy(newSuperblock.S_mtime[:], today)
	copy(newSuperblock.S_umtime[:], today)
	newSuperblock.S_mnt_count = 0

	//fmt.Println("create_ext2 not invoked")
	if fs_ == "2fs" {
		file, err := Utilities.AbrirArchivo(diskFileName)
		if err != nil {
			defer file.Close()
			return err
		}

		// file, err := os.Open(diskFileName)
		// if err != nil {
		// 	defer file.Close()
		// 	return err
		// }

		//var file *os.File
		err = create_ext2(n, TempMBR.Partitions[index], newSuperblock, today, file)
		if err != nil {
			defer file.Close()
			return err
		}
	} else {
		fmt.Println("EXT3")
	}

	//Close bin file
	//defer file.Close()

	fmt.Println("======End MKFS======")

	return nil
}

func setupSuperblock(superblock *Types.SuperBlock, partition Types.Partition, inodesCount int32) error {
	if inodesCount <= 0 {
		return fmt.Errorf("inodesCount must be positive")
	}

	superblock.S_filesystem_type = 2
	superblock.S_bm_inode_start = partition.Start + int32(binary.Size(Types.SuperBlock{}))
	superblock.S_bm_block_start = superblock.S_bm_inode_start + inodesCount

	if superblock.S_bm_block_start >= partition.Size {
		return fmt.Errorf("block bitmap start position is outside of the partition")
	}

	superblock.S_inode_start = superblock.S_bm_block_start + 3*inodesCount

	if superblock.S_inode_start >= partition.Size {
		return fmt.Errorf("inode start position is outside of the partition")
	}

	superblock.S_block_start = superblock.S_inode_start + inodesCount*int32(binary.Size(Types.Inode{}))

	if superblock.S_block_start >= partition.Size {
		return fmt.Errorf("block start position is outside of the partition")
	}

	if superblock.S_free_inodes_count <= 0 || superblock.S_free_blocks_count <= 0 {
		return fmt.Errorf("free inodes or blocks count is already zero or negative")
	}

	superblock.S_free_inodes_count -= 1
	superblock.S_free_blocks_count -= 1
	superblock.S_free_inodes_count -= 1
	superblock.S_free_blocks_count -= 1

	superblock.S_inodes_count = inodesCount

	return nil
}

func createRootInode(date string) (Types.Inode, error) {
	var inode Types.Inode
	inode.I_uid = 1
	inode.I_gid = 1
	inode.I_size = 0
	if len(date) > len(inode.I_atime) {
		return Types.Inode{}, fmt.Errorf("date string too long for inode timestamp fields")
	}
	copy(inode.I_atime[:], date)
	copy(inode.I_ctime[:], date)
	copy(inode.I_mtime[:], date)
	// Assuming "664" is the correct permission setting
	copy(inode.I_perm[:], "664")

	for i := range inode.I_block {
		inode.I_block[i] = -1
	}

	// Validate or correct this according to your filesystem's block numbering scheme
	inode.I_block[0] = 0

	// Add any additional validation or setup as needed
	return inode, nil
}

func create_ext2(inodesCount int32, partition Types.Partition, newSuperblock Types.SuperBlock, date string, file *os.File) error {
	fmt.Println("======Start CREATE EXT2======")
	fmt.Println("N:", inodesCount)
	fmt.Println("Superblock:", newSuperblock)
	fmt.Println("Date:", date)

	err := setupSuperblock(&newSuperblock, partition, inodesCount)
	if err != nil {
		fmt.Println("Error setting up superblock:", err)
		return err
	}

	fmt.Println("Superblock:", newSuperblock)
	fmt.Println("inodesCount:", inodesCount)

	for i := int32(0); i < inodesCount; i++ {
		err := Utilities.WriteObject(file, byte(0), int64(newSuperblock.S_bm_inode_start+i))
		if err != nil {
			fmt.Println("Error: ", err)
		}
	}

	for i := int32(0); i < 3*inodesCount; i++ {
		err := Utilities.WriteObject(file, byte(0), int64(newSuperblock.S_bm_block_start+i))
		if err != nil {
			fmt.Println("Error: ", err)
		}
	}

	var newInode Types.Inode
	for i := int32(0); i < 15; i++ {
		newInode.I_block[i] = -1
	}

	for i := int32(0); i < inodesCount; i++ {
		err := Utilities.WriteObject(file, newInode, int64(newSuperblock.S_inode_start+i*int32(binary.Size(Types.Inode{}))))
		if err != nil {
			fmt.Println("Error: ", err)
		}
	}

	var newFileblock Types.FileBlock
	for i := int32(0); i < 3*inodesCount; i++ {
		err := Utilities.WriteObject(file, newFileblock, int64(newSuperblock.S_block_start+i*int32(binary.Size(Types.FileBlock{}))))
		if err != nil {
			fmt.Println("Error: ", err)
		}
	}

	Inode0, err := createRootInode(date)
	if err != nil {
		fmt.Println("Error creating root inode:", err)
		return err
	}

	// // . | 0
	// // .. | 0
	// // users.txt | 1
	// //

	var Folderblock0 Types.DirectoryBlock //Bloque 0 -> carpetas
	Folderblock0.B_content[0].B_inodo = 0
	copy(Folderblock0.B_content[0].B_name[:], ".")
	Folderblock0.B_content[1].B_inodo = 0
	copy(Folderblock0.B_content[1].B_name[:], "..")
	Folderblock0.B_content[1].B_inodo = 1
	copy(Folderblock0.B_content[1].B_name[:], "users.txt")

	var Inode1 Types.Inode //Inode 1
	Inode1.I_uid = 1
	Inode1.I_gid = 1
	Inode1.I_size = int32(binary.Size(Types.DirectoryBlock{}))
	copy(Inode1.I_atime[:], date)
	copy(Inode1.I_ctime[:], date)
	copy(Inode1.I_mtime[:], date)
	copy(Inode1.I_perm[:], "0")
	copy(Inode1.I_perm[:], "664")

	for i := int32(0); i < 15; i++ {
		Inode1.I_block[i] = -1
	}

	Inode1.I_block[0] = 1

	data := "1,G,root\n1,U,root,root,123\n"
	var Fileblock1 Types.FileBlock //Bloque 1 -> archivo
	copy(Fileblock1.B_content[:], data)

	// // Inodo 0 -> Bloque 0 -> Inodo 1 -> Bloque 1
	// // Crear la carpeta raiz /
	// // Crear el archivo users.txt "1,G,root\n1,U,root,root,123\n"

	// write superblock
	err = Utilities.WriteObject(file, newSuperblock, int64(partition.Start))
	if err != nil {
		return err
	}

	// write bitmap inodes
	err = Utilities.WriteObject(file, byte(1), int64(newSuperblock.S_bm_inode_start))
	if err != nil {
		return err
	}

	err = Utilities.WriteObject(file, byte(1), int64(newSuperblock.S_bm_inode_start+1))
	if err != nil {
		return err
	}

	// write bitmap blocks
	err = Utilities.WriteObject(file, byte(1), int64(newSuperblock.S_bm_block_start))
	if err != nil {
		return err
	}

	err = Utilities.WriteObject(file, byte(1), int64(newSuperblock.S_bm_block_start+1))
	if err != nil {
		return err
	}

	fmt.Println("Inode 0:", int64(newSuperblock.S_inode_start))
	fmt.Println("Inode 1:", int64(newSuperblock.S_inode_start+int32(binary.Size(Types.Inode{}))))

	// write inodes
	err = Utilities.WriteObject(file, Inode0, int64(newSuperblock.S_inode_start)) //Inode 0
	if err != nil {
		return err
	}
	err = Utilities.WriteObject(file, Inode1, int64(newSuperblock.S_inode_start+int32(binary.Size(Types.Inode{})))) //Inode 1
	if err != nil {
		return err
	}
	// write blocks
	err = Utilities.WriteObject(file, Folderblock0, int64(newSuperblock.S_block_start)) //Bloque 0
	if err != nil {
		return err
	}
	err = Utilities.WriteObject(file, Fileblock1, int64(newSuperblock.S_block_start+int32(binary.Size(Types.FileBlock{})))) //Bloque 1
	if err != nil {
		return err
	}

	fmt.Println("======End CREATE EXT2======")
	fmt.Println("Superblock:", newSuperblock)
	fmt.Println("inodesCount:", inodesCount)
	fmt.Println("======End CREATE EXT2======")

	return nil
}

func GenerateDotCodeTree(inodes []Types.Inode, blocks []Types.DirectoryBlock) string {
	var builder strings.Builder

	// Begin the DOT graph
	builder.WriteString("digraph filesystem {\n")
	builder.WriteString("\trankdir=LR;\n")
	builder.WriteString("\tnode [shape=record];\n")

	// Iterate over inodes
	for i, inode := range inodes {
		hasValidBlock := false // Flag to track if the inode has a valid block
		for _, blockNum := range inode.I_block {
			if blockNum != -1 {
				hasValidBlock = true
				break // No need to check further blocks
			}
		}

		if hasValidBlock { // Only output inodes with valid blocks
			builder.WriteString(fmt.Sprintf("\tInodo%d [label=\"{Inodo %d", i, i))
			for b, blockNum := range inode.I_block {
				if blockNum != -1 {
					builder.WriteString(fmt.Sprintf("|<f%d> AD%d", b, blockNum))
				}
			}
			builder.WriteString("}\"];\n")
		}
	}

	// Iterate over blocks
	for i, block := range blocks {
		builder.WriteString(fmt.Sprintf("\tBloque%d [label=\"", i))
		for j, entry := range block.B_content {
			entryName := strings.Trim(string(entry.B_name[:]), "\x00")
			if entryName != "" {
				builder.WriteString(fmt.Sprintf("<f%d> %s", j, entryName))
				if j < len(block.B_content)-1 {
					builder.WriteString("|")
				}
			}
		}
		builder.WriteString("\"];\n")
	}

	// Link inodes to blocks
	for i, inode := range inodes {
		for b, blockNum := range inode.I_block {
			if blockNum != -1 {
				builder.WriteString(fmt.Sprintf("\tInodo%d:f%d -> Bloque%d;\n", i, b, blockNum))
			}
		}
	}

	// Close the DOT graph
	builder.WriteString("}\n")

	return builder.String()
}

// Read the superblock from the file to find out where inodes and blocks start.
func ReadSuperBlock(filePath string, partitionStart int32) (Types.SuperBlock, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return Types.SuperBlock{}, err
	}
	defer file.Close()

	// Seek to the start of the partition to read the superblock
	_, err = file.Seek(int64(partitionStart), io.SeekStart)
	if err != nil {
		return Types.SuperBlock{}, err
	}

	var superblock Types.SuperBlock
	err = binary.Read(file, binary.LittleEndian, &superblock)
	if err != nil {
		return Types.SuperBlock{}, err
	}

	return superblock, nil
}

func ReadInodesFromFile(filePath string, superblock Types.SuperBlock) ([]Types.Inode, error) {
	file, err := os.Open(filePath)
	if err != nil {
		defer file.Close()
		return nil, err
	}
	defer file.Close()

	fmt.Println("ReadInodesFromFile.Superblock:", superblock)

	// Seek to the inode start
	_, err = file.Seek(int64(superblock.S_inode_start), io.SeekStart)
	if err != nil {
		return nil, err
	} else {
		fmt.Println("ReadInodesFromFile.Inode start:", superblock.S_inode_start)
	}

	numInodes := superblock.S_inodes_count
	fmt.Println("ReadInodesFromFile.numInodes:", numInodes)
	inodes := make([]Types.Inode, numInodes)
	for i := int32(0); i < numInodes; i++ {
		inode := Types.Inode{}
		err = binary.Read(file, binary.LittleEndian, &inode)
		if err != nil {
			return nil, err
		}
		inodes[i] = inode
	}

	return inodes, nil
}

func ReadDirectoryBlocksFromFile(filePath string, superblock Types.SuperBlock) ([]Types.DirectoryBlock, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Seek to the block start
	_, err = file.Seek(int64(superblock.S_block_start), io.SeekStart)
	if err != nil {
		return nil, err
	}

	// Assuming only one block for the root directory
	blocks := make([]Types.DirectoryBlock, 1)
	for i := range blocks {
		block := Types.DirectoryBlock{}
		err = binary.Read(file, binary.LittleEndian, &block)
		if err != nil {
			return nil, err
		}
		blocks[i] = block
	}

	return blocks, nil
}

func PrintMBR(data Types.MBR) {
	fmt.Println(fmt.Sprintf("CreationDate: %s, fit: %s, size: %d", string(data.MbrFechaCreacion[:]), string(data.DskFit[:]), data.MbrTamano))
	for i := 0; i < 4; i++ {
		PrintPartition(data.Partitions[i])
	}
}

func PrintPartition(data Types.Partition) {
	fmt.Println(fmt.Sprintf("Name: %s, type: %s, start: %d, size: %d, status: %s, id: %s", string(data.Name[:]), string(data.Type[:]), data.Start, data.Size, string(data.Status[:]), string(data.Id[:])))
}
