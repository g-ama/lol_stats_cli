# lol_stats

A CLI tool written in Go for tracking your League of Legends ranked performance.

It pulls your recent solo/duo ranked games from the Riot Games API, scores each
one with a role-aware heuristic, and renders the results in your terminal —
either as a color-coded grid of your last 20 games, or as a detailed scorecard
for a single game.

## How it works

1. **Account lookup** — your Riot ID (`Username#Tagline`) is resolved to a PUUID
   and cached locally, so you only enter it once.
2. **Match fetch** — the match-v5 endpoint returns your recent ranked match IDs,
   and each match is fetched concurrently (one goroutine per match).
3. **Scoring** — for each game, your participant record is extracted and scored
   out of ~100 by a formula specific to your lane. Top, jungle, mid, bot, and
   support are weighted differently (e.g. supports are graded heavily on vision
   score and CC, ADCs on damage and CS/min).
4. **Persistence** — scored games are written to disk, so viewing stats is
   instant and doesn't consume API rate limit. You only re-fetch when you ask
   for it with `--load`.
5. **Display** — results are printed with ANSI colors.

### Score bands

The performance grid colors each game by its score:

| Color  | Score    | Meaning                          |
| ------ | -------- | -------------------------------- |
| Gray   | 0        | Unscored (unrecognized lane)     |
| Red    | 1–49     | Poor game                        |
| Yellow | 50–79    | Average game                     |
| Green  | 80+      | Strong game                      |

## Requirements

- Go 1.24.6 or later
- A Riot Games API key — get one at [developer.riotgames.com](https://developer.riotgames.com/)
- An account on the **Americas** routing region (the API host is hardcoded to
  `americas.api.riotgames.com`)

## Setup

Clone the repo and create a `.env` file in the project root:

```sh
echo 'API_KEY=RGAPI-your-key-here' > .env
```

`.env` is gitignored. Note that development keys from the Riot developer portal
expire every 24 hours, so you'll need to refresh this value.

Build the binary:

```sh
go build -o lol_stats
```

Or run without building:

```sh
go run . stats
```

## Usage

```
lol_stats stats [flags]

Flags:
  -l, --load        Fetch the last 20 ranked games from the Riot API
  -g, --game int    Show the detailed scorecard for a single game by index
  -h, --help        Help for stats
```

### First run

On your very first run with `--load`, you'll be prompted for your Riot ID:

```sh
./lol_stats stats --load
```

```
There is no account file, press Y/N to proceed and make one or to terminate
y
Username: YourName
Tagline: NA1
```

Your PUUID is saved after that and you won't be asked again.

### Viewing the performance grid

```sh
./lol_stats stats
```

Prints a 4×5 grid of your last 20 games, each cell showing the game index
colored by its score — a quick visual read on how your recent climb is going.

### Viewing a single game

Pass the index shown in the grid:

```sh
./lol_stats stats --game 7
```

```
════════════════════════════════════════════════════════
VICTORY
════════════════════════════════════════════════════════

Player: YourName#NA1
Champion: Jinx (Level 16)
Position: BOTTOM

Score: 87.4

━━━ Combat Stats ━━━
  K/D/A:  14 / 3 / 8  (KDA: 7.33)
  Damage Dealt:  38,204
  Damage Taken:  19,551
  CC Duration:   42.0s

━━━ Economy ━━━
  Gold Earned:  17,832
  CS:           284
  CS/min:       8.9

━━━ Vision & Support ━━━
  Vision Score:  21
```

### Refreshing your data

`--load` re-fetches and overwrites the stored history. Combine it with `--game`
to refresh and immediately inspect a game:

```sh
./lol_stats stats --load --game 3
```

## Stored files

State lives under `~/lol_stats/`:

| File           | Contents                                         |
| -------------- | ------------------------------------------------ |
| `account.json` | Your resolved PUUID and username                 |
| `history.json` | The last fetched batch of scored games           |

Delete `account.json` to re-run the account setup prompt with a different Riot ID.

## Project layout

```
main.go                      Entry point
cmd/
  root.go                    Root cobra command
  lol.go                     `stats` command, flags, account setup prompt
internal/
  api/client.go              Riot API calls (account, match IDs, matches)
  model/model.go             API response types and config struct
  stats/stats.go             Per-role performance scoring formulas
  persistence/persistence.go Reading/writing account and history JSON
  printer/printer.go         ANSI grid and single-game scorecard rendering
```

## Limitations

- Only ranked solo/duo (queue `420`) games are considered.
- Locked to the Americas routing region.
- Games where your lane resolves to `NONE` are skipped, and any lane outside
  the five standard positions scores `0`.
- The grid renderer expects exactly 20 stored games; a shorter history will
  cause it to fail.
- `--game 1` falls through to the grid view because of the `game > 1` check, so
  the first game can't currently be inspected individually.
