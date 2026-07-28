module github.com/DataDog/prof-correctness

go 1.25.1

require (
	github.com/google/pprof v0.0.0-20260507013755-92041b743c96
	github.com/grafana/jfr-parser v0.17.1
	github.com/klauspost/compress v1.18.4
	github.com/pierrec/lz4/v4 v4.1.25
	github.com/xeipuuv/gojsonschema v1.2.0
	go.opentelemetry.io/collector/pdata v1.62.0
	go.opentelemetry.io/collector/pdata/pprofile v0.156.0
)

require (
	github.com/google/gnostic v0.7.1 // indirect
	github.com/google/gnostic-models v0.7.0 // indirect
	github.com/grafana/pyroscope/api v1.5.0 // indirect
	github.com/hashicorp/go-version v1.9.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/planetscale/vtprotobuf v0.6.1-0.20250313105119-ba97887b0a25 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20180127040702-4e3ac2762d5f // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	go.opentelemetry.io/collector/featuregate v1.62.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/net v0.55.0 // indirect
	golang.org/x/sys v0.45.0 // indirect
	golang.org/x/text v0.39.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260414002931-afd174a4e478 // indirect
	google.golang.org/grpc v1.82.0 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

replace github.com/grafana/jfr-parser => github.com/r1viollet/pyroscope-jfr-parser v0.0.0-20260728101739-fe03b6cb52f6
