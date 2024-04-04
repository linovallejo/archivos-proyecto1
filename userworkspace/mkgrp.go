package userworkspace

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	Global "proyecto1/global"
	Types "proyecto1/types"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

var CurrentSession Global.Sesion

func ExtractMkgrpParams(params []string) (string, error) {
	var name string = ""

	if len(params) == 0 {
		return "", fmt.Errorf("No se encontraron parámetros")
	}
	var parametrosObligatoriosOk bool = false
	nameOk := false
	for _, param1 := range params {
		if strings.HasPrefix(param1, "-name=") {
			nameOk = true
		}
	}

	parametrosObligatoriosOk = nameOk

	if !parametrosObligatoriosOk {
		return "", fmt.Errorf("No se encontraron parámetros obligatorios")
	}

	for _, param := range params {
		switch {
		case strings.HasPrefix(param, "-name="):
			name = strings.TrimPrefix(param, "-name=")
		}
	}

	return name, nil
}

func EjecutarMkgrp(groupName string, pathUsersFile string, CurrentSession Global.Sesion) (string, error) {

	//if IsLoginFlag {
	if true {

		// Validacion que sea el Usuario ROOT
		//if CurrentSession.Id_user == 1 && CurrentSession.Id_grp == 1 {
		if true {

			// Validacion si el grupo ya existe
			result, err := buscarGrupo(groupName, pathUsersFile, CurrentSession.PartitionId, CurrentSession.Path)
			if err != nil {
				return "", fmt.Errorf("Error: %s\n", err)
			}
			if result != 1 {
				newGrp_id, _ := nuevoGrupoId(pathUsersFile, CurrentSession.PartitionId, CurrentSession.Path)
				nuevoGrupo := strconv.Itoa(newGrp_id) + ",G," + groupName + "\n"
				fmt.Printf("nuevoGrupo: %s\n", nuevoGrupo)
				//setToFileUsersTxt(nuevoGrupo)
				return "", fmt.Errorf("Grupo creado con exito!\n")

			} else {
				return "", fmt.Errorf("Error: Ya existe un grupo con ese nombre!\n")
			}

		} else {
			return "", fmt.Errorf("Error: Solo el usuario root puede ejecutar este comando!\n")
		}

	} else {
		return "", fmt.Errorf("Error: Necesita iniciar sesion para poder ejecutar este comando!\n")
	}
}

/* Metodo para agregar un grupo/usuario al archivo users.txt de una particion
 * @param string newData: Datos del nuevo grupo/usuario
 */

func nuevoGrupoId(pathUsersFile string, partitionId string, diskFileName string) (int, error) {

	fileContents, err := ReturnFileContents(pathUsersFile, partitionId, diskFileName)
	if err != nil {
		return 0, err
	}

	var objectType string = ""
	var objectId int = 0

	for _, filaU_G := range fileContents {
		cleanString := strings.ReplaceAll(filaU_G, "\x00", "")
		parts := strings.Split(cleanString, ",")
		if cleanString != "" && len(parts) > 1 {
			objectType = strings.TrimSpace(parts[1])
			if strings.Compare(objectType, "G") == 0 {
				objectId, _ = strconv.Atoi(strings.TrimSpace(parts[0]))
			}
		}
		// parts := strings.Split(filaU_G, ",")
		// objectType = strings.TrimSpace(parts[1])
		// if strings.Compare(objectType, "G") == 0 {
		// 	objectId, _ = strconv.Atoi(strings.TrimSpace(parts[0]))
		// }
	}

	return objectId + 1, nil
}

func buscarGrupo(groupName string, pathUsersFile string, partitionId string, diskFileName string) (int, error) {

	fileContents, err := ReturnFileContents(pathUsersFile, partitionId, diskFileName)
	if err != nil {
		return 0, err
	}

	var objectType string = ""
	var name string = ""

	for _, filaU_G := range fileContents {
		cleanString := strings.ReplaceAll(filaU_G, "\x00", "")
		parts := strings.Split(cleanString, ",")
		if cleanString != "" && len(parts) > 1 {
			objectType = strings.TrimSpace(parts[1])
			name = strings.TrimSpace(parts[2])
			if strings.Compare(objectType, "G") == 0 {
				if strings.Compare(name, groupName) == 0 {
					return 1, nil
				}
			}
		}

	}

	return 0, nil

	// // Apertura del archivo del disco binario
	// disco_actual, err := os.OpenFile(Global.SesionActual.Path, os.O_RDWR, 0660)
	// if err != nil {
	// 	fmt.Printf("Error: %s\n", err)
	// }
	// defer disco_actual.Close()

	// // Estructuras necesarias a utilizar
	// superB := Types.SuperBlock{}
	// inodo := Types.Inode{}

	// // Tamaño de algunas estructuras
	// var inodoTable Types.Inode
	// const i_size = unsafe.Sizeof(inodoTable)

	// // --------Se extrae el SB del disco---------
	// var sbsize int = int(binary.Size(superB))
	// disco_actual.Seek(int64(CurrentSession.PartitionStart), 0)
	// data := leerEnArchivo(disco_actual, sbsize)
	// buffer := bytes.NewBuffer(data)
	// err = binary.Read(buffer, binary.LittleEndian, &superB)
	// if err != nil {
	// 	fmt.Printf("Binary.Read failed: %s\n", err)
	// }

	// // --------Se extrae el InodoTable del archivo user.txt---------
	// var inodosize int = int(binary.Size(inodo))
	// disco_actual.Seek(int64(superB.S_inode_start)+int64(i_size), 0)
	// dataI := leerEnArchivo(disco_actual, inodosize)
	// bufferI := bytes.NewBuffer(dataI)
	// err = binary.Read(bufferI, binary.LittleEndian, &inodo)
	// if err != nil {
	// 	fmt.Printf("Binary.Read failed: %s\n", err)
	// }

	// // Almacenara lo extraido del archivo
	// contenidoFile := ""

	// // Se recorren los iblock para conocer los punteros
	// for i := 0; i < 15; i++ {

	// 	// El 255 representa al -1
	// 	if int(inodo.I_block[i]) != 255 {
	// 		archivo := Types.FileBlock{}
	// 		disco_actual.Seek(int64(superB.S_block_start), 0)

	// 		for j := 0; j <= int(inodo.I_block[i]); j++ {
	// 			// --------Se extrae el Bloque Archivo USERS.TXT---------
	// 			var basize int = int(binary.Size(archivo))
	// 			data := leerEnArchivo(disco_actual, basize)
	// 			buff := bytes.NewBuffer(data)
	// 			err = binary.Read(buff, binary.LittleEndian, &archivo)
	// 			if err != nil {
	// 				fmt.Printf("Binary.Read failed: %s\n", err)
	// 			}
	// 		}

	// 		//Concatenar el contenido de cada bloque perteneciente al archivo users.txt
	// 		contenidoFile += byteToStr(archivo.B_content[:])
	// 	}
	// }

	// disco_actual.Close()

	// var arregloU_G []string = strings.Split(contenidoFile, "\n")

	// for _, filaU_G := range arregloU_G {

	// 	// Se verifica que la fila obtenida del contenido no venga vacia
	// 	if filaU_G != "" {
	// 		var data []string = strings.Split(filaU_G, ",")

	// 		// Verificar ID que no se un U/G eliminado
	// 		if strings.Compare(data[0], "0") != 0 {

	// 			// Verificar que sea tipo Grupo
	// 			if strings.Compare(data[1], "G") == 0 {
	// 				group := data[2]
	// 				if strings.Compare(group, grp_name) == 0 {
	// 					idG, _ := strconv.Atoi(data[0])
	// 					return idG
	// 				}

	// 			}
	// 		}
	// 	}
	// }

	// return -1
}

func getc(f *os.File) byte {
	b := make([]byte, 1)
	_, err := f.Read(b)

	if err != nil {
		fmt.Println(err)
	}

	return b[0]
}

func leerEnArchivo(file *os.File, n int) []byte { //leemos n bytes del DD y lo devolvemos
	Arraybytes := make([]byte, n)   //molde q contendra lo q leemos
	_, err := file.Read(Arraybytes) // recogemos la info q nos interesa y la guardamos en el molde

	if err != nil { //si es error lo reportamos
		fmt.Println(err)
	}
	return Arraybytes
}

func escribirEnArchivo(file *os.File, bytes []byte) { //escribe dentro de un file
	_, err := file.Write(bytes)

	if err != nil {
		fmt.Println(err)
	}
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

func setToFileUsersTxt(newData string) {
	// Apertura del archivo del disco binario
	disco_actual, err := os.OpenFile(CurrentSession.Path, os.O_RDWR, 0660)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}
	defer disco_actual.Close()

	// Estructuras necesarias a utilizar
	superB := Types.SuperBlock{}
	inodo := Types.Inode{}
	inodo1 := Types.Inode{}
	archivo := Types.FileBlock{}

	// Tamaño de algunas estructuras
	var inodoTable Types.Inode
	const i_size = unsafe.Sizeof(inodoTable)

	var ba Types.FileBlock
	const ba_size = unsafe.Sizeof(ba)

	// --------Se extrae el SB del disco---------
	var sbsize int = int(binary.Size(superB))
	disco_actual.Seek(int64(CurrentSession.PartitionStart), 0)
	data := leerEnArchivo(disco_actual, sbsize)
	buffer := bytes.NewBuffer(data)
	err = binary.Read(buffer, binary.BigEndian, &superB)
	if err != nil {
		fmt.Printf("Binary.Read failed: %s\n", err)
	}

	// --------Se extrae el InodoTable del archivo user.txt---------
	var inodosize int = int(binary.Size(inodo))
	disco_actual.Seek(int64(superB.S_inode_start)+int64(i_size), 0)
	dataI := leerEnArchivo(disco_actual, inodosize)
	bufferI := bytes.NewBuffer(dataI)
	err = binary.Read(bufferI, binary.BigEndian, &inodo)
	if err != nil {
		fmt.Printf("Binary.Read failed: %s\n", err)
	}

	blockIndex := 0

	for i := 0; i < 12; i++ {

		// El 255 representa al -1
		if int(inodo.I_block[i]) != 255 {
			blockIndex = int(inodo.I_block[i]) // Indice del ultimo bloque utilizado del archivo
		}
	}

	disco_actual.Seek(int64(superB.S_block_start)+int64(ba_size)*int64(blockIndex), 0)
	// --------Se extrae un Bloque Archivo---------
	var basize int = int(binary.Size(archivo))
	dataA := leerEnArchivo(disco_actual, basize)
	buff := bytes.NewBuffer(dataA)
	err = binary.Read(buff, binary.BigEndian, &archivo)
	if err != nil {
		fmt.Printf("Binary.Read failed: %s\n", err)
	}

	// Calcular si el bloque tipo archivo tiene aun espacio para escribir
	content := byteToStr(archivo.B_content[:])
	blockInUse := len(content)
	blockFree := 63 - blockInUse
	contador := 0
	escribir := len(newData)

	// Si la nueva data aun cabe en el bloque
	if escribir <= blockFree {
		// Escribir byte a byte la newData
		for i := 0; i < 64; i++ {

			if archivo.B_content[i] == 0 && contador < escribir {
				archivo.B_content[i] = newData[contador]

				if newData[contador] == '\n' {
					contador = escribir
					break
				}
				contador++
			}
		}

		// Almacenar el bloque archivo modificado
		disco_actual.Seek(int64(superB.S_block_start)+int64(ba_size)*int64(blockIndex), 0)
		sBloqueA := &archivo
		var binario1 bytes.Buffer
		binary.Write(&binario1, binary.BigEndian, sBloqueA)
		escribirEnArchivo(disco_actual, binario1.Bytes())

		disco_actual.Seek(int64(superB.S_inode_start)+int64(i_size), 0)
		var inodosize1 int = int(binary.Size(inodo1))
		dataI1 := leerEnArchivo(disco_actual, inodosize1)
		bufferI1 := bytes.NewBuffer(dataI1)
		err = binary.Read(bufferI1, binary.BigEndian, &inodo1)
		if err != nil {
			fmt.Printf("Binary.Read failed: %s\n", err)
		}

		// Actualizacion del tamaño del inodo y la fecha
		inodo1.I_size = int32(inodo1.I_size + int32(escribir))
		copy(inodo1.I_mtime[:], time.Now().String())

		// Almacenar el inodo actualizado
		disco_actual.Seek(int64(superB.S_inode_start)+int64(i_size), 0)
		sInodo1 := &inodo1
		var binario2 bytes.Buffer
		binary.Write(&binario2, binary.BigEndian, sInodo1)
		escribirEnArchivo(disco_actual, binario2.Bytes())

	} else {
		aux := ""
		aux2 := ""
		// Este indice se actualiza con el recorrido de los 2 for siguientes
		i := 0

		// Este for no tiene indice nuevo, utiliza el 'i' declarado arriba
		for i <= blockFree {
			aux += string(newData[i])
			i++
		}

		for i < escribir {
			aux2 += string(newData[i])
			i++
		}

		// Guardamos lo que quepa en el primer bloque
		// Escribir byte a byte la newData
		for i := 0; i < 64; i++ {

			if archivo.B_content[i] == 0 && contador < len(aux) {
				archivo.B_content[i] = aux[contador]

				if aux[contador] == '\n' {
					contador = len(aux)
					break
				}
				contador++
			}
		}

		// Esto no hace nada, solo es para desaparecer
		if contador != 1 {
			fmt.Print("")
		}
		// La alerta de que 'contador' no se utiliza =D

		// Almacenar el bloque archivo modificado
		disco_actual.Seek(int64(superB.S_block_start)+int64(ba_size)*int64(blockIndex), 0)
		sBloqueA := &archivo
		var binario1 bytes.Buffer
		binary.Write(&binario1, binary.BigEndian, sBloqueA)
		escribirEnArchivo(disco_actual, binario1.Bytes())

		// Nuevo bloque archivo, se almacena el resto de la newData
		auxArchivo := Types.FileBlock{}
		copy(auxArchivo.B_content[:], aux2)
		bit := buscarBit(disco_actual, "B", string(CurrentSession.Fit[:]))

		// Guardamos el bloque en el bitmap y en la tabla de bloques
		disco_actual.Seek(int64(superB.S_bm_block_start)+int64(bit), 0)
		Fputc('2', disco_actual)

		disco_actual.Seek(int64(superB.S_block_start)+int64(ba_size)*int64(bit), 0)
		sBloqueA1 := &auxArchivo
		var binario2 bytes.Buffer
		binary.Write(&binario2, binary.BigEndian, sBloqueA1)
		escribirEnArchivo(disco_actual, binario2.Bytes())

		// Guardamos el modificado del inodo
		disco_actual.Seek(int64(superB.S_inode_start)+int64(i_size), 0)
		var inodosize1 int = int(binary.Size(inodo1))
		dataI1 := leerEnArchivo(disco_actual, inodosize1)
		bufferI1 := bytes.NewBuffer(dataI1)
		err = binary.Read(bufferI1, binary.BigEndian, &inodo1)
		if err != nil {
			fmt.Printf("Binary.Read failed: %s\n", err)
		}

		// Actualizacion del tamaño del inodo y la fecha
		inodo1.I_size = int32(inodo1.I_size + int32(escribir))
		copy(inodo1.I_mtime[:], time.Now().String())
		copy(inodo1.I_mtime[:], time.Now().String())
		inodo1.I_block[blockIndex] = int32(bit)
		disco_actual.Seek(int64(superB.S_inode_start)+int64(i_size), 0)
		sInodo1 := &inodo1
		var binario3 bytes.Buffer
		binary.Write(&binario3, binary.BigEndian, sInodo1)
		escribirEnArchivo(disco_actual, binario3.Bytes())

		// Guardamos la nueva cantidad de bloques libres y el primer bloque libre
		superB.S_first_blo = superB.S_first_blo + 1
		superB.S_free_blocks_count = superB.S_free_blocks_count - 1
		disco_actual.Seek(int64(CurrentSession.PartitionStart), 0)
		//Se escribe el superbloque al inicio de la particion
		s1 := &superB
		var binario4 bytes.Buffer
		binary.Write(&binario4, binary.BigEndian, s1)
		escribirEnArchivo(disco_actual, binario4.Bytes()) //meto el superbloque en el inicio de la particion

	}
	disco_actual.Close()

}

/* Funcion que devuelve el bit libre en el bitmap de inodos/bloques segun el ajuste
 * @param FILE fp: archivo en el que se esta leyendo
 * @param string tipo: tipo de bit a buscar (Inodo/Bloque)
 * @param string fit: ajuste de la particion
 * @return -1 = Ya no existen bloques libres | # bit libre en el bitmap
 */
func buscarBit(file *os.File, tipo string, fit string) int {
	super := Types.SuperBlock{}
	inicio_bm := 0
	bit_libre := -1
	tam_bm := 0

	file.Seek(int64(CurrentSession.PartitionStart), 0)
	// --------Se extrae el SB del disco---------
	var sbsize int = int(binary.Size(super))
	data := leerEnArchivo(file, sbsize)
	buffer := bytes.NewBuffer(data)
	err := binary.Read(buffer, binary.BigEndian, &super)
	if err != nil {
		fmt.Printf("Binary.Read failed: %s\n", err)
	}

	// Si es inodo
	if tipo == "I" {
		tam_bm = int(super.S_inodes_count)
		inicio_bm = int(super.S_bm_inode_start)

		// Si es bloque
	} else if tipo == "B" {
		tam_bm = int(super.S_blocks_count)
		inicio_bm = int(super.S_bm_block_start)
	}

	//-------------- Tipo de ajuste a utilizar-------------------
	if fit == "F" { // Primer ajuste

		for i := 0; i < tam_bm; i++ {
			file.Seek(int64(inicio_bm+i), 0)
			dataBMI := getc(file)          // me devuelve el dato en byte
			bufINT := int(dataBMI)         // Lo convierto en int
			buffer := strconv.Itoa(bufINT) // Convierto el int a string

			if buffer == "0" {
				bit_libre = i
				return bit_libre
			}
		}

		if bit_libre == -1 {
			return -1
		}

	} else if fit == "B" { // Mejor ajuste

		libres := 0
		auxLibres := -1

		for i := 0; i < tam_bm; i++ { // Primer recorrido
			file.Seek(int64(inicio_bm+i), 0)
			dataBMI := getc(file)          // me devuelve el dato en byte
			bufINT := int(dataBMI)         // Lo convierto en int
			buffer := strconv.Itoa(bufINT) // Convierto el int a string

			if buffer == "0" {
				libres++
				if i+1 == tam_bm {
					if auxLibres == -1 || auxLibres == 0 {
						auxLibres = libres
					} else {
						if auxLibres > libres {
							auxLibres = libres
						}
					}
					libres = 0
				}
			} else if buffer == "1" {
				if auxLibres == -1 || auxLibres == 0 {
					auxLibres = libres
				} else {
					if auxLibres > libres {
						auxLibres = libres
					}
				}
				libres = 0
			}
		}

		for i := 0; i < tam_bm; i++ { // Segundo recorrido
			file.Seek(int64(inicio_bm+i), 0)
			dataBMI := getc(file)          // me devuelve el dato en byte
			bufINT := int(dataBMI)         // Lo convierto en int
			buffer := strconv.Itoa(bufINT) // Convierto el int a string

			if buffer == "0" {
				libres++
				if i+1 == tam_bm {
					res := (i + 1) - libres
					return res
				}
			} else if buffer == "1" {
				if auxLibres == libres {
					res := (i + 1) - libres
					return res
				}
				libres = 0
			}
		}

		return -1
	} else if fit == "W" { // Peor ajuste
		libres := 0
		auxLibres := -1

		for i := 0; i < tam_bm; i++ { // Primer recorrido
			file.Seek(int64(inicio_bm+i), 0)
			dataBMI := getc(file)          // me devuelve el dato en byte
			bufINT := int(dataBMI)         // Lo convierto en int
			buffer := strconv.Itoa(bufINT) // Convierto el int a string

			if buffer == "0" {
				libres++
				if i+1 == tam_bm {
					if auxLibres == -1 || auxLibres == 0 {
						auxLibres = libres
					} else {
						if auxLibres < libres {
							auxLibres = libres
						}
					}
					libres = 0
				}
			} else if buffer == "1" {
				if auxLibres == -1 || auxLibres == 0 {
					auxLibres = libres
				} else {
					if auxLibres < libres {
						auxLibres = libres
					}
				}
				libres = 0
			}
		}

		for i := 0; i < tam_bm; i++ { // Segundo recorrido
			file.Seek(int64(inicio_bm+i), 0)
			dataBMI := getc(file)          // me devuelve el dato en byte
			bufINT := int(dataBMI)         // Lo convierto en int
			buffer := strconv.Itoa(bufINT) // Convierto el int a string

			if buffer == "0" {
				libres++
				if i+1 == tam_bm {
					res := (i + 1) - libres
					return res
				}
			} else if buffer == "1" {
				if auxLibres == libres {
					res := (i + 1) - libres
					return res
				}
				libres = 0
			}
		}

		return -1
	}

	return 0
}

// Fputc handles fputc().
//
// Writes a character to the stream and advances the position indicator.
//
// The character is written at the position indicated by the internal position
// indicator of the stream, which is then automatically advanced by one.
func Fputc(c int32, f *os.File) int32 {
	n, err := f.Write([]byte{byte(c)})
	if err != nil {
		return 0
	}

	return int32(n)
}

/*
Funcion para obtener un nuevo ID para el nuevo grupo
@return id del ultimo grupo +1
*/
func getNewGrp_id() int {
	// Apertura del archivo del disco binario
	disco_actual, err := os.OpenFile(CurrentSession.Path, os.O_RDWR, 0660)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return 0
	}
	defer disco_actual.Close()

	// Estructuras necesarias a utilizar
	superB := Types.SuperBlock{}
	inodo := Types.Inode{}

	// Tamaño de algunas estructuras
	var inodoTable Types.Inode
	const i_size = unsafe.Sizeof(inodoTable)

	// --------Se extrae el SB del disco---------
	var sbsize int = int(binary.Size(superB))
	disco_actual.Seek(int64(CurrentSession.PartitionStart), 0)
	data := leerEnArchivo(disco_actual, sbsize)
	buffer := bytes.NewBuffer(data)
	err = binary.Read(buffer, binary.BigEndian, &superB)
	if err != nil {
		fmt.Printf("Binary.Read failed: %s\n", err)
	}

	// --------Se extrae el InodoTable del archivo user.txt---------
	var inodosize int = int(binary.Size(inodo))
	disco_actual.Seek(int64(superB.S_inode_start)+int64(i_size), 0)
	dataI := leerEnArchivo(disco_actual, inodosize)
	bufferI := bytes.NewBuffer(dataI)
	err = binary.Read(bufferI, binary.BigEndian, &inodo)
	if err != nil {
		fmt.Printf("Binary.Read failed: %s\n", err)
	}

	// Almacenara lo extraido del archivo
	contenidoFile := ""
	id_auxiliar := -1

	// Se recorren los iblock para conocer los punteros
	for i := 0; i < 15; i++ {

		// El 255 representa al -1
		if int(inodo.I_block[i]) != 255 {
			archivo := Types.FileBlock{}
			disco_actual.Seek(int64(superB.S_block_start), 0)

			for j := 0; j <= int(inodo.I_block[i]); j++ {
				// --------Se extrae el Bloque Archivo USERS.TXT---------
				var basize int = int(binary.Size(archivo))
				data := leerEnArchivo(disco_actual, basize)
				buff := bytes.NewBuffer(data)
				err = binary.Read(buff, binary.BigEndian, &archivo)
				if err != nil {
					fmt.Printf("Binary.Read failed: %s\n", err)
				}
			}

			//Concatenar el contenido de cada bloque perteneciente al archivo users.txt
			contenidoFile += byteToStr(archivo.B_content[:])
		}
	}

	disco_actual.Close()

	var arregloU_G []string = strings.Split(contenidoFile, "\n")

	for _, filaU_G := range arregloU_G {

		// Se verifica que la fila obtenida del contenido no venga vacia
		if filaU_G != "" {
			var data []string = strings.Split(filaU_G, ",")

			// Verificar ID que no se un U/G eliminado
			if strings.Compare(data[0], "0") != 0 {

				// Verificar que sea tipo Grupo
				if strings.Compare(data[1], "G") == 0 {

					idG, _ := strconv.Atoi(data[0])
					id_auxiliar = idG

				}
			}
		}
	}

	// Se retorna el ultimo id del tipo Grupo encontrado sumandole 1
	return id_auxiliar + 1
}
