package pkggogit

import (
	"github.com/spf13/viper"

	defs "github.com/alphacodinggroup/ponti-backend/pkg/repo-tools/go-git/v5/defs"
)

func Bootstrap(repoRemoteUrlKey, repoLocalPathKey, repoBranchKey string) (defs.Client, error) {
	config := newConfig(
		viper.GetString(repoRemoteUrlKey),
		viper.GetString(repoLocalPathKey),
		viper.GetString(repoBranchKey),
	)

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return newClient(config)
}
