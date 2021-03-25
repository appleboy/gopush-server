package gorush

import (
	"context"
	"log"
	"os"
	"sync"
	"testing"

	"github.com/appleboy/go-fcm"
	"github.com/appleboy/gorush/config"
	"github.com/stretchr/testify/assert"
)

func init() {
	PushConf, _ = config.LoadConf("")
	if err := InitLog(); err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	wg := &sync.WaitGroup{}
	wg.Add(int(PushConf.Core.WorkerNum))
	FCMClients = make(map[string]*fcm.Client)
	InitWorkers(ctx, wg, PushConf.Core.WorkerNum, PushConf.Core.QueueNum)

	if err := InitAppStatus(); err != nil {
		log.Fatal(err)
	}
}

func TestMissingAndroidAPIKey(t *testing.T) {
	PushConf, _ = config.LoadConf("")
	tenantId := "tenant_id1"
	tenant := PushConf.Tenants[tenantId]

	tenant.Android.Enabled = true
	tenant.Android.APIKey = ""

	err := CheckPushConf()

	assert.Error(t, err)
	assert.Equal(t, "missing Android API Key for tenant "+tenantId, err.Error())
}

func TestMissingKeyForInitFCMClient(t *testing.T) {
	tenantId := "1"
	client, err := InitFCMClient(tenantId, "")

	assert.Nil(t, client)
	assert.Error(t, err)
	assert.Equal(t, "missing Android API Key for tenant "+tenantId, err.Error())
}

func TestPushToAndroidWrongToken(t *testing.T) {
	PushConf, _ = config.LoadConf("")
	tenantId := "tenant_id1"
	tenant := PushConf.Tenants[tenantId]

	tenant.Android.Enabled = true
	tenant.Android.APIKey = os.Getenv("ANDROID_API_KEY")

	req := PushNotification{
		TenantId: tenantId,
		Tokens:   []string{"aaaaaa", "bbbbb"},
		Platform: PlatformAndroid,
		Message:  "Welcome",
	}

	// Android Success count: 0, Failure count: 2
	PushToAndroid(req)
}

func TestPushToAndroidRightTokenForJSONLog(t *testing.T) {
	PushConf, _ = config.LoadConf("")
	tenantId := "tenant_id1"
	tenant := PushConf.Tenants[tenantId]

	tenant.Android.Enabled = true
	tenant.Android.APIKey = os.Getenv("ANDROID_API_KEY")
	// log for json
	PushConf.Log.Format = "json"

	androidToken := os.Getenv("ANDROID_TEST_TOKEN")

	req := PushNotification{
		TenantId: tenantId,
		Tokens:   []string{androidToken},
		Platform: PlatformAndroid,
		Message:  "Welcome",
	}

	PushToAndroid(req)
}

func TestPushToAndroidRightTokenForStringLog(t *testing.T) {
	PushConf, _ = config.LoadConf("")
	tenantId := "tenant_id1"
	tenant := PushConf.Tenants[tenantId]

	tenant.Android.Enabled = true
	tenant.Android.APIKey = os.Getenv("ANDROID_API_KEY")

	androidToken := os.Getenv("ANDROID_TEST_TOKEN")

	req := PushNotification{
		TenantId: tenantId,
		Tokens:   []string{androidToken},
		Platform: PlatformAndroid,
		Message:  "Welcome",
	}

	PushToAndroid(req)
}

func TestOverwriteAndroidAPIKey(t *testing.T) {
	PushConf, _ = config.LoadConf("")
	tenantId := "tenant_id1"
	tenant := PushConf.Tenants[tenantId]

	PushConf.Core.Sync = true
	tenant.Android.Enabled = true
	tenant.Android.APIKey = os.Getenv("ANDROID_API_KEY")

	androidToken := os.Getenv("ANDROID_TEST_TOKEN")

	req := PushNotification{
		TenantId: tenantId,
		Tokens:   []string{androidToken, "bbbbb"},
		Platform: PlatformAndroid,
		Message:  "Welcome",
		// overwrite android api key
		APIKey: "1234",

		log: &[]LogPushEntry{},
	}

	// FCM server error: 401 error: 401 Unauthorized (Wrong API Key)
	PushToAndroid(req)

	assert.Len(t, *req.log, 2)
}

func TestFCMMessage(t *testing.T) {
	var req PushNotification
	var err error

	// the message must specify at least one registration ID
	req = PushNotification{
		TenantId: "test",
		Message:  "Test",
		Tokens:   []string{},
	}

	err = CheckMessage(req)
	assert.Error(t, err)

	// the token must not be empty
	req = PushNotification{
		TenantId: "test",
		Message:  "Test",
		Tokens:   []string{""},
	}

	err = CheckMessage(req)
	assert.Error(t, err)

	// ignore check token length if send topic message
	req = PushNotification{
		TenantId: "test",
		Message:  "Test",
		Platform: PlatformAndroid,
		To:       "/topics/foo-bar",
	}

	err = CheckMessage(req)
	assert.NoError(t, err)

	// "condition": "'dogs' in topics || 'cats' in topics",
	req = PushNotification{
		TenantId:  "test",
		Message:   "Test",
		Platform:  PlatformAndroid,
		Condition: "'dogs' in topics || 'cats' in topics",
	}

	err = CheckMessage(req)
	assert.NoError(t, err)

	// the message may specify at most 1000 registration IDs
	req = PushNotification{
		TenantId: "test",
		Message:  "Test",
		Platform: PlatformAndroid,
		Tokens:   make([]string, 1001),
	}

	err = CheckMessage(req)
	assert.Error(t, err)

	// the message's TimeToLive field must be an integer
	// between 0 and 2419200 (4 weeks)
	timeToLive := uint(2419201)
	req = PushNotification{
		TenantId:   "test",
		Message:    "Test",
		Platform:   PlatformAndroid,
		Tokens:     []string{"XXXXXXXXX"},
		TimeToLive: &timeToLive,
	}

	err = CheckMessage(req)
	assert.Error(t, err)

	// Pass
	timeToLive = uint(86400)
	req = PushNotification{
		TenantId:   "test",
		Message:    "Test",
		Platform:   PlatformAndroid,
		Tokens:     []string{"XXXXXXXXX"},
		TimeToLive: &timeToLive,
	}

	err = CheckMessage(req)
	assert.NoError(t, err)
}

func TestCheckAndroidMessage(t *testing.T) {
	PushConf, _ = config.LoadConf("")
	tenantId := "tenant_id1"
	tenant := PushConf.Tenants[tenantId]

	tenant.Android.Enabled = true
	tenant.Android.APIKey = os.Getenv("ANDROID_API_KEY")

	timeToLive := uint(2419201)
	req := PushNotification{
		TenantId:   tenantId,
		Tokens:     []string{"aaaaaa", "bbbbb"},
		Platform:   PlatformAndroid,
		Message:    "Welcome",
		TimeToLive: &timeToLive,
	}

	PushToAndroid(req)
}

func TestAndroidNotificationStructure(t *testing.T) {
	test := "test"
	timeToLive := uint(100)
	req := PushNotification{
		Tokens:                []string{"a", "b"},
		Message:               "Welcome",
		To:                    test,
		Priority:              "high",
		CollapseKey:           "1",
		ContentAvailable:      true,
		DelayWhileIdle:        true,
		TimeToLive:            &timeToLive,
		RestrictedPackageName: test,
		DryRun:                true,
		Title:                 test,
		Sound:                 test,
		Data: D{
			"a": "1",
			"b": 2,
		},
		Notification: &fcm.Notification{
			Color: test,
			Tag:   test,
			Body:  "",
		},
	}

	notification := GetAndroidNotification(req)

	assert.Equal(t, test, notification.To)
	assert.Equal(t, "high", notification.Priority)
	assert.Equal(t, "1", notification.CollapseKey)
	assert.True(t, notification.ContentAvailable)
	assert.True(t, notification.DelayWhileIdle)
	assert.Equal(t, uint(100), *notification.TimeToLive)
	assert.Equal(t, test, notification.RestrictedPackageName)
	assert.True(t, notification.DryRun)
	assert.Equal(t, test, notification.Notification.Title)
	assert.Equal(t, test, notification.Notification.Sound)
	assert.Equal(t, test, notification.Notification.Color)
	assert.Equal(t, test, notification.Notification.Tag)
	assert.Equal(t, "Welcome", notification.Notification.Body)
	assert.Equal(t, "1", notification.Data["a"])
	assert.Equal(t, 2, notification.Data["b"])

	// test empty body
	req = PushNotification{
		Tokens: []string{"a", "b"},
		To:     test,
		Notification: &fcm.Notification{
			Body: "",
		},
	}
	notification = GetAndroidNotification(req)

	assert.Equal(t, test, notification.To)
	assert.Equal(t, "", notification.Notification.Body)
}
