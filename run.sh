#!/bin/bash

# Clear screen 
clear

# Delete .dsk files
rm -f ./disks/MIA/P1/*.dsk

# Delete files with specific patterns
rm -f ./reportes/mbr*.*
rm -f ./reportes/disk*.*
rm -f ./reportes/reporte*.*
rm -rf /home/linovallejo/archivos/reports/*.*

# Run Go program
go run .