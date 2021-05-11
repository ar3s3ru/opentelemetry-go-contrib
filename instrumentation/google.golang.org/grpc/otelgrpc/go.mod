module go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc

go 1.14

replace go.opentelemetry.io/contrib => ../../../../

require (
	github.com/golang/protobuf v1.5.2
	github.com/stretchr/testify v1.7.0
	go.opentelemetry.io/contrib v0.20.0
	go.opentelemetry.io/otel v0.20.0
	go.opentelemetry.io/otel/metric v0.20.0
	go.opentelemetry.io/otel/oteltest v0.20.0
	go.opentelemetry.io/otel/trace v0.20.0
	go.uber.org/goleak v1.1.10
	golang.org/x/net v0.0.0-20210510120150-4163338589ed // indirect
	golang.org/x/sys v0.0.0-20210511113859-b0526f3d8744 // indirect
	google.golang.org/genproto v0.0.0-20210510173355-fb37daa5cd7a // indirect
	google.golang.org/grpc v1.37.0
)
