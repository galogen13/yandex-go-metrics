package buildinfo

import (
	"fmt"
)

const BuildInfoNotAvaluable = "N/A"

func PrintBuildInfo(buildVersion, buildDate, buildCommit string) {

	buildInfo := `
Build version: %s
Build date: %s
Build commit: %s
`
	fmt.Printf(buildInfo, buildVersion, buildDate, buildCommit)

}
