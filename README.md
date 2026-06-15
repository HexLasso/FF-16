# FF-16 (Find Frequent 16-bit)

## What does FF-16 do?

FF-16 is a static analysis tool that finds frequently occurring local 16-bit patterns across the entire file. It can help to understand file layout and locate structures from frequent patterns.

## Command line usage

```
ff-16 [filename] [-d <filename>] [<-bpc <1..256>|-cpf <1..65536>>] [-g <0..127>] [-t <1..255>]

  <filename>      Target file
  -d <filename>   Dictionary file  (Default: dict.csv)
  -bpc <1..256>   Blocks per chunk (Default: 1)
  -cpf <1..65536> Chunks per file  (Default: not specified)
  -g <0..127>     Max gaps         (Default: 31)
  -t <1..255>     Freq threshold   (Default: 5)
```

## Understanding the results

The `Offset` and `Size` columns indicate the data region in the file.

The `Pattern` column shows the frequent pattern.

The `Ascii` column shows the text representation of the pattern, or `.` if not printable.

The `Bpc` column indicates whether the result is displayed per block (`Bpc = 1`) or summarized into chunks (`Bpc > 1`).

The `Dict` column shows the corresponding text for the pattern from the dictionary.

### Block (`Bpc = 1`)

The `Freq` column indicates the number of pattern hits in the data region. The minimum value is as per defined by Freq threshold (`-t`), and the maximum value is 255. Example output:

```
Offset   Size     Pattern      Ascii Bpc Freq Dict
00000000 00000100 00 00         |..|   1   77 -
00000100 00000100 00 00         |..|   1  127 -
00000200 00000100 00 00         |..|   1  165 -
00000300 00000100 00 00         |..|   1  255 -
00000400 00000100 00 00         |..|   1   13 -
00000500 00000100 01 64         |.d|   1   10 -
00000600 00000100 CC CC         |..|   1   33 -
00000700 00000100 00 00         |..|   1   22 -
00000800 00000100 00 00         |..|   1   34 -
```

### Chunk (`Bpc > 1`)

The `Freq` column indicates how many blocks in a chunk contain the pattern. The maximum value is `Bpc`. Example output:

```
Offset   Size     Pattern      Ascii Bpc Freq Dict
00000000 00002D00 00 00         |..|  45   13 -
00002D00 00002D00 CC CC         |..|  45   33 -
00005A00 00002D00 00 00         |..|  45   19 -
00008700 00002D00 00 +(23) 00   |..|  45   20 -
0000B400 00002D00 00 +(3) 00    |..|  45   24 -
0000E100 00002D00 00 +(3) 00    |..|  45   35 -
00010E00 00002D00 00 +(3) 00    |..|  45   34 -
00013B00 00002D00 00 +(3) 00    |..|  45   23 -
00016800 00002D00 FF +(27) FF   |..|  45   17 -
00019500 00002D00 00 +(3) 00    |..|  45   23 -
```

## Examples

### Simple

The simplest way to run an analysis.

```
go run ff-16.go .\sample.bin
```

### Using `-bpc`

Using the `-bpc` argument to set the number of blocks per chunk to `4` and control output granularity.

```
go run ff-16.go .\sample.bin -bpc 4
```

### Using `-cpf`

Using the `-cpf` argument to combine block results into `40` chunks and control total output length.

```
go run ff-16.go .\sample.bin -cpf 40
```

### Dictionary usage

In the dictionary file (e.g. `mydict.csv`), each line contains a pattern and the text to be displayed for that pattern, separated by `;`.

```
00 +(7) 00; QWORD
00 +(3) 00; DWORD
CC CC; INT 3
```

The dictionary file is specified with the `-d` argument.

```
go run ff-16.go .\sample.bin -d .\mydict.csv
```

## Terminologies

| Term | Description |
| --- | --- |
| Block | A block consists of a sequence of bytes. A block is always 256 bytes in size, except for the last block, if the file size is not a multiple of 256. |
| Chunk | A chunk consists of one block or a sequence of blocks. |
| Pattern | A pattern is a frequent two-byte data sequence in a block, with or without a gap between the bytes. |
| Gap | The number of bytes to skip between the two bytes in the pattern. |
| Blocks Per Chunk (BPC) | The number of blocks in a chunk. |
| Chunks Per File (CPF) | The number of chunks in the file. |
| Pattern frequency | The number of occurrences of a given pattern in a block. |
| Frequency threshold | The threshold defines the minimum statistically significant pattern frequency. |
| Dictionary | A list of pattern-description pairs in a user-editable CSV file used for pattern lookup. |

## Use cases

* Quick understanding of the layout of the file
* Finding redundancy in high-entropy data
