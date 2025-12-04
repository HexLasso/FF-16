// TODO reword ERRORs/WARNINGs
// TODO revise if variable names and rename them if needed
// TODO revise function namings and rename them if needed
// TODO write test files
// TODO revie Freq threashold
// TODO revise max gaps
// TODO revise max blocks per chunk
package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Last update instead of version
const LastUpdate = "11-Jul-2023"

// Block size is fixed to 256 bytes
const BlockSize = 256

// Default values for cmd parameters
const DefDictFileName = "dict.csv"
const DefBlocksPerChunk = 1
const DefMaxGap = 31
const DefThreshold = 5

// Operation ranges
const BlocksPerChunkLo = 1
const BlocksPerChunkHi = 256
const ChunksPerFileLo = 1
const ChunksPerFileHi = 65536
const MaxGapLo = 0
const MaxGapHi = 127
const ThresholdLo = 1
const ThresholdHi = 255
const FileSizeLo = 0
const FileSizeHi = BlocksPerChunkHi * ChunksPerFileHi
const DictColumnWidth = 2
const MaxDictRows = 65536

// Worst case array size for varying gaps
// +1 for gap=0
var GapArr [MaxGapHi + 1]int

type PatternInfo struct {
	Byte1 byte // First byte in the pattern
	Byte2 byte // Second byte in the pattern
	Gap   int  // Gap (in bytes) between the fist and the second bytes
	Hits  int  // Counter for the pattern hits
}

func Check(e error) {
	if e != nil {
		panic(e)
	}
}

func IsOutOfRangeValue(val int, lo int, hi int) bool {
	if (val < lo) || (val > hi) {
		return true
	}
	return false
}

func ToPrintable(byte1 byte, byte2 byte) string {
	var b1 byte = '.'
	var b2 byte = '.'
	if (byte1 >= 0x20) && (byte1 <= 0x7E) {
		b1 = byte1
	}
	if (byte2 >= 0x20) && (byte2 <= 0x7E) {
		b2 = byte2
	}
	return fmt.Sprintf("%c%c", b1, b2)
}

func Help() {
	fmt.Printf("FF-16 searches for frequent 16-bit patterns in file. Last update: %s\n\n", LastUpdate)
	fmt.Printf("ff-16 [filename] [-d <filename>] [<-bpc <%d..%d>|-cpf <%d..%d>>] [-g <%d..%d>] [-t <%d..%d>]\n\n",
		BlocksPerChunkLo, BlocksPerChunkHi, ChunksPerFileLo, ChunksPerFileHi, MaxGapLo, MaxGapHi, ThresholdLo, ThresholdHi)
	fmt.Printf("  <filename>      Target file\n")
	fmt.Printf("  -d <filename>   Dictionary file  (Default: %s)\n", DefDictFileName)
	fmt.Printf("  -bpc <%d..%d>   Blocks per chunk (Default: %d)\n", BlocksPerChunkLo, BlocksPerChunkHi, DefBlocksPerChunk)
	fmt.Printf("  -cpf <%d..%d> Chunks per file  (Default: not specified)\n", ChunksPerFileLo, ChunksPerFileHi)
	fmt.Printf("  -g <%d..%d>     Max gaps         (Default: %d)\n", MaxGapLo, MaxGapHi, DefMaxGap)
	fmt.Printf("  -t <%d..%d>     Freq threshold   (Default: %d)\n\n", ThresholdLo, ThresholdHi, DefThreshold)
}

func main() {
	// At least one parameter (i.e. filename) is mandatory
	if len(os.Args) < 2 {
		Help()
		return
	}

	// Parameter parsing
	fileName := ""
	dictFileName := DefDictFileName
	blocksPerChunk := DefBlocksPerChunk
	chunkPerFile := 0
	gap := DefMaxGap
	threshold := DefThreshold
	bpcSet := false
	cpfSet := false
	missingValue := false
	badValueFormat := false
	outOfRangeValue := false
	var err error = nil
	for i := 1; i < len(os.Args); i++ {
		arg := strings.ToLower(os.Args[i])
		if strings.EqualFold(arg, "-d") {
			if len(os.Args) <= i+1 {
				missingValue = true
			} else {
				dictFileName = os.Args[i+1]
				i++
			}
		} else if strings.EqualFold(arg, "-bpc") {
			if len(os.Args) <= i+1 {
				missingValue = true
			} else {
				blocksPerChunk, err = strconv.Atoi(os.Args[i+1])
				if err != nil {
					badValueFormat = true
				}
				if IsOutOfRangeValue(blocksPerChunk, BlocksPerChunkLo, BlocksPerChunkHi) {
					outOfRangeValue = true
				}
				i++
			}
			bpcSet = true
		} else if strings.EqualFold(arg, "-cpf") {
			if len(os.Args) <= i+1 {
				missingValue = true
			} else {
				chunkPerFile, err = strconv.Atoi(os.Args[i+1])
				if err != nil {
					badValueFormat = true
				}
				if IsOutOfRangeValue(chunkPerFile, ChunksPerFileLo, ChunksPerFileHi) {
					outOfRangeValue = true
				}
				i++
			}
			cpfSet = true
		} else if strings.EqualFold(arg, "-g") {
			if len(os.Args) <= i+1 {
				missingValue = true
			} else {
				gap, err = strconv.Atoi(os.Args[i+1])
				if err != nil {
					badValueFormat = true
				}
				if IsOutOfRangeValue(gap, MaxGapLo, MaxGapHi) {
					outOfRangeValue = true
				}
				i++
			}
		} else if strings.EqualFold(arg, "-t") {
			if len(os.Args) <= i+1 {
				missingValue = true
			} else {
				threshold, err = strconv.Atoi(os.Args[i+1])
				if err != nil {
					badValueFormat = true
				}
				if IsOutOfRangeValue(threshold, ThresholdLo, ThresholdHi) {
					outOfRangeValue = true
				}
				i++
			}
		} else {
			// Target filename is not prefixed with "-"
			if fileName == "" {
				fileName = os.Args[i]
			} else {
				fmt.Printf("ERROR: Unknown parameter: \"%s\".\n\n", os.Args[i])
				Help()
				return
			}
		}

		if missingValue {
			fmt.Printf("ERROR: Missing value for: \"%s\".\n\n", os.Args[i])
			Help()
			return
		}

		if badValueFormat {
			fmt.Printf("ERROR: The value: \"%s\" has a bad format in: \"%s\".\n\n", os.Args[i], os.Args[i-1])
			Help()
			return
		}

		if outOfRangeValue {
			fmt.Printf("ERROR: The value: \"%s\" is out of range for: \"%s\".\n\n", os.Args[i], os.Args[i-1])
			Help()
			return
		}

		if bpcSet && cpfSet {
			fmt.Printf("ERROR: The parameters \"-bpc\" and \"-cpf\" are mutually exclusive. Use only one of them.\n\n")
			Help()
			return
		}
	}

	// Input file
	inFile, err := os.Open(fileName)
	if err != nil {
		fmt.Printf("ERROR: Unable to access \"%s\".\n\n", fileName)
		Help()
		return
	}
	fi, err := inFile.Stat()
	Check(err)
	if fi.IsDir() {
		fmt.Printf("ERROR: The supplied parameter \"%s\" is a directory. You need to supply a file.\n\n", fileName)
		Help()
		return
	}
	inFileSize := int(fi.Size())
	if IsOutOfRangeValue(inFileSize, FileSizeLo, FileSizeHi) {
		fmt.Printf("ERROR: The file is too large in size. The file size should be up to %dMB.\n\n", FileSizeHi/1024/1024)
		Help()
		return
	}

	// Dictionary file (i.e. CSV file)
	csvRecordCount := 0
	var csvRecords [][]string
	csvFileBuf, err := os.ReadFile(dictFileName)
	if err == nil {
		// TODO check duplicate entries
		r := csv.NewReader(strings.NewReader(string(csvFileBuf)))
		r.Comma = ';'
		r.Comment = '#'
		csvRecords, err = r.ReadAll()
		Check(err)
		csvRecordCount = len(csvRecords)
		// Verification
		if csvRecordCount > MaxDictRows {
			fmt.Printf("ERROR: Too many rows in dictionary.\n\n")
			Help()
			return
		}
		if len(csvRecords[0]) != DictColumnWidth {
			fmt.Printf("ERROR: Column width must be 2.\n\n")
			Help()
			return
		}
	} else {
		fmt.Printf("WARNING: File not found: %s\n", dictFileName)
	}

	// Init
	buffer := make([]byte, BlockSize)
	blockFreqTable := make(map[string]PatternInfo)
	chunkFreqTable := make(map[string]PatternInfo)
	for i := 0; i <= gap; i++ {
		GapArr[i] = i
	}

	// Init chunk size
	chunkSize := blocksPerChunk * BlockSize

	// Calculate chunksize for chunks-per-file parameter
	if chunkPerFile != 0 {
		if inFileSize/chunkPerFile < BlockSize {
			// Chunks per file parameter is too big
			chunkPerFile = 0
			fmt.Printf("WARNING: Chunks per file parameter is too big considering the file size. You cannot split the file to chunks less than 256 byte\n")
		} else {
			chunkSize = ((inFileSize / chunkPerFile) / BlockSize) * BlockSize
			if chunkSize == 0 {
				// The whole file will be one chunk
				chunkSize = (inFileSize / 256) * 256
			}
			blocksPerChunk = chunkSize / 256

			if IsOutOfRangeValue(blocksPerChunk, BlocksPerChunkLo, BlocksPerChunkHi) {
				fmt.Printf("ERROR: The calculated BPC is too big\n\n")
				Help()
				return
			}
		}
	}

	// Set chunk or block mode for printing
	blockMode := true
	if chunkSize != BlockSize {
		blockMode = false
	}

	// Print header
	fmt.Printf("Offset   Size     Pattern      Ascii Bpc Freq Dict\n")

	actualChunkSize := 0
	// Read block of data in each iteration
	for fileOffs := 0; fileOffs < inFileSize; fileOffs += BlockSize {
		// Read block of data
		bytesRead, err := inFile.Read(buffer)
		Check(err)

		// Build pattern frequency table for the block
		for skipIdx := 0; skipIdx <= gap; skipIdx++ {
			for bufIdx := 0; bufIdx < bytesRead-1-GapArr[skipIdx]; bufIdx++ {
				key := fmt.Sprintf("%02x +(%d) %02x \n", buffer[bufIdx], GapArr[skipIdx], buffer[bufIdx+GapArr[skipIdx]+1])
				hits := blockFreqTable[key].Hits + 1
				blockFreqTable[key] = PatternInfo{
					Byte1: buffer[bufIdx],
					Byte2: buffer[bufIdx+GapArr[skipIdx]+1],
					Gap:   GapArr[skipIdx],
					Hits:  hits}
			}
		}

		// Get the top pattern of the block
		top := 0
		topKey := ""
		for k, v := range blockFreqTable {
			if v.Hits > top {
				top = v.Hits
				topKey = k
			}
		}
		// If there are multiple blocks with the same hits choose deterministically
		for k, v := range blockFreqTable {
			if v.Hits == blockFreqTable[topKey].Hits {
				if topKey > k {
					topKey = k
				}
			}
		}

		hex := "-"
		printable := "-"
		hitFreq := "-"
		dict := "-"

		// Threshold applies for blocks only
		if blockFreqTable[topKey].Hits >= threshold {
			if blockFreqTable[topKey].Gap == 0 {
				hex = fmt.Sprintf("%02X %02X", blockFreqTable[topKey].Byte1, blockFreqTable[topKey].Byte2)
			} else {
				hex = fmt.Sprintf("%02X +(%d) %02X", blockFreqTable[topKey].Byte1, blockFreqTable[topKey].Gap, blockFreqTable[topKey].Byte2)
			}

			printable = ToPrintable(blockFreqTable[topKey].Byte1, blockFreqTable[topKey].Byte2)
			printable = "|" + printable + "|"

			for i := 0; i < csvRecordCount; i++ {
				if strings.EqualFold((csvRecords[i])[0], hex) {
					dict = strings.Trim((csvRecords[i])[1], " ")
				}
			}

			hitFreq = strconv.Itoa(blockFreqTable[topKey].Hits)
		}

		// Update pattern frequency table for the chunk with the top pattern from the block
		hits := chunkFreqTable[hex].Hits + 1

		chunkFreqTable[hex] = PatternInfo{
			Byte1: blockFreqTable[topKey].Byte1,
			Byte2: blockFreqTable[topKey].Byte2,
			Gap:   blockFreqTable[topKey].Gap,
			Hits:  hits}

		if blockMode {
			// Block mode
			fmt.Printf("%08X %08x %-12s %5s %3d %4s %s\n", fileOffs, bytesRead, hex, printable, blocksPerChunk, hitFreq, dict)
		} else {
			// Chunk mode
			actualChunkSize += bytesRead

			firstChunk := false
			if (fileOffs+bytesRead)-chunkSize <= 0 {
				firstChunk = true
			}
			lastChunk := false
			if fileOffs+bytesRead == inFileSize {
				lastChunk = true
			}

			// Print Chunk info when all the Blocks in the Chunk processed OR if it's the last Chunk in the file
			if ((fileOffs != 0) && ((fileOffs+BlockSize)%chunkSize == 0)) || lastChunk {
				// Get the top block of the chunk
				top := 0
				topKey := ""
				for k, v := range chunkFreqTable {
					if v.Hits > top {
						top = v.Hits
						topKey = k
					}
				}
				// If there are multiple blocks with the same hits choose deterministically
				for k, v := range chunkFreqTable {
					if v.Hits == chunkFreqTable[topKey].Hits {
						if topKey > k {
							topKey = k
						}
					}
				}

				printable = "-"
				hex = "-"
				if topKey != "-" {
					if chunkFreqTable[topKey].Gap == 0 {
						hex = fmt.Sprintf("%02X %02X", chunkFreqTable[topKey].Byte1, chunkFreqTable[topKey].Byte2)
					} else {
						hex = fmt.Sprintf("%02X +(%d) %02X", chunkFreqTable[topKey].Byte1, chunkFreqTable[topKey].Gap, chunkFreqTable[topKey].Byte2)
					}

					printable = ToPrintable(chunkFreqTable[topKey].Byte1, chunkFreqTable[topKey].Byte2)
					printable = "|" + printable + "|"
				}
				dict = "-"
				for i := 0; i < csvRecordCount; i++ {
					if strings.EqualFold((csvRecords[i])[0], topKey) {
						dict = strings.Trim((csvRecords[i])[1], " ")
					}
				}
				hitFreq = strconv.Itoa(top)

				offset := (fileOffs + BlockSize) - chunkSize

				if firstChunk && lastChunk {
					offset = 0
				}

				actualBpc := actualChunkSize / BlockSize
				// If the block in the chunk is less than 256 bytes in size, it still counts as a block
				if actualChunkSize%BlockSize != 0 {
					actualBpc++
				}

				fmt.Printf("%08X %08X %-12s %5s %3d %4s %s\n", offset, actualChunkSize, hex, printable, actualBpc, hitFreq, dict)

				// Clear chunk frequency table for the next chunk
				for k := range chunkFreqTable {
					delete(chunkFreqTable, k)
				}

				// Reset actual chunk size
				actualChunkSize = 0
			} // if
		} // else
		// Clear block frequency table for the next block
		for key := range blockFreqTable {
			delete(blockFreqTable, key)
		}
	} // for
}
