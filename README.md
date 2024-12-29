# synth

## play

Plays test tones.

```bash
sudo dnf install alsa-lib-devel
go run ./cmd/play
```

## synth

Generate test tones, and save them to a WAV file.

```bash
go run ./cmd/synth beep.wav
```
