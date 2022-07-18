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

// Package fileutils provides convenience functions for opening, closing and
// writing to files.
package fileutils

import (
	"os"

	"github.com/sirupsen/logrus"
)

// CloseFileOrLog tries to close the given file. If it fails to do so, the error
// is logged.
func CloseFileOrLog(f *os.File) {
	if errClose := f.Close(); errClose != nil {
		logrus.WithFields(logrus.Fields{
			"error": errClose,
			"path":  f.Name(),
		}).Error("failed to close file")
	}
}
