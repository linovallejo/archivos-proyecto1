# Clear the screen for better visibility
Clear-Host

# Delete .dsk files in the specified disks directory
del .\disks\MIA\P1\*.dsk

# Delete files starting with "disk" and "mbr" in the reportes directory
del .\reportes\mbr*.*
del .\reportes\disk*.* 
del .\reportes\reporte*.*
del .\reportes\*.jpg
del .\reportes\*.png
del .\reportes\*.dot



# Run the Go program
go run .
