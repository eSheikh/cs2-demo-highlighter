package demoinfocs

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	demoparser "github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs"
	"github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/common"
	"github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/events"

	"github.com/eSheikh/cs2-demo-highlighter/internal/demo"
	"github.com/eSheikh/cs2-demo-highlighter/internal/model"
)

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) Parse(ctx context.Context, demoPath string, steamID string) (model.ParsedDemo, error) {
	if err := demo.ValidatePath(demoPath); err != nil {
		return model.ParsedDemo{}, err
	}

	file, err := os.Open(demoPath)
	if err != nil {
		return model.ParsedDemo{}, fmt.Errorf("open demo file: %w", err)
	}
	defer file.Close()

	parser := demoparser.NewParser(file)
	defer parser.Close()

	result := newParsedDemo(demoPath)
	roundWinners := make(map[int]common.Team)
	registerHandlers(parser, steamID, &result, roundWinners)

	if err := parseDemo(ctx, parser); err != nil {
		return model.ParsedDemo{}, err
	}
	if result.TickRate <= 0 {
		result.TickRate = parser.TickRate()
	}
	applyRoundWinners(result.Kills, roundWinners)

	return result, nil
}

func newParsedDemo(demoPath string) model.ParsedDemo {
	return model.ParsedDemo{
		Demo:     filepath.Base(demoPath),
		TickRate: 0,
		Kills:    make([]model.KillEvent, 0),
	}
}

func registerHandlers(
	parser demoparser.Parser,
	steamID string,
	result *model.ParsedDemo,
	roundWinners map[int]common.Team,
) {
	parser.RegisterEventHandler(func(e events.TickRateInfoAvailable) {
		result.TickRate = e.TickRate
	})

	parser.RegisterEventHandler(func(e events.RoundEnd) {
		roundWinners[parser.GameState().TotalRoundsPlayed()] = e.Winner
	})

	parser.RegisterEventHandler(func(e events.Kill) {
		kill, ok := buildKillEvent(parser, steamID, e)
		if !ok {
			return
		}
		result.Kills = append(result.Kills, kill)
	})
}

func buildKillEvent(parser demoparser.Parser, steamID string, e events.Kill) (model.KillEvent, bool) {
	if e.Killer == nil || e.Victim == nil {
		return model.KillEvent{}, false
	}
	if steamIDFromUint64(e.Killer.SteamID64) != steamID {
		return model.KillEvent{}, false
	}

	weaponName := ""
	if e.Weapon != nil {
		weaponName = e.Weapon.String()
	}

	killerTeam := e.Killer.Team
	alliesAlive, enemiesAlive := aliveCountsBeforeKill(parser.GameState().Participants(), killerTeam)

	return model.KillEvent{
		Tick:       parser.GameState().IngameTick(),
		Time:       parser.CurrentTime(),
		Round:      parser.GameState().TotalRoundsPlayed(),
		KillerID:   steamIDFromUint64(e.Killer.SteamID64),
		KillerSlot: killerSlotFromPlayer(e.Killer),
		VictimID:   steamIDFromUint64(e.Victim.SteamID64),
		Weapon:     weaponName,
		IsInSmoke:  e.ThroughSmoke,
		IsBlinded:  e.AttackerBlind,
		IsWallbang: e.PenetratedObjects > 0,
		IsNoScope:  e.NoScope,
		IsHeadshot: e.IsHeadshot,
		KillerTeam: int(killerTeam),

		AlliesAliveBefore:  alliesAlive,
		EnemiesAliveBefore: enemiesAlive,
	}, true
}

func parseDemo(ctx context.Context, parser demoparser.Parser) (err error) {
	if parser == nil {
		return errors.New("demo parser is nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if ctx.Err() != nil {
		return fmt.Errorf("demo parsing cancelled: %w", ctx.Err())
	}

	stopCancelWatcher := context.AfterFunc(ctx, parser.Cancel)
	defer stopCancelWatcher()

	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("demo parser panicked on invalid input: %v", recovered)
		}
	}()

	if err = parser.ParseToEnd(); err != nil {
		switch {
		case errors.Is(err, demoparser.ErrCancelled) && ctx.Err() != nil:
			return fmt.Errorf("demo parsing cancelled: %w", ctx.Err())
		case errors.Is(err, demoparser.ErrUnexpectedEndOfDemo):
			return fmt.Errorf("demo file is truncated or corrupted: %w", err)
		default:
			return fmt.Errorf("demo parsing failed: %w", err)
		}
	}

	return nil
}

func steamIDFromUint64(id uint64) string {
	return strconv.FormatUint(id, 10)
}

func killerSlotFromPlayer(p *common.Player) int {
	if p == nil {
		return 0
	}
	// In CS2 demos spec_player aligns with EntityID in practice.
	switch {
	case p.EntityID > 0:
		return p.EntityID
	// Fallback for edge-cases where only UserID is available.
	case p.UserID > 0:
		return p.UserID + 1
	default:
		return p.UserID
	}
}

func applyRoundWinners(kills []model.KillEvent, roundWinners map[int]common.Team) {
	for i := range kills {
		winner, exists := roundWinners[kills[i].Round]
		if !exists {
			continue
		}
		kills[i].RoundWon = winner == common.Team(kills[i].KillerTeam)
	}
}

func aliveCountsBeforeKill(participants demoparser.Participants, killerTeam common.Team) (allies int, enemies int) {
	allies = countAlivePlayers(participants.TeamMembers(killerTeam))
	enemies = countAlivePlayers(participants.TeamMembers(opponentTeam(killerTeam)))
	return allies, enemies
}

func countAlivePlayers(players []*common.Player) int {
	alive := 0
	for _, player := range players {
		if player == nil || !player.IsAlive() {
			continue
		}
		alive++
	}
	return alive
}

func opponentTeam(team common.Team) common.Team {
	switch team {
	case common.TeamTerrorists:
		return common.TeamCounterTerrorists
	case common.TeamCounterTerrorists:
		return common.TeamTerrorists
	default:
		return common.TeamUnassigned
	}
}
