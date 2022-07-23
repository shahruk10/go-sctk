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

// Package textutils provides convenience functions for parsing and manipulating
// text.
package textutils

import (
	"encoding/csv"
	"strings"

	"github.com/sirupsen/logrus"
)

// FieldsWithQuoted splits the given string into fields, based on the given
// delimiter. This method handles the case where a quote may appear in an
// unquoted field and a non-doubled quote may appear in a quoted field.
func FieldsWithQuoted(s string, delimiter rune) []string {
	r := csv.NewReader(strings.NewReader(s))
	r.Comma = delimiter
	r.LazyQuotes = true

	parts, err := r.Read()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"string":    s,
			"delimiter": string(delimiter),
			"error":     err,
		}).Error("failed to split fields")
	}

	return parts
}
