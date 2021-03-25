package gorush

import (
	"github.com/appleboy/gorush/config"
	"github.com/appleboy/gorush/storage"

	"github.com/appleboy/go-fcm"
	"github.com/msalihkarakasli/go-hms-push/push/core"
	"github.com/sideshow/apns2"
	"github.com/sirupsen/logrus"
)

var (
	// PushConf is gorush config
	PushConf config.ConfYaml
	// QueueNotification is chan type
	QueueNotification chan PushNotification
	// ApnsClient is apns client
	ApnsClients map[string]*apns2.Client
	// FCMClient is apns client
	FCMClients map[string]*fcm.Client
	// HMSClient is Huawei push client
	HMSClients map[string]*core.HMSClient
	// LogAccess is log server request log
	LogAccess *logrus.Logger
	// LogError is log server error log
	LogError *logrus.Logger
	// StatStorage implements the storage interface
	StatStorage storage.Storage
	// MaxConcurrentIOSPushes pool to limit the number of concurrent iOS pushes
	MaxConcurrentIOSPushes map[string]chan struct{}
)
