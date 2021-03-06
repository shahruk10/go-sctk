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
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/shahruk10/go-sctk/internal/sctk/embedded"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

const (
	filePerm = 0777
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
		// We leave out the pra file here, because for non-english alphabets, the
		// spacing between words in the pra file added to align reference and
		// hypotheses doesn't always work properly (due to diacritics and font
		// ligatures). We instead parse the sgml file and generate our own pra file,
		// with aligned ref and hyp shown in markdown tables.
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

	return genAlignmentFileFromSgml(outDir)
}

func genAlignmentFileFromSgml(outDir string) error {
	sgmlFiles, err := filepath.Glob(path.Join(outDir, "*.sgml"))
	if err != nil {
		logrus.WithFields(log.Fields{
			"err": err,
		}).Error("no sgml files were produced, cannot generate alignment file")
	}

	// Fixed for now.
	formats := []TableFormat{TableFormatMarkdown, TableFormatHTML, TableFormatCSV}

	for _, sgmlFile := range sgmlFiles {
		aligned, err := ReadAlignmentSgml(sgmlFile)
		if err != nil {
			return err
		}

		for _, format := range formats {
			var ext string
			switch {
			case format == TableFormatMarkdown:
				ext = ".pra.md"
			case format == TableFormatHTML:
				ext = ".pra.html"
			case format == TableFormatCSV:
				ext = ".pra.csv"
			default:
				ext = ".pra"
			}

			outFile := strings.ReplaceAll(sgmlFile, ".sgml", ext)
			if err := WriteAlignment(outFile, aligned, format); err != nil {
				return err
			}
		}

		// Also dump the alignments as JSON to easily parse back later for further
		// processing if required.
		jsonData, err := json.MarshalIndent(aligned, "", " ")
		if err != nil {
			return err
		}

		ext := ".pra.json"
		outFile := strings.ReplaceAll(sgmlFile, ".sgml", ext)

		if err := os.WriteFile(outFile, jsonData, filePerm); err != nil {
			return err
		}
	}

	return nil
}
