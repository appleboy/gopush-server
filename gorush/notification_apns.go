package gorush

import (
	"crypto/ecdsa"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"net"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/certificate"
	"github.com/sideshow/apns2/payload"
	"github.com/sideshow/apns2/token"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/http2"
)

var (
	idleConnTimeout = 90 * time.Second
	tlsDialTimeout  = 20 * time.Second
	tcpKeepAlive    = 60 * time.Second
)

// DialTLS is the default dial function for creating TLS connections for
// non-proxied HTTPS requests.
var DialTLS = func(cfg *tls.Config) func(network, addr string) (net.Conn, error) {
	return func(network, addr string) (net.Conn, error) {
		dialer := &net.Dialer{
			Timeout:   tlsDialTimeout,
			KeepAlive: tcpKeepAlive,
		}
		return tls.DialWithDialer(dialer, network, addr, cfg)
	}
}

// Sound sets the aps sound on the payload.
type Sound struct {
	Critical int     `json:"critical,omitempty"`
	Name     string  `json:"name,omitempty"`
	Volume   float32 `json:"volume,omitempty"`
}

// InitAPNSClient use for initialize APNs Client.
func InitAPNSClient() error {
	if PushConf.Ios.Enabled {
		var err error
		var authKey *ecdsa.PrivateKey
		var certificateKey tls.Certificate
		var ext string

		if PushConf.Ios.KeyPath != "" {
			ext = filepath.Ext(PushConf.Ios.KeyPath)

			switch ext {
			case ".p12":
				certificateKey, err = certificate.FromP12File(PushConf.Ios.KeyPath, PushConf.Ios.Password)
			case ".pem":
				certificateKey, err = certificate.FromPemFile(PushConf.Ios.KeyPath, PushConf.Ios.Password)
			case ".p8":
				authKey, err = token.AuthKeyFromFile(PushConf.Ios.KeyPath)
			default:
				err = errors.New("wrong certificate key extension")
			}

			if err != nil {
				LogError.Error("Cert Error:", err.Error())

				return err
			}
		} else if PushConf.Ios.KeyBase64 != "" {
			ext = "." + PushConf.Ios.KeyType
			key, err := base64.StdEncoding.DecodeString(PushConf.Ios.KeyBase64)
			if err != nil {
				LogError.Error("base64 decode error:", err.Error())

				return err
			}
			switch ext {
			case ".p12":
				certificateKey, err = certificate.FromP12Bytes(key, PushConf.Ios.Password)
			case ".pem":
				certificateKey, err = certificate.FromPemBytes(key, PushConf.Ios.Password)
			case ".p8":
				authKey, err = token.AuthKeyFromBytes(key)
			default:
				err = errors.New("wrong certificate key type")
			}

			if err != nil {
				LogError.Error("Cert Error:", err.Error())

				return err
			}
		}

		if ext == ".p8" {
			if PushConf.Ios.KeyID == "" || PushConf.Ios.TeamID == "" {
				msg := "You should provide ios.KeyID and ios.TeamID for P8 token"
				LogError.Error(msg)
				return errors.New(msg)
			}
			token := &token.Token{
				AuthKey: authKey,
				// KeyID from developer account (Certificates, Identifiers & Profiles -> Keys)
				KeyID: PushConf.Ios.KeyID,
				// TeamID from developer account (View Account -> Membership)
				TeamID: PushConf.Ios.TeamID,
			}

			ApnsClient, err = newApnsTokenClient(token)
		} else {
			ApnsClient, err = newApnsClient(certificateKey)
		}

		if err != nil {
			LogError.Error("Transport Error:", err.Error())

			return err
		}
	}

	return nil
}

func newApnsClient(certificate tls.Certificate) (*apns2.Client, error) {
	var client *apns2.Client

	if PushConf.Ios.Production {
		client = apns2.NewClient(certificate).Production()
	} else {
		client = apns2.NewClient(certificate).Development()
	}

	if PushConf.Core.HTTPProxy == "" {
		return client, nil
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{certificate},
	}

	if len(certificate.Certificate) > 0 {
		tlsConfig.BuildNameToCertificate()
	}

	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
		DialTLS:         DialTLS(tlsConfig),
		Proxy:           http.DefaultTransport.(*http.Transport).Proxy,
		IdleConnTimeout: idleConnTimeout,
	}

	transportErr := http2.ConfigureTransport(transport)
	if transportErr != nil {
		return nil, transportErr
	}

	client.HTTPClient.Transport = transport

	return client, nil
}

func newApnsTokenClient(token *token.Token) (*apns2.Client, error) {
	var client *apns2.Client

	if PushConf.Ios.Production {
		client = apns2.NewTokenClient(token).Production()
	} else {
		client = apns2.NewTokenClient(token).Development()
	}

	if PushConf.Core.HTTPProxy == "" {
		return client, nil
	}

	transport := &http.Transport{
		DialTLS:         DialTLS(nil),
		Proxy:           http.DefaultTransport.(*http.Transport).Proxy,
		IdleConnTimeout: idleConnTimeout,
	}

	transportErr := http2.ConfigureTransport(transport)
	if transportErr != nil {
		return nil, transportErr
	}

	client.HTTPClient.Transport = transport

	return client, nil
}

func iosAlertDictionary(payload *payload.Payload, req PushNotification) *payload.Payload {
	// Alert dictionary

	if len(req.Title) > 0 {
		payload.AlertTitle(req.Title)
	}

	if len(req.Alert.Title) > 0 {
		payload.AlertTitle(req.Alert.Title)
	}

	// Apple Watch & Safari display this string as part of the notification interface.
	if len(req.Alert.Subtitle) > 0 {
		payload.AlertSubtitle(req.Alert.Subtitle)
	}

	if len(req.Alert.TitleLocKey) > 0 {
		payload.AlertTitleLocKey(req.Alert.TitleLocKey)
	}

	if len(req.Alert.LocArgs) > 0 {
		payload.AlertLocArgs(req.Alert.LocArgs)
	}

	if len(req.Alert.TitleLocArgs) > 0 {
		payload.AlertTitleLocArgs(req.Alert.TitleLocArgs)
	}

	if len(req.Alert.Body) > 0 {
		payload.AlertBody(req.Alert.Body)
	}

	if len(req.Alert.LaunchImage) > 0 {
		payload.AlertLaunchImage(req.Alert.LaunchImage)
	}

	if len(req.Alert.LocKey) > 0 {
		payload.AlertLocKey(req.Alert.LocKey)
	}

	if len(req.Alert.Action) > 0 {
		payload.AlertAction(req.Alert.Action)
	}

	if len(req.Alert.ActionLocKey) > 0 {
		payload.AlertActionLocKey(req.Alert.ActionLocKey)
	}

	// General
	if len(req.Category) > 0 {
		payload.Category(req.Category)
	}

	if len(req.Alert.SummaryArg) > 0 {
		payload.AlertSummaryArg(req.Alert.SummaryArg)
	}

	if req.Alert.SummaryArgCount > 0 {
		payload.AlertSummaryArgCount(req.Alert.SummaryArgCount)
	}

	return payload
}

// GetIOSNotification use for define iOS notification.
// The iOS Notification Payload
// ref: https://developer.apple.com/library/content/documentation/NetworkingInternet/Conceptual/RemoteNotificationsPG/PayloadKeyReference.html#//apple_ref/doc/uid/TP40008194-CH17-SW1
func GetIOSNotification(req PushNotification) *apns2.Notification {
	notification := &apns2.Notification{
		ApnsID:     req.ApnsID,
		Topic:      req.Topic,
		CollapseID: req.CollapseID,
	}

	if req.Expiration != nil {
		notification.Expiration = time.Unix(*req.Expiration, 0)
	}

	if len(req.Priority) > 0 {
		if req.Priority == "normal" {
			notification.Priority = apns2.PriorityLow
		} else if req.Priority == "high" {
			notification.Priority = apns2.PriorityHigh
		}
	}

	if len(req.PushType) > 0 {
		notification.PushType = apns2.EPushType(req.PushType)
	}

	payload := payload.NewPayload()

	// add alert object if message length > 0
	if len(req.Message) > 0 {
		payload.Alert(req.Message)
	}

	// zero value for clear the badge on the app icon.
	if req.Badge != nil && *req.Badge >= 0 {
		payload.Badge(*req.Badge)
	}

	if req.MutableContent {
		payload.MutableContent()
	}

	switch req.Sound.(type) {
	// from http request binding
	case map[string]interface{}:
		result := &Sound{}
		_ = mapstructure.Decode(req.Sound, &result)
		payload.Sound(result)
	// from http request binding for non critical alerts
	case string:
		payload.Sound(&req.Sound)
	case Sound:
		payload.Sound(&req.Sound)
	}

	if len(req.SoundName) > 0 {
		payload.SoundName(req.SoundName)
	}

	if req.SoundVolume > 0 {
		payload.SoundVolume(req.SoundVolume)
	}

	if req.ContentAvailable {
		payload.ContentAvailable()
	}

	if len(req.URLArgs) > 0 {
		payload.URLArgs(req.URLArgs)
	}

	if len(req.ThreadID) > 0 {
		payload.ThreadID(req.ThreadID)
	}

	for k, v := range req.Data {
		payload.Custom(k, v)
	}

	payload = iosAlertDictionary(payload, req)

	notification.Payload = payload

	return notification
}

func getApnsClient(req PushNotification) (client *apns2.Client) {
	if req.Production {
		client = ApnsClient.Production()
	} else if req.Development {
		client = ApnsClient.Development()
	} else {
		if PushConf.Ios.Production {
			client = ApnsClient.Production()
		} else {
			client = ApnsClient.Development()
		}
	}
	return
}

// PushToIOS provide send notification to APNs server.
func PushToIOS(req PushNotification) bool {
	LogAccess.Debug("Start push notification for iOS")

	var (
		retryCount = 0
		maxRetry   = PushConf.Ios.MaxRetry
	)

	if req.Retry > 0 && req.Retry < maxRetry {
		maxRetry = req.Retry
	}

Retry:
	var (
		isError   = false
		newTokens []string
	)

	notification := GetIOSNotification(req)
	client := getApnsClient(req)

	var wg sync.WaitGroup
	for _, token := range req.Tokens {
		// occupy push slot
		MaxConcurrentIOSPushes <- struct{}{}
		wg.Add(1)
		go func(token string) {
			notification.DeviceToken = token

			// send ios notification
			res, err := client.Push(notification)

			if err != nil || res.StatusCode != 200 {
				if err == nil {
					// error message:
					// ref: https://github.com/sideshow/apns2/blob/master/response.go#L14-L65
					err = errors.New(res.Reason)
				}
				// apns server error
				LogPush(FailedPush, token, req, err, res.Reason)

				if PushConf.Core.Sync {
					req.AddLog(getLogPushEntry(FailedPush, token, req, err, res.Reason))
				} else if PushConf.Core.FeedbackURL != "" {
					go func(logger *logrus.Logger, log LogPushEntry, url string, timeout int64) {
						err := DispatchFeedback(log, url, timeout)
						if err != nil {
							logger.Error(err)
						}
					}(LogError, getLogPushEntry(FailedPush, token, req, err, res.Reason), PushConf.Core.FeedbackURL, PushConf.Core.FeedbackTimeout)
				}

				StatStorage.AddIosError(1)
				// We should retry only "retryable" statuses. More info about response:
				// https://developer.apple.com/documentation/usernotifications/setting_up_a_remote_notification_server/handling_notification_responses_from_apns
				if res.StatusCode >= http.StatusInternalServerError {
					newTokens = append(newTokens, token)
				}
				isError = true
			}

			if res.Sent() && !isError {
				successLogPush(SucceededPush, token, req)
				StatStorage.AddIosSuccess(1)
			}
			// free push slot
			<-MaxConcurrentIOSPushes
			wg.Done()
		}(token)
	}
	wg.Wait()

	if isError && retryCount < maxRetry {
		retryCount++

		// resend fail token
		req.Tokens = newTokens
		goto Retry
	}

	return isError
}
