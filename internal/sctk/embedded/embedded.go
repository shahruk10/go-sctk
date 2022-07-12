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

// Package embedded embed SCTK tools and provides executable paths to them for
// other packages.
package embedded

import (
	"embed"
	"fmt"
	"os"
	"path"
)

//go:embed bin
var sctk embed.FS

const (
	scliteBin  = "sclite"
	scStatsBin = "sc_stats"
)

const (
	executablePerm = 0777
)

// Sclite returns the path to the sclite executable. If the executable is not
// embedded or cannot be written to the user cache directory, this function will
// written an error.
func Sclite() (string, error) {
	return getBinPath(scliteBin)
}

// ScStats returns the path to the sc_stats executable. If the executable is not
// embedded or cannot be written to the user cache directory, this function will
// written an error.
func ScStats() (string, error) {
	return getBinPath(scStatsBin)
}

// getBinPath returns the executable path of the given binary. If the executable
// is not embedded or cannot be written to the user cache directory, this
// function will written an error.
func getBinPath(binName string) (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("failed to get binary cache directory: %w", err)
	}

	binDir := path.Join(cacheDir, "sctk")
	if err := os.MkdirAll(binDir, executablePerm); err != nil {
		return "", fmt.Errorf("failed to create sctk binary directory: %w", err)
	}

	binPath := path.Join(binDir, binName)
	if _, err := os.Stat(binPath); os.IsExist(err) {
		return binPath, nil
	}

	if err := writeBinToUserCache(binName, binPath); err != nil {
		return "", err
	}

	return binPath, nil
}

// writeBinToUserCache writes the SCTK binary with the given name from the
// embedded FS to the provided path and sets executable permissions.
func writeBinToUserCache(binName, binPath string) error {
	const (
		executablePerm = 0777
	)

	data, err := sctk.ReadFile(path.Join("bin", binName))
	if err != nil {
		return fmt.Errorf("%q not embedded with this tool", binName)
	}

	return os.WriteFile(binPath, data, executablePerm)
}
