package Mount

import (
	"fmt"
	Fdisk "proyecto1/commands/fdisk"
	Types "proyecto1/types"
	Utilities "proyecto1/utils"
	"strconv"
	"strings"
	"unicode"
)

func MountPartition(mbr *Types.MBR, diskFileName string, driveletter string, name string) (int, error) {
	// fmt.Println("======Start MOUNT======")
	// fmt.Println("Driveletter:", driveletter)
	// fmt.Println("Name:", name)

	// Open bin file
	file, err := Utilities.AbrirArchivo(diskFileName)
	if err != nil {
		return -1, err
	}

	//fmt.Println("-------------")

	var index int = -1
	var count = 0
	// Iterate over the partitions
	//fmt.Println("################")
	var partitionStatusStr string = ""
	var partitionType string = ""

	for i := 0; i < 4; i++ {
		//fmt.Println("*********************")
		partitionStatus := mbr.Partitions[i].Status[0]
		partitionStatusStr = strconv.Itoa(int(partitionStatus))
		partitionType = string(mbr.Partitions[i].Type[:])

		// fmt.Println(string(mbr.Partitions[i].Id[:]))
		// fmt.Println(string(mbr.Partitions[i].Name[:]))
		// fmt.Println(partitionStatus)
		// fmt.Println(strconv.Itoa(int(mbr.Partitions[i].Size))[:])
		if mbr.Partitions[i].Size != 0 {
			//fmt.Printf("Count: %d\n", count)
			count++
			if strings.Contains(string(mbr.Partitions[i].Name[:]), name) {
				//fmt.Println("Gotcha!")
				index = i
				//fmt.Printf("Index: %d\n", index)
				break
			}
		}
	}

	if index < 0 {
		defer file.Close()
		//fmt.Println("Partition not found")
		return -1, fmt.Errorf("partición no existe")
	}

	if partitionType == "E" || partitionType == "L" {
		return -1, fmt.Errorf("no es posible montar una partición extendida o lógica")
	}

	// if index >= 0 {
	// 	defer file.Close()
	// 	//fmt.Println("Partition not found")
	// 	return -1, fmt.Errorf("partición no existe misha")

	// }

	if partitionStatusStr == "1" {
		defer file.Close()
		return -1, fmt.Errorf("no es posible montar una particion que ya se encuentra montada")
	}

	// id = DriveLetter + Correlative + 19

	id := strings.ToUpper(driveletter) + strconv.Itoa(count) + "23"

	//copy(mbr.Partitions[index].Status[:], "1")
	mbr.Partitions[index].Status[0] = 1
	copy(mbr.Partitions[index].Id[:], id)

	// Overwrite the MBR
	Fdisk.WriteMBR(diskFileName, *mbr)

	var TempMBR2 Types.MBR
	// Read object from bin file
	if err := Utilities.ReadObject(file, &TempMBR2, 0); err != nil {
		defer file.Close()
		return -1, err
	}

	// Print object
	// PrintMBR(TempMBR2)

	// Close bin file
	defer file.Close()

	//fmt.Println("======End MOUNT======")

	return 0, nil
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

func ExtractMountParams(params []string) (string, string, error) {
	var driveletter string = ""
	var name string = ""

	if len(params) == 0 {
		return "", "", fmt.Errorf("No se encontraron parámetros")
	}
	var parametrosObligatoriosOk bool = false
	driveletterOk := false
	nameOk := false
	for _, param1 := range params {
		if strings.HasPrefix(param1, "-driveletter=") {
			driveletterOk = true
		} else if strings.HasPrefix(param1, "-name=") {
			nameOk = true
		}
	}

	if driveletterOk && nameOk {
		parametrosObligatoriosOk = true
	}

	if !parametrosObligatoriosOk {
		return "", "", fmt.Errorf("No se encontraron parámetros obligatorios")
	}

	for _, param := range params {
		switch {
		case strings.HasPrefix(param, "-driveletter="):
			driveletter = strings.TrimPrefix(param, "-driveletter=")
			// Validar la letra de la partición
			if strings.TrimSpace(driveletter) == "" {
				return "", "", fmt.Errorf("La letra del drive es un parametro obligatorio")
			} else if len(driveletter) != 1 || !unicode.IsLetter(rune(driveletter[0])) {
				return "", "", fmt.Errorf("La letra de la partición debe ser un único carácter alfabérico")
			}
		case strings.HasPrefix(param, "-name="):
			name = strings.TrimPrefix(param, "-name=")
			if strings.TrimSpace(name) == "" {
				return "", "", fmt.Errorf("Parametro name es obligatorio")
			}

			// Validar el nombre de la partición
			if len(name) > 16 {
				return "", "", fmt.Errorf("El nombre de la partición no puede exceder los 16 caracteres")
			}

		}

	}

	return driveletter, name, nil
}

func UnmountPartition(mbr *Types.MBR, id string, diskFileName string) (int, error) {
	// fmt.Println("======Start UNMOUNT======")
	// fmt.Println("Id:", id)
	// fmt.Println("Disk File Name:", diskFileName)

	var partitionStatusStr string = ""
	var partitionIndex int = -1
	for i := 0; i < 4; i++ {
		partitionStatus := mbr.Partitions[i].Status[0]
		partitionStatusStr = strconv.Itoa(int(partitionStatus))
		if strings.Contains(string(mbr.Partitions[i].Id[:]), id) {
			partitionIndex = i
			break
		}
	}
	if partitionStatusStr == "0" {
		return -1, fmt.Errorf("la partición ya no se encuentra montada")
	}
	if partitionIndex >= 0 {
		//fmt.Println("Partition Index:", partitionIndex)
		var partitionToUnmount Types.Partition = mbr.Partitions[partitionIndex]
		//fmt.Println("Partition to unmount:", partitionToUnmount)
		partitionToUnmount.Status[0] = 0
		partitionToUnmount.Id = [4]byte{0, 0, 0, 0}
		//fmt.Println("Partition unmounted:", partitionToUnmount)

		mbr.Partitions[partitionIndex] = partitionToUnmount

		err := Fdisk.WriteMBR(diskFileName, *mbr)
		if err != nil {
			return -1, err
		}

		//fmt.Println("======End UNMOUNT======")
		return 0, nil
	} else {
		return -1, fmt.Errorf("no se encontró la partición")
	}
}

func ExtractUnmountParams(params []string) (string, error) {
	var id string = ""

	if len(params) == 0 {
		return "", fmt.Errorf("No se encontraron parámetros")
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
		return "", fmt.Errorf("No se encontraron parámetros obligatorios")
	}

	for _, param := range params {
		switch {
		case strings.HasPrefix(param, "-id="):
			id = strings.TrimPrefix(param, "-id=")
			// Validar el id de la partición
			// TODO
		}
	}

	return id, nil
}

func ValidatePartitionId(mbr *Types.MBR, id string) (string, error) {
	var partitionId string = ""
	for i := 0; i < 4; i++ {
		if strings.Contains(string(mbr.Partitions[i].Id[:]), id) {
			partitionId = string(mbr.Partitions[i].Id[:])
			break
		}
	}
	if partitionId == "" {
		return "", fmt.Errorf("No se encontró la partición con el id especificado")
	}
	return partitionId, nil
}

func GetPartitionStart(mbr *Types.MBR, id string) (int32, error) {
	var partitionStart int32 = 0
	for i := 0; i < 4; i++ {
		if strings.Contains(string(mbr.Partitions[i].Id[:]), id) {
			partitionStart = mbr.Partitions[i].Start
			break
		}
	}
	return partitionStart, nil
}
