module github.com/gaia-pipeline/gaia

require (
	github.com/GeertJohan/go.rice v1.0.2
	github.com/Pallinder/go-randomdata v1.2.0
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751
	github.com/casbin/casbin/v2 v2.37.0
	github.com/containerd/containerd v1.5.5 // indirect
	github.com/daaku/go.zipexe v1.0.1 // indirect
	github.com/docker/docker v20.10.8+incompatible
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/fatih/color v1.12.0 // indirect
	github.com/gaia-pipeline/flag v1.7.4-pre
	github.com/gaia-pipeline/protobuf v0.0.0-20180812091451-7be8a901b55a
	github.com/go-openapi/jsonreference v0.19.6 // indirect
	github.com/go-openapi/spec v0.20.3 // indirect
	github.com/go-openapi/swag v0.19.15 // indirect
	github.com/gofrs/uuid v4.0.0+incompatible
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/golang/protobuf v1.5.2
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/google/go-github v17.0.0+incompatible
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/gorilla/mux v1.7.3 // indirect
	github.com/hashicorp/go-hclog v0.16.2
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/go-memdb v1.3.2
	github.com/hashicorp/go-plugin v1.4.3
	github.com/hashicorp/yamux v0.0.0-20210826001029-26ff87cf9493 // indirect
	github.com/kevinburke/ssh_config v1.1.0 // indirect
	github.com/labstack/echo/v4 v4.5.0
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/mitchellh/go-testing-interface v1.14.1 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/oklog/run v1.1.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/robfig/cron v1.2.0
	github.com/sergi/go-diff v1.2.0 // indirect
	github.com/speza/casbin-bolt-adapter v0.0.0-20200919192425-e2008c12e733
	github.com/stretchr/testify v1.6.1
	github.com/swaggo/echo-swagger v1.1.3
	github.com/swaggo/files v0.0.0-20210815190702-a29dd2bc99b2 // indirect
	github.com/swaggo/swag v1.7.1
	github.com/xanzy/ssh-agent v0.3.1 // indirect
	go.etcd.io/bbolt v1.3.6
	golang.org/x/crypto v0.0.0-20210817164053-32db794688a5
	golang.org/x/net v0.0.0-20210913180222-943fd674d43e // indirect
	golang.org/x/oauth2 v0.0.0-20210819190943-2bc19b11175f
	golang.org/x/sys v0.0.0-20210910150752-751e447fb3d0 // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/time v0.0.0-20210723032227-1f47c861a9ac // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20210909211513-a8c4777a87af // indirect
	google.golang.org/grpc v1.40.0
	gopkg.in/src-d/go-git.v4 v4.13.1
	gopkg.in/yaml.v2 v2.4.0
)

go 1.13

replace github.com/swaggo/swag => github.com/swaggo/swag v1.6.10-0.20201104153820-3f47d68f8872

replace github.com/ugorji/go/codec => github.com/ugorji/go/codec v1.2.0
