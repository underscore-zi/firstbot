package subscriptions

import (
	"github.com/sirupsen/logrus"
	"github.com/underscore-zi/firstbot/pkg/eventsub"
)

// OnSubscribed is a simple implementation of the SubscriptionHandler's OnSubscribed method that can be called in-place
// of implementing it yourself. It simply logs the subscription creation.
func OnSubscribed(logger *logrus.Logger, sub eventsub.Subscription) {
	logger.WithFields(logrus.Fields{
		"type":   sub.Type,
		"id":     sub.ID,
		"status": sub.Status,
	}).Info("Subscription created")
}

// OnRevoke is a simple implementation of the SubscriptionHandler's OnRevoke method that can be called in-place
// of implementing it yourself. It simply logs the subscription revocation.
func OnRevoke(logger *logrus.Logger, sub eventsub.Subscription) {
	logger.WithFields(logrus.Fields{
		"type":   sub.Type,
		"id":     sub.ID,
		"status": sub.Status,
	}).Info("Subscription revoked")
}

// OnEvent is a simple implementation of the SubscriptionHandler's OnEvent method that can be called in-place
// of implementing it yourself. It simply logs the event.
func OnEvent(logger *logrus.Logger, sub eventsub.Subscription, event eventsub.Event) {
	logger.WithFields(logrus.Fields{
		"type":   sub.Type,
		"id":     sub.ID,
		"status": sub.Status,
		"event":  event,
	}).Info("Event received")
}
