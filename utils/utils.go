package Utils

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func AbrirArchivo(name string) (*os.File, error) {
	file, err := os.OpenFile(name, os.O_RDWR, 0644)
	if err != nil {
		fmt.Println("Err OpenFile==", err)
		return nil, err
	}
	return file, nil
}

func WriteObject(file *os.File, data interface{}, position int64) error {
	file.Seek(position, 0)
	err := binary.Write(file, binary.LittleEndian, data)
	if err != nil {
		fmt.Println("Err WriteObject==", err)
		return err
	}
	return nil
}

// Function to Read an object from a bin file
func ReadObject(file *os.File, data interface{}, position int64) error {
	file.Seek(position, 0)
	err := binary.Read(file, binary.LittleEndian, data)
	if err != nil {
		fmt.Println("Err ReadObject==", err)
		return err
	}
	return nil
}

func LimpiarConsola() {
	fmt.Print("\033[H\033[2J")
}

func LineaDoble(longitud int) {
	fmt.Println(strings.Repeat("=", longitud))
}

func PrintCopyright() {
	LineaDoble(60)
	fmt.Println("Lino Antonio Garcia Vallejo")
	fmt.Println("Carné: 9017323")
	LineaDoble(60)
}

func CleanPartitionName(name []byte) string {
	n := bytes.IndexByte(name, 0)
	if n == -1 {
		n = len(name)
	}
	return string(name[:n])
}

func EnsurePathExists(path string) error {
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return os.MkdirAll(dir, os.ModePerm)
	}
	return nil
}
