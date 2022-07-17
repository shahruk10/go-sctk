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

// Package score implements subcommands to evaluate different scores such as
// word and character error rates for ASR hypotheses against reference
// transcripts.
package score

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/shahruk10/go-sctk/internal/score"
	"github.com/shahruk10/go-sctk/internal/sctk"
)

// Config for the score subcommand.
type Config struct {
	outDir     string
	refFile    string
	hypFiles   []sctk.Hypothesis
	fileFormat score.FileFormat
	normCfg    score.NormalizeConfig
	scliteCfg  sctk.ScliteCfg
}

// Cmd creates and returns a pointer to the ffcli.Command for the score
// subcommand
func Cmd() *ffcli.Command {
	cfg := Config{}
	fs := flag.NewFlagSet("sctk score", flag.ExitOnError)

	// Will parse these into config field with the correct type later.
	var (
		hypArgs   stringArray
		delimiter string
		quoteChar string
	)

	fs.StringVar(&cfg.outDir, "out", "",
		"(Required) Path to output directory where scores and reports will be written.\n")

	fs.StringVar(&cfg.refFile, "ref", "",
		"(Required) Path to file containing reference text.\n")

	fs.Var(&hypArgs, "hyp",
		`(Required) Path hypothesis file to score. It can either simply be the filepath
to the hypothesis file, or it can be in the form <name>,<filepath> where
<name> is identifies the system that generated the hypothesis. The <name> will
be used in the generated reports. If no name is provided, the name will be set
automatically. This argument may be provided multiple times to point to score
multiple hypotheses at once.
`)

	fs.StringVar(&delimiter, "delimiter", ",",
		`The delimiter used in reference and hypotheses files. By default, the program expects
comma delimited files (.csv) The program needs at least two columns per row, containing
<utteranceID> and <transcript>. By default, the first and second columns are assumed
to contain <utteranceID> and <transcript> respectively. This can be changed by using
--id-col and --trn-col arguments.
`)

	fs.StringVar(&quoteChar, "quote-char", "\"",
		`The character used to indicate the start and end of a block of text where any instances
of the delimiter character can be ignored.
`)

	fs.IntVar(&cfg.fileFormat.ColID, "col-id", 0,
		"The column index (zero based, positive only) containing <utteranceID>.\n")

	fs.IntVar(&cfg.fileFormat.ColTrn, "col-trn", 1,
		"The column index (zero based, positive only) containing <transcript>.\n")

	fs.BoolVar(&cfg.fileFormat.IgnoreFirstRow, "ignore-first", false,
		"If true, will ignore the first row in the provided files, assuming it is the header row.\n")

	fs.BoolVar(&cfg.scliteCfg.CER, "cer", false,
		"If true, will evaluate character error rate instead of word error rate.\n")

	fs.IntVar(&cfg.scliteCfg.LineWidth, "line-width", 1000,
		`When printing the text alignments for the output option "pralign", lines will be wrapped
when they reach this many characters
`)

	fs.StringVar(&cfg.scliteCfg.Encoding, "encoding", "utf-8",
		"What text encoding to use for interpreting text.\n")

	fs.BoolVar(&cfg.normCfg.CaseSensitive, "case-sensitive", false,
		"If true, scoring will be case sensitive.\n")

	fs.BoolVar(&cfg.normCfg.NormalizeUnicode, "normalize-unicode", false,
		"If true, unicode normalization wil be applied reference and hypothesis text before scoring.\n")

	shortUsage := `
sctk score \
  --ignore-first=true --delimiter="," --col-id=1 --col-trn=2 \
  --case-sensitive=false --normalize-unicode=true \
  --out=./wer --ref=truth.csv --hyp=output1.csv
`

	return &ffcli.Command{
		Name:       "score",
		FlagSet:    fs,
		ShortUsage: shortUsage,
		ShortHelp:  "Score hypothesis transcripts against provided reference transcripts.",
		Exec: func(_ context.Context, args []string) (err error) {
			if err := cfg.parseHypArgs(hypArgs); err != nil {
				fs.Usage()
				return err
			}

			if len(delimiter) != 1 || len(quoteChar) != 1 {
				return fmt.Errorf("demiliter and quote-char must be a single rune")
			}

			cfg.fileFormat.Delimiter = []rune(delimiter)[0]
			cfg.fileFormat.QuoteChar = []rune(quoteChar)[0]

			if err := cfg.checkArgs(); err != nil {
				fs.Usage()
				return err
			}

			return cfg.runScore(context.Background())
		},
	}
}

// Defining a stringArray type so that we can parse multiple instances of the
// -hyp flag.
type stringArray []string

func (i *stringArray) String() string {
	return strings.Join(*i, " ")
}

func (i *stringArray) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func (cfg *Config) parseHypArgs(hypArgs stringArray) error {
	var hypPath, hypName string
	i := 0

	for _, val := range hypArgs {

		switch parts := strings.Split(val, ","); len(parts) {
		case 0:
			return fmt.Errorf("hypothesis file not specified after -hyp flag")
		case 1:
			i++
			hypName, hypPath = fmt.Sprintf("hyp%d", i), parts[0]
		case 2:
			hypName, hypPath = parts[0], parts[1]
		default:
			return fmt.Errorf(
				"expected at most 2 comma delimited fields in -hyp flag value, got %d", len(parts),
			)
		}

		cfg.hypFiles = append(
			cfg.hypFiles,
			sctk.Hypothesis{SystemName: hypName, FilePath: hypPath},
		)
	}

	return nil
}

func (cfg *Config) checkArgs() error {
	if err := cfg.fileFormat.Validate(); err != nil {
		return err
	}

	if err := cfg.scliteCfg.Validate(); err != nil {
		return err
	}

	if _, err := os.Stat(cfg.refFile); os.IsNotExist(err) {
		return fmt.Errorf("specified reference file does not exist: %q", cfg.refFile)
	}

	for _, f := range cfg.hypFiles {
		if _, err := os.Stat(f.FilePath); os.IsNotExist(err) {
			return fmt.Errorf("specified hypothesis file does not exist: %q", f.FilePath)
		}
	}

	return nil
}

// runScore executes sclite and sc_stat on specified reference and hypothesis
// files to generate error analysis reports.
func (cfg *Config) runScore(ctx context.Context) error {
	return score.Score(
		ctx, cfg.fileFormat, cfg.normCfg, cfg.scliteCfg,
		cfg.outDir, cfg.refFile, cfg.hypFiles,
	)
}
