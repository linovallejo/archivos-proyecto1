package main

import (
	"bufio"
	"fmt"
	"os"
	"path"
	Fdisk "proyecto1/commands/fdisk"
	Mkdisk "proyecto1/commands/mkdisk"
	Rep "proyecto1/commands/rep"
	Reportes "proyecto1/reportes"
	Utils "proyecto1/utils"
	"strings"
)

var rutaDiscos string = "./disks/MIA/P1/"
var archivoBinarioDiscoActual string = ""
var ajusteParticionActual string = "" // first fit, best fit, worst fit

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
	if comando != "execute" || path == "" {
		fmt.Println("Comando no reconocido o ruta de archivo faltante. Uso: execute <ruta_al_archivo_de_scripts>")
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
			archivoBinarioDiscoActual = mkdisk(params[1:])
		} else if strings.HasPrefix(command, "fdisk") {
			params := strings.Fields(command)
			fdisk(params[1:])
		} else if strings.HasPrefix(command, "rep") {
			params := strings.Fields(command)
			rep(archivoBinarioDiscoActual, params[1:])
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
		if strings.HasPrefix(part, "-path=") {
			path = strings.TrimPrefix(part, "-path=")
			break
		}
	}

	return command, path
}

func mkdisk(params []string) string {
	size, unit, diskFit, err := Mkdisk.ExtractMkdiskParams(params)
	if err != nil {
		fmt.Println("Error al procesar los parámetros MKDISK:", err)
		return ""
	}

	// Tamaño del disco en bytes
	sizeInBytes, err := Mkdisk.CalculateDiskSize(size, unit)
	if err != nil {
		fmt.Println("Error:", err)
		return ""
	}

	// Construye el nombre del disco apropiado
	var filename string = Mkdisk.ConstructFileName(rutaDiscos)

	ajusteParticionActual = diskFit

	// Creación del disco con el tamaño calculado en bytes
	Mkdisk.CreateDiskWithSize(filename, sizeInBytes)

	fmt.Println("Disco creado con éxito!")

	return filename
}

func fdisk(params []string) {
	size, driveletter, name, unit, parttype, fit, delete, addValue, err := Fdisk.ExtractFdiskParams(params)
	// size, driveletter, unit, letter, name, fit, parttype, err := extractFdiskParams(params)

	if err != nil {
		fmt.Println("Error al procesar los parámetros FDISK:", err)
		return
	}

	// Leer el MBR existente
	filename := driveletter + ".dsk"
	archivoBinarioDisco, err := Fdisk.ValidateFileName(rutaDiscos, filename)
	if err != nil {
		fmt.Println(err)
		return
	}

	mbr, err := Fdisk.ReadMBR(archivoBinarioDisco)
	if err != nil {
		fmt.Println("Error leyendo el MBR:", err)
		return
	}

	// Validar el nombre de la partición
	err = Fdisk.ValidatePartitionName(&mbr, name, delete)
	if err != nil {
		fmt.Println("Error al validar el nombre de la partición:", err)
	}

	// Parametro delete
	if delete == "full" {
		Fdisk.DeletePartition(&mbr, archivoBinarioDisco, name)
	}

	// Parametro add
	if addValue > 0 || addValue < 0 {
		Fdisk.AdjustPartitionSize(&mbr, name, addValue, unit)
		return
	}

	// Validar la creación de la partición
	err = Fdisk.ValidatePartitionTypeCreation(&mbr, parttype)
	if err != nil {
		fmt.Println("Error al validar la creación de la partición:", err)
	}

	// Tamaño de la partición en bytes
	sizeInBytes, err := Fdisk.CalculateSize(size, unit)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Ajustar y crear la partición
	err = Fdisk.AdjustAndCreatePartition(&mbr, sizeInBytes, fit)
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

func rep(diskFileName string, params []string) {
	reportFileName, reportPath, err := Rep.ExtractRepParams(params)

	if err != nil {
		fmt.Println("Error al procesar los parámetros REP:", err)
	}

	// Leer el MBR existente
	mbr, err := Fdisk.ReadMBR(diskFileName)
	if err != nil {
		fmt.Println("Error leyendo el MBR:", err)
		return
	}

	dotCode := Fdisk.GenerateDotCode(&mbr)

	nombreArchivoDot := path.Join(reportPath, reportFileName+".dot")
	nombreArchivoPng := path.Join(reportPath, reportFileName+".png")

	Reportes.CrearArchivo(nombreArchivoDot)
	Reportes.EscribirArchivo(dotCode, nombreArchivoDot)
	Reportes.Ejecutar(nombreArchivoPng, nombreArchivoDot)
	// Reportes.VerReporte(nombreArchivoPng)
}
