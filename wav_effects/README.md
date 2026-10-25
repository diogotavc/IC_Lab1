# wav_effects

Applies effect to a WAVE file.

How to build:

```sh
go build -o wav_effects .
```

How to run

```sh
./wav_effects [-echo <echo delay time> -echoC <echo coefficient>] [-multiEcho <multi echo delay time> -multiEchoC <multi echo coefficient>] [-ampMod <amplitude for modulation>] [-tvd <delay>]
```

If a option is given, an output file is written to the current directory:

- The echo effect is output into the file `echo.wav`
- The multi echo effect is output into the file `multiEcho.wav`
- The amplitude modulation effect is output into the file `ampMod.wav`
- The delay effect is output into the file `tvd.wav`
