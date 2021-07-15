package model

// This is a list of all the current versions including any patches.
// It should be maintained in chronological order with most current
// release at the front of the list.
var versions = []string{
	"0.1.0",
}

var (
	CurrentVersion string = versions[0]
	BuildNumber    string
	BuildTime      string
	BuildHash      string
)
