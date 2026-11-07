package appcommon

import "github.com/guilhermearpassos/database-monitoring/internal/runtimes"

type Service interface { //Agent extracts data
	Register(grpcRuntime *runtimes.GRPCServerRuntime, TaskRuntime *runtimes.BackGroundTaskRuntime) error
}
