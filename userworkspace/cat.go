package userworkspace

import (
	"fmt"
	"strings"
)

func ExtractCatParams(params []string) (string, error) {
	var pfile string = ""

	if len(params) == 0 {
		return "", fmt.Errorf("No se encontraron parámetros")
	}
	var parametrosObligatoriosOk bool = false
	pfileOk := false
	for _, param1 := range params {
		if strings.HasPrefix(param1, "-file=") {
			pfileOk = true
		}
	}

	parametrosObligatoriosOk = pfileOk

	if !parametrosObligatoriosOk {
		return "", fmt.Errorf("No se encontraron parámetros obligatorios")
	}

	for _, param := range params {
		switch {
		case strings.HasPrefix(param, "-file="):
			pfile = strings.TrimPrefix(param, "-file=")
		}
	}

	return pfile, nil
}

func EjecutarCat(pathUsersFile string, partitionId string, diskFileName string) ([]string, error) {
	fileContents, err := ReturnFileContents(pathUsersFile, partitionId, diskFileName)
	if err != nil {
		//fmt.Println(err)
		return nil, err
	} else {
		//fmt.Println("fileContents:", fileContents)
		return fileContents, nil
	}
}
