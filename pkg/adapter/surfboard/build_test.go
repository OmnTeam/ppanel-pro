package surfboard

import (
	"testing"
	"time"

	"github.com/OmnTeam/ppanel-pro/pkg/adapter/proxy"

	"github.com/OmnTeam/ppanel-pro/pkg/uuidx"
)

func TestBuildSurfboard(t *testing.T) {
	siteName := "test"
	user := UserInfo{
		UUID:         uuidx.NewUUID().String(),
		Upload:       0,
		Download:     0,
		TotalTraffic: 0,
		ExpiredDate:  time.Now().AddDate(0, 1, 1),
		SubscribeURL: "https://test.com",
	}
	conf := BuildSurfboard(proxy.Adapter{}, siteName, user)
	t.Log(string(conf))
}
