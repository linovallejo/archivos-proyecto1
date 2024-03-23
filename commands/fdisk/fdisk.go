package Fdisk

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	Types "proyecto1/types"
	Utils "proyecto1/utils"
	"strconv"
	"strings"
	"unicode"
)

type Space struct {
	Start int64 // Inicio del espacio libre
	Size  int64 // Tamaño del espacio libre
}

func ExtractFdiskParams(params []string) (int64, string, string, string, string, string, string, int64, error) {
	var size int64
	var driveletter, name, unit, parttype, fit, delete, add string
	var addValue int64 = 0
	var err error

	if len(params) == 0 {
		return 0, "", "", "", "", "", "", 0, fmt.Errorf("No se encontraron parámetros")
	}

	var parametrosObligatoriosOk bool = false
	sizeOk := false
	driveletterOk := false
	nameOk := false
	addOk := false
	deleteOk := false
	for _, param1 := range params {
		if strings.HasPrefix(param1, "-size=") {
			sizeOk = true
		} else if strings.HasPrefix(param1, "-driveletter=") {
			driveletterOk = true
		} else if strings.HasPrefix(param1, "-name=") {
			nameOk = true
		} else if strings.HasPrefix(param1, "-add=") {
			addOk = true
		} else if strings.HasPrefix(param1, "-delete=") {
			deleteOk = true
		}
	}

	if !sizeOk && (addOk || deleteOk) {
		sizeOk = true
	}

	if sizeOk && driveletterOk && nameOk {
		parametrosObligatoriosOk = true
	}

	if !parametrosObligatoriosOk {
		return 0, "", "", "", "", "", "", 0, fmt.Errorf("No se encontraron parámetros obligatorios")
	}

	for _, param := range params {
		switch {
		case strings.HasPrefix(param, "-size="):
			sizeStr := strings.TrimPrefix(param, "-size=")
			if strings.TrimSpace(sizeStr) == "" {
				return 0, "", "", "", "", "", "", 0, fmt.Errorf("Parametro tamaño es obligatorio")
			}

			var err error
			size, err = strconv.ParseInt(sizeStr, 10, 64)
			if err != nil || size <= 0 {
				return 0, "", "", "", "", "", "", 0, fmt.Errorf("Parametro tamaño invalido")
			}
		case strings.HasPrefix(param, "-driveletter="):
			driveletter = strings.TrimPrefix(param, "-driveletter=")
			if strings.TrimSpace(driveletter) == "" {
				return 0, "", "", "", "", "", "", 0, fmt.Errorf("Parametro driveletter es obligatorio")
			}
			// Validar la letra de la partición
			if len(driveletter) != 1 || !unicode.IsLetter(rune(driveletter[0])) {
				return 0, "", "", "", "", "", "", 0, fmt.Errorf("La letra de la partición debe ser un único carácter alfabérico")
			}
		case strings.HasPrefix(param, "-name="):
			name = strings.TrimPrefix(param, "-name=")
			if strings.TrimSpace(name) == "" {
				return 0, "", "", "", "", "", "", 0, fmt.Errorf("Parametro name es obligatorio")
			}
			// Validar el nombre de la partición
			if len(name) > 16 {
				return 0, "", "", "", "", "", "", 0, fmt.Errorf("El nombre de la partición no puede exceder los 16 caracteres")
			}
		case strings.HasPrefix(param, "-unit="):
			unit = strings.TrimPrefix(param, "-unit=")
			unit = strings.ToLower(unit)
			// Validar la unidad de la partición
			if unit != "b" && unit != "k" && unit != "m" {
				return 0, "", "", "", "", "", "", 0, fmt.Errorf("Parametro unidad invalido")
			}
		case strings.HasPrefix(param, "-type="):
			parttype = strings.TrimPrefix(param, "-type=")
			// Validar el tipo de la partición
			if parttype != "P" && parttype != "E" && parttype != "L" {
				return 0, "", "", "", "", "", "", 0, fmt.Errorf("El tipo de la partición debe ser 'P', 'E' o 'L'")
			}
		case strings.HasPrefix(param, "-fit="):
			fit = strings.TrimPrefix(param, "-fit=")
			// Validar el ajuste de la partición
			fit = strings.ToLower(fit)
			if fit != "bf" && fit != "ff" && fit != "wf" {
				return 0, "", "", "", "", "", "", 0, fmt.Errorf("Parametro fit invalido")
			}
		case strings.HasPrefix(param, "-delete="):
			delete = strings.TrimPrefix(param, "-delete=")
			// Validar el parametro delete
			if delete != "fast" && delete != "full" {
				return 0, "", "", "", "", "", "", 0, fmt.Errorf("Parametro delete invalido")
			}
		case strings.HasPrefix(param, "-add="):
			add = strings.TrimPrefix(param, "-add=")
			// Validar el parametro add
			addValue, err = strconv.ParseInt(add, 10, 64)
			if err != nil {
				fmt.Println("Parametro add invalido")
				continue
			}
			// if addValue < 0 {
			// 	fmt.Println("Parametro add es negativo")
			// }
		}
	}

	if (size == 0 && !addOk && !deleteOk) || driveletter == "" || name == "" {
		return 0, "", "", "", "", "", "", 0, fmt.Errorf("Parametro obligatorio faltante")
	}

	// Unidad por defecto es Kilobytes
	if unit == "" {
		unit = "K"
	}

	return size, driveletter, name, unit, parttype, fit, delete, addValue, nil
}

func ReadMBR(filename string) (*Types.MBR, error) {
	var mbr Types.MBR

	file, err := os.Open(filename)
	if err != nil {
		return &mbr, err
	}
	defer file.Close()

	// Omitir el byte nulo inicial
	_, err = file.Seek(1, io.SeekStart)
	if err != nil {
		return &mbr, err
	}

	err = binary.Read(file, binary.LittleEndian, &mbr)
	return &mbr, err
}

func WriteMBR(filename string, mbr Types.MBR) error {
	file, err := os.OpenFile(filename, os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	// Omitir el byte nulo inicial
	_, err = file.Seek(1, io.SeekStart)
	if err != nil {
		return err
	}

	err = binary.Write(file, binary.LittleEndian, &mbr)
	return err
}

func calculateTotalUsedSpace(mbr Types.MBR) (int32, error) {
	var totalUsedSpace int32 = 0
	for _, partition := range mbr.Partitions {
		if partition.Status != [1]byte{0} {
			if partition.Size < 0 {
				return 0, fmt.Errorf("tamano de partición invalido: %d", partition.Size)
			}
			newTotal := totalUsedSpace + partition.Size
			// Check for overflow
			if newTotal < totalUsedSpace {
				return 0, fmt.Errorf("integer overflow detected when calculating total used space")
			}
			totalUsedSpace = newTotal
		}
	}
	return totalUsedSpace, nil
}

func createPartition(mbr *Types.MBR, start int64, size int32, unit string, typePart, fit, name string, diskFileName string) error {
	//fmt.Println("createPartition: ", start, size, unit, typePart, fit, name, diskFileName)

	var sizeInBytes int32 = 0
	unit = strings.ToLower(unit)

	switch unit {
	case "b":
		sizeInBytes = size
	case "k":
		sizeInBytes = size * 1024
	case "m":
		sizeInBytes = size * 1024 * 1024
	default:
		return fmt.Errorf("Unidad invalida")
	}

	if typePart == "P" || typePart == "E" {
		var count = 0
		var gap = int32(0)
		// Iterate over the partitions
		for i := 0; i < 4; i++ {
			if mbr.Partitions[i].Size != 0 {
				count++
				gap = mbr.Partitions[i].Start + mbr.Partitions[i].Size
			}
		}

		for i := 0; i < 4; i++ {
			if mbr.Partitions[i].Size == 0 {
				mbr.Partitions[i].Size = sizeInBytes

				if count == 0 {
					mbr.Partitions[i].Start = int32(binary.Size(mbr))
				} else {
					mbr.Partitions[i].Start = gap
				}

				copy(mbr.Partitions[i].Name[:], name)
				copy(mbr.Partitions[i].Fit[:], fit)
				//copy(mbr.Partitions[i].Status[:], "0")
				mbr.Partitions[i].Status[0] = 0
				copy(mbr.Partitions[i].Type[:], typePart)
				//mbr.Partitions[i].Type[0] = typePart[0]
				mbr.Partitions[i].Correlative = int32(count + 1)
				break
			}
		}
	}

	if typePart == "E" {
		// Instead of directly accessing logicalPartitions, use AddOrUpdateLogicalPartition
		AddOrUpdateLogicalPartition(diskFileName, &LogicalPartitionInfo{
			ExtendedStart: int32(start), // Assuming start can be safely converted to int32
			FirstEBR:      nil,          // Initially, there's no EBR
		})

		_, exists1 := GetLogicalPartition(diskFileName)
		// if exists1 {
		// 	fmt.Println("particion lógica creada correctamente")
		// } else {
		// 	fmt.Println("crear una partición lógica requiere una partición extendida en el disco")
		// }
		if !exists1 {
			return fmt.Errorf("crear una partición lógica requiere una partición extendida en el disco")
		}
	}

	// Logical Partition Check
	if typePart == "L" {
		//fmt.Println("logical partition to be created")
		// Use GetLogicalPartition to check if the extended partition exists
		info, exists := GetLogicalPartition(diskFileName)
		if !exists {
			return fmt.Errorf("crear una partición lógica requiere una partición extendida en el disco")
		}

		var extendedPartition *Types.Partition
		var err error
		extendedPartition, err = GetExtendedPartition(diskFileName)
		if err != nil {
			return fmt.Errorf("crear una partición lógica requiere una partición extendida en el MBR")
		}
		var sizeInBytesExtended int32 = int32(extendedPartition.Size)
		// fmt.Println("sizeInBytesExtended:", sizeInBytesExtended)
		// fmt.Println("sizeInBytesLogical:", sizeInBytes)

		if info.FirstEBR == nil {
			if sizeInBytesExtended < sizeInBytes {
				return fmt.Errorf("logical partition size exceeds extended partition size")
			}
		} else {
			var totalLogicalPartitionsSize int32 = 0
			currentEBR := info.FirstEBR
			for currentEBR != nil {
				totalLogicalPartitionsSize += currentEBR.PartSize
				currentEBR = currentEBR.PartNext
			}

			fmt.Println("totalLogicalPartitionsSize:", totalLogicalPartitionsSize)
			totalLogicalPartitionsSize += sizeInBytes

			if totalLogicalPartitionsSize > sizeInBytesExtended {
				return fmt.Errorf("combined size of logical partitions exceeds extended partition size")
			}
		}

		// Assuming the logic to find space in extended and create a new EBR is encapsulated elsewhere...
		newEBR := &Types.EBR{
			// ... Fill in the EBR fields (start, size, etc.)
			PartStart: 0,
			PartSize:  sizeInBytes,
		}
		copy(newEBR.PartFit[:], fit)
		copy(newEBR.PartName[:], name)
		newEBR.PartMount = [1]byte{0}

		// Assuming there's logic to correctly set FirstEBR or add the new EBR to the existing chain,
		// which might involve more functions in the `partition` package for manipulating the EBR chain.
		if info.FirstEBR == nil {
			// Direct modification is no longer appropriate; you might need a function to update this.
			SetFirstEBR(diskFileName, newEBR)
			//fmt.Println("First EBR set for disk", diskFileName)
		} else {
			// Similarly, logic to add the EBR to the chain would be encapsulated in a function
			AddEBRToChain(diskFileName, newEBR)
			//fmt.Println("EBR added to chain for disk", diskFileName)
		}

	}

	WriteMBR(diskFileName, *mbr)

	return nil
}

func GetExtendedPartition(diskFileName string) (*Types.Partition, error) {
	var partition Types.Partition
	mbr, err := ReadMBR(diskFileName)
	if err != nil {
		return nil, err
	}

	for _, p := range mbr.Partitions {
		if p.Type == [1]byte{'E'} {
			partition = p
			return &partition, nil
		}
	}

	return nil, fmt.Errorf("No se encontró la partición extendida")
}

func AdjustAndCreatePartition(mbr *Types.MBR, size int32, unit, typePart, fit, name string, diskFileName string) error {
	spaces := calculateAvailableSpaces(mbr)
	var selectedSpace *Space

	fit = strings.ToLower(fit)
	switch fit {
	case "ff":
		selectedSpace = findFirstFit(spaces, size)
	case "bf":
		selectedSpace = findBestFit(spaces, size)
	case "wf":
		selectedSpace = findWorstFit(spaces, size)
	default: //WF por defecto
		selectedSpace = findWorstFit(spaces, size)
	}

	if selectedSpace == nil {
		return fmt.Errorf("No se encontró espacio adecuado para la partición")
	}

	// Aquí, usarías selectedSpace.Start como la posición de inicio de tu nueva partición
	// y procederías a crear la partición con el tamaño especificado.
	err := createPartition(mbr, selectedSpace.Start, size, unit, typePart, fit, name, diskFileName)
	if err != nil {
		return err
	}

	return nil
}

func calculateAvailableSpaces(mbr *Types.MBR) []Space {
	var spaces []Space
	var lastPosition int32 = 1 // Asume que el disco comienza en la posición 1

	for _, partition := range mbr.Partitions {
		if partition.Status[0] != 0 { // Asume que Status != 0 significa partición ocupada
			if lastPosition < partition.Start {
				// Hay espacio libre entre la última posición y el inicio de la partición actual
				spaces = append(spaces, Space{Start: int64(lastPosition), Size: int64(partition.Start) - int64(lastPosition)})
			}
			lastPosition = partition.Start + partition.Size
		}
	}

	// Considera el espacio hasta el final del disco
	if lastPosition < mbr.MbrTamano {
		spaces = append(spaces, Space{Start: int64(lastPosition), Size: int64(mbr.MbrTamano) - int64(lastPosition)})
	}

	return spaces
}

func findFirstFit(spaces []Space, size int32) *Space {
	for _, space := range spaces {
		if space.Size >= int64(size) {
			return &space
		}
	}
	return nil
}

func findBestFit(spaces []Space, size int32) *Space {
	var bestSpace *Space
	for _, space := range spaces {
		if space.Size >= int64(size) && (bestSpace == nil || space.Size < bestSpace.Size) {
			bestSpace = &space
		}
	}
	return bestSpace
}

func findWorstFit(spaces []Space, size int32) *Space {
	var worstSpace *Space
	for _, space := range spaces {
		if space.Size >= int64(size) && (worstSpace == nil || space.Size > worstSpace.Size) {
			worstSpace = &space
		}
	}
	return worstSpace
}
func ValidatePartitionTypeCreation(mbr *Types.MBR, partType string) error {
	var countP, countE int = 1, 1

	for _, partition := range mbr.Partitions {
		// fmt.Println("Particion:", partition)
		// fmt.Println("Particion:", Utils.CleanPartitionName(partition.Name[:]))
		// fmt.Println("Type:", string(partition.Type[0]))
		// fmt.Printf("Debug Type: %s\n", string(partition.Type[:]))
		switch string(partition.Type[:]) {
		case "P": // Asume que 'P' representa una partición Primaria
			countP++
		case "E": // Asume que 'E' representa una partición Extendida
			countE++
		}
	}

	// fmt.Println("countP:", countP)
	// fmt.Println("countE:", countE)

	if partType == "E" && countE > 1 {
		return fmt.Errorf("Ya existe una partición extendida en el disco")
	}

	if (partType == "P" || partType == "E") && (countP+countE) > 5 {
		return fmt.Errorf("No se pueden crear más particiones primarias o extendidas (límite de 4)")
	}

	// Para L, asegúrate de que ya existe una partición Extendida
	if partType == "L" && countE == 0 {
		return fmt.Errorf("Debe existir una partición extendida para crear una partición lógica")
	}

	return nil
}

func ValidatePartitionsSizeAgainstDiskSize(mbr *Types.MBR, newPartitionSize int64) error {

	var totalSizePartitions int64 = 0
	for _, partition := range mbr.Partitions {
		if string(partition.Type[:]) != "L" {
			//fmt.Println("partition.Type:", string(partition.Type[:]))
			//fmt.Println("partition.Size:", partition.Size)

			totalSizePartitions += int64(partition.Size)
			//fmt.Println("totalSizePartitions:", totalSizePartitions)
		}
	}

	if (totalSizePartitions + newPartitionSize) > int64(mbr.MbrTamano) {
		return fmt.Errorf("El tamaño de las particiones supera el tamaño del disco")
	}

	return nil
}

func ValidateFileName(path string, filename string) (string, error) {
	fullPath := filepath.Join(path, filename)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		// El archivo no existe
		return "", fmt.Errorf("el archivo %s no existe", fullPath)
	}
	// El archivo existe
	return fullPath, nil
}

func GenerateDotCodeMbr(mbr *Types.MBR, diskFileName string) string {
	var builder strings.Builder

	mbrFechaCreacion := Utils.CleanPartitionName(mbr.MbrFechaCreacion[:])

	builder.WriteString("digraph G {\n")
	builder.WriteString("    node [shape=none];\n")
	builder.WriteString("    rankdir=\"LR\";\n")

	builder.WriteString("    struct1 [label=<<TABLE BORDER=\"0\" CELLBORDER=\"1\" CELLSPACING=\"0\">\n")

	// Nodo MBR
	builder.WriteString("    <TR>")
	builder.WriteString("    <TD BGCOLOR=\"#4A235A\" COLSPAN=\"2\"><FONT COLOR=\"white\">REPORTE DE MBR</FONT></TD>")
	builder.WriteString("    </TR>\n")
	builder.WriteString("    <TR>")
	builder.WriteString("    <TD BGCOLOR=\"#FFFFFF\">mbr_tamano:</TD>")
	builder.WriteString("    <TD BGCOLOR=\"#FFFFFF\">" + strconv.Itoa(int(mbr.MbrTamano)) + "</TD>")
	builder.WriteString("    </TR>\n")
	builder.WriteString("    <TR>")
	builder.WriteString("    <TD BGCOLOR=\"#E8DAEF\">mbr_fecha_creacion:</TD>")
	builder.WriteString("    <TD BGCOLOR=\"#E8DAEF\">" + mbrFechaCreacion + "</TD>")
	builder.WriteString("    </TR>\n")
	builder.WriteString("    <TR>")
	builder.WriteString("    <TD BGCOLOR=\"#FFFFFF\">mbr_disk_signature:</TD>")
	builder.WriteString("    <TD BGCOLOR=\"#FFFFFF\">" + strconv.Itoa(int(mbr.MbrDiskSignature)) + "</TD>")
	builder.WriteString("    </TR>\n")

	// Nodos de particiones
	for _, partition := range mbr.Partitions {
		partitionName := Utils.CleanPartitionName(partition.Name[:])
		partitionStatus := partition.Status[0]
		partitionStatusStr := strconv.Itoa(int(partitionStatus))

		if string(partition.Type[:]) == "P" || string(partition.Type[:]) == "E" {

			builder.WriteString("    <TR>")
			builder.WriteString("    <TD BGCOLOR=\"#4A235A\" COLSPAN=\"2\"><FONT COLOR=\"white\">Particion</FONT></TD>")
			builder.WriteString("    </TR>\n")
			builder.WriteString("    <TR>")
			builder.WriteString("    <TD BGCOLOR=\"#FFFFFF\">part_status:</TD>")
			builder.WriteString("    <TD BGCOLOR=\"#FFFFFF\">" + string(partitionStatusStr) + "</TD>")
			builder.WriteString("    </TR>\n")
			builder.WriteString("    <TR>")
			builder.WriteString("    <TD BGCOLOR=\"#E8DAEF\">part_type:</TD>")
			builder.WriteString("    <TD BGCOLOR=\"#E8DAEF\">" + string(partition.Type[:]) + "</TD>")
			builder.WriteString("    </TR>\n")
			builder.WriteString("    <TR>")
			builder.WriteString("    <TD BGCOLOR=\"#FFFFFF\">part_fit:</TD>")
			builder.WriteString("    <TD BGCOLOR=\"#FFFFFF\">" + string(partition.Fit[:]) + "</TD>")
			builder.WriteString("    </TR>\n")
			builder.WriteString("    <TR>")
			builder.WriteString("    <TD BGCOLOR=\"#E8DAEF\">part_start:</TD>")
			builder.WriteString("    <TD BGCOLOR=\"#E8DAEF\">" + strconv.Itoa(int(partition.Start)) + "</TD>")
			builder.WriteString("    </TR>\n")
			builder.WriteString("    <TR>")
			builder.WriteString("    <TD BGCOLOR=\"#E8DAEF\">part_size:</TD>")
			builder.WriteString("    <TD BGCOLOR=\"#E8DAEF\">" + strconv.Itoa(int(partition.Size)) + "</TD>")
			builder.WriteString("    </TR>\n")
			builder.WriteString("    <TR>")
			builder.WriteString("    <TD BGCOLOR=\"#E8DAEF\">part_name:</TD>")
			builder.WriteString("    <TD BGCOLOR=\"#E8DAEF\">" + partitionName + "</TD>")
			builder.WriteString("    </TR>\n")
		}
	}

	logicalPartitions, _ := GetLogicalPartition(diskFileName)
	if logicalPartitions != nil {

		currentEBR := logicalPartitions.FirstEBR
		partitionNumber := 1

		for currentEBR != nil {

			builder.WriteString("    <TR>")
			builder.WriteString("    <TD BGCOLOR=\"#4A235A\" COLSPAN=\"2\"><FONT COLOR=\"white\">Particion Logica</FONT></TD>")
			builder.WriteString("    </TR>\n")

			partitionName := Utils.CleanPartitionName(currentEBR.PartName[:])
			partNextStr := "-1"
			if currentEBR.PartNext != nil {
				partNextStr = fmt.Sprintf("%p", currentEBR.PartNext) // %p formats as a pointer (base 16 notation)
			}

			builder.WriteString("    <TR>")
			builder.WriteString("    <TD BGCOLOR=\"#FFFFFF\">part_status:</TD>")
			builder.WriteString("    <TD BGCOLOR=\"#FFFFFF\">0</TD>")
			builder.WriteString("    </TR>\n")
			builder.WriteString("    <TR>")
			builder.WriteString("    <TD BGCOLOR=\"#FFFFFF\">part_next:</TD>")
			builder.WriteString("    <TD BGCOLOR=\"#FFFFFF\">" + partNextStr + "</TD>")
			builder.WriteString("    </TR>\n")
			builder.WriteString("    <TR>")
			builder.WriteString("    <TD BGCOLOR=\"#FFFFFF\">part_type:</TD>")
			builder.WriteString("    <TD BGCOLOR=\"#FFFFFF\">L</TD>")
			builder.WriteString("    </TR>\n")
			builder.WriteString("    <TR>")
			builder.WriteString("    <TD BGCOLOR=\"#FFFFFF\">part_fit:</TD>")
			builder.WriteString("    <TD BGCOLOR=\"#FFFFFF\">" + string(currentEBR.PartFit[:]) + "</TD>")
			builder.WriteString("    </TR>\n")
			builder.WriteString("    <TR>")
			builder.WriteString("    <TD BGCOLOR=\"#E8DAEF\">part_start:</TD>")
			builder.WriteString("    <TD BGCOLOR=\"#E8DAEF\">" + strconv.Itoa(int(currentEBR.PartStart)) + "</TD>")
			builder.WriteString("    </TR>\n")
			builder.WriteString("    <TR>")
			builder.WriteString("    <TD BGCOLOR=\"#E8DAEF\">part_size:</TD>")
			builder.WriteString("    <TD BGCOLOR=\"#E8DAEF\">" + strconv.Itoa(int(currentEBR.PartSize)) + "</TD>")
			builder.WriteString("    </TR>\n")
			builder.WriteString("    <TR>")
			builder.WriteString("    <TD BGCOLOR=\"#E8DAEF\">part_name:</TD>")
			builder.WriteString("    <TD BGCOLOR=\"#E8DAEF\">" + partitionName + "</TD>")
			builder.WriteString("    </TR>\n")

			// Move to the next EBR in the chain
			currentEBR = currentEBR.PartNext
			partitionNumber++
		}
	}

	builder.WriteString("    </TABLE>>];\n")

	//EBR Section

	if logicalPartitions != nil {
		builder.WriteString("    struct2 [label=<<TABLE BORDER=\"0\" CELLBORDER=\"1\" CELLSPACING=\"0\">\n")

		// Nodo MBR
		builder.WriteString("    <TR>")
		builder.WriteString("    <TD BGCOLOR=\"#4A235A\" COLSPAN=\"2\"><FONT COLOR=\"white\">EBR</FONT></TD>")
		builder.WriteString("    </TR>\n")

		currentEBR := logicalPartitions.FirstEBR
		partitionNumber := 1

		for currentEBR != nil {

			builder.WriteString("    <TR>")
			builder.WriteString("    <TD BGCOLOR=\"#4A235A\" COLSPAN=\"2\"><FONT COLOR=\"white\">Particion</FONT></TD>")
			builder.WriteString("    </TR>\n")

			partitionName := Utils.CleanPartitionName(currentEBR.PartName[:])
			partNextStr := "-1"
			if currentEBR.PartNext != nil {
				partNextStr = fmt.Sprintf("%p", currentEBR.PartNext) // %p formats as a pointer (base 16 notation)
			}

			builder.WriteString("    <TR>")
			builder.WriteString("    <TD BGCOLOR=\"#FFFFFF\">part_status:</TD>")
			builder.WriteString("    <TD BGCOLOR=\"#FFFFFF\">0</TD>")
			builder.WriteString("    </TR>\n")
			builder.WriteString("    <TR>")
			builder.WriteString("    <TD BGCOLOR=\"#FFFFFF\">part_next:</TD>")
			builder.WriteString("    <TD BGCOLOR=\"#FFFFFF\">" + partNextStr + "</TD>")
			builder.WriteString("    </TR>\n")
			builder.WriteString("    <TR>")
			builder.WriteString("    <TD BGCOLOR=\"#FFFFFF\">part_type:</TD>")
			builder.WriteString("    <TD BGCOLOR=\"#FFFFFF\">L</TD>")
			builder.WriteString("    </TR>\n")
			builder.WriteString("    <TR>")
			builder.WriteString("    <TD BGCOLOR=\"#FFFFFF\">part_fit:</TD>")
			builder.WriteString("    <TD BGCOLOR=\"#FFFFFF\">" + string(currentEBR.PartFit[:]) + "</TD>")
			builder.WriteString("    </TR>\n")
			builder.WriteString("    <TR>")
			builder.WriteString("    <TD BGCOLOR=\"#E8DAEF\">part_start:</TD>")
			builder.WriteString("    <TD BGCOLOR=\"#E8DAEF\">" + strconv.Itoa(int(currentEBR.PartStart)) + "</TD>")
			builder.WriteString("    </TR>\n")
			builder.WriteString("    <TR>")
			builder.WriteString("    <TD BGCOLOR=\"#E8DAEF\">part_size:</TD>")
			builder.WriteString("    <TD BGCOLOR=\"#E8DAEF\">" + strconv.Itoa(int(currentEBR.PartSize)) + "</TD>")
			builder.WriteString("    </TR>\n")
			builder.WriteString("    <TR>")
			builder.WriteString("    <TD BGCOLOR=\"#E8DAEF\">part_name:</TD>")
			builder.WriteString("    <TD BGCOLOR=\"#E8DAEF\">" + partitionName + "</TD>")
			builder.WriteString("    </TR>\n")

			// Move to the next EBR in the chain
			currentEBR = currentEBR.PartNext
			partitionNumber++
		}

		builder.WriteString("    </TABLE>>];\n")

		builder.WriteString("    struct1 -> struct2 [style=invis];\n")

	}

	builder.WriteString("}\n")

	return builder.String()
}

func GenerateDotCodeDisk(mbr *Types.MBR, diskFileName string) (string, error) {
	var dot bytes.Buffer

	dot.WriteString("digraph G {\n")
	dot.WriteString("node [shape=plaintext]\n")
	dot.WriteString("struct1 [label=<\n")
	dot.WriteString("<table border=\"0\" cellborder=\"1\" cellspacing=\"0\" cellpadding=\"4\">\n")

	// var espacioLibreDisco = Mkdisk.CalcularEspacioLibreDisco(mbr)

	//fmt.Println("before GetLogicalPartition")
	logicalPartitions, _ := GetLogicalPartition(diskFileName)
	// fmt.Println("Espacio libre:", espacioLibreDisco)
	//fmt.Println("after GetLogicalPartition")

	dot.WriteString("<tr>\n")
	dot.WriteString("<td rowspan=\"2\">MBR</td>\n") // MBR siempre está presente

	var espacioLibre int32 = 0
	var espacioLibreTemporal int32 = 0
	var porcentajeLibre int32 = 0
	var espacioLibreFinal int32 = 0
	var porcentajeLibreFinal int32 = 0
	var espacioTotalDisco int32 = 0
	espacioTotalDisco = int32(mbr.MbrTamano)
	var porcentajeParticion int32 = 0
	var espacioOcupado int32 = 0
	var particionesLibres int32 = 0
	var espacioOcupadoExtendida int32 = 0
	var dotLogicalPartitions bytes.Buffer

	fmt.Println("Espacio total:", espacioTotalDisco)
	fmt.Println("Espacio libre antes del loop:", espacioLibre)
	for _, partition := range mbr.Partitions {
		partitionName := Utils.CleanPartitionName(partition.Name[:])
		partitionStatus := partition.Status[0]
		fmt.Println("Particion:", partitionName, "Status:", partitionStatus, "Size:", partition.Size, "Espacio libre:", espacioLibre)

		if partitionStatus != 1 && partition.Type[0] == 'P' {
			if partition.Size > 0 {
				//espacioLibre += espacioLibre + partition.Size
				fmt.Println("Espacio libre antes de sumar:", espacioLibre)
				espacioLibre += partition.Size
				fmt.Println("Espacio libre despues de sumar:", espacioLibre)
			} else {
				particionesLibres++
			}
		} else {
			if espacioLibre > 0 {
				porcentajeLibre = int32((100 * espacioLibre) / espacioTotalDisco)
				if porcentajeLibre > 0 {
					dot.WriteString(fmt.Sprintf("<td rowspan=\"2\">%s<br/><FONT POINT-SIZE='6'>%d %% del disco</FONT></td>\n", "Libre1", porcentajeLibre))
					espacioLibreTemporal += espacioLibre
					espacioLibre = 0
				}
			}
			espacioOcupado += partition.Size
			porcentajeParticion = int32((100 * partition.Size) / espacioTotalDisco)
			if (partition.Type[0] == 'P') && (porcentajeParticion > 0) {
				dot.WriteString(fmt.Sprintf("<td rowspan=\"2\">%s<br/><FONT POINT-SIZE='6'>%d %% del disco</FONT></td>\n", partitionName, porcentajeParticion))
				porcentajeParticion = 0
			} else if (partition.Type[0] == 'E') && (porcentajeParticion > 0) {
				if logicalPartitions != nil && logicalPartitions.FirstEBR != nil {
					currentEBR := logicalPartitions.FirstEBR
					partitionNumber := 0
					var espacioTotalLogicalPartitions int32 = 0
					for currentEBR != nil {
						// Print details of the logical partition
						fmt.Printf("Partition %d:\n", partitionNumber)
						fmt.Printf("  Name: %s\n", currentEBR.PartName[:])
						fmt.Printf("  Start: %d\n", currentEBR.PartStart)
						fmt.Printf("  Size: %d\n", currentEBR.PartSize)

						dotLogicalPartitions.WriteString("<td>EBR</td>\n")
						partitionName = Utils.CleanPartitionName(currentEBR.PartName[:])
						espacioTotalLogicalPartitions += int32(currentEBR.PartSize)
						porcentajeParticion = int32((100 * currentEBR.PartSize) / espacioTotalDisco)
						dotLogicalPartitions.WriteString(fmt.Sprintf("<td>%s<br/><FONT POINT-SIZE='6'>%d %% del disco</FONT></td>\n", partitionName, porcentajeParticion))

						// Move to the next EBR in the chain
						currentEBR = currentEBR.PartNext
						partitionNumber++
					}
					fmt.Println("Espacio total:", espacioTotalLogicalPartitions)
					espacioOcupadoExtendida = partition.Size
					fmt.Println("Extendida:", espacioOcupadoExtendida)
					var colSpan int = 0
					colSpan = partitionNumber * 2
					if (espacioOcupadoExtendida - espacioTotalLogicalPartitions) > 0 {
						colSpan++
						porcentajeLibre = int32((100 * (espacioOcupadoExtendida - espacioTotalLogicalPartitions)) / espacioTotalDisco)
						dotLogicalPartitions.WriteString(fmt.Sprintf("<td>%s<br/><FONT POINT-SIZE='6'>%d %% del disco</FONT></td>\n", "Libre", porcentajeLibre))
					}
					dot.WriteString("<td colspan=\"" + strconv.Itoa(colSpan) + "\">Extendida</td>\n")

					//dot.WriteString(fmt.Sprintf("<td>%s<br/><FONT POINT-SIZE='6'>%d %% del disco</FONT></td>\n", partitionName, porcentajeParticion))
				}

				porcentajeParticion = 0
			}
		}
	}

	fmt.Println("Espacio total disco:", espacioTotalDisco)
	fmt.Println("Espacio ocupado primarias:", espacioOcupado)
	fmt.Println("Espacio ocupado extendida:", espacioOcupadoExtendida)
	fmt.Println("Espacio libre:", espacioLibre)
	Utils.LineaDoble(60)

	espacioLibreFinal = espacioTotalDisco - espacioOcupado - espacioLibreTemporal
	fmt.Printf("Espacio libre final: %d\n", espacioLibreFinal)
	if particionesLibres > 0 || espacioLibreFinal > 0 {
		//espacioLibreFinal = espacioLibreFinal + espacioLibre
		//fmt.Println("Espacio libre:", espacioLibreFinal)

		fmt.Println("Particiones libres:", particionesLibres)
		fmt.Println("Queda Espacio libre:", espacioLibreFinal)
		//espacioLibre = espacioTotalDisco - espacioOcupado
		porcentajeLibreFinal = int32((100 * espacioLibreFinal) / espacioTotalDisco)
		fmt.Printf("Porcentaje Libre: %d\n", porcentajeLibreFinal)
		dot.WriteString(fmt.Sprintf("<td rowspan=\"2\">%s<br/><FONT POINT-SIZE='6'>%d %% del disco</FONT></td>\n", "Libre2", porcentajeLibreFinal))
		fmt.Println("dot.WriteString")
	}

	if dotLogicalPartitions.String() != "" {
		dot.WriteString("</tr>\n")
		dot.WriteString("<tr>\n")
		dot.WriteString(dotLogicalPartitions.String())
	}

	dot.WriteString("</tr>\n")

	dot.WriteString("</table>\n")
	dot.WriteString(">];\n")
	dot.WriteString("}\n")
	//fmt.Printf("dot.String(): %s\n", dot.String())

	return dot.String(), nil
}

func ValidatePartitionName(mbr *Types.MBR, name string, delete string, addValue int64) error {
	partitionExists := false

	// fmt.Println("Validando nombre de la partición:", name)
	// fmt.Println("Delete:", delete)

	var partitionName string = ""
	for _, partition := range mbr.Partitions {
		partitionName = Utils.CleanPartitionName(partition.Name[:])
		//fmt.Println("Particion:", string(partition.Name[:]))
		if strings.TrimSpace(partitionName) == strings.TrimSpace(name) {
			//fmt.Println("Particion encontrada:", string(partition.Name[:]))
			partitionExists = true
			break
		}
	}

	if delete == "full" || addValue < 0 || addValue > 0 {
		if !partitionExists {
			return fmt.Errorf("la partición a eliminar %s no existe", name)
		}
	} else {
		if partitionExists {
			return fmt.Errorf("el nombre de la partición %s ya está en uso", name)
		}
	}

	return nil
}

func CalculateSize(size int64, unit string) (int64, error) {
	unit = strings.ToLower(unit)
	switch unit {
	case "b":
		return size, nil
	case "k":
		return size * 1024, nil
	case "m":
		return size * 1024 * 1024, nil
	default:
		return 0, fmt.Errorf("unidad invalida")
	}
}

func DeletePartition(mbr *Types.MBR, filename string, name string) error {
	// Verificar la existencia de la partición
	partitionIndex := -1
	var partitionName string = ""
	for i, partition := range mbr.Partitions {
		partitionName = Utils.CleanPartitionName(partition.Name[:])

		if strings.TrimSpace(partitionName) == strings.TrimSpace(name) {
			partitionIndex = i
			break
		}
	}
	if partitionIndex == -1 {
		return fmt.Errorf("la partición '%s' no existe", name)
	}

	// Confirmar la eliminación de la partición
	fmt.Printf("¿Estás seguro de que quieres eliminar la partición '%s'? [s/N]: ", name)
	var response string
	fmt.Scanln(&response)
	if response != "s" && response != "S" {
		return fmt.Errorf("eliminación cancelada por el usuario")
	}

	//fmt.Println("getPartitionDetails")
	// Rellenar con `\0`
	start, size, err := getPartitionDetails(mbr, name)
	if err != nil {
		fmt.Println(err)
	}
	// else {
	// 	fmt.Printf("La partición '%s' comienza en %d bytes y tiene un tamaño de %d bytes.\n", name, start, size)
	// }

	// TODO
	// Eliminar la partición
	// Si la partición es extendida, también elimina sus particiones lógicas
	// if string(mbr.Partitions[partitionIndex].Type[:]) == "E" {
	// 	// TODO
	// 	// Implementar lógica para eliminar particiones lógicas si es necesario
	// 	return nil
	// }

	mbr.Partitions[partitionIndex] = Types.Partition{} // Asigna una partición vacía
	WriteMBR(filename, *mbr)

	//fmt.Println("cleanPartitionSpace")
	err = cleanPartitionSpace(filename, int64(start), int64(size))
	if err != nil {
		fmt.Printf("Error al limpiar el espacio de la partición: %v\n", err)
	} else {
		fmt.Println("Espacio de la partición limpiado exitosamente.")
	}

	return nil
}

func cleanPartitionSpace(filename string, startPosition int64, size int64) error {
	// Abrir el archivo del disco
	file, err := os.OpenFile(filename, os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	// Navegar hasta el inicio de la partición
	_, err = file.Seek(startPosition, io.SeekStart)
	if err != nil {
		return err
	}

	// Crear un slice de bytes nulos para escribir
	zeros := make([]byte, size)
	_, err = file.Write(zeros)
	if err != nil {
		return err
	}

	return nil
}

func getPartitionDetails(mbr *Types.MBR, name string) (int32, int32, error) {
	var partitionName string = ""
	for _, partition := range mbr.Partitions {
		partitionName = Utils.CleanPartitionName(partition.Name[:])
		if strings.TrimSpace(partitionName) == strings.TrimSpace(name) {
			return partition.Start, partition.Size, nil
		}
	}
	return 0, 0, fmt.Errorf("la partición '%s' no se encontró", name)
}

func canAdjustPartitionSize(mbr *Types.MBR, partitionIndex int, sizeInBytes int64) (bool, error) {
	currentPartition := mbr.Partitions[partitionIndex]
	// fmt.Println("Partición actual:", currentPartition)
	// fmt.Println("Espacio a ajustar:", sizeInBytes)

	var operationType string = ""
	if sizeInBytes < 0 {
		operationType = "reducir"
	} else {
		operationType = "aumentar"
	}

	if operationType == "aumentar" {
		var totalUsedSpace int64 = 0
		for _, partition := range mbr.Partitions {
			if partition.Status != [1]byte{0} { // Asume Status != 0 como partición activa
				totalUsedSpace += int64(partition.Size) // Convert partition.Size to int64 before adding
			}
		}
		//fmt.Println("Espacio utilizado:", totalUsedSpace)

		// Espacio disponible en el disco
		spaceAvailable := int64(mbr.MbrTamano) - totalUsedSpace
		//fmt.Println("Espacio disponible:", spaceAvailable)

		if sizeInBytes > spaceAvailable {
			return false, fmt.Errorf("no hay suficiente espacio disponible para ajustar la partición") // No hay suficiente espacio para ajustar la partición
		} else {
			return true, nil
		}
	} else if operationType == "reducir" {
		var partitionSize int64 = int64(currentPartition.Size)

		if (partitionSize - int64(math.Abs(float64(sizeInBytes)))) < 0 {
			return false, fmt.Errorf("el tamaño de la partición no puede ser reducido a menos de 0")
		} else {
			return true, nil
		}
	}

	// if sizeInBytes < 0 && (partitionSize-int64(math.Abs(float64(sizeInBytes)))) < 0 {
	// 	return false // No se puede reducir la partición a tamaño negativo
	// }

	// //if (sizeInBytes < 0 && partitionSize - math.Abs(sizeInBytes) < 0) || (sizeInBytes > 0 && partitionSize + sizeInBytes > int64(mbr.MbrTamano)) {

	// // Verificar si se reduce el tamaño y se mantiene positivo
	// if sizeInBytes < 0 && (int64(currentPartition.Size)+sizeInBytes) < 0 {
	// 	return false // No se puede reducir la partición a tamaño negativo
	// }

	// // Verificar si se puede expandir la partición
	// if sizeInBytes > 0 {
	// 	// Calcula el espacio total utilizado por todas las particiones

	// 	// Verificar si el espacio disponible es suficiente para la expansión
	// 	if sizeInBytes > spaceAvailable {
	// 		return false // No hay suficiente espacio para expandir
	// 	}
	// }
	return true, nil
}

func findPartitionByName(mbr *Types.MBR, partitionName string) (int, Types.Partition) {
	for i, partition := range mbr.Partitions {
		// Asume que el nombre de la partición se almacena en un array de bytes y necesita ser convertido a string
		if string(partition.Name[:]) == partitionName {
			return i, partition
		}
	}
	// Retorna -1 y una partición vacía si no se encuentra
	return -1, Types.Partition{}
}

func convertUnitToAddValue(addValue int64, unit string) int64 {
	unit = strings.ToUpper(unit)
	switch unit {
	case "B":
		return addValue
	case "K":
		return addValue * 1024
	case "M":
		return addValue * 1024 * 1024
	default:
		return 0
	}
}

func AdjustPartitionSize(mbr *Types.MBR, partitionName string, addValue int64, unit string, diskFileName string) error {
	// Conversión del valor add según la unidad
	var sizeInBytes int64 = convertUnitToAddValue(addValue, unit)

	// fmt.Println("Espacio a ajustar:", sizeInBytes)
	// fmt.Println("Partición:", partitionName)

	// partitionIndex, _ := findPartitionByName(mbr, partitionName)
	// if partitionIndex == -1 {
	// 	return fmt.Errorf("la partición '%s' no se encontró", partitionName)
	// }

	var partitionIndex int64 = 0

	for i, partition := range mbr.Partitions {
		// Asume que el nombre de la partición se almacena en un array de bytes y necesita ser convertido a string
		if string(partition.Name[:]) == partitionName {
			partitionIndex = int64(i)
		}
	}

	//fmt.Println("Partición encontrada:", partitionIndex)

	// Verifica si se puede agregar o quitar espacio
	canAdjust, err := canAdjustPartitionSize(mbr, int(partitionIndex), sizeInBytes)
	if err != nil {
		return err
	}
	if !canAdjust {
		return fmt.Errorf(err.Error())
	}

	//fmt.Println("Ajustando el tamaño de la partición...")

	// Ajusta el tamaño de la partición
	mbr.Partitions[partitionIndex].Size += int32(sizeInBytes)

	WriteMBR(diskFileName, *mbr)

	// Opcional: Rellenar con '\0' si se reduce el tamaño y se especifica "Full"

	return nil
}

func printMBRState(mbr *Types.MBR) {
	fmt.Println("Estado actual del MBR:")
	fmt.Printf("Tamaño del Disco: %d\n", mbr.MbrTamano)
	fmt.Printf("Firma del Disco: %d\n", mbr.MbrDiskSignature)
	for i, p := range mbr.Partitions {
		fmt.Printf("Partición %d: %+v\n", i+1, p)
	}
}

func PrintLogicalPartitions(info *LogicalPartitionInfo) {
	if info == nil || info.FirstEBR == nil {
		fmt.Println("No logical partitions found.")
		return
	}

	currentEBR := info.FirstEBR
	partitionNumber := 1

	// Assuming 'LogicalPartitionInfo' and 'EBR' are defined with the necessary fields
	// And 'EBR' has fields like 'Size', 'Start', and potentially 'Name' for the partition it describes
	fmt.Println("List of Logical Partitions:")
	for currentEBR != nil {
		// Print details of the logical partition
		fmt.Printf("Partition %d:\n", partitionNumber)
		fmt.Printf("  Name: %s\n", currentEBR.PartName[:])
		fmt.Printf("  Start: %d\n", currentEBR.PartStart)
		fmt.Printf("  Size: %d\n", currentEBR.PartSize)

		// Move to the next EBR in the chain
		currentEBR = currentEBR.PartNext
		partitionNumber++
	}
}

func DeleteLogicalPartition(info *LogicalPartitionInfo, partitionName string) error {
	if info == nil || info.FirstEBR == nil {
		return fmt.Errorf("no logical partitions to delete")
	}

	var previousEBR *Types.EBR = nil
	currentEBR := info.FirstEBR

	// Search for the EBR corresponding to the partition to be deleted
	for currentEBR != nil && string(currentEBR.PartName[:]) == partitionName {
		previousEBR = currentEBR
		currentEBR = currentEBR.PartNext
	}

	// If the partition was not found
	if currentEBR == nil {
		return fmt.Errorf("logical partition with name %s not found", partitionName)
	}

	// If the partition to be deleted is the first in the list
	if previousEBR == nil {
		info.FirstEBR = currentEBR.PartNext // Bypass the deleted partition's EBR
	} else {
		// If the partition is in the middle or end of the list
		previousEBR.PartNext = currentEBR.PartNext // Bypass the deleted partition's EBR
	}

	return nil
}

func GetPartitionId(mbr *Types.MBR, name string) (string, error) {
	partitionId := ""
	partitionExists := false

	var partitionName string = ""
	for _, partition := range mbr.Partitions {
		partitionName = Utils.CleanPartitionName(partition.Name[:])
		//fmt.Println("Particion:", string(partition.Name[:]))
		if strings.TrimSpace(partitionName) == strings.TrimSpace(name) {
			//fmt.Println("Particion encontrada:", string(partition.Name[:]))
			partitionExists = true
			partitionId = string(partition.Id[:])
			break
		}
	}

	if !partitionExists {
		return "", fmt.Errorf("la partición %s no existe", name)
	}

	return partitionId, nil
}

func ValidatePartitionId(mbr *Types.MBR, id string) (string, error) {
	//partitionId := ""
	partitionExists := false

	var partitionId string = ""
	for _, partition := range mbr.Partitions {
		//fmt.Printf("Fdisk Particion: %+v\n", partition)
		//fmt.Printf("Fdisk Particion Id: %s\n", string(partition.Id[:]))
		partitionId = Utils.CleanPartitionName(partition.Id[:])
		//fmt.Printf("Fdisk Clean Particion Id: %s\n", partitionId)
		//fmt.Println("Particion:", string(partition.Name[:]))
		if strings.TrimSpace(partitionId) == strings.TrimSpace(id) {
			//fmt.Println("Particion encontrada:", string(partition.Name[:]))
			partitionExists = true
			//partitionId = string(partition.Id[:])
			break
		}
	}

	if !partitionExists {
		return "", fmt.Errorf("la partición %s no existe", id)
	}

	return partitionId, nil
}
