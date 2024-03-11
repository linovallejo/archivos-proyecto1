package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	Fdisk "proyecto1/commands/fdisk"
	Mkdisk "proyecto1/commands/mkdisk"
	Mount "proyecto1/commands/mount"
	Rep "proyecto1/commands/rep"
	Rmdisk "proyecto1/commands/rmdisk"
	Reportes "proyecto1/reportes"
	Types "proyecto1/types"
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
		switch {
		case strings.HasPrefix(command, "mkdisk"):
			params := strings.Fields(command)
			archivoBinarioDiscoActual = mkdisk(params[1:])

			fmt.Println("despues del mkdisk")
			var TempMBR2 *Types.MBR
			TempMBR2, err = Fdisk.ReadMBR(archivoBinarioDiscoActual)
			if err != nil {
				fmt.Println("Error leyendo el MBR:", err)
				return
			}

			Utils.PrintMBRv3(TempMBR2)
		case strings.HasPrefix(command, "fdisk"):
			params := strings.Fields(command)

			fmt.Println("antes del fdisk")
			var TempMBR2 *Types.MBR
			TempMBR2, err = Fdisk.ReadMBR(archivoBinarioDiscoActual)
			if err != nil {
				fmt.Println("Error leyendo el MBR:", err)
				return
			}
			Utils.PrintMBRv3(TempMBR2)

			fdisk(params[1:])

			fmt.Println("despues del fdisk")
			var TempMBR3 *Types.MBR
			TempMBR3, err = Fdisk.ReadMBR(archivoBinarioDiscoActual)
			if err != nil {
				fmt.Println("Error leyendo el MBR:", err)
				return
			}
			Utils.PrintMBRv3(TempMBR3)
		case strings.HasPrefix(command, "rmdisk"):
			params := strings.Fields(command)
			rmdisk(params[1:])
		case strings.HasPrefix(command, "mount"):
			params := strings.Fields(command)

			fmt.Println("antes del mount")
			var TempMBR2 *Types.MBR
			TempMBR2, err = Fdisk.ReadMBR(archivoBinarioDiscoActual)
			if err != nil {
				fmt.Println("Error leyendo el MBR:", err)
				return
			}
			Utils.PrintMBRv3(TempMBR2)

			mount(params[1:])

			fmt.Println("despues del mount")
			var TempMBR3 *Types.MBR
			TempMBR3, err = Fdisk.ReadMBR(archivoBinarioDiscoActual)
			if err != nil {
				fmt.Println("Error leyendo el MBR:", err)
				return
			}
			Utils.PrintMBRv3(TempMBR3)
		case strings.HasPrefix(command, "unmount"):
			params := strings.Fields(command)
			unmount(params[1:])
		case strings.HasPrefix(command, "rep"):
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
	Mkdisk.CreateDiskWithSize(filename, int32(sizeInBytes), diskFit)

	fmt.Println("Disco creado con éxito!")

	return filename
}

func fdisk(params []string) {
	size, driveletter, name, unit, typePart, fit, delete, addValue, err := Fdisk.ExtractFdiskParams(params)
	//fmt.Println("size:", size, "driveletter:", driveletter, "name:", name, "unit:", unit, "typePart:", typePart, "fit:", fit, "delete:", delete, "addValue:", addValue)

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
	err = Fdisk.ValidatePartitionName(mbr, name, delete)
	if err != nil {
		fmt.Println("Error al validar el nombre de la partición:", err)
	}

	// Parametro delete
	if delete == "full" {
		Fdisk.DeletePartition(mbr, archivoBinarioDisco, name)
	}

	// Parametro add
	if addValue > 0 || addValue < 0 {
		Fdisk.AdjustPartitionSize(mbr, name, addValue, unit)
		return
	}

	// Validar la creación de la partición
	err = Fdisk.ValidatePartitionTypeCreation(mbr, typePart)
	if err != nil {
		fmt.Println("Error al validar la creación de la partición:", err)
	}

	// Tamaño de la partición en bytes
	// sizeInBytes, err := Fdisk.CalculateSize(size, unit)
	// if err != nil {
	// 	fmt.Println("Error:", err)
	// 	return
	// }

	// Ajustar y crear la partición
	err = Fdisk.AdjustAndCreatePartition(mbr, int32(size), unit, typePart, fit, name, archivoBinarioDiscoActual)
	if err != nil {
		fmt.Println("Error al ajustar y crear la partición:", err)
	} else {
		fmt.Println("main.go - Partición creada exitosamente.")
	}
}

func rmdisk(params []string) {
	driveletter, err := Rmdisk.ExtractRmdiskParams(params)
	if err != nil {
		fmt.Println("Error al procesar los parámetros RMDISK:", err)
	}

	// Leer el MBR existente
	filename := driveletter + ".dsk"
	archivoBinarioDisco, err := Fdisk.ValidateFileName(rutaDiscos, filename)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = Rmdisk.RemoveDisk(archivoBinarioDisco)
	if err != nil {
		fmt.Println(err)
	}
}

func mount(params []string) {
	driveletter, name, err := Mount.ExtractMountParams(params)
	if err != nil {
		fmt.Println("Error al procesar los parámetros MOUNT:", err)
	}
	var result int

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

	result, err = Mount.MountPartition(mbr, archivoBinarioDisco, driveletter, name)
	if err != nil {
		fmt.Println(err)
	}

	if result == 0 {
		fmt.Println("Partición montada exitosamente.")
	} else {
		fmt.Println("Error al montar la partición.")
	}

}

func unmount(params []string) {
	id, err := Mount.ExtractUnmountParams(params)
	if err != nil {
		fmt.Println("Error al procesar los parámetros UNMOUNT:", err)
	}

	Mount.UnmountPartition(id)
}

func rep(diskFileName string, params []string) {
	// Leer el MBR existente
	fmt.Println("diskFileName:", diskFileName)
	mbr, err := Fdisk.ReadMBR(diskFileName)
	if err != nil {
		fmt.Println("Error leyendo el MBR:", err)
		return
	}
	Utils.LineaDoble(60)
	fmt.Println("mbr in rep:", mbr)
	for i, p := range mbr.Partitions {
		fmt.Printf("Partición %d: %+v\n", i+1, p)
	}

	reportName, reportPathAndFileName, err := Rep.ExtractRepParams(params)

	if err != nil {
		fmt.Println("Error al procesar los parámetros REP:", err)
	}

	// Leer el MBR existente
	// mbr, err := Fdisk.ReadMBR(diskFileName)
	// if err != nil {
	// 	fmt.Println("Error leyendo el MBR:", err)
	// 	return
	// }

	var dotCode string
	switch reportName {
	case "mbr":
		dotCode = Fdisk.GenerateDotCodeMbr(mbr)
	case "disk":
		dotCode = Fdisk.GenerateDotCodeDisk(mbr)
	}

	extension := filepath.Ext(reportPathAndFileName)
	pathWithoutExt := reportPathAndFileName[:len(reportPathAndFileName)-len(extension)]

	nombreArchivoDot := pathWithoutExt + ".dot"
	nombreArchivoReporte := reportPathAndFileName
	switch extension {
	case ".pdf":
		nombreArchivoReporte = pathWithoutExt + ".pdf"
	case ".txt":
		nombreArchivoReporte = pathWithoutExt + ".txt"
	case ".png":
		nombreArchivoReporte = pathWithoutExt + ".png"
	default:
		nombreArchivoReporte = reportPathAndFileName
	}

	Reportes.CrearArchivo(nombreArchivoDot)
	Reportes.EscribirArchivo(dotCode, nombreArchivoDot)
	Reportes.Ejecutar(nombreArchivoReporte, nombreArchivoDot, extension)
	// Reportes.VerReporte(nombreArchivoPng)
}
