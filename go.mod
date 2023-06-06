module github.com/erda-project/erda

go 1.19

require (
	bou.ke/monkey v1.0.2
	erda.cloud/rocketmq v0.0.0-20221216094959-4f1e33965edc
	github.com/360EntSecGroup-Skylar/excelize/v2 v2.3.2
	github.com/ClickHouse/clickhouse-go/v2 v2.10.0
	github.com/DATA-DOG/go-sqlmock v1.5.0
	github.com/GoogleCloudPlatform/spark-on-k8s-operator v0.0.0-20201215015655-2e8b733f5ad0
	github.com/Masterminds/semver v1.5.0
	github.com/Shopify/sarama v1.29.1
	github.com/WeiZhang555/tabwriter v0.0.0-20200115015932-e5c45f4da38d
	github.com/ahmetb/go-linq/v3 v3.2.0
	github.com/alecthomas/assert v0.0.0-20170929043011-405dbfeb8e38
	github.com/alibabacloud-go/darabonba-openapi v0.1.9
	github.com/alibabacloud-go/darabonba-openapi/v2 v2.0.4
	github.com/alibabacloud-go/dingtalk v1.2.1
	github.com/alibabacloud-go/mse-20190531/v3 v3.0.3119
	github.com/alibabacloud-go/tea v1.1.19
	github.com/alibabacloud-go/tea-utils/v2 v2.0.1
	github.com/aliyun/alibaba-cloud-sdk-go v1.61.1600
	github.com/aliyun/aliyun-log-go-sdk v0.1.19
	github.com/aliyun/aliyun-mns-go-sdk v0.0.0-20210305050620-d1b5875bda58
	github.com/aliyun/aliyun-oss-go-sdk v2.1.4+incompatible
	github.com/andrianbdn/iospng v0.0.0-20180730113000-dccef1992541
	github.com/antlr/antlr4/runtime/Go/antlr v0.0.0-20220314183648-97c793e446ba
	github.com/antonmedv/expr v1.9.0
	github.com/apache/thrift v0.14.2
	github.com/appscode/go v0.0.0-20191119085241-0887d8ec2ecc
	github.com/asaskevich/govalidator v0.0.0-20210307081110-f21760c49a8d
	github.com/bluele/gcache v0.0.2
	github.com/bmizerany/assert v0.0.0-20160611221934-b7ed37b82869
	github.com/bugaolengdeyuxiaoer/go-ansiterm v0.0.0-20211110022506-7e621b9b6908
	github.com/buger/jsonparser v1.1.1
	github.com/buraksezer/consistent v0.9.0
	github.com/c2h5oh/datasize v0.0.0-20200112174442-28bbd4740fee
	github.com/caarlos0/env v0.0.0-20180521112546-3e0f30cbf50b
	github.com/cespare/xxhash v1.1.0
	github.com/cespare/xxhash/v2 v2.2.0
	github.com/coreos/etcd v3.3.25+incompatible
	github.com/davecgh/go-spew v1.1.1
	github.com/dlclark/regexp2 v1.4.0
	github.com/docker/docker v20.10.22+incompatible
	github.com/docker/spdystream v0.2.0
	github.com/doug-martin/goqu/v9 v9.18.0
	github.com/dustin/go-humanize v1.0.1
	github.com/elastic/cloud-on-k8s/v2 v2.0.0-20221101080201-221fa719f43b
	github.com/elazarl/goproxy v0.0.0-20200421181703-e76ad31c14f6
	github.com/erda-project/erda-infra v1.0.8
	github.com/erda-project/erda-infra/tools v0.0.0-20220922112628-1965c0260662
	github.com/erda-project/erda-oap-thirdparty-protocol v0.0.0-20210907135609-15886a136d5b
	github.com/erda-project/erda-proto-go v1.4.0
	github.com/erda-project/erda-sourcecov v0.1.0
	github.com/fatih/color v1.13.0
	github.com/fntlnz/mountinfo v0.0.0-20171106231217-40cb42681fad
	github.com/fsnotify/fsnotify v1.5.4
	github.com/getkin/kin-openapi v0.76.0
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/gin-gonic/gin v1.7.0
	github.com/go-echarts/go-echarts/v2 v2.2.4
	github.com/go-eden/routine v1.0.1
	github.com/go-errors/errors v1.0.1
	github.com/go-openapi/loads v0.19.4
	github.com/go-openapi/spec v0.19.8
	github.com/go-openapi/strfmt v0.19.5
	github.com/go-playground/validator v9.31.0+incompatible
	github.com/go-redis/redis v6.15.9+incompatible
	github.com/go-sql-driver/mysql v1.6.0
	github.com/gobwas/glob v0.2.3
	github.com/gocql/gocql v0.0.0-20210707082121-9a3953d1826d
	github.com/gofrs/flock v0.8.0
	github.com/gofrs/uuid v4.0.0+incompatible
	github.com/gogap/errors v0.0.0-20200228125012-531a6449b28c
	github.com/golang-collections/collections v0.0.0-20130729185459-604e922904d3
	github.com/golang-jwt/jwt v3.2.1+incompatible
	github.com/golang/mock v1.6.0
	github.com/golang/protobuf v1.5.2
	github.com/golang/snappy v0.0.4
	github.com/google/go-cmp v0.5.9
	github.com/google/go-jsonnet v0.18.0
	github.com/google/uuid v1.3.0
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/schema v1.1.0
	github.com/gorilla/websocket v1.4.2
	github.com/hashicorp/go-multierror v1.1.1
	github.com/hpcloud/tail v1.0.0
	github.com/influxdata/influxql v0.0.0-00010101000000-000000000000
	github.com/jinzhu/copier v0.3.2
	github.com/jinzhu/gorm v1.9.16
	github.com/jinzhu/now v1.1.5
	github.com/jmespath/go-jmespath v0.4.0
	github.com/json-iterator/go v1.1.12
	github.com/labstack/echo v3.3.10+incompatible
	github.com/labstack/gommon v0.3.0
	github.com/lestrrat/go-file-rotatelogs v0.0.0-20180223000712-d3151e2a480f
	github.com/libgit2/git2go/v33 v33.0.9
	github.com/magiconair/properties v1.8.6
	github.com/mcuadros/go-version v0.0.0-20190830083331-035f6764e8d2
	github.com/mholt/archiver v2.1.0+incompatible
	github.com/minio/md5-simd v1.1.2
	github.com/minio/minio-go v0.0.0-20190308013636-b32976861da0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/mapstructure v1.5.0
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826
	github.com/olivere/elastic v6.2.35+incompatible
	github.com/opentracing/opentracing-go v1.2.0
	github.com/otiai10/copy v1.5.0
	github.com/parnurzeal/gorequest v0.2.16
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pelletier/go-toml v1.9.5
	github.com/pingcap/parser v0.0.0-20201022083903-fbe80b0c40bb
	github.com/pingcap/tidb v1.1.0-beta.0.20200921100526-29e8c0913100
	github.com/pkg/browser v0.0.0-20210911075715-681adbf594b8
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.13.0
	github.com/prometheus/common v0.37.0
	github.com/prometheus/prometheus v2.3.2+incompatible
	github.com/pyroscope-io/client v0.6.1-0.20230130114945-a64d920d2fba
	github.com/pyroscope-io/pyroscope v0.37.2
	github.com/qri-io/starlib v0.4.2-0.20200213133954-ff2e8cd5ef8d
	github.com/rakyll/statik v0.1.7
	github.com/rancher/apiserver v0.0.0-20210922180056-297b6df8d714
	github.com/rancher/dynamiclistener v0.2.1-0.20200714201033-9c1939da3af9
	github.com/rancher/remotedialer v0.2.6-0.20220104192242-f3837f8d649a
	github.com/rancher/steve v0.0.0-20221031182508-a10fe811f58f
	github.com/rancher/wrangler v0.8.11-0.20211214201934-f5aa5d9f2e81
	github.com/recallsong/go-utils v1.1.2-0.20210826100715-fce05eefa294
	github.com/rifflock/lfshook v0.0.0-20180920164130-b9218ef580f5
	github.com/robfig/cron v1.2.0
	github.com/robfig/cron/v3 v3.0.1
	github.com/russross/blackfriday/v2 v2.1.0
	github.com/sabhiram/go-gitignore v0.0.0-20201211210132-54b8a0bf510f
	github.com/satori/go.uuid v1.2.0
	github.com/scylladb/gocqlx v1.5.0
	github.com/shirou/gopsutil v3.21.11+incompatible
	github.com/shirou/gopsutil/v3 v3.22.2
	github.com/shogo82148/androidbinary v1.0.2
	github.com/shopspring/decimal v1.3.1
	github.com/sirupsen/logrus v1.9.0
	github.com/spf13/afero v1.8.2
	github.com/spf13/cast v1.5.0
	github.com/spf13/cobra v1.7.0
	github.com/spotify/flink-on-k8s-operator v0.4.0
	github.com/stretchr/testify v1.8.2
	github.com/t-tiger/gorm-bulk-insert v1.3.0
	github.com/tealeg/xlsx v1.0.5
	github.com/tealeg/xlsx/v3 v3.3.0
	github.com/varstr/uaparser v0.0.0-20170929040706-6aabb7c4e98c
	github.com/xormplus/builder v0.0.0-20181220055446-b12ceebee76f
	github.com/xormplus/core v0.0.0-20181016121923-6bfce2eb8867
	github.com/xormplus/xorm v0.0.0-20181212020813-da46657160ff
	go.etcd.io/etcd v0.5.0-alpha.5.0.20200910180754-dd1b699fc489
	go.opentelemetry.io/otel v1.14.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.3.0
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.2.0
	go.opentelemetry.io/otel/sdk v1.13.0
	go.opentelemetry.io/otel/trace v1.14.0
	go.opentelemetry.io/proto/otlp v0.11.0
	go.uber.org/automaxprocs v1.5.1
	go.uber.org/ratelimit v0.2.0
	golang.org/x/net v0.7.0
	golang.org/x/sync v0.1.0
	golang.org/x/text v0.7.0
	golang.org/x/time v0.0.0-20220609170525-579cf78fd858
	google.golang.org/genproto v0.0.0-20220915135415-7fd63a7952de
	google.golang.org/grpc v1.48.0
	google.golang.org/protobuf v1.28.1
	gopkg.in/Knetic/govaluate.v3 v3.0.0
	gopkg.in/flosch/pongo2.v3 v3.0.0-20141028000813-5e81b817a0c4
	gopkg.in/igm/sockjs-go.v2 v2.0.0
	gopkg.in/oauth2.v3 v3.12.0
	gopkg.in/square/go-jose.v2 v2.5.1
	gopkg.in/stretchr/testify.v1 v1.2.2
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.1
	gorm.io/driver/mysql v1.3.2
	gorm.io/driver/sqlite v1.3.1
	gorm.io/gorm v1.23.5
	gorm.io/plugin/soft_delete v1.1.0
	gotest.tools v2.2.0+incompatible
	helm.sh/helm/v3 v3.6.2
	howett.net/plist v1.0.0
	istio.io/api v0.0.0-20200715212100-dbf5277541ef
	istio.io/client-go v0.0.0-20201005161859-d8818315d678
	k8s.io/api v0.25.3
	k8s.io/apiextensions-apiserver v0.25.0
	k8s.io/apimachinery v0.25.3
	k8s.io/apiserver v0.24.0
	k8s.io/autoscaler/vertical-pod-autoscaler v0.11.0
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/klog v1.0.0
	k8s.io/klog/v2 v2.80.1
	k8s.io/kube-aggregator v0.24.0
	k8s.io/kubectl v0.21.0
	k8s.io/kubernetes v1.21.0
	k8s.io/utils v0.0.0-20220728103510-ee6ede2d64ed
	modernc.org/mathutil v1.0.0
	sigs.k8s.io/controller-runtime v0.13.1
	sigs.k8s.io/sig-storage-lib-external-provisioner/v6 v6.3.0
	sigs.k8s.io/yaml v1.3.0
)

require (
	cloud.google.com/go v0.102.1 // indirect
	cloud.google.com/go/storage v1.22.1 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1 // indirect
	github.com/BurntSushi/toml v0.4.1 // indirect
	github.com/Chronokeeper/anyxml v0.0.0-20160530174208-54457d8e98c6 // indirect
	github.com/ClickHouse/ch-go v0.53.0 // indirect
	github.com/CloudyKit/fastprinter v0.0.0-20200109182630-33d98a066a53 // indirect
	github.com/CloudyKit/jet v2.1.2+incompatible // indirect
	github.com/DataDog/zstd v1.4.1 // indirect
	github.com/MakeNowJust/heredoc v0.0.0-20170808103936-bb23615498cd // indirect
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver/v3 v3.1.1 // indirect
	github.com/Masterminds/sprig/v3 v3.2.2 // indirect
	github.com/Masterminds/squirrel v1.5.0 // indirect
	github.com/Microsoft/go-winio v0.5.2 // indirect
	github.com/Microsoft/hcsshim v0.9.4 // indirect
	github.com/PuerkitoBio/purell v1.1.1 // indirect
	github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578 // indirect
	github.com/XSAM/otelsql v0.9.0 // indirect
	github.com/agrison/go-tablib v0.0.0-20160310143025-4930582c22ee // indirect
	github.com/agrison/mxj v0.0.0-20160310142625-1269f8afb3b4 // indirect
	github.com/alecthomas/colour v0.1.0 // indirect
	github.com/alecthomas/repr v0.0.0-20210301060118-828286944d6a // indirect
	github.com/alibabacloud-go/alibabacloud-gateway-spi v0.0.4 // indirect
	github.com/alibabacloud-go/debug v0.0.0-20190504072949-9472017b5c68 // indirect
	github.com/alibabacloud-go/endpoint-util v1.1.0 // indirect
	github.com/alibabacloud-go/openapi-util v0.1.0 // indirect
	github.com/alibabacloud-go/tea-utils v1.3.9 // indirect
	github.com/alibabacloud-go/tea-xml v1.1.2 // indirect
	github.com/aliyun/credentials-go v1.1.2 // indirect
	github.com/andres-erbsen/clock v0.0.0-20160526145045-9e14626cd129 // indirect
	github.com/andybalholm/brotli v1.0.5 // indirect
	github.com/armon/go-radix v1.0.0 // indirect
	github.com/baiyubin/aliyun-sts-go-sdk v0.0.0-20180326062324-cfa1a18b161f // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blang/semver v3.5.1+incompatible // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/bndr/gotabulate v1.1.2 // indirect
	github.com/brahma-adshonor/gohook v1.1.9 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/cenkalti/backoff/v4 v4.2.0 // indirect
	github.com/clbanning/mxj/v2 v2.5.5 // indirect
	github.com/confluentinc/confluent-kafka-go v1.5.2 // indirect
	github.com/containerd/cgroups v1.0.4 // indirect
	github.com/containerd/containerd v1.6.8 // indirect
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd v0.0.0-20191104093116-d3cd4ed1dbcf // indirect
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // indirect
	github.com/cyphar/filepath-securejoin v0.2.2 // indirect
	github.com/cznic/mathutil v0.0.0-20181122101859-297441e03548 // indirect
	github.com/danjacques/gofslock v0.0.0-20191023191349-0a45f885bc37 // indirect
	github.com/deislabs/oras v0.11.1 // indirect
	github.com/dgraph-io/badger/v2 v2.2007.2 // indirect
	github.com/dgraph-io/ristretto v0.1.0 // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/dgryski/go-farm v0.0.0-20190423205320-6a90982ecee2 // indirect
	github.com/docker/cli v20.10.20+incompatible // indirect
	github.com/docker/distribution v2.8.1+incompatible // indirect
	github.com/docker/docker-credential-helpers v0.7.0 // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-metrics v0.0.1 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/dsnet/compress v0.0.1 // indirect
	github.com/eapache/go-resiliency v1.2.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20180814174437-776d5712da21 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/elastic/go-licenser v0.4.0 // indirect
	github.com/elastic/go-sysinfo v1.7.1 // indirect
	github.com/elastic/go-ucfg v0.8.6 // indirect
	github.com/elastic/go-windows v1.0.1 // indirect
	github.com/envoyproxy/protoc-gen-validate v0.1.0 // indirect
	github.com/evanphx/json-patch v4.12.0+incompatible // indirect
	github.com/exponent-io/jsonpath v0.0.0-20151013193312-d6023ce2651d // indirect
	github.com/facebookgo/stack v0.0.0-20160209184415-751773369052 // indirect
	github.com/fastly/go-utils v0.0.0-20180712184237-d95a45783239 // indirect
	github.com/fatih/camelcase v1.0.0 // indirect
	github.com/fatih/structs v1.1.0 // indirect
	github.com/felixge/httpsnoop v1.0.3 // indirect
	github.com/frankban/quicktest v1.14.5 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-faster/city v1.0.1 // indirect
	github.com/go-faster/errors v0.6.1 // indirect
	github.com/go-kit/kit v0.12.0 // indirect
	github.com/go-kit/log v0.2.0 // indirect
	github.com/go-logfmt/logfmt v0.5.1 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-logr/zapr v1.2.3 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/go-openapi/analysis v0.19.5 // indirect
	github.com/go-openapi/errors v0.19.2 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.19.6 // indirect
	github.com/go-openapi/swag v0.21.1 // indirect
	github.com/go-playground/locales v0.14.0 // indirect
	github.com/go-playground/universal-translator v0.18.0 // indirect
	github.com/go-playground/validator/v10 v10.4.1 // indirect
	github.com/gogap/stack v0.0.0-20150131034635-fef68dddd4f8 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/glog v1.0.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/google/btree v1.0.1 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/googleapis/gnostic v0.5.5 // indirect
	github.com/gosuri/uitable v0.0.4 // indirect
	github.com/gregjones/httpcache v0.0.0-20190212212710-3befbb6ad0cc // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.16.0 // indirect
	github.com/hailocab/go-hostpool v0.0.0-20160125115350-e80d13ce29ed // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-uuid v1.0.2 // indirect
	github.com/hashicorp/go-version v1.6.0 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/hashicorp/yamux v0.0.0-20210826001029-26ff87cf9493 // indirect
	github.com/huandu/xstrings v1.3.1 // indirect
	github.com/igm/sockjs-go v3.0.0+incompatible // indirect
	github.com/imdario/mergo v0.3.13 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jcchavezs/porto v0.1.0 // indirect
	github.com/jcmturner/aescts/v2 v2.0.0 // indirect
	github.com/jcmturner/dnsutils/v2 v2.0.0 // indirect
	github.com/jcmturner/gofork v1.0.0 // indirect
	github.com/jcmturner/gokrb5/v8 v8.4.2 // indirect
	github.com/jcmturner/rpc/v2 v2.0.3 // indirect
	github.com/jehiah/go-strftime v0.0.0-20171201141054-1d33003b3869 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jmoiron/sqlx v1.3.1 // indirect
	github.com/joeshaw/multierror v0.0.0-20140124173710-69b34d4ec901 // indirect
	github.com/jonboulle/clockwork v0.2.2 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/jpillora/backoff v1.0.0 // indirect
	github.com/klauspost/compress v1.16.0 // indirect
	github.com/klauspost/cpuid/v2 v2.0.1 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/lann/builder v0.0.0-20180802200727-47ae307949d0 // indirect
	github.com/lann/ps v0.0.0-20150810152359-62de8c46ede0 // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/lestrrat/go-envload v0.0.0-20180220120943-6ed08b54a570 // indirect
	github.com/lestrrat/go-strftime v0.0.0-20180220042222-ba3bf9c1d042 // indirect
	github.com/lib/pq v1.10.1 // indirect
	github.com/liggitt/tabwriter v0.0.0-20181228230101-89fcab3d43de // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/mattn/go-runewidth v0.0.10 // indirect
	github.com/mattn/go-sqlite3 v2.0.1+incompatible // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/miekg/dns v1.1.43 // indirect
	github.com/mitchellh/copystructure v1.1.1 // indirect
	github.com/mitchellh/go-wordwrap v1.0.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.1 // indirect
	github.com/moby/spdystream v0.2.0 // indirect
	github.com/moby/term v0.0.0-20210619224110-3f7ff695adc6 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/monochromegane/go-gitignore v0.0.0-20200626010858-205db1a8cc00 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/mwitkow/go-conntrack v0.0.0-20190716064945-2f068394615f // indirect
	github.com/mwitkow/go-proto-validators v0.3.2 // indirect
	github.com/mxk/go-flowrate v0.0.0-20140419014527-cca7078d478f // indirect
	github.com/nwaples/rardecode v1.1.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.0-rc2 // indirect
	github.com/paulmach/orb v0.9.0 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/peterbourgon/diskv/v3 v3.0.1 // indirect
	github.com/pierrec/lz4 v2.6.0+incompatible // indirect
	github.com/pierrec/lz4/v4 v4.1.17 // indirect
	github.com/pingcap/errors v0.11.5-0.20200917111840-a15ef68f753d // indirect
	github.com/pingcap/kvproto v0.0.0-20200818080353-7aaed8998596 // indirect
	github.com/pingcap/log v0.0.0-20200828042413-fce0951f1463 // indirect
	github.com/pingcap/tipb v0.0.0-20200618092958-4fad48b4c8c3 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/procfs v0.8.0 // indirect
	github.com/pyroscope-io/godeltaprof v0.1.0 // indirect
	github.com/pyroscope-io/jfr-parser v0.6.0 // indirect
	github.com/rancher/lasso v0.0.0-20210616224652-fc3ebd901c08 // indirect
	github.com/rancher/norman v0.0.0-20210423002317-8e6ffc77a819 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475 // indirect
	github.com/recallsong/unmarshal v1.0.0 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20190728182440-6a916e37a237 // indirect
	github.com/richardlehane/mscfb v1.0.3 // indirect
	github.com/richardlehane/msoleps v1.0.1 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/rogpeppe/fastuuid v1.2.0 // indirect
	github.com/rogpeppe/go-internal v1.9.0 // indirect
	github.com/rubenv/sql-migrate v0.0.0-20200616145509-8d140a17f351 // indirect
	github.com/russross/blackfriday v1.6.0 // indirect
	github.com/santhosh-tekuri/jsonschema v1.2.4 // indirect
	github.com/scylladb/go-reflectx v1.0.1 // indirect
	github.com/segmentio/asm v1.2.0 // indirect
	github.com/segmentio/kafka-go v0.4.31 // indirect
	github.com/sergi/go-diff v1.2.0 // indirect
	github.com/shabbyrobe/xmlwriter v0.0.0-20200208144257-9fca06d00ffa // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/tebeka/strftime v0.1.5 // indirect
	github.com/tidwall/gjson v1.14.1 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	github.com/tikv/pd v1.1.0-beta.0.20200907080620-6830f5bb92a2 // indirect
	github.com/tjfoc/gmsm v1.3.2 // indirect
	github.com/tklauser/go-sysconf v0.3.10 // indirect
	github.com/tklauser/numcpus v0.4.0 // indirect
	github.com/tmc/grpc-websocket-proxy v0.0.0-20201229170055-e5319fda7802 // indirect
	github.com/uber/jaeger-client-go v2.25.0+incompatible // indirect
	github.com/uber/jaeger-lib v2.4.1+incompatible // indirect
	github.com/ugorji/go/codec v1.1.7 // indirect
	github.com/ulikunitz/xz v0.5.10 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.6.0 // indirect
	github.com/valyala/fasttemplate v1.2.1 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20180127040702-4e3ac2762d5f // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	github.com/xlab/treeprint v0.0.0-20181112141820-a009c3971eca // indirect
	github.com/xuri/efp v0.0.0-20201016154823-031c29024257 // indirect
	github.com/yusufpapurcu/wmi v1.2.2 // indirect
	go.elastic.co/apm/module/apmzap/v2 v2.1.0 // indirect
	go.elastic.co/apm/v2 v2.1.0 // indirect
	go.elastic.co/fastjson v1.1.0 // indirect
	go.mongodb.org/mongo-driver v1.11.1 // indirect
	go.opencensus.io v0.23.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.28.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.27.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/internal/retry v1.3.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.3.0 // indirect
	go.opentelemetry.io/otel/internal/metric v0.25.0 // indirect
	go.opentelemetry.io/otel/metric v0.36.0 // indirect
	go.starlark.net v0.0.0-20200306205701-8dd3e2ee1dd5 // indirect
	go.uber.org/atomic v1.10.0 // indirect
	go.uber.org/multierr v1.9.0 // indirect
	go.uber.org/zap v1.24.0 // indirect
	golang.org/x/arch v0.0.0-20190312162104-788fe5ffcd8c // indirect
	golang.org/x/crypto v0.1.0 // indirect
	golang.org/x/exp v0.0.0-20230321023759-10a507213a29 // indirect
	golang.org/x/mod v0.6.0 // indirect
	golang.org/x/oauth2 v0.1.0 // indirect
	golang.org/x/sys v0.6.0 // indirect
	golang.org/x/term v0.5.0 // indirect
	golang.org/x/tools v0.2.0 // indirect
	golang.org/x/xerrors v0.0.0-20220609144429-65e65417b02f // indirect
	gomodules.xyz/jsonpatch/v2 v2.2.0 // indirect
	google.golang.org/api v0.96.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	gopkg.in/fsnotify.v1 v1.4.7 // indirect
	gopkg.in/gorp.v1 v1.7.2 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	istio.io/gogo-genproto v0.0.0-20190930162913-45029607206a // indirect
	k8s.io/cli-runtime v0.21.2 // indirect
	k8s.io/component-base v0.25.0 // indirect
	k8s.io/kube-openapi v0.0.0-20220803162953-67bda5d908f1 // indirect
	moul.io/http2curl v1.0.0 // indirect
	rsc.io/letsencrypt v0.0.3 // indirect
	sigs.k8s.io/apiserver-network-proxy/konnectivity-client v0.0.30 // indirect
	sigs.k8s.io/cli-utils v0.16.0 // indirect
	sigs.k8s.io/kustomize/api v0.8.8 // indirect
	sigs.k8s.io/kustomize/kyaml v0.10.17 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.3 // indirect
)

replace (
	erda.cloud/rocketmq => github.com/erda-project/rocketmq-operator v0.0.0-20221222075906-f28c42d7bf23
	github.com/ClickHouse/clickhouse-go/v2 => github.com/ClickHouse/clickhouse-go/v2 v2.6.5
	github.com/containerd/containerd => github.com/containerd/containerd v1.4.4
	github.com/coreos/bbolt => go.etcd.io/bbolt v1.3.5
	github.com/docker/spdystream => github.com/docker/spdystream v0.0.0-20160310174837-449fdfce4d96
	github.com/erda-project/erda-infra => github.com/erda-project/erda-infra v0.0.0-20220928061127-9fca81f05238
	github.com/erda-project/erda-proto-go v1.4.0 => ./api/proto-go
	github.com/getkin/kin-openapi => github.com/getkin/kin-openapi v0.49.0
	github.com/go-logr/logr => github.com/go-logr/logr v0.4.0
	github.com/go-logr/zapr => github.com/go-logr/zapr v0.4.0
	github.com/google/gnostic => github.com/googleapis/gnostic v0.4.0
	github.com/influxdata/influxql => github.com/erda-project/influxql v1.1.0-ex
	github.com/olivere/elastic v6.2.35+incompatible => github.com/erda-project/elastic v0.0.1-ex
	github.com/pingcap/pd/v4 => github.com/tikv/pd v1.0.8
	github.com/pyroscope-io/pyroscope => github.com/erda-project/pyroscope v0.0.0-20230605032132-b8ff7cd367ec
	github.com/rancher/apiserver => github.com/rancher/apiserver v0.0.0-20210519053359-f943376c4b42
	github.com/rancher/remotedialer => github.com/erda-project/remotedialer v0.2.6-0.20210713103000-da03eb9e4b23
	github.com/rancher/steve => github.com/rancher/steve v0.0.0-20210520191028-52f86dce9bd4
	go.etcd.io/bbolt => github.com/coreos/bbolt v1.3.5
	go.opentelemetry.io/otel => go.opentelemetry.io/otel v1.2.0
	go.opentelemetry.io/otel/exporters/stdout => go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.2.0
	go.opentelemetry.io/otel/metric => go.opentelemetry.io/otel/metric v0.25.0
	go.opentelemetry.io/otel/sdk => go.opentelemetry.io/otel/sdk v1.2.0
	go.opentelemetry.io/proto/otlp v0.11.0 => github.com/recallsong/opentelemetry-proto-go/otlp v0.11.1-0.20211202093058-995eca7123f5
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
	k8s.io/api => k8s.io/api v0.21.2
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.21.2
	k8s.io/apimachinery => k8s.io/apimachinery v0.21.2
	k8s.io/apiserver => k8s.io/apiserver v0.21.2
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.21.2
	k8s.io/client-go => k8s.io/client-go v0.21.2
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.21.2
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.21.2
	k8s.io/code-generator => k8s.io/code-generator v0.21.2
	k8s.io/component-base => k8s.io/component-base v0.21.2
	k8s.io/component-helpers => k8s.io/component-helpers v0.21.2
	k8s.io/controller-manager => k8s.io/controller-manager v0.21.2
	k8s.io/cri-api => k8s.io/cri-api v0.21.2
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.21.2
	k8s.io/klog => k8s.io/klog v1.0.0
	k8s.io/klog/v2 => k8s.io/klog/v2 v2.9.0
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.21.2
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.21.2
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20210305001622-591a79e4bda7
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.21.2
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.21.2
	k8s.io/kubectl => k8s.io/kubectl v0.21.2
	k8s.io/kubelet => k8s.io/kubelet v0.21.2
	k8s.io/kubernetes => k8s.io/kubernetes v1.21.2
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.21.2
	k8s.io/metrics => k8s.io/metrics v0.21.2
	k8s.io/mount-utils => k8s.io/mount-utils v0.21.2
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.21.2
	sigs.k8s.io/apiserver-network-proxy/konnectivity-client => sigs.k8s.io/apiserver-network-proxy/konnectivity-client v0.0.19
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.9.2
)
