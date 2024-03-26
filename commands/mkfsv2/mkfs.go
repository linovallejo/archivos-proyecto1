package MkfsV2

import (
	//strct "File_Manager_GO/structs"
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	Types "proyecto1/types"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

type PARTICIONMONTADA struct {
	Letra  [1]byte
	Estado int
	Nombre string
}

type DISCOMONTADO struct {
	Path        string
	Numero      int
	Estado      int
	Particiones [99]PARTICIONMONTADA
}

var Discos [26]DISCOMONTADO

func EjecutarMKFS(id string, type_format string, file_system string) {

	// Se limpia la cadena Consola para recolectar informacion de una nueva ejecucion
	//Consola = ""
	sizePartition := 0
	startPartition := 0
	error_ := 0
	pathDisco_Partition := ""

	// Obtener todos los datos por referencia relacionados con la particion montanda en la cual se formateara el sistema de archivos
	getDatosID(id, &pathDisco_Partition, &startPartition, &sizePartition, &error_)

	// Si 'error_' obtiene un 1, el path del ID montado no existe
	if error_ == 1 {
		return
	}

	// SE FORMATEAN SOLO PRIMARIAS, LAS LOGICAS NO SE IMPLEMENTARON

	if file_system == "3fs" {
		formatoEXT3(startPartition, sizePartition, pathDisco_Partition)
	} else {
		formatoEXT2(startPartition, sizePartition, pathDisco_Partition)
	}
}

func getDatosID(id string, path *string, part_startPartition *int, part_sizePartition *int, error_ *int) {

	// Ejemplo ID: 6413f
	numDiskString := id[2 : len(id)-1]          // Ejemplo: 13 en string
	numDisk, err := strconv.Atoi(numDiskString) // Ejemplo: 13 en int
	if err != nil {
		//msg_error(err)
		*error_ = 1
		return
	}
	letraPartition := id[len(id)-1] // Ejemplo: f

	existePath := false
	pathDisco := ""
	sizePartition := -1
	startPartition := -1
	nombreParticion := ""

	// Obtener el nombre de la particion y el path del disco
	for i := 0; i < 26; i++ {

		// NOTA: El arreglo [26]Discos esta declarado global desde mount.go, por ello se puede utilizar aqui
		// NOTA: Esto se puede pues los archivos pertenecen al mismo package
		if Discos[i].Numero == numDisk {

			for j := 0; j < 99; j++ {

				if string(Discos[i].Particiones[j].Letra[:]) == string(letraPartition) && Discos[i].Particiones[j].Estado == 1 {
					pathDisco = Discos[i].Path
					nombreParticion = Discos[i].Particiones[j].Nombre
					existePath = true
					break
				}
			}
		}
	}

	if !existePath {
		//Consola += "Error: ID no reconocido o la particion no esta montada\n"
		*error_ = 1
		return
	}

	// Obtener el part_start de la particion y su tamaño
	disco_actual, err := os.OpenFile(pathDisco, os.O_RDWR, 0660)
	if err != nil {
		//msg_error(err)
		*error_ = 1
		return
	}
	defer disco_actual.Close()

	mbr_auxiliar := Types.MBR{}
	// --------Se extrae el MBR del disco---------
	var size int = int(binary.Size(mbr_auxiliar))
	disco_actual.Seek(0, 0)
	data := leerEnFILE(disco_actual, size)
	buffer := bytes.NewBuffer(data)
	err = binary.Read(buffer, binary.BigEndian, &mbr_auxiliar)
	if err != nil {
		//Consola += "Binary.Read failed\n"
		//msg_error(err)
		fmt.Println(err)
	}

	// Solo se busca en las particiones primarias
	// NOTA: no se implementa la creacion del sistema de archivos en las logicas
	for i := 0; i < 4; i++ {

		if strings.Compare(string(mbr_auxiliar.Partitions[i].Name[:]), nombreParticion) == 0 {
			sizePartition = int(mbr_auxiliar.Partitions[i].Size)
			startPartition = int(mbr_auxiliar.Partitions[i].Start)
		}
	}

	disco_actual.Close()

	*error_ = 0
	*path = pathDisco
	*part_startPartition = startPartition
	*part_sizePartition = sizePartition
}

/*
Metodo encargado de formatear una particion con formato EXT2

	@param int part_start: Byte donde inicia la particion en el disco
	@param int part_size: Tamano de la particion
	@param string path: ruta del disco
*/
func formatoEXT2(part_start int, part_size int, path string) {
	// Tamaño de cada estructura para el calculo
	var sb Types.SuperBlock
	const sb_size = unsafe.Sizeof(sb)

	var inodo Types.Inode
	const i_size = unsafe.Sizeof(inodo)

	var ba Types.FileBlock
	const ba_size = unsafe.Sizeof(ba)

	var bc Types.DirectoryBlock
	const bc_size = unsafe.Sizeof(bc)

	// Numero de inodos sacado de la formula
	n := (part_size - int(sb_size)) / (4 + int(i_size) + 3*int(ba_size))
	num_bloques := 3 * n

	// Fecha actual
	fechaActual := time.Now().String()

	// ESCRITURA EN EL SUPER BLOQUE
	copy(sb.S_filesystem_type[:], "2")
	copy(sb.S_inodes_count[:], strconv.Itoa(n))
	copy(sb.S_blocks_count[:], strconv.Itoa(num_bloques))
	copy(sb.S_free_blocks_count[:], strconv.Itoa(num_bloques-2))
	copy(sb.S_free_inodes_count[:], strconv.Itoa(n-2))
	copy(sb.S_mtime[:], fechaActual)
	copy(sb.S_umtime[:], fechaActual)
	copy(sb.S_mnt_count[:], "0")
	copy(sb.S_magic[:], strconv.Itoa(int(0xEF53)))
	copy(sb.S_inode_size[:], strconv.Itoa(int(i_size)))
	copy(sb.S_block_size[:], strconv.Itoa(int(ba_size)))
	copy(sb.S_first_ino[:], "2")
	copy(sb.S_first_blo[:], "2")
	copy(sb.S_bm_inode_start[:], strconv.Itoa(part_start+int(sb_size)))
	copy(sb.S_bm_block_start[:], strconv.Itoa(part_start+int(sb_size)+n))
	copy(sb.S_inode_start[:], strconv.Itoa(part_start+int(sb_size)+n+num_bloques))
	copy(sb.S_block_start[:], strconv.Itoa(part_start+int(sb_size)+n+num_bloques+(int(i_size)*n)))

	// Instancia de un inodo para los datos de la carpeta raiz
	inodoC := Types.Inode{}
	// Instancia de un inodo para los datos del archivo user.txt
	inodoA := Types.Inode{}
	// Instancia de un bloque de carpetas para los datos de la carpete raiz
	bloqueC := Types.DirectoryBlock{}
	// Instancia de un bloque de archivos para los datos del archivo user.txt
	bloqueA := Types.FileBlock{}

	disco_actual, err := os.OpenFile(path, os.O_RDWR, 0660)
	if err != nil {
		// msg_error(err)
		fmt.Println(err)
		return
	}
	defer disco_actual.Close()

	/*--------------------- SUPERBLOQUE ------------------------*/
	//Se escribe el superbloque al inicio de la particion
	disco_actual.Seek(int64(part_start), 0)
	s1 := &sb //creo pointer
	var binario3 bytes.Buffer
	binary.Write(&binario3, binary.BigEndian, s1)
	escribirDentroFILE(disco_actual, binario3.Bytes()) //meto el superbloque en el inicio de la particion

	/*--------------------- BITMAP DE INODOS ------------------------*/
	// SB -> BMI se rellena de 0's
	// Array que almacenara los bits maps inodos
	var BitMap_ino = make([]byte, n)
	// Inicio en 2 pues esos inodos (0 y 1) son para la carpeta y el archivo user.txt
	for i := 2; i < n; i++ {
		BitMap_ino[i] = 0
	}
	/*--------------------- BIT PARA / Y USER.TXT EN BM ------------------------*/
	//Se ocupan los inodos de ambos bit maps debido a carpeta y archivo iniciales
	BitMap_ino[0] = 1 //carpeta root por parte de usuario de serie
	BitMap_ino[1] = 1 //user.txt por parte de usuario de serie

	// Guardar el bit map inodos por ende calculo su posicion fisica
	StartBitMap_ino := part_start + int(sb_size)
	disco_actual.Seek(int64(StartBitMap_ino), 0)

	s2 := &BitMap_ino
	var binario4 bytes.Buffer
	binary.Write(&binario4, binary.BigEndian, s2)
	escribirDentroFILE(disco_actual, binario4.Bytes())

	/*--------------------- BITMAP DE BLOQUES ------------------------*/
	// SB -> BMI -> BMB se rellena de 0's
	// Array que almacenara los bits maps bloques
	var BitMap_blq = make([]byte, num_bloques)
	// Inicio en 2 pues esos bloques (0 y 1) son para la carpeta y el archivo user.txt
	for i := 2; i < num_bloques; i++ {
		BitMap_blq[i] = 0
	}
	/*--------------------- BIT PARA / Y USER.TXT EN BM ------------------------*/
	// Defino que tipo de objetos se guardan en sincronia con el bitmap inodos que guarda el creador
	BitMap_blq[0] = 1 //1 carpeta para diferenciarlos en reporte etc.
	BitMap_blq[1] = 2 //2 archivo
	//Guardar el bit map bloque por ende calculo la posicion con ayuda de la posicion bitMap inodos
	StartBitMap_blq := StartBitMap_ino + n
	disco_actual.Seek(int64(StartBitMap_blq), 0)

	s3 := &BitMap_blq
	var binario5 bytes.Buffer
	binary.Write(&binario5, binary.BigEndian, s3)
	escribirDentroFILE(disco_actual, binario5.Bytes())

	/*-----------------INODO PARA CARPETA ROOT -----------------------*/
	copy(inodoC.I_uid[:], "1")  // User ID propietario
	copy(inodoC.I_gid[:], "1")  // Group ID
	copy(inodoC.I_size[:], "0") // Para las carpetas siempre sera 0
	copy(inodoC.I_atime[:], fechaActual)
	copy(inodoC.I_ctime[:], fechaActual)
	copy(inodoC.I_mtime[:], fechaActual)
	inodoC.I_block[0] = 0 // Apuntador al bloqueCarpeta de ROOT
	for i := 1; i < 15; i++ {
		// El 255 simularia el -1, posicion default
		inodoC.I_block[i] = 255
	}
	copy(inodoC.I_type[:], "0") // 0, es carpeta
	copy(inodoC.I_perm[:], "664")

	// Se guarda el struct del inodo de la carpeta root
	disco_actual.Seek(int64(sb.S_inode_start), 0)
	sInodoC := &inodoC
	var binario6 bytes.Buffer
	binary.Write(&binario6, binary.BigEndian, sInodoC)
	escribirDentroFILE(disco_actual, binario6.Bytes())

	/*----------------- BLOQUE PARA CARPETA ROOT -----------------------*/
	copy(bloqueC.B_content[0].B_name[:], ".")  // Representa al bloque actual
	copy(bloqueC.B_content[0].B_inodo[:], "0") // Apuntador que en este caso coincide con carpetaRoot.i_block[0] = 0, esto por default

	copy(bloqueC.B_content[1].B_name[:], "..") // Representa al bloque padre
	copy(bloqueC.B_content[1].B_inodo[:], "0")

	copy(bloqueC.B_content[2].B_name[:], "users.txt") // Nombre del archivo que contiene
	copy(bloqueC.B_content[2].B_inodo[:], "1")        // Apuntador hacia el inodo asociado

	copy(bloqueC.B_content[3].B_name[:], ".") // Reset
	copy(bloqueC.B_content[3].B_inodo[:], "-1")

	// Se guarda el struct del bloque de la carpeta root
	disco_actual.Seek(int64(sb.S_block_start), 0)
	sBloqueC := &bloqueC
	var binario7 bytes.Buffer
	binary.Write(&binario7, binary.BigEndian, sBloqueC)
	escribirDentroFILE(disco_actual, binario7.Bytes())

	/*-----------------INODO PARA USERS.TXT -----------------------*/
	contenido := "1,G,root\n1,U,root,root,123\n" // Data que tendra el archivo users.txt
	content_size := len(contenido)

	copy(inodoA.I_uid[:], "1")
	copy(inodoA.I_gid[:], "1")
	copy(inodoA.I_size[:], strconv.Itoa(content_size)) // Se puede poner por default 27 o se calcula el tamaño
	copy(inodoA.I_atime[:], fechaActual)
	copy(inodoA.I_ctime[:], fechaActual)
	copy(inodoA.I_mtime[:], fechaActual)
	inodoA.I_block[0] = 1 // Apuntador al archivo .txt
	for i := 1; i < 15; i++ {
		// El 255 simularia el -1, posicion default
		inodoA.I_block[i] = 255
	}
	copy(inodoA.I_type[:], "1") // 1, es archivo
	copy(inodoA.I_perm[:], "664")

	// Se guarda el struct del inodo del archivo users.txt
	// Al inode_start se le suma el tamaño de otro inodo struct, pues en el start esta el struct de Carpeta root
	disco_actual.Seek(int64(sb.S_inode_start)+int64(i_size), 0)
	sInodoA := &inodoA
	var binario8 bytes.Buffer
	binary.Write(&binario8, binary.BigEndian, sInodoA)
	escribirDentroFILE(disco_actual, binario8.Bytes())

	/*----------------- BLOQUE PARA USERS.TXT -----------------------*/
	copy(bloqueA.B_content[:], contenido)

	disco_actual.Seek(int64(sb.S_block_start)+int64(bc_size), 0)
	sBloqueA := &bloqueA
	var binario9 bytes.Buffer
	binary.Write(&binario9, binary.BigEndian, sBloqueA)
	escribirDentroFILE(disco_actual, binario9.Bytes())

	// Consola += "EXT2\n"
	// Consola += "...\n"
	// Consola += "Disco formateado con exito\n"

	disco_actual.Close()
}

/*
Metodo encargado de formatear una particion con formato EXT3

	@param int part_start: Byte donde inicia la particion en el disco
	@param int part_size: Tamano de la particion
	@param string path: ruta del disco
*/
func formatoEXT3(part_start int, part_size int, path string) {
	//Consola += "Por peticion, no se implemento 3FS"
	return
}

func leerEnFILE(file *os.File, n int) []byte { //leemos n bytes del DD y lo devolvemos
	Arraybytes := make([]byte, n)   //molde q contendra lo q leemos
	_, err := file.Read(Arraybytes) // recogemos la info q nos interesa y la guardamos en el molde

	if err != nil { //si es error lo reportamos
		fmt.Println(err)
	}
	return Arraybytes
}

func escribirDentroFILE(file *os.File, bytes []byte) { //escribe dentro de un file
	_, err := file.Write(bytes)

	if err != nil {
		//msg_error(err)
		fmt.Println(err)
	}
}
