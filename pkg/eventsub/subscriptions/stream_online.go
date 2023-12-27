package subscriptions

import (
	"github.com/sirupsen/logrus"
	"github.com/underscore-zi/firstbot/pkg/eventsub"
)

type EventCallback func(event eventsub.Event)

type StreamOnline struct {
	EventSub  *eventsub.Client
	Logger    *logrus.Logger
	ChannelID string
	Callback  EventCallback
}

func (s StreamOnline) OnSubscribed(sub eventsub.Subscription) {
	OnSubscribed(s.Logger, sub)
}

func (s StreamOnline) OnRevoke(sub eventsub.Subscription) {
	OnRevoke(s.Logger, sub)
}

func (s StreamOnline) OnEvent(sub eventsub.Subscription, event eventsub.Event) {
	if s.Callback != nil {
		s.Callback(event)
	} else {
		OnEvent(s.Logger, sub, event)
	}
}

func (s StreamOnline) Register() (err error) {
	cond := map[string]interface{}{
		"broadcaster_user_id": s.ChannelID,
	}
	err = s.EventSub.Subscribe("stream.online", "1", cond, s)
	return
}
