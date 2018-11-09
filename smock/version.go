package smock

import "fmt"

const (
	Name    = "smock"
	version = "0.9.1"
)

var (
	commit string
	date   string
)

func FmtVersion() string {
	if commit == "" || date == "" {
		return version
	}
	return fmt.Sprintf("v%s, build %s, date %s", version, commit, date)
}
