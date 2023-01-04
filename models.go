package lib

import "time"

type Repos struct {
	Repo string
	Nome string
	Sw   int
	Tipo string
}
type loginAuth struct {
	username, password string
}
type BranchResStruct struct {
	Name   string `json:"name"`
	Target struct {
		Hash string `json:"hash"`
	} `json:"target"`
}

type TagCreateResStruct struct {
	Type  string `json:"type"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}
type ProfileInfo struct {
	Code    int `json:"code"`
	Session struct {
		GrantSession struct {
			Gruppo      string `json:"gruppo"`
			NomeCognome string `json:"nome_cognome"`
			Email       string `json:"email"`
		} `json:"grant_session"`
		Market struct {
			Binval  string `json:"binval"`
			Decval  int    `json:"decval"`
			Details struct {
				Book        string `json:"book"`
				Corefactory string `json:"corefactory"`
				Corporate   string `json:"corporate"`
				Fashion     string `json:"fashion"`
				Food        string `json:"food"`
			} `json:"details"`
		} `json:"market"`
	} `json:"session"`
}
type MasterConn struct {
	Host        string `json:"host"`
	MetaName    string `json:"metaName"`
	User        string `json:"user"`
	Pass        string `json:"pass"`
	Domain      string `json:"domain"`
	AccessToken string `json:"accssToken"`
	Cluster     string `json:"cluster"`
}
type CompareDbRes struct {
	Tbl         string
	Column_name string
	Columns     string
	Tipo        string
}

type DeployDbLog struct {
	Log    string
	Errore int32
}

type CompareIndex struct {
	Tbl         string
	NomeIdx     string
	Index       string
	NomeColonna string
	Unique      string
}
type MergeResponse struct {
	Log   string
	Error string
}
type CreateBranchResponse struct {
	Type  string `json:"type"`
	Error struct {
		Message string `json:"message"`
		Data    struct {
			Key string `json:"key"`
		} `json:"data"`
	} `json:"error"`
}
type RestyResStruct struct {
	Type  string `json:"type"`
	ID    int    `json:"id"`
	Links struct {
		Diffstat struct {
			Href string `json:"href"`
		} `json:"diffstat"`
		Commits struct {
			Href string `json:"href"`
		} `json:"commits"`
		Comments struct {
			Href string `json:"href"`
		} `json:"comments"`
		Merge struct {
			Href string `json:"href"`
		} `json:"merge"`
		Diff struct {
			Href string `json:"href"`
		} `json:"diff"`
	} `json:"links"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}
type RestyResConflict struct {
	Pagelen int `json:"pagelen"`
	Values  []struct {
		Status string `json:"status"`
		Old    struct {
			Path        string `json:"path"`
			EscapedPath string `json:"escaped_path"`
			Type        string `json:"type"`
			Links       struct {
				Self struct {
					Href string `json:"href"`
				} `json:"self"`
			} `json:"links"`
		} `json:"old"`
		LinesRemoved int `json:"lines_removed"`
		LinesAdded   int `json:"lines_added"`
		New          struct {
			Path        string `json:"path"`
			EscapedPath string `json:"escaped_path"`
			Type        string `json:"type"`
			Links       struct {
				Self struct {
					Href string `json:"href"`
				} `json:"self"`
			} `json:"links"`
		} `json:"new"`
		Type string `json:"type"`
	} `json:"values"`
	Page int `json:"page"`
	Size int `json:"size"`
}
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
	Probes     []Probes
}
type Probes struct {
	Category            string
	Type                string
	Command             string
	HttpHost            string
	HttpPort            int
	HttpPath            string
	HttpHeaders         string
	HttpScheme          string
	TcpHost             string
	TcpPort             int
	GrpcPort            int
	InitialDelaySeconds int
	PeriodSeconds       int
	TimeoutSeconds      int
	SuccessThreshold    int
	FailureThreshold    int
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
	Enviro                              string `json:"enviro"`
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
	ApiHost                             string                 `json:"apiHost"`
	ApiToken                            string                 `json:"apiToken"`
	Autopilot                           string                 `json:"autopilot"`
	CloudNet                            string                 `json:"cloudNet"`
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
	ApiHost            string `json:"apiHost"`
	ApiToken           string `json:"apiToken"`
	CloudNet           string `json:"cloudNet"`
	Autopilot          string `json:"autopilot"`
}
type ClusterAccess struct {
	Domain            string `json="domain"`
	AccessToken       string `json="accessToken"`
	ReffappCustomerID string `json="reffappCustomerID"`
}
type DeploymntStatus struct {
	Items []struct {
		Metadata struct {
			Name      string `json:"name"`
			Namespace string `json:"namespace"`
		} `json:"metadata,omitempty"`
		Spec struct {
			Replicas int `json:"replicas"`
			Selector struct {
				MatchLabels struct {
					App     string `json:"app"`
					Version string `json:"version"`
				} `json:"matchLabels"`
			} `json:"selector"`
			Template struct {
				Spec struct {
					Containers []struct {
						Env []struct {
							Name  string `json:"name"`
							Value string `json:"value"`
						} `json:"env"`
						EnvFrom []struct {
							ConfigMapRef struct {
								Name string `json:"name"`
							} `json:"configMapRef"`
						} `json:"envFrom"`
						Image           string `json:"image"`
						ImagePullPolicy string `json:"imagePullPolicy"`
						Name            string `json:"name"`
						Ports           []struct {
							ContainerPort int    `json:"containerPort"`
							Protocol      string `json:"protocol"`
						} `json:"ports"`
						Resources struct {
							Limits struct {
								CPU    string `json:"cpu"`
								Memory string `json:"memory"`
							} `json:"limits"`
							Requests struct {
								CPU    string `json:"cpu"`
								Memory string `json:"memory"`
							} `json:"requests"`
						} `json:"resources"`
						TerminationMessagePath   string `json:"terminationMessagePath"`
						TerminationMessagePolicy string `json:"terminationMessagePolicy"`
						VolumeMounts             []struct {
							MountPath string `json:"mountPath"`
							Name      string `json:"name"`
							SubPath   string `json:"subPath,omitempty"`
						} `json:"volumeMounts"`
					} `json:"containers"`
					Volumes []struct {
						HostPath struct {
							Path string `json:"path"`
							Type string `json:"type"`
						} `json:"hostPath,omitempty"`
						Name                  string `json:"name"`
						PersistentVolumeClaim struct {
							ClaimName string `json:"claimName"`
						} `json:"persistentVolumeClaim,omitempty"`
						ConfigMap struct {
							DefaultMode int    `json:"defaultMode"`
							Name        string `json:"name"`
						} `json:"configMap,omitempty"`
					} `json:"volumes"`
				} `json:"spec"`
			} `json:"template"`
		} `json:"spec"`
		Status struct {
			AvailableReplicas int `json:"availableReplicas"`
			Conditions        []struct {
				LastTransitionTime time.Time `json:"lastTransitionTime"`
				LastUpdateTime     time.Time `json:"lastUpdateTime"`
				Message            string    `json:"message"`
				Reason             string    `json:"reason"`
				Status             string    `json:"status"`
				Type               string    `json:"type"`
			} `json:"conditions"`
			ObservedGeneration int `json:"observedGeneration"`
			ReadyReplicas      int `json:"readyReplicas"`
			Replicas           int `json:"replicas"`
			UpdatedReplicas    int `json:"updatedReplicas"`
		} `json:"status,omitempty"`
	} `json:"items"`
}
type HttpHeadersJson struct {
	Name  string
	Value string
}
