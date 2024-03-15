package Rep

import (
	"fmt"
	Utils "proyecto1/utils"
	"strings"
)

func isValidReportName(name string) bool {
	validNames := []string{"mbr", "disk", "inode", "journaling", "block", "bm_inode", "bm_block", "tree", "sb", "file", "ls"}
	for _, validName := range validNames {
		if name == validName {
			return true
		}
	}
	return false
}

func ExtractRepParams(params []string) (string, string, string, error) {
	var id string = ""
	var name string = ""
	var path string = ""

	for _, param := range params {
		if strings.HasPrefix(param, "-id=") {
			id = strings.TrimPrefix(param, "-id=")
			// Validar el id de la partición
			// TODO
		} else if strings.HasPrefix(param, "-name=") {
			name = strings.TrimPrefix(param, "-name=")
			if !isValidReportName(name) {
				return "", "", "", fmt.Errorf("Parametro nombre de reporte invalido")
			}
		} else if strings.HasPrefix(param, "-path=") {
			path = strings.TrimPrefix(param, "-path=")
			trimmedPath := strings.Trim(path, "\"")

			if err := Utils.EnsurePathExists(trimmedPath); err != nil {
				return "", "", "", err
			}
		}
	}

	return id, name, path, nil
}
