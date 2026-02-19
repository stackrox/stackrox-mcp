module github.com/stackrox/stackrox-mcp

go 1.25

require (
	github.com/google/jsonschema-go v0.4.2
	github.com/modelcontextprotocol/go-sdk v1.3.1
	github.com/pkg/errors v0.9.1
	github.com/spf13/viper v1.21.0
	github.com/stackrox/rox v0.0.0-20210914215712-9ac265932e28
	github.com/stretchr/testify v1.11.1
	golang.stackrox.io/grpc-http1 v0.5.1
	google.golang.org/grpc v1.77.0
)

require (
	github.com/coder/websocket v1.8.14 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/golang/glog v1.2.5 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.3 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/planetscale/vtprotobuf v0.6.1-0.20240409071808-615f978279ca // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/sagikazarmark/locafero v0.11.0 // indirect
	github.com/segmentio/asm v1.1.3 // indirect
	github.com/segmentio/encoding v0.5.3 // indirect
	github.com/sourcegraph/conc v0.3.1-0.20240121214520-5f936abd7ae8 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/stackrox/scanner v0.0.0-20240830165150-d133ba942d59 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/yosida95/uritemplate/v3 v3.0.2 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/net v0.46.1-0.20251013234738-63d1a5100f82 // indirect
	golang.org/x/oauth2 v0.33.0 // indirect
	golang.org/x/sys v0.37.0 // indirect
	golang.org/x/text v0.30.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20251022142026-3a174f9686a8 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251103181224-f26f9409b101 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// StackRox library - pinned to specific commit SHA.
// Additional two libraries have to be replaced, because go is not able to resolve version "v0.0.0" used for them.
replace (
	github.com/heroku/docker-registry-client => github.com/stackrox/docker-registry-client v0.2.1
	github.com/operator-framework/helm-operator-plugins => github.com/stackrox/helm-operator v0.8.1-0.20250929095149-d1ee3c386305

	github.com/stackrox/rox => github.com/stackrox/stackrox v0.0.0-20251113103849-f9a0378795b1
)
