package common

import "errors"

var (
	ERR_LOCK_ALADY_REQUIRED = errors.New("锁已被占用")
)
