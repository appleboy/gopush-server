package gorush

import (
	"context"
	"os"
	"testing"

	"github.com/appleboy/gorush/config"
	"github.com/stretchr/testify/assert"
)

func TestCorrectConf(t *testing.T) {
	PushConf, _ = config.LoadConf("")
	tenantId := "tenant_id1"

	tenant := PushConf.Tenants[tenantId]
	tenant.Android.Enabled = true
	tenant.Android.APIKey = "xxxxx"

	tenant.Ios.Enabled = true
	tenant.Ios.KeyPath = "../certificate/certificate-valid.pem"

	tenant.Huawei.Enabled = false

	err := CheckPushConf()

	assert.NoError(t, err)
}

func TestSendMultipleNotifications(t *testing.T) {
	ctx := context.Background()
	PushConf, _ = config.LoadConf("")

	tenantId := "tenant_id1"
	tenant := PushConf.Tenants[tenantId]

	tenant.Ios.Enabled = true
	tenant.Ios.KeyPath = "../certificate/certificate-valid.pem"

	err := InitAPNSClient(tenantId, *tenant)
	assert.Nil(t, err)

	tenant.Android.Enabled = true
	tenant.Android.APIKey = os.Getenv("ANDROID_API_KEY")

	androidToken := os.Getenv("ANDROID_TEST_TOKEN")

	req := RequestPush{
		Notifications: []PushNotification{
			// ios
			{
				Tokens:   []string{"11aa01229f15f0f0c52029d8cf8cd0aeaf2365fe4cebc4af26cd6d76b7919ef7"},
				Platform: PlatformIos,
				Message:  "Welcome",
			},
			// android
			{
				Tokens:   []string{androidToken, "bbbbb"},
				Platform: PlatformAndroid,
				Message:  "Welcome",
			},
		},
	}

	count, logs := queueNotification(ctx, req, tenantId)
	assert.Equal(t, 3, count)
	assert.Equal(t, 0, len(logs))
}

func TestDisabledAndroidNotifications(t *testing.T) {
	ctx := context.Background()
	PushConf, _ = config.LoadConf("")
	tenantId := "tenant_id1"

	tenant := PushConf.Tenants[tenantId]

	tenant.Ios.Enabled = true
	tenant.Ios.KeyPath = "../certificate/certificate-valid.pem"
	err := InitAPNSClient(tenantId, *tenant)
	assert.Nil(t, err)

	tenant.Android.Enabled = false
	tenant.Android.APIKey = os.Getenv("ANDROID_API_KEY")

	androidToken := os.Getenv("ANDROID_TEST_TOKEN")

	req := RequestPush{
		Notifications: []PushNotification{
			// ios
			{
				Tokens:   []string{"11aa01229f15f0f0c52029d8cf8cd0aeaf2365fe4cebc4af26cd6d76b7919ef7"},
				Platform: PlatformIos,
				Message:  "Welcome",
			},
			// android
			{
				Tokens:   []string{androidToken, "bbbbb"},
				Platform: PlatformAndroid,
				Message:  "Welcome",
			},
		},
	}

	count, logs := queueNotification(ctx, req, tenantId)
	assert.Equal(t, 1, count)
	assert.Equal(t, 0, len(logs))
}

func TestSyncModeForNotifications(t *testing.T) {
	ctx := context.Background()
	PushConf, _ = config.LoadConf("")
	tenantId := "tenant_id1"

	tenant := PushConf.Tenants[tenantId]

	tenant.Ios.Enabled = true
	tenant.Ios.KeyPath = "../certificate/certificate-valid.pem"
	err := InitAPNSClient(tenantId, *tenant)
	assert.Nil(t, err)

	tenant.Android.Enabled = true
	tenant.Android.APIKey = os.Getenv("ANDROID_API_KEY")

	// enable sync mode
	PushConf.Core.Sync = true

	androidToken := os.Getenv("ANDROID_TEST_TOKEN")

	req := RequestPush{
		Notifications: []PushNotification{
			// ios
			{
				Tokens:   []string{"11aa01229f15f0f0c52029d8cf8cd0aeaf2365fe4cebc4af26cd6d76b7919ef7"},
				Platform: PlatformIos,
				Message:  "Welcome",
			},
			// android
			{
				Tokens:   []string{androidToken, "bbbbb"},
				Platform: PlatformAndroid,
				Message:  "Welcome",
			},
		},
	}

	count, logs := queueNotification(ctx, req, tenantId)

	assert.Equal(t, 3, count)
	assert.Equal(t, 2, len(logs))
}

func TestSyncModeForTopicNotification(t *testing.T) {
	ctx := context.Background()
	PushConf, _ = config.LoadConf("")
	tenantId := "tenant_id1"

	tenant := PushConf.Tenants[tenantId]

	tenant.Android.Enabled = true
	tenant.Android.APIKey = os.Getenv("ANDROID_API_KEY")
	PushConf.Log.HideToken = false

	// enable sync mode
	PushConf.Core.Sync = true

	req := RequestPush{
		Notifications: []PushNotification{
			// android
			{
				// error:InvalidParameters
				// Check that the provided parameters have the right name and type.
				To:       "/topics/foo-bar@@@##",
				Platform: PlatformAndroid,
				Message:  "This is a Firebase Cloud Messaging Topic Message!",
			},
			// android
			{
				// success
				To:       "/topics/foo-bar",
				Platform: PlatformAndroid,
				Message:  "This is a Firebase Cloud Messaging Topic Message!",
			},
			// android
			{
				// success
				Condition: "'dogs' in topics || 'cats' in topics",
				Platform:  PlatformAndroid,
				Message:   "This is a Firebase Cloud Messaging Topic Message!",
			},
		},
	}

	count, logs := queueNotification(ctx, req, tenantId)
	assert.Equal(t, 2, count)
	assert.Equal(t, 1, len(logs))
}

func TestSyncModeForDeviceGroupNotification(t *testing.T) {
	ctx := context.Background()
	PushConf, _ = config.LoadConf("")
	tenantId := "tenant_id1"

	tenant := PushConf.Tenants[tenantId]

	tenant.Android.Enabled = true
	tenant.Android.APIKey = os.Getenv("ANDROID_API_KEY")
	PushConf.Log.HideToken = false

	// enable sync mode
	PushConf.Core.Sync = true

	req := RequestPush{
		Notifications: []PushNotification{
			// android
			{
				To:       "aUniqueKey",
				Platform: PlatformAndroid,
				Message:  "This is a Firebase Cloud Messaging Device Group Message!",
			},
		},
	}

	count, logs := queueNotification(ctx, req, tenantId)
	assert.Equal(t, 1, count)
	assert.Equal(t, 1, len(logs))
}

func TestSetProxyURL(t *testing.T) {
	err := SetProxy("87.236.233.92:8080")
	assert.Error(t, err)
	assert.Equal(t, "parse \"87.236.233.92:8080\": invalid URI for request", err.Error())

	err = SetProxy("a.html")
	assert.Error(t, err)

	err = SetProxy("http://87.236.233.92:8080")
	assert.NoError(t, err)
}
