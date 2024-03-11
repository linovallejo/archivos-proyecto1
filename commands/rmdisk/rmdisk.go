package rmdisk

import (
	"fmt"
	"os"
	"strings"
	"unicode"
)

func ExtractRmdiskParams(params []string) (string, error) {
	var driveletter string

	if len(params) == 0 {
		return "", fmt.Errorf("No se encontraron parámetros")
	}

	var parametrosObligatoriosOk bool = false
	for _, param1 := range params {
		if strings.HasPrefix(param1, "-driveletter=") {
			parametrosObligatoriosOk = true
			break
		}
	}

	if !parametrosObligatoriosOk {
		return "", fmt.Errorf("No se encontraron parámetros obligatorios")
	}

	for _, param := range params {
		switch {
		case strings.HasPrefix(param, "-driveletter="):
			driveletter = strings.TrimPrefix(param, "-driveletter=")
			// Validar la letra de la partición
			if strings.TrimSpace(driveletter) == "" {
				return "", fmt.Errorf("La letra del drive es un parametro obligatorio")
			} else if len(driveletter) != 1 || !unicode.IsLetter(rune(driveletter[0])) {
				return "", fmt.Errorf("La letra de la partición debe ser un único carácter alfabético")
			}
		}
	}

	return driveletter, nil
}

func RemoveDisk(diskFileName string) error {
	err := os.Remove(diskFileName)
	if err != nil {
		return err
	}

	return nil
}
