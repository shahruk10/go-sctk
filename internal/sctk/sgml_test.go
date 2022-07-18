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
	"os"
	"path"
	"testing"

	"github.com/google/go-cmp/cmp"
)

//nolint: funlen // table tests can be long.
func TestReadAlignmentSgml(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		sgmlPath string
		wantErr  bool
		want     *AlignedHypothesis
	}{
		{
			name:     "non existing file",
			sgmlPath: "testdata/sgml/does-not-exist.bangla.trn.sgml",
			wantErr:  true,
		},
		{
			name:     "missing speaker id",
			sgmlPath: "testdata/sgml/bad1.bangla.trn.sgml",
			wantErr:  true,
		},
		{
			name:     "missing sentence id",
			sgmlPath: "testdata/sgml/bad2.bangla.trn.sgml",
			wantErr:  true,
		},
		{
			name:     "non integer word count",
			sgmlPath: "testdata/sgml/bad3.bangla.trn.sgml",
			wantErr:  true,
		},
		{
			name:     "invalid word count",
			sgmlPath: "testdata/sgml/bad4.bangla.trn.sgml",
			wantErr:  true,
		},
		{
			name:     "invalid sequence count",
			sgmlPath: "testdata/sgml/bad5.bangla.trn.sgml",
			wantErr:  true,
		},
		{
			name:     "premature EOF",
			sgmlPath: "testdata/sgml/bad6.bangla.trn.sgml",
			wantErr:  true,
		},
		{
			name:     "fewer than expected parts in aligned word string",
			sgmlPath: "testdata/sgml/bad7.bangla.trn.sgml",
			wantErr:  true,
		},
		{
			sgmlPath: "testdata/sgml/good.bangla.trn.sgml",
			wantErr:  false,
			name:     "good",
			want: &AlignedHypothesis{
				SystemName: "bangla",
				Speakers: map[string]SpeakerSentences{
					"common": {
						"common_voice_bn_30620258.mp3": &AlignedSentence{
							SystemName: "bangla",
							SpeakerID:  "common",
							SentenceID: "common_voice_bn_30620258.mp3",
							Sequence:   0,
							WordCount:  5,
							Words: []AlignedWord{
								{"C", "তার", "তার"},
								{"C", "পিতার", "পিতার"},
								{"C", "নাম", "নাম"},
								{"C", "কালীপ্রসন্ন", "কালীপ্রসন্ন"},
								{"S", "ভট্টাচার্য।", "ভট্টাচার্য"},
							},
						},
						"common_voice_bn_30620259.mp3": &AlignedSentence{
							SystemName: "bangla",
							SpeakerID:  "common",
							SentenceID: "common_voice_bn_30620259.mp3",
							Sequence:   1,
							WordCount:  8,
							Words: []AlignedWord{
								{"C", "ভৌগোলিক", "ভৌগোলিক"},
								{"C", "অবস্থান", "অবস্থান"},
								{"C", "অনুযায়ী", "অনুযায়ী"},
								{"D", "শহরটির", ""},
								{"D", "পূর্ব", ""},
								{"D", "দিকে", ""},
								{"S", "কাশ্মীর", "চাহরটিরপূর্বদিকেকাশ্মির"},
								{"S", "অবস্থিত।", "অবস্থিত"},
							},
						},
						"common_voice_bn_30620260.mp3": &AlignedSentence{
							SystemName: "bangla",
							SpeakerID:  "common",
							SentenceID: "common_voice_bn_30620260.mp3",
							Sequence:   2,
							WordCount:  5,
							Words: []AlignedWord{
								{"C", "এটি", "এটি"},
								{"S", "বিশ্বব্যাপি", "বিশ্বব্যাপী"},
								{"I", "", "একই"},
								{"C", "হয়ে", "হয়ে"},
								{"S", "থাকে।", "থাকে"},
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(subT *testing.T) {
			subT.Parallel()

			got, err := ReadAlignmentSgml(tc.sgmlPath)

			if err != nil && !tc.wantErr {
				subT.Errorf("got unexpected error, want=nil, got=%v", err)
				return
			}

			if err == nil && tc.wantErr {
				subT.Errorf("did not get expected error, want=non-nil, got=%v", err)
				return
			}

			if err != nil && tc.wantErr {
				return
			}

			compareAlignedHypotheses(subT, tc.want, got)
		})
	}
}

func TestWriteAlignmentSgml(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		sgmlPath string
		wantPath string
		format   TableFormat
		wantErr  bool
	}{
		{
			name:     "html",
			sgmlPath: "testdata/sgml/good.bangla.trn.sgml",
			wantPath: "testdata/sgml/good.bangla.trn.pra.html",
			format:   TableFormatHTML,
			wantErr:  false,
		},
		{
			name:     "markdown",
			sgmlPath: "testdata/sgml/good.bangla.trn.sgml",
			wantPath: "testdata/sgml/good.bangla.trn.pra.md",
			format:   TableFormatMarkdown,
			wantErr:  false,
		},
		{
			name:     "csv",
			sgmlPath: "testdata/sgml/good.bangla.trn.sgml",
			wantPath: "testdata/sgml/good.bangla.trn.pra.csv",
			format:   TableFormatCSV,
			wantErr:  false,
		},
		{
			name:     "text",
			sgmlPath: "testdata/sgml/good.bangla.trn.sgml",
			wantPath: "testdata/sgml/good.bangla.trn.pra.txt",
			format:   TableFormatTxt,
			wantErr:  false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(subT *testing.T) {
			subT.Parallel()

			aligned, err := ReadAlignmentSgml(tc.sgmlPath)
			if err != nil {
				subT.Fatalf("unexpected error while reading sgml file, want=nil, got=%v", err)
			}

			outDir := subT.TempDir()
			outFile := path.Join(outDir, "alignment.tmp")

			gotErr := WriteAlignment(outFile, aligned, tc.format)

			if gotErr != nil && !tc.wantErr {
				subT.Errorf("got unexpected error, want=nil, got=%v", gotErr)
				return
			}

			if gotErr == nil && tc.wantErr {
				subT.Errorf("did not get expected error, want=non-nil, got=%v", gotErr)
				return
			}

			if gotErr != nil && tc.wantErr {
				return
			}

			compareFiles(subT, tc.wantPath, outFile)
		})
	}
}

func compareAlignedHypotheses(t *testing.T, want, got *AlignedHypothesis) {
	t.Helper()

	if want.SystemName != got.SystemName {
		t.Errorf("unexpected system name, want=%s, got=%s", want.SystemName, got.SystemName)
	}

	if len(want.Speakers) != len(got.Speakers) {
		t.Errorf("unexpected number of speakers, want=%d, got=%d", len(want.Speakers), len(got.Speakers))
	}

	for spk, wantSentences := range want.Speakers {
		gotSentences, ok := got.Speakers[spk]
		if !ok {
			t.Errorf("missing speaker in output: %q", spk)
			continue
		}

		if len(wantSentences) != len(gotSentences) {
			t.Errorf(
				"unexpected number of sentences for speaker %q, want=%d, got=%d",
				spk, len(wantSentences), len(gotSentences),
			)
		}

		for sentID, wantSent := range wantSentences {
			gotSent, ok := gotSentences[sentID]
			if !ok {
				t.Errorf("missing sentence for speaker %q, want=%q", spk, sentID)
				continue
			}

			if diff := cmp.Diff(wantSent, gotSent); diff != "" {
				t.Errorf("unexpected aligned sentence for speaker=%q, sent=%q (-want, +got):\n%s", spk, sentID, diff)
			}
		}
	}
}

func compareFiles(t *testing.T, wantPath, gotPath string) {
	t.Helper()

	want, err := os.ReadFile(wantPath)
	if err != nil {
		t.Errorf("unexpected error while reading reference file, want=nil, got=%v", err)
	}

	got, err := os.ReadFile(gotPath)
	if err != nil {
		t.Errorf("unexpected error while reading output file, want=nil, got=%v", err)
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("differences between reference and output files (-want, +got):\n%s", diff)
	}
}
