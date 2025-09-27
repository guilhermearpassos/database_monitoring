
.PHONY: proto
proto:
	@protoc -I ./api --go_out=proto --go_opt=paths=source_relative --go-grpc_out=proto --go-grpc_opt=paths=source_relative --go-vtproto_out=proto --go-vtproto_opt=paths=source_relative,features=marshal+unmarshal+size ./api/database_monitoring/v1/*.proto
	@protoc -I ./api --go_out=proto --go_opt=paths=source_relative --go-grpc_out=proto --go-grpc_opt=paths=source_relative --go-vtproto_out=proto --go-vtproto_opt=paths=source_relative,features=marshal+unmarshal+size ./api/database_monitoring/v1/collector/*.proto
