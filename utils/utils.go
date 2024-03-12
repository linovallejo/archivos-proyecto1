package Utils

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	Types "proyecto1/types"
	"strconv"
	"strings"
	"time"
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

func PrintMBRv2(mbr Types.MBR) {
	LineaDoble(80)
	fmt.Println("*** MBR ***")
	fmt.Printf("MBR Size: %d\n", mbr.MbrTamano)
	creationDate, _ := strconv.ParseInt(string(mbr.MbrFechaCreacion[:]), 10, 64)
	fmt.Printf("MBR Creation Date: %s\n", time.Unix(creationDate, 0).Format("2006-01-02 15:04:05"))
	fmt.Printf("MBR Disk Signature: %d\n", mbr.MbrDiskSignature)
	fmt.Printf("Disk Fit: %c\n", mbr.DskFit[0])

	fmt.Println("\n*** Partitions ***")
	for i, part := range mbr.Partitions {
		fmt.Printf("\nPartition %d:\n", i+1)
		fmt.Printf("  Status: %c\n", part.Status[0])
		fmt.Printf("  Type: %c\n", part.Type[0])
		fmt.Printf("  Fit: %c\n", part.Fit[0])
		fmt.Printf("  Start: %d\n", part.Start)
		fmt.Printf("  Size: %d\n", part.Size)
		fmt.Printf("  Name: %s\n", string(part.Name[:])) // Convert byte array to string
		fmt.Printf("  Correlative: %d\n", part.Correlative)
		fmt.Printf("  Id: %v\n", part.Id) // Print byte array
	}
	LineaDoble(80)
}

func ReturnFitType(fitType string) [1]byte {
	var fitByte byte
	switch fitType {
	case "BF":
		fitByte = 'B'
	case "FF":
		fitByte = 'F'
	case "WF":
		fitByte = 'W'
	default:
		fitByte = 'F' // FF es el valor predeterminado
	}
	return [1]byte{fitByte}
}

func PrintMBRv3(mbr *Types.MBR) {
	LineaDoble(80)
	fmt.Println("*** MBR ***")
	fmt.Printf("MBR Size: %d\n", mbr.MbrTamano)
	creationDate, _ := strconv.ParseInt(string(mbr.MbrFechaCreacion[:]), 10, 64)
	fmt.Printf("MBR Creation Date: %s\n", time.Unix(creationDate, 0).Format("2006-01-02 15:04:05"))
	fmt.Printf("MBR Disk Signature: %d\n", mbr.MbrDiskSignature)
	fmt.Printf("Disk Fit: %c\n", mbr.DskFit[0])

	fmt.Println("\n*** Partitions ***")
	for i, part := range mbr.Partitions {
		fmt.Printf("\nPartition %d:\n", i+1)
		fmt.Printf("  Status: %c\n", part.Status[0])
		fmt.Printf("  Type: %c\n", part.Type[0])
		fmt.Printf("  Fit: %c\n", part.Fit[0])
		fmt.Printf("  Start: %d\n", part.Start)
		fmt.Printf("  Size: %d\n", part.Size)
		fmt.Printf("  Name: %s\n", string(part.Name[:])) // Convert byte array to string
		fmt.Printf("  Correlative: %d\n", part.Correlative)
		fmt.Printf("  Id: %v\n", part.Id)            // Print byte array
		fmt.Printf("  Id: %s\n", string(part.Id[:])) // Print string
	}
	LineaDoble(80)
}

func PrintMounted(mbr *Types.MBR) {
	LineaDoble(80)
	// fmt.Println("*** MBR ***")
	// fmt.Printf("MBR Size: %d\n", mbr.MbrTamano)
	// creationDate, _ := strconv.ParseInt(string(mbr.MbrFechaCreacion[:]), 10, 64)
	// fmt.Printf("MBR Creation Date: %s\n", time.Unix(creationDate, 0).Format("2006-01-02 15:04:05"))
	// fmt.Printf("MBR Disk Signature: %d\n", mbr.MbrDiskSignature)
	// fmt.Printf("Disk Fit: %c\n", mbr.DskFit[0])

	fmt.Println("\n*** Particiones Montadas ***")
	for i, part := range mbr.Partitions {
		if part.Status[0] == 1 {
			fmt.Printf("\nPartition %d:\n", i+1)
			fmt.Printf("  Status: %d\n", part.Status[0])
			fmt.Printf("  Type: %c\n", part.Type[0])
			fmt.Printf("  Fit: %c\n", part.Fit[0])
			fmt.Printf("  Start: %d\n", part.Start)
			fmt.Printf("  Size: %d\n", part.Size)
			fmt.Printf("  Name: %s\n", string(part.Name[:]))
			fmt.Printf("  Correlative: %d\n", part.Correlative)
			fmt.Printf("  Id: %v\n", part.Id)
			fmt.Printf("  Id: %s\n", string(part.Id[:]))

		}
	}
	LineaDoble(80)
}
