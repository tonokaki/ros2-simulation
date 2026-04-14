package service

import "context"

// ROSBridgeService はROS2ブリッジとの通信インターフェース
type ROSBridgeService interface {
	SendTaskCommand(ctx context.Context, taskID, robotID, action, targetLocation string) error
}
