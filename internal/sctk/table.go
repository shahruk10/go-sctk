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
	"fmt"
	"sort"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

// TableFormat determines in which format the alignments are written to a file.
type TableFormat int

const (
	TableFormatMarkdown TableFormat = iota
	TableFormatHTML
	TableFormatCSV
	TableFormatTxt
)

// ToTable generates a report in the specified format, with the alignments for
// each speaker sentence presented in a table, along with other auxillary
// information.
func (a *AlignedHypothesis) ToTable(f TableFormat) string {
	// Sorting speakers alphabetically.
	speakers := make([]string, 0, len(a.Speakers))
	sentCount := 0
	for spk := range a.Speakers {
		speakers = append(speakers, spk)
		sentCount += len(a.Speakers[spk])
	}

	sort.Strings(speakers)

	w := strings.Builder{}

	w.WriteString(getSectionHeader("SYSTEM ALIGNMENT", f, 2))
	w.WriteString(getBodyText(fmt.Sprintf("System Name = %s", a.SystemName), f))
	w.WriteString(getBodyText(fmt.Sprintf("Speakers = %d", len(a.Speakers)), f))
	w.WriteString(getBodyText(fmt.Sprintf("Sentences = %d", sentCount), f))
	w.WriteString(getSectionFooter(f))

	for _, spk := range speakers {
		sents := a.Speakers[spk]

		w.WriteString(getSectionHeader(spk, f, 3))
		w.WriteString(sents.ToTable(f))
		w.WriteString(getSectionFooter(f))
	}

	return w.String()
}

// ToTable generates a table for each sentence showing the reference and
// hypothesis, along with the alignment label for each token. It also prints
// some stats about the number of errors between the reference and hypothesis.
func (s SpeakerSentences) ToTable(f TableFormat) string {
	// Sorting speaker sentences by sequence order.
	sortedSents := make([]string, 0, len(s))
	for sent := range s {
		sortedSents = append(sortedSents, sent)
	}

	sort.SliceStable(sortedSents, func(i, j int) bool {
		si, sj := sortedSents[i], sortedSents[j]
		return s[si].Sequence < s[sj].Sequence
	})

	w := strings.Builder{}

	for _, sent := range sortedSents {
		w.WriteString(getSectionHeader(sent, f, 4))
		w.WriteString(getBodyText(s[sent].Stats(), f))
		w.WriteString(s[sent].ToTable(f))
		w.WriteString(getSectionFooter(f))
	}

	return w.String()
}

// Stats returns a string containing the proportion of words that the ASR system
// got right, substituted, deleted and inserted.
func (s *AlignedSentence) Stats() string {
	var cor, sub, del, ins float32
	for _, w := range s.Words {
		switch {
		case w.Label == "S":
			sub++
		case w.Label == "D":
			del++
		case w.Label == "I":
			ins++
		default:
			cor++
		}
	}

	total := float32(s.WordCount) + 1e-10

	return fmt.Sprintf(
		"Cor=%3.1f%%\tSub=%3.1f%%\tDel=%3.1f%%\tIns=%3.1f%%\n",
		100*cor/total, 100*sub/total, 100*del/total, 100*ins/total,
	)
}

// ToTable generates a table with three rows containing the reference and
// hypothesis sentence, along with the alignment label for each token.
func (s *AlignedSentence) ToTable(f TableFormat) string {
	t := table.NewWriter()

	colCfg := make([]table.ColumnConfig, 0, s.WordCount+1)

	// Column numbers are indexed from 1.
	colCfg = append(colCfg, table.ColumnConfig{
		Number: 1,
		Align:  text.AlignLeft,
		VAlign: text.VAlignMiddle,
	})

	for i := 0; i < s.WordCount; i++ {
		// Column numbers are indexed from 1. The configs added here are for the
		// columns after the first column, so we are using i+2 here.
		colCfg = append(colCfg, table.ColumnConfig{
			Number: i + 2,
			Align:  text.AlignCenter,
			VAlign: text.VAlignMiddle,
		})
	}

	rows := []table.Row{
		make(table.Row, s.WordCount+1), // REF
		make(table.Row, s.WordCount+1), // HYP
		make(table.Row, s.WordCount+1), // EVAL
	}

	rows[0][0] = "REF"
	rows[1][0] = strings.ToUpper(s.SystemName)
	rows[2][0] = "EVAL"

	for i, w := range s.Words {
		rows[0][i+1] = w.Ref
		rows[1][i+1] = w.Hyp

		// Don't print eval label if correct (C).
		if w.Label == "C" {
			rows[2][i+1] = ""
		} else {
			rows[2][i+1] = w.Label
		}
	}

	t.AppendHeader(nil)
	t.AppendRows(rows)
	t.SetColumnConfigs(colCfg)

	return renderTable(t, f)
}

func getSectionHeader(title string, f TableFormat, level int) string {
	switch f {
	case TableFormatHTML:
		headerTag := fmt.Sprintf("h%d", level)
		return fmt.Sprintf("\n<%s>%s</%s>\n", headerTag, title, headerTag)
	case TableFormatMarkdown:
		headerTag := strings.Repeat("#", level)
		return fmt.Sprintf("\n%s %s\n", headerTag, title)
	default:
		return fmt.Sprintf("\n%s\n", title)
	}
}

func getSectionFooter(f TableFormat) string {
	switch f {
	case TableFormatHTML:
		return "\n<br>\n"
	case TableFormatMarkdown:
		return "\n---\n"
	default:
		return "\n\n"
	}
}

func getBodyText(s string, f TableFormat) string {
	switch f {
	case TableFormatHTML:
		return fmt.Sprintf("\n<p>%s</p>\n", s)
	case TableFormatMarkdown:
		return fmt.Sprintf("\n- %s\n", s)
	default:
		return fmt.Sprintf("\n%s\n", s)
	}
}

func renderTable(t table.Writer, f TableFormat) string {
	switch f {
	case TableFormatHTML:
		return t.RenderHTML()
	case TableFormatMarkdown:
		return t.RenderMarkdown()
	case TableFormatCSV:
		return t.RenderCSV()
	default:
		return t.Render()
	}
}
