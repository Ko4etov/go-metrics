package buildinfo

import (
	"fmt"
)

type BuildInfo struct {
	Version string
	Date    string
	Commit  string
}

func New(version, date, commit string) *BuildInfo {
	return &BuildInfo{
		Version: version,
		Date:    date,
		Commit:  commit,
	}
}

func (bi *BuildInfo) Print() {
	fmt.Printf("Build version: %s\n", bi.getVersion())
	fmt.Printf("Build date: %s\n", bi.getDate())
	fmt.Printf("Build commit: %s\n", bi.getCommit())
}

func (bi *BuildInfo) getVersion() string {
	if bi.Version == "" {
		return "N/A"
	}
	return bi.Version
}

func (bi *BuildInfo) getDate() string {
	if bi.Date == "" {
		return "N/A"
	}
	return bi.Date
}

func (bi *BuildInfo) getCommit() string {
	if bi.Commit == "" {
		return "N/A"
	}
	return bi.Commit
}