module github.com/teamkeel/keel

go 1.22

toolchain go1.23.2

require (
	github.com/99designs/gqlgen v0.17.16
	github.com/Masterminds/semver/v3 v3.2.1
	github.com/PaesslerAG/jsonpath v0.1.1
	github.com/TylerBrock/colorjson v0.0.0-20200706003622-8a50f05110d2
	github.com/alecthomas/participle/v2 v2.0.0-beta.5
	github.com/aws/aws-lambda-go v1.47.0
	github.com/aws/aws-sdk-go v1.44.298
	github.com/aws/aws-sdk-go-v2 v1.32.4
	github.com/aws/aws-sdk-go-v2/config v1.27.23
	github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs v1.43.2
	github.com/aws/aws-sdk-go-v2/service/lambda v1.65.0
	github.com/aws/aws-sdk-go-v2/service/s3 v1.65.2
	github.com/aws/aws-sdk-go-v2/service/secretsmanager v1.34.2
	github.com/aws/aws-sdk-go-v2/service/sqs v1.37.0
	github.com/aws/aws-sdk-go-v2/service/ssm v1.55.2
	github.com/aws/aws-sdk-go-v2/service/sts v1.32.2
	github.com/aws/smithy-go v1.22.0
	github.com/bmatcuk/doublestar/v4 v4.2.0
	github.com/bykof/gostradamus v1.0.4
	github.com/charmbracelet/bubbles v0.16.1
	github.com/charmbracelet/bubbletea v0.25.0
	github.com/charmbracelet/lipgloss v0.10.0
	github.com/coreos/go-oidc v2.2.1+incompatible
	github.com/dchest/uniuri v1.2.0
	github.com/docker/docker v24.0.9+incompatible
	github.com/docker/go-connections v0.4.0
	github.com/evanw/esbuild v0.24.0
	github.com/fatih/camelcase v1.0.0
	github.com/goccy/go-yaml v1.12.0
	github.com/golang-jwt/jwt/v4 v4.4.2
	github.com/hashicorp/go-version v1.6.0
	github.com/iancoleman/strcase v0.2.0
	github.com/jackc/pgx/v5 v5.5.5
	github.com/jinzhu/inflection v1.0.0
	github.com/jonbretman/gotestpretty v0.0.0-20200908080245-691cf3e7550c
	github.com/ka-weihe/fast-levenshtein v0.0.0-20201227151214-4c99ee36a1ba
	github.com/karlseguin/typed v1.1.8
	github.com/lib/pq v1.10.9
	github.com/manifoldco/promptui v0.9.0
	github.com/nleeper/goment v1.4.4
	github.com/nsf/jsondiff v0.0.0-20230430225905-43f6cf3098c1
	github.com/otiai10/copy v1.7.0
	github.com/pulumi/pulumi-aws/sdk/v6 v6.55.0
	github.com/pulumi/pulumi/sdk/v3 v3.136.1
	github.com/radovskyb/watcher v1.0.7
	github.com/relvacode/iso8601 v1.3.0
	github.com/rs/cors v1.8.2
	github.com/samber/lo v1.28.0
	github.com/sanity-io/litter v1.5.5
	github.com/segmentio/ksuid v1.0.4
	github.com/sergi/go-diff v1.3.2-0.20230802210424-5b0b94c5c0d3
	github.com/spf13/cobra v1.8.0
	github.com/spf13/viper v1.15.0
	github.com/stretchr/testify v1.9.0
	github.com/teamkeel/graphql v0.8.2-0.20230531102419-995b8ab035b6
	github.com/twitchtv/twirp v8.1.3+incompatible
	github.com/vincent-petithory/dataurl v1.0.0
	github.com/xeipuuv/gojsonschema v1.2.0
	go.opentelemetry.io/otel v1.31.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.21.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.21.0
	go.opentelemetry.io/otel/sdk v1.31.0
	go.opentelemetry.io/otel/trace v1.31.0
	go.opentelemetry.io/proto/otlp v1.0.0
	golang.org/x/exp v0.0.0-20240604190554-fc45aab8b7f8
	golang.org/x/oauth2 v0.17.0
	golang.org/x/sync v0.7.0
	google.golang.org/protobuf v1.34.0
	gopkg.in/yaml.v3 v3.0.1
	gorm.io/driver/postgres v1.5.0
	gorm.io/gorm v1.25.1
)

require (
	dario.cat/mergo v1.0.0 // indirect
	github.com/BurntSushi/toml v1.2.1 // indirect
	github.com/PaesslerAG/gval v1.0.0 // indirect
	github.com/ProtonMail/go-crypto v1.0.0 // indirect
	github.com/aead/chacha20 v0.0.0-20180709150244-8b13a72661da // indirect
	github.com/agext/levenshtein v1.2.3 // indirect
	github.com/apparentlymart/go-textseg/v13 v13.0.0 // indirect
	github.com/atotto/clipboard v0.1.4 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.6.6 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.17.23 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.9 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.23 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.23 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.0 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.3.21 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.12.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.4.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.12.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.18.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.22.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.26.1 // indirect
	github.com/aymanbagabas/go-osc52/v2 v2.0.1 // indirect
	github.com/blang/semver v3.5.1+incompatible // indirect
	github.com/cenkalti/backoff/v4 v4.2.1 // indirect
	github.com/cheggaaa/pb v1.0.29 // indirect
	github.com/chzyer/readline v0.0.0-20180603132655-2972be24d48e // indirect
	github.com/cloudflare/circl v1.3.7 // indirect
	github.com/containerd/console v1.0.4-0.20230313162750-1ae8d489ac81 // indirect
	github.com/cyphar/filepath-securejoin v0.2.4 // indirect
	github.com/djherbis/times v1.5.0 // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/fatih/color v1.16.0 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/go-git/gcfg v1.5.1-0.20230307220236-3a3c6141e376 // indirect
	github.com/go-git/go-billy/v5 v5.5.0 // indirect
	github.com/go-git/go-git/v5 v5.12.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/golang/glog v1.2.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.16.0 // indirect
	github.com/grpc-ecosystem/grpc-opentracing v0.0.0-20180507213350-8e809c8a8645 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/hashicorp/hcl/v2 v2.17.0 // indirect
	github.com/hokaccha/go-prettyjson v0.0.0-20211117102719-0474bc63780f // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/kevinburke/ssh_config v1.2.0 // indirect
	github.com/lucasb-eyer/go-colorful v1.2.0 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mattn/go-localereader v0.0.1 // indirect
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/mitchellh/go-ps v1.0.0 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/muesli/ansi v0.0.0-20230316100256-276c6243b2f6 // indirect
	github.com/muesli/cancelreader v0.2.2 // indirect
	github.com/muesli/reflow v0.3.0 // indirect
	github.com/muesli/termenv v0.15.2 // indirect
	github.com/nxadm/tail v1.4.11 // indirect
	github.com/opentracing/basictracer-go v1.1.0 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/pelletier/go-toml/v2 v2.0.6 // indirect
	github.com/pgavlin/fx v0.1.6 // indirect
	github.com/pjbgf/sha1cd v0.3.0 // indirect
	github.com/pkg/term v1.1.0 // indirect
	github.com/pquerna/cachecontrol v0.2.0 // indirect
	github.com/pulumi/appdash v0.0.0-20231130102222-75f619a67231 // indirect
	github.com/pulumi/esc v0.9.1 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/rogpeppe/go-internal v1.12.0 // indirect
	github.com/sabhiram/go-gitignore v0.0.0-20210923224102-525f6e181f06 // indirect
	github.com/santhosh-tekuri/jsonschema/v5 v5.0.0 // indirect
	github.com/skeema/knownhosts v1.2.2 // indirect
	github.com/spf13/afero v1.9.3 // indirect
	github.com/spf13/cast v1.5.0 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/subosito/gotenv v1.4.2 // indirect
	github.com/texttheater/golang-levenshtein v1.0.1 // indirect
	github.com/tkuchiki/go-timezone v0.2.0 // indirect
	github.com/uber/jaeger-client-go v2.30.0+incompatible // indirect
	github.com/uber/jaeger-lib v2.4.1+incompatible // indirect
	github.com/xanzy/ssh-agent v0.3.3 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20180127040702-4e3ac2762d5f // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/zclconf/go-cty v1.13.2 // indirect
	go.opentelemetry.io/otel/metric v1.31.0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	golang.org/x/mod v0.18.0 // indirect
	golang.org/x/term v0.21.0 // indirect
	golang.org/x/text v0.16.0 // indirect
	golang.org/x/tools v0.22.0 // indirect
	golang.org/x/xerrors v0.0.0-20220907171357-04be3eba64a2 // indirect
	google.golang.org/appengine v1.6.8 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240227224415-6ceb2ff114de // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240311173647-c811ad7063a7 // indirect
	google.golang.org/grpc v1.63.2 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/square/go-jose.v2 v2.6.0 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	lukechampine.com/frand v1.4.2 // indirect
)

require (
	github.com/Microsoft/go-winio v0.6.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/docker/distribution v2.8.2+incompatible // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/moby/term v0.0.0-20210619224110-3f7ff695adc6 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.2 // indirect
	github.com/pkg/errors v0.9.1
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/sirupsen/logrus v1.9.0
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/crypto v0.24.0
	golang.org/x/net v0.26.0 // indirect
	golang.org/x/sys v0.26.0 // indirect
	gotest.tools/v3 v3.3.0 // indirect
)
