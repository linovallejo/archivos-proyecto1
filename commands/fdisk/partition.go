package Fdisk

import (
	"fmt"
	Types "proyecto1/types"
)

var logicalPartitions map[string]*LogicalPartitionInfo

type LogicalPartitionInfo struct {
	ExtendedStart int32
	FirstEBR      *Types.EBR
}

func init() {
	logicalPartitions = make(map[string]*LogicalPartitionInfo)
}

// Exported function to add or update logical partitions
func AddOrUpdateLogicalPartition(diskFileName string, info *LogicalPartitionInfo) {
	logicalPartitions[diskFileName] = info
}

// Exported function to retrieve a logical partition
func GetLogicalPartition(diskFileName string) (*LogicalPartitionInfo, bool) {
	//fmt.Println("GetLogicalPartition: ", logicalPartitions[diskFileName])
	info, exists := logicalPartitions[diskFileName]
	return info, exists
}

func SetFirstEBR(diskFileName string, newEBR *Types.EBR) error {
	// Retrieve the logical partition info for the given disk file name
	info, exists := logicalPartitions[diskFileName]
	if !exists {
		return fmt.Errorf("extended partition does not exist for disk %s", diskFileName)
	}

	// Set the first EBR for this logical partition
	info.FirstEBR = newEBR
	//fmt.Println("First EBR set for disk", diskFileName)
	return nil
}

// AddEBRToChain appends a new EBR to the chain of EBRs for the logical partition associated with the given disk file name.
func AddEBRToChain(diskFileName string, newEBR *Types.EBR) error {
	// Retrieve the logical partition info for the given disk file name
	info, exists := logicalPartitions[diskFileName]
	if !exists {
		return fmt.Errorf("extended partition does not exist for disk %s", diskFileName)
	}

	// Find the last EBR in the chain
	currentEBR := info.FirstEBR
	if currentEBR == nil {
		return fmt.Errorf("no initial EBR set for disk %s; use SetFirstEBR first", diskFileName)
	}
	for currentEBR.PartNext != nil {
		currentEBR = currentEBR.PartNext
	}

	// Append the new EBR to the end of the chain
	currentEBR.PartNext = newEBR
	//fmt.Println("EBR added to chain for disk", diskFileName)
	return nil
}
