package gorush

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/appleboy/go-fcm"
	"github.com/appleboy/gorush/core"
	"github.com/appleboy/gorush/logx"
	"github.com/appleboy/gorush/status"
	"github.com/sirupsen/logrus"
)

// InitFCMClient use for initialize FCM Client.
func InitFCMClient(key string) (*fcm.Client, error) {
	var err error

	if key == "" {
		return nil, errors.New("Missing Android API Key")
	}

	if key != PushConf.Android.APIKey {
		return fcm.NewClient(key)
	}

	if FCMClient == nil {
		FCMClient, err = fcm.NewClient(key)
		return FCMClient, err
	}

	return FCMClient, nil
}

// GetAndroidNotification use for define Android notification.
// HTTP Connection Server Reference for Android
// https://firebase.google.com/docs/cloud-messaging/http-server-ref
func GetAndroidNotification(req PushNotification) (*fcm.Message, error) {
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

	jsonMarshall, err := json.Marshal(notification)
	if err != nil {
		LogError.Error("Failed to marshal the default message! Error is " + err.Error())
		return nil, err
	}

	LogAccess.Debugf("Default message is %s", string(jsonMarshall))

	return notification, nil
}

// PushToAndroid provide send notification to Android server.
func PushToAndroid(req PushNotification) {
	logx.LogAccess.Debug("Start push notification for Android")

	var (
		client     *fcm.Client
		retryCount = 0
		maxRetry   = PushConf.Android.MaxRetry
	)

	if req.Retry > 0 && req.Retry < maxRetry {
		maxRetry = req.Retry
	}

	// check message
	err := CheckMessage(req)
	if err != nil {
		logx.LogError.Error("request error: " + err.Error())
		return
	}

Retry:
	notification, _ := GetAndroidNotification(req)

	if req.APIKey != "" {
		client, err = InitFCMClient(req.APIKey)
	} else {
		client, err = InitFCMClient(PushConf.Android.APIKey)
	}

	if err != nil {
		// FCM server error
		logx.LogError.Error("FCM server error: " + err.Error())
		return
	}

	res, err := client.Send(notification)
	if err != nil {
		// Send Message error
		logx.LogError.Error("FCM server send message error: " + err.Error())

		if req.IsTopic() {
			if PushConf.Core.Sync {
				req.AddLog(createLogPushEntry(core.FailedPush, req.To, req, err))
			} else if PushConf.Core.FeedbackURL != "" {
				go func(logger *logrus.Logger, log logx.LogPushEntry, url string, timeout int64) {
					err := DispatchFeedback(log, url, timeout)
					if err != nil {
						logger.Error(err)
					}
				}(logx.LogError, createLogPushEntry(core.FailedPush, req.To, req, err), PushConf.Core.FeedbackURL, PushConf.Core.FeedbackTimeout)
			}
			status.StatStorage.AddAndroidError(1)
		} else {
			for _, token := range req.Tokens {
				if PushConf.Core.Sync {
					req.AddLog(createLogPushEntry(core.FailedPush, token, req, err))
				} else if PushConf.Core.FeedbackURL != "" {
					go func(logger *logrus.Logger, log logx.LogPushEntry, url string, timeout int64) {
						err := DispatchFeedback(log, url, timeout)
						if err != nil {
							logger.Error(err)
						}
					}(logx.LogError, createLogPushEntry(core.FailedPush, token, req, err), PushConf.Core.FeedbackURL, PushConf.Core.FeedbackTimeout)
				}
			}
			status.StatStorage.AddAndroidError(int64(len(req.Tokens)))
		}
		return
	}

	if !req.IsTopic() {
		logx.LogAccess.Debug(fmt.Sprintf("Android Success count: %d, Failure count: %d", res.Success, res.Failure))
	}

	status.StatStorage.AddAndroidSuccess(int64(res.Success))
	status.StatStorage.AddAndroidError(int64(res.Failure))

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

			logPush(core.FailedPush, to, req, result.Error)
			if PushConf.Core.Sync {
				req.AddLog(createLogPushEntry(core.FailedPush, to, req, result.Error))
			} else if PushConf.Core.FeedbackURL != "" {
				go func(logger *logrus.Logger, log logx.LogPushEntry, url string, timeout int64) {
					err := DispatchFeedback(log, url, timeout)
					if err != nil {
						logger.Error(err)
					}
				}(logx.LogError, createLogPushEntry(core.FailedPush, to, req, result.Error), PushConf.Core.FeedbackURL, PushConf.Core.FeedbackTimeout)
			}
			continue
		}

		logPush(core.SucceededPush, to, req, nil)
	}

	// result from Send messages to topics
	if req.IsTopic() {
		to := ""
		if req.To != "" {
			to = req.To
		} else {
			to = req.Condition
		}
		logx.LogAccess.Debug("Send Topic Message: ", to)
		// Success
		if res.MessageID != 0 {
			logPush(core.SucceededPush, to, req, nil)
		} else {
			// failure
			logPush(core.FailedPush, to, req, res.Error)
			if PushConf.Core.Sync {
				req.AddLog(createLogPushEntry(core.FailedPush, to, req, res.Error))
			}
		}
	}

	// Device Group HTTP Response
	if len(res.FailedRegistrationIDs) > 0 {
		newTokens = append(newTokens, res.FailedRegistrationIDs...)

		logPush(core.FailedPush, notification.To, req, errors.New("device group: partial success or all fails"))
		if PushConf.Core.Sync {
			req.AddLog(createLogPushEntry(core.FailedPush, notification.To, req, errors.New("device group: partial success or all fails")))
		}
	}

	if len(newTokens) > 0 && retryCount < maxRetry {
		retryCount++

		// resend fail token
		req.Tokens = newTokens
		goto Retry
	}
}

func createLogPushEntry(status, token string, req PushNotification, err error) logx.LogPushEntry {
	return logx.GetLogPushEntry(&logx.InputLog{
		ID:        req.ID,
		Status:    status,
		Token:     token,
		Message:   req.Message,
		Platform:  req.Platform,
		Error:     err,
		HideToken: PushConf.Log.HideToken,
		Format:    PushConf.Log.Format,
	})
}

func logPush(status, token string, req PushNotification, err error) {
	logx.LogPush(&logx.InputLog{
		ID:        req.ID,
		Status:    status,
		Token:     token,
		Message:   req.Message,
		Platform:  req.Platform,
		Error:     err,
		HideToken: PushConf.Log.HideToken,
		Format:    PushConf.Log.Format,
	})
}
