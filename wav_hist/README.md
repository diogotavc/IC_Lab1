# wav_hist

Generate a histogram of the samples of a WAVE file.

How to build:

```sh
go build -o wav_hist .
```

How to run:

```sh
./wav_hist <path to wav file> [optional bin exponent]
```

It will output histograms as CSV files as `channelHistogramX.csv` for channel X.
Stereo files will also have their MID and SIDE histograms output on
`minHistogram.csv` and `sideHistogram.csv`

Each CSV file has two columns:

- The start value for the bin
- The count
