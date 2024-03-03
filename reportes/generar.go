package Reportes

import (
	"fmt"
	"os"
	"os/exec"
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

func EscribirArchivo(contenido string, nombre_archivo string) {
	var file, err = os.OpenFile(nombre_archivo, os.O_RDWR, 0644)
	if err != nil {
		return
	}
	defer file.Close()
	_, err = file.WriteString(contenido)
	if err != nil {
		return
	}
	err = file.Sync()
	if err != nil {
		return
	}
}

func Ejecutar(nombre_imagen string, archivo string, extension string) {
	path, _ := exec.LookPath("dot")
	cmd, _ := exec.Command(path, "-Tjpg", archivo).Output()
	mode := 0777
	_ = os.WriteFile(nombre_imagen, cmd, os.FileMode(mode))
	//dot ejemplo.dot -Tjpg ejemplo.jpg -o
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
