package rule

import (
	"context"
	"regexp"
	"time"

	"github.com/jiuhuche120/jreminder/pkg/config"
	"github.com/jiuhuche120/jreminder/pkg/event"
	"github.com/jiuhuche120/jreminder/pkg/tool"
	"github.com/procyon-projects/chrono"
	"github.com/sirupsen/logrus"
)

var _ Rule = (*GithubRuleOne)(nil)
var _ Rule = (*GithubRuleTwo)(nil)

type GithubRuleOne struct {
	ruleID    string
	token     string
	members   map[string]config.Member
	scheduler chrono.TaskScheduler
	webhooks  []string
	// check main branch merged config
	repository string
	project    string
	base       string
	head       string
	cron       string
}

func NewGithubRuleOne(ruleID, token string, members map[string]config.Member, webhooks []string, repository, project string, rule *config.CheckMainBranchMerged) *GithubRuleOne {
	return &GithubRuleOne{
		ruleID:    ruleID,
		token:     token,
		members:   members,
		scheduler: chrono.NewDefaultTaskScheduler(),
		webhooks:  webhooks,

		repository: repository,
		project:    project,
		base:       rule.Base,
		head:       rule.Head,
		cron:       rule.Cron,
	}
}

func (g *GithubRuleOne) ID() string {
	return g.ruleID
}

func (g *GithubRuleOne) Call(ctx context.Context, ch chan *event.Event, log *logrus.Logger) {
	task, err := g.scheduler.ScheduleWithCron(func(ctx context.Context) {
		isWorkingDay, err := tool.IsWorkingDay()
		if err != nil {
			log.WithFields(logrus.Fields{
				"id":    g.ruleID,
				"error": err,
			}).Error("get working day failed")
		}
		if isWorkingDay {
			pulls, err := tool.GetAllPullRequests(g.token, g.repository, g.project)
			if err != nil {
				log.WithFields(logrus.Fields{
					"id":    g.ruleID,
					"error": err,
				}).Error("get all pull requests failed")
			}
			var hookPulls []tool.PullRequest
			for i := 0; i < len(pulls); i++ {
				reg := regexp.MustCompile(g.head)
				if pulls[i].State == "open" && reg.FindString(pulls[i].Base.Ref) != "" {
					flag := false
					for j := 0; j < len(pulls); j++ {
						if i == j {
							continue
						}
						// open status pull request merged to master
						if i != j && pulls[i].Title == pulls[j].Title && pulls[j].Base.Ref == g.base && pulls[j].State == "open" {
							flag = true
							break
						}
						// close status pull request merged to master
						merged, err := tool.IsMerged(pulls[j])
						if err != nil {
							break
						}
						if i != j && pulls[i].Title == pulls[j].Title && pulls[j].Base.Ref == g.base && pulls[j].State == "closed" && merged {
							flag = true
							break
						}
					}
					if !flag {
						log.WithFields(logrus.Fields{
							"id": g.ruleID,
						}).Infof("the pull request [%v] from %v to %v is lost", pulls[i].Title, pulls[i].Head.Ref, g.base)
						pulls[i].DingTalk = g.members[pulls[i].User.Login].Phone
						hookPulls = append(hookPulls, pulls[i])
					}
				}
			}
			if len(hookPulls) > 0 {
				ch <- event.NewEvent(g.ruleID, g.webhooks, tool.NewMsg(hookPulls, "需要合并分支到master分支"))
			}
		} else {
			log.WithFields(logrus.Fields{
				"id": g.ruleID,
			}).Error("today is not working day, skip")
		}
	}, g.cron)
	if err != nil {
		log.WithFields(logrus.Fields{
			"id":    g.ruleID,
			"error": err,
		}).Error("task has been scheduled")
	}
	// ctx is done cancel task
	go func() {
		for {
			select {
			case <-ctx.Done():
				task.Cancel()
				return
			}
		}
	}()
}

type GithubRuleTwo struct {
	ruleID    string
	token     string
	members   map[string]config.Member
	scheduler chrono.TaskScheduler
	webhooks  []string
	// check pull request timeout
	repository string
	project    string
	timeout    string
	cron       string
}

func NewGithubRuleTwo(ruleID, token string, members map[string]config.Member, webhooks []string, repository, project string, rule *config.CheckPullRequestTimeout) *GithubRuleTwo {
	return &GithubRuleTwo{
		ruleID:    ruleID,
		token:     token,
		members:   members,
		scheduler: chrono.NewDefaultTaskScheduler(),
		webhooks:  webhooks,

		repository: repository,
		project:    project,
		timeout:    rule.Timeout,
		cron:       rule.Cron,
	}
}

func (g *GithubRuleTwo) ID() string {
	return g.ruleID
}

func (g *GithubRuleTwo) Call(ctx context.Context, ch chan *event.Event, log *logrus.Logger) {
	task, err := g.scheduler.ScheduleWithCron(func(ctx context.Context) {
		isWorkingDay, err := tool.IsWorkingDay()
		if err != nil {
			log.WithFields(logrus.Fields{
				"id":    g.ruleID,
				"error": err,
			}).Error("get working day failed")
		}
		if isWorkingDay {
			pulls, err := tool.GetAllPullRequests(g.token, g.repository, g.project)
			if err != nil {
				log.WithFields(logrus.Fields{
					"id":    g.ruleID,
					"error": err,
				}).Error("get all pull requests failed")
			}
			var hookPulls []tool.PullRequest
			for i := 0; i < len(pulls); i++ {
				if pulls[i].State == "open" {
					timeout, err := time.ParseDuration(g.timeout)
					if err != nil {
						log.WithFields(logrus.Fields{
							"id":    g.ruleID,
							"error": err,
						}).Error("parse timeout err")
					}
					if time.Since(pulls[i].CreateAt) >= timeout {
						log.WithFields(logrus.Fields{
							"id": g.ruleID,
						}).Infof("the pull request [%v] is timeout", pulls[i].Title)
						pulls[i].DingTalk = g.members[pulls[i].User.Login].Phone
						hookPulls = append(hookPulls, pulls[i])
					}
				}
			}
			if len(hookPulls) > 0 {
				ch <- event.NewEvent(g.ruleID, g.webhooks, tool.NewMsg(hookPulls, "PR存活超时"))
			}
		} else {
			log.WithFields(logrus.Fields{
				"id": g.ruleID,
			}).Error("today is not working day, skip")
		}
	}, g.cron)
	if err != nil {
		log.WithFields(logrus.Fields{
			"id":    g.ruleID,
			"error": err,
		}).Error("task has been scheduled")
	}
	// ctx is done cancel task
	go func() {
		for {
			select {
			case <-ctx.Done():
				task.Cancel()
				return
			}
		}
	}()
}
