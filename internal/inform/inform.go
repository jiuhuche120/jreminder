package inform

import (
	"context"
	"fmt"
	"sync"

	"github.com/jiuhuche120/jreminder/pkg/event"
	"github.com/jiuhuche120/jreminder/pkg/rule"
	"github.com/sirupsen/logrus"
)

type Inform struct {
	ch     chan *event.Event
	ctx    context.Context
	cancel context.CancelFunc
	rules  map[string]rule.Rule
	log    *logrus.Logger
	module string
}

func NewInform(ctx context.Context, cancel context.CancelFunc, ch chan *event.Event, log *logrus.Logger) *Inform {
	return &Inform{
		ch:     ch,
		ctx:    ctx,
		cancel: cancel,
		rules:  map[string]rule.Rule{},
		log:    log,
		module: "inform",
	}
}

func (i *Inform) RegisterRule(rule rule.Rule) error {
	_, ok := i.rules[rule.ID()]
	if ok {
		return fmt.Errorf("rule %v already registered", rule.ID())
	}
	i.rules[rule.ID()] = rule
	return nil
}

func (i *Inform) Start() error {
	if len(i.rules) == 0 {
		return fmt.Errorf("empty rule")
	}
	wg := sync.WaitGroup{}
	wg.Add(len(i.rules))
	for _, r := range i.rules {
		go func(r rule.Rule) {
			defer wg.Done()
			r.Call(i.ctx, i.ch, i.log)
			i.log.WithFields(logrus.Fields{
				"module": i.module,
			}).Infof("rule [%v] started successful", r.ID())
		}(r)
	}
	wg.Wait()

	i.log.WithFields(logrus.Fields{
		"module": i.module,
	}).Info("inform started successful")
	return nil
}

func (i *Inform) Stop() {
	i.cancel()
}
