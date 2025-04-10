package lib

type Microservice struct {
	Nome             string       `json:"nome,omitempty"`
	Descrizione      string       `json:"descrizione,omitempty"`
	Namespace        string       `json:"namespace,omitempty"`
	BuildVersione    string       `json:"buildVersione,omitempty"`
	Virtualservice   string       `json:"virtualService,omitempty"`
	Public           int          `json:"public,omitempty"`
	DatabasebEnable  string       `json:"databasebEnable,omitempty"`
	Hpa              Hpa          `json:"hpa,omitempty"`
	Pod              []Pod        `json:"pod,omitempty"`
	Istanza          IstanzaMicro `json:"istanza,omitempty"`
	Affinity         []Affinity   `json:"affinity,omitempty"`
	SwCore           bool         `json:"swCore,omitempty"`
	ScaleToZero      bool         `json:"scaleToZero,omitempty"`
	IsApp            bool         `json:"isApp,omitempty"`
	SwDb             int          `json:"swDb,omitempty"`
	NomeApp          string       `json:"nomeApp,omitempty"`
	RefappID         string       `json:"refappID,omitempty"`
	RefappCustomerID string       `json:"refappCustomerID,omitempty"`
	RefAppCode       string       `json:"refAppCode,omitempty"`
}
type Hpa struct {
	MinReplicas   string `json:"min,omitempty"`
	MaxReplicas   string `json:"max,omitempty"`
	CpuTipoTarget string `json:"cpuTipoTarget,omitempty"`
	CpuTarget     string `json:"cpuTarget,omitempty"`
	MemTipoTarget string `json:"memTipoTarget,omitempty"`
	MemTarget     string `json:"memTarget,omitempty"`
}
type Pod struct {
	Id         string      `json:"id,omitempty"`
	Docker     string      `json:"docker,omitempty"`
	GitRepo    string      `json:"gitRepo,omitempty"`
	Descr      string      `json:"descr,omitempty"`
	Dockerfile string      `json:"dockerfile,omitempty"`
	Tipo       string      `json:"tipo,omitempty"`
	Vpn        int         `json:"vpn,omitempty"`
	Workdir    string      `json:"workdir,omitempty"`
	Branch     Branch      `json:"branch,omitempty"`
	Resource   Resource    `json:"resource,omitempty"`
	PodBuild   *PodBuild   `json:"podBuild,omitempty"`
	Mount      []Mount     `json:"mount,omitempty"`
	Service    []Service   `json:"service,omitempty"`
	Probes     []Probes    `json:"probes,omitempty"`
	ConfigMap  []ConfigMap `json:"configMap,omitempty"`
}
type Affinity struct {
	Key      string   `json:"key,omitempty"`
	Operator string   `json:"operator,omitempty"`
	Values   []string `json:"values,omitempty"`
}
type PodBuild struct {
	Versione     string
	Merged       string
	Tag          string
	MasterDev    string
	ReleaseNote  string
	SprintBranch string
	Sha          string
}
type Mount struct {
	Nome       string `json:"nome,omitempty"`
	Mount      string `json:"mount,omitempty"`
	Subpath    string `json:"subpath,omitempty"`
	ClaimName  string `json:"claimName,omitempty"`
	FromSecret bool   `json:"fromSecret,omitempty"`
}
type ConfigMap struct {
	ConfigType string `json:"configType,omitempty"`
	MountType  string `json:"mountType,omitempty"`
	MountPath  string `json:"mountPath,omitempty"`
	Name       string `json:"name,omitempty"`
	Content    string `json:"content,omitempty"`
	Env        string `json:"env,omitempty"`
	Cluster    string `json:"cluster,omitempty"`
}
type Service struct {
	Tipo       string     `json:"tipo,omitempty"`
	Port       string     `json:"port,omitempty"`
	Versione   string     `json:"versione,omitempty"`
	TipoDeploy string     `json:"tipoDeploy,omitempty"`
	Endpoint   []Endpoint `json:"endpoint,omitempty"`
}
type Probes struct {
	Category            string `json:"category,omitempty"`
	Type                string `json:"type,omitempty"`
	Command             string `json:"command,omitempty"`
	HttpHost            string `json:"httpHost,omitempty"`
	HttpPort            int    `json:"httpPort,omitempty"`
	HttpPath            string `json:"httpPath,omitempty"`
	HttpHeaders         string `json:"httpHeaders,omitempty"`
	HttpScheme          string `json:"httpScheme,omitempty"`
	TcpHost             string `json:"tcpHost,omitempty"`
	TcpPort             int    `json:"tcpPort,omitempty"`
	GrpcPort            int    `json:"grpcPort,omitempty"`
	InitialDelaySeconds int    `json:"initialDelaySeconds,omitempty"`
	PeriodSeconds       int    `json:"periodSeconds,omitempty"`
	TimeoutSeconds      int    `json:"timeoutSeconds,omitempty"`
	SuccessThreshold    int    `json:"successThreshold,omitempty"`
	FailureThreshold    int    `json:"failureThreshold,omitempty"`
}
type Endpoint struct {
	MicroserviceSrc string   `json:"microserviceSrc,omitempty"`
	DockerSrc       string   `json:"dockerSrc,omitempty"`
	TypeSrvSrc      string   `json:"typeSrvSrc,omitempty"`
	RouteSrc        string   `json:"routeSrc,omitempty"`
	RewriteSrc      string   `json:"rewriteSrc,omitempty"`
	NamespaceSrc    string   `json:"namespaceSrc,omitempty"`
	VersionSrc      string   `json:"versionSrc,omitempty"`
	MicroserviceDst string   `json:"microserviceDst,omitempty"`
	DockerDst       string   `json:"dockerDst,omitempty"`
	TypeSrvDst      string   `json:"typeSrvDst,omitempty"`
	RouteDst        string   `json:"routeDst,omitempty"`
	RewriteDst      string   `json:"rewriteDst,omitempty"`
	NamespaceDst    string   `json:"namespaceDst,omitempty"`
	VersionDst      string   `json:"versionDst,omitempty"`
	Domain          string   `json:"domain,omitempty"`
	Market          string   `json:"market,omitempty"`
	Partner         string   `json:"partner,omitempty"`
	Customer        string   `json:"customer,omitempty"`
	ClusterDomain   string   `json:"clusterDomain,omitempty"`
	Priority        string   `json:"priority,omitempty"`
	AllowedMethod   string   `json:"allowedMethod,omitempty"`
	AllowHeaders    []string `json:"allowHeaders,omitempty"`
}
type RouteMs struct {
	Microservice string         `json:"microservice,omitempty"`
	Istanza      string         `json:"istanza,omitempty"`
	Docker       []RouteDocker  `json:"docker,omitempty"`
	Version      []RouteVersion `json:"version,omitempty"`
}
type RouteDocker struct {
	Docker  string    `json:"docker,omitempty"`
	Service []Service `json:"service,omitempty"`
}
type RouteVersion struct {
	CanaryProduction string `json:"canaryProduction,omitempty"`
	Versione         string `json:"versione,omitempty"`
}

type IstanzaMicro struct {
	Istanza            string          `json:"istanza,omitempty"`
	Cluster            string          `json:"cluster,omitempty"`
	Customer           string          `json:"customer,omitempty"`
	Microservice       string          `json:"microservice,omitempty"`
	ProjectID          string          `json:"projectID,omitempty"`
	Owner              string          `json:"owner,omitempty"`
	Profile            string          `json:"profile,omitempty"`
	Enviro             string          `json:"enviro,omitempty"`
	TipoDeploy         string          `json:"tipoDeploy,omitempty"`
	Version            []RouteVersion  `json:"version,omitempty"`
	CustomerSalt       string          `json:"customSalt,omitempty"`
	PodName            string          `json:"podName,omitempty"`
	ClusterDomain      string          `json:"clusterDomain,omitempty"`
	ClusterDomainOvr   bool            `json:"clusterDomainOvr,omitempty"`
	ClusterDomainEnv   string          `json:"clusterDomainEnv,omitempty"`
	ClusterDomainProd  string          `json:"clusterDomainProd,omitempty"`
	ClusterDomainStage string          `json:"clusterDomainStage,omitempty"`
	Token              string          `json:"token,omitempty"`
	ClusterRefAppID    string          `json:"clusterRefAppID,omitempty"`
	ClusterExtIP       string          `json:"clusterExtIP,omitempty"`
	Ambiente           int32           `json:"ambiente,omitempty"`
	Tags               []string        `json:"tags,omitempty"`
	AmbID              int             `json:"ambID,omitempty"`
	Monolith           int32           `json:"monolith,omitempty"`
	ProfileDeploy      int32           `json:"profileDeploy,omitempty"`
	ProfileInt         int32           `json:"profileInt,omitempty"`
	DbMetaConnMs       []DbMetaConnMs  `json:"dbMetaConnMs,omitempty"`
	DbDataConnMs       []DbDataConnMs  `json:"dbDataConnMs,omitempty"`
	MasterHost         string          `json:"masterHost,omitempty"`
	MasterName         string          `json:"masterName,omitempty"`
	MasterHostData     string          `json:"masterHostData,omitempty"`
	MasterHostMeta     string          `json:"masterHostMeta,omitempty"`
	MasterUser         string          `json:"masterUser,omitempty"`
	MasterPass         string          `json:"masterPass,omitempty"`
	SwMultiEnvironment string          `json:"swMultiEnvironment,omitempty"`
	AttributiMS        []AttributiMS   `json:"attributiMS,omitempty"`
	ApiHost            string          `json:"apiHost,omitempty"`
	ApiToken           string          `json:"apiToken,omitempty"`
	Autopilot          string          `json:"autopilot,omitempty"`
	CloudNet           string          `json:"cloudNet,omitempty"`
	DeploymentEnv      []DeploymentEnv `json:"deploymentEnv,omitempty"`
	LayerDue           *LayerMesh      `json:"layerDue,omitempty"`
	LayerTre           *LayerMesh      `json:"layerTre,omitempty"`
}
type DeploymentEnv struct {
	Env           string `json:"env,omitempty"`
	DeploymentEnv string `json:"deploymentEnv,omitempty"`
}
type DbMetaConnMs struct {
	MetaHost     string `json:"metaHost,omitempty"`
	MetaName     string `json:"metaName,omitempty"`
	MetaUser     string `json:"metaUser,omitempty"`
	MetaPass     string `json:"metaPass,omitempty"`
	MetaMicroAmb string `json:"metaMicroAmb,omitempty"`
	Cluster      string `json:"cluster,omitempty"`
}
type DbDataConnMs struct {
	DataHost     string `json:"dataHost,omitempty"`
	DataName     string `json:"dataName,omitempty"`
	DataUser     string `json:"dataUser,omitempty"`
	DataPass     string `json:"dataPass,omitempty"`
	DataMicroAmb string `json:"dataMicroAmb,omitempty"`
	Cluster      string `json:"cluster,omitempty"`
}
type IstanzaMicroVersioni struct {
	Microservice string `json:"microservice,omitempty"`
	TipoVersione string `json:"tipoVersione,omitempty"`
	Versione     string `json:"versione,omitempty"`
}
type AttributiMS struct {
	Partner  string `json:"partner,omitempty"`
	Market   string `json:"market,omitempty"`
	Customer string `json:"customer,omitempty"`
}
type LayerMesh struct {
	AppName string `json:"AppName"`
	Gw      []Gw   `json:"gw"`
	Vs      Vs     `json:"vs"`
	Se      Se     `json:"se"`
}
type Gw struct {
	ExtDominio []string `json:"extDominio"`
	IntDominio []string `json:"intDominio"`
	Name       string   `json:"name"`
	Number     string   `json:"number"`
	Protocol   string   `json:"protocol"`
}
type Se struct {
	Ip    string   `json:"ip"`
	Hosts []string `json:"hosts"`
}
type Vs struct {
	ExternalHost []string    `json:"externalHost"`
	InternalHost []string    `json:"internalHost"`
	VsDetails    []VsDetails `json:"vsDetails"`
}
type VsDetails struct {
	DestinationHost string `json:"destinationHost"`
	Prefix          string `json:"prefix"`
	Authority       string `json:"authority"`
}
