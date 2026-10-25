# wav_lossy_enc

Compress WAVE files.

How to build:

```sh
go build -o wav_lossy_enc .
```

How to run:

```sh
./wav_lossy_enc [--block-size block size] <input file path> <output file path>
```

The block size argument must be a power of two, and must be less than or equal to 32768.
