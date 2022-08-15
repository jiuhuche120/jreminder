package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	cfg, err := LoadConfig()
	require.Nil(t, err)
	// log level
	require.Equal(t, "info", cfg.Log.Level)
	// github token
	require.Equal(t, "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", cfg.Github.Token)
	// repository info
	require.Equal(t, "meshplus", cfg.Repositories["bitxhub"].Repository)
	require.Equal(t, "bitxhub", cfg.Repositories["bitxhub"].Project)
	require.Equal(t, []string{"checkMainBranchMerged.rule1", "checkPullRequestTimeout.rule1"}, cfg.Repositories["bitxhub"].Rules)
	require.Equal(t, "dingtalk", cfg.Repositories["bitxhub"].Webhook)
	require.Equal(t, "meshplus", cfg.Repositories["pier"].Repository)
	require.Equal(t, "pier", cfg.Repositories["pier"].Project)
	require.Equal(t, []string{"checkMainBranchMerged.rule1", "checkPullRequestTimeout.rule1"}, cfg.Repositories["pier"].Rules)
	require.Equal(t, "dingtalk", cfg.Repositories["pier"].Webhook)
	// rule info
	require.Equal(t, "master", cfg.Rules.CheckMainBranchMerged["rule1"].Base)
	require.Equal(t, "release*", cfg.Rules.CheckMainBranchMerged["rule1"].Head)
	require.Equal(t, "0 30 16 * * *", cfg.Rules.CheckMainBranchMerged["rule1"].Cron)
	require.Equal(t, "72h", cfg.Rules.CheckPullRequestTimeout["rule1"].Timeout)
	require.Equal(t, "0 30 16 * * *", cfg.Rules.CheckPullRequestTimeout["rule1"].Cron)
	// member info
	require.Equal(t, 8, len(cfg.Members))
	// webhook info
	require.Equal(t, "钉钉机器人", cfg.Webhook["dingtalk"].Name)
	require.Equal(t, "https://oapi.dingtalk.com/robot/send?access_token=89240674ce78edb5f96b7d8607a3516b81ab55ce7cb60aa9c9551dabefefbf49", cfg.Webhook["dingtalk"].Webhook)
}
