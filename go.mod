module github.com/alexandreLamarre/pprof-server

go 1.23.0

replace github.com/alexandreLamarre/otelbpf/receiver/pprofreceiver => github.com/alexandreLamarre/otelcol-bpf/receiver/pprofreceiver v0.0.0-20240822220648-dbccb34e2f62

replace github.com/google/pprof v0.0.0-20240727154555-813a5fbdbec8 => github.com/alexandreLamarre/pprof v0.0.0-20240823000903-9c0b95314838

require (
	github.com/alexandreLamarre/otelbpf/receiver/pprofreceiver v0.0.0-20240822220648-dbccb34e2f62
	github.com/google/pprof v0.0.0-20240727154555-813a5fbdbec8
	github.com/samber/lo v1.47.0
	github.com/sirupsen/logrus v1.9.3
	github.com/spf13/cobra v1.8.1
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.53.0
	go.opentelemetry.io/proto/otlp v1.3.1
	google.golang.org/grpc v1.65.0
	google.golang.org/protobuf v1.34.2
)

require (
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.20.0 // indirect
	github.com/ianlancetaylor/demangle v0.0.0-20240312041847-bd984b5ce465 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	go.opentelemetry.io/collector/component v0.107.0 // indirect
	go.opentelemetry.io/collector/config/configtelemetry v0.107.0 // indirect
	go.opentelemetry.io/collector/consumer v0.107.0 // indirect
	go.opentelemetry.io/collector/consumer/consumerprofiles v0.107.0 // indirect
	go.opentelemetry.io/collector/pdata v1.13.0 // indirect
	go.opentelemetry.io/collector/pdata/pprofile v0.107.0 // indirect
	go.opentelemetry.io/collector/receiver v0.107.0 // indirect
	go.opentelemetry.io/otel v1.28.0 // indirect
	go.opentelemetry.io/otel/metric v1.28.0 // indirect
	go.opentelemetry.io/otel/trace v1.28.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/net v0.26.0 // indirect
	golang.org/x/sys v0.22.0 // indirect
	golang.org/x/text v0.16.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240528184218-531527333157 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240701130421-f6361c86f094 // indirect
)
