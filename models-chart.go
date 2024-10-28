package lib

type ChartLayerTre struct {
	Namespace       Namespace       `json:"namespace"`
	Gateway         Gateway         `json:"gateway"`
	Service         ServiceChart    `json:"service"`
	ConfigMap       ConfigMapChart  `json:"configmap"`
	Deployment      Deployment      `json:"deployment"`
	Hpa             HpaChart        `json:"hpa"`
	VirtualService  VirtualService  `json:"virtualservice"`
	DestinationRule Destinationrule `json:"destinationrule"`
}
type Namespace struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Metadata   struct {
		Labels struct {
			IstioInjection           string `json:"istio-injection,omitempty"`
			KubernetesIoMetadataName string `json:"kubernetes.io/metadata.name,omitempty"`
		} `json:"labels,omitempty"`
		Name            string `json:"name,omitempty"`
		ResourceVersion string `json:"resourceVersion,omitempty"`
	} `json:"metadata,omitempty"`
	Spec struct {
		Finalizers []string `json:"finalizers,omitempty"`
	} `json:"spec,omitempty"`
}

// --- HPA ---
type HpaChart struct {
	APIVersion string       `json:"apiVersion,omitempty"`
	Kind       string       `json:"kind,omitempty"`
	Metadata   *MetadataHpa `json:"metadata,omitempty"`
	Spec       *SpecHpa     `json:"spec,omitempty"`
}
type SpecHpa struct {
	MaxReplicas    int                `json:"maxReplicas,omitempty"`
	Metrics        []Metrics          `json:"metrics,omitempty"`
	MinReplicas    int                `json:"minReplicas,omitempty"`
	ScaleTargetRef *ScaleTargetRefHpa `json:"scaleTargetRef,omitempty"`
}
type ScaleTargetRefHpa struct {
	APIVersion string `json:"apiVersion,omitempty"`
	Kind       string `json:"kind,omitempty"`
	Name       string `json:"name,omitempty"`
}
type Metrics struct {
	Resource struct {
		Name   string `json:"name,omitempty"`
		Target struct {
			AverageUtilization int    `json:"averageUtilization,omitempty"`
			Type               string `json:"type,omitempty"`
		} `json:"target,omitempty"`
	} `json:"resource,omitempty"`
	Type string `json:"type,omitempty"`
}
type MetadataHpa struct {
	Name            string `json:"name,omitempty"`
	Namespace       string `json:"namespace,omitempty"`
	ResourceVersion string `json:"resourceVersion,omitempty"`
}

// --- GATEWAY ---
type Gateway struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Metadata   struct {
		Name            string `json:"name,omitempty"`
		Namespace       string `json:"namespace,omitempty"`
		ResourceVersion string `json:"resourceVersion,omitempty"`
	} `json:"metadata,omitempty"`
	Spec struct {
		Selector struct {
			App string `json:"app,omitempty"`
		} `json:"selector,omitempty"`
		Servers []Server `json:"servers,omitempty"`
	} `json:"spec,omitempty"`
}
type Server struct {
	Hosts []string `json:"hosts,omitempty"`
	Port  struct {
		Name     string `json:"name,omitempty"`
		Number   int    `json:"number,omitempty"`
		Protocol string `json:"protocol,omitempty"`
	} `json:"port,omitempty"`
	TLS *TLS `json:"tls,omitempty"`
}
type TLS struct {
	HTTPSRedirect  bool   `json:"httpsRedirect,omitempty"`
	CredentialName string `json:"credentialName,omitempty"`
	Mode           string `json:"mode,omitempty"`
}

// --- SERVICE ---
type ServiceChart struct {
	APIVersion string       `json:"apiVersion,omitempty"`
	Kind       string       `json:"kind,omitempty"`
	Metadata   *MetadataSrv `json:"metadata,omitempty"`
	Spec       *SpecSvc     `json:"spec,omitempty"`
	Status     *StatusSvc   `json:"status,omitempty"`
}
type MetadataSrv struct {
	Labels          *LabelsSrv `json:"labels,omitempty"`
	Name            string     `json:"name,omitempty"`
	Namespace       string     `json:"namespace,omitempty"`
	ResourceVersion string     `json:"resourceVersion,omitempty"`
}
type LabelsSrv struct {
	App string `json:"app,omitempty"`
}
type SpecSvc struct {
	ClusterIP             string        `json:"clusterIP,omitempty"`
	ClusterIPs            []string      `json:"clusterIPs,omitempty"`
	InternalTrafficPolicy string        `json:"internalTrafficPolicy,omitempty"`
	IPFamilies            []string      `json:"ipFamilies,omitempty"`
	IPFamilyPolicy        string        `json:"ipFamilyPolicy,omitempty"`
	Ports                 []ServicePort `json:"ports,omitempty"`
	Selector              *SelectorSvc  `json:"selector,omitempty"`
	SessionAffinity       string        `json:"sessionAffinity,omitempty"`
	Type                  string        `json:"type,omitempty"`
}
type StatusSvc struct {
	LoadBalancer *LoadBalancerSvc `json:"loadBalancer,omitempty"`
}
type LoadBalancerSvc struct {
}
type ServicePort struct {
	Name       string `json:"name,omitempty"`
	Port       int32  `json:"port,omitempty"`
	Protocol   string `json:"protocol,omitempty"`
	TargetPort int32  `json:"targetPort,omitempty"`
}
type SelectorSvc struct {
	App string `json:"app,omitempty"`
}

// --- CONFIGMAP ---
type ConfigMapChart struct {
	APIVersion string       `json:"apiVersion,omitempty"`
	Data       interface{}  `json:"data,omitempty"`
	Kind       string       `json:"kind,omitempty"`
	Metadata   *MetadataCfg `json:"metadata,omitempty"`
}
type MetadataCfg struct {
	Name            string `json:"name,omitempty"`
	Namespace       string `json:"namespace,omitempty"`
	ResourceVersion string `json:"resourceVersion,omitempty"`
}

// --- DEPLOYMENT ---
type Deployment struct {
	APIVersion string       `json:"apiVersion,omitempty"`
	Kind       string       `json:"kind,omitempty"`
	Metadata   *MetadataDpl `json:"metadata,omitempty"`
	Spec       *SpecDpl     `json:"spec,omitempty"`
}
type MetadataDpl struct {
	Labels          LabelsDpl `json:"labels,omitempty"`
	Name            string    `json:"name,omitempty"`
	Namespace       string    `json:"namespace,omitempty"`
	ResourceVersion string    `json:"resourceVersion,omitempty"`
}
type LabelsDpl struct {
	App string `json:"app,omitempty"`
}
type SpecDpl struct {
	Replicas             int           `json:"replicas,omitempty"`
	RevisionHistoryLimit int           `json:"revisionHistoryLimit,omitempty"`
	Selector             *SpecSelector `json:"selector,omitempty"`
	Strategy             *SpecStrategy `json:"strategy,omitempty"`
	Template             *TemplateDpl  `json:"template,omitempty"`
}
type SpecSelector struct {
	MatchLabels map[string]string `json:"matchLabels,omitempty"`
}
type SpecStrategy struct {
	RollingUpdate RollingUpdate `json:"rollingUpdate,omitempty"`
	Type          string        `json:"type,omitempty"`
}
type RollingUpdate struct {
	MaxSurge       int `json:"maxSurge,omitempty"`
	MaxUnavailable int `json:"maxUnavailable,omitempty"`
}
type TemplateDpl struct {
	Metadata *TemplateMetadataDpl `json:"metadata,omitempty"`
	Spec     *TemplateSpecDpl     `json:"spec,omitempty"`
}
type TemplateMetadataDpl struct {
	CreationTimestamp interface{}       `json:"creationTimestamp,omitempty"`
	Labels            map[string]string `json:"labels,omitempty"`
}
type TemplateSpecDpl struct {
	Containers                    []Containers `json:"containers,omitempty"`
	DNSPolicy                     string       `json:"dnsPolicy,omitempty"`
	RestartPolicy                 string       `json:"restartPolicy,omitempty"`
	SchedulerName                 string       `json:"schedulerName,omitempty"`
	TerminationGracePeriodSeconds int          `json:"terminationGracePeriodSeconds,omitempty"`
	Volumes                       []Volumes    `json:"volumes,omitempty"`
	Affinity                      *AffinityDpl `json:"affinity,omitempty"`
}
type Containers struct {
	Env             []Env     `json:"env,omitempty"`
	Image           string    `json:"image,omitempty"`
	ImagePullPolicy string    `json:"imagePullPolicy,omitempty"`
	Name            string    `json:"name,omitempty"`
	Ports           []Ports   `json:"ports,omitempty"`
	LivenessProbe   *ProbeDpl `json:"livenessProbe,omitempty"`
	StartupProbe    *ProbeDpl `json:"startupProbe,omitempty"`
	ReadinessProbe  *ProbeDpl `json:"ReadinessProbe,omitempty"`
	Resources       struct {
		Limits struct {
			CPU    string `json:"cpu,omitempty"`
			Memory string `json:"memory,omitempty"`
		} `json:"limits,omitempty"`
		Requests struct {
			CPU    string `json:"cpu,omitempty"`
			Memory string `json:"memory,omitempty"`
		} `json:"requests,omitempty"`
	} `json:"resources,omitempty"`
	TerminationMessagePath   string         `json:"terminationMessagePath,omitempty"`
	TerminationMessagePolicy string         `json:"terminationMessagePolicy,omitempty"`
	VolumeMounts             []VolumeMounts `json:"volumeMounts,omitempty"`
}
type VolumeMounts struct {
	MountPath string  `json:"mountPath,omitempty"`
	Name      string  `json:"name,omitempty"`
	SubPath   *string `json:"subPath,omitempty"`
}
type ProbeDtlsDpl struct {
	Command     string `json:"command,omitempty"`
	Host        string `json:"host,omitempty"`
	HttpHeaders string `json:"httpHeaders,omitempty"`
	Path        string `json:"path,omitempty"`
	Port        int    `json:"port,omitempty"`
	Scheme      string `json:"scheme,omitempty"`
	Name        string `json:"name,omitempty"`
	Value       string `json:"value,omitempty"`
}
type ProbeDpl struct {
	HttpGet             *ProbeDtlsDpl `json:"httpGet,omitempty"`
	Exec                *ProbeDtlsDpl `json:"exec,omitempty"`
	TcpSocket           *ProbeDtlsDpl `json:"tcpSocket,omitempty"`
	Grpc                *ProbeDtlsDpl `json:"grpc,omitempty"`
	InitialDelaySeconds int           `json:"initialDelaySeconds,omitempty"`
	PeriodSeconds       int           `json:"periodSeconds,omitempty"`
	TimeoutSeconds      int           `json:"timeoutSeconds,omitempty"`
	SuccessThreshold    int           `json:"successThreshold,omitempty"`
	FailureThreshold    int           `json:"failureThreshold,omitempty"`
}
type Ports struct {
	ContainerPort int    `json:"containerPort,omitempty"`
	Protocol      string `json:"protocol,omitempty"`
}
type Env struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}
type Volumes struct {
	HostPath  *HostPath        `json:"hostPath,omitempty"`
	Name      string           `json:"name,omitempty"`
	ConfigMap *VolumeConfigMap `json:"configMap,omitempty"`
	Secret    *VolumeSecret    `json:"secret,omitempty"`
}
type HostPath struct {
	Path string `json:"path,omitempty"`
	Type string `json:"type,omitempty"`
}
type VolumeConfigMap struct {
	DefaultMode int    `json:"defaultMode,omitempty"`
	Name        string `json:"name,omitempty"`
}
type VolumeSecret struct {
	DefaultMode int    `json:"defaultMode,omitempty"`
	SecretName  string `json:"secretName,omitempty"`
}
type AffinityDpl struct {
	NodeAffinity *NodeAffinity `json:"nodeAffinity"`
}
type NodeAffinity struct {
	RequiredDuringSchedulingIgnoredDuringExecution *RequiredDuringSchedulingIgnoredDuringExecution `json:"requiredDuringSchedulingIgnoredDuringExecution,omitempty"`
}
type RequiredDuringSchedulingIgnoredDuringExecution struct {
	NodeSelectorTerms []*NodeSelectorTerms `json:"nodeSelectorTerms"`
}
type NodeSelectorTerms struct {
	MatchExpressions []*MatchExpressions `json:"matchExpressions"`
}
type MatchExpressions struct {
	Key      string   `json:"key"`
	Operator string   `json:"operator"`
	Values   []string `json:"values"`
}

// --- VIRTUALSERVICE ---
type VirtualService struct {
	APIVersion string      `json:"apiVersion,omitempty"`
	Kind       string      `json:"kind,omitempty"`
	Metadata   *MetadataVs `json:"metadata,omitempty"`
	Spec       *SpecVs     `json:"spec,omitempty"`
}
type SpecVs struct {
	Gateways []string `json:"gateways,omitempty"`
	Hosts    []string `json:"hosts,omitempty"`
	HTTP     []HTTP   `json:"http,omitempty"`
	TCP      []TCP    `json:"tcp,omitempty"`
}
type MetadataVs struct {
	Name            string `json:"name,omitempty"`
	Namespace       string `json:"namespace,omitempty"`
	ResourceVersion string `json:"resourceVersion,omitempty"`
}
type TCP struct {
	Match []Match `json:"match,omitempty"`
	Route []Route `json:"route,omitempty"`
}
type HTTP struct {
	CorsPolicy *CorsPolicy `json:"corsPolicy,omitempty"`
	Match      []Match     `json:"match,omitempty"`
	Retries    *Retries    `json:"retries,omitempty"`
	Route      []Route     `json:"route,omitempty"`
	Rewrite    *Rewrite    `json:"rewrite,omitempty"`
}
type Retries struct {
	Attempts int    `json:"attempts,omitempty"`
	RetryOn  string `json:"retryOn,omitempty"`
}
type CorsPolicy struct {
	AllowHeaders []string `json:"allowHeaders"`
	AllowMethods []string `json:"allowMethods"`
	AllowOrigin  []string `json:"allowOrigin"`
}
type Match struct {
	URI     *URI     `json:"uri,omitempty"`
	Port    int      `json:"port,omitempty"`
	Headers *Headers `json:"headers,omitempty"`
}
type Headers struct {
	Customer   *Customer   `json:"customer,omitempty"`
	Partner    *Partner    `json:"partner,omitempty"`
	Market     *Market     `json:"market,omitempty"`
	CanaryMode *CanaryMode `json:"canary-mode,omitempty"`
}
type CanaryMode struct {
	Exact string `json:"exact,omitempty"`
}
type Market struct {
	Exact string `json:"exact,omitempty"`
}
type Customer struct {
	Exact string `json:"exact,omitempty"`
}
type Partner struct {
	Exact string `json:"exact,omitempty"`
}
type URI struct {
	Prefix string `json:"prefix,omitempty"`
	Exact  string `json:"exact,omitempty"`
	Regex  string `json:"regex,omitempty"`
}
type Rewrite struct {
	URI       string `json:"uri,omitempty"`
	Authority string `json:"authority,omitempty"`
}
type Route struct {
	Destination Destination `json:"destination,omitempty"`
	Weight      int         `json:"weight,omitempty"`
}
type Destination struct {
	Host   string `json:"host,omitempty"`
	Port   *Port  `json:"port,omitempty"`
	Subset string `json:"subset,omitempty"`
}
type Port struct {
	Number int `json:"number,omitempty"`
}

// --- DESTINATIONRULE ---
type MetadataDr struct {
	Name            string `json:"name,omitempty"`
	Namespace       string `json:"namespace,omitempty"`
	ResourceVersion string `json:"resourceVersion,omitempty"`
}
type Destinationrule struct {
	APIVersion string      `json:"apiVersion,omitempty"`
	Kind       string      `json:"kind,omitempty"`
	Metadata   *MetadataDr `json:"metadata,omitempty"`
	Spec       *SpecDr     `json:"spec,omitempty"`
}
type SpecDr struct {
	Host    string    `json:"host,omitempty"`
	Subsets []Subsets `json:"subsets,omitempty"`
}
type Subsets struct {
	Labels Labels `json:"labels,omitempty"`
	Name   string `json:"name,omitempty"`
}
type Labels struct {
	Version string `json:"version,omitempty"`
}
