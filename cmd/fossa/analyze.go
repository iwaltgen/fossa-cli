package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/fossas/fossa-cli/module"
	logging "github.com/op/go-logging"
	"github.com/urfave/cli"
)

var analysisLogger = logging.MustGetLogger("analyze")

type analyzeConfig struct {
	output          bool
	allowUnresolved bool
	noUpload        bool
}

func analyzeCmd(c *cli.Context) {
	config, err := initialize(c)
	if err != nil {
		buildLogger.Fatalf("Could not load configuration: %s", err.Error())
	}
	if len(config.modules) == 0 {
		buildLogger.Fatal("No modules specified.")
	}

	analysis, err := doAnalyze(config.modules, config.analyzeConfig.allowUnresolved)
	if err != nil {
		analysisLogger.Fatalf("Analysis failed: %s", err.Error())
	}
	analysisLogger.Debugf("Analysis complete: %#v", analysis)

	if config.analyzeConfig.output {
		normalModules, err := normalizeAnalysis(analysis)
		if err != nil {
			mainLogger.Fatalf("Could not normalize build data: %s", err.Error())
		}
		buildData, err := json.Marshal(normalModules)
		if err != nil {
			mainLogger.Fatalf("Could not marshal analysis results: %s", err.Error())
		}
		fmt.Println(string(buildData))
	}

	if config.analyzeConfig.noUpload {
		fmt.Fprintln(os.Stderr, "Analysis succeeded!")
		return
	}

	err = doUpload(config, analysis)
	if err != nil {
		analysisLogger.Fatalf("Upload failed: %s", err.Error())
	}
}

type analysisKey struct {
	builder module.Builder
	module  module.Module
}

type analysis map[analysisKey][]module.Dependency

func doAnalyze(modules []moduleConfig, allowUnresolved bool) (analysis, error) {
	analysisLogger.Debugf("Running analysis on modules: %#v", modules)
	dependencies := make(analysis)

	for _, moduleConfig := range modules {
		builder, m, err := resolveModuleConfig(moduleConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve modules: " + err.Error())
		}

		err = builder.Initialize()
		if err != nil {
			return nil, fmt.Errorf("failed to initialize build: " + err.Error())
		}

		isBuilt, err := builder.IsBuilt(m, allowUnresolved)
		if err != nil {
			return nil, fmt.Errorf("could not determine whether module %#v is built: %#v", m.Name, err.Error())
		}
		if !isBuilt {
			return nil, fmt.Errorf("module " + m.Name + " does not appear to be built (try first running your build or `fossa build`, and then running `fossa`)")
		}

		deps, err := builder.Analyze(m, allowUnresolved)
		if err != nil {
			return nil, fmt.Errorf("analysis failed on module " + m.Name + ": " + err.Error())
		}
		dependencies[analysisKey{
			builder: builder,
			module:  m,
		}] = deps
	}

	return dependencies, nil
}
