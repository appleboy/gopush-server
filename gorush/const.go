package gorush

const (
	// PlatformIos constant is 1 for iOS
	PlatformIos = iota + 1
	// PlatformAndroid constant is 2 for Android
	PlatformAndroid
	// PlatformHuawei constant is 3 for Huawei
	PlatformHuawei
)

const (
	// SucceededPush is log block
	SucceededPush = "succeeded-push"
	// FailedPush is log block
	FailedPush = "failed-push"
)
