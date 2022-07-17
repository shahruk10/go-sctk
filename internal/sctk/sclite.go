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

// Package sctk wraps SCTK tools and provides a simpler interface for generating
// reports and scoring ASR hypotheses submitted in a variety of formats against
// reference transcripts.
package sctk

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"

	"github.com/shahruk10/go-sctk/internal/sctk/embedded"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

// ScliteCfg configures report generation options for sclite.
type ScliteCfg struct {
	LineWidth int
	Encoding  string
	Reports   []string
	CER       bool
}

// Validate checks whether all configured options are valid and supported by
// sclite.
func (c *ScliteCfg) Validate() error {
	const (
		minLineWidth    = 100
		allowedEncoding = "ascii|utf-8"
		allowedReports  = "sum|rsum|pralign|all|sgml|stdout|lur|snt|spk|dtl|prf|wws|nl.sgml|none"
	)

	encodingCheck := regexp.MustCompile("^(" + allowedEncoding + ")$")
	reportCheck := regexp.MustCompile("^(" + allowedReports + ")$")

	if c.LineWidth <= 0 {
		return fmt.Errorf("line width must be >= %d", minLineWidth)
	}

	if !encodingCheck.MatchString(c.Encoding) {
		return fmt.Errorf(
			"unsupported encoding option %q, supported %s", c.Encoding, allowedEncoding,
		)
	}

	for _, r := range c.Reports {
		if !reportCheck.MatchString(r) {
			return fmt.Errorf(
				"unsupported report option %q, supported %s", r, allowedReports,
			)
		}
	}

	return nil
}

// RunSclite executes the sclite tool on the given reference and hypothesis
// files, and evaluates word error rates. It also generates alignments between
// the reference and hypotheses, and optionally character error rate as well.
func RunSclite(
	ctx context.Context, cfg ScliteCfg, outDir, refFile string, hypFiles []Hypothesis,
) error {
	const (
		filePerm = 0777
	)

	if len(hypFiles) == 0 {
		return fmt.Errorf("no hypothesis files provided")
	}

	args := []string{
		"-i", "swb", // UttID format utt ID (swb = switchboard).
		"-r", refFile, "trn", // Reference file and format.
		"-O", outDir,
		"-l", fmt.Sprintf("%d", cfg.LineWidth),
		"-e", cfg.Encoding,
	}

	// Use default reports if unspecified.
	if len(cfg.Reports) == 0 {
		cfg.Reports = []string{"sum", "rsum", "dtl", "sgml"}
	}

	args = append(args, "-o")
	args = append(args, cfg.Reports...)

	// Normalization step before running sclite adjusts for sensitivity; so sclite
	// is set to be always case sensitive.
	args = append(args, "-s")

	if cfg.CER {
		args = append(args, "-c")
	}

	for _, hyp := range hypFiles {
		args = append(args, "-h", hyp.FilePath, "trn", hyp.SystemName)
	}

	scliteBin, err := embedded.Sclite()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(outDir, filePerm); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	cmd := exec.CommandContext(ctx, scliteBin, args...)

	if stderr, err := cmd.CombinedOutput(); err != nil {
		logrus.WithFields(log.Fields{
			"stderr": string(stderr),
		}).Error("sclite encountered errors")
	}

	return err
}
