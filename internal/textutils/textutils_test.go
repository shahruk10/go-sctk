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

package textutils

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestFieldsWithQuoted(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		input     string
		delimiter rune
		want      []string
	}{
		{
			name:      "bengali characters",
			input:     `C,"আ","আ":C,"র","র":S,"ো","ও":C,"ক","ক":C,"ম","ম":C,"প","প":C,"্","্":C,"র","র":C,"চ","চ":C,"ল","ল":C,"ি","ি":C,"ত","ত":C,"এ","এ":C,"ক","ক":C,"ট","ট":C,"ি","ি":C,"র","র":C,"ূ","ূ":C,"প","প":C,"ভ","ভ":C,"ে","ে":C,"দ","দ":C,"হ","হ":C,"ল","ল":D,""",:D,""",:D,"ন",:D,"ম",:C,"ো","ো":S,"ব","ন":C,"া","া":C,"ম","ম":I,,"ও":S,"্","ভ":S,""","া":S,""","ব":C,"।","।"`,
			delimiter: ':',
			want:      []string{`C,"আ","আ"`, `C,"র","র"`, `S,"ো","ও"`, `C,"ক","ক"`, `C,"ম","ম"`, `C,"প","প"`, `C,"্","্"`, `C,"র","র"`, `C,"চ","চ"`, `C,"ল","ল"`, `C,"ি","ি"`, `C,"ত","ত"`, `C,"এ","এ"`, `C,"ক","ক"`, `C,"ট","ট"`, `C,"ি","ি"`, `C,"র","র"`, `C,"ূ","ূ"`, `C,"প","প"`, `C,"ভ","ভ"`, `C,"ে","ে"`, `C,"দ","দ"`, `C,"হ","হ"`, `C,"ল","ল"`, `D,""",`, `D,""",`, `D,"ন",`, `D,"ম",`, `C,"ো","ো"`, `S,"ব","ন"`, `C,"া","া"`, `C,"ম","ম"`, `I,,"ও"`, `S,"্","ভ"`, `S,""","া"`, `S,""","ব"`, `C,"।","।"`},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(subT *testing.T) {
			got := FieldsWithQuoted(tc.input, tc.delimiter)

			if diff := cmp.Diff(tc.want, got); diff != "" {
				subT.Errorf("unexpected output, (-want, +got):\n%s", diff)
			}
		})
	}
}
