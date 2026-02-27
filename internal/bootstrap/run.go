package bootstrap

import (
	"context"
	"errors"
	"flag"
	"log"

	"github.com/eSheikh/cs2-demo-highlighter/internal/hlae"
	"github.com/eSheikh/cs2-demo-highlighter/internal/model"
	"github.com/eSheikh/cs2-demo-highlighter/internal/parser/demoinfocs"
	"github.com/eSheikh/cs2-demo-highlighter/internal/repository/jsonrepo"
	"github.com/eSheikh/cs2-demo-highlighter/internal/service"
)

func Run(ctx context.Context, args []string, logger *log.Logger) error {
	cfg, err := ParseConfig(args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	orchestrator := service.NewOrchestrator(
		demoinfocs.NewParser(),
		service.NewHighlightService(),
	)

	result, err := orchestrator.Build(ctx, cfg.DemoPath, cfg.SteamID)
	if err != nil {
		return err
	}

	if err := jsonrepo.New(cfg.OutputPath).Save(ctx, result); err != nil {
		return err
	}
	logOutputSaved(logger, cfg.OutputPath)

	if err := writeHLAEScripts(cfg, result, logger); err != nil {
		return err
	}

	return nil
}

func writeHLAEScripts(cfg Config, result model.HighlightResult, logger *log.Logger) error {
	if !cfg.HLAE.Enabled() {
		return nil
	}

	if err := writeHLAEScriptFile(cfg.HLAE.ScriptPath, hlae.BuildScript(result, cfg.HLAE), logger); err != nil {
		return err
	}
	if !cfg.HLAE.HeadshotMontageEnabled() {
		return nil
	}

	return writeHLAEScriptFile(cfg.HLAE.HeadshotMontageScriptPath, hlae.BuildHeadshotMontageScript(result, cfg.HLAE), logger)
}

func writeHLAEScriptFile(path string, content string, logger *log.Logger) error {
	if err := writeTextFile(path, content); err != nil {
		return err
	}
	logOutputSaved(logger, path)
	return nil
}

func logOutputSaved(logger *log.Logger, outputPath string) {
	if logger == nil {
		return
	}
	logger.Printf("saved %s", outputPath)
}
