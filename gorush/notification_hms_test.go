package gorush

import (
	"context"
	"log"
	"sync"
	"testing"

	"github.com/appleboy/gorush/config"
	"github.com/msalihkarakasli/go-hms-push/push/core"
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
	HMSClients = make(map[string]*core.HMSClient)
	InitWorkers(ctx, wg, PushConf.Core.WorkerNum, PushConf.Core.QueueNum)

	if err := InitAppStatus(); err != nil {
		log.Fatal(err)
	}
}

func TestMissingHuaweiAPIKey(t *testing.T) {
	PushConf, _ = config.LoadConf("")
	tenantId := "tenant_id1"
	tenant := PushConf.Tenants[tenantId]

	tenant.Huawei.Enabled = true
	tenant.Huawei.APIKey = ""

	err := CheckPushConf()

	assert.Error(t, err)
	assert.Equal(t, "missing Huawei API Key for tenant "+tenantId, err.Error())
}

func TestMissingHuaweiAPPId(t *testing.T) {
	PushConf, _ = config.LoadConf("")
	tenantId := "tenant_id1"
	tenant := PushConf.Tenants[tenantId]

	tenant.Huawei.Enabled = true
	tenant.Huawei.APPId = ""

	err := CheckPushConf()

	assert.Error(t, err)
	assert.Equal(t, "missing Huawei APP Id for tenant "+tenantId, err.Error())
}

func TestMissingKeyForInitHMSClient(t *testing.T) {
	client, err := InitHMSClient("", "", "APP_ID")

	assert.Nil(t, client)
	assert.Error(t, err)
	assert.Equal(t, "missing Huawei API Key", err.Error())
}

func TestMissingAppIDForInitHMSClient(t *testing.T) {
	client, err := InitHMSClient("", "APP_KEY", "")

	assert.Nil(t, client)
	assert.Error(t, err)
	assert.Equal(t, "missing Huawei APP Id", err.Error())
}
