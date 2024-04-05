package userworkspace

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	Fdisk "proyecto1/commands/fdisk"
	Global "proyecto1/global"
	Types "proyecto1/types"
	Utilities "proyecto1/utils"
	"strconv"
	"strings"
	"time"
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
	if Global.Usuario.Status {

		// Validacion que sea el Usuario ROOT
		if CurrentSession.Id_user == 1 && CurrentSession.Id_grp == 1 {

			// Validacion si el grupo ya existe
			result, err := buscarGrupo(groupName, pathUsersFile, CurrentSession.PartitionId, CurrentSession.Path)
			if err != nil {
				return "", fmt.Errorf("Error: %s\n", err)
			}
			if result != 1 {
				newGrp_id, _ := nuevoGrupoId(pathUsersFile, CurrentSession.PartitionId, CurrentSession.Path)
				nuevoGrupo := strconv.Itoa(newGrp_id) + ",G," + groupName + "\n"
				fmt.Printf("nuevoGrupo: %s\n", nuevoGrupo)
				setToFileUsersTxt(nuevoGrupo, pathUsersFile, CurrentSession.PartitionId, CurrentSession.Path)
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
// func actualizarArchivoUsersTxt(contentToAdd string, pathUsersFile string, partitionId string, diskFileName string) (int, error) {

// 	fileContents, err := ReturnFileContents(pathUsersFile, partitionId, diskFileName)
// 	if err != nil {
// 		return 0, err
// 	}

// 	var objectType string = ""
// 	var objectId int = 0

// 	content := byteToStr(archivo.B_content[:])
// 	blockInUse := len(content)
// 	blockFree := 63 - blockInUse
// 	contador := 0
// 	escribir := len(newData)

// 	for _, filaU_G := range fileContents {
// 		cleanString := strings.ReplaceAll(filaU_G, "\x00", "")
// 		parts := strings.Split(cleanString, ",")
// 		if cleanString != "" && len(parts) > 1 {
// 			objectType = strings.TrimSpace(parts[1])
// 			if strings.Compare(objectType, "G") == 0 {
// 				objectId, _ = strconv.Atoi(strings.TrimSpace(parts[0]))
// 			}
// 		}
// 		// parts := strings.Split(filaU_G, ",")
// 		// objectType = strings.TrimSpace(parts[1])
// 		// if strings.Compare(objectType, "G") == 0 {
// 		// 	objectId, _ = strconv.Atoi(strings.TrimSpace(parts[0]))
// 		// }
// 	}

// 	return objectId + 1, nil
// }

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

func setToFileUsersTxt(newData string, pathUsersFile string, partitionId string, diskFileName string) ([]string, error) {
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
		defer file.Close()
		return nil, fmt.Errorf("Error abriendo el archivo")
	}

	defer file.Close()

	var tempSuperblock Types.SuperBlock
	// Read object from bin file
	if err := Utilities.ReadObject(file, &tempSuperblock, int64(TempMBR.Partitions[index].Start)); err != nil {
		return nil, fmt.Errorf("Error leyendo el superbloque")
	}

	TempStepsPath := strings.Split(pathUsersFile, string(filepath.Separator))
	StepsPath := TempStepsPath[1:]

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

	var archivo *Types.FileBlock
	archivo = GetFileBlockDataOriginal(tempInode, file, tempSuperblock)
	if archivo == nil {
		//fmt.Println("User not found")
		return nil, fmt.Errorf("User not found")
	}

	currentSize := tempInode.I_size
	// Calcular si el bloque tipo archivo tiene aun espacio para escribir
	blockFree := int(len(tempInode.I_block))*64 - int(currentSize)
	if blockFree < len(newData) {
		fmt.Println("No hay suficiente espacio")
		return nil, fmt.Errorf("No hay suficiente espacio")
	} else {
		//fmt.Println("Hay suficiente espacio")
		offset := currentSize
		for i := 0; i < len(newData); i++ {
			archivo.B_content[offset] = newData[i]
			offset++
			if newData[i] == '\n' {
				break
			}
		}
		tempInode.I_size += int32(len(newData))
		var today string = time.Now().Format("02/01/2006")
		copy(tempInode.I_mtime[:], today)
		//fmt.Printf("Inode: %+v\n", tempInode)
		//fmt.Printf("FileBlock: %+v\n", archivo)
		// Write object to bin file
		if err := Utilities.WriteObject(file, &tempInode, int64(tempSuperblock.S_inode_start+indexInode*int32(binary.Size(Types.Inode{})))); err != nil {
			return nil, fmt.Errorf("Error escribiendo el inode")
		} else {
			fmt.Println("Inode written")
		}

		if err := Utilities.WriteObject(file, &archivo, int64(tempSuperblock.S_block_start+tempInode.I_block[0]*int32(binary.Size(Types.FileBlock{})))); err != nil {
			return nil, fmt.Errorf("Error escribiendo el archivo")
		} else {
			fmt.Println("FileBlock written")
		}
	}

	return []string{newData}, nil
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
