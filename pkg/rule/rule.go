package rule

import (
	"context"

	"github.com/jiuhuche120/jreminder/pkg/event"
	"github.com/sirupsen/logrus"
)

type Rule interface {
	ID() string
	Call(ctx context.Context, ch chan *event.Event, log *logrus.Logger)
}
