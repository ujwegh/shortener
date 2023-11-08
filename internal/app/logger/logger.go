package logger

import (
	"go.uber.org/zap"
)

var Log *zap.Logger = zap.NewNop()

func Initialize(zl *zap.Logger) {
	Log = zl
}
