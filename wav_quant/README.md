# wav_quant

Removes bits from WAVE audio file

How to build:

```sh
go build -o wav_quant .
```

How to run:

```sh
./wav_quant <path to wav> <bits to subtract>
```

The output is stored in a file named `quantized.wav`.
