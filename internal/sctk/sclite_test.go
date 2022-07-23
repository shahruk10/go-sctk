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

package sctk

import (
	"context"
	"os"
	"path"
	"regexp"
	"testing"

	"github.com/google/go-cmp/cmp"
)

//nolint: funlen // table tests can be long.
func TestRunSclite(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		ref     string
		hyp     []Hypothesis
		cfg     ScliteCfg
		wantErr bool
	}{
		{
			name: "good1_wer",
			ref:  "testdata/sclite/good1_ref.trn",
			hyp: []Hypothesis{
				{SystemName: "good1_hyp1", FilePath: "testdata/sclite/good1_hyp1.trn"},
				{SystemName: "good1_hyp2", FilePath: "testdata/sclite/good1_hyp2.trn"},
			},
			cfg: ScliteCfg{
				LineWidth: 120,
				Encoding:  "utf-8",
				CER:       false,
			},
			wantErr: false,
		},
		{
			name: "good1_cer",
			ref:  "testdata/sclite/good1_ref.trn",
			hyp: []Hypothesis{
				{SystemName: "good1_hyp1", FilePath: "testdata/sclite/good1_hyp1.trn"},
				{SystemName: "good1_hyp2", FilePath: "testdata/sclite/good1_hyp2.trn"},
			},
			cfg: ScliteCfg{
				LineWidth: 120,
				Encoding:  "utf-8",
				CER:       true,
			},
			wantErr: false,
		},
		{
			name: "bad_config1",
			cfg: ScliteCfg{
				LineWidth: 120,
				Encoding:  "utf-16",
				CER:       false,
			},
			wantErr: true,
		},
		{
			name: "bad_config2",
			cfg: ScliteCfg{
				LineWidth: -1,
				Encoding:  "utf-8",
				CER:       false,
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(subT *testing.T) {
			subT.Parallel()

			gotErr := tc.cfg.Validate()

			checkError(subT, tc.wantErr, gotErr)
			if gotErr != nil && tc.wantErr {
				return
			}

			ctx := context.Background()
			outDir := subT.TempDir()

			gotErr = RunSclite(ctx, tc.cfg, outDir, tc.ref, tc.hyp)

			checkError(subT, tc.wantErr, gotErr)
			if gotErr != nil && tc.wantErr {
				return
			}

			for _, hyp := range tc.hyp {
				filesToCompare := []string{
					hyp.SystemName + ".trn.sys",
					hyp.SystemName + ".trn.sgml",
					hyp.SystemName + ".trn.dtl",
				}

				for _, name := range filesToCompare {
					wantPath := path.Join("testdata", "sclite", "wer", name)
					if tc.cfg.CER {
						wantPath = path.Join("testdata", "sclite", "cer", name)
					}

					gotPath := path.Join(outDir, name)

					compareFiles(subT, wantPath, gotPath)
				}
			}
		})
	}
}

// compareFiles compares the contents of the files at the given paths by
// generating the diff between them. Some normalization steps are applied before
// doing so, such as converting timestamps to a fixed string.
func compareFiles(t *testing.T, wantPath, gotPath string) {
	t.Helper()

	want, err := os.ReadFile(wantPath)
	if err != nil {
		t.Errorf("unexpected error while reading reference file, want=nil, got=%v", err)
		return
	}

	got, err := os.ReadFile(gotPath)
	if err != nil {
		t.Errorf("unexpected error while reading output file, want=nil, got=%v", err)
		return
	}

	want = removeCreationDate(t, want)
	got = removeCreationDate(t, got)

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("differences between reference and output files (-want, +got):\n%s", diff)
		return
	}
}

// removeCreationDate removes the date embedded in files generated by SCTK for
// testing purposes, since pre-generated test reference files will always have a
// different timestamp.
func removeCreationDate(t *testing.T, data []byte) []byte {
	const (
		fixedDateStr = `creation_date="Tue Jul 12 12:12:12 2022"`
	)

	t.Helper()

	r := regexp.MustCompile("creation_date=\"[^\"]+\"")

	return r.ReplaceAll(data, []byte(fixedDateStr))
}

func checkError(t *testing.T, wantErr bool, gotErr error) {
	if gotErr != nil && !wantErr {
		t.Errorf("got unexpected error, want=nil, got=%v", gotErr)
		return
	}

	if gotErr == nil && wantErr {
		t.Errorf("did not get expected error, want=non-nil, got=%v", gotErr)
		return
	}
}
