package gorush

import (
	"context"
	"crypto/tls"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/appleboy/gorush/config"

	"github.com/appleboy/gofight/v2"
	"github.com/buger/jsonparser"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

var goVersion = runtime.Version()

func initTest() {
	PushConf, _ = config.LoadConf("")
	PushConf.Core.Mode = "test"
}

// testRequest is testing url string if server is running
func testRequest(t *testing.T, url string) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Timeout:   time.Second * 10,
		Transport: tr,
	}

	resp, err := client.Get(url)
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Println("close body err:", err)
		}
	}()

	assert.NoError(t, err)

	_, ioerr := ioutil.ReadAll(resp.Body)
	assert.NoError(t, ioerr)
	assert.Equal(t, "200 OK", resp.Status, "should get a 200")
}

func TestPrintGoRushVersion(t *testing.T) {
	SetVersion("3.0.0")
	ver := GetVersion()
	PrintGoRushVersion()

	assert.Equal(t, "3.0.0", ver)
}

func TestRunNormalServer(t *testing.T) {
	initTest()

	gin.SetMode(gin.TestMode)

	go func() {
		assert.NoError(t, RunHTTPServer(context.Background()))
	}()
	// have to wait for the goroutine to start and run the server
	// otherwise the main thread will complete
	time.Sleep(5 * time.Millisecond)

	testRequest(t, "http://localhost:8088/api/stat/go")
}

func TestRunTLSServer(t *testing.T) {
	initTest()

	PushConf.Core.SSL = true
	PushConf.Core.Port = "8087"
	PushConf.Core.CertPath = "../certificate/localhost.cert"
	PushConf.Core.KeyPath = "../certificate/localhost.key"

	go func() {
		assert.NoError(t, RunHTTPServer(context.Background()))
	}()
	// have to wait for the goroutine to start and run the server
	// otherwise the main thread will complete
	time.Sleep(5 * time.Millisecond)

	testRequest(t, "https://localhost:8087/api/stat/go")
}

func TestRunTLSBase64Server(t *testing.T) {
	cert := `LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUMrekNDQWVPZ0F3SUJBZ0lKQUxiWkVEdlVRckZLTUEwR0NTcUdTSWIzRFFFQkJRVUFNQlF4RWpBUUJnTlYKQkFNTUNXeHZZMkZzYUc5emREQWVGdzB4TmpBek1qZ3dNek13TkRGYUZ3MHlOakF6TWpZd016TXdOREZhTUJReApFakFRQmdOVkJBTU1DV3h2WTJGc2FHOXpkRENDQVNJd0RRWUpLb1pJaHZjTkFRRUJCUUFEZ2dFUEFEQ0NBUW9DCmdnRUJBTWoxK3hnNGpWTHpWbkI1ajduMXVsMzBXRUU0QkN6Y05GeGc1QU9CNUg1cSt3amUwWVlpVkZnNlBReXYKR0NpcHFJUlhWUmRWUTFoSFNldW5ZR0tlOGxxM1NiMVg4UFVKMTJ2OXVSYnBTOURLMU93cWs4cnNQRHU2c1ZUTApxS0tnSDFaOHlhenphUzBBYlh1QTVlOWdPL1J6aWpibnBFUCtxdU00ZHVlaU1QVkVKeUxxK0VvSVFZK01NOE1QCjhkWnpMNFhabDd3TDRVc0NON3JQY082VzN0bG5UMGlPM2g5Yy9ZbTJoRmh6K0tOSjlLUlJDdnRQR1pFU2lndEsKYkhzWEgwOTlXRG84di9XcDUvZXZCdy8rSkQwb3B4bUNmSElCQUxIdDl2NTNSdnZzRFoxdDMzUnB1NUM4em5FWQpZMkF5N05neGhxanFvV0pxQTQ4bEplQTBjbHNDQXdFQUFhTlFNRTR3SFFZRFZSME9CQllFRkMwYlRVMVhvZmVoCk5LSWVsYXNoSXNxS2lkRFlNQjhHQTFVZEl3UVlNQmFBRkMwYlRVMVhvZmVoTktJZWxhc2hJc3FLaWREWU1Bd0cKQTFVZEV3UUZNQU1CQWY4d0RRWUpLb1pJaHZjTkFRRUZCUUFEZ2dFQkFBaUpMOElNVHdOWDlYcVFXWURGZ2tHNApBbnJWd1FocmVBcUM5clN4RENqcXFuTUhQSEd6Y0NlRE1MQU1vaDBrT3kyMG5vd1VHTnRDWjB1QnZuWDJxMWJOCmcxanQrR0JjTEpEUjNMTDRDcE5PbG0zWWhPeWN1TmZXTXhUQTdCWGttblNyWkQvN0toQXJzQkVZOGF1bHh3S0oKSFJnTmxJd2Uxb0ZEMVlkWDFCUzVwcDR0MjVCNlZxNEEzRk1NVWtWb1dFNjg4bkUxNjhodlFnd2pySGtnSGh3ZQplTjhsR0UyRGhGcmFYbldtRE1kd2FIRDNIUkZHaHlwcElGTitmN0JxYldYOWdNK1QyWVJUZk9iSVhMV2JxSkxECjNNay9Oa3hxVmNnNGVZNTR3SjF1ZkNVR0FZQUlhWTZmUXFpTlV6OG5od0szdDQ1TkJWVDl5L3VKWHFuVEx5WT0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=`
	key := `LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFb2dJQkFBS0NBUUVBeVBYN0dEaU5Vdk5XY0htUHVmVzZYZlJZUVRnRUxOdzBYR0RrQTRIa2ZtcjdDTjdSCmhpSlVXRG85REs4WUtLbW9oRmRWRjFWRFdFZEo2NmRnWXA3eVdyZEp2VmZ3OVFuWGEvMjVGdWxMME1yVTdDcVQKeXV3OE83cXhWTXVvb3FBZlZuekpyUE5wTFFCdGU0RGw3MkE3OUhPS051ZWtRLzZxNHpoMjU2SXc5VVFuSXVyNApTZ2hCajR3end3L3gxbk12aGRtWHZBdmhTd0kzdXM5dzdwYmUyV2RQU0k3ZUgxejlpYmFFV0hQNG8wbjBwRkVLCiswOFprUktLQzBwc2V4Y2ZUMzFZT2p5Lzlhbm45NjhIRC80a1BTaW5HWUo4Y2dFQXNlMzIvbmRHKyt3Tm5XM2YKZEdtN2tMek9jUmhqWURMczJER0dxT3FoWW1vRGp5VWw0RFJ5V3dJREFRQUJBb0lCQUdUS3FzTjlLYlNmQTQycQpDcUkwVXVMb3VKTU5hMXFzbno1dUFpNllLV2dXZEE0QTQ0bXBFakNtRlJTVmhVSnZ4V3VLK2N5WUlRelh4SVdECkQxNm5aZHFGNzJBZUNXWjlKeVNzdnZaMDBHZktNM3kzNWlSeTA4c0pXZ096bWNMbkdKQ2lTZXlLc1FlM0hUSkMKZGhEWGJYcXZzSFRWUFpnMDFMVGVEeFVpVGZmVThOTUtxUjJBZWNRMnNURHdYRWhBblR5QXRuemwvWGFCZ0Z6dQpVNkc3RnpHTTV5OWJ4a2ZRVmt2eStERUprSEdOT2p6d2NWZkJ5eVZsNjEwaXhtRzF2bXhWajlQYldtSVBzVVY4CnlTbWpodkRRYk9mb3hXMGg5dlRsVHFHdFFjQnc5NjJvc25ERE1XRkNkTTdsek8wVDdSUm5QVkdJUnBDSk9LaHEKa2VxSEt3RUNnWUVBOHd3SS9pWnVnaG9UWFRORzlMblFRL1dBdHNxTzgwRWpNVFVoZW81STFrT3ptVXowOXB5aAppQXNVRG9OMC8yNnRaNVdOamxueVp1N2R2VGMveDNkVFpwbU5ub284Z2NWYlFORUNEUnpxZnVROVBQWG0xU041CjZwZUJxQXZCdjc4aGpWMDVhWHpQRy9WQmJlaWc3bDI5OUVhckVBK2Evb0gzS3JnRG9xVnFFMEVDZ1lFQTA2dkEKWUptZ2c0ZlpSdWNBWW9hWXNMejlaOXJDRmpUZTFQQlRtVUprYk9SOHZGSUhIVFRFV2kvU3V4WEwwd0RTZW9FMgo3QlFtODZnQ0M3L0tnUmRyem9CcVo1cVM5TXYyZHNMZ1k2MzVWU2dqamZaa1ZMaUgxVlJScFNRT2JZbmZveXNnCmdhdGNIU0tNRXhkNFNMUUJ5QXVJbVhQK0w1YXlEQmNFSmZicVNwc0NnWUI3OElzMWIwdXpOTERqT2g3WTlWaHIKRDJxUHpFT1JjSW9Oc2RaY3RPb1h1WGFBbW1uZ3lJYm01UjlaTjFnV1djNDdvRndMVjNyeFdxWGdzNmZtZzhjWAo3djMwOXZGY0M5UTQvVnhhYTRCNUxOSzluM2dUQUlCUFRPdGxVbmwrMm15MXRmQnRCcVJtMFc2SUtiVEhXUzVnCnZ4akVtL0NpRUl5R1VFZ3FUTWdIQVFLQmdCS3VYZFFvdXRuZzYzUXVmd0l6RHRiS1Z6TUxRNFhpTktobWJYcGgKT2F2Q25wK2dQYkIrTDdZbDhsdEFtVFNPSmdWWjBoY1QwRHhBMzYxWngrMk11NThHQmw0T2JsbmNobXdFMXZqMQpLY1F5UHJFUXhkb1VUeWlzd0dmcXZyczhKOWltdmIrejkvVTZUMUtBQjhXaTNXVmlYelByNE1zaWFhUlhnNjQyCkZJZHhBb0dBWjcvNzM1ZGtoSmN5T2ZzK0xLc0xyNjhKU3N0b29yWE9ZdmRNdTErSkdhOWlMdWhuSEVjTVZXQzgKSXVpaHpQZmxvWnRNYkdZa1pKbjhsM0JlR2Q4aG1mRnRnVGdaR1BvVlJldGZ0MkxERkxuUHhwMnNFSDVPRkxzUQpSK0sva0FPdWw4ZVN0V3VNWE9GQTlwTXpHa0dFZ0lGSk1KT3lhSk9OM2tlZFFJOGRlQ009Ci0tLS0tRU5EIFJTQSBQUklWQVRFIEtFWS0tLS0tCg==`
	initTest()

	PushConf.Core.SSL = true
	PushConf.Core.Port = "8089"
	PushConf.Core.CertPath = ""
	PushConf.Core.KeyPath = ""
	PushConf.Core.CertBase64 = cert
	PushConf.Core.KeyBase64 = key

	go func() {
		assert.NoError(t, RunHTTPServer(context.Background()))
	}()
	// have to wait for the goroutine to start and run the server
	// otherwise the main thread will complete
	time.Sleep(5 * time.Millisecond)

	testRequest(t, "https://localhost:8089/api/stat/go")
}

func TestRunAutoTLSServer(t *testing.T) {
	initTest()
	PushConf.Core.AutoTLS.Enabled = true
	go func() {
		assert.NoError(t, RunHTTPServer(context.Background()))
	}()
	// have to wait for the goroutine to start and run the server
	// otherwise the main thread will complete
	time.Sleep(5 * time.Millisecond)
}

func TestLoadTLSCertError(t *testing.T) {
	initTest()

	PushConf.Core.SSL = true
	PushConf.Core.Port = "8087"
	PushConf.Core.CertPath = "../config/config.yml"
	PushConf.Core.KeyPath = "../config/config.yml"

	assert.Error(t, RunHTTPServer(context.Background()))
}

func TestMissingTLSCertConfg(t *testing.T) {
	initTest()

	PushConf.Core.SSL = true
	PushConf.Core.Port = "8087"
	PushConf.Core.CertPath = ""
	PushConf.Core.KeyPath = ""
	PushConf.Core.CertBase64 = ""
	PushConf.Core.KeyBase64 = ""

	err := RunHTTPServer(context.Background())
	assert.Error(t, RunHTTPServer(context.Background()))
	assert.Equal(t, "missing https cert config", err.Error())
}

func TestRootHandler(t *testing.T) {
	initTest()

	r := gofight.New()

	// log for json
	PushConf.Log.Format = "json"

	r.GET("/").
		Run(routerEngine(), func(r gofight.HTTPResponse, rq gofight.HTTPRequest) {
			data := r.Body.Bytes()

			value, _ := jsonparser.GetString(data, "text")

			assert.Equal(t, "Welcome to notification server.", value)
			assert.Equal(t, http.StatusOK, r.Code)
			assert.Equal(t, "application/json; charset=utf-8", r.HeaderMap.Get("Content-Type"))
		})
}

func TestAPIStatusGoHandler(t *testing.T) {
	initTest()

	r := gofight.New()

	r.GET("/api/stat/go").
		Run(routerEngine(), func(r gofight.HTTPResponse, rq gofight.HTTPRequest) {
			data := r.Body.Bytes()

			value, _ := jsonparser.GetString(data, "go_version")

			assert.Equal(t, goVersion, value)
			assert.Equal(t, http.StatusOK, r.Code)
		})
}

func TestAPIStatusAppHandler(t *testing.T) {
	initTest()

	r := gofight.New()

	appVersion := "v1.0.0"
	SetVersion(appVersion)

	r.GET("/api/stat/app").
		Run(routerEngine(), func(r gofight.HTTPResponse, rq gofight.HTTPRequest) {
			data := r.Body.Bytes()

			value, _ := jsonparser.GetString(data, "version")

			assert.Equal(t, appVersion, value)
			assert.Equal(t, http.StatusOK, r.Code)
		})
}

func TestAPIConfigHandler(t *testing.T) {
	initTest()

	r := gofight.New()

	r.GET("/api/config").
		Run(routerEngine(), func(r gofight.HTTPResponse, rq gofight.HTTPRequest) {
			assert.Equal(t, http.StatusCreated, r.Code)
		})
}

func TestMissingNotificationsParameter(t *testing.T) {
	initTest()
	tenantId := "tenant_id1"
	tenant := PushConf.Tenants[tenantId]
	tenant.PushURI = "/api/push/" + tenantId

	r := gofight.New()

	// missing notifications parameter.
	r.POST(tenant.PushURI).
		Run(routerEngine(), func(r gofight.HTTPResponse, rq gofight.HTTPRequest) {
			assert.Equal(t, http.StatusBadRequest, r.Code)
			assert.Equal(t, "application/json; charset=utf-8", r.HeaderMap.Get("Content-Type"))
		})
}

func TestEmptyNotifications(t *testing.T) {
	initTest()
	tenantId := "tenant_id1"
	tenant := PushConf.Tenants[tenantId]
	tenant.PushURI = "/api/push/" + tenantId

	r := gofight.New()

	// notifications is empty.
	r.POST(tenant.PushURI).
		SetJSON(gofight.D{
			"notifications": []PushNotification{},
		}).
		Run(routerEngine(), func(r gofight.HTTPResponse, rq gofight.HTTPRequest) {
			assert.Equal(t, http.StatusBadRequest, r.Code)
		})
}

func TestMutableContent(t *testing.T) {
	initTest()
	tenantId := "tenant_id1"
	tenant := PushConf.Tenants[tenantId]
	tenant.PushURI = "/api/push/" + tenantId

	r := gofight.New()

	// notifications is empty.
	r.POST(tenant.PushURI).
		SetJSON(gofight.D{
			"notifications": []gofight.D{
				{
					"tokens":          []string{"aaaaa", "bbbbb"},
					"platform":        PlatformAndroid,
					"message":         "Welcome",
					"mutable_content": 1,
					"topic":           "test",
					"badge":           1,
					"alert": gofight.D{
						"title": "title",
						"body":  "body",
					},
				},
			},
		}).
		Run(routerEngine(), func(r gofight.HTTPResponse, rq gofight.HTTPRequest) {
			// json: cannot unmarshal number into Go struct field PushNotification.mutable_content of type bool
			assert.Equal(t, http.StatusBadRequest, r.Code)
		})
}

func TestOutOfRangeMaxNotifications(t *testing.T) {
	initTest()
	tenantId := "tenant_id1"
	tenant := PushConf.Tenants[tenantId]
	tenant.PushURI = "/api/push/" + tenantId

	PushConf.Core.MaxNotification = int64(1)

	r := gofight.New()

	// notifications is empty.
	r.POST(tenant.PushURI).
		SetJSON(gofight.D{
			"notifications": []gofight.D{
				{
					"tokens":   []string{"aaaaa", "bbbbb"},
					"platform": PlatformAndroid,
					"message":  "Welcome",
				},
				{
					"tokens":   []string{"aaaaa", "bbbbb"},
					"platform": PlatformAndroid,
					"message":  "Welcome",
				},
			},
		}).
		Run(routerEngine(), func(r gofight.HTTPResponse, rq gofight.HTTPRequest) {
			assert.Equal(t, http.StatusBadRequest, r.Code)
		})
}

func TestSuccessPushHandler(t *testing.T) {
	t.Skip()
	initTest()
	tenantId := "tenant_id1"
	tenant := PushConf.Tenants[tenantId]
	tenant.PushURI = "/api/push/" + tenantId

	tenant.Android.Enabled = true
	tenant.Android.APIKey = os.Getenv("ANDROID_API_KEY")

	androidToken := os.Getenv("ANDROID_TEST_TOKEN")

	r := gofight.New()

	r.POST(tenant.PushURI).
		SetJSON(gofight.D{
			"notifications": []gofight.D{
				{
					"tokens":   []string{androidToken, "bbbbb"},
					"platform": PlatformAndroid,
					"message":  "Welcome",
				},
			},
		}).
		Run(routerEngine(), func(r gofight.HTTPResponse, rq gofight.HTTPRequest) {
			assert.Equal(t, http.StatusOK, r.Code)
		})
}

func TestSuccessPushHandlerMultiTenant(t *testing.T) {
	t.Skip()
	initTest()
	tenantId1 := "tenant_id1"
	tenant1 := config.SectionTenant{}
	PushConf.Tenants[tenantId1] = &tenant1
	tenant1.PushURI = "/api/push/" + tenantId1
	tenant1.Ios.Enabled = false
	tenant1.Huawei.Enabled = false
	tenant1.Android.Enabled = true
	tenant1.Android.APIKey = os.Getenv("ANDROID_API_KEY_1")
	androidToken1 := os.Getenv("ANDROID_TEST_TOKEN_1")

	tenantId2 := "tenant_id2"
	tenant2 := config.SectionTenant{}
	PushConf.Tenants[tenantId2] = &tenant2
	tenant2.PushURI = "/api/push/"
	tenant2.Ios.Enabled = false
	tenant2.Huawei.Enabled = false
	tenant2.Android.Enabled = true
	tenant2.Android.APIKey = os.Getenv("ANDROID_API_KEY_2")
	androidToken2 := os.Getenv("ANDROID_TEST_TOKEN_2")

	tenantId3 := "tenant_id3"
	tenant3 := config.SectionTenant{}
	PushConf.Tenants[tenantId3] = &tenant3
	tenant3.PushURI = "/api/push/" + tenantId3
	tenant3.Ios.Enabled = false
	tenant3.Huawei.Enabled = false
	tenant3.Android.Enabled = true
	tenant3.Android.APIKey = os.Getenv("ANDROID_API_KEY_3")

	r := gofight.New()

	r.POST(tenant1.PushURI).
		SetJSON(gofight.D{
			"notifications": []gofight.D{
				{
					"tokens":   []string{androidToken1},
					"platform": PlatformAndroid,
					"message":  "Welcome 1",
				},
			},
		}).
		Run(routerEngine(), func(r gofight.HTTPResponse, rq gofight.HTTPRequest) {
			assert.Equal(t, http.StatusOK, r.Code)
		})

	r.POST(tenant2.PushURI).
		SetJSON(gofight.D{
			"notifications": []gofight.D{
				{
					"tokens":   []string{androidToken2},
					"platform": PlatformAndroid,
					"message":  "Welcome 2",
				},
			},
		}).
		Run(routerEngine(), func(r gofight.HTTPResponse, rq gofight.HTTPRequest) {
			assert.Equal(t, http.StatusBadRequest, r.Code)
		})

	r.POST(tenant3.PushURI).
		SetJSON(gofight.D{
			"notifications": []gofight.D{
				{
					"tokens":   []string{""},
					"platform": PlatformAndroid,
					"message":  "Welcome 3",
				},
			},
		}).
		Run(routerEngine(), func(r gofight.HTTPResponse, rq gofight.HTTPRequest) {
			assert.Equal(t, http.StatusOK, r.Code)
		})
}

func TestSysStatsHandler(t *testing.T) {
	initTest()

	r := gofight.New()

	r.GET("/sys/stats").
		Run(routerEngine(), func(r gofight.HTTPResponse, rq gofight.HTTPRequest) {
			assert.Equal(t, http.StatusOK, r.Code)
		})
}

func TestMetricsHandler(t *testing.T) {
	initTest()

	r := gofight.New()

	r.GET("/metrics").
		Run(routerEngine(), func(r gofight.HTTPResponse, rq gofight.HTTPRequest) {
			assert.Equal(t, http.StatusOK, r.Code)
		})
}

func TestGETHeartbeatHandler(t *testing.T) {
	initTest()

	r := gofight.New()

	r.GET("/healthz").
		Run(routerEngine(), func(r gofight.HTTPResponse, rq gofight.HTTPRequest) {
			assert.Equal(t, http.StatusOK, r.Code)
		})
}

func TestHEADHeartbeatHandler(t *testing.T) {
	initTest()

	r := gofight.New()

	r.HEAD("/healthz").
		Run(routerEngine(), func(r gofight.HTTPResponse, rq gofight.HTTPRequest) {
			assert.Equal(t, http.StatusOK, r.Code)
		})
}

func TestVersionHandler(t *testing.T) {
	SetVersion("3.0.0")
	initTest()

	r := gofight.New()

	r.GET("/version").
		Run(routerEngine(), func(r gofight.HTTPResponse, rq gofight.HTTPRequest) {
			assert.Equal(t, http.StatusOK, r.Code)
			data := r.Body.Bytes()

			value, _ := jsonparser.GetString(data, "version")

			assert.Equal(t, "3.0.0", value)
		})
}

func TestDisabledHTTPServer(t *testing.T) {
	initTest()
	PushConf.Core.Enabled = false
	err := RunHTTPServer(context.Background())
	PushConf.Core.Enabled = true

	assert.Nil(t, err)
}
