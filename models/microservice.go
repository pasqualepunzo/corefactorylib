package models

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
	cpuReq string
	cpuLim string
	memReq string
	memLim string
}

type Hpa struct {
	minReplicas   string
	maxReplicas   string
	cpuTipoTarget string
	cpuTarget     string
	memTipoTarget string
	memTarget     string
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
