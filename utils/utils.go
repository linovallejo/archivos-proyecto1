package Utils

import (
	"bytes"
	"fmt"
	"strings"
)

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
