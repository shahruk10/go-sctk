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

// A Hypothesis points to a file containing transcripts from an ASR system,
// along with the name of the system. The name will be used in generated
// reports.
type Hypothesis struct {
	SystemName string
	FilePath   string
}
