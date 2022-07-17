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

package score

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/dlclark/regexp2"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"golang.org/x/text/unicode/norm"

	"github.com/shahruk10/go-sctk/internal/sctk"
)

// An Utt contains the transcript of an utterance along with its ID.
type Utt struct {
	ID         string
	Transcript string
}

// NormalizeConfig specifies how to normalize utterance transcripts.
type NormalizeConfig struct {
	CaseSensitive    bool
	NormalizeUnicode bool
}

// FileFormat specifies the expected format of reference and hypotheses files.
type FileFormat struct {
	Delimiter      rune
	QuoteChar      rune
	ColTrn         int
	ColID          int
	IgnoreFirstRow bool
}

// Validate checks whether the options configured for the file format are
// consistent, and supported.
func (f *FileFormat) Validate() error {
	if f.ColID < 0 || f.ColTrn < 0 {
		return fmt.Errorf("column index for transcript and ID must be >=0")
	}

	if f.ColID == f.ColTrn {
		return fmt.Errorf("column index for transcript and ID must not be the same")
	}

	return nil
}

// normalizeFiles parses the reference and hypotheses files, and normalizes them
// based on the provided configs. The normalized files are written to the
// provided output directory; the normalized reference file is named ref.trn,
// while hypotheses files are named based on their system name.
func normalizeFiles(
	ctx context.Context, fileFormat FileFormat, cfg NormalizeConfig,
	outDir, refFile string, hypFiles []sctk.Hypothesis,
) (string, []sctk.Hypothesis, error) {
	// Read reference transcripts.
	refUtts, err := readTranscriptFile(refFile, fileFormat)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read reference file: %w", err)
	}

	// Write normalized reference transcripts into format expected by SCTK.
	normalizeUtts(refUtts, cfg)

	refNorm := path.Join(outDir, "ref.trn")
	if err := writeTranscriptFile(refUtts, refNorm); err != nil {
		return "", nil, fmt.Errorf("failed to write normalized reference file: %w", err)
	}

	// Getting the set of reference utt IDs. Will filter utts from hypotheses that
	// do not have a reference utt.
	refIDs := make(map[string]struct{})
	for _, utt := range refUtts {
		refIDs[utt.ID] = struct{}{}
	}

	if len(refIDs) == 0 {
		return "", nil, fmt.Errorf("reference file does not contain any utterances")
	}

	// Read hypothesis transcripts and write out normalized version in the format
	// expected by SCTK.
	normHypFiles := make([]sctk.Hypothesis, 0, len(hypFiles))

	for _, hyp := range hypFiles {
		hypUtts, err := readTranscriptFile(hyp.FilePath, fileFormat)
		if err != nil {
			return "", nil, fmt.Errorf("failed to read hypothesis file: %w", err)
		}

		normalizeUtts(hypUtts, cfg)

		hypUtts = filterUtts(hypUtts, refIDs)
		if len(hypUtts) == 0 {
			return "", nil, fmt.Errorf(
				"no utterance IDs in common between reference file and %q", hyp.FilePath,
			)
		}

		sanitizedName := sanitizeSystemName(hyp.SystemName)
		hypNorm := path.Join(outDir, sanitizedName+".trn")

		if err := writeTranscriptFile(hypUtts, hypNorm); err != nil {
			return "", nil, fmt.Errorf("failed to write normalized hypothesis file: %w", err)
		}

		normHypFiles = append(normHypFiles, sctk.Hypothesis{
			SystemName: sanitizedName,
			FilePath:   hypNorm,
		})
	}

	return refNorm, normHypFiles, nil
}

// normalizeUtts applies different normalization processes in-place on the
// provided list of utts.
func normalizeUtts(utts []Utt, cfg NormalizeConfig) {
	for i := range utts {
		trn := utts[i].Transcript

		if !cfg.CaseSensitive {
			trn = strings.ToLower(trn)
		}

		if cfg.NormalizeUnicode {
			trn = norm.NFC.String(trn)
			trn = removeZW(trn)
		}

		utts[i].Transcript = trn
	}
}

// removeZW removes all optional occurrences of ZWNJ or ZWJ from Bangla text.
func removeZW(s string) string {
	const (
		zw = "\u200D" // Zero Width Joiner
	)

	// The non-printing characters U+200C (ZWNJ) and U+200D (ZWJ) are used in
	// Bangla to optionally control the appearance of ligatures, except in one
	// special situation: after RA (র) and before HOSONTO + YA (য্‌), the presence
	// or absence of ZWJ (formerly ZWNJ) changes the visual appearance of the
	// involved consonants in a meaningful way.
	//
	// This occurrences of ZWJ must be preserved, while all other occurrences are
	// advisory and can be removed for most purposes. After RA and before HOSONTO
	// + YA, this function changes ZWNJ to ZWJ and preserves ZWJ; and removes ZWNJ
	// and ZWJ everywhere else.
	//
	// Using regexp2 package, since Go's regexp currently doesn't support positive
	// lookbehind.
	zwStandardize := regexp2.MustCompile(
		"(?<=\u09b0)[\u200c\u200d]+(?=\u09cd\u09af)", regexp2.Unicode,
	)

	zwDelete := regexp2.MustCompile(
		"(?<!\u09b0)[\u200c\u200d](?!\u09cd\u09af)", regexp2.Unicode,
	)

	s, err := zwStandardize.Replace(s, zw, -1, -1)
	if err != nil {
		logrus.WithFields(log.Fields{
			"text":  s,
			"error": err,
		}).Error("failed to standardize zero-width joiners")

		return s
	}

	s, err = zwDelete.Replace(s, "", -1, -1)
	if err != nil {
		logrus.WithFields(log.Fields{
			"text":  s,
			"error": err,
		}).Error("failed to remove zero-width joiners")
	}

	return s
}

// filterUtts removes utterances in the given list whose IDs do not appear in
// the provided set of reference IDs. The filtered list of utts is returned.
func filterUtts(utts []Utt, refIDs map[string]struct{}) []Utt {
	n := 0
	for _, utt := range utts {
		if _, ok := refIDs[utt.ID]; ok {
			utts[n] = utt
			n++
		}
	}

	return utts[:n]
}

// readTranscriptFile reads the utterance data from the given transcript file
// based on the provided file format.
func readTranscriptFile(
	filePath string, fileFormat FileFormat,
) ([]Utt, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read reference file: %w", err)
	}

	defer closeFileOrLog(f)

	scanner := bufio.NewScanner(f)
	ldx := 0

	if fileFormat.IgnoreFirstRow {
		ldx++
		scanner.Scan()
	}

	utts := make([]Utt, 0)

	maxColsExpected := fileFormat.ColTrn
	if maxColsExpected < fileFormat.ColID {
		maxColsExpected = fileFormat.ColID
	}

	for scanner.Scan() {
		ldx++

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := fieldsWithQuoted(line, fileFormat.Delimiter, fileFormat.QuoteChar)

		if len(parts) < maxColsExpected {
			return nil, fmt.Errorf(
				"expected transcript file to contain at least %d columns, got %d on line %d",
				maxColsExpected, len(parts), ldx,
			)
		}

		ID, trn := sanitizeUttID(parts[fileFormat.ColID]), parts[fileFormat.ColTrn]
		utts = append(utts, Utt{ID, trn})
	}

	return utts, nil
}

// writeTranscriptFile writes the provided list of utterances to the given path
// in the format expected by SCTK tools - "<transcript>(<uttID>)".
func writeTranscriptFile(
	utts []Utt, filePath string,
) error {
	f, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create output transcript file: %w", err)
	}

	defer closeFileOrLog(f)

	w := bufio.NewWriter(f)

	// Sorting utts by ID.
	sort.SliceStable(utts, func(i, j int) bool {
		switch c := strings.Compare(utts[i].ID, utts[j].ID); c {
		case -1:
			return true
		default:
			return false
		}
	})

	for _, utt := range utts {
		line := fmt.Sprintf("%s (%s)\n", utt.Transcript, utt.ID)
		if _, err := w.WriteString(line); err != nil {
			return fmt.Errorf("failed to write line to transcript file: %w", err)
		}
	}

	if err := w.Flush(); err != nil {
		return fmt.Errorf("failed to flush write buffer to transcript file: %w", err)
	}

	return nil
}

// closeFileOrLog tries to close the given file. If it fails to do so, the error
// is logged.
func closeFileOrLog(f *os.File) {
	if errClose := f.Close(); errClose != nil {
		log.WithFields(log.Fields{
			"error": errClose,
			"path":  f.Name(),
		}).Error("failed to close file")
	}
}

// sanitizeSystemName converts the given string representing a system name to all
// lower case, and replaces and spaces with underscores.
func sanitizeSystemName(name string) string {
	parts := strings.Fields(strings.ToLower(name))
	return strings.Join(parts, "_")
}

// sanitizeUttID converts the given string representing an utterance ID by
// replacing spaces with underscores
func sanitizeUttID(ID string) string {
	parts := strings.Fields(ID)
	return strings.Join(parts, "_")
}

// fieldsWithQuoted splits the given string into fields, based on the given
// delimiter and quote character. If the delimiter occurs between quoteChars,
// that part of the string won't be split.
func fieldsWithQuoted(s string, delimiter, quoteChar rune) []string {
	hasQuotes := false
	inQuoted := false

	parts := strings.FieldsFunc(s, func(r rune) bool {
		if r == quoteChar {
			inQuoted = !inQuoted
			hasQuotes = true
		}

		return !inQuoted && r == delimiter
	})

	if hasQuotes {
		for i := range parts {
			parts[i] = strings.Trim(parts[i], string(quoteChar))
		}
	}

	return parts
}
