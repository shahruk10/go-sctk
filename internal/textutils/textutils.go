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

import "strings"

// fieldsWithQuoted splits the given string into fields, based on the given
// delimiter and quote character. If the delimiter occurs between quoteChars,
// that part of the string won't be split.
func FieldsWithQuoted(s string, delimiter, quoteChar rune) []string {
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
