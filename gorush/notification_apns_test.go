package gorush

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/appleboy/gorush/config"
	"github.com/buger/jsonparser"
	"github.com/sideshow/apns2"
	"github.com/stretchr/testify/assert"
)

const certificateValidP12 = `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`

const certificateValidPEM = `QmFnIEF0dHJpYnV0ZXMKICAgIGxvY2FsS2V5SUQ6IDhDIDFBIDlGIDAwIDY2IEJEIDI0IDQyIEI5IDVEIDFFIEVCIEZFIDVFIDhCIENBIDA0IDNEIDczIDgzIAogICAgZnJpZW5kbHlOYW1lOiBBUE5TLzIgUHJpdmF0ZSBLZXkKc3ViamVjdD0vQz1OWi9TVD1XZWxsaW5ndG9uL0w9V2VsbGluZ3Rvbi9PPUludGVybmV0IFdpZGdpdHMgUHR5IEx0ZC9PVT05WkVINjJLUlZWL0NOPUFQTlMvMiBEZXZlbG9wbWVudCBJT1MgUHVzaCBTZXJ2aWNlczogY29tLnNpZGVzaG93LkFwbnMyCmlzc3Vlcj0vQz1OWi9TVD1XZWxsaW5ndG9uL0w9V2VsbGluZ3Rvbi9PPUFQTlMvMiBJbmMuL09VPUFQTlMvMiBXb3JsZHdpZGUgRGV2ZWxvcGVyIFJlbGF0aW9ucy9DTj1BUE5TLzIgV29ybGR3aWRlIERldmVsb3BlciBSZWxhdGlvbnMgQ2VydGlmaWNhdGlvbiBBdXRob3JpdHkKLS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUQ2ekNDQXRNQ0FRSXdEUVlKS29aSWh2Y05BUUVMQlFBd2djTXhDekFKQmdOVkJBWVRBazVhTVJNd0VRWUQKVlFRSUV3cFhaV3hzYVc1bmRHOXVNUk13RVFZRFZRUUhFd3BYWld4c2FXNW5kRzl1TVJRd0VnWURWUVFLRXd0QgpVRTVUTHpJZ1NXNWpMakV0TUNzR0ExVUVDeE1rUVZCT1V5OHlJRmR2Y214a2QybGtaU0JFWlhabGJHOXdaWElnClVtVnNZWFJwYjI1ek1VVXdRd1lEVlFRREV6eEJVRTVUTHpJZ1YyOXliR1IzYVdSbElFUmxkbVZzYjNCbGNpQlMKWld4aGRHbHZibk1nUTJWeWRHbG1hV05oZEdsdmJpQkJkWFJvYjNKcGRIa3dIaGNOTVRZd01UQTRNRGd6TkRNdwpXaGNOTWpZd01UQTFNRGd6TkRNd1dqQ0JzakVMTUFrR0ExVUVCaE1DVGxveEV6QVJCZ05WQkFnVENsZGxiR3hwCmJtZDBiMjR4RXpBUkJnTlZCQWNUQ2xkbGJHeHBibWQwYjI0eElUQWZCZ05WQkFvVEdFbHVkR1Z5Ym1WMElGZHAKWkdkcGRITWdVSFI1SUV4MFpERVRNQkVHQTFVRUN4TUtPVnBGU0RZeVMxSldWakZCTUQ4R0ExVUVBeE00UVZCTwpVeTh5SUVSbGRtVnNiM0J0Wlc1MElFbFBVeUJRZFhOb0lGTmxjblpwWTJWek9pQmpiMjB1YzJsa1pYTm9iM2N1ClFYQnVjekl3Z2dFaU1BMEdDU3FHU0liM0RRRUJBUVVBQTRJQkR3QXdnZ0VLQW9JQkFRRFkwYzFUS0I1b1pQd1EKN3QxQ3dNSXJ2cUI2R0lVM3RQeTZSaGNrWlhUa09COFllQldKN1VLZkN6OEhHSEZWb21CUDBUNU9VYmVxUXpxVwpZSmJRelo4YTZaTXN6YkwwbE80WDkrKzNPaTUvVHRBd09VT0s4ck9GTjI1bTJLZnNheUhRWi80dldTdEsyRndtCjVhSmJHTGxwSC9iLzd6MUQ0dmhtTWdvQnVUMUl1eWhHaXlGeGxaOUV0VGxvRnZzcU0xRTVmWVpPU1pBQ3lYVGEKSzR2ZGdiUU1nVVZzSTcxNEZBZ0xUbEswVWVpUmttS20zcGRidGZWYnJ0aHpJK0lIWEtJdFVJeStGbjIwUFJNaApkU25henRTejd0Z0JXQ0l4MjJxdmNZb2dIV2lPZ1VZSU03NzJ6RTJ5OFVWT3I4RHNpUmxzT0hTQTdFSTRNSmNRCkcyRlVxMlovQWdNQkFBRXdEUVlKS29aSWh2Y05BUUVMQlFBRGdnRUJBR3lmeU8ySE1nY2RlQmN6M2J0NUJJTFgKZjdSQTIvVW1WSXdjS1IxcW90VHNGK1BuQm1jSUxleU9RZ0RlOXRHVTVjUmM3OWtEdDNKUm1NWVJPRklNZ0ZSZgpXZjIydU9LdGhvN0dRUWFLdkcrYmtnTVZkWUZSbEJIbkYrS2VxS0g4MXFiOXArQ1Q0SXcwR2VoSUwxRGlqRkxSClZJQUlCWXB6NG9CUENJRTFJU1ZUK0ZnYWYzSkFoNTlrYlBiTnc5QUlEeGFCdFA4RXV6U1ROd2ZieG9HYkNvYlMKV2kxVThJc0N3UUZ0OHRNMW00WlhEMUNjWklyR2RyeWVBaFZrdktJSlJpVTVRWVdJMm5xWk4rSnFRdWNtOWFkMAptWU81bUprSW9iVWE0K1pKaENQS0VkbWdwRmJSR2swd1Z1YURNOUN2NlAyc3JzWUFqYU80eTNWUDBHdk5LUkk9Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0KQmFnIEF0dHJpYnV0ZXMKICAgIGxvY2FsS2V5SUQ6IDhDIDFBIDlGIDAwIDY2IEJEIDI0IDQyIEI5IDVEIDFFIEVCIEZFIDVFIDhCIENBIDA0IDNEIDczIDgzIAogICAgZnJpZW5kbHlOYW1lOiBBUE5TLzIgUHJpdmF0ZSBLZXkKS2V5IEF0dHJpYnV0ZXM6IDxObyBBdHRyaWJ1dGVzPgotLS0tLUJFR0lOIFJTQSBQUklWQVRFIEtFWS0tLS0tCk1JSUVvd0lCQUFLQ0FRRUEyTkhOVXlnZWFHVDhFTzdkUXNEQ0s3NmdlaGlGTjdUOHVrWVhKR1YwNURnZkdIZ1YKaWUxQ253cy9CeGh4VmFKZ1Q5RStUbEczcWtNNmxtQ1cwTTJmR3VtVExNMnk5SlR1Ri9mdnR6b3VmMDdRTURsRAppdkt6aFRkdVp0aW43R3NoMEdmK0wxa3JTdGhjSnVXaVd4aTVhUi8yLys4OVErTDRaaklLQWJrOVNMc29Sb3NoCmNaV2ZSTFU1YUJiN0tqTlJPWDJHVGttUUFzbDAyaXVMM1lHMERJRkZiQ085ZUJRSUMwNVN0Rkhva1pKaXB0NlgKVzdYMVc2N1ljeVBpQjF5aUxWQ012aFo5dEQwVElYVXAyczdVcys3WUFWZ2lNZHRxcjNHS0lCMW9qb0ZHQ0RPKwo5c3hOc3ZGRlRxL0E3SWtaYkRoMGdPeENPRENYRUJ0aFZLdG1md0lEQVFBQkFvSUJBUUNXOFpDSStPQWFlMXRFCmlwWjlGMmJXUDNMSExYVG84RllWZENBK1ZXZUlUazNQb2lJVWtKbVYwYVdDVWhEc3RndG81ZG9EZWo1c0NUdXIKWHZqL3luYWVyTWVxSkZZV2tld2p3WmNnTHlBWnZ3dU8xdjdmcDlFMHgvOVRHRGZuampuUE5lYXVuZHhXMGNOdAp6T1kzbDBIVkhzeTlKcGUzUURjQUpvdnk0VHY1K2hGWTRrRHhVQkdzeWp2aFNjVmdLZzV0TGtKY2xtM3NPdS9MCkd5THFwd05JM09KQWRNSXVWRDROMkJaMWFPRWFwNm1wMnk4SWUwL1I0WVdjYVo1QTRQdzd4VVBsNlNYYzl1dWEKLzc4UVRFUnRQQzZlanlDQmlFMDVhOG0zUTNpdWQzWHRubHl3czJLd2hnQkFmRTZNNHpSL2YzT1FCN1pJWE1oeQpacG1aWnc1eEFvR0JBUFluODRJcmxJUWV0V1FmdlBkTTdLemdoNlVESEN1Z25sQ0RnaHdZcFJKR2k4aE1mdVpWCnhOSXJZQUp6TFlEUTAxbEZKUkpnV1hUY2JxejlOQnoxbmhnK2NOT3oxL0tZKzM4ZXVkZWU2RE5ZbXp0UDdqRFAKMmpuYVMrZHRqQzhoQVhPYm5GcUcrTmlsTURMTHU2YVJtckphSW1ialNyZnlMaUU2bXZKN3U4MW5Bb0dCQU9GOQpnOTN3WjBtTDFyazJzNVd3SEdUTlUvSGFPdG1XUzR6N2tBN2Y0UWFSdWIrTXdwcFptbURaUEhwaVpYN0JQY1p6CmlPUFFoK3huN0lxUkdvUVdCTHlrQlZ0OHpaRm9MWkpvQ1IzbjYzbGV4NUE0cC8wUHAxZ0ZaclIreFg4UFlWb3MKM3llZWlXeVBLc1hYTmMwczVRd0haY1g2V2I4RUhUaFRYR0NCZXRjcEFvR0FNZVFKQzlJUGFQUGNhZTJ3M0NMQQpPWTNNa0ZwZ0JFdXFxc0RzeHdzTHNmZVFiMGxwMHYrQlErTzhzdUpyVDVlRHJxMUFCVWgzK1NLUVlBbDEzWVMrCnhVVXFrdzM1YjljbjZpenRGOUhDV0YzV0lLQmpzNHI5UFFxTXBkeGpORTRwUUNoQytXb3YxNkVyY3JBdVdXVmIKaUZpU2JtNFUvOUZiSGlzRnFxMy9jM01DZ1lCK3Z6U3VQZ0Z3MzcrMG9FRFZ0UVpneXVHU29wNU56Q052ZmIvOQovRzNhYVhORmJuTzhtdjBoenpvbGVNV2dPRExuSis0Y1VBejNIM3RnY0N1OWJ6citaaHYwenZRbDlhOFlDbzZGClZ1V1BkVzByYmcxUE84dE91TXFBVG5ubzc5WkMvOUgzelM5bDdCdVkxVjJTbE5leXFUM1Z5T0ZGYzZTUkVwcHMKVEp1bDhRS0JnQXhuUUI4TUE3elBVTHUxY2x5YUpMZHRFZFJQa0tXTjdsS1lwdGMwZS9WSGZTc0t4c2VXa2ZxaQp6Z1haNTFrUVRyVDZaYjZIWVJmd0MxbU1YSFdSS1J5WWpBbkN4VmltNllRZCtLVlQ0OWlSRERBaUlGb01HQTRpCnZ2Y0lsbmVxT1paUERJb0tKNjBJak8vRFpIV2t3NW1MamFJclQrcVEzWEFHZEpBMTNoY20KLS0tLS1FTkQgUlNBIFBSSVZBVEUgS0VZLS0tLS0K`

const authkeyInvalidP8 = `TUlHSEFnRUFNQk1HQnlxR1NNNDlBZ0VHQ0NxR1NNNDlBd0VIQkcwd2F3SUJBUVFnRWJWemZQblpQeGZBeXhxRQpaVjA1bGFBb0pBbCsvNlh0Mk80bU9CNjExc09oUkFOQ0FBU2dGVEtqd0pBQVU5NWcrKy92ektXSGt6QVZtTk1JCnRCNXZUalpPT0l3bkViNzBNc1daRkl5VUZEMVA5R3dzdHo0K2FrSFg3dkk4Qkg2aEhtQm1mWlpaCg==`

const authkeyValidP8 = `LS0tLS1CRUdJTiBQUklWQVRFIEtFWS0tLS0tCk1JR0hBZ0VBTUJNR0J5cUdTTTQ5QWdFR0NDcUdTTTQ5QXdFSEJHMHdhd0lCQVFRZ0ViVnpmUG5aUHhmQXl4cUUKWlYwNWxhQW9KQWwrLzZYdDJPNG1PQjYxMXNPaFJBTkNBQVNnRlRLandKQUFVOTVnKysvdnpLV0hrekFWbU5NSQp0QjV2VGpaT09Jd25FYjcwTXNXWkZJeVVGRDFQOUd3c3R6NCtha0hYN3ZJOEJINmhIbUJtZmVRbAotLS0tLUVORCBQUklWQVRFIEtFWS0tLS0tCg==`

func init() {
	ApnsClients = make(map[string]*apns2.Client)
	MaxConcurrentIOSPushes = make(map[string]chan struct{})
}

func TestDisabledConf(t *testing.T) {
	PushConf, _ = config.LoadConf("")
	tenantId := "tenant_id1"
	tenant := PushConf.Tenants[tenantId]

	tenant.Android.Enabled = false
	tenant.Huawei.Enabled = false
	tenant.Ios.Enabled = false

	err := CheckPushConf()

	assert.Error(t, err)
	assert.Equal(t, "please enable iOS, Android or Huawei config in yml config for tenant "+tenantId, err.Error())
}

func TestMissingIOSCertificate(t *testing.T) {
	PushConf, _ = config.LoadConf("")
	tenantId := "tenant_id1"
	tenant := PushConf.Tenants[tenantId]

	tenant.Ios.Enabled = true
	tenant.Ios.KeyPath = ""
	tenant.Ios.KeyBase64 = ""

	err := CheckPushConf()

	assert.Error(t, err)
	assert.Equal(t, "missing iOS certificate key for tenant "+tenantId, err.Error())

	tenant.Ios.KeyPath = "test.pem"
	err = CheckPushConf()

	assert.Error(t, err)
	assert.Equal(t, "certificate file does not exist for tenant "+tenantId, err.Error())
}

func TestIOSNotificationStructure(t *testing.T) {
	var dat map[string]interface{}
	unix := time.Now().Unix()

	test := "test"
	expectBadge := 0
	message := "Welcome notification Server"
	expiration := int64(time.Now().Unix())
	req := PushNotification{
		ApnsID:     test,
		Topic:      test,
		Expiration: &expiration,
		Priority:   "normal",
		Message:    message,
		Badge:      &expectBadge,
		Sound: Sound{
			Critical: 1,
			Name:     test,
			Volume:   1.0,
		},
		ContentAvailable: true,
		Data: D{
			"key1": "test",
			"key2": 2,
		},
		Category: test,
		URLArgs:  []string{"a", "b"},
	}

	notification := GetIOSNotification(req)

	dump, _ := json.Marshal(notification.Payload)
	data := []byte(string(dump))

	if err := json.Unmarshal(data, &dat); err != nil {
		panic(err)
	}

	alert, _ := jsonparser.GetString(data, "aps", "alert")
	badge, _ := jsonparser.GetInt(data, "aps", "badge")
	soundName, _ := jsonparser.GetString(data, "aps", "sound", "name")
	soundCritical, _ := jsonparser.GetInt(data, "aps", "sound", "critical")
	soundVolume, _ := jsonparser.GetFloat(data, "aps", "sound", "volume")
	contentAvailable, _ := jsonparser.GetInt(data, "aps", "content-available")
	category, _ := jsonparser.GetString(data, "aps", "category")
	key1 := dat["key1"].(interface{})
	key2 := dat["key2"].(interface{})
	aps := dat["aps"].(map[string]interface{})
	urlArgs := aps["url-args"].([]interface{})

	assert.Equal(t, test, notification.ApnsID)
	assert.Equal(t, test, notification.Topic)
	assert.Equal(t, unix, notification.Expiration.Unix())
	assert.Equal(t, ApnsPriorityLow, notification.Priority)
	assert.Equal(t, message, alert)
	assert.Equal(t, expectBadge, int(badge))
	assert.Equal(t, expectBadge, *req.Badge)
	assert.Equal(t, test, soundName)
	assert.Equal(t, 1.0, soundVolume)
	assert.Equal(t, int64(1), soundCritical)
	assert.Equal(t, 1, int(contentAvailable))
	assert.Equal(t, "test", key1)
	assert.Equal(t, 2, int(key2.(float64)))
	assert.Equal(t, test, category)
	assert.Contains(t, urlArgs, "a")
	assert.Contains(t, urlArgs, "b")
}

func TestIOSSoundAndVolume(t *testing.T) {
	var dat map[string]interface{}

	test := "test"
	message := "Welcome notification Server"
	req := PushNotification{
		ApnsID:   test,
		Topic:    test,
		Priority: "normal",
		Message:  message,
		Sound: Sound{
			Critical: 3,
			Name:     test,
			Volume:   4.5,
		},
	}

	notification := GetIOSNotification(req)

	dump, _ := json.Marshal(notification.Payload)
	data := []byte(string(dump))

	if err := json.Unmarshal(data, &dat); err != nil {
		panic(err)
	}

	alert, _ := jsonparser.GetString(data, "aps", "alert")
	soundName, _ := jsonparser.GetString(data, "aps", "sound", "name")
	soundCritical, _ := jsonparser.GetInt(data, "aps", "sound", "critical")
	soundVolume, _ := jsonparser.GetFloat(data, "aps", "sound", "volume")

	assert.Equal(t, test, notification.ApnsID)
	assert.Equal(t, test, notification.Topic)
	assert.Equal(t, ApnsPriorityLow, notification.Priority)
	assert.Equal(t, message, alert)
	assert.Equal(t, test, soundName)
	assert.Equal(t, 4.5, soundVolume)
	assert.Equal(t, int64(3), soundCritical)

	req.SoundName = "foobar"
	req.SoundVolume = 5.5
	notification = GetIOSNotification(req)
	dump, _ = json.Marshal(notification.Payload)
	data = []byte(string(dump))

	if err := json.Unmarshal(data, &dat); err != nil {
		panic(err)
	}

	soundName, _ = jsonparser.GetString(data, "aps", "sound", "name")
	soundVolume, _ = jsonparser.GetFloat(data, "aps", "sound", "volume")
	soundCritical, _ = jsonparser.GetInt(data, "aps", "sound", "critical")
	assert.Equal(t, 5.5, soundVolume)
	assert.Equal(t, int64(1), soundCritical)
	assert.Equal(t, "foobar", soundName)

	req = PushNotification{
		ApnsID:   test,
		Topic:    test,
		Priority: "normal",
		Message:  message,
		Sound: map[string]interface{}{
			"critical": 3,
			"name":     "test",
			"volume":   4.5,
		},
	}

	notification = GetIOSNotification(req)
	dump, _ = json.Marshal(notification.Payload)
	data = []byte(string(dump))

	if err := json.Unmarshal(data, &dat); err != nil {
		panic(err)
	}

	soundName, _ = jsonparser.GetString(data, "aps", "sound", "name")
	soundVolume, _ = jsonparser.GetFloat(data, "aps", "sound", "volume")
	soundCritical, _ = jsonparser.GetInt(data, "aps", "sound", "critical")
	assert.Equal(t, 4.5, soundVolume)
	assert.Equal(t, int64(3), soundCritical)
	assert.Equal(t, "test", soundName)

	req = PushNotification{
		ApnsID:   test,
		Topic:    test,
		Priority: "normal",
		Message:  message,
		Sound:    "default",
	}

	notification = GetIOSNotification(req)
	dump, _ = json.Marshal(notification.Payload)
	data = []byte(string(dump))

	if err := json.Unmarshal(data, &dat); err != nil {
		panic(err)
	}

	soundName, _ = jsonparser.GetString(data, "aps", "sound")
	assert.Equal(t, "default", soundName)
}

func TestIOSSummaryArg(t *testing.T) {
	var dat map[string]interface{}

	test := "test"
	message := "Welcome notification Server"
	req := PushNotification{
		ApnsID:   test,
		Topic:    test,
		Priority: "normal",
		Message:  message,
		Alert: Alert{
			SummaryArg:      "test",
			SummaryArgCount: 3,
		},
	}

	notification := GetIOSNotification(req)

	dump, _ := json.Marshal(notification.Payload)
	data := []byte(string(dump))

	if err := json.Unmarshal(data, &dat); err != nil {
		panic(err)
	}

	assert.Equal(t, test, notification.ApnsID)
	assert.Equal(t, test, notification.Topic)
	assert.Equal(t, ApnsPriorityLow, notification.Priority)
	assert.Equal(t, "test", dat["aps"].(map[string]interface{})["alert"].(map[string]interface{})["summary-arg"])
	assert.Equal(t, float64(3), dat["aps"].(map[string]interface{})["alert"].(map[string]interface{})["summary-arg-count"])
}

// Silent Notification which payload’s aps dictionary must not contain the alert, sound, or badge keys.
// ref: https://goo.gl/m9xyqG
func TestSendZeroValueForBadgeKey(t *testing.T) {
	var dat map[string]interface{}

	test := "test"
	message := "Welcome notification Server"
	req := PushNotification{
		ApnsID:           test,
		Topic:            test,
		Priority:         "normal",
		Message:          message,
		Sound:            test,
		ContentAvailable: true,
		MutableContent:   true,
		ThreadID:         test,
	}

	notification := GetIOSNotification(req)

	dump, _ := json.Marshal(notification.Payload)
	data := []byte(string(dump))

	if err := json.Unmarshal(data, &dat); err != nil {
		log.Println(err)
		panic(err)
	}

	alert, _ := jsonparser.GetString(data, "aps", "alert")
	badge, _ := jsonparser.GetInt(data, "aps", "badge")
	sound, _ := jsonparser.GetString(data, "aps", "sound")
	threadID, _ := jsonparser.GetString(data, "aps", "thread-id")
	contentAvailable, _ := jsonparser.GetInt(data, "aps", "content-available")
	mutableContent, _ := jsonparser.GetInt(data, "aps", "mutable-content")

	if req.Badge != nil {
		t.Errorf("req.Badge must be nil")
	}

	assert.Equal(t, test, notification.ApnsID)
	assert.Equal(t, test, notification.Topic)
	assert.Equal(t, ApnsPriorityLow, notification.Priority)
	assert.Equal(t, message, alert)
	assert.Equal(t, 0, int(badge))
	assert.Equal(t, test, sound)
	assert.Equal(t, test, threadID)
	assert.Equal(t, 1, int(contentAvailable))
	assert.Equal(t, 1, int(mutableContent))

	// Add Bage
	expectBadge := 10
	req.Badge = &expectBadge

	notification = GetIOSNotification(req)

	dump, _ = json.Marshal(notification.Payload)
	data = []byte(string(dump))

	if err := json.Unmarshal(data, &dat); err != nil {
		log.Println(err)
		panic(err)
	}

	if req.Badge == nil {
		t.Errorf("req.Badge must be equal %d", *req.Badge)
	}

	badge, _ = jsonparser.GetInt(data, "aps", "badge")
	assert.Equal(t, expectBadge, *req.Badge)
	assert.Equal(t, expectBadge, int(badge))
}

// Silent Notification:
// The payload’s aps dictionary must include the content-available key with a value of 1.
// The payload’s aps dictionary must not contain the alert, sound, or badge keys.
// ref: https://goo.gl/m9xyqG
func TestCheckSilentNotification(t *testing.T) {
	var dat map[string]interface{}

	test := "test"
	req := PushNotification{
		ApnsID:           test,
		Topic:            test,
		CollapseID:       test,
		Priority:         "normal",
		ContentAvailable: true,
	}

	notification := GetIOSNotification(req)

	dump, _ := json.Marshal(notification.Payload)
	data := []byte(string(dump))

	if err := json.Unmarshal(data, &dat); err != nil {
		log.Println(err)
		panic(err)
	}

	assert.Equal(t, test, notification.CollapseID)
	assert.Equal(t, test, notification.ApnsID)
	assert.Equal(t, test, notification.Topic)
	assert.Nil(t, dat["aps"].(map[string]interface{})["alert"])
	assert.Nil(t, dat["aps"].(map[string]interface{})["sound"])
	assert.Nil(t, dat["aps"].(map[string]interface{})["badge"])
}

//URL: https://goo.gl/5xFo3C
//Example 2
//{
//    "aps" : {
//        "alert" : {
//            "title" : "Game Request",
//            "body" : "Bob wants to play poker",
//            "action-loc-key" : "PLAY"
//        },
//        "badge" : 5
//    },
//    "acme1" : "bar",
//    "acme2" : [ "bang",  "whiz" ]
//}
func TestAlertStringExample2ForIos(t *testing.T) {
	var dat map[string]interface{}

	test := "test"
	title := "Game Request"
	body := "Bob wants to play poker"
	actionLocKey := "PLAY"
	req := PushNotification{
		ApnsID:   test,
		Topic:    test,
		Priority: "normal",
		Alert: Alert{
			Title:        title,
			Body:         body,
			ActionLocKey: actionLocKey,
		},
	}

	notification := GetIOSNotification(req)

	dump, _ := json.Marshal(notification.Payload)
	data := []byte(string(dump))

	if err := json.Unmarshal(data, &dat); err != nil {
		log.Println(err)
		panic(err)
	}

	assert.Equal(t, title, dat["aps"].(map[string]interface{})["alert"].(map[string]interface{})["title"])
	assert.Equal(t, body, dat["aps"].(map[string]interface{})["alert"].(map[string]interface{})["body"])
	assert.Equal(t, actionLocKey, dat["aps"].(map[string]interface{})["alert"].(map[string]interface{})["action-loc-key"])
}

// URL: https://goo.gl/5xFo3C
// Example 3
// {
//     "aps" : {
//         "alert" : "You got your emails.",
//         "badge" : 9,
//         "sound" : "bingbong.aiff"
//     },
//     "acme1" : "bar",
//     "acme2" : 42
// }
func TestAlertStringExample3ForIos(t *testing.T) {
	var dat map[string]interface{}

	test := "test"
	badge := 9
	sound := "bingbong.aiff"
	req := PushNotification{
		ApnsID:           test,
		Topic:            test,
		Priority:         "normal",
		ContentAvailable: true,
		Message:          test,
		Badge:            &badge,
		Sound:            sound,
	}

	notification := GetIOSNotification(req)

	dump, _ := json.Marshal(notification.Payload)
	data := []byte(string(dump))

	if err := json.Unmarshal(data, &dat); err != nil {
		log.Println(err)
		panic(err)
	}

	assert.Equal(t, sound, dat["aps"].(map[string]interface{})["sound"])
	assert.Equal(t, float64(badge), dat["aps"].(map[string]interface{})["badge"].(float64))
	assert.Equal(t, test, dat["aps"].(map[string]interface{})["alert"])
}

func TestMessageAndTitle(t *testing.T) {
	var dat map[string]interface{}

	test := "test"
	message := "Welcome notification Server"
	title := "Welcome notification Server title"
	req := PushNotification{
		ApnsID:           test,
		Topic:            test,
		Priority:         "normal",
		Message:          message,
		Title:            title,
		ContentAvailable: true,
	}

	notification := GetIOSNotification(req)

	dump, _ := json.Marshal(notification.Payload)
	data := []byte(string(dump))

	if err := json.Unmarshal(data, &dat); err != nil {
		log.Println(err)
		panic(err)
	}

	alert, _ := jsonparser.GetString(data, "aps", "alert")
	alertBody, _ := jsonparser.GetString(data, "aps", "alert", "body")
	alertTitle, _ := jsonparser.GetString(data, "aps", "alert", "title")

	assert.Equal(t, test, notification.ApnsID)
	assert.Equal(t, ApnsPriorityLow, notification.Priority)
	assert.Equal(t, message, alertBody)
	assert.Equal(t, title, alertTitle)
	assert.NotEqual(t, message, alert)

	// Add alert body
	messageOverride := "Welcome notification Server overridden"
	req.Alert.Body = messageOverride

	notification = GetIOSNotification(req)

	dump, _ = json.Marshal(notification.Payload)
	data = []byte(string(dump))

	if err := json.Unmarshal(data, &dat); err != nil {
		log.Println(err)
		panic(err)
	}

	alertBodyOverridden, _ := jsonparser.GetString(data, "aps", "alert", "body")
	alertTitle, _ = jsonparser.GetString(data, "aps", "alert", "title")
	assert.Equal(t, messageOverride, alertBodyOverridden)
	assert.NotEqual(t, message, alertBodyOverridden)
	assert.Equal(t, title, alertTitle)
}

func TestIOSAlertNotificationStructure(t *testing.T) {
	var dat map[string]interface{}

	test := "test"
	req := PushNotification{
		Message: "Welcome",
		Title:   test,
		Alert: Alert{
			Action:       test,
			ActionLocKey: test,
			Body:         test,
			LaunchImage:  test,
			LocArgs:      []string{"a", "b"},
			LocKey:       test,
			Subtitle:     test,
			TitleLocArgs: []string{"a", "b"},
			TitleLocKey:  test,
		},
	}

	notification := GetIOSNotification(req)

	dump, _ := json.Marshal(notification.Payload)
	data := []byte(string(dump))

	if err := json.Unmarshal(data, &dat); err != nil {
		log.Println(err)
		panic(err)
	}

	action, _ := jsonparser.GetString(data, "aps", "alert", "action")
	actionLocKey, _ := jsonparser.GetString(data, "aps", "alert", "action-loc-key")
	body, _ := jsonparser.GetString(data, "aps", "alert", "body")
	launchImage, _ := jsonparser.GetString(data, "aps", "alert", "launch-image")
	locKey, _ := jsonparser.GetString(data, "aps", "alert", "loc-key")
	title, _ := jsonparser.GetString(data, "aps", "alert", "title")
	subtitle, _ := jsonparser.GetString(data, "aps", "alert", "subtitle")
	titleLocKey, _ := jsonparser.GetString(data, "aps", "alert", "title-loc-key")
	aps := dat["aps"].(map[string]interface{})
	alert := aps["alert"].(map[string]interface{})
	titleLocArgs := alert["title-loc-args"].([]interface{})
	locArgs := alert["loc-args"].([]interface{})

	assert.Equal(t, test, action)
	assert.Equal(t, test, actionLocKey)
	assert.Equal(t, test, body)
	assert.Equal(t, test, launchImage)
	assert.Equal(t, test, locKey)
	assert.Equal(t, test, title)
	assert.Equal(t, test, subtitle)
	assert.Equal(t, test, titleLocKey)
	assert.Contains(t, titleLocArgs, "a")
	assert.Contains(t, titleLocArgs, "b")
	assert.Contains(t, locArgs, "a")
	assert.Contains(t, locArgs, "b")
}

func TestDisabledIosNotifications(t *testing.T) {
	ctx := context.Background()
	PushConf, _ = config.LoadConf("")
	tenantId := "tenant_id1"
	tenant := PushConf.Tenants[tenantId]

	tenant.Ios.Enabled = false
	tenant.Ios.KeyPath = "../certificate/certificate-valid.pem"
	err := InitAPNSClient("tenant_id1", *tenant)
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
				Tokens:   []string{androidToken, androidToken + "_"},
				Platform: PlatformAndroid,
				Message:  "Welcome",
			},
		},
	}

	count, logs := queueNotification(ctx, req, tenantId)
	assert.Equal(t, 2, count)
	assert.Equal(t, 0, len(logs))
}

func TestWrongIosCertificateExt(t *testing.T) {
	PushConf, _ = config.LoadConf("")
	tenantId := "tenant_id1"
	tenant := PushConf.Tenants[tenantId]

	tenant.Ios.Enabled = true
	tenant.Ios.KeyPath = "test"
	err := InitAPNSClient("tenant_id1", *tenant)

	assert.Error(t, err)
	assert.Equal(t, "wrong certificate key extension", err.Error())

	tenant.Ios.KeyPath = ""
	tenant.Ios.KeyBase64 = "abcd"
	tenant.Ios.KeyType = "abcd"
	err = InitAPNSClient("tenant_id1", *tenant)

	assert.Error(t, err)
	assert.Equal(t, "wrong certificate key type", err.Error())
}

func TestAPNSClientDevHost(t *testing.T) {
	PushConf, _ = config.LoadConf("")
	tenantId := "tenant_id1"
	tenant := PushConf.Tenants[tenantId]

	tenant.Ios.Enabled = true
	tenant.Ios.KeyPath = "../certificate/certificate-valid.p12"
	err := InitAPNSClient(tenantId, *tenant)
	assert.Nil(t, err)
	assert.Equal(t, apns2.HostDevelopment, ApnsClients[tenantId].Host)

	tenant.Ios.KeyPath = ""
	tenant.Ios.KeyBase64 = certificateValidP12
	tenant.Ios.KeyType = "p12"
	err = InitAPNSClient("tenant_id1", *tenant)
	assert.Nil(t, err)
	assert.Equal(t, apns2.HostDevelopment, ApnsClients[tenantId].Host)
}

func TestAPNSClientProdHost(t *testing.T) {
	PushConf, _ = config.LoadConf("")
	tenantId := "tenant_id1"
	tenant := PushConf.Tenants[tenantId]

	tenant.Ios.Enabled = true
	tenant.Ios.Production = true
	tenant.Ios.KeyPath = "../certificate/certificate-valid.pem"
	err := InitAPNSClient(tenantId, *tenant)
	assert.Nil(t, err)
	assert.Equal(t, apns2.HostProduction, ApnsClients[tenantId].Host)

	tenant.Ios.KeyPath = ""
	tenant.Ios.KeyBase64 = certificateValidPEM
	tenant.Ios.KeyType = "pem"
	err = InitAPNSClient(tenantId, *tenant)
	assert.Nil(t, err)
	assert.Equal(t, apns2.HostProduction, ApnsClients[tenantId].Host)
}

func TestAPNSClientInvaildToken(t *testing.T) {
	PushConf, _ = config.LoadConf("")
	tenantId := "tenant_id1"
	tenant := PushConf.Tenants[tenantId]

	tenant.Ios.Enabled = true
	tenant.Ios.KeyPath = "../certificate/authkey-invalid.p8"
	err := InitAPNSClient(tenantId, *tenant)
	assert.Error(t, err)

	tenant.Ios.KeyPath = ""
	tenant.Ios.KeyBase64 = authkeyInvalidP8
	tenant.Ios.KeyType = "p8"
	err = InitAPNSClient(tenantId, *tenant)
	assert.Error(t, err)

	// empty key-id or team-id
	tenant.Ios.Enabled = true
	tenant.Ios.KeyPath = "../certificate/authkey-valid.p8"
	err = InitAPNSClient(tenantId, *tenant)
	assert.Error(t, err)

	tenant.Ios.KeyID = "key-id"
	tenant.Ios.TeamID = ""
	err = InitAPNSClient(tenantId, *tenant)
	assert.Error(t, err)

	tenant.Ios.KeyID = ""
	tenant.Ios.TeamID = "team-id"
	err = InitAPNSClient(tenantId, *tenant)
	assert.Error(t, err)
}

func TestAPNSClientVaildToken(t *testing.T) {
	PushConf, _ = config.LoadConf("")
	tenantId := "tenant_id1"
	tenant := PushConf.Tenants[tenantId]

	tenant.Ios.Enabled = true
	tenant.Ios.KeyPath = "../certificate/authkey-valid.p8"
	tenant.Ios.KeyID = "key-id"
	tenant.Ios.TeamID = "team-id"
	err := InitAPNSClient(tenantId, *tenant)
	assert.NoError(t, err)
	assert.Equal(t, apns2.HostDevelopment, ApnsClients[tenantId].Host)

	tenant.Ios.Production = true
	err = InitAPNSClient(tenantId, *tenant)
	assert.NoError(t, err)
	assert.Equal(t, apns2.HostProduction, ApnsClients[tenantId].Host)

	// test base64
	tenant.Ios.Production = false
	tenant.Ios.KeyPath = ""
	tenant.Ios.KeyBase64 = authkeyValidP8
	tenant.Ios.KeyType = "p8"
	err = InitAPNSClient(tenantId, *tenant)
	assert.NoError(t, err)
	assert.Equal(t, apns2.HostDevelopment, ApnsClients[tenantId].Host)

	tenant.Ios.Production = true
	err = InitAPNSClient(tenantId, *tenant)
	assert.NoError(t, err)
	assert.Equal(t, apns2.HostProduction, ApnsClients[tenantId].Host)
}

func TestAPNSClientUseProxy(t *testing.T) {
	PushConf, _ = config.LoadConf("")
	tenantId := "tenant_id1"
	tenant := PushConf.Tenants[tenantId]

	tenant.Ios.Enabled = true
	tenant.Ios.KeyPath = "../certificate/certificate-valid.p12"
	PushConf.Core.HTTPProxy = "http://127.0.0.1:8080"
	_ = SetProxy(PushConf.Core.HTTPProxy)
	err := InitAPNSClient(tenantId, *tenant)
	assert.Nil(t, err)
	assert.Equal(t, apns2.HostDevelopment, ApnsClients[tenantId].Host)

	req, _ := http.NewRequest("GET", apns2.HostDevelopment, nil)
	actualProxyURL, err := ApnsClients[tenantId].HTTPClient.Transport.(*http.Transport).Proxy(req)
	assert.Nil(t, err)

	expectedProxyURL, _ := url.ParseRequestURI(PushConf.Core.HTTPProxy)
	assert.Equal(t, expectedProxyURL, actualProxyURL)

	tenant.Ios.KeyPath = "../certificate/authkey-valid.p8"
	tenant.Ios.TeamID = "example.team"
	tenant.Ios.KeyID = "example.key"
	err = InitAPNSClient(tenantId, *tenant)
	assert.Nil(t, err)
	assert.Equal(t, apns2.HostDevelopment, ApnsClients[tenantId].Host)
	assert.NotNil(t, ApnsClients[tenantId].Token)

	req, _ = http.NewRequest("GET", apns2.HostDevelopment, nil)
	actualProxyURL, err = ApnsClients[tenantId].HTTPClient.Transport.(*http.Transport).Proxy(req)
	assert.Nil(t, err)

	expectedProxyURL, _ = url.ParseRequestURI(PushConf.Core.HTTPProxy)
	assert.Equal(t, expectedProxyURL, actualProxyURL)

	http.DefaultTransport.(*http.Transport).Proxy = nil
}

func TestPushToIOS(t *testing.T) {
	PushConf, _ = config.LoadConf("")
	tenantId := "tenant_id1"
	tenant := PushConf.Tenants[tenantId]

	MaxConcurrentIOSPushes[tenantId] = make(chan struct{}, tenant.Ios.MaxConcurrentPushes)

	tenant.Ios.Enabled = true
	tenant.Ios.KeyPath = "../certificate/certificate-valid.pem"
	err := InitAPNSClient(tenantId, *tenant)
	assert.Nil(t, err)
	err = InitAppStatus()
	assert.Nil(t, err)

	req := PushNotification{
		TenantId: tenantId,
		Tokens:   []string{"11aa01229f15f0f0c52029d8cf8cd0aeaf2365fe4cebc4af26cd6d76b7919ef7", "11aa01229f15f0f0c52029d8cf8cd0aeaf2365fe4cebc4af26cd6d76b7919ef1"},
		Platform: 1,
		Message:  "Welcome",
	}

	// send fail
	PushToIOS(req)
}

func TestApnsHostFromRequest(t *testing.T) {
	PushConf, _ = config.LoadConf("")
	tenantId := "tenant_id1"
	tenant := PushConf.Tenants[tenantId]

	tenant.Ios.Enabled = true
	tenant.Ios.KeyPath = "../certificate/certificate-valid.pem"
	err := InitAPNSClient(tenantId, *tenant)
	assert.Nil(t, err)
	err = InitAppStatus()
	assert.Nil(t, err)

	req := PushNotification{
		TenantId:   tenantId,
		Production: true,
	}
	client := getApnsClient(req)
	assert.Equal(t, apns2.HostProduction, client.Host)

	req = PushNotification{
		TenantId:    tenantId,
		Development: true,
	}
	client = getApnsClient(req)
	assert.Equal(t, apns2.HostDevelopment, client.Host)

	req = PushNotification{
		TenantId: tenantId,
	}
	tenant.Ios.Production = true
	client = getApnsClient(req)
	assert.Equal(t, apns2.HostProduction, client.Host)

	tenant.Ios.Production = false
	client = getApnsClient(req)
	assert.Equal(t, apns2.HostDevelopment, client.Host)
}
