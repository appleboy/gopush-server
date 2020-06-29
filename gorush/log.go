package gorush

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/sirupsen/logrus"
)

var (
	green   = string([]byte{27, 91, 57, 55, 59, 52, 50, 109})
	white   = string([]byte{27, 91, 57, 48, 59, 52, 55, 109})
	yellow  = string([]byte{27, 91, 57, 55, 59, 52, 51, 109})
	red     = string([]byte{27, 91, 57, 55, 59, 52, 49, 109})
	blue    = string([]byte{27, 91, 57, 55, 59, 52, 52, 109})
	magenta = string([]byte{27, 91, 57, 55, 59, 52, 53, 109})
	cyan    = string([]byte{27, 91, 57, 55, 59, 52, 54, 109})
	reset   = string([]byte{27, 91, 48, 109})
)

// LogPushEntry is push response log
type LogPushEntry struct {
	ID                string `json:"notif_id,omitempty"`
	Type              string `json:"type"`
	Platform          string `json:"platform"`
	Token             string `json:"token"`
	Message           string `json:"message"`
	Error             string `json:"error"`
	ErrorResponseCode string `json:"error_response_code"`
}

var isTerm bool

func init() {
	isTerm = isatty.IsTerminal(os.Stdout.Fd())
}

// InitLog use for initial log module
func InitLog() error {

	var err error

	// init logger
	LogAccess = logrus.New()
	LogError = logrus.New()

	if !isTerm {
		LogAccess.SetFormatter(&logrus.JSONFormatter{})
		LogError.SetFormatter(&logrus.JSONFormatter{})
	} else {
		LogAccess.Formatter = &logrus.TextFormatter{
			TimestampFormat: "2006/01/02 - 15:04:05",
			FullTimestamp:   true,
		}

		LogError.Formatter = &logrus.TextFormatter{
			TimestampFormat: "2006/01/02 - 15:04:05",
			FullTimestamp:   true,
		}
	}

	// set logger
	if err = SetLogLevel(LogAccess, PushConf.Log.AccessLevel); err != nil {
		return errors.New("Set access log level error: " + err.Error())
	}

	if err = SetLogLevel(LogError, PushConf.Log.ErrorLevel); err != nil {
		return errors.New("Set error log level error: " + err.Error())
	}

	if err = SetLogOut(LogAccess, PushConf.Log.AccessLog); err != nil {
		return errors.New("Set access log path error: " + err.Error())
	}

	if err = SetLogOut(LogError, PushConf.Log.ErrorLog); err != nil {
		return errors.New("Set error log path error: " + err.Error())
	}

	return nil
}

// SetLogOut provide log stdout and stderr output
func SetLogOut(log *logrus.Logger, outString string) error {
	switch outString {
	case "stdout":
		log.Out = os.Stdout
	case "stderr":
		log.Out = os.Stderr
	default:
		f, err := os.OpenFile(outString, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)

		if err != nil {
			return err
		}

		log.Out = f
	}

	return nil
}

// SetLogLevel is define log level what you want
// log level: panic, fatal, error, warn, info and debug
func SetLogLevel(log *logrus.Logger, levelString string) error {
	level, err := logrus.ParseLevel(levelString)

	if err != nil {
		return err
	}

	log.Level = level

	return nil
}

func colorForPlatForm(platform int) string {
	switch platform {
	case PlatFormIos:
		return blue
	case PlatFormAndroid:
		return yellow
	default:
		return reset
	}
}

func typeForPlatForm(platform int) string {
	switch platform {
	case PlatFormIos:
		return "ios"
	case PlatFormAndroid:
		return "android"
	default:
		return ""
	}
}

func hideToken(token string, markLen int) string {
	if len(token) == 0 {
		return ""
	}

	if len(token) < markLen*2 {
		return strings.Repeat("*", len(token))
	}

	start := token[len(token)-markLen:]
	end := token[0:markLen]

	result := strings.Replace(token, start, strings.Repeat("*", markLen), -1)
	result = strings.Replace(result, end, strings.Repeat("*", markLen), -1)

	return result
}

func getLogPushEntry(status, token string, req PushNotification, errPush error, errorResponseCode string) LogPushEntry {
	var errMsg string

	plat := typeForPlatForm(req.Platform)

	if errPush != nil {
		errMsg = errPush.Error()
	}

	if PushConf.Log.HideToken {
		token = hideToken(token, 10)
	}

	return LogPushEntry{
		ID:                req.ID,
		Type:              status,
		Platform:          plat,
		Token:             token,
		Message:           req.Message,
		Error:             errMsg,
		ErrorResponseCode: errorResponseCode,
	}
}

func successLogPush(status, token string, req PushNotification) {
	LogPush(status, token, req, nil, "")
}

// LogPush record user push request and server response.
func LogPush(status, token string, req PushNotification, errPush error, errorResponseCode string) {
	var platColor, resetColor, output string

	if isTerm {
		platColor = colorForPlatForm(req.Platform)
		resetColor = reset
	}

	log := getLogPushEntry(status, token, req, errPush, errorResponseCode)

	if PushConf.Log.Format == "json" {
		logJSON, _ := json.Marshal(log)

		output = string(logJSON)
	} else {
		var typeColor string
		switch status {
		case SucceededPush:
			if isTerm {
				typeColor = green
			}

			output = fmt.Sprintf("|%s %s %s| %s%s%s [%s] %s",
				typeColor, log.Type, resetColor,
				platColor, log.Platform, resetColor,
				log.Token,
				log.Message,
			)
		case FailedPush:
			if isTerm {
				typeColor = red
			}

			output = fmt.Sprintf("|%s %s %s| %s%s%s [%s] | %s | Error Message: %s | ErrorResponseCode: %s",
				typeColor, log.Type, resetColor,
				platColor, log.Platform, resetColor,
				log.Token,
				log.Message,
				log.Error,
				log.ErrorResponseCode,
			)
		}
	}

	switch status {
	case SucceededPush:
		LogAccess.Info(output)
	case FailedPush:
		LogError.Error(output)
	}
}
