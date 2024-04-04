package userworkspace

import (
	"encoding/binary"
	"fmt"
	Fdisk "proyecto1/commands/fdisk"
	Global "proyecto1/global"
	Types "proyecto1/types"
	Utils "proyecto1/utils"
	"strconv"
	"strings"
)

// type Sesion struct {
// 	Id_user     int
// 	Id_grp      int
// 	Start_SB    int
// 	System_type int
// 	User_name   string
// 	Path        string
// 	Fit         [1]byte
// }

// var CurrentSession Sesion
// var IsLoginFlag bool = false

func ExtractLoginParams(params []string) (string, string, string, error) {
	var user string = ""
	var pass string = ""
	var id string = ""

	if len(params) == 0 {
		return "", "", "", fmt.Errorf("No se encontraron parámetros")
	}
	var parametrosObligatoriosOk bool = false
	userOk := false
	passOk := false
	idOk := false
	for _, param1 := range params {
		if strings.HasPrefix(param1, "-id=") {
			idOk = true
		} else if strings.HasPrefix(param1, "-user=") {
			userOk = true
		} else if strings.HasPrefix(param1, "-pass=") {
			passOk = true
		}
	}

	parametrosObligatoriosOk = idOk && userOk && passOk

	if !parametrosObligatoriosOk {
		return "", "", "", fmt.Errorf("No se encontraron parámetros obligatorios")
	}

	for _, param := range params {
		switch {
		case strings.HasPrefix(param, "-id="):
			id = strings.TrimPrefix(param, "-id=")
			// Validar el id de la partición
			// TODO
		case strings.HasPrefix(param, "-user="):
			user = strings.TrimPrefix(param, "-user=")
		case strings.HasPrefix(param, "-pass="):
			pass = strings.TrimPrefix(param, "-pass=")
		}
	}

	return user, pass, id, nil
}

func Login(user string, pass string, id string, diskFileName string) error {
	// fmt.Println("======Start LOGIN======")
	// fmt.Println("User:", user)
	// fmt.Println("Pass:", pass)
	// fmt.Println("Id:", id)

	if Global.Usuario.Status {
		//fmt.Println("User already logged in")
		return fmt.Errorf("Usuario ya conectado")
	}

	var login bool = false

	var TempMBR *Types.MBR
	// Leer el MBR existente
	TempMBR, err := Fdisk.ReadMBR(diskFileName)
	if err != nil {
		//fmt.Println("Error leyendo el MBR:", err)
		return fmt.Errorf("Error leyendo el MBR")
	}

	var partitionStatusStr string = ""
	var index int = -1
	var partitionStart int32 = 0
	var partitionFit string = ""
	// Iterate over the partitions
	for i := 0; i < 4; i++ {
		// fmt.Println("Partition id:", string(TempMBR.Partitions[i].Id[:]))
		// fmt.Println("Partition name:", Utils.CleanPartitionName(TempMBR.Partitions[i].Name[:]))
		//fmt.Println("Partition size:", TempMBR.Partitions[i].Size)
		//fmt.Println("Partition start:", TempMBR.Partitions[i].Start)
		// fmt.Println("Partition status:", string(TempMBR.Partitions[i].Status[:]))
		partitionStatus := TempMBR.Partitions[i].Status[0]
		partitionStatusStr = strconv.Itoa(int(partitionStatus))
		// fmt.Println("Partition status:", partitionStatusStr)

		if TempMBR.Partitions[i].Size != 0 {
			if strings.Contains(string(TempMBR.Partitions[i].Id[:]), id) {
				//fmt.Println("Partition found")
				if strings.Contains(partitionStatusStr, "1") {
					//fmt.Println("Partition is mounted")
					index = i
					partitionStart = TempMBR.Partitions[i].Start
					partitionFit = string(TempMBR.Partitions[i].Fit[:])
				} else {
					//fmt.Println("Partition is not mounted")
					return fmt.Errorf("Partition is not mounted")
				}
				break
			}
		}
	}

	if index != -1 {
		//fmt.Println("Partition found")
	} else {
		//fmt.Println("Partition not found")
		return fmt.Errorf("Partition not found")
	}

	file, err := Utils.OpenFile(diskFileName)
	if err != nil {
		return fmt.Errorf("Error abriendo el archivo")
	}

	var tempSuperblock Types.SuperBlock
	// Read object from bin file
	if err := Utils.ReadObject(file, &tempSuperblock, int64(TempMBR.Partitions[index].Start)); err != nil {
		return fmt.Errorf("Error leyendo el superbloque")
	}

	indexInode := InitSearch("/users.txt", file, tempSuperblock)
	//Utils.LineaDoble(80)
	//fmt.Println("Index Inode:", indexInode)
	if indexInode == -1 {
		//fmt.Println("User not found")
		return fmt.Errorf("User not found")
	}

	var tempInode Types.Inode
	if err := Utils.ReadObject(file, &tempInode, int64(tempSuperblock.S_inode_start+indexInode*int32(binary.Size(Types.Inode{})))); err != nil {
		return fmt.Errorf("Error leyendo el inode")
	}
	//fmt.Printf("Inode: %+v\n", tempInode)

	if tempInode.I_block[0] == -1 {
		//fmt.Println("User not found")
		return fmt.Errorf("User not found")
	}
	// else {
	// 	fmt.Println("User found")
	// }

	data := GetInodeFileDataOriginal(tempInode, file, tempSuperblock)
	//fmt.Println("Data:", data)

	//fmt.Println("Fileblock------------")
	// Dividir la cadena en líneas
	userId := ""
	//groupName := ""
	groupId := ""
	lines := strings.Split(data, "\n")

	// login -user=root -pass=123 -id=A119

	//fmt.Println("-------------------------------------------")
	// Iterar a través de las líneas
	for _, line := range lines {
		// Imprimir cada línea
		//fmt.Println("Line:", line)
		words := strings.Split(line, ",")
		//fmt.Println("Words:", words)

		if len(words) == 5 {
			if (strings.Contains(words[3], user)) && (strings.Contains(words[4], pass)) {
				login = true
				userId = words[0]
				break
			}
		}
	}

	for _, line := range lines {
		// Imprimir cada línea
		//fmt.Println("Line:", line)
		words := strings.Split(line, ",")
		//fmt.Println("Words:", words)

		if len(words) == 3 {
			if (strings.Contains(words[2], "root")) && (strings.Contains(words[1], "G")) {
				groupId = words[0]
				//groupName = words[2]
				break
			}
		}
	}

	//fmt.Println("-------------------------------------------")

	// Print object
	//fmt.Println("Inode", tempInode.I_block)

	// Close bin file
	defer file.Close()

	if login {
		//fmt.Println("User logged in")
		Global.Usuario.Id = userId
		Global.Usuario.Status = true

		// Convert id from string to int
		idInt, err := strconv.Atoi(userId)
		if err != nil {
			return err
		}

		Global.SesionActual.Id_user = idInt
		Global.SesionActual.User_name = user
		Global.SesionActual.Path = diskFileName
		Global.SesionActual.PartitionId = id
		Global.SesionActual.PartitionStart = partitionStart
		Global.SesionActual.System_type = int(tempSuperblock.S_filesystem_type)
		copy(Global.SesionActual.Fit[:], []byte(partitionFit))
		//Global.SesionActual.Id_grp = buscarGrupo(groupName)

		idInt, err = strconv.Atoi(groupId)
		if err != nil {
			return err
		}

		Global.SesionActual.Id_grp = idInt
	}

	//fmt.Println("======End LOGIN======")

	// var tempDirectoryBlock Types.DirectoryBlock
	// if err := Utils.ReadObject(file, &tempDirectoryBlock, int64(tempSuperblock.S_block_start+tempInode.I_block[0]*int32(binary.Size(Types.DirectoryBlock{})))); err != nil {
	// 	return
	// }

	// for _, folder := range tempDirectoryBlock.B_content {
	// 	fmt.Println("Folder === Name:", strings.TrimSpace(string(folder.B_name[:])), "B_inodo", folder.B_inodo)
	// }

	// if strings.TrimSpace(string(tempDirectoryBlock.B_content[0].B_name[:])) == user {
	// 	fmt.Println("User found")
	// } else {
	// 	fmt.Println("User not found")
	// 	return
	//}

	return nil

}
