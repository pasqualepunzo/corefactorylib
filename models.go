package lib

type Microservice struct {
	Nome             string `json:"nome"`
	Descrizione      string `json:"descrizione"`
	Namespace        string `json:"namespace"`
	VersMicroservice string `json:"versMs"`
	Virtualservice   string `json:"virtualService"`
	Replicas         string `json:"replicas"`
	Hpa              Hpa
	Pod              []Pod
}

type Pod struct {
	Id         string
	Docker     string
	GitRepo    string
	Descr      string
	Dockerfile string
	Tipo       string
	Vpn        int
	Workdir    string
	Mount      []Mount
	Service    []Service
	Branch     Branch
	Resource   Resource
	PodBuild   PodBuild
}

type PodBuild struct {
	Versione     string
	Merged       string
	Tag          string
	MasterDev    string
	ReleaseNote  string
	SprintBranch string
}
type Resource struct {
	CpuReq string
	CpuLim string
	MemReq string
	MemLim string
}

type Hpa struct {
	MinReplicas   string
	MaxReplicas   string
	CpuTipoTarget string
	CpuTarget     string
	MemTipoTarget string
	MemTarget     string
}

type Branch struct {
	Branch  string
	Version string
	Sha     string
}
type Mount struct {
	Nome      string
	Mount     string
	Subpath   string
	ClaimName string
}

type Service struct {
	Tipo     string
	Port     string
	Endpoint []Endpoint
}

type Endpoint struct {
	MicroserviceSrc, DockerSrc, TypeSrvSrc, RouteSrc, RewriteSrc, NamespaceSrc, VersionSrc string
	MicroserviceDst, DockerDst, TypeSrvDst, RouteDst, RewriteDst, NamespaceDst, VersionDst string
	Domain, Market, Partner, Customer                                                      string
}

type CallGetResponse struct {
	Kind      string
	BodyJson  map[string]interface{}
	BodyArray []map[string]interface{}
	Errore    int32
	Log       string
}

type LoggaErrore struct {
	Errore int32
	Log    string
}

type IresRequest struct {
	Istanza          string `json:"istanza"`
	AppID            string `json:"appID"`
	RefAppID         string `json:"refAppID"`
	CustomerID       string `json:"customerID"`
	Profile          int32  `json:"profile"`
	Monolith         string `json:"monolith"`
	Tags             string `json:"tags"`
	ProfileDeploy    string `json:"profileDeploy"`
	PodName          string `json:"podName"`
	RefAppCustomerID string `json:"refAppCustomerID"`
	CustomerDomain   string `json:"customerDomain"`
	Enviro           string `json:"enviro"`
}

type HealthPod struct {
	Nome    string
	Version string
	Status  string
}

type KillemallStruct struct {
	ClusterContext   string `json:"clusterContext"`
	DeploymentToKill string `json:"deploymentToKill"`
	Namespace        string `json:"namespace"`
}
type IstanzaMicro struct {
	Istanza                             string `json:"istanza"`
	Cluster                             string `json:"cluster"`
	Customer                            string `json:"customer"`
	Microservice                        string `json:"microservice"`
	ProjectID                           string `json:"projectID"`
	Owner                               string `json:"owner"`
	Profile                             string `json:"profile"`
	TipoDeploy                          string `json:"tipoDeploy"`
	Version                             string `json:"version"`
	CustomerSalt                        string `json:"customSalt"`
	MonolithDomain                      string `json:"monolithDomain"`
	NomeApp                             string `json:"nomeApp"`
	RefappID                            string `json:"refappID"`
	PodName                             string `json:"podName"`
	ClusterDomain                       string `json:"clusterDomain"`
	Token                               string `json:"token"`
	ClusterRefAppID                     string `json:"clusterRefAppID"`
	RefappCustomerID                    string `json:"refappCustomerID"`
	Ambiente                            int32  `json:"ambiente"`
	Tags                                []string
	AmbID                               int `json:"ambID"`
	Monolith, ProfileDeploy, ProfileInt int32
	DbMetaConnMs                        []DbMetaConnMs `json:"dbMetaConnMs"`
	DbDataConnMs                        []DbDataConnMs `json:"dbDataConnMs"`
	MasterHost                          string         `json:"masterHost"`
	MasterName                          string         `json:"masterName"`
	MasterHostData                      string         `json:"masterHostData"`
	MasterHostMeta                      string         `json:"masterHostMeta"`
	MasterUser                          string         `json:"masterUser"`
	MasterPass                          string         `json:"masterPass"`
	SwMultiEnvironment                  string         `json:"swMultiEnvironment"`
	SwCore                              bool
	SwDb                                int
	IstanzaMicroVersioni                []IstanzaMicroVersioni `json:"istanzaMicroVersioni"`
	AttributiMS                         []AttributiMS          `json:"attributiMS"`
	//Clusters                            map[string]clusterSt   `json:"clusters"`
}

type AttributiMS struct {
	Partner  string `json:"partner"`
	Market   string `json:"market"`
	Customer string `json:"customer"`
}

type DbMetaConnMs struct {
	MetaHost     string `json:"metaHost"`
	MetaName     string `json:"metaName"`
	MetaUser     string `json:"metaUser"`
	MetaPass     string `json:"metaPass"`
	MetaMicroAmb string `json:"metaMicroAmb"`
	Cluster      string `json:"cluster"`
}
type DbDataConnMs struct {
	DataHost     string `json:"dataHost"`
	DataName     string `json:"sataName"`
	DataUser     string `json:"sataUser"`
	DataPass     string `json:"sataPass"`
	DataMicroAmb string `json:"dataMicroAmb"`
	Cluster      string `json:"cluster"`
}
type IstanzaMicroVersioni struct {
	TipoVersione string `json:"tipoVersione"`
	Versione     string `json:"versione"`
}

type ClusterSt struct {
	ProjectID          string `json:"projectID"`
	Owner              string `json:"owner"`
	Profile            string `json:"profile"`
	ProfileInt         int32  `json:"profileInt"`
	Domain             string `json:"domain"`
	Token              string `json:"token"`
	MasterHost         string `json:"masterHost"`
	MasterUser         string `json:"masterUser"`
	MasterPasswd       string `json:"masterPasswd"`
	Ambiente           int32  `json:"ambiente"`
	RefappID           string `json:"refappID"`
	SwMultiEnvironment string `json:"swMultiEnvironment"`
	AccessToken        string `json:"accssToken"`
}
