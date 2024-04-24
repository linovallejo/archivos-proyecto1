#!/bin/bash

# Clear screen 
clear

# Delete .dsk files
rm -f ./disks/MIA/P1/*.dsk

# Delete files with specific patterns
rm -f ./reportes/mbr*.*
rm -f ./reportes/disk*.*
rm -f ./reportes/reporte*.*
rm -rf ./reportes/*.jpg
rm -rf ./reportes/*.png
rm -rf ./reportes/*.pdf
rm -rf ./reportes/*.txt
rm -rf ./reportes/*.dot
rm -rf /home/linovallejo/archivos/reports/*.*

# Run Go program
go run .
