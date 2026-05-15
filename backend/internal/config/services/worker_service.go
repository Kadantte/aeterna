package services

import (
	"github.com/alpyxn/aeterna/backend/internal/config/common"
)

type WorkerModule struct{}

func (WorkerModule) Name() string { return "WorkerModule" }
func (WorkerModule) Section() string {
	return "worker"
}

func init() {
	common.Register(WorkerModule{})
}

type WorkerSection struct {
	BaseURL string
}

func (WorkerModule) LoadAndValidate() (WorkerSection, error) {
	return WorkerSection{
		BaseURL: common.WithDefault(common.GetenvTrim("BASE_URL"), common.DefaultWorkerBaseURL),
	}, nil
}
