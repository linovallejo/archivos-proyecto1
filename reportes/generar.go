package Reportes

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func CrearArchivo(nombre_archivo string) {
	var _, err = os.Stat(nombre_archivo)

	if os.IsNotExist(err) {
		var file, err = os.Create(nombre_archivo)
		if err != nil {
			fmt.Printf("Error creating file: %v", err)
			return
		}
		defer file.Close()
	}
}

func EscribirArchivo(contenido string, nombre_archivo string) error {
	// Ensure necessary directories exist
	dir := filepath.Dir(nombre_archivo)

	err := os.MkdirAll(dir, 0755) // More permissive for directory creation
	if err != nil {
		return fmt.Errorf("error creating directory: %w", err)
	}

	file, err := os.OpenFile(nombre_archivo, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	_, err = file.WriteString(contenido)
	if err != nil {
		return fmt.Errorf("error writing to file: %w", err)
	}
	err = file.Sync()
	if err != nil {
		return fmt.Errorf("error syncing file: %w", err)
	}

	return nil
}
func Ejecutar(nombre_imagen string, archivo string, extension string) error {
	newExtension := strings.TrimPrefix(extension, ".")
	path, err := exec.LookPath("dot")
	if err != nil {
		return fmt.Errorf("error finding 'dot' executable: %w", err)
	}

	fileFormat := "-T" + newExtension
	cmd := exec.Command(path, fileFormat, archivo, "-o", nombre_imagen)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error executing 'dot' command: %w", err)
	}

	return nil
}

func VerReporte(nombre_imagen string) {
	fmt.Println("Abriendo reporte..." + nombre_imagen)
	var cmd *exec.Cmd
	//cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", nombre_imagen)
	cmd = exec.Command("cmd", "/C", "start", "", nombre_imagen)

	err := cmd.Start()
	if err != nil {
		fmt.Println("Error al abrir el archivo:", err)
	}
}
