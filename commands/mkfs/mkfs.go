package Mkfs

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	Fdisk "proyecto1/commands/fdisk"
	Types "proyecto1/types"
	Utilities "proyecto1/utils"
	Utils "proyecto1/utils"
	"strconv"
	"strings"
	"time"
	"unsafe"
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
				//fmt.Println("Partition found")
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
		return fmt.Errorf("partición no existe")
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
	superblock.S_blocks_count = 3 * inodesCount

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

	// Inicializa bitmap de inodos
	for i := int32(0); i < inodesCount; i++ {
		err := Utilities.WriteObject(file, byte(0), int64(newSuperblock.S_bm_inode_start+i))
		if err != nil {
			fmt.Println("Error: ", err)
		}
	}

	// Inicializa bitmap de bloques
	for i := int32(0); i < 3*inodesCount; i++ {
		err := Utilities.WriteObject(file, byte(0), int64(newSuperblock.S_bm_block_start+i))
		if err != nil {
			fmt.Println("Error: ", err)
		}
	}

	// Inicializa nuevo inodo y todos sus bloques
	var newInode Types.Inode
	for i := int32(0); i < 15; i++ {
		newInode.I_block[i] = -1
	}

	// Inicializa todos los inodos
	for i := int32(0); i < inodesCount; i++ {
		err := Utilities.WriteObject(file, newInode, int64(newSuperblock.S_inode_start+i*int32(binary.Size(Types.Inode{}))))
		if err != nil {
			fmt.Println("Error: ", err)
		}
	}

	// Inicializa todos los bloques
	var newFileblock Types.FileBlock
	for i := int32(0); i < 3*inodesCount; i++ {
		err := Utilities.WriteObject(file, newFileblock, int64(newSuperblock.S_block_start+i*int32(binary.Size(Types.FileBlock{}))))
		if err != nil {
			fmt.Println("Error: ", err)
		}
	}

	// . | 0
	// Nodo 0
	Inode0, err := createRootInode(date)
	if err != nil {
		fmt.Println("Error creating root inode:", err)
		return err
	}

	// // . | 0
	// // .. | 0
	// // users.txt | 1
	// //

	// Inodo 0
	var Folderblock0 Types.DirectoryBlock //Bloque 0 -> carpetas
	Folderblock0.B_content[0].B_inodo = 0
	copy(Folderblock0.B_content[0].B_name[:], ".")
	Folderblock0.B_content[1].B_inodo = 0
	copy(Folderblock0.B_content[1].B_name[:], "..")
	// Folderblock.B_content[2] apuntando a inodo 1 que es donde va a estar su contenido: 1,G,root\n1,U,root,root,555
	Folderblock0.B_content[2].B_inodo = 1
	copy(Folderblock0.B_content[1].B_name[:], "users.txt")

	//Inode 1
	var Inode1 Types.Inode
	Inode1.I_uid = 1
	Inode1.I_gid = 1
	Inode1.I_size = int32(binary.Size(Types.DirectoryBlock{}))
	copy(Inode1.I_atime[:], date)
	copy(Inode1.I_ctime[:], date)
	copy(Inode1.I_mtime[:], date)
	copy(Inode1.I_perm[:], "0")
	copy(Inode1.I_perm[:], "664")

	// Inodo 1 inicializacion de sus bloques
	for i := int32(0); i < 15; i++ {
		Inode1.I_block[i] = -1
	}

	// Marcando bloque 0 del Inodo 1 como utilizado
	Inode1.I_block[0] = 1

	// Contenido de users.txt
	data := "1,G,root\n1,U,root,root,555\n"

	// Seteando Bloque de archivo 1 con el contenido de users.txt
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

	// marcar en el bitmap de inodos, los nodos 0 y 1 como ocupados
	err = Utilities.WriteObject(file, byte(1), int64(newSuperblock.S_bm_inode_start))
	if err != nil {
		return err
	}
	err = Utilities.WriteObject(file, byte(1), int64(newSuperblock.S_bm_inode_start+1))
	if err != nil {
		return err
	}

	// marcar en el bitmap de bloques, los bloques 0 y 1 como ocupados
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
		//fmt.Println("Inode:", i, "Inode:", inode)
		hasValidBlock := false // Flag to track if the inode has a valid block
		for _, blockNum := range inode.I_block {
			if blockNum != -1 {
				hasValidBlock = true
				break // No need to check further blocks
			}
		}

		if hasValidBlock { // Only output inodes with valid blocks
			fmt.Printf("Inode with valid block %d: %+v\n", i, inode)
			builder.WriteString(fmt.Sprintf("\tInodo%d [label=\"{Inodo %d", i, i))
			for b, blockNum := range inode.I_block {
				if blockNum != -1 {
					builder.WriteString(fmt.Sprintf("|<f%d> AD%d", b, blockNum))
				}
			}
			builder.WriteString("}\"];\n")
		}
		if i > 5 {
			break
		}
	}

	// Iterate over blocks
	for i, block := range blocks {
		fmt.Printf("Block %d: %+v\n", i, block)
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
	Utils.LineaDoble(80)
	fmt.Println("Link inodes to blocks")
	for i, inode := range inodes {
		//fmt.Println("Inode:", i, "Inode:", inode)
		for b, blockNum := range inode.I_block {
			//fmt.Println("Block:", b, "Block:", blockNum)
			if blockNum != -1 {
				builder.WriteString(fmt.Sprintf("\tInodo%d:f%d -> Bloque%d;\n", i, b, blockNum))
			}
		}
	}
	fmt.Println("======End DOT code======")

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

func ReadAllUsedInodesFromFile(filePath string, superblock Types.SuperBlock) ([]Types.Inode, error) {
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
	usedInodesIndex := 0                            // Keep track of where to insert in usedInodes
	usedInodes := make([]Types.Inode, 0, numInodes) // Allocate initially with capacity

	for i := int32(0); i < numInodes; i++ {
		inode := Types.Inode{}
		err = binary.Read(file, binary.LittleEndian, &inode)
		if err != nil {
			return nil, err
		}
		if inode.I_block[0] != -1 {
			usedInodes = append(usedInodes, inode) // Efficiently add only used inodes
			usedInodesIndex++
		}
	}

	return usedInodes, nil
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

func ReadBlockBitmap(file *os.File, superblock Types.SuperBlock) ([]bool, error) {
	// The total number of bytes needed to represent the block bitmap
	bitmapSize := (superblock.S_blocks_count + 7) / 8

	// Seek to the start of the block bitmap
	_, err := file.Seek(int64(superblock.S_bm_block_start), io.SeekStart)
	if err != nil {
		return nil, err
	}

	bitmapBytes := make([]byte, bitmapSize)
	_, err = file.Read(bitmapBytes)
	if err != nil {
		return nil, err
	}

	blockBitmap := make([]bool, superblock.S_blocks_count)
	for i := int32(0); i < superblock.S_blocks_count; i++ {
		// Calculate byte index and bit index within that byte
		byteIndex := i / 8
		bitIndex := i % 8

		// Check if the bit at bitIndex in byteIndex is set
		blockBitmap[i] = (bitmapBytes[byteIndex] & (1 << bitIndex)) != 0
	}

	return blockBitmap, nil
}

func ReadAllUsedBlocksFromFile(filePath string, superblock Types.SuperBlock) ([]Types.DirectoryBlock, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	blockSize := int64(superblock.S_block_size)
	blockBitmap, err := ReadBlockBitmap(file, superblock)
	if err != nil {
		return nil, err
	}

	var usedBlocks []Types.DirectoryBlock
	for i, used := range blockBitmap {
		if used {
			// Calculate the offset for the block
			offset := int64(superblock.S_block_start) + int64(i)*blockSize

			// Seek to the block's start
			_, err = file.Seek(offset, io.SeekStart)
			if err != nil {
				return nil, err
			}

			// Read the block
			block := Types.DirectoryBlock{}
			err = binary.Read(file, binary.LittleEndian, &block)
			if err != nil {
				return nil, err
			}

			// Add the block to the slice of used blocks
			usedBlocks = append(usedBlocks, block)
		}
	}

	return usedBlocks, nil
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

func GraficarArbol(path string, part_start_Partition int, usedInodes []Types.Inode, usedBlocks []Types.DirectoryBlock) (string, error) {
	// Se limpia el strings que almacena el codigo dot del reporte
	var RepDot string

	RepDot = ""

	// Apertura del archivo del disco binario
	disco_actual, err := os.OpenFile(path, os.O_RDWR, 0660)
	if err != nil {
		//msg_error(err)
		return "", err
	}
	defer disco_actual.Close()

	// Estructuras necesarias a utilizar
	//superB := Types.SuperBlock{}
	//inodo := Types.Inode{}
	// carpeta := Types.DirectoryBlock{}
	// archivo := Types.FileBlock{}
	// apuntador := Types.PointerBlock{}

	// Tamaño de algunas estructuras
	var inodoTable Types.Inode
	const i_size = unsafe.Sizeof(inodoTable)

	var blockCarpeta Types.DirectoryBlock
	const bc_size = unsafe.Sizeof(blockCarpeta)

	var blockArchivo Types.FileBlock
	const ba_size = unsafe.Sizeof(blockArchivo)

	var blockApuntador Types.PointerBlock
	const bapu_size = unsafe.Sizeof(blockApuntador)

	RepDot += "digraph G{\n\n"
	RepDot += "    rankdir=\"LR\" \n"

	// Creamos lo inodos
	start := time.Now()

	// Pre-processing
	//inodeMap := make(map[int]Types.Inode)
	//blockMap := make(map[int]string) // Values: "folder", "file", "pointer"

	// for inodeNum, inodo := range usedInodes { // Iterate with index

	// 	RepDot += fmt.Sprintf("  inodo_%d [ shape=plaintext fontname=\"Century Gothic\" label=<\n", inodeNum)
	// 	RepDot += "  <table bgcolor=\"royalblue\" border=\"0\" >"
	// 	RepDot += "  <tr> <td colspan=\"2\"><b>Inode " + strconv.Itoa(inodeNum) + "</b></td></tr>\n"

	// 	// ... (Other inode attributes) ...

	// 	RepDot += "  <tr> <td bgcolor=\"lightsteelblue\"> i_blocks </td> <td bgcolor=\"white\"> "
	// 	for _, blockNum := range inodo.I_block {
	// 		if blockNum != 255 {
	// 			RepDot += strconv.Itoa(int(blockNum)) + " "
	// 		}
	// 	}
	// 	RepDot += "</td> </tr>\n"

	// 	RepDot += "  </table>>]\n\n"
	// }
	// Optimized Main Loop
	// for inodeNum, inodo := range usedInodes {
	// 	RepDot += fmt.Sprintf("  inodo_%d [ shape=record fontname=\"Century Gothic\" label=\"{ Inode %d | { i_uid: %s | i_gid: %s | ... } | { i_blocks: %v } }\" ]\n\n", inodeNum, inodeNum, string(inodo.I_uid), string(inodo.I_gid), inodo.I_block)
	// }

	// Optimized Main Loop
	for inodeNum, inodo := range usedInodes {
		RepDot += fmt.Sprintf("  inodo_%d [ shape=plaintext fontname=\"Century Gothic\" label=<\n", inodeNum)
		RepDot += "  <table border=\"0\" cellborder=\"1\" cellspacing=\"0\" >"
		RepDot += fmt.Sprintf("  <tr> <td colspan=\"2\"><b>Inode %d</b></td></tr>\n", inodeNum)

		for blockNum := 0; blockNum < 15; blockNum++ {
			bgColor := "white"
			if blockNum >= 12 {
				bgColor = "lightsteelblue" // Or your preferred color
			}
			RepDot += fmt.Sprintf("  <tr> <td bgcolor=\"%s\">pt%d</td> <td bgcolor=\"%s\"> %d </td> </tr>\n", bgColor, blockNum+1, bgColor, inodo.I_block[blockNum])
		}

		RepDot += "  </table>>]\n\n"
	}

	// Similar process for blockMap

	elapsed := time.Since(start)
	fmt.Println("myFunction() took:", elapsed)

	RepDot += "\n\n}"
	disco_actual.Close()

	//Consola += "Reporte Tree generado con exito!\n"
	fmt.Println("Reporte Tree generado con exito!")

	return RepDot, nil
}

func GraficarTREE(path string, part_start_Partition int) (string, error) {
	// Se limpia el strings que almacena el codigo dot del reporte
	var RepDot string

	RepDot = ""

	// Apertura del archivo del disco binario
	disco_actual, err := os.OpenFile(path, os.O_RDWR, 0660)
	if err != nil {
		//msg_error(err)
		return "", err
	}
	defer disco_actual.Close()

	// Estructuras necesarias a utilizar
	superB := Types.SuperBlock{}
	inodo := Types.Inode{}
	carpeta := Types.DirectoryBlock{}
	archivo := Types.FileBlock{}
	apuntador := Types.PointerBlock{}

	// Tamaño de algunas estructuras
	var inodoTable Types.Inode
	const i_size = unsafe.Sizeof(inodoTable)

	var blockCarpeta Types.DirectoryBlock
	const bc_size = unsafe.Sizeof(blockCarpeta)

	var blockArchivo Types.FileBlock
	const ba_size = unsafe.Sizeof(blockArchivo)

	var blockApuntador Types.PointerBlock
	const bapu_size = unsafe.Sizeof(blockApuntador)

	// --------Se extrae el SB del disco---------
	var sbsize int = int(binary.Size(superB))
	disco_actual.Seek(int64(part_start_Partition), 0)
	data := leerEnFILE(disco_actual, sbsize)
	//data := Utils.ReadObject(disco_actual, superB, int64(part_start_Partition))
	buffer := bytes.NewBuffer(data)
	err = binary.Read(buffer, binary.BigEndian, &superB)
	if err != nil {
		//Consola += "Binary.Read failed\n"
		//msg_error(err)
		fmt.Println("Binary.Read failed")
		return "", err
	}

	aux := superB.S_bm_inode_start
	i := 0

	RepDot += "digraph G{\n\n"
	RepDot += "    rankdir=\"LR\" \n"

	// Creamos lo inodos
	start := time.Now()
	for aux < superB.S_bm_block_start {

		disco_actual.Seek(int64(superB.S_bm_inode_start)+int64(i), 0)
		aux++
		port := 0
		dataBMI := getc(disco_actual)  // me devuelve el dato en byte
		bufINT := int(dataBMI)         // Lo convierto en int
		buffer := strconv.Itoa(bufINT) // Convierto el int a string

		if buffer == "1" {
			var inodosize int = int(binary.Size(inodo))
			disco_actual.Seek(int64(superB.S_inode_start)+int64(i_size)*int64(i), 0)
			data := leerEnFILE(disco_actual, inodosize)
			buffer := bytes.NewBuffer(data)
			err = binary.Read(buffer, binary.BigEndian, &inodo)
			if err != nil {
				//Consola += "Binary.Read failed\n"
				//msg_error(err)
				fmt.Println("Binary.Read failed")
			}

			RepDot += "    inodo_" + strconv.Itoa(i) + " [ shape=plaintext fontname=\"Century Gothic\" label=<\n"
			RepDot += "   <table bgcolor=\"royalblue\" border=\"0\" >"
			RepDot += "    <tr> <td colspan=\"2\"><b>Inode " + strconv.Itoa(i) + "</b></td></tr>\n"
			RepDot += "    <tr> <td bgcolor=\"lightsteelblue\"> i_uid </td> <td bgcolor=\"white\"> " + string(inodo.I_uid) + " </td>  </tr>\n"
			RepDot += "    <tr> <td bgcolor=\"lightsteelblue\"> i_gid </td> <td bgcolor=\"white\"> " + string(inodo.I_gid) + " </td>  </tr>\n"
			RepDot += "    <tr> <td bgcolor=\"lightsteelblue\"> i_size </td><td bgcolor=\"white\"> " + string(inodo.I_size) + " </td> </tr>\n"
			RepDot += "    <tr> <td bgcolor=\"lightsteelblue\"> i_atime </td> <td bgcolor=\"white\"> " + byteToStr(inodo.I_atime[:]) + " </td> </tr>\n"
			RepDot += "    <tr> <td bgcolor=\"lightsteelblue\"> i_ctime </td> <td bgcolor=\"white\"> " + byteToStr(inodo.I_ctime[:]) + " </td> </tr>\n"
			RepDot += "    <tr> <td bgcolor=\"lightsteelblue\"> i_mtime </td> <td bgcolor=\"white\"> " + byteToStr(inodo.I_mtime[:]) + " </td> </tr>\n"

			for b := 0; b < 15; b++ {
				RepDot += "    <tr> <td bgcolor=\"lightsteelblue\"> i_block_" + strconv.Itoa(port) + " </td> <td bgcolor=\"white\" port=\"f" + strconv.Itoa(b) + "\"> " + strconv.Itoa(int(inodo.I_block[b])) + " </td></tr>\n"
				port++
			}

			RepDot += "    <tr> <td bgcolor=\"lightsteelblue\"> i_type </td> <td bgcolor=\"white\"> " + byteToStr(inodo.I_type[:]) + " </td>  </tr>\n"
			RepDot += "    <tr> <td bgcolor=\"lightsteelblue\"> i_perm </td> <td bgcolor=\"white\"> " + byteToStr(inodo.I_perm[:]) + " </td>  </tr>\n"
			RepDot += "   </table>>]\n\n"

			// Creamos los bloques relacionados al inodo
			for j := 0; j < 15; j++ {
				port = 0

				// El 255 representa al -1
				if int(inodo.I_block[j]) != 255 {

					disco_actual.Seek(int64(superB.S_bm_block_start)+int64(inodo.I_block[j]), 0)

					buffINT := int(getc(disco_actual))
					buffer := strconv.Itoa(buffINT)

					if buffer == "1" { // Bloque carpeta
						disco_actual.Seek(int64(superB.S_block_start)+int64(bc_size)*int64(inodo.I_block[j]), 0)
						// --------Se extrae el Bloque Carpeta del disco---------
						var bcsize int = int(binary.Size(carpeta))
						data := leerEnFILE(disco_actual, bcsize)
						buff := bytes.NewBuffer(data)
						err = binary.Read(buff, binary.BigEndian, &carpeta)
						if err != nil {
							// Consola += "Binary.Read failed\n"
							// msg_error(err)
							fmt.Println("Binary.Read failed")
						}

						RepDot += "    bloque_" + strconv.Itoa(int(inodo.I_block[j])) + " [shape=plaintext fontname=\"Century Gothic\" label=< \n"
						RepDot += "   <table bgcolor=\"seagreen\" border=\"0\">\n"
						RepDot += "    <tr> <td colspan=\"2\"><b>Folder block " + strconv.Itoa(int(inodo.I_block[j])) + "</b></td></tr>\n"
						RepDot += "    <tr> <td bgcolor=\"mediumseagreen\"> b_name </td> <td bgcolor=\"mediumseagreen\"> b_inode </td></tr>\n"

						for c := 0; c < 4; c++ {
							RepDot += "    <tr> <td bgcolor=\"white\" > " + byteToStr(carpeta.B_content[c].B_name[:]) + " </td> <td bgcolor=\"white\"  port=\"f" + strconv.Itoa(port) + "\"> " + string(carpeta.B_content[c].B_inodo) + " </td></tr>\n"
							port++
						}

						RepDot += "   </table>>]\n\n"

						// Relacion de bloques a inodos
						for c := 0; c < 4; c++ {
							if carpeta.B_content[c].B_inodo != -1 {

								if strings.Compare(byteToStr(carpeta.B_content[c].B_name[:]), ".") != 0 && strings.Compare(byteToStr(carpeta.B_content[c].B_name[:]), "..") != 0 {
									RepDot += "    bloque_" + strconv.Itoa(int(inodo.I_block[j])) + ":f" + strconv.Itoa(c) + " -> inodo_" + string(carpeta.B_content[c].B_inodo) + ";\n"
								}
							}
						}

					} else if buffer == "2" { // Bloque archivo
						disco_actual.Seek(int64(superB.S_block_start)+int64(ba_size)*int64(inodo.I_block[j]), 0)
						// --------Se extrae el Bloque Archivo del disco---------
						var basize int = int(binary.Size(archivo))
						data := leerEnFILE(disco_actual, basize)
						buff := bytes.NewBuffer(data)
						err = binary.Read(buff, binary.BigEndian, &archivo)
						if err != nil {
							// Consola += "Binary.Read failed\n"
							// msg_error(err)
							fmt.Println("Binary.Read failed")
						}

						RepDot += "    bloque_" + strconv.Itoa(int(inodo.I_block[j])) + " [shape=plaintext fontname=\"Century Gothic\" label=< \n"
						RepDot += "   <table border=\"0\" bgcolor=\"sandybrown\">\n"
						RepDot += "    <tr> <td> <b>File block " + strconv.Itoa(int(inodo.I_block[j])) + "</b></td></tr>\n"
						RepDot += "    <tr> <td bgcolor=\"white\"> " + byteToStr(archivo.B_content[:]) + " </td></tr>\n"
						RepDot += "   </table>>]\n\n"

					} else if buffer == "3" { // Bloque apuntador
						disco_actual.Seek(int64(superB.S_block_start)+int64(bapu_size)*int64(inodo.I_block[j]), 0)
						// --------Se extrae el Bloque Apuntador del disco---------
						var bapusize int = int(binary.Size(apuntador))
						data := leerEnFILE(disco_actual, bapusize)
						buff := bytes.NewBuffer(data)
						err = binary.Read(buff, binary.BigEndian, &apuntador)
						if err != nil {
							// Consola += "Binary.Read failed\n"
							// msg_error(err)
						}

						RepDot += "    bloque_" + strconv.Itoa(int(inodo.I_block[j])) + " [shape=plaintext fontname=\"Century Gothic\" label=< \n"
						RepDot += "   <table border=\"0\" bgcolor=\"khaki\">\n"
						RepDot += "    <tr> <td> <b>Pointer block " + strconv.Itoa(int(inodo.I_block[j])) + "</b></td></tr>\n"

						for a := 0; a < 16; a++ {
							RepDot += "    <tr> <td bgcolor=\"white\" port=\"f" + strconv.Itoa(port) + "\">" + string(apuntador.B_pointers[a]) + "</td> </tr>\n"
							port++
						}

						RepDot += "   </table>>]\n\n"

						// Bloques carpeta/archivo  del bloque de apuntadores
						for x := 0; x < 16; x++ {
							port = 0

							// El 255 representa al -1
							if int(apuntador.B_pointers[x]) != 255 {
								disco_actual.Seek(int64(superB.S_bm_block_start)+int64(apuntador.B_pointers[x]), 0)

								buffINT := int(getc(disco_actual))
								buffer := strconv.Itoa(buffINT)

								if buffer == "1" {
									disco_actual.Seek(int64(superB.S_block_start)+int64(bc_size)*int64(apuntador.B_pointers[x]), 0)
									// --------Se extrae el Bloque Carpeta del disco---------
									var bcsize int = int(binary.Size(carpeta))
									data := leerEnFILE(disco_actual, bcsize)
									buff := bytes.NewBuffer(data)
									err = binary.Read(buff, binary.BigEndian, &carpeta)
									if err != nil {
										// Consola += "Binary.Read failed\n"
										// msg_error(err)
									}

									RepDot += "    bloque_" + string(apuntador.B_pointers[x]) + " [shape=plaintext fontname=\"Century Gothic\" label=< \n"
									RepDot += "   <table bgcolor=\"seagreen\" border=\"0\">\n"
									RepDot += "    <tr> <td colspan=\"2\"><b>Folder block " + string(apuntador.B_pointers[x]) + "</b></td></tr>\n"
									RepDot += "    <tr> <td bgcolor=\"mediumseagreen\"> b_name </td> <td bgcolor=\"mediumseagreen\"> b_inode </td></tr>\n"

									for c := 0; c < 4; c++ {
										RepDot += "    <tr> <td bgcolor=\"white\" > " + byteToStr(carpeta.B_content[c].B_name[:]) + " </td> <td bgcolor=\"white\"  port=\"f" + strconv.Itoa(port) + "\"> " + string(carpeta.B_content[c].B_inodo) + " </td></tr>\n"
										port++
									}

									RepDot += "   </table>>]\n\n"

									// Relacion de bloques a inodos
									for c := 0; c < 4; c++ {
										if carpeta.B_content[c].B_inodo != -1 {

											if strings.Compare(byteToStr(carpeta.B_content[c].B_name[:]), ".") != 0 && strings.Compare(byteToStr(carpeta.B_content[c].B_name[:]), "..") != 0 {
												RepDot += "    bloque_" + string(apuntador.B_pointers[x]) + ":f" + strconv.Itoa(c) + " -> inodo_" + string(carpeta.B_content[c].B_inodo) + ";\n"
											}
										}
									}
								} else if buffer == "2" {
									disco_actual.Seek(int64(superB.S_block_start)+int64(ba_size)*int64(apuntador.B_pointers[x]), 0)
									// --------Se extrae el Bloque Archivo del disco---------
									var basize int = int(binary.Size(archivo))
									data := leerEnFILE(disco_actual, basize)
									buff := bytes.NewBuffer(data)
									err = binary.Read(buff, binary.BigEndian, &archivo)
									if err != nil {
										// Consola += "Binary.Read failed\n"
										// msg_error(err)
										fmt.Println("Binary.Read failed")
									}

									RepDot += "    bloque_" + string(apuntador.B_pointers[x]) + " [shape=plaintext fontname=\"Century Gothic\" label=< \n"
									RepDot += "   <table border=\"0\" bgcolor=\"sandybrown\">\n"
									RepDot += "    <tr> <td> <b>File block " + string(apuntador.B_pointers[x]) + "</b></td></tr>\n"
									RepDot += "    <tr> <td bgcolor=\"white\"> " + byteToStr(archivo.B_content[:]) + " </td></tr>\n"
									RepDot += "   </table>>]\n\n"

								}
								//else if buffer == "3" {
								// 	// NO SE IMPLEMENTO
								// 	Consola += ""
								// }
							}
						}

						// Relacion de bloques apuntador a bloques archivos/carpetas
						for b := 0; b < 16; b++ {
							// El 255 representa al -1
							if int(apuntador.B_pointers[b]) != 255 {
								RepDot += "    bloque_" + strconv.Itoa(int(inodo.I_block[j])) + ":f" + strconv.Itoa(b) + " -> bloque_" + string(apuntador.B_pointers[b]) + ";\n"
							}
						}
					}
					// Relacion de inodos a bloques
					RepDot += "    inodo_" + strconv.Itoa(i) + ":f" + strconv.Itoa(j) + " -> bloque_" + strconv.Itoa(int(inodo.I_block[j])) + "; \n"
				}
			}
		}
		i++

	}
	elapsed := time.Since(start)
	fmt.Println("myFunction() took:", elapsed)

	RepDot += "\n\n}"
	disco_actual.Close()

	//Consola += "Reporte Tree generado con exito!\n"
	fmt.Println("Reporte Tree generado con exito!")

	return RepDot, nil
}

/*
Metodo que lee un byte del archivo en la posicion en donde se encuentra el puntero
*/
func getc(f *os.File) byte {
	b := make([]byte, 1)
	_, err := f.Read(b)

	if err != nil { //si es error lo reportamos
		fmt.Println("getc error:", err)
	}

	return b[0]
}

func byteToStr(array []byte) string { //paso de []byte a string (SIRVE EN ESPECIAL PARA OBTENER UN VALOR NUMERICO)
	contador := 0
	str := ""
	for {
		if contador == len(array) { //significa que termine de recorrel el array
			break
		} else {
			//if array[contador] == uint8(56) { //no hago nada
			//str += "0"
			//}
			if array[contador] == uint8(0) {
				array[contador] = uint8(0) //asigno \00 (creo) y finalizo
				break
			} else if array[contador] != 0 {
				str += string(array[contador]) //le agrego a mi cadena un valor real
			}
		}
		contador++
	}

	return str
}
func byteToInt(part []byte) int {

	fus := -7777777777777777777 //numero malo xd

	ff1 := byteToStr(part[:]) //el tam del DD q posee el mbr lo obtengo en string
	//fmt.Println("/////desencadenando     " + ff1)
	partSize, err := strconv.Atoi(ff1) //valor string lo convierto a int
	//comprobamos q no exista error
	if err != nil {
		//msg_error(err)
		return fus
	}
	fus = partSize

	return fus
}

func leerEnFILE(file *os.File, n int) []byte { //leemos n bytes del DD y lo devolvemos
	Arraybytes := make([]byte, n)   //molde q contendra lo q leemos
	_, err := file.Read(Arraybytes) // recogemos la info q nos interesa y la guardamos en el molde

	if err != nil { //si es error lo reportamos
		fmt.Println(err)
	}
	return Arraybytes
}
