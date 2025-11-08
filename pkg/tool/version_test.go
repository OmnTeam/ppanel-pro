package tool

import (
	"testing"

	"github.com/OmnTeam/ppanel-pro/pkg/constant"
)

func TestExtractVersionNumber(t *testing.T) {
	versionNumber := ExtractVersionNumber(constant.Version)
	t.Log(versionNumber)
}
