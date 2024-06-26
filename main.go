package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	Fdisk "proyecto1/commands/fdisk"
	Mkdisk "proyecto1/commands/mkdisk"
	Mkfs "proyecto1/commands/mkfs"
	Mount "proyecto1/commands/mount"
	Rep "proyecto1/commands/rep"
	Rmdisk "proyecto1/commands/rmdisk"
	Command "proyecto1/commands/validations"
	Global "proyecto1/global"
	Reportes "proyecto1/reportes"
	Types "proyecto1/types"
	UserWorkspace "proyecto1/userworkspace"
	Utils "proyecto1/utils"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

var rutaDiscos string = "./disks/MIA/P1/"

var archivoBinarioDiscoActual string = ""
var ajusteParticionActual string = "" // first fit, best fit, worst fit

// var CurrentSession UserWorkspace.Sesion
var IsLoginFlag bool = false

var CurrentSession Global.Sesion
var PathUsersFile string = ""

func main() {
	app := fiber.New()
	app.Use(cors.New())
	app.Get("/hello", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	app.Post("/execute", func(c *fiber.Ctx) error {
		// Define a struct to match incoming JSON
		type CommandRequest struct {
			Command string `json:"command"`
		}

		var cmdReq CommandRequest
		if err := c.BodyParser(&cmdReq); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Error parsing request")
		}

		if cmdReq.Command == "" {
			return c.Status(fiber.StatusBadRequest).SendString("No command provided")
		}

		result := executeCommand(cmdReq.Command)
		return c.SendString(result)
	})

	app.Get("/list-disks", listDisksHandler)

	app.Get("/list-mounted-partitions-by-disk/:diskFileName", listMountedPartitionsByDiskHandler)

	app.Post("/login", loginHandler)

	//app.Static("/reportes", "./reportes")

	app.Get("/reports", reportsHandler)

	app.Post("/get-report", getReportHandler)

	app.Get("/get-root-directory-contents", getRootDirectoryContentsHandler)

	app.Post("/get-file-contents", getFileContentsHandler)

	app.Listen(":4000")
}

func executeCommand(input string) string {
	var resultado string = ""
	var output strings.Builder

	// comando, path := parseCommand(input)
	// if comando != "execute" || path == "" {
	// 	return "Comando no reconocido o ruta de archivo faltante. Uso: execute <ruta_al_archivo_de_scripts>"
	// }

	// path = strings.Trim(path, `"'`)
	// output.WriteString(fmt.Sprintf("Leyendo el archivo de scripts de: %s\n", path))

	// content, err := os.ReadFile(path)
	// if err != nil {
	// 	return fmt.Sprintf("Error leyendo el archivo de scripts: %v\n", err)
	// }

	//contentStr := string(content)
	contentStr := input
	contentStr = strings.Replace(contentStr, "\r\n", "\n", -1) // Convertir CRLF a LF
	commands := strings.Split(contentStr, "\n")

	for _, command := range commands {
		command = strings.TrimSpace(command)
		if command == "" || strings.HasPrefix(command, "#") {
			continue
		}

		var commandLower string = strings.ToLower(command)

		err := Command.ValidarComando(commandLower)
		if err != nil {
			fmt.Println(err)
		} else {
			err = Command.ValidarParametros(commandLower)
			if err != nil {
				fmt.Println(err)
			} else {
				switch {
				case strings.HasPrefix(commandLower, "mkdisk"):
					params := strings.Fields(command)
					archivoBinarioDiscoActual, resultado = mkdisk(params[1:])
				case strings.HasPrefix(commandLower, "fdisk"):
					params := strings.Fields(command)
					fdisk(params[1:])
				case strings.HasPrefix(commandLower, "mount"):
					params := strings.Fields(command)
					mount(params[1:])
				case strings.HasPrefix(commandLower, "mkfs"):
					params := strings.Fields(command)
					mkfs(params[1:])
				case strings.HasPrefix(commandLower, "rep"):
					params := strings.Fields(command)
					rep(params[1:])
				}
			}
		}

		// Simulated execution, replace with actual logic
		//output.WriteString(fmt.Sprintf("Procesando comando: %s\n", command))

		output.WriteString(fmt.Sprintf("%s\n", resultado))
	}

	return output.String()
}

func parseCommand(input string) (string, string) {
	parts := strings.Fields(input)
	if len(parts) < 2 {
		return "", ""
	}
	return parts[0], strings.Join(parts[1:], " ")
}

// func parseCommand(input string) (string, string) {
// 	parts := strings.Fields(input)
// 	if len(parts) < 2 {
// 		return "", ""
// 	}

// 	command := parts[0]
// 	var path string

// 	for _, part := range parts[1:] {
// 		if strings.HasPrefix(part, "-path=") {
// 			path = strings.TrimPrefix(part, "-path=")
// 			break
// 		}
// 	}

// 	return command, path
// }

func mkdisk(params []string) (string, string) {
	size, unit, diskFit, err := Mkdisk.ExtractMkdiskParams(params)
	if err != nil {
		errorMessage := fmt.Sprintf("Error al procesar los parámetros MKDISK: %v", err)
		fmt.Println(errorMessage)
		return "", errorMessage
	}

	// Tamaño del disco en bytes
	sizeInBytes, err := Mkdisk.CalculateDiskSize(size, unit)
	if err != nil {
		fmt.Println("Error:", err)
		errorMessage := fmt.Sprintf(err.Error())
		return "", errorMessage
	}

	// Construye el nombre del disco apropiado
	var filename string = Mkdisk.ConstructFileName(rutaDiscos)

	ajusteParticionActual = diskFit

	// Creación del disco con el tamaño calculado en bytes
	Mkdisk.CreateDiskWithSize(filename, int32(sizeInBytes), diskFit)

	successMessage := fmt.Sprintf("Disco %s creado exitosamente!", filename)

	fmt.Println(successMessage) // This prints the success message to the console.

	return filename, successMessage
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
	//fmt.Println("filename:", filename)
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
	err = Fdisk.ValidatePartitionName(mbr, name, delete, addValue)
	if err != nil {
		fmt.Println("Error al validar el nombre de la partición:", err)
	}

	var sizeInBytes int64 = 0
	unit = strings.ToLower(unit)
	switch unit {
	case "b":
		sizeInBytes = size
	case "k":
		sizeInBytes = size * 1024
	case "m":
		sizeInBytes = size * 1024 * 1024
	}

	if typePart != "L" {
		err = Fdisk.ValidatePartitionsSizeAgainstDiskSize(mbr, sizeInBytes)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
	}

	// Parametro delete
	if delete == "full" {
		var partitionType string = ""
		var partitionName string = ""
		for _, part := range mbr.Partitions {
			partitionName = Utils.CleanPartitionName(part.Name[:])
			if strings.TrimSpace(partitionName) == strings.TrimSpace(name) {
				partitionType = string(part.Type[:])
			}
		}

		///fmt.Println("partitionType to be deleted:", partitionType)
		if partitionType == "P" || partitionType == "E" {
			err = Fdisk.DeletePartition(mbr, archivoBinarioDisco, name)
			if err != nil {
				fmt.Println("Error al eliminar la partición:", err)
			} else {
				fmt.Println("Partición eliminada exitosamente.")
			}
		} else {
			logicalPartitions, _ := Fdisk.GetLogicalPartition(archivoBinarioDisco)
			Fdisk.PrintLogicalPartitions(logicalPartitions)

			err = Fdisk.DeleteLogicalPartition(logicalPartitions, name)
			if err != nil {
				fmt.Println("Error al eliminar la partición:", err)
			}

			///Fdisk.PrintLogicalPartitions(logicalPartitions)

		}
		return
	}

	// Parametro add
	///fmt.Println("addValue:", addValue)
	if addValue > 0 || addValue < 0 {
		err = Fdisk.AdjustPartitionSize(mbr, name, addValue, unit, archivoBinarioDisco)
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			fmt.Println("Tamaño de la partición ajustado exitosamente.")
		}
		// fmt.Println("despues del addValue negativo")
		// var TempMBR2 *Types.MBR
		// TempMBR2, err = Fdisk.ReadMBR(archivoBinarioDisco)
		// if err != nil {
		// 	fmt.Println("Error leyendo el MBR:", err)
		// 	return
		// }

		// Utils.PrintMBRv3(TempMBR2)

		return
	}

	// Validar la creación de la partición
	err = Fdisk.ValidatePartitionTypeCreation(mbr, typePart)
	if err != nil {
		fmt.Println("Error al validar la creación de la partición:", err)
		return
	}

	// Ajustar y crear la partición
	err = Fdisk.AdjustAndCreatePartition(mbr, int32(size), unit, typePart, fit, name, archivoBinarioDisco)
	if err != nil {
		fmt.Println("Error al ajustar y crear la partición:", err)
	} else {
		fmt.Println("Partición creada exitosamente.")
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

	fmt.Printf("Disco %s eliminado exitosamente!\n", filename)
}

func mount(params []string) {
	driveletter, name, err := Mount.ExtractMountParams(params)
	if err != nil {
		fmt.Println("Error al procesar los parámetros MOUNT:", err)
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
		//fmt.Println("Error leyendo el MBR:", err)
		return
	}

	_, err = Mount.MountPartition(mbr, archivoBinarioDisco, driveletter, name)
	if err != nil {
		fmt.Println("Error al montar la partición: ", err)
	} else {
		partitionId, err := Fdisk.GetPartitionId(mbr, name)
		if err != nil {
			fmt.Printf("Error al buscar la partición %s: %e", name, err)
			return
		}
		//fmt.Printf("Partición %s encontrada exitosamente antes del mount.\n", partitionId)

		fmt.Printf("Partición %s montada exitosamente.\n", partitionId)
	}

	// if result == 0 && strings.TrimSpace(partitionId) != "" {
	// 	fmt.Printf("Partición %s montada exitosamente.", partitionId)
	// }

	// else {
	// 	fmt.Println("Error al montar la partición: ", err)
	// }

	// mbr, err = Fdisk.ReadMBR(archivoBinarioDisco)
	// if err != nil {
	// 	fmt.Println("Error leyendo el MBR:", err)
	// 	return
	// }
	// Utils.PrintMounted(mbr)

}

func unmount(params []string) {
	id, err := Mount.ExtractUnmountParams(params)
	if err != nil {
		fmt.Println("Error al procesar los parámetros UNMOUNT:", err)
	}

	driveletter := string(id[0])
	filename := driveletter + ".dsk"
	//fmt.Println("filename:", filename)

	archivoBinarioDisco, err := Fdisk.ValidateFileName(rutaDiscos, filename)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("archivoBinarioDisco:", archivoBinarioDisco)

	// Leer el MBR existente
	mbr, err := Fdisk.ReadMBR(archivoBinarioDisco)
	if err != nil {
		fmt.Println("Error leyendo el MBR:", err)
		return
	}

	///fmt.Println("mbr in unmount:", mbr)
	///fmt.Println("id:", id)

	_, err = Mount.ValidatePartitionId(mbr, id)
	if err != nil {
		fmt.Println(err)
		return
	}
	/// else {
	/// 	fmt.Println("Partición encontrada.")
	/// }

	_, err = Mount.UnmountPartition(mbr, id, archivoBinarioDisco)
	if err != nil {
		fmt.Printf("Error al desmontar la partición: %s\n", err)
	} else {
		fmt.Println("Partición desmontada exitosamente.")
	}

	// var TempMBR2 *Types.MBR
	// TempMBR2, err = Fdisk.ReadMBR(archivoBinarioDisco)
	// if err != nil {
	// 	fmt.Println("Error leyendo el MBR:", err)
	// } else {
	// 	Utils.PrintMBRv3(TempMBR2)
	// 	//Utils.PrintMounted(TempMBR2)
	// }
}

func mkfs(params []string) {
	id, type_, fs, err := Mkfs.ExtractMkfsParams(params)
	if err != nil {
		fmt.Println("Error al procesar los parámetros MKFS:", err)
	}

	driveletter := string(id[0])
	filename := driveletter + ".dsk"
	//fmt.Println("filename in rep:", filename)

	archivoBinarioDisco, err := Fdisk.ValidateFileName(rutaDiscos, filename)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Leer el MBR existente
	mbr, err := Fdisk.ReadMBR(archivoBinarioDisco)
	if err != nil {
		fmt.Println("Error leyendo el MBR:", err)
		return
	}

	/// fmt.Println("mbr in mkfs:", mbr)
	/// fmt.Println("id:", id)

	_, err = Fdisk.ValidatePartitionId(mbr, id)
	if err != nil {
		fmt.Println(err)
		return
	}
	///else {
	///	fmt.Println("Partición encontrada.")
	///}

	err = Mkfs.MakeFileSystem(archivoBinarioDisco, id, type_, fs)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("Sistema de archivos %s creado exitosamente.\n", fs)

		// var partitionStart int32 = 0
		// partitionStart, err = Mount.GetPartitionStart(mbr, id)
		// if err != nil {
		// 	fmt.Println("Error getting partition start:", err)
		// 	return
		// }
		// else {
		// 	fmt.Println("Partition start:", partitionStart)
		// }

		// superblock, err := Mkfs.ReadSuperBlock(archivoBinarioDisco, partitionStart)
		// if err != nil {
		// 	fmt.Println("Error reading superblock:", err)
		// 	return
		// }

		///fmt.Println("Superblock:", superblock)

		// inodes, err := Mkfs.ReadInodesFromFile(archivoBinarioDisco, superblock)
		// if err != nil {
		// 	fmt.Println("Error reading inodes:", err)
		// 	return
		// }

		///fmt.Println("Inodes:", inodes[0])

		// directoryBlocks, err := Mkfs.ReadDirectoryBlocksFromFile(archivoBinarioDisco, superblock)
		// if err != nil {
		// 	fmt.Println("Error reading directory blocks:", err)
		// 	return
		// }

		// fmt.Println("Directory Blocks:", directoryBlocks)

	}
}

func rep(params []string) {
	//fmt.Println("rep")
	id, reportName, reportPathAndFileName, err := Rep.ExtractRepParams(params)
	reportPathAndFileName = strings.ReplaceAll(reportPathAndFileName, "\"", "")
	//fmt.Printf("id: %s, reportName: %s, reportPathAndFileName: %s\n", id, reportName, reportPathAndFileName)

	if err != nil {
		fmt.Println("Error al procesar los parámetros REP:", err)
	}

	reportPathAndFileName = filepath.Clean(reportPathAndFileName)

	driveletter := string(id[0])
	filename := driveletter + ".dsk"
	//fmt.Println("filename in rep:", filename)

	archivoBinarioDisco, err := Fdisk.ValidateFileName(rutaDiscos, filename)
	if err != nil {
		fmt.Println(err)
		return
	}

	//fmt.Println("archivoBinarioDisco:", archivoBinarioDisco)

	// Leer el MBR existente
	mbr, err := Fdisk.ReadMBR(archivoBinarioDisco)
	if err != nil {
		fmt.Println("Error leyendo el MBR:", err)
		return
	}

	var dotCode string
	switch reportName {
	case "mbr":
		//fmt.Printf("Identificador: [%s]\n", id)
		_, err = Fdisk.ValidatePartitionId(mbr, id)
		if err != nil {
			fmt.Println(err)
			return
		} else {
			//fmt.Println("Generated Mbr report")

			dotCode = Fdisk.GenerateDotCodeMbr(mbr, archivoBinarioDisco)
		}

	case "disk":
		//fmt.Printf("Identificador: [%s]\n", id)

		// fmt.Println("mbr in rep antes de generar el reporte disk:", mbr)
		// var TempMBR2 *Types.MBR
		// TempMBR2, err = Fdisk.ReadMBR(archivoBinarioDisco)
		// if err != nil {
		// 	fmt.Println("Error leyendo el MBR:", err)
		// 	return
		// }

		// Utils.PrintMBRv3(TempMBR2)

		_, err = Fdisk.ValidatePartitionId(mbr, id)
		if err != nil {
			fmt.Println(err)
			return
		} else {
			//fmt.Println("Generated Disk report")
			dotCode, err = Fdisk.GenerateDotCodeDisk(mbr, archivoBinarioDisco)
			if err != nil {
				fmt.Println("Error generating disk report:", err)
			}
			// else {
			// 	fmt.Println(dotCode)
			// }
		}
	case "tree":
		// Leer el MBR existente
		mbr, err := Fdisk.ReadMBR(archivoBinarioDisco)
		if err != nil {
			fmt.Println("Error leyendo el MBR:", err)
			return
		}

		/// fmt.Println("mbr in rep:", mbr)
		/// fmt.Println("id:", id)

		_, err = Mount.ValidatePartitionId(mbr, id)
		if err != nil {
			fmt.Println(err)
			return
		}
		// else {
		// 	fmt.Println("Partición encontrada.")
		// }

		var partitionStart int32 = 0
		partitionStart, err = Mount.GetPartitionStart(mbr, id)

		if err != nil {
			fmt.Println(err)
			return
		}
		// else {
		// 	fmt.Println("Start:", partitionStart)
		// }

		superblock, err := Mkfs.ReadSuperBlock(archivoBinarioDisco, partitionStart)
		if err != nil {
			fmt.Println("Error reading superblock:", err)
			return
		}

		//fmt.Println("Superblock in rep:", superblock)

		inodes, err := Mkfs.ReadAllUsedInodesFromFile(archivoBinarioDisco, superblock)
		if err != nil {
			fmt.Println("Error reading inodes:", err)
			return
		}

		//fmt.Println("Inodes len in rep:", len(inodes))

		// for i, inode := range inodes {
		// 	fmt.Println("Inode:", i, "Inode:", inode)
		// 	if i > 10 {
		// 		break
		// 	}
		// }

		//fmt.Println("Inodes in rep:", inodes[0])

		// entries, err := Mkfs.ReadBlock0AndTraverseContents(archivoBinarioDisco, superblock)
		// if err != nil {
		// 	fmt.Println("Error traversing Block 0:", err)
		// 	return
		// }

		// // Example: Print the names of the entries in Block 0
		// for _, entry := range entries {
		// 	fmt.Printf("Entry Name: %s, Inode: %d\n", string(entry.B_name[:]), entry.B_inodo)
		// }

		dotCode, err = Mkfs.GraficarArbol(archivoBinarioDisco, int(partitionStart), inodes)
		//dotCode = Mkfs.GenerateDotCodeTree(inodes, directoryBlocks)

		//dotCode, err = Mkfs.GraficarTREE(archivoBinarioDisco, int(partitionStart))
		if err != nil {
			fmt.Println("Error generating tree report:", err)
			return
		} else {
			fmt.Println("Reporte Tree generado con exito!")
		}
	case "sb":
		// Leer el MBR existente
		mbr, err := Fdisk.ReadMBR(archivoBinarioDisco)
		if err != nil {
			fmt.Println("Error leyendo el MBR:", err)
			return
		}

		/// fmt.Println("mbr in rep:", mbr)
		/// fmt.Println("id:", id)

		_, err = Mount.ValidatePartitionId(mbr, id)
		if err != nil {
			fmt.Println(err)
			return
		}
		// else {
		// 	fmt.Println("Partición encontrada.")
		// }

		var partitionStart int32 = 0
		partitionStart, err = Mount.GetPartitionStart(mbr, id)

		if err != nil {
			fmt.Println(err)
			return
		}
		// else {
		// 	fmt.Println("Start:", partitionStart)
		// }

		superblock, err := Mkfs.ReadSuperBlock(archivoBinarioDisco, partitionStart)
		if err != nil {
			fmt.Println("Error reading superblock:", err)
			return
		}

		//fmt.Println("Superblock in rep:", superblock)

		dotCode, err = Mkfs.GraficarSB(&superblock)

		if err != nil {
			fmt.Println("Error generating sb report:", err)
			return
		} else {
			fmt.Println("Reporte SB generado con exito!")
		}

	case "inode":
		// Leer el MBR existente
		mbr, err := Fdisk.ReadMBR(archivoBinarioDisco)
		if err != nil {
			fmt.Println("Error leyendo el MBR:", err)
			return
		}

		/// fmt.Println("mbr in rep:", mbr)
		/// fmt.Println("id:", id)

		_, err = Mount.ValidatePartitionId(mbr, id)
		if err != nil {
			fmt.Println(err)
			return
		}
		// else {
		// 	fmt.Println("Partición encontrada.")
		// }

		var partitionStart int32 = 0
		partitionStart, err = Mount.GetPartitionStart(mbr, id)

		if err != nil {
			fmt.Println(err)
			return
		}
		// else {
		// 	fmt.Println("Start:", partitionStart)
		// }

		superblock, err := Mkfs.ReadSuperBlock(archivoBinarioDisco, partitionStart)
		if err != nil {
			fmt.Println("Error reading superblock:", err)
			return
		}

		//fmt.Println("Superblock in rep:", superblock)

		allUsedInodes, err := Mkfs.ReadAllUsedInodesFromFile(archivoBinarioDisco, superblock)
		if err != nil {
			fmt.Println("Error reading inodes:", err)
			return
		}

		dotCode, err = Mkfs.GraficarInodos(allUsedInodes)

		if err != nil {
			fmt.Println("Error generating inode report:", err)
			return
		} else {
			fmt.Println("Reporte Inode generado con exito!")
		}
	}

	//fmt.Println("Dot Code:", dotCode)

	fmt.Printf("reportPathAndFileName: %s\n", reportPathAndFileName)

	dir, file := filepath.Split(reportPathAndFileName)
	extension := filepath.Ext(file)
	baseName := file[:len(file)-len(extension)]

	//extension := filepath.Ext(reportPathAndFileName)
	//fmt.Printf("extension: %s\n", extension)
	//pathWithoutExt := reportPathAndFileName[:len(reportPathAndFileName)-len(extension)]
	//fmt.Printf("pathWithoutExt: %s\n", pathWithoutExt)

	nombreArchivoDot := filepath.Join(dir, id+"_"+baseName+".dot")
	nombreArchivoReporte := filepath.Join(dir, id+"_"+baseName+extension)
	//fmt.Printf("nombreArchivoReporte: %s\n", nombreArchivoReporte)
	// switch extension {
	// case ".pdf":
	// 	nombreArchivoReporte = pathWithoutExt + ".pdf"
	// case ".txt":
	// 	nombreArchivoReporte = pathWithoutExt + ".txt"
	// case ".png":
	// 	nombreArchivoReporte = pathWithoutExt + ".png"
	// case ".jpg":
	// 	nombreArchivoReporte = pathWithoutExt + ".jpg"
	// default:
	// 	nombreArchivoReporte = reportPathAndFileName
	// }

	Reportes.CrearArchivo(nombreArchivoDot)
	Reportes.EscribirArchivo(dotCode, nombreArchivoDot)
	//fmt.Printf("extension: %s\n", extension)
	Reportes.Ejecutar(nombreArchivoReporte, nombreArchivoDot, extension)
	// Reportes.VerReporte(nombreArchivoPng)
}

func login(params []string) {
	user, pass, id, err := UserWorkspace.ExtractLoginParams(params)
	if err != nil {
		fmt.Println(err)
		return
	}

	driveletter := string(id[0])
	filename := driveletter + ".dsk"
	//fmt.Println("filename in rep:", filename)

	archivoBinarioDisco, err := Fdisk.ValidateFileName(rutaDiscos, filename)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = UserWorkspace.Login(user, pass, id, archivoBinarioDisco)
	if err != nil {
		fmt.Println(err)
		return
	} else {
		if Global.Usuario.Status {
			IsLoginFlag = true
			fmt.Println("Login exitoso")
		} else {
			fmt.Println("Login fallido")
		}
	}
}

func logout() {
	if Global.Usuario.Status == false {
		fmt.Println("No hay ninguna sesión activa")
		return
	}
	Global.Usuario.Status = false
	Global.Usuario.Id = ""
	Global.SesionActual = Global.Sesion{}

	fmt.Println("Logout exitoso")
}

func mkgrp(params []string) {
	name, err := UserWorkspace.ExtractMkgrpParams(params)
	if err != nil {
		fmt.Println(err)
		return
	}

	// fmt.Printf("Global.SesionActual in mkgrp: %v\n", Global.SesionActual)

	// name, err := UserWorkspace.ExtractMkgrpParams(params)
	// if err != nil {
	// 	fmt.Println("Error al procesar los parámetros MKFS:", err)
	// }

	// filename := "A.dsk"
	// //fmt.Println("filename in rep:", filename)

	// archivoBinarioDisco, err := Fdisk.ValidateFileName(rutaDiscos, filename)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// _, err = UserWorkspace.EjecutarMkgrp(name)
	// if err != nil {
	// 	fmt.Println(err)
	// } else {
	// 	fmt.Println("mkgrp exitoso")
	// }

	// var filePath string = ".\\users.txt"
	// var partitionId string = "A123"

	// fileContents, err := UserWorkspace.ReturnFileContents(filePath, partitionId, archivoBinarioDisco)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// } else {
	// 	fmt.Println("fileContents:", fileContents)
	// }

	_, err = UserWorkspace.EjecutarMkgrp(name, PathUsersFile, Global.SesionActual)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("mkgrp exitoso")
	}
}

func cat(params []string) {
	file, err := UserWorkspace.ExtractCatParams(params)
	if err != nil {
		fmt.Println(err)
		return
	}

	var fileContents []string = []string{}
	fileContents, err = UserWorkspace.EjecutarCat(file, Global.SesionActual.PartitionId, Global.SesionActual.Path)

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Contenido del archivo:\n", fileContents)
	}

}

func listDisksHandler(c *fiber.Ctx) error {
	disks, err := getDiskFiles(rutaDiscos)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to read directory")
	}
	return c.JSON(disks)

}

func getDiskFiles(directoryPath string) ([]string, error) {
	files, err := os.ReadDir(directoryPath)
	if err != nil {
		return nil, err
	}

	var disks []string
	for _, file := range files {
		info, err := file.Info()
		if err != nil {
			continue // or handle the error in a way that fits your requirements
		}
		if strings.HasSuffix(file.Name(), ".dsk") && !info.IsDir() {
			disks = append(disks, file.Name())
		}
	}
	return disks, nil
}

func listMountedPartitionsByDiskHandler(c *fiber.Ctx) error {
	partitions, err := getMountedPartitionsHandler(c.Params("diskFileName"))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to get mounted partitions")
	}
	return c.JSON(partitions)
}

func getMountedPartitionsHandler(diskFileName string) ([]Types.DiskPartitionDto, error) {
	var diskFileNameFullPath string = rutaDiscos + diskFileName
	partitions, err := Mkfs.GetMountedPartitionsByDisk(diskFileNameFullPath)
	if err != nil {
		return nil, err
	}
	return partitions, nil
}

func loginHandler(c *fiber.Ctx) error {
	var loginRequest Types.LoginRequestDto
	if err := c.BodyParser(&loginRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Error parsing request")
	}
	if loginRequest.Username == "" || loginRequest.Password == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Username and password are required")
	}
	fmt.Printf("loginRequest: %+v\n", loginRequest)
	err := executeLogin(loginRequest.Username, loginRequest.Password, loginRequest.PartitionId)
	if err != nil && err.Error() != "Usuario ya conectado" {
		return c.Status(fiber.StatusBadRequest).SendString("Error logging in")
	}
	return c.SendString("Login successful")
}

func executeLogin(user string, pass string, partitionId string) error {

	driveletter := string(partitionId[0])
	filename := driveletter + ".dsk"
	//fmt.Println("filename in rep:", filename)

	archivoBinarioDisco, err := Fdisk.ValidateFileName(rutaDiscos, filename)
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = UserWorkspace.Login(user, pass, partitionId, archivoBinarioDisco)
	if err != nil {
		fmt.Println(err)
		return err
	} else {
		if Global.Usuario.Status {
			IsLoginFlag = true
			fmt.Println("Login exitoso")
		} else {
			fmt.Println("Login fallido")
		}
	}

	return nil
}

func reportsHandler(c *fiber.Ctx) error {
	//partitionId := c.Params("partitionId")
	partitionId := c.Query("partitionId")
	reports, err := listReports("./reportes", partitionId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to read reports directory")
	}
	return c.JSON(reports)
	//return c.SendString("Reports")
}

func listReports(directory string, partitionId string) ([]Types.ReportDto, error) {
	fmt.Println("directory:", directory)
	fmt.Println("partitionId:", partitionId)
	var reports []Types.ReportDto
	files, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		fmt.Println("Checking file:", file.Name())
		if !file.IsDir() && strings.HasPrefix(strings.ToLower(file.Name()), strings.ToLower(partitionId+"_")) && (strings.HasSuffix(strings.ToLower(file.Name()), ".dot") || strings.HasSuffix(strings.ToLower(file.Name()), ".jpg")) {
			parts := strings.Split(file.Name(), "_")
			result := parts[len(parts)-1]
			result = strings.TrimSuffix(result, ".dot")
			report := Types.ReportDto{
				ReportFileName: result,
				DotFileName:    file.Name()[:len(file.Name())-4] + ".dot",
			}
			reports = append(reports, report)
		}
	}
	return reports, nil
}

func getReportHandler(c *fiber.Ctx) error {
	var getReportRequest Types.GetReportDto
	if err := c.BodyParser(&getReportRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Error parsing request")
	}
	var reportPath string = "./reportes/" + getReportRequest.DotFileName
	if _, err := os.Stat(reportPath); os.IsNotExist(err) {
		return c.Status(fiber.StatusNotFound).SendString("Report not found")
	}
	file, err := os.Open(reportPath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to open report file")
	}
	defer file.Close()

	dotContent, err := io.ReadAll(file)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to read report file"})
	}

	// Sending the content of the DOT file as a JSON object
	return c.JSON(fiber.Map{"dotCode": string(dotContent)})
}

func getRootDirectoryContentsHandler(c *fiber.Ctx) error {
	partitionId := c.Query("partitionId")
	rootDirectoryContents, err := getRootDirectoryContents(partitionId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to get root directory contents")
	}

	// Send the structured response to the frontend
	response := Types.FileExplorerResponse{Items: rootDirectoryContents}
	return c.JSON(response)
}

func getRootDirectoryContents(partitionId string) ([]Types.FileExplorerItem, error) {

	driveletter := string(partitionId[0])
	filename := driveletter + ".dsk"
	//fmt.Println("filename in rep:", filename)

	archivoBinarioDisco, err := Fdisk.ValidateFileName(rutaDiscos, filename)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	// Leer el MBR existente
	mbr, err := Fdisk.ReadMBR(archivoBinarioDisco)
	if err != nil {
		fmt.Println("Error leyendo el MBR:", err)
		return nil, err
	}

	_, err = Mount.ValidatePartitionId(mbr, partitionId)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	var partitionStart int32 = 0
	partitionStart, err = Mount.GetPartitionStart(mbr, partitionId)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	superblock, err := Mkfs.ReadSuperBlock(archivoBinarioDisco, partitionStart)
	if err != nil {
		fmt.Println("Error reading superblock:", err)
		return nil, err
	}

	inodes, err := Mkfs.ReadAllUsedInodesFromFile(archivoBinarioDisco, superblock)
	if err != nil {
		fmt.Println("Error reading inodes:", err)
		return nil, err
	}

	var blocks []interface{}
	//var directoryBlocks []Types.DirectoryBlock
	var error2 error

	_, blocks, error2 = Mkfs.ReadFileSystemStructures(archivoBinarioDisco, partitionStart, inodes)
	if error2 != nil {
		fmt.Println(error2)
		return nil, error2
	}

	var items []Types.FileExplorerItem

	rootInode := inodes[0]

	for _, blockNum := range rootInode.I_block {
		if blockNum != -1 { // Check for valid block number
			block := blocks[blockNum]
			switch b := block.(type) {
			case Types.DirectoryBlock:
				for _, content := range b.B_content {
					name := Utils.CleanPartitionName(content.B_name[:]) // Convert name from bytes to string
					if name == "" {
						continue
					}
					inodeIndex := content.B_inodo
					isFolder := inodes[inodeIndex].I_type[0] == 0 // '0' denotes a directory
					items = append(items, Types.FileExplorerItem{Name: name, Inode: inodeIndex, IsFolder: isFolder})
				}
			case Types.FileBlock:
				// Assuming we have metadata in blocks for files (adjust if necessary)
				content := Utils.CleanPartitionName(b.B_content[:]) // Simplified representation
				if content != "" {
					items = append(items, Types.FileExplorerItem{
						Name:     content,
						Inode:    blockNum, // Use block index as an inode index placeholder
						IsFolder: false,
					})
				}
			}
		}
	}

	fmt.Println("Items:", items)
	return items, nil
}

func getFileContentsHandler(c *fiber.Ctx) error {
	var getFileContentsRequest Types.GetFileContentsRequest
	if err := c.BodyParser(&getFileContentsRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Error parsing request")
	}
	fmt.Printf("getFileContentsRequest: %+v\n", getFileContentsRequest)
	fileContents, err := getFileContents(getFileContentsRequest)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to get file contents")
	}
	return c.JSON(fileContents)
}

func getFileContents(getFileContentsRequest Types.GetFileContentsRequest) (Types.FileContentsResponse, error) {

	driveletter := string(getFileContentsRequest.PartitionId[0])
	filename := driveletter + ".dsk"
	archivoBinarioDisco, err := Fdisk.ValidateFileName(rutaDiscos, filename)
	if err != nil {
		fmt.Println(err)
		return Types.FileContentsResponse{}, err
	}

	var fileContents Types.FileContentsResponse
	var error2 error
	fileContents.Contents = ""

	var contents []string = []string{}
	fmt.Println("getFileContentsRequest.Path:", getFileContentsRequest.Path)
	fmt.Println("getFileContentsRequest.PartitionId:", getFileContentsRequest.PartitionId)
	fmt.Println("archivoBinarioDisco:", archivoBinarioDisco)
	contents, error2 = UserWorkspace.EjecutarCat(normalizePath(getFileContentsRequest.Path), getFileContentsRequest.PartitionId, archivoBinarioDisco)
	if error2 != nil {
		fmt.Println(error2)
		return Types.FileContentsResponse{}, error2
	}
	contentsString := strings.Join(contents, "\n")
	contents = []string{Utils.CleanPartitionName([]byte(contentsString))}
	fileContents.Contents = strings.Join(contents, "\n")

	return fileContents, nil
}

func normalizePath(path string) string {
	return strings.ReplaceAll(path, "\\", "/") // Convert all backslashes to forward slashes
}
