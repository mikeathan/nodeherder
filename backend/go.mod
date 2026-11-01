module node-herder

go 1.24.0

toolchain go1.24.8

require (
	github.com/eclipse/paho.mqtt.golang v1.5.0
	github.com/golang-jwt/jwt/v5 v5.3.0
	github.com/google/uuid v1.6.0
	github.com/gorilla/websocket v1.5.3
	github.com/joho/godotenv v1.5.1
	github.com/mark3labs/mcp-go v0.43.2
	github.com/miekg/dns v1.1.55
	github.com/olahol/melody v1.2.1
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/sirupsen/logrus v1.9.3
	github.com/stretchr/testify v1.10.0
	github.com/t-tomalak/logrus-easy-formatter v0.0.0-20190827215021-c074f06c5816
	go.etcd.io/bbolt v1.4.3
	go.uber.org/atomic v1.11.0
	golang.org/x/oauth2 v0.32.0
	google.golang.org/protobuf v1.34.2
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
)

require (
	github.com/flynn/noise v1.0.1-0.20220214164934-d803f5c4b0f4 // indirect
	golang.org/x/crypto v0.32.0 // indirect
	golang.org/x/mod v0.17.0 // indirect
	golang.org/x/tools v0.21.1-0.20240508182429-e35e4ccd0d2d // indirect
)

require (
	cloud.google.com/go/compute/metadata v0.3.0 // indirect
	github.com/bahlo/generic-list-go v0.2.0 // indirect
	github.com/buger/jsonparser v1.1.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/invopop/jsonschema v0.13.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mycontroller-org/esphome_api v1.4.0
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/spf13/cast v1.7.1 // indirect
	github.com/wk8/go-ordered-map/v2 v2.1.8 // indirect
	github.com/yosida95/uritemplate/v3 v3.0.2 // indirect
	golang.org/x/net v0.34.0 // indirect
	golang.org/x/sync v0.10.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
	gopkg.in/yaml.v3 v3.0.1
)

replace github.com/mycontroller-org/esphome_api => ./third_party/esphome_api
