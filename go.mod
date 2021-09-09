module github.com/erda-project/erda

go 1.16

require (
	bou.ke/monkey v1.0.2
	github.com/360EntSecGroup-Skylar/excelize/v2 v2.3.2
	github.com/DATA-DOG/go-sqlmock v1.5.0
	github.com/GoogleCloudPlatform/spark-on-k8s-operator v0.0.0-20201215015655-2e8b733f5ad0
	github.com/Masterminds/semver v1.5.0
	github.com/alecthomas/assert v0.0.0-20170929043011-405dbfeb8e38
	github.com/alecthomas/colour v0.1.0 // indirect
	github.com/alecthomas/repr v0.0.0-20210301060118-828286944d6a // indirect
	github.com/aliyun/alibaba-cloud-sdk-go v1.61.426
	github.com/aliyun/aliyun-log-go-sdk v0.1.19
	github.com/aliyun/aliyun-mns-go-sdk v0.0.0-20210305050620-d1b5875bda58
	github.com/aliyun/aliyun-oss-go-sdk v2.1.4+incompatible
	github.com/andrianbdn/iospng v0.0.0-20180730113000-dccef1992541
	github.com/appscode/go v0.0.0-20191119085241-0887d8ec2ecc
	github.com/baiyubin/aliyun-sts-go-sdk v0.0.0-20180326062324-cfa1a18b161f // indirect
	github.com/bluele/gcache v0.0.2
	github.com/bmizerany/assert v0.0.0-20160611221934-b7ed37b82869
	github.com/buger/jsonparser v1.1.1
	github.com/c2h5oh/datasize v0.0.0-20200112174442-28bbd4740fee
	github.com/caarlos0/env v0.0.0-20180521112546-3e0f30cbf50b
	github.com/cespare/xxhash v1.1.0
	github.com/confluentinc/confluent-kafka-go v1.5.2
	github.com/coreos/etcd v3.3.25+incompatible
	github.com/cznic/mathutil v0.0.0-20181122101859-297441e03548
	github.com/davecgh/go-spew v1.1.1
	github.com/dsnet/compress v0.0.1 // indirect
	github.com/dustin/go-humanize v1.0.0
	github.com/elastic/cloud-on-k8s v0.0.0-20210205172912-5ce0eca90c60
	github.com/erda-project/erda-infra v0.0.0-20210907124129-709cc0069f2f
	github.com/erda-project/erda-proto-go v0.0.0-20210909085352-edd80839d294
	github.com/extrame/ole2 v0.0.0-20160812065207-d69429661ad7 // indirect
	github.com/extrame/xls v0.0.1
	github.com/facebookgo/stack v0.0.0-20160209184415-751773369052 // indirect
	github.com/fastly/go-utils v0.0.0-20180712184237-d95a45783239 // indirect
	github.com/fatih/color v1.10.0
	github.com/fsnotify/fsnotify v1.4.9
	github.com/getkin/kin-openapi v0.49.0
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/gin-gonic/gin v1.7.0
	github.com/go-openapi/loads v0.19.4
	github.com/go-openapi/spec v0.19.8
	github.com/go-openapi/strfmt v0.19.5
	github.com/go-playground/validator v9.31.0+incompatible
	github.com/go-redis/redis v6.15.9+incompatible
	github.com/go-sql-driver/mysql v1.5.0
	github.com/gocql/gocql v0.0.0-20210707082121-9a3953d1826d
	github.com/gofrs/flock v0.8.0
	github.com/gofrs/uuid v4.0.0+incompatible
	github.com/gogap/errors v0.0.0-20200228125012-531a6449b28c
	github.com/gogap/stack v0.0.0-20150131034635-fef68dddd4f8 // indirect
	github.com/golang-collections/collections v0.0.0-20130729185459-604e922904d3
	github.com/golang-jwt/jwt v3.2.1+incompatible
	github.com/golang/mock v1.5.0
	github.com/google/uuid v1.2.0
	github.com/googlecloudplatform/flink-operator v0.0.0-00010101000000-000000000000
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/schema v1.1.0
	github.com/gorilla/websocket v1.4.2
	github.com/hashicorp/go-multierror v1.1.0
	github.com/hpcloud/tail v1.0.0
	github.com/igm/sockjs-go v3.0.0+incompatible // indirect
	github.com/influxdata/influxql v0.0.0-00010101000000-000000000000
	github.com/jehiah/go-strftime v0.0.0-20171201141054-1d33003b3869 // indirect
	github.com/jinzhu/gorm v1.9.16
	github.com/jinzhu/now v1.1.2
	github.com/jmespath/go-jmespath v0.4.0
	github.com/json-iterator/go v1.1.11
	github.com/labstack/echo v3.3.10+incompatible
	github.com/labstack/gommon v0.3.0
	github.com/lestrrat/go-envload v0.0.0-20180220120943-6ed08b54a570 // indirect
	github.com/lestrrat/go-file-rotatelogs v0.0.0-20180223000712-d3151e2a480f
	github.com/lestrrat/go-strftime v0.0.0-20180220042222-ba3bf9c1d042 // indirect
	github.com/libgit2/git2go/v30 v30.0.5
	github.com/magiconair/properties v1.8.5
	github.com/mcuadros/go-version v0.0.0-20190830083331-035f6764e8d2
	github.com/mholt/archiver v2.1.0+incompatible
	github.com/minio/md5-simd v1.1.2
	github.com/minio/minio-go v0.0.0-20190308013636-b32976861da0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/mapstructure v1.4.1
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826
	github.com/mwitkow/go-proto-validators v0.3.2
	github.com/nwaples/rardecode v1.1.0 // indirect
	github.com/olivere/elastic v6.2.35+incompatible
	github.com/otiai10/copy v1.5.0
	github.com/parnurzeal/gorequest v0.2.16
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pelletier/go-toml v1.9.3
	github.com/pingcap/errors v0.11.5-0.20200917111840-a15ef68f753d
	github.com/pingcap/parser v0.0.0-20201022083903-fbe80b0c40bb
	github.com/pingcap/tidb v1.1.0-beta.0.20200921100526-29e8c0913100
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.11.0
	github.com/rakyll/statik v0.1.7
	github.com/rancher/apiserver v0.0.0-20210519053359-f943376c4b42
	github.com/rancher/dynamiclistener v0.2.1-0.20200714201033-9c1939da3af9
	github.com/rancher/remotedialer v0.2.6-0.20210318171128-d1ebd5202be4
	github.com/rancher/steve v0.0.0-20210520191028-52f86dce9bd4
	github.com/rancher/wrangler v0.8.1-0.20210423003607-f71a90542852
	github.com/recallsong/go-utils v1.1.2-0.20210826100715-fce05eefa294
	github.com/rifflock/lfshook v0.0.0-20180920164130-b9218ef580f5
	github.com/robfig/cron v1.2.0
	github.com/russross/blackfriday/v2 v2.0.1
	github.com/sabhiram/go-gitignore v0.0.0-20201211210132-54b8a0bf510f
	github.com/satori/go.uuid v1.2.0
	github.com/scylladb/gocqlx v1.5.0
	github.com/shirou/gopsutil/v3 v3.21.3
	github.com/shogo82148/androidbinary v1.0.2
	github.com/shopspring/decimal v1.2.0
	github.com/sirupsen/logrus v1.8.1
	github.com/sony/sonyflake v1.0.0
	github.com/spf13/cast v1.3.1
	github.com/spf13/cobra v1.2.1
	github.com/stretchr/testify v1.7.0
	github.com/t-tiger/gorm-bulk-insert v1.3.0
	github.com/tealeg/xlsx v1.0.5
	github.com/tealeg/xlsx/v3 v3.2.4-0.20210615062226-d5ce25722f69
	github.com/tebeka/strftime v0.1.5 // indirect
	github.com/ulikunitz/xz v0.5.10 // indirect
	github.com/varstr/uaparser v0.0.0-20170929040706-6aabb7c4e98c
	github.com/xormplus/builder v0.0.0-20181220055446-b12ceebee76f
	github.com/xormplus/core v0.0.0-20181016121923-6bfce2eb8867
	github.com/xormplus/xorm v0.0.0-20181212020813-da46657160ff
	go.uber.org/atomic v1.8.0 // indirect
	go.uber.org/multierr v1.7.0 // indirect
	go.uber.org/ratelimit v0.2.0
	golang.org/x/net v0.0.0-20210903162142-ad29c8ab022f
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/sys v0.0.0-20210906170528-6f6e22806c34 // indirect
	golang.org/x/text v0.3.7
	google.golang.org/genproto v0.0.0-20210903162649-d08c68adba83
	google.golang.org/grpc v1.40.0
	google.golang.org/protobuf v1.27.1
	gopkg.in/Knetic/govaluate.v3 v3.0.0
	gopkg.in/flosch/pongo2.v3 v3.0.0-20141028000813-5e81b817a0c4
	gopkg.in/igm/sockjs-go.v2 v2.0.0
	gopkg.in/oauth2.v3 v3.12.0
	gopkg.in/stretchr/testify.v1 v1.2.2
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	gorm.io/driver/mysql v1.0.5
	gorm.io/gorm v1.21.8
	gotest.tools v2.2.0+incompatible
	helm.sh/helm/v3 v3.6.2
	howett.net/plist v0.0.0-20201203080718-1454fab16a06
	istio.io/api v0.0.0-20200715212100-dbf5277541ef
	istio.io/client-go v0.0.0-20201005161859-d8818315d678
	k8s.io/api v0.21.2
	k8s.io/apiextensions-apiserver v0.21.2
	k8s.io/apimachinery v0.21.2
	k8s.io/apiserver v0.21.2
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/kubectl v0.21.0
	k8s.io/kubernetes v1.21.2
	kmodules.xyz/monitoring-agent-api v0.0.0-20200125202117-d3b3e33ce41f
	kmodules.xyz/objectstore-api v0.0.0-20200214040336-fe8f39a4210d
	kmodules.xyz/offshoot-api v0.0.0-20200216080509-45ee6418d1c1
	modernc.org/mathutil v1.0.0
	moul.io/http2curl v1.0.0 // indirect
	rsc.io/letsencrypt v0.0.3 // indirect
	sigs.k8s.io/controller-runtime v0.9.2
	sigs.k8s.io/sig-storage-lib-external-provisioner/v6 v6.3.0
	sigs.k8s.io/yaml v1.2.0
)

replace (
	github.com/google/gnostic => github.com/googleapis/gnostic v0.4.0
	github.com/googlecloudplatform/flink-operator => github.com/erda-project/flink-on-k8s-operator v0.0.0-20210828094530-28e003581cf2
	github.com/influxdata/influxql => github.com/erda-project/influxql v1.1.0-ex
	github.com/olivere/elastic v6.2.35+incompatible => github.com/erda-project/elastic v0.0.1-ex
	github.com/rancher/remotedialer => github.com/erda-project/remotedialer v0.2.6-0.20210713103000-da03eb9e4b23
	go.etcd.io/bbolt v1.3.5 => github.com/coreos/bbolt v1.3.5
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
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.21.2
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.21.2
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.21.2
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.21.2
	k8s.io/kubectl => k8s.io/kubectl v0.21.2
	k8s.io/kubelet => k8s.io/kubelet v0.21.2
	k8s.io/kubernetes => k8s.io/kubernetes v1.21.2
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.21.2
	k8s.io/metrics => k8s.io/metrics v0.21.2
	k8s.io/mount-utils => k8s.io/mount-utils v0.21.2
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.21.2
)
