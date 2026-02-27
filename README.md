# CS2 Demo Highlighter

> **Notice**
> This repository is public, but it is still a personal home project.
> Interfaces, output format details, and recording behavior may change.
> Bugs and edge cases are possible.
> Use it with caution and validate results before relying on them in production.

Russian version: [README.ru.md](./README.ru.md)

CS2 Demo Highlighter is a CLI tool that parses `.dem` files, extracts player-centric highlight events, and generates HLAE scripts for automated recording.

The project is focused on extraction and recording orchestration, not on a full post-production pipeline.

## Features

- Demo parsing via `github.com/markus-wa/demoinfocs-golang/v5`
- Highlight event detection:
  - `kill_in_smoke`
  - `kill_blinded`
  - `wallbang`
  - `noscope`
  - `round_multikill`
  - `clutch_win`
  - `headshot_kill`
  - `headshot_collection` (summary event)
- HLAE script generation based on `mirv_streams` (without `startmovie`)
- POV lock using `spec_player <slot>`
- Pre-roll and post-roll segment extension
- Automatic jumps between segments (`demo_pause -> demo_gototick -> demo_resume`)
- Optional in-recording jumps for `round_multikill` when kill gaps are large

## Outputs

The tool can generate:

- `highlights.json`: normalized highlight metadata
- `highlights.cfg`: HLAE script for regular highlight clips
- `headshots.cfg`: HLAE script for one-file headshot montage with jump cuts

## Requirements

- Go `1.26+`
- A valid non-empty CS2 `.dem` file
- Target player SteamID64 (17 digits)
- HLAE setup for CS2 recording (AfxHookSource2)

## Installation

```bash
git clone https://github.com/eSheikh/cs2-demo-highlighter.git
cd cs2-demo-highlighter
go mod download
```

## Quick Start

```bash
go run ./cmd/highlighter \
  --demo /path/to/match.dem \
  --steamid 7656119XXXXXXXXXX \
  --out highlights.json \
  --hlae highlights.cfg \
  --hlae-headshots headshots.cfg \
  --hlae-path highlights \
  --hlae-preset afxFfmpegYuv420p
```

Run tests:

```bash
go test ./...
```

## CLI Reference

| Flag                    | Default               | Description                                                                               |
| ----------------------- | --------------------- | ----------------------------------------------------------------------------------------- |
| `--demo`                | -                     | Path to input `.dem` file (required)                                                      |
| `--steamid`             | -                     | Target SteamID64 (required, 17 digits)                                                    |
| `--out`                 | `highlights.json`     | Output JSON path (empty disables JSON output)                                             |
| `--hlae`                | `highlights.cfg`      | Output path for regular HLAE script                                                       |
| `--hlae-headshots`      | `headshots.cfg`       | Output path for headshot montage HLAE script                                              |
| `--hlae-headshots-name` | `headshot_collection` | Recording name for montage output                                                         |
| `--hlae-path`           | `highlights`          | Prefix used in `mirv_streams record name`                                                 |
| `--hlae-preset`         | `afxFfmpegYuv420p`    | HLAE FFmpeg preset                                                                        |
| `--hlae-fps`            | `60`                  | Recording frame rate                                                                      |
| `--hlae-preroll`        | `3`                   | Seconds added before each event                                                           |
| `--hlae-postroll`       | `2`                   | Seconds added after each event                                                            |
| `--hlae-kill-gap`       | `10`                  | Seconds between kills in `round_multikill` to trigger an in-recording jump (`0` disables) |

Disable headshot montage script generation:

```bash
go run ./cmd/highlighter ... --hlae-headshots ""
```

Disable JSON output:

```bash
go run ./cmd/highlighter ... --out ""
```

## Recording Workflows

### Regular highlights (`highlights.cfg`)

1. Launch CS2 through HLAE.
2. Load demo: `playdemo <demo_name>`.
3. Paste `highlights.cfg` into the HLAE console.
4. Wait for `All N segments recorded`.
5. Script ends with `disconnect` and returns to main menu.

Result: multiple output files, one per segment.

### Headshot montage (`headshots.cfg`)

1. Load the same demo.
2. Paste `headshots.cfg`.
3. Recording starts once, jumps across headshot segments, then stops once.
4. Script ends with `disconnect` and returns to main menu.

Result: one montage-oriented output file.

## Generated File Examples

### `highlights.json`

```json
{
  "demo": "mirage.dem",
  "steamid": "7656119XXXXXXXXXX",
  "tick_rate": 64,
  "highlights": [
    {
      "type": "round_multikill",
      "round": 16,
      "tick_start": 112258,
      "tick_end": 112610,
      "kills": 3,
      "weapon": "M4A1",
      "player_slot": 10
    }
  ]
}
```

### `highlights.cfg`

```cfg
mirv_streams settings edit afxDefault settings afxFfmpegYuv420p;
mirv_streams record fps 60;
spec_show_xray 0;

mirv_cmd addAtTick 112066 "spec_player 10; host_framerate 60; mirv_streams record name highlights_hl_0005_r16_round_multikill; mirv_streams record start";
mirv_cmd addAtTick 112738 "mirv_streams record end; host_framerate 0";
mirv_cmd addAtTick 112739 "demo_pause; demo_gototick 118230; spec_player 10; demo_resume";
```

### `headshots.cfg`

```cfg
mirv_streams settings edit afxDefault settings afxFfmpegYuv420p;
mirv_streams record fps 60;

mirv_cmd addAtTick 26746 "spec_player 10; host_framerate 60; mirv_streams record name highlights_headshot_collection; mirv_streams record start";
mirv_cmd addAtTick 27067 "demo_pause; demo_gototick 59674; spec_player 10; demo_resume";
mirv_cmd addAtTick 118664 "mirv_streams record end; host_framerate 0";
```

## Validation and Error Handling

- Fail-fast config validation before parsing:
  - empty demo path
  - invalid extension (non-`.dem`)
  - missing / non-regular / empty file
  - invalid SteamID64 format
  - leading/trailing spaces in CLI string flags are trimmed before validation and execution
- Parser safety behavior:
  - defensive demo-path validation
  - parser panic conversion to regular error
  - explicit error for truncated/corrupted demos
  - `context` cancellation support

## Architecture

- `cmd/highlighter`: CLI entrypoint
- `internal/bootstrap`: config parsing and pipeline bootstrapping
- `internal/parser`: demo event extraction (`demoinfocs`)
- `internal/service`: highlight rules and domain logic
- `internal/hlae`: segment planning and script rendering
- `internal/repository`: persistence layer

## Limitations

- Output quality depends on demo integrity and parser event fidelity.
- Clutch detection is rule-based, not model/vision-based.
- Headshot montage is playback jump-cut automation, not NLE post-production.

## Roadmap

1. New highlight types (`awp_flick`, `360`, etc.).
2. Highlight rule selection.
3. Add audio to recorded highlight videos.

## License

This project is licensed under the MIT License. See [LICENSE](./LICENSE).
