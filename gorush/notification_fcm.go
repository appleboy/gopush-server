package gorush

import (
	"errors"
	"fmt"

	"github.com/appleboy/go-fcm"
	"github.com/sirupsen/logrus"
)

// InitFCMClient use for initialize FCM Client.
func InitFCMClient(tenantId string, key string) (*fcm.Client, error) {
	var err error

	if key == "" {
		return nil, errors.New("missing Android API Key for tenant " + tenantId)
	}

	if FCMClients[tenantId] == nil {
		FCMClients[tenantId], err = fcm.NewClient(key)
		return FCMClients[tenantId], err
	}

	return FCMClients[tenantId], nil
}

// GetAndroidNotification use for define Android notification.
// HTTP Connection Server Reference for Android
// https://firebase.google.com/docs/cloud-messaging/http-server-ref
func GetAndroidNotification(req PushNotification) *fcm.Message {
	notification := &fcm.Message{
		To:                    req.To,
		Condition:             req.Condition,
		CollapseKey:           req.CollapseKey,
		ContentAvailable:      req.ContentAvailable,
		MutableContent:        req.MutableContent,
		DelayWhileIdle:        req.DelayWhileIdle,
		TimeToLive:            req.TimeToLive,
		RestrictedPackageName: req.RestrictedPackageName,
		DryRun:                req.DryRun,
	}

	if len(req.Tokens) > 0 {
		notification.RegistrationIDs = req.Tokens
	}

	if req.Priority == "high" || req.Priority == "normal" {
		notification.Priority = req.Priority
	}

	// Add another field
	if len(req.Data) > 0 {
		notification.Data = make(map[string]interface{})
		for k, v := range req.Data {
			notification.Data[k] = v
		}
	}

	n := &fcm.Notification{}
	isNotificationSet := false
	if req.Notification != nil {
		isNotificationSet = true
		n = req.Notification
	}

	if len(req.Message) > 0 {
		isNotificationSet = true
		n.Body = req.Message
	}

	if len(req.Title) > 0 {
		isNotificationSet = true
		n.Title = req.Title
	}

	if len(req.Image) > 0 {
		isNotificationSet = true
		n.Image = req.Image
	}

	if v, ok := req.Sound.(string); ok && len(v) > 0 {
		isNotificationSet = true
		n.Sound = v
	}

	if isNotificationSet {
		notification.Notification = n
	}

	// handle iOS apns in fcm

	if len(req.Apns) > 0 {
		notification.Apns = req.Apns
	}

	return notification
}

// PushToAndroid provide send notification to Android server.
func PushToAndroid(req PushNotification) {
	LogAccess.Debug("Start push notification for Android")
	if req.TenantId == "" {
		LogError.Error("missing tenant id for Android notification")
		return
	}

	var (
		client     *fcm.Client
		retryCount = 0
		maxRetry   = PushConf.Tenants[req.TenantId].Android.MaxRetry
	)

	if req.Retry > 0 && req.Retry < maxRetry {
		maxRetry = req.Retry
	}

	// check message
	err := CheckMessage(req)
	if err != nil {
		LogError.Error("request error: " + err.Error())
		return
	}

Retry:
	notification := GetAndroidNotification(req)

	if req.APIKey != "" {
		client, err = InitFCMClient(req.TenantId, req.APIKey)
	} else {
		client, err = InitFCMClient(req.TenantId, PushConf.Tenants[req.TenantId].Android.APIKey)
	}

	if err != nil {
		// FCM server error
		LogError.Error("FCM server error: " + err.Error())
		return
	}

	res, err := client.Send(notification)
	if err != nil {
		// Send Message error
		LogError.Error("FCM server send message error: " + err.Error())

		if req.IsTopic() {
			if PushConf.Core.Sync {
				req.AddLog(getLogPushEntry(FailedPush, req.To, req, err))
			} else if PushConf.Core.FeedbackURL != "" {
				go func(logger *logrus.Logger, log LogPushEntry, url string, timeout int64) {
					err := DispatchFeedback(log, url, timeout)
					if err != nil {
						logger.Error(err)
					}
				}(LogError, getLogPushEntry(FailedPush, req.To, req, err), PushConf.Core.FeedbackURL, PushConf.Core.FeedbackTimeout)
			}
			StatStorage.AddAndroidError(1)
		} else {
			for _, token := range req.Tokens {
				if PushConf.Core.Sync {
					req.AddLog(getLogPushEntry(FailedPush, token, req, err))
				} else if PushConf.Core.FeedbackURL != "" {
					go func(logger *logrus.Logger, log LogPushEntry, url string, timeout int64) {
						err := DispatchFeedback(log, url, timeout)
						if err != nil {
							logger.Error(err)
						}
					}(LogError, getLogPushEntry(FailedPush, token, req, err), PushConf.Core.FeedbackURL, PushConf.Core.FeedbackTimeout)
				}
			}
			StatStorage.AddAndroidError(int64(len(req.Tokens)))
		}
		return
	}

	if !req.IsTopic() {
		LogAccess.Debug(fmt.Sprintf("Android Success count: %d, Failure count: %d", res.Success, res.Failure))
	}

	StatStorage.AddAndroidSuccess(int64(res.Success))
	StatStorage.AddAndroidError(int64(res.Failure))

	var newTokens []string
	// result from Send messages to specific devices
	for k, result := range res.Results {
		to := ""
		if k < len(req.Tokens) {
			to = req.Tokens[k]
		} else {
			to = req.To
		}

		if result.Error != nil {
			// We should retry only "retryable" statuses. More info about response:
			// https://firebase.google.com/docs/cloud-messaging/http-server-ref#downstream-http-messages-plain-text
			if !result.Unregistered() {
				newTokens = append(newTokens, to)
			}

			LogPush(FailedPush, to, req, result.Error)
			if PushConf.Core.Sync {
				req.AddLog(getLogPushEntry(FailedPush, to, req, result.Error))
			} else if PushConf.Core.FeedbackURL != "" {
				go func(logger *logrus.Logger, log LogPushEntry, url string, timeout int64) {
					err := DispatchFeedback(log, url, timeout)
					if err != nil {
						logger.Error(err)
					}
				}(LogError, getLogPushEntry(FailedPush, to, req, result.Error), PushConf.Core.FeedbackURL, PushConf.Core.FeedbackTimeout)
			}
			continue
		}

		LogPush(SucceededPush, to, req, nil)
	}

	// result from Send messages to topics
	if req.IsTopic() {
		to := ""
		if req.To != "" {
			to = req.To
		} else {
			to = req.Condition
		}
		LogAccess.Debug("Send Topic Message: ", to)
		// Success
		if res.MessageID != 0 {
			LogPush(SucceededPush, to, req, nil)
		} else {
			// failure
			LogPush(FailedPush, to, req, res.Error)
			if PushConf.Core.Sync {
				req.AddLog(getLogPushEntry(FailedPush, to, req, res.Error))
			}
		}
	}

	// Device Group HTTP Response
	if len(res.FailedRegistrationIDs) > 0 {
		newTokens = append(newTokens, res.FailedRegistrationIDs...)

		LogPush(FailedPush, notification.To, req, errors.New("device group: partial success or all fails"))
		if PushConf.Core.Sync {
			req.AddLog(getLogPushEntry(FailedPush, notification.To, req, errors.New("device group: partial success or all fails")))
		}
	}

	if len(newTokens) > 0 && retryCount < maxRetry {
		retryCount++

		// resend fail token
		req.Tokens = newTokens
		goto Retry
	}
}
