// Copyright (2022 -- present) Shahruk Hossain <shahruk10@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//		 http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// ==============================================================================

package main

import (
	"context"
	"flag"
	"os"

	"github.com/peterbourgon/ff/v3/ffcli"
	log "github.com/sirupsen/logrus"

	"github.com/shahruk10/go-sctk/cmd/sctk/score"
)

func main() {
	var (
		rootFlagSet = flag.NewFlagSet("sctk", flag.ExitOnError)
	)

	// Setting logger format.
	log.SetFormatter(&log.TextFormatter{
		DisableTimestamp:       true,
		DisableLevelTruncation: true,
		DisableQuote:           true,
		PadLevelText:           true,
	})

	root := &ffcli.Command{
		ShortUsage: "sctk [flags] <subcommand>",
		ShortHelp:  "sctk wraps several functionality provided by the SCTK toolkit.",
		FlagSet:    rootFlagSet,
	}

	// subcommands
	root.Subcommands = []*ffcli.Command{
		score.Cmd(),
	}

	if err := root.Parse(os.Args[1:]); err != nil {
		root.FlagSet.Usage()
		os.Exit(1)
	}

	// running command
	if err := root.Run(context.Background()); err != nil {
		log.WithFields(log.Fields{"error": err}).Error("failed to run command")
		os.Exit(1)
	}
}
