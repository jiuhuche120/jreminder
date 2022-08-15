package router

import (
	"context"
	"fmt"

	"github.com/jiuhuche120/jreminder/pkg/event"
	"github.com/jiuhuche120/jreminder/pkg/hook"
	"github.com/sirupsen/logrus"
)

type Router struct {
	ch     chan *event.Event
	ctx    context.Context
	cancel context.CancelFunc
	hooks  map[string]hook.Webhook
	log    *logrus.Logger
	module string
}

func NewRouter(ctx context.Context, cancel context.CancelFunc, ch chan *event.Event, log *logrus.Logger) *Router {
	return &Router{
		ch:     ch,
		ctx:    ctx,
		cancel: cancel,
		hooks:  make(map[string]hook.Webhook),
		log:    log,
		module: "router",
	}
}

func (r *Router) RegisterHook(hook hook.Webhook) error {
	_, ok := r.hooks[hook.ID()]
	if ok {
		return fmt.Errorf("register hook already registered: %v", hook.ID())
	}
	r.hooks[hook.ID()] = hook
	return nil
}

func (r *Router) Start() error {
	if len(r.hooks) == 0 {
		return fmt.Errorf("empty hook")
	}
	go func() {
		for {
			select {
			case e := <-r.ch:
				//err := r.callHook(e)
				//if err != nil {
				//	r.log.WithFields(logrus.Fields{
				//		"module": r.module,
				//		"id":     e.ID,
				//		"error":  err,
				//	}).Error("callHook failed")
				//}
				r.log.WithFields(logrus.Fields{
					"module": r.module,
					"id":     e.ID,
				}).Info("callHook succeeded")
			case <-r.ctx.Done():
				return
			}
		}
	}()

	r.log.WithFields(logrus.Fields{
		"module": r.module,
	}).Info("Router started succeeded")
	return nil
}

func (r *Router) Stop() {
	r.cancel()
}

func (r *Router) callHook(event *event.Event) error {
	for _, webhook := range event.Webhooks {
		hk, ok := r.hooks[webhook]
		if !ok {
			return fmt.Errorf("hook %v not found", webhook)
		}
		err := hk.Call(event)
		if err != nil {
			return err
		}
	}
	return nil
}
