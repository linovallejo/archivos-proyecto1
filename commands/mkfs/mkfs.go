package FileSystem

import (
	"encoding/binary"
	"fmt"
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
					//defer file.Close()
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
		//defer file.Close()
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

	if fs_ == "2fs" {
		file, err := Utilities.AbrirArchivo(diskFileName)
		if err != nil {
			return err
		}
		err = create_ext2(n, TempMBR.Partitions[index], newSuperblock, today, file)
		if err != nil {
			return err
		}
	} else {
		fmt.Println("EXT3")
	}

	// Close bin file
	//defer file.Close()

	fmt.Println("======End MKFS======")

	return nil
}

func create_ext2(inodesCount int32, partition Types.Partition, newSuperblock Types.SuperBlock, date string, file *os.File) error {
	fmt.Println("======Start CREATE EXT2======")
	fmt.Println("N:", inodesCount)
	fmt.Println("Superblock:", newSuperblock)
	fmt.Println("Date:", date)

	newSuperblock.S_filesystem_type = 2
	newSuperblock.S_bm_inode_start = partition.Start + int32(binary.Size(Types.SuperBlock{}))
	newSuperblock.S_bm_block_start = newSuperblock.S_bm_inode_start + inodesCount
	newSuperblock.S_inode_start = newSuperblock.S_bm_block_start + 3*inodesCount
	newSuperblock.S_block_start = newSuperblock.S_inode_start + inodesCount*int32(binary.Size(Types.Inode{}))

	newSuperblock.S_free_inodes_count -= 1
	newSuperblock.S_free_blocks_count -= 1

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

	var Inode0 Types.Inode //Inode 0
	Inode0.I_uid = 1
	Inode0.I_gid = 1
	Inode0.I_size = 0
	copy(Inode0.I_atime[:], date)
	copy(Inode0.I_ctime[:], date)
	copy(Inode0.I_mtime[:], date)
	copy(Inode0.I_perm[:], "0")
	copy(Inode0.I_perm[:], "664")

	for i := int32(0); i < 15; i++ {
		Inode0.I_block[i] = -1
	}

	Inode0.I_block[0] = 0

	// . | 0
	// .. | 0
	// users.txt | 1
	//

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

	// Inodo 0 -> Bloque 0 -> Inodo 1 -> Bloque 1
	// Crear la carpeta raiz /
	// Crear el archivo users.txt "1,G,root\n1,U,root,root,123\n"

	var err error
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
		builder.WriteString(fmt.Sprintf("\tInodo%d [label=\"{Inodo %d", i, i))
		for b, blockNum := range inode.I_block {
			if blockNum != -1 {
				builder.WriteString(fmt.Sprintf("|<f%d> AD%d", b, blockNum))
			}
		}
		builder.WriteString("}\"];\n")
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

func PrintMBR(data Types.MBR) {
	fmt.Println(fmt.Sprintf("CreationDate: %s, fit: %s, size: %d", string(data.MbrFechaCreacion[:]), string(data.DskFit[:]), data.MbrTamano))
	for i := 0; i < 4; i++ {
		PrintPartition(data.Partitions[i])
	}
}

func PrintPartition(data Types.Partition) {
	fmt.Println(fmt.Sprintf("Name: %s, type: %s, start: %d, size: %d, status: %s, id: %s", string(data.Name[:]), string(data.Type[:]), data.Start, data.Size, string(data.Status[:]), string(data.Id[:])))
}
