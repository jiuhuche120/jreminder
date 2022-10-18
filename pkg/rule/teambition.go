package rule

import (
	"context"
	"time"

	"github.com/jiuhuche120/jreminder/pkg/config"
	"github.com/jiuhuche120/jreminder/pkg/event"
	"github.com/jiuhuche120/jreminder/pkg/tool"
	"github.com/jiuhuche120/jreminder/pkg/types"
	"github.com/procyon-projects/chrono"
	"github.com/sirupsen/logrus"
)

var _ Rule = (*TeambitionTimeoutRule)(nil)

type TeambitionTimeoutRule struct {
	ruleID            string
	email             string
	password          string
	teambitionMembers map[string]string
	scheduler         chrono.TaskScheduler
	webhooks          []string
	holiday           string

	project string
	app     string
	cron    string
}

func NewTeambitionTimeoutRule(ruleID string, email string, password string, members map[string]config.Member, webhooks []string, holiday, project, app, cron string) *TeambitionTimeoutRule {
	teambitionMembers := make(map[string]string)
	for _, v := range members {
		if v.Name != "" {
			teambitionMembers[v.Name] = v.Phone
		}
	}
	return &TeambitionTimeoutRule{
		ruleID:            ruleID,
		email:             email,
		password:          password,
		teambitionMembers: teambitionMembers,
		scheduler:         chrono.NewDefaultTaskScheduler(),
		webhooks:          webhooks,
		holiday:           holiday,

		project: project,
		app:     app,
		cron:    cron,
	}
}

func (t *TeambitionTimeoutRule) ID() string {
	return t.ruleID
}

func (t *TeambitionTimeoutRule) Call(ctx context.Context, ch chan *event.Event, log *logrus.Logger) {
	task, err := t.scheduler.ScheduleWithCron(func(ctx context.Context) {
		if tool.IsWorkingDay(t.holiday) {
			cookie, err := tool.GetTeambitionCookie(t.email, t.password)
			if err != nil {
				log.WithFields(logrus.Fields{
					"id":    t.ruleID,
					"error": err,
				}).Error("get teambition cookie failed")
				return
			}
			tasks, err := tool.GetTeambitionAllSubTask(cookie, t.project, t.app)
			if err != nil {
				log.WithFields(logrus.Fields{
					"id":    t.ruleID,
					"error": err,
				}).Error("get teambition subtasks failed")
				return
			}
			var hookTasks []*types.SubTask
			for _, task := range tasks {
				if !task.IsDone && (task.StartDate.IsZero() || task.DueDate.IsZero() || time.Since(task.DueDate) > 0) {
					log.WithFields(logrus.Fields{
						"id": t.ruleID,
					}).Infof("task [%v] is timed out", task.Content)
					task.DingTalk = t.teambitionMembers[task.Executor.Name]
					hookTasks = append(hookTasks, task)
				}
			}
			if len(hookTasks) != 0 {
				ch <- event.NewEvent(t.ruleID, t.webhooks, tool.NewTeambitionMsg(hookTasks, "tb任务存在异常"))
			}
		} else {
			log.WithFields(logrus.Fields{
				"id": t.ruleID,
			}).Info("today is not working day, skip")
			return
		}
	}, t.cron)
	if err != nil {
		log.WithFields(logrus.Fields{
			"id":    t.ruleID,
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
