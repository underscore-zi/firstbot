package subscriptions

import (
	"github.com/sirupsen/logrus"
	"github.com/underscore-zi/firstbot/pkg/eventsub"
)

type StreamOffline struct {
	EventSub  *eventsub.Client
	Logger    *logrus.Logger
	ChannelID string
	Callback  func(eventsub.Event)
}

func (s StreamOffline) OnSubscribed(sub eventsub.Subscription) {
	OnSubscribed(s.Logger, sub)
}

func (s StreamOffline) OnRevoke(sub eventsub.Subscription) {
	OnRevoke(s.Logger, sub)
}

func (s StreamOffline) OnEvent(sub eventsub.Subscription, event eventsub.Event) {
	if s.Callback != nil {
		s.Callback(event)
	} else {
		OnEvent(s.Logger, sub, event)
	}
}

func (s StreamOffline) Register() (err error) {
	cond := map[string]interface{}{
		"broadcaster_user_id": s.ChannelID,
	}
	err = s.EventSub.Subscribe("stream.offline", "1", cond, s)
	return
}
