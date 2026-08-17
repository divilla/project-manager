package main

import (
	"apih/skeleton/internal/definition"
	"apih/skeleton/internal/domain"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

var InvalidPathError = errors.New("invalid path")

func main() {
	workDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	if len(os.Args) > 1 {
		subdir := os.Args[1]
		path := filepath.Join(workDir, subdir)
		info, err := os.Stat(path)
		if err != nil {
			log.Fatal(fmt.Errorf("%w: %w", InvalidPathError, err))
		}
		if !info.IsDir() {
			log.Fatal(fmt.Errorf("%w: %v", InvalidPathError, path))
		}
		workDir = path
	}
	fmt.Printf("Working Directory: %s\n\n", workDir)

	ctx, cancel := context.WithCancel(context.Background())
	_ = cancel

	suite := &domain.Suite{WorkDir: workDir}

	loader := definition.NewLoader()
	if err = loader.LoadDirectoryStructure(ctx, suite); err != nil {
		log.Fatal(err)
	}
	if err = loader.LoadDirectoryFiles(ctx, suite); err != nil {
		log.Fatal(err)
	}
	if err = loader.DecodeBaseDefinitions(ctx, suite); err != nil {
		log.Fatal(err)
	}

	decoder := definition.NewDecoder()
	if err = decoder.DecodeFiles(ctx, suite); err != nil {
		log.Fatal(err)
	}
	if err = decoder.ValidateDefaultsDefinitions(ctx, suite); err != nil {
		log.Fatal(err)
	}
	if err = decoder.ValidateStepsDefinitions(ctx, suite); err != nil {
		log.Fatal(err)
	}

	resolver := definition.NewResolver()
	if err = resolver.ResolveDefaults(ctx, suite); err != nil {
		log.Fatal(err)
	}
	if err = resolver.ResolveSteps(ctx, suite); err != nil {
		log.Fatal(err)
	}
}
