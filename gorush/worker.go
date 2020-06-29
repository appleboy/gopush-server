package gorush

import (
	"context"
	"errors"
	"sync"
)

// InitWorkers for initialize all workers.
func InitWorkers(ctx context.Context, wg *sync.WaitGroup, workerNum int64, queueNum int64) {
	LogAccess.Info("worker number is ", workerNum, ", queue number is ", queueNum)
	QueueNotification = make(chan PushNotification, queueNum)
	for i := int64(0); i < workerNum; i++ {
		go startWorker(ctx, wg, i)
	}
}

// SendNotification is send message to iOS or Android
func SendNotification(ctx context.Context, req PushNotification) {
	if PushConf.Core.Sync {
		defer req.WaitDone()
	}

	switch req.Platform {
	case PlatFormIos:
		PushToIOS(req)
	case PlatFormAndroid:
		PushToAndroid(req)
	}
}

func startWorker(ctx context.Context, wg *sync.WaitGroup, num int64) {
	defer wg.Done()
	for notification := range QueueNotification {
		SendNotification(ctx, notification)
	}
	LogAccess.Info("closed the worker num ", num)
}

// markFailedNotification adds failure logs for all tokens in push notification
func markFailedNotification(notification *PushNotification, reason string) {
	LogError.Error(reason)
	for _, token := range notification.Tokens {
		notification.AddLog(getLogPushEntry(FailedPush, token, *notification, errors.New(reason), reason))
	}
	notification.WaitDone()
}

// queueNotification add notification to queue list.
func queueNotification(ctx context.Context, req RequestPush) (int, []LogPushEntry) {
	var count int
	wg := sync.WaitGroup{}
	newNotification := []*PushNotification{}
	for i := range req.Notifications {
		notification := &req.Notifications[i]
		switch notification.Platform {
		case PlatFormIos:
			if !PushConf.Ios.Enabled {
				continue
			}
		case PlatFormAndroid:
			if !PushConf.Android.Enabled {
				continue
			}
		}
		newNotification = append(newNotification, notification)
	}

	log := make([]LogPushEntry, 0, count)
	for _, notification := range newNotification {
		if PushConf.Core.Sync {
			notification.wg = &wg
			notification.log = &log
			notification.AddWaitCount()
		}
		if !tryEnqueue(*notification, QueueNotification) {
			markFailedNotification(notification, "max capacity reached")
		}
		count += len(notification.Tokens)
		// Count topic message
		if notification.To != "" {
			count++
		}
	}

	if PushConf.Core.Sync {
		wg.Wait()
	}

	StatStorage.AddTotalCount(int64(count))

	return count, log
}

// tryEnqueue tries to enqueue a job to the given job channel. Returns true if
// the operation was successful, and false if enqueuing would not have been
// possible without blocking. Job is not enqueued in the latter case.
func tryEnqueue(job PushNotification, jobChan chan<- PushNotification) bool {
	select {
	case jobChan <- job:
		return true
	default:
		return false
	}
}
