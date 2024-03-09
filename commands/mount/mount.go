package Mount

import (
	"fmt"
	Types "proyecto1/types"
	Utilities "proyecto1/utils"
	"strconv"
	"strings"
	"unicode"
)

func MountPartition(diskFileName string, driveletter string, name string) (int, error) {
	fmt.Println("======Start MOUNT======")
	fmt.Println("Driveletter:", driveletter)
	fmt.Println("Name:", name)

	// Open bin file
	file, err := Utilities.AbrirArchivo(diskFileName)
	if err != nil {
		return -1, err
	}

	var TempMBR Types.MBR
	// Read object from bin file
	if err := Utilities.ReadObject(file, &TempMBR, 0); err != nil {
		return -1, err
	}

	// Print object
	PrintMBR(TempMBR)

	fmt.Println("-------------")

	var index int = -1
	var count = 0
	// Iterate over the partitions
	for i := 0; i < 4; i++ {
		if TempMBR.Partitions[i].Size != 0 {
			count++
			if strings.Contains(string(TempMBR.Partitions[i].Name[:]), name) {
				index = i
				break
			}
		}
	}

	if index != -1 {
		fmt.Println("Partition found")
		PrintPartition(TempMBR.Partitions[index])
	} else {
		fmt.Println("Partition not found")
		return -1, nil
	}

	// id = DriveLetter + Correlative + 19

	id := strings.ToUpper(driveletter) + strconv.Itoa(count) + "19"

	copy(TempMBR.Partitions[index].Status[:], "1")
	copy(TempMBR.Partitions[index].Id[:], id)

	// Overwrite the MBR
	if err := Utilities.WriteObject(file, TempMBR, 0); err != nil {
		return -1, err
	}

	var TempMBR2 Types.MBR
	// Read object from bin file
	if err := Utilities.ReadObject(file, &TempMBR2, 0); err != nil {
		return -1, err
	}

	// Print object
	PrintMBR(TempMBR2)

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

	for _, param := range params {
		switch {
		case strings.HasPrefix(param, "-driveletter="):
			driveletter = strings.TrimPrefix(param, "-driveletter=")
			// Validar la letra de la partición
			if len(driveletter) != 1 || !unicode.IsLetter(rune(driveletter[0])) {
				return "", "", fmt.Errorf("La letra de la partición debe ser un único carácter alfabérico")
			}
		case strings.HasPrefix(param, "-name="):
			name = strings.TrimPrefix(param, "-name=")
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
