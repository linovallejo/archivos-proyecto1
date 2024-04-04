package main

import (
	"bufio"
	"fmt"
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
	UserWorkspace "proyecto1/userworkspace"
	Utils "proyecto1/utils"
	"runtime"
	"strings"
)

var rutaDiscos string = "./disks/MIA/P1/"

var archivoBinarioDiscoActual string = ""
var ajusteParticionActual string = "" // first fit, best fit, worst fit

// var CurrentSession UserWorkspace.Sesion
var IsLoginFlag bool = false

var CurrentSession Global.Sesion
var PathUsersFile string = ""

func main() {
	Utils.LimpiarConsola()
	Utils.PrintCopyright()
	fmt.Println("Sistema de Archivos ext2/ext3 - Proyecto 1")

	if runtime.GOOS == "windows" {
		PathUsersFile = ".\\users.txt"
	} else {
		PathUsersFile = "./users.txt"
		// PathUsersFile = filepath.Join("/home", "lino", "users.txt")
	}

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
		command = strings.TrimSpace(command)
		if command != "" {
			fmt.Println("_______________________________________________________")
			fmt.Println("Procesando comando: ", command)
		}
		if command == "" || strings.HasPrefix(command, "#") {
			continue
		}

		var commandLower string = strings.ToLower(command)

		err = Command.ValidarComando(commandLower)
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
					archivoBinarioDiscoActual = mkdisk(params[1:])

					// fmt.Println("despues del mkdisk")
					// var TempMBR2 *Types.MBR
					// TempMBR2, err = Fdisk.ReadMBR(archivoBinarioDiscoActual)
					// if err != nil {
					// 	fmt.Println("Error leyendo el MBR:", err)
					// 	return
					// }

					// Utils.PrintMBRv3(TempMBR2)
				case strings.HasPrefix(commandLower, "fdisk"):
					params := strings.Fields(command)

					// fmt.Println("antes del fdisk")
					// var TempMBR2 *Types.MBR
					// TempMBR2, err = Fdisk.ReadMBR(archivoBinarioDiscoActual)
					// if err != nil {
					// 	fmt.Println("Error leyendo el MBR:", err)
					// 	return
					// }
					// Utils.PrintMBRv3(TempMBR2)

					fdisk(params[1:])

					// fmt.Println("despues del fdisk")
					// var TempMBR3 *Types.MBR
					// TempMBR3, err = Fdisk.ReadMBR(archivoBinarioDiscoActual)
					// if err != nil {
					// 	fmt.Println("Error leyendo el MBR:", err)
					// 	return
					// }
					// Utils.PrintMBRv3(TempMBR3)

					// logicalPartitions, _ := Fdisk.GetLogicalPartition(archivoBinarioDiscoActual)
					// Fdisk.PrintLogicalPartitions(logicalPartitions)

				case strings.HasPrefix(commandLower, "rmdisk"):
					// fmt.Println("¿Está seguro de que desea eliminar el disco? [s/N]:")
					// var response string
					// _, err := fmt.Scanln(&response)
					// if err != nil || (response != "s" && response != "S") {
					// 	fmt.Println("Operación de eliminación cancelada.")
					// 	return
					// }

					params := strings.Fields(command)
					rmdisk(params[1:])
				case strings.HasPrefix(commandLower, "mount"):
					params := strings.Fields(command)

					// fmt.Println("antes del mount")
					// var TempMBR2 *Types.MBR
					// TempMBR2, err = Fdisk.ReadMBR(archivoBinarioDiscoActual)
					// if err != nil {
					// 	fmt.Println("Error leyendo el MBR:", err)
					// 	return
					// }
					// Utils.PrintMBRv3(TempMBR2)

					mount(params[1:])

					// fmt.Println("despues del mount")
					// var TempMBR3 *Types.MBR
					// TempMBR3, err = Fdisk.ReadMBR(archivoBinarioDiscoActual)
					// if err != nil {
					// 	fmt.Println("Error leyendo el MBR:", err)
					// 	return
					// }
					// Utils.PrintMBRv3(TempMBR3)
				case strings.HasPrefix(commandLower, "unmount"):
					params := strings.Fields(command)
					unmount(params[1:])
				case strings.HasPrefix(commandLower, "rep"):
					params := strings.Fields(command)
					rep(params[1:])
				case strings.HasPrefix(commandLower, "pause"):
					fmt.Println("Presione cualquier tecla para continuar...")
					fmt.Scanln()
				case strings.HasPrefix(commandLower, "login"):
					params := strings.Fields(command)
					login(params[1:])
				// case strings.HasPrefix(commandLower, "logout"):
				// 	logout()
				case strings.HasPrefix(commandLower, "mkfs"):
					params := strings.Fields(command)

					// fmt.Println("antes del mkfs")
					// var TempMBR2 *Types.MBR
					// TempMBR2, err = Fdisk.ReadMBR(archivoBinarioDiscoActual)
					// if err != nil {
					// 	fmt.Println("Error leyendo el MBR:", err)
					// 	return
					// }
					// Utils.PrintMBRv3(TempMBR2)

					mkfs(params[1:])

					// fmt.Println("despues del mkfs")
					// var TempMBR3 *Types.MBR
					// TempMBR3, err = Fdisk.ReadMBR(archivoBinarioDiscoActual)
					// if err != nil {
					// 	fmt.Println("Error leyendo el MBR:", err)
					// 	return
					// }
					// Utils.PrintMBRv3(TempMBR3)
				case strings.HasPrefix(commandLower, "mkgrp"):
					params := strings.Fields(command)
					mkgrp(params[1:])
				}
			}

		}

	}
	//Rmdisk.RemoveDisk(archivoBinarioDiscoActual)
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

	fmt.Printf("Disco %s creado exitosamente!\n", filename)

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
	id, reportName, reportPathAndFileName, err := Rep.ExtractRepParams(params)
	reportPathAndFileName = strings.ReplaceAll(reportPathAndFileName, "\"", "")
	//fmt.Printf("id: %s, reportName: %s, reportPathAndFileName: %s\n", id, reportName, reportPathAndFileName)

	if err != nil {
		fmt.Println("Error al procesar los parámetros REP:", err)
	}

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

		fmt.Println("Superblock in rep:", superblock)

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

	//fmt.Printf("reportPathAndFileName: %s\n", reportPathAndFileName)

	extension := filepath.Ext(reportPathAndFileName)
	//fmt.Printf("extension: %s\n", extension)
	pathWithoutExt := reportPathAndFileName[:len(reportPathAndFileName)-len(extension)]
	//fmt.Printf("pathWithoutExt: %s\n", pathWithoutExt)

	nombreArchivoDot := pathWithoutExt + ".dot"
	nombreArchivoReporte := reportPathAndFileName
	//fmt.Printf("nombreArchivoReporte: %s\n", nombreArchivoReporte)
	switch extension {
	case ".pdf":
		nombreArchivoReporte = pathWithoutExt + ".pdf"
	case ".txt":
		nombreArchivoReporte = pathWithoutExt + ".txt"
	case ".png":
		nombreArchivoReporte = pathWithoutExt + ".png"
	case ".jpg":
		nombreArchivoReporte = pathWithoutExt + ".jpg"
	default:
		nombreArchivoReporte = reportPathAndFileName
	}

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

// func logout() {
// 	err := UserWorkspace.Logout()
// 	if err != nil {
// 		fmt.Println(err)
// 	} else {
// 		fmt.Println("Logout exitoso")
// 	}
// }

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
