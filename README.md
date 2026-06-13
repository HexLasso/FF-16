# FF-16 (Find Frequent 16-bit)

## Intro

The purpose of FF-16 is to assist in identifying structural redundancies in a given file.

Structural redundancies are more likely to be understood by users than coincidental redundancies.

Understanding structural redundancies can further assist in understanding the layout of the file.

In order to do this, **FF-16 finds frequent 16-bit patterns in all regions of the file.**

Finding 16-bit patterns is computationally manageable on desktops.

Byte-boundary matching can highlight structural redundancies with negligible coincidental redundancies in common data sets.

FF-16 is a CLI tool written in Go.

## Command line syntax

```
ff-16 [filename] [-d <filename>] [<-bpc <1..256>|-cpf <1..65536>>] [-g <0..127>] [-t <1..255>]

  <filename>      Target file
  -d <filename>   Dictionary file  (Default: dict.csv)
  -bpc <1..256>   Blocks per chunk (Default: 1)
  -cpf <1..65536> Chunks per file  (Default: not specified)
  -g <0..127>     Max gaps         (Default: 31)
  -t <1..255>     Freq threshold   (Default: 5)
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

## File, chunk, block and pattern

A file consist of chunks. A chunk consists of blocks.

```
([<Blk><Blk><Blk>][<Blk><Blk><Blk>][<Blk><Blk><Blk>][<Blk><Blk><Blk>])

() File
[] Chunk
<> Block
```

The structure of a pattern is like this.
```
00 +(01) 00
|    |   |
|    |   Second byte of pattern
|    Gap in bytes
First byte of pattern
```
