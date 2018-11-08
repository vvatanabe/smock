package smock

import "fmt"

const (
	Name    = "smock"
	version = ""
)

var (
	commit string
	date   string
)

func FmtVersion() string {
	if commit == "" || date == "" {
		return version
	}
	return fmt.Sprintf("%s, build %s, date %s", version, commit, date)
}
