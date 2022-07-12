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
	"os/exec"

	"github.com/shahruk10/go-sctk/internal/sctk/embedded"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

// RunSclite executes the sclite tool on the given reference and hypothesis
// files, and evaluates word error rates. It also generates alignments between
// the reference and hypotheses, and optionally character error rate as well.
func RunSclite(
	ctx context.Context, outDir, refFile string, hypFiles []Hypothesis, evalCER bool,
) error {
	if len(hypFiles) == 0 {
		return fmt.Errorf("no hypothesis files provided")
	}

	args := []string{
		"-r", refFile, "trn", // Reference file and format.
		"-i", "swb", // UttID format utt ID (swb = switchboard).
		"-O", outDir, // Specify output dir.
		"-o", "sum", "rsum", "pralign", "dtl", "sgml", // Reports to generate.
		"-l", "120", // Set line width.
		"-s", // Case sensitive alignments (case is pre-normalized when needed)
		"-e", "utf-8",
	}

	for _, hyp := range hypFiles {
		args = append(args, "-h", hyp.FilePath, "trn", hyp.SystemName)
	}

	if evalCER {
		args = append(args, "-c")
	}

	scliteBin, err := embedded.Sclite()
	if err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, scliteBin, args...)

	if stderr, err := cmd.CombinedOutput(); err != nil {
		logrus.WithFields(log.Fields{
			"stderr": string(stderr),
		}).Error("sclite encountered errors")
	}

	return err
}
