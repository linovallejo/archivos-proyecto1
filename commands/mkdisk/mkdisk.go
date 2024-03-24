package Mkdisk

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	Types "proyecto1/types"
	Utils "proyecto1/utils"
	"strconv"
	"strings"
	"time"
)

func ExtractMkdiskParams(params []string) (int64, string, string, error) {
	var size int64
	var unit string = "M" // Megabytes por defecto
	var fit string = "FF" // First Fit por defecto

	if len(params) == 0 {
		return 0, "", "", fmt.Errorf("No se encontraron parámetros")
	}

	var parametrosObligatoriosOk bool = false
	for _, param1 := range params {
		if strings.HasPrefix(param1, "-size=") {
			parametrosObligatoriosOk = true
			break
		}
	}

	if !parametrosObligatoriosOk {
		return 0, "", "", fmt.Errorf("No se encontraron parámetros obligatorios")
	}

	for _, param := range params {
		if strings.HasPrefix(param, "-size=") {
			sizeStr := strings.TrimPrefix(param, "-size=")
			var err error
			if strings.TrimSpace(sizeStr) == "" {
				return 0, "", "", fmt.Errorf("Parametro tamaño es obligatorio")
			}
			size, err = strconv.ParseInt(sizeStr, 10, 64)
			if err != nil || size <= 0 {
				return 0, "", "", fmt.Errorf("Parametro tamaño invalido")
			}
		} else if strings.HasPrefix(param, "-unit=") {
			unit = strings.TrimPrefix(param, "-unit=")
			unit = strings.ToUpper(unit)
			if unit != "K" && unit != "M" {
				return 0, "", "", fmt.Errorf("Parametro unidad invalido")
			}
		} else if strings.HasPrefix(param, "-fit=") {
			fit = strings.TrimPrefix(param, "-fit=")
			fit = strings.ToLower(fit)
			if fit != "ff" && fit != "bf" && fit != "wf" {
				return 0, "", "", fmt.Errorf("Parametro ajuste invalido")
			}
		}
	}

	if size == 0 {
		return 0, "", "", fmt.Errorf("Parametro tamaño es obligatorio")
	}

	return size, unit, fit, nil
}

func CalculateDiskSize(size int64, unit string) (int64, error) {
	switch unit {
	case "K":
		return size * 1024, nil
	case "M":
		return size * 1024 * 1024, nil
	default:
		return 0, fmt.Errorf("Unidad invalida")
	}
}

func CreateDiskWithSize(filename string, size int32, fit string) {
	var mbr Types.MBR
	mbr.MbrTamano = size
	currentTime := time.Now()
	copy(mbr.MbrFechaCreacion[:], currentTime.Format("2006-01-02T15:04:05"))
	mbr.MbrDiskSignature = generateDiskSignature()
	mbr.DskFit = Utils.ReturnFitType(fit)

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

	err = file.Truncate(int64(size))
	if err != nil {
		fmt.Printf("Error al asignar espacio en disco: %v\n", err)
		return
	}

	// fmt.Println("Disco creado correctamente con tamaño:", size, "bytes.")

	// printMBRState(&mbr)
}

func ConstructFileName(path string) string {
	const baseName = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	for _, letter := range baseName {
		filename := fmt.Sprintf("%c.dsk", letter) // Asigna el nombre de archivo actual aquí
		fullPath := filepath.Join(path, filename)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			return fullPath // Retorna inmediatamente el primer nombre de archivo disponible
		}
	}
	return "" // Puedes retornar un error o un valor vacío si todos los nombres están ocupados
}

func printMBRState(mbr *Types.MBR) {
	fmt.Println("Estado actual del MBR:")
	fmt.Printf("Tamaño del Disco: %d\n", mbr.MbrTamano)
	fmt.Printf("Firma del Disco: %d\n", mbr.MbrDiskSignature)
	for i, p := range mbr.Partitions {
		fmt.Printf("Partición %d: %+v\n", i+1, p)
	}
}

func generateDiskSignature() (signature int32) {
	source := rand.NewSource(time.Now().UnixNano())
	generator := rand.New(source)
	signature = generator.Int31()
	return
}

func CalcularEspacioLibreDisco(mbr *Types.MBR) int32 {
	espacioTotal := mbr.MbrTamano
	espacioUsado := int32(0)

	for _, partition := range mbr.Partitions {
		// Sumar el tamaño de las particiones asignadas al espacio usado
		if partition.Status[0] == 0 || partition.Status[0] == 1 { // Asume que el status '0' significa no asignado
			espacioUsado += partition.Size
		}
	}

	espacioLibre := espacioTotal - espacioUsado
	return espacioLibre
}
