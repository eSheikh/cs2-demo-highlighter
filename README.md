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
  - `headshot_kill`
  - `round_multikill`
  - `clutch_win`
- Highlight type filtering (`--types`)
- Flexible render targets: any set of highlight types as either **clips** (one recording per segment) or a **montage** (one continuous recording with jump cuts)
- HLAE script generation based on `mirv_streams` (without `startmovie`)
- POV lock using `spec_player <slot>`
- Pre-roll and post-roll segment extension
- Automatic jumps between segments (`demo_pause -> demo_gototick -> demo_resume`)
- Optional in-recording jumps for `round_multikill` when kill gaps are large

## Outputs

- `highlights.json`: normalized highlight metadata
- One `.cfg` per render target (see [Render targets](#render-targets)). By default a single clips script covering every highlight type.

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
  --clips highlights.cfg
```

Run tests:

```bash
go test ./...
```

## Interactive mode (TUI)

An interactive terminal UI walks through demo path → player selection → parsing (with a live progress bar) → highlight-type selection → cfg generation:

```bash
go run ./cmd/tui /path/to/match.dem
```

The demo path argument is optional and only pre-fills the first field. In the results screen, `space` toggles highlight types and `c` / `m` write a clips / montage `.cfg` for the selected types.

## Render targets

Recording output is configured with repeatable `--clips` and `--montage` flags. Each flag produces one `.cfg` file and has the form:

```
[types=]path.cfg
```

- `types` — comma-separated highlight types (omit, or use `all`, for every type). The value is split on the **first** `=`, so Windows drive-letter paths (`C:\...`) are preserved.
- `path.cfg` — output script path. Its base name is also used as a trailing segment of the `mirv_streams record name`, so multiple targets record into distinct folders.

If neither flag is given, the tool defaults to a single clips target of all types written to `highlights.cfg`.

Examples:

```bash
# Clips of every highlight (default behavior, explicit)
go run ./cmd/highlighter ... --clips highlights.cfg

# Separate outputs in one run: clutch clips + a headshot montage
go run ./cmd/highlighter ... \
  --clips clutch_win,wallbang=clutches.cfg \
  --montage headshot_kill=headshots.cfg

# A montage of every smoke kill, and another of every noscope
go run ./cmd/highlighter ... \
  --montage kill_in_smoke=smokes.cfg \
  --montage noscope=noscopes.cfg
```

## CLI Reference

| Flag              | Default            | Description                                                                               |
| ----------------- | ------------------ | ----------------------------------------------------------------------------------------- |
| `--demo`          | -                  | Path to input `.dem` file (required)                                                      |
| `--steamid`       | -                  | Target SteamID64 (required, 17 digits)                                                    |
| `--out`           | `highlights.json`  | Output JSON path (empty disables JSON output)                                             |
| `--types`         | (all)              | Comma-separated highlight types kept in the result (empty/`all` = every type)             |
| `--clips`         | `highlights.cfg`   | Clips render target `[types=]path.cfg` (repeatable)                                        |
| `--montage`       | -                  | Montage render target `[types=]path.cfg` (repeatable)                                      |
| `--hlae-path`     | current directory  | Output directory used in `mirv_streams record name`                                       |
| `--hlae-preset`   | `afxFfmpegYuv420p` | HLAE FFmpeg preset                                                                        |
| `--hlae-fps`      | `60`               | Recording frame rate                                                                      |
| `--hlae-preroll`  | `3`                | Seconds added before each event                                                           |
| `--hlae-postroll` | `2`                | Seconds added after each event                                                            |
| `--hlae-kill-gap` | `10`               | Seconds between kills in `round_multikill` to trigger an in-recording jump (`0` disables) |

Disable JSON output:

```bash
go run ./cmd/highlighter ... --out ""
```

## Recording Workflows

### Clips (one recording per segment)

1. Launch CS2 through HLAE.
2. Load demo: `playdemo <demo_name>`.
3. Paste the clips `.cfg` into the HLAE console.
4. Wait for `All N segments recorded`.
5. Script ends with `disconnect` and returns to the main menu.

Result: multiple output files, one per segment.

### Montage (one continuous recording)

1. Load the same demo.
2. Paste the montage `.cfg`.
3. Recording starts once, jumps across the selected segments, then stops once.
4. Script ends with `disconnect` and returns to the main menu.

Result: one montage-oriented output file.

## Generated File Examples

### `highlights.json`

Rounds are 1-based (round 1 is the first round).

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
      "time_start_sec": 1754.03,
      "time_end_sec": 1759.53,
      "kills": 3,
      "kill_ticks": [112258, 112430, 112610],
      "victims": ["7656119XXXXXXXXXX", "7656119XXXXXXXXXX", "7656119XXXXXXXXXX"],
      "weapon": "M4A1",
      "player_slot": 10,
      "steamid": "7656119XXXXXXXXXX",
      "demo": "mirage.dem",
      "segment_tick_start": 112258,
      "segment_tick_end": 112610
    }
  ]
}
```

### Clips `.cfg` (abridged)

The setup block is emitted once, followed by per-segment `mirv_cmd addAtTick` lines. Output is comment-free because the CS2/HLAE console can break on comment lines.

```cfg
mirv_cvar_unhide_all;
mirv_cmd clear;
mirv_streams record end;
mirv_streams record name "<hlae-path>/<steamid>/<date>/<target>";
mirv_streams settings edit afxDefault settings afxFfmpegYuv420p;
mirv_streams record fps 60;
...

mirv_cmd addAtTick 112066 "spec_player 10; host_framerate 60; mirv_streams record start";
mirv_cmd addAtTick 112738 "mirv_streams record end; host_framerate 0";
```

The `112066`/`112738` ticks are the JSON example's `112258`/`112610` extended by the 3s pre-roll and 2s post-roll (at 64 tick), and `spec_player 10` matches its `player_slot`.

## Validation and Error Handling

- Fail-fast config validation before parsing:
  - empty demo path
  - invalid extension (non-`.dem`)
  - missing / non-regular / empty file
  - invalid SteamID64 format
  - unknown highlight type in `--types` / render targets
  - leading/trailing spaces in CLI string flags are trimmed before validation and execution
- Parser safety behavior:
  - defensive demo-path validation
  - parser panic conversion to regular error
  - explicit error for truncated/corrupted demos
  - `context` cancellation support

## Architecture

- `cmd/highlighter`: CLI entrypoint
- `internal/bootstrap`: flag parsing and the CLI run (file output lives here)
- `internal/engine`: I/O-free core — roster listing, parse + highlight extraction, parse-progress streaming
- `internal/parser`: demo event extraction (`demoinfocs`)
- `internal/service`: highlight rules and domain logic
- `internal/hlae`: render targets, segment planning, script rendering
- `internal/repository`: persistence layer
- `internal/model`: shared types

## Limitations

- Output quality depends on demo integrity and parser event fidelity.
- Clutch detection is rule-based, not model/vision-based.
- Montages are playback jump-cut automation, not NLE post-production.

## Roadmap

1. New highlight types (`awp_flick`, `360`, etc.).
2. Automated HLAE launch/recording (`recorder`).
3. Add audio to recorded highlight videos.

## License

This project is licensed under the MIT License. See [LICENSE](./LICENSE).
