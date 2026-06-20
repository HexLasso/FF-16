// FF-16, stands for Find Frequent 16-bit, searches for frequent 16-bit patterns in a given file
//
// The latest version can be accessed at https://github.com/HexLasso/FF-16
package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Last update
const LastUpdate = "20-Jun-2026"

// Block size is 256 bytes
const BlockSize = 256

// Zero filter values
const (
	ZeroFilterNone   = 0 // No filtering
	ZeroFilterBoth   = 1 // Exclude 00 00 patterns
	ZeroFilterEither = 2 // Exclude patterns containing 00
)

// Default values for cmd parameters
const DefDictFileName = "dict.csv"
const DefBlocksPerChunk = 1
const DefMaxGap = 31
const DefThreshold = 5
const DefZeroFilter = ZeroFilterNone

// Operation ranges
const BlocksPerChunkLo = 1
const BlocksPerChunkHi = 256
const ChunksPerFileLo = 1
const ChunksPerFileHi = 65536
const MaxGapLo = 0
const MaxGapHi = 127
const ThresholdLo = 1
const ThresholdHi = 255
const ZeroFilterLo = 0
const ZeroFilterHi = 2
const FileSizeLo = 0
const FileSizeHi = BlocksPerChunkHi * ChunksPerFileHi
const DictColumnCount = 2
const MaxDictRows = 65536

// Help hint for errors
const HelpHint = "Run \"ff-16\" without parameters for help."

// Worst case array size for varying gaps
// +1 for gap=0
var GapTable [MaxGapHi + 1]int

type PatternInfo struct {
	First  byte // First byte of the pattern
	Second byte // Second byte of the pattern
	Gap    int  // Number of bytes between First and Second
	Hits   int  // Count of matches
}

func PanicIfError(e error) {
	if e != nil {
		panic(e)
	}
}

func IsOutOfRange(val int, lo int, hi int) bool {
	if (val < lo) || (val > hi) {
		return true
	}
	return false
}

func ToPrintable(first byte, second byte) string {
	var firstPrintable byte = '.'
	var secondPrintable byte = '.'
	if (first >= 0x20) && (first <= 0x7E) {
		firstPrintable = first
	}
	if (second >= 0x20) && (second <= 0x7E) {
		secondPrintable = second
	}
	return fmt.Sprintf("%c%c", firstPrintable, secondPrintable)
}

func Help() {
	fmt.Printf("FF-16 searches for frequent 16-bit patterns in file. Last update: %s\n\n", LastUpdate)
	fmt.Printf("ff-16 [filename] [-d <filename>] [<-bpc <%d..%d>|-cpf <%d..%d>>] [-g <%d..%d>] [-t <%d..%d>] [-z <%d..%d>]\n\n",
		BlocksPerChunkLo, BlocksPerChunkHi, ChunksPerFileLo, ChunksPerFileHi, MaxGapLo, MaxGapHi, ThresholdLo, ThresholdHi, ZeroFilterLo, ZeroFilterHi)
	fmt.Printf("  <filename>      Target file\n")
	fmt.Printf("  -d <filename>   Dictionary file  (Default: %s)\n", DefDictFileName)
	fmt.Printf("  -bpc <%d..%d>   Blocks per chunk (Default: %d)\n", BlocksPerChunkLo, BlocksPerChunkHi, DefBlocksPerChunk)
	fmt.Printf("  -cpf <%d..%d> Chunks per file  (Default: not specified)\n", ChunksPerFileLo, ChunksPerFileHi)
	fmt.Printf("  -g <%d..%d>     Max gaps         (Default: %d)\n", MaxGapLo, MaxGapHi, DefMaxGap)
	fmt.Printf("  -t <%d..%d>     Freq threshold   (Default: %d)\n", ThresholdLo, ThresholdHi, DefThreshold)
	fmt.Printf("  -z <%d..%d>       Byte 00 filter   (Default: %d)\n\n", ZeroFilterLo, ZeroFilterHi, DefZeroFilter)
}

func main() {
	// At least one parameter (i.e. filename) is required
	if len(os.Args) < 2 {
		Help()
		return
	}

	// Parameter parsing
	fileName := ""
	dictFileName := DefDictFileName
	blocksPerChunk := DefBlocksPerChunk
	chunksPerFile := 0
	gap := DefMaxGap
	threshold := DefThreshold
	zero := DefZeroFilter
	bpcSet := false
	cpfSet := false
	missingValue := false
	invalidValue := false
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
					invalidValue = true
				}
				if IsOutOfRange(blocksPerChunk, BlocksPerChunkLo, BlocksPerChunkHi) {
					outOfRangeValue = true
				}
				i++
			}
			bpcSet = true
		} else if strings.EqualFold(arg, "-cpf") {
			if len(os.Args) <= i+1 {
				missingValue = true
			} else {
				chunksPerFile, err = strconv.Atoi(os.Args[i+1])
				if err != nil {
					invalidValue = true
				}
				if IsOutOfRange(chunksPerFile, ChunksPerFileLo, ChunksPerFileHi) {
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
					invalidValue = true
				}
				if IsOutOfRange(gap, MaxGapLo, MaxGapHi) {
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
					invalidValue = true
				}
				if IsOutOfRange(threshold, ThresholdLo, ThresholdHi) {
					outOfRangeValue = true
				}
				i++
			}
		} else if strings.EqualFold(arg, "-z") {
			if len(os.Args) <= i+1 {
				missingValue = true
			} else {
				zero, err = strconv.Atoi(os.Args[i+1])
				if err != nil {
					invalidValue = true
				}
				if IsOutOfRange(zero, ZeroFilterLo, ZeroFilterHi) {
					outOfRangeValue = true
				}
				i++
			}
		} else {
			if fileName == "" {
				fileName = os.Args[i]
			} else {
				fmt.Printf("ERROR: Unknown parameter \"%s\".\n", os.Args[i])
				fmt.Printf("%s\n", HelpHint)
				return
			}
		}

		if missingValue {
			fmt.Printf("ERROR: Missing value for \"%s\".\n", os.Args[i])
			fmt.Printf("%s\n", HelpHint)
			return
		}

		if invalidValue {
			fmt.Printf("ERROR: The value \"%s\" is invalid for \"%s\".\n", os.Args[i], os.Args[i-1])
			fmt.Printf("%s\n", HelpHint)
			return
		}

		if outOfRangeValue {
			fmt.Printf("ERROR: The value \"%s\" is out of range for \"%s\".\n", os.Args[i], os.Args[i-1])
			fmt.Printf("%s\n", HelpHint)
			return
		}

		if bpcSet && cpfSet {
			fmt.Printf("ERROR: The parameters \"-bpc\" and \"-cpf\" are mutually exclusive and cannot be used together.\n")
			fmt.Printf("%s\n", HelpHint)
			return
		}
	}

	// Opening the input file
	inFile, err := os.Open(fileName)
	if err != nil {
		fmt.Printf("ERROR: File not found \"%s\".\n", fileName)
		return
	}
	fi, err := inFile.Stat()
	PanicIfError(err)
	if fi.IsDir() {
		fmt.Printf("ERROR: The supplied parameter \"%s\" is a directory, but a file was expected.\n", fileName)
		return
	}
	inFileSize := int(fi.Size())
	if IsOutOfRange(inFileSize, FileSizeLo, FileSizeHi) {
		fmt.Printf("ERROR: File too large. Maximum allowed size is %d MB.\n", FileSizeHi/1024/1024)
		return
	}

	// Reading the dictionary (i.e. csv file)
	dictRecordCount := 0
	var dictRecords [][]string
	csvFileContent, err := os.ReadFile(dictFileName)
	if err == nil {
		r := csv.NewReader(strings.NewReader(string(csvFileContent)))
		r.Comma = ';'
		r.Comment = '#'
		dictRecords, err = r.ReadAll()
		PanicIfError(err)
		dictRecordCount = len(dictRecords)
		// Verification
		if dictRecordCount > MaxDictRows {
			fmt.Printf("ERROR: Dictionary file \"%s\" contains %d rows, but the maximum allowed is %d.\n", dictFileName, dictRecordCount, MaxDictRows)
			return
		}
		columnCount := len(dictRecords[0])
		if columnCount != DictColumnCount {
			fmt.Printf("ERROR: Dictionary file \"%s\" contains %d columns, but it must contain %d columns.\n", dictFileName, columnCount, DictColumnCount)
			return
		}
	} else {
		fmt.Printf("WARNING: Dictionary file not found \"%s\".\n", dictFileName)
	}

	// Init
	blockBuf := make([]byte, BlockSize)
	blockFreqTable := make(map[string]PatternInfo)
	chunkFreqTable := make(map[string]PatternInfo)
	for i := 0; i <= gap; i++ {
		GapTable[i] = i
	}

	// Calculate chunk size from bpc
	chunkSize := blocksPerChunk * BlockSize

	if chunksPerFile != 0 {
		// Cpf is set
		if inFileSize/chunksPerFile < BlockSize {
			// Cpf is too large
			chunksPerFile = 0
			fmt.Printf("WARNING: The \"-cpf\" value is too large for this file. The file cannot be split into chunks smaller than 256 bytes. The \"-cpf\" parameter will be ignored, continuing with \"-bpc=1\" instead.\n")
		} else {
			// Recalculate chunk size from cpf
			chunkSize = ((inFileSize / chunksPerFile) / BlockSize) * BlockSize
			if chunkSize == 0 {
				// Fixup chunk size as per one chunk per file
				chunkSize = (inFileSize / BlockSize) * BlockSize
			}

			// Calculate bpc
			blocksPerChunk = chunkSize / BlockSize

			if IsOutOfRange(blocksPerChunk, BlocksPerChunkLo, BlocksPerChunkHi) {
				fmt.Printf("ERROR: The calculated BPC exceeds the maximum allowed value. Try increasing the \"-cpf\" value.\n")
				return
			}
		}
	}

	// Set block/chunk mode for printing
	blockMode := (chunkSize == BlockSize)

	// Print header
	fmt.Printf("Offset   Size     Pattern      Ascii Bpc Freq Dict\n")

	actualChunkSize := 0
	for fileOffs := 0; fileOffs < inFileSize; fileOffs += BlockSize {
		// Read block of data
		bytesRead, err := inFile.Read(blockBuf)
		PanicIfError(err)

		// Build pattern frequency table for the block
		for gapIdx := 0; gapIdx <= gap; gapIdx++ {
			for bufIdx := 0; bufIdx < bytesRead-1-GapTable[gapIdx]; bufIdx++ {
				if zero == ZeroFilterBoth {
					// Exclude 00 00 patterns
					if (blockBuf[bufIdx] == 0x00) && (blockBuf[bufIdx+GapTable[gapIdx]+1] == 0x00) {
						continue
					}
				} else if zero == ZeroFilterEither {
					// Exclude patterns containing 00
					if (blockBuf[bufIdx] == 0x00) || (blockBuf[bufIdx+GapTable[gapIdx]+1] == 0x00) {
						continue
					}
				}
				key := fmt.Sprintf("%02x +(%d) %02x \n", blockBuf[bufIdx], GapTable[gapIdx], blockBuf[bufIdx+GapTable[gapIdx]+1])
				hits := blockFreqTable[key].Hits + 1
				blockFreqTable[key] = PatternInfo{
					First:  blockBuf[bufIdx],
					Second: blockBuf[bufIdx+GapTable[gapIdx]+1],
					Gap:    GapTable[gapIdx],
					Hits:   hits}
			}
		}

		// Get the top pattern of the block
		topHits := 0
		topKey := ""
		for k, v := range blockFreqTable {
			if v.Hits > topHits {
				topHits = v.Hits
				topKey = k
			}
		}

		// If there are multiple patterns with the same hits, choose deterministically
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

		// Threshold applies to blocks only
		if blockFreqTable[topKey].Hits >= threshold {
			if blockFreqTable[topKey].Gap == 0 {
				hex = fmt.Sprintf("%02X %02X", blockFreqTable[topKey].First, blockFreqTable[topKey].Second)
			} else {
				hex = fmt.Sprintf("%02X +(%d) %02X", blockFreqTable[topKey].First, blockFreqTable[topKey].Gap, blockFreqTable[topKey].Second)
			}

			printable = ToPrintable(blockFreqTable[topKey].First, blockFreqTable[topKey].Second)
			printable = "|" + printable + "|"

			for i := 0; i < dictRecordCount; i++ {
				if strings.EqualFold((dictRecords[i])[0], hex) {
					dict = strings.Trim((dictRecords[i])[1], " ")
				}
			}

			hitFreq = strconv.Itoa(blockFreqTable[topKey].Hits)
		}

		// Update pattern frequency table for the chunk with the top pattern of the block
		hits := chunkFreqTable[hex].Hits + 1

		chunkFreqTable[hex] = PatternInfo{
			First:  blockFreqTable[topKey].First,
			Second: blockFreqTable[topKey].Second,
			Gap:    blockFreqTable[topKey].Gap,
			Hits:   hits}

		if blockMode {
			// Block mode
			fmt.Printf("%08X %08X %-12s %5s %3d %4s %s\n", fileOffs, bytesRead, hex, printable, blocksPerChunk, hitFreq, dict)
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

			// Print chunk info
			//   when all the blocks in the chunk are processed OR
			//   if this is the last chunk
			if ((fileOffs != 0) && ((fileOffs+BlockSize)%chunkSize == 0)) || lastChunk {
				// Get the top block of the chunk
				topHits := 0
				topKey := ""
				for k, v := range chunkFreqTable {
					if v.Hits > topHits {
						topHits = v.Hits
						topKey = k
					}
				}
				// If there are multiple blocks with the same hits, choose deterministically
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
						hex = fmt.Sprintf("%02X %02X", chunkFreqTable[topKey].First, chunkFreqTable[topKey].Second)
					} else {
						hex = fmt.Sprintf("%02X +(%d) %02X", chunkFreqTable[topKey].First, chunkFreqTable[topKey].Gap, chunkFreqTable[topKey].Second)
					}

					printable = ToPrintable(chunkFreqTable[topKey].First, chunkFreqTable[topKey].Second)
					printable = "|" + printable + "|"
				}
				dict = "-"
				for i := 0; i < dictRecordCount; i++ {
					if strings.EqualFold((dictRecords[i])[0], topKey) {
						dict = strings.Trim((dictRecords[i])[1], " ")
					}
				}
				hitFreq = strconv.Itoa(topHits)

				// Offset is always the multiple of chunk size if cpf>1
				offset := (fileOffs + bytesRead) - actualChunkSize

				if firstChunk && lastChunk {
					offset = 0
				}

				actualBpc := actualChunkSize / BlockSize

				if actualChunkSize%BlockSize != 0 {
					// The block in the chunk is smaller than 256 bytes, it is still counted as a block
					actualBpc++
				}

				fmt.Printf("%08X %08X %-12s %5s %3d %4s %s\n", offset, actualChunkSize, hex, printable, actualBpc, hitFreq, dict)

				// Clear chunk frequency table for the next chunk
				for k := range chunkFreqTable {
					delete(chunkFreqTable, k)
				}

				// Reset actual chunk size
				actualChunkSize = 0
			} // if - Print chunk info
		} // else - Chunk mode

		// Clear block frequency table for the next block
		for key := range blockFreqTable {
			delete(blockFreqTable, key)
		}
	} // for - Read block of data
}
