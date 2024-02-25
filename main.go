package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	Types "proyecto1/types"
	Utils "proyecto1/utils"
	"strconv"
	"strings"
	"time"
)

var rutaDiscos string = "./disks/MIA/P1/"
var archivoBinarioDisco string = "primario.dsk"

func main() {
	Utils.LimpiarConsola()
	Utils.PrintCopyright()
	fmt.Println("Procesador de Comandos - Proyecto 1")

	var input string
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Ingrese el comando:")
	scanner.Scan()
	input = scanner.Text()

	comando, path := parseCommand(input)
	if comando != "EXECUTE" || path == "" {
		fmt.Println("Comando no reconocido o ruta de archivo faltante. Uso: EXECUTE <ruta_al_archivo_de_scripts>")
		return
	}

	path = strings.Trim(path, `"'`)

	fmt.Printf("Leyendo el archivo de scripts de: %s\n", path)

	content, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("Error leyendo el archivo de scripts: %v\n", err)
		return
	}

	contentStr := string(content)
	contentStr = strings.Replace(contentStr, "\r\n", "\n", -1) // Convertir CRLF a LF
	commands := strings.Split(contentStr, "\n")

	for _, command := range commands {
		if strings.HasPrefix(command, "MKDISK") {
			params := strings.Fields(command)
			mkdisk(archivoBinarioDisco, params[1:])
		} else if strings.HasPrefix(command, "FDISK") {
			params := strings.Fields(command)
			fdisk(params[1:])
		} else if command == "REP" {
			//rep(archivoBinarioDisco)
			fmt.Println("REP")
		}
	}
}

func parseCommand(input string) (string, string) {
	parts := strings.Fields(input)
	if len(parts) < 2 {
		return "", ""
	}

	command := parts[0]
	var path string

	for _, part := range parts[1:] {
		if strings.HasPrefix(part, "->path=") {
			path = strings.TrimPrefix(part, "->path=")
			break
		}
	}

	return command, path
}

func mkdisk(filename string, params []string) {
	size, unit, err := extractMKDISKParams(params)
	if err != nil {
		fmt.Println("Error al procesar los parámetros MKDISK:", err)
		return
	}

	// Tamaño del disco en bytes
	sizeInBytes, err := calculateDiskSize(size, unit)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Creación del disco con el tamaño calculado en bytes
	createDiskWithSize(filename, sizeInBytes)
}

func extractMKDISKParams(params []string) (int64, string, error) {
	var size int64
	var unit string = "M" // Megabytes por defecto
	var fit string = "FF" // First Fit por defecto

	for _, param := range params {
		if strings.HasPrefix(param, "-size=") {
			sizeStr := strings.TrimPrefix(param, "-size=")
			var err error
			size, err = strconv.ParseInt(sizeStr, 10, 64)
			if err != nil || size <= 0 {
				return 0, "", fmt.Errorf("Parametro tamaño invalido")
			}
		} else if strings.HasPrefix(param, "-unit=") {
			unit = strings.TrimPrefix(param, "-unit=")
			if unit != "K" && unit != "M" {
				return 0, "", fmt.Errorf("Parametro unidad invalido")
			}
		} else if strings.HasPrefix(param, "-fit=") {
			fit = strings.TrimPrefix(param, "-fit=")
			if fit != "FF" && fit != "BF" && fit != "WF" {
				return 0, "", fmt.Errorf("Parametro ajuste invalido")
			}
		}
	}

	if size == 0 {
		return 0, "", fmt.Errorf("Parametro tamaño es obligatorio")
	}

	return size, unit, nil
}

func calculateDiskSize(size int64, unit string) (int64, error) {
	switch unit {
	case "K":
		return size * 1024, nil
	case "M":
		return size * 1024 * 1024, nil
	default:
		return 0, fmt.Errorf("Unidad invalida")
	}
}

func createDiskWithSize(filename string, size int64) {
	var mbr Types.MBR
	mbr.MbrTamano = size
	currentTime := time.Now()
	copy(mbr.MbrFechaCreacion[:], currentTime.Format("2006-01-02T15:04:05"))
	mbr.MbrDiskSignature = 123456789 // Example signature

	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Error creando el disco: %v\n", err)
		return
	}
	defer file.Close()

	_, err = file.Write([]byte{'\x00'})
	if err != nil {
		fmt.Printf("Error escribiendo el caracter inicial: %v\n", err)
		return
	}

	err = binary.Write(file, binary.LittleEndian, &mbr)
	if err != nil {
		fmt.Printf("Error escribiendo el MBR: %v\n", err)
		return
	}

	err = file.Truncate(size)
	if err != nil {
		fmt.Printf("Error al asignar espacio en disco: %v\n", err)
		return
	}

	fmt.Println("Disco creado correctamente con tamaño:", size, "bytes.")
}

func extractFDISKParams(params []string) (int64, string, string, string, error) {
	var size int64
	var unit, letter, name string

	for _, param := range params {
		switch {
		case strings.HasPrefix(param, "->size="):
			sizeStr := strings.TrimPrefix(param, "->size=")
			var err error
			size, err = strconv.ParseInt(sizeStr, 10, 64)
			if err != nil || size <= 0 {
				return 0, "", "", "", fmt.Errorf("Parametro tamaño invalido")
			}
		case strings.HasPrefix(param, "->unit="):
			unit = strings.TrimPrefix(param, "->unit=")
			if unit != "B" && unit != "K" && unit != "M" {
				return 0, "", "", "", fmt.Errorf("Parametro unidad invalido")
			}
		case strings.HasPrefix(param, "->letter="):
			letter = strings.TrimPrefix(param, "->letter=")
			// TODO: Agregue la validación de la letra (si es necesario)
		case strings.HasPrefix(param, "->name="):
			name = strings.TrimPrefix(param, "->name=")
			// TODO: Agregue la validación del nombre (si es necesario)
		}
	}

	if size == 0 || letter == "" || name == "" {
		return 0, "", "", "", fmt.Errorf("Parametro obligatorio faltante")
	}

	// Unidad por defecto es Kilobytes
	if unit == "" {
		unit = "K"
	}

	return size, unit, letter, name, nil
}

func fdisk(params []string) {
	size, unit, letter, name, err := extractFDISKParams(params)
	if err != nil {
		fmt.Println("Error al procesar los parámetros FDISK:", err)
		return
	}

	// Leer el MBR existente
	mbr, err := readMBR(archivoBinarioDisco)
	if err != nil {
		fmt.Println("Error leyendo el MBR:", err)
		return
	}

	// Crear la partición
	err = createPartition(&mbr, size, unit, letter, name)
	if err != nil {
		fmt.Println("Error creando la particion:", err)
	} else {
		// Escribir el MR actualizado con la nueva partición en el disco
		err = writeMBR(archivoBinarioDisco, mbr)
		if err != nil {
			fmt.Println("Error writing updated MBR:", err)
		} else {
			fmt.Println("Particion creada exitosamente.")
		}
	}
}

func readMBR(filename string) (Types.MBR, error) {
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

func cleanPartitionName(name []byte) string {
	n := bytes.IndexByte(name, 0)
	if n == -1 {
		n = len(name)
	}
	return string(name[:n])
}
