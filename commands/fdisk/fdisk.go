package Fdisk

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
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

	for _, param := range params {
		switch {
		case strings.HasPrefix(param, "-size="):
			sizeStr := strings.TrimPrefix(param, "-size=")
			var err error
			size, err = strconv.ParseInt(sizeStr, 10, 64)
			if err != nil || size <= 0 {
				return 0, "", "", "", "", "", "", 0, fmt.Errorf("Parametro tamaño invalido")
			}
		case strings.HasPrefix(param, "-driveletter="):
			driveletter = strings.TrimPrefix(param, "-driveletter=")
			// Validar la letra de la partición
			if len(driveletter) != 1 || !unicode.IsLetter(rune(driveletter[0])) {
				return 0, "", "", "", "", "", "", 0, fmt.Errorf("La letra de la partición debe ser un único carácter alfabérico")
			}
		case strings.HasPrefix(param, "-name="):
			name = strings.TrimPrefix(param, "-name=")
			// Validar el nombre de la partición
			if len(name) > 16 {
				return 0, "", "", "", "", "", "", 0, fmt.Errorf("El nombre de la partición no puede exceder los 16 caracteres")
			}
		case strings.HasPrefix(param, "-unit="):
			unit = strings.TrimPrefix(param, "-unit=")
			// Validar la unidad de la partición
			if unit != "B" && unit != "K" && unit != "M" {
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
			if fit != "BF" && fit != "FF" && fit != "WF" {
				return 0, "", "", "", "", "", "", 0, fmt.Errorf("Parametro fit invalido")
			}
		case strings.HasPrefix(param, "-delete"):
			delete = strings.TrimPrefix(param, "-delete")
			// Validar el parametro delete
			if delete != "fast" && delete != "full" {
				return 0, "", "", "", "", "", "", 0, fmt.Errorf("Parametro delete invalido")
			}
		case strings.HasPrefix(param, "-add"):
			add = strings.TrimPrefix(param, "-add")
			// Validar el parametro add
			addValue, err := strconv.Atoi(add)
			if err != nil {
				fmt.Println("Parametro add invalido")
				continue
			}
			if addValue < 0 {
				fmt.Println("Parametro add es negativo")
			}
		}
	}

	if size == 0 || driveletter == "" || name == "" {
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

func writeMBR(filename string, mbr Types.MBR) error {
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

func calculateTotalUsedSpace(mbr Types.MBR) (int64, error) {
	var totalUsedSpace int64 = 0
	for _, partition := range mbr.Partitions {
		if partition.Status != 0 {
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

func createPartition(mbr *Types.MBR, start int64, size int64, unit string, typePart, fit, name string, diskFileName string) error {
	var partitionName [16]byte
	copy(partitionName[:], name)

	var sizeInBytes int64 = 0
	// fmt.Println("El tamaño de la partición es: ", size, " en ", unit)
	switch unit {
	case "B":
		sizeInBytes = size
	case "K":
		sizeInBytes = size * 1024
	case "M":
		sizeInBytes = size * 1024 * 1024
	default:
		return fmt.Errorf("Unidad invalida")
	}

	// fmt.Println("El tamaño calculado de la partición es: ", sizeInBytes)
	// Comprbar si hay un slot disponible para la partición disponible y validar la unicidad del nombre
	partitionIndex := -1
	for i, partition := range mbr.Partitions {
		if partition.Status == 0 { // Asume que 0 indica un slot disponible
			if partitionIndex == -1 {
				partitionIndex = i
			}
		} else if string(partition.Name[:]) == name {
			return fmt.Errorf("Nombre de particion ya existe")
		}
	}

	if partitionIndex == -1 {
		return fmt.Errorf("No hay slots disponibles para la nueva partición")
	}

	// fmt.Println("El indice de La partición es: ", partitionIndex)

	// Verifica si hay espacio suficiente para la nueva partición (asumiendo una asignación lineal para simplificar).
	var totalUsedSpace int64 = 0
	var err error

	totalUsedSpace, err = calculateTotalUsedSpace(*mbr)
	if err != nil {
		fmt.Println("Error calculando el espacio total usado: ", err)
		return err
	}

	// fmt.Println("El valor de totalUsedSpace es: ", totalUsedSpace)
	// fmt.Println("El valor de sizeInBytes es: ", sizeInBytes)
	// fmt.Println("El valor de mbr.MbrTamano es: ", mbr.MbrTamano)

	if (totalUsedSpace + sizeInBytes) > mbr.MbrTamano {
		return fmt.Errorf("No hay suficiente espacio para la nueva partición")
	}

	// fmt.Println("Espacio disponible: ", mbr.MbrTamano-totalUsedSpace)

	// Convertir typePart y fit a sus representaciones de byte correspondientes
	var typeByte [1]byte
	typeByte[0] = typePart[0] // 'P', 'E' o 'L'
	var fitByte byte
	switch fit {
	case "BF":
		fitByte = 'B'
	case "FF":
		fitByte = 'F'
	case "WF":
		fitByte = 'W'
	default:
		fitByte = 'F' // FF es el valor predeterminado
	}

	// fmt.Println("El valor de typeByte es: ", typeByte[0])
	// fmt.Println("El valor de fitByte es: ", fitByte)

	// Crear y "setear" la nueva partición
	newPartition := Types.Partition{
		Status: 1,
		Type:   typeByte,
		Fit:    fitByte,
		Start:  start,
		Size:   sizeInBytes,
		Name:   partitionName,
	}

	// fmt.Println("La nueva partición es: ", newPartition)

	// fmt.Println("Particion creada con exito")
	// fmt.Println(newPartition.Name)
	// fmt.Println("------------------------------------------------")

	copy(newPartition.Name[:], name)
	mbr.Partitions[partitionIndex] = newPartition

	writeMBR(diskFileName, *mbr)

	return nil
}

func AdjustAndCreatePartition(mbr *Types.MBR, size int64, unit, typePart, fit, name string, diskFileName string) error {
	spaces := calculateAvailableSpaces(mbr)
	var selectedSpace *Space

	switch fit {
	case "FF":
		selectedSpace = findFirstFit(spaces, size)
	case "BF":
		selectedSpace = findBestFit(spaces, size)
	case "WF":
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
	var lastPosition int64 = 1 // Asume que el disco comienza en la posición 1

	for _, partition := range mbr.Partitions {
		if partition.Status != 0 { // Asume que Status != 0 significa partición ocupada
			if lastPosition < partition.Start {
				// Hay espacio libre entre la última posición y el inicio de la partición actual
				spaces = append(spaces, Space{Start: lastPosition, Size: partition.Start - lastPosition})
			}
			lastPosition = partition.Start + partition.Size
		}
	}

	// Considera el espacio hasta el final del disco
	if lastPosition < mbr.MbrTamano {
		spaces = append(spaces, Space{Start: lastPosition, Size: mbr.MbrTamano - lastPosition})
	}

	return spaces
}

func findFirstFit(spaces []Space, size int64) *Space {
	for _, space := range spaces {
		if space.Size >= size {
			return &space
		}
	}
	return nil
}

func findBestFit(spaces []Space, size int64) *Space {
	var bestSpace *Space
	for _, space := range spaces {
		if space.Size >= size && (bestSpace == nil || space.Size < bestSpace.Size) {
			bestSpace = &space
		}
	}
	return bestSpace
}

func findWorstFit(spaces []Space, size int64) *Space {
	var worstSpace *Space
	for _, space := range spaces {
		if space.Size >= size && (worstSpace == nil || space.Size > worstSpace.Size) {
			worstSpace = &space
		}
	}
	return worstSpace
}

// Asume que tienes una función que acepta estos parámetros y crea la partición.

func ValidatePartitionTypeCreation(mbr *Types.MBR, partType string) error {
	var countP, countE int

	for _, partition := range mbr.Partitions {
		switch partition.Type {
		case [1]byte{'P'}: // Asume que 'P' representa una partición Primaria
			countP++
		case [1]byte{'E'}: // Asume que 'E' representa una partición Extendida
			countE++
		}
	}

	if partType == "E" && countE > 0 {
		return fmt.Errorf("Ya existe una partición extendida en el disco")
	}

	if (partType == "P" || partType == "E") && (countP+countE) >= 4 {
		return fmt.Errorf("No se pueden crear más particiones primarias o extendidas (límite de 4)")
	}

	// Para L, asegúrate de que ya existe una partición Extendida
	if partType == "L" && countE == 0 {
		return fmt.Errorf("Debe existir una partición extendida para crear una partición lógica")
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

func GenerateDotCodeMbr(mbr *Types.MBR) string {
	var builder strings.Builder

	builder.WriteString("digraph G {\n")
	builder.WriteString("    node [shape=none];\n")
	builder.WriteString("    rankdir=\"LR\";\n")

	builder.WriteString("    struct1 [label=<<TABLE BORDER=\"0\" CELLBORDER=\"1\" CELLSPACING=\"0\">\n")

	builder.WriteString("    <TR>")
	// Nodo MBR
	builder.WriteString("    <TD BGCOLOR=\"yellow\">MBR</TD>")

	// Nodos de particiones
	for i, partition := range mbr.Partitions {
		if partition.Status != 0 { // Asumiendo que Status != 0 significa que la partición esta disponible.
			partitionName := Utils.CleanPartitionName(partition.Name[:])
			if partitionName == "" {
				partitionName = fmt.Sprintf("Partition%d", i+1)
			}
			builder.WriteString(fmt.Sprintf("    <TD BGCOLOR=\"green\">%s</TD>", partitionName))
		}
	}

	builder.WriteString("</TR>\n")

	builder.WriteString("    </TABLE>>];\n")

	builder.WriteString("}\n")

	return builder.String()
}

func GenerateDotCodeDisk(mbr *Types.MBR) string {
	var dot bytes.Buffer

	dot.WriteString("digraph G {\n")
	dot.WriteString("node [shape=plaintext]\n")
	dot.WriteString("struct1 [label=<\n")
	dot.WriteString("<table border=\"0\" cellborder=\"1\" cellspacing=\"0\" cellpadding=\"4\">\n")

	// Primera fila para MBR y particiones que no sean lógicas
	dot.WriteString("<tr>\n")
	dot.WriteString("<td>MBR</td>\n") // MBR siempre está presente
	for _, p := range mbr.Partitions {
		if p.Status != 0 && p.Type != [1]byte{'L'} { // Revisar si la partición no es lógica y está activa
			partitionName := string(p.Name[:])                              // Convertir el nombre de la partición a string
			partitionName = Utils.CleanPartitionName([]byte(partitionName)) // Convertir partitionName a []byte antes de pasarlo como argumento
			dot.WriteString(fmt.Sprintf("<td>%s<br/>%d bytes</td>\n", partitionName, p.Size))
		}
	}
	dot.WriteString("</tr>\n")

	// Segunda fila para particiones lógicas si existe una extendida
	for _, p := range mbr.Partitions {
		if p.Status != 0 && p.Type == [1]byte{'E'} { // Si hay una extendida, asumimos que hay lógicas
			dot.WriteString("<tr>\n")
			dot.WriteString("<td colspan=\"3\">Extendida</td>\n") // Colspan basado en la cantidad de lógicas
			for _, subP := range mbr.Partitions {
				if subP.Status != 0 && subP.Type == [1]byte{'L'} { // Revisar si la partición es lógica
					partitionName := string(subP.Name[:]) // Convertir el nombre de la partición a string
					partitionName = Utils.CleanPartitionName([]byte(partitionName))
					dot.WriteString(fmt.Sprintf("<td>%s<br/>%d bytes</td>\n", partitionName, subP.Size))
				}
			}
			dot.WriteString("</tr>\n")
			break // Solo una fila para las lógicas
		}
	}

	dot.WriteString("</table>\n")
	dot.WriteString(">];\n")
	dot.WriteString("}\n")

	return dot.String()
}

func ValidatePartitionName(mbr *Types.MBR, name string, delete string) error {
	partitionExists := false
	for _, partition := range mbr.Partitions {
		if string(partition.Name[:]) == name {
			partitionExists = true
			break
		}
	}

	if delete == "full" {
		if !partitionExists {
			return fmt.Errorf("la partición a eliminar '%s' no existe", name)
		}
	} else {
		if partitionExists {
			return fmt.Errorf("el nombre de la partición '%s' ya está en uso", name)
		}
	}

	return nil
}

func CalculateSize(size int64, unit string) (int64, error) {
	switch unit {
	case "K":
		return size * 1024, nil
	case "M":
		return size * 1024 * 1024, nil
	default:
		return 0, fmt.Errorf("unidad invalida")
	}
}

func DeletePartition(mbr *Types.MBR, filename string, partitionName string) error {
	// Verificar la existencia de la partición
	partitionIndex := -1
	for i, p := range mbr.Partitions {
		if string(p.Name[:]) == partitionName {
			partitionIndex = i
			break
		}
	}
	if partitionIndex == -1 {
		return fmt.Errorf("la partición '%s' no existe", partitionName)
	}

	// Confirmar la eliminación de la partición
	fmt.Printf("¿Estás seguro de que quieres eliminar la partición '%s'? [s/N]: ", partitionName)
	var response string
	fmt.Scanln(&response)
	if response != "s" && response != "S" {
		return fmt.Errorf("eliminación cancelada por el usuario")
	}

	// Eliminar la partición
	// Si la partición es extendida, también elimina sus particiones lógicas
	if string(mbr.Partitions[partitionIndex].Type[:]) == "E" {
		// TODO
		// Implementar lógica para eliminar particiones lógicas si es necesario
		return nil
	}
	mbr.Partitions[partitionIndex] = Types.Partition{} // Asigna una partición vacía

	// Rellenar con `\0`
	start, size, err := getPartitionDetails(mbr, partitionName)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("La partición '%s' comienza en %d bytes y tiene un tamaño de %d bytes.\n", partitionName, start, size)
	}

	err = cleanPartitionSpace(filename, start, size)
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

func getPartitionDetails(mbr *Types.MBR, partitionName string) (int64, int64, error) {
	for _, partition := range mbr.Partitions {
		if string(partition.Name[:]) == partitionName {
			return partition.Start, partition.Size, nil
		}
	}
	return 0, 0, fmt.Errorf("la partición '%s' no se encontró", partitionName)
}

func canAdjustPartitionSize(mbr *Types.MBR, partitionIndex int, sizeInBytes int64) bool {
	currentPartition := mbr.Partitions[partitionIndex]

	// Verificar si se reduce el tamaño y se mantiene positivo
	if sizeInBytes < 0 && (currentPartition.Size+sizeInBytes) < 0 {
		return false // No se puede reducir la partición a tamaño negativo
	}

	// Verificar si se puede expandir la partición
	if sizeInBytes > 0 {
		// Calcula el espacio total utilizado por todas las particiones
		var totalUsedSpace int64 = 0
		for _, partition := range mbr.Partitions {
			if partition.Status != 0 { // Asume Status != 0 como partición activa
				totalUsedSpace += partition.Size
			}
		}

		// Espacio disponible en el disco
		spaceAvailable := mbr.MbrTamano - totalUsedSpace

		// Verificar si el espacio disponible es suficiente para la expansión
		if sizeInBytes > spaceAvailable {
			return false // No hay suficiente espacio para expandir
		}
	}

	return true // El ajuste de tamaño es viable
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

func AdjustPartitionSize(mbr *Types.MBR, partitionName string, addValue int64, unit string) error {
	// Conversión del valor add según la unidad
	var sizeInBytes int64 = convertUnitToAddValue(addValue, unit)

	partitionIndex, _ := findPartitionByName(mbr, partitionName)
	if partitionIndex == -1 {
		return fmt.Errorf("la partición '%s' no se encontró", partitionName)
	}

	// Verifica si se puede agregar o quitar espacio
	if !canAdjustPartitionSize(mbr, partitionIndex, sizeInBytes) {
		return fmt.Errorf("no es posible ajustar el tamaño de la partición '%s'", partitionName)
	}

	// Ajusta el tamaño de la partición
	mbr.Partitions[partitionIndex].Size += sizeInBytes

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
