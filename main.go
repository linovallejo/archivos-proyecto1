package main

import (
	"bufio"
	"fmt"
	"os"
	Fdisk "proyecto1/commands/fdisk"
	Mkdisk "proyecto1/commands/mkdisk"
	Utils "proyecto1/utils"
	"strconv"
	"strings"
	"unicode"
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
		if strings.HasPrefix(command, "mkdisk") {
			params := strings.Fields(command)
			mkdisk(archivoBinarioDisco, params[1:])
		} else if strings.HasPrefix(command, "fdisk") {
			params := strings.Fields(command)
			fdisk(params[1:])
		} else if command == "rep" {
			//rep(archivoBinarioDisco)
			fmt.Println("rep")
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
	size, unit, err := Mkdisk.ExtractMKDISKParams(params)
	if err != nil {
		fmt.Println("Error al procesar los parámetros MKDISK:", err)
		return
	}

	// Tamaño del disco en bytes
	sizeInBytes, err := Mkdisk.CalculateDiskSize(size, unit)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Creación del disco con el tamaño calculado en bytes
	Mkdisk.CreateDiskWithSize(filename, sizeInBytes)
}

func extractFdiskParams(params []string) (int64, string, string, string, string, string, string, error) {
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

func fdisk(params []string) {
	size, _, _, _, _, fit, parttype, err := extractFdiskParams(params)
	// size, driveletter, unit, letter, name, fit, parttype, err := extractFdiskParams(params)

	if err != nil {
		fmt.Println("Error al procesar los parámetros FDISK:", err)
		return
	}

	// Leer el MBR existente
	mbr, err := Fdisk.ReadMBR(archivoBinarioDisco)
	if err != nil {
		fmt.Println("Error leyendo el MBR:", err)
		return
	}

	// Validar la creación de la partición
	err = Fdisk.ValidatePartitionCreation(&mbr, parttype)
	if err != nil {
		fmt.Println("Error al validar la creación de la partición:", err)
	}

	err = Fdisk.AdjustAndCreatePartition(&mbr, size, fit)
	if err != nil {
		fmt.Println("Error al ajustar y crear la partición:", err)
	} else {
		// Escribir el MBR actualizado con la nueva partición en el disco
		// err = writeMBR(archivoBinarioDisco, mbr)
		// if err != nil {
		// 	fmt.Println("Error escribiendo el MBR actualizado:", err)
		// } else {
		// 	fmt.Println("Partición creada exitosamente.")
		// }
		fmt.Println("Partición creada exitosamente.")

	}

	// Crear la partición
	// err = createPartition(&mbr, size, unit, letter, name)
	// if err != nil {
	// 	fmt.Println("Error creando la particion:", err)
	// } else {
	// 	// Escribir el MR actualizado con la nueva partición en el disco
	// 	err = writeMBR(archivoBinarioDisco, mbr)
	// 	if err != nil {
	// 		fmt.Println("Error writing updated MBR:", err)
	// 	} else {
	// 		fmt.Println("Particion creada exitosamente.")
	// 	}
	// }
}
