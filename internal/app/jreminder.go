package app

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/jiuhuche120/jreminder/internal/inform"
	"github.com/jiuhuche120/jreminder/internal/logger"
	"github.com/jiuhuche120/jreminder/internal/router"
	"github.com/jiuhuche120/jreminder/pkg/config"
	"github.com/jiuhuche120/jreminder/pkg/event"
	"github.com/jiuhuche120/jreminder/pkg/hook"
	"github.com/jiuhuche120/jreminder/pkg/rule"
	"github.com/sirupsen/logrus"
)

type Jreminder struct {
	Inform *inform.Inform
	Router *router.Router

	ctx    context.Context
	cancel context.CancelFunc
}

func NewJreminder() (*Jreminder, error) {
	var j Jreminder
	ctx, cancel := context.WithCancel(context.Background())
	j.ctx = ctx
	j.cancel = cancel
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}
	ch := make(chan *event.Event)
	l := logger.NewLogger(cfg)

	err = j.newInform(cfg, ch, l)
	if err != nil {
		return nil, err
	}
	err = j.newRouter(cfg, ch, l)
	if err != nil {
		return nil, err
	}
	return &j, nil
}

func (j *Jreminder) Start() error {
	if err := j.Inform.Start(); err != nil {
		return fmt.Errorf("start inform: %v", err)
	}
	if err := j.Router.Start(); err != nil {
		return fmt.Errorf("start router: %v", err)
	}
	return nil
}

func (j *Jreminder) Stop() {
	j.cancel()
}

func (j *Jreminder) newInform(cfg *config.Config, ch chan *event.Event, log *logrus.Logger) error {
	i := inform.NewInform(j.ctx, j.cancel, ch, log)
	rules, err := loadRules(cfg)
	if err != nil {
		return err
	}
	for _, r := range rules {
		err := i.RegisterRule(r)
		if err != nil {
			return err
		}
	}
	j.Inform = i
	return nil
}

func (j *Jreminder) newRouter(cfg *config.Config, ch chan *event.Event, log *logrus.Logger) error {
	r := router.NewRouter(j.ctx, j.cancel, ch, log)
	hooks, err := loadWebhooks(cfg)
	if err != nil {
		return err
	}
	for _, hk := range hooks {
		err := r.RegisterHook(hk)
		if err != nil {
			return err
		}
	}
	j.Router = r
	return nil
}

func loadRules(cfg *config.Config) ([]rule.Rule, error) {
	var rules []rule.Rule
	GitHubRuleOneMap := make(map[string]*config.CheckMainBranchMerged)
	GitHubRuleTwoMap := make(map[string]*config.CheckPullRequestTimeout)
	TeambitionRuleMap := make(map[string]*config.CheckTeambitionTimeout)
	for k, v := range cfg.Rules.CheckMainBranchMerged {
		ID := fmt.Sprintf("%v.%v", "checkMainBranchMerged", k)
		GitHubRuleOneMap[ID] = v
	}
	for k, v := range cfg.Rules.CheckPullRequestTimeout {
		ID := fmt.Sprintf("%v.%v", "checkPullRequestTimeout", k)
		GitHubRuleTwoMap[ID] = v
	}
	for k, v := range cfg.Rules.CheckTeambitionTimeout {
		ID := fmt.Sprintf("%v.%v", "checkTeambitionTimeout", k)
		TeambitionRuleMap[ID] = v
	}
	for k, v := range cfg.Repositories {
		for _, r := range v.Rules {
			rule1, ok1 := GitHubRuleOneMap[r]
			if ok1 {
				ruleID := fmt.Sprintf("%v.%v", k, r)
				bytes, err := ioutil.ReadFile(cfg.Holiday.Path)
				if err != nil {
					return nil, err
				}
				one := rule.NewGithubRuleOne(ruleID, cfg.Github.Token, cfg.Members, v.Webhook, string(bytes), v.Repository, v.Project, rule1)
				rules = append(rules, one)
			}
			rule2, ok2 := GitHubRuleTwoMap[r]
			if ok2 {
				ruleID := fmt.Sprintf("%v.%v", k, r)
				bytes, err := ioutil.ReadFile(cfg.Holiday.Path)
				if err != nil {
					return nil, err
				}
				two := rule.NewGithubRuleTwo(ruleID, cfg.Github.Token, cfg.Members, v.Webhook, string(bytes), v.Repository, v.Project, rule2)
				rules = append(rules, two)
			}
			if !ok1 && !ok2 {
				return nil, fmt.Errorf("error load rules %s", r)
			}
		}
	}
	for k, v := range cfg.Teambitions {
		for _, r := range v.Rules {
			teambitionRule, ok := TeambitionRuleMap[r]
			if ok {
				ruleID := fmt.Sprintf("%v.%v", k, r)
				bytes, err := ioutil.ReadFile(cfg.Holiday.Path)
				if err != nil {
					return nil, err
				}
				timeoutRule := rule.NewTeambitionTimeoutRule(ruleID, cfg.Account.Email, cfg.Account.Password, cfg.Members, v.Webhook, string(bytes), v.Project, v.App, teambitionRule.Cron)
				rules = append(rules, timeoutRule)
			} else {
				return nil, fmt.Errorf("error load rules %s", r)
			}
		}
	}
	return rules, nil
}

func loadWebhooks(cfg *config.Config) ([]hook.Webhook, error) {
	var hooks []hook.Webhook
	for k, v := range cfg.Webhook {
		dingTalk := hook.NewDingTalk(k, v.Webhook)
		hooks = append(hooks, dingTalk)
	}
	return hooks, nil
}
