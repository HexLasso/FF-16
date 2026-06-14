# FF-16 (Find Frequent 16-bit)

## What FF-16 does?

FF-16 sequentially splits the file into blocks and finds frequently occurring 16-bit patterns within each block. It can aggregate results into chunks to reduce output length while retaining full coverage. 

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

If no argument is provided, FF-16 displays the usage (above).

For analysis, the target file is the only parameter that needs specifying, and the analysis will start with default values.

Blocks per chunk (`-bpc`) and Chunks per file (`-cpf`) parameters are mutually exclusive. Use only one. If none is used, the results will be displayed for each block (verbose). Otherwise, the results will be aggregated into chunks (reduced verbosity).

Block per chunk is 1 (`-bpc 1`). Result for each block. Verbose.
```
(<Blk><Blk><Blk><Blk><Blk><Blk><Blk><Blk><Blk><Blk><Blk><Blk>)
```

Block per chunk is not 1 (`-bpc` or `-cpf` is specified). Result for each chunk. Reduced verbosity.
```
([<Blk><Blk><Blk>][<Blk><Blk><Blk>][<Blk><Blk><Blk>][<Blk><Blk><Blk>])

() File
[] Chunk
<> Block
```
If Max gap is not specified (`-g`), FF-16 will operate with `-g 31`. It means the search will involve any gaps between 0 and 31 between the first byte of the pattern and the second byte of the pattern. That is to find up to 32 bytes long structures.

This is how a pattern looks like.

```
00 +(31) 00
|    |   |
|    |   Second byte of pattern
|    Gap in bytes
First byte of pattern
```

Frequency threshold means if the pattern occurs at least that many times as specified, it's considered statistically significant and will be ranked against the other patterns. What is the point of this? For example, if a given block is a high-entropy block, it controls the coincidental matches, i.e., the noise. But if the user wants to see coincidental redundancy in high-entropy data, they may lower the threshold. For another example, the user may decide they only want to see very strong signals and increase the threshold.

Using a dictionary (`-d`) can be useful if the user wants to display a text next to the pattern. The dictionary file is a CSV file. Each line defines the pattern and the text to be displayed with semicolon between them. For example:
```
00 +(7) 00; QWORD
00 +(3) 00; DWORD
CC CC; INT 3
```

## Terminologies

| Term | Description |
| --- | --- |
| Block | A block consists of a sequence of bytes. A block is always 256 bytes in size, except for the last block, if the file size is not a multiple of 256. |
| Chunk | A chunk consists of one block or a sequence of blocks. |
| Pattern | A pattern is a frequent two-byte data sequence in a block, with or without a gap between the bytes. |
| Gaps | The number of bytes to skip between the two bytes in the pattern. |
| Block Per Chunk (BPC) | The number of blocks in a chunk. |
| Chunk Per File (CPF) | The number of chunks in the file. |
| Pattern frequency | The number of occurrences of a given pattern in a block. |
| Frequency threshold | The boundary to define the statistically significant pattern frequency. |
| Dictionary | A list of pattern-description pairs in a user-editable CSV file in which the pattern is looked up. |

## Use cases

* Quick understanding of the layout of the file
* Finding redundancy in high-entropy data
