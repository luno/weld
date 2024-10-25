module example

go 1.23.2

replace github.com/luno/weld => ../../../

require (
	github.com/luno/jettison v0.0.0-20240722160230-b42bd507a5f6
	github.com/luno/weld v0.0.0
	google.golang.org/grpc v1.63.2
	google.golang.org/protobuf v1.33.0
)

require (
	github.com/go-stack/stack v1.8.1 // indirect
	golang.org/x/net v0.24.0 // indirect
	golang.org/x/sys v0.19.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	golang.org/x/xerrors v0.0.0-20231012003039-104605ab7028 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240227224415-6ceb2ff114de // indirect
)
