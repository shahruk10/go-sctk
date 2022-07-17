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

// Package score wraps SCTK tools and provides a simpler interface for scoring
// ASR hypotheses submitted in a variety of formats against reference transcripts.
package score

import (
	"context"
	"fmt"

	"github.com/shahruk10/go-sctk/internal/sctk"
)

func Score(
	ctx context.Context, fileFormat FileFormat, normCfg NormalizeConfig, scliteCfg sctk.ScliteCfg,
	outDir, refFile string, hypFiles []sctk.Hypothesis,
) error {
	normRef, normHypFiles, err := normalizeFiles(
		ctx, fileFormat, normCfg, outDir, refFile, hypFiles,
	)
	if err != nil {
		return err
	}

	if err := sctk.RunSclite(ctx, scliteCfg, outDir, normRef, normHypFiles); err != nil {
		return fmt.Errorf("failed to run sclite: %w", err)
	}

	return nil
}
