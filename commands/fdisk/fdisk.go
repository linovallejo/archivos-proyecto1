package Fdisk

import (
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

func ExtractFdiskParams(params []string) (int64, string, string, string, string, string, string, error) {
	var size int64
	var driveletter, unit, letter, name, fit, parttype string

	for _, param := range params {
		switch {
		case strings.HasPrefix(param, "-size="):
			sizeStr := strings.TrimPrefix(param, "-size=")
			var err error
			size, err = strconv.ParseInt(sizeStr, 10, 64)
			if err != nil || size <= 0 {
				return 0, "", "", "", "", "", "", fmt.Errorf("Parametro tamaño invalido")
			}
		case strings.HasPrefix(param, "-driveletter="):
			driveletter = strings.TrimPrefix(param, "-driveletter=")
			// Validar la letra de la partición
			if len(driveletter) != 1 || !unicode.IsLetter(rune(driveletter[0])) {
				return 0, "", "", "", "", "", "", fmt.Errorf("La letra de la partición debe ser un único carácter alfabérico")
			}
		case strings.HasPrefix(param, "-name="):
			name = strings.TrimPrefix(param, "-name=")
			// Validar el nombre de la partición
			if len(name) > 16 {
				return 0, "", "", "", "", "", "", fmt.Errorf("El nombre de la partición no puede exceder los 16 caracteres")
			}
		case strings.HasPrefix(param, "-unit="):
			unit = strings.TrimPrefix(param, "-unit=")
			if unit != "B" && unit != "K" && unit != "M" {
				return 0, "", "", "", "", "", "", fmt.Errorf("Parametro unidad invalido")
			}
		case strings.HasPrefix(param, "-type="):
			parttype = strings.TrimPrefix(param, "-type=")
			// Validar el tipo de la partición
			if parttype != "P" && parttype != "E" && parttype != "L" {
				return 0, "", "", "", "", "", "", fmt.Errorf("El tipo de la partición debe ser 'P', 'E' o 'L'")
			}
		case strings.HasPrefix(param, "-fit="):
			fit = strings.TrimPrefix(param, "-fit=")
			if fit != "BF" && fit != "FF" && fit != "WF" {
				return 0, "", "", "", "", "", "", fmt.Errorf("Parametro fit invalido")
			}
		}
	}

	if size == 0 || letter == "" || name == "" {
		return 0, "", "", "", "", "", "", fmt.Errorf("Parametro obligatorio faltante")
	}

	// Unidad por defecto es Kilobytes
	if unit == "" {
		unit = "K"
	}

	return size, driveletter, unit, letter, name, fit, parttype, nil
}

func ReadMBR(filename string) (Types.MBR, error) {
	var mbr Types.MBR

	file, err := os.Open(filename)
	if err != nil {
		return mbr, err
	}
	defer file.Close()

	// Omitir el byte nulo inicial
	_, err = file.Seek(1, io.SeekStart)
	if err != nil {
		return mbr, err
	}

	err = binary.Read(file, binary.LittleEndian, &mbr)
	return mbr, err
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

func createPartition(mbr *Types.MBR, size int64, unit string, letter string, name string) error {
	var sizeInBytes int64
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

	// Verifica si hay espacio suficiente para la nueva partición (asumiendo una asignación lineal para simplificar).
	var totalUsedSpace int64 = 0
	for _, partition := range mbr.Partitions {
		if partition.Status != 0 {
			totalUsedSpace += partition.Size
		}
	}
	if totalUsedSpace+sizeInBytes > mbr.MbrTamano {
		return fmt.Errorf("No hay suficiente espacio para la nueva partición")
	}

	// Crear y "setear" la nueva partición
	newPartition := Types.Partition{
		Status: 1,
		Type:   0,
		Fit:    0,
		Start:  totalUsedSpace + 1,
		Size:   sizeInBytes,
		Name:   [16]byte{},
	}
	copy(newPartition.Name[:], name)
	mbr.Partitions[partitionIndex] = newPartition

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
func AdjustAndCreatePartition(mbr *Types.MBR, size int64, fit string) error {
	spaces := calculateAvailableSpaces(mbr)
	var selectedSpace *Space

	switch fit {
	case "FF":
		selectedSpace = findFirstFit(spaces, size)
	case "BF":
		selectedSpace = findBestFit(spaces, size)
	case "WF":
		selectedSpace = findWorstFit(spaces, size)
	}

	if selectedSpace == nil {
		return fmt.Errorf("No se encontró espacio adecuado para la partición")
	}

	// Aquí, usarías selectedSpace.Start como la posición de inicio de tu nueva partición
	// y procederías a crear la partición con el tamaño especificado.
	createPartition(mbr, size, "B", "", "")

	return nil
}

func ValidatePartitionCreation(mbr *Types.MBR, partType string) error {
	var countP, countE int

	for _, partition := range mbr.Partitions {
		switch partition.Type {
		case 'P': // Asume que 'P' representa una partición Primaria
			countP++
		case 'E': // Asume que 'E' representa una partición Extendida
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

func ConstructAndValidateFileName(path string, filename string) (string, error) {
	fullPath := filepath.Join(path, filename)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		// El archivo no existe
		return "", fmt.Errorf("el archivo %s no existe", fullPath)
	}
	// El archivo existe
	return fullPath, nil
}

func GenerateDotCode(mbr *Types.MBR) string {
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
