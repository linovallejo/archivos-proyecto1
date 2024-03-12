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
	fmt.Println("======Start MOUNT======")
	fmt.Println("Driveletter:", driveletter)
	fmt.Println("Name:", name)

	// Open bin file
	file, err := Utilities.AbrirArchivo(diskFileName)
	if err != nil {
		return -1, err
	}

	fmt.Println("-------------")

	var index int = -1
	var count = 0
	// Iterate over the partitions
	for i := 0; i < 4; i++ {
		if mbr.Partitions[i].Size != 0 {
			count++
			if strings.Contains(string(mbr.Partitions[i].Name[:]), name) {
				index = i
				break
			}
		}
	}

	if index != -1 {
		fmt.Println("Partition found")
	} else {
		defer file.Close()
		fmt.Println("Partition not found")
		return -1, nil
	}

	if strings.Contains(string(mbr.Partitions[index].Status[:]), "1") {
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

	fmt.Println("======End MOUNT======")

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

func UnmountPartition(id string) {
	fmt.Println("======Start UNMOUNT======")
	fmt.Println("Id:", id)
	fmt.Println("======End UNMOUNT======")
}

func ExtractUnmountParams(params []string) (string, error) {
	var id string = ""

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
