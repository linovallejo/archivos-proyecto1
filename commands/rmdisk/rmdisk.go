package rmdisk

import (
	"fmt"
	"os"
	"strings"
	"unicode"
)

func ExtractRmdiskParams(params []string) (string, error) {
	var driveletter string

	for _, param := range params {
		switch {
		case strings.HasPrefix(param, "-driveletter="):
			driveletter = strings.TrimPrefix(param, "-driveletter=")
			// Validar la letra de la partición
			if len(driveletter) != 1 || !unicode.IsLetter(rune(driveletter[0])) {
				return "", fmt.Errorf("La letra de la partición debe ser un único carácter alfabérico")
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
