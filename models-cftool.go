package lib

type EnviroStruc struct {
	CodiceEnv string
	Namespace string
	DescEnv   string
	SwBuild   string
}
type Cluster struct {
	Nome             string `json:"nome"`
	Customer         string `json:"customer"`
	Profile          int32  `json:"profile"`
	ProfileDescr     string `json:"profileDescr"`
	Dominio          string `json:"dominio"`
	AccessToken      string `json:"accessToken"`
	Ambiente         int32  `json:"ambiente"`
	RefappCustomerID string `json:"refappCustomerID"`
}
type App struct {
	AppID               string
	RefAppID            string
	NomeApp             string
	PodName             string
	Docker              string
	Monolith            string
	Resource            string
	RefAppCustomerID    string
	RefAppCustomerToken string
}
type ActionDocker struct {
	Name        string
	Value       string
	Description string
}
type ActionDb struct {
	Name        string
	Value       string
	Description string
}
type JiraIssue struct {
	Key    string `json:"key"`
	Fields struct {
		Summary string `json:"summary"`
	} `json:"fields"`
}
type LoginRes struct {
	Err                     string       `json:"err"`
	Desc                    string       `json:"desc"`
	Token                   string       `json:"token"`
	Login                   string       `json:"login"`
	MarketDec               int64        `json:"marketdec"`
	Team                    string       `json:"team"`
	Stage                   string       `json:"stage"`
	Nome                    string       `json:"nome"`
	Email                   string       `json:"email"`
	CurrentSprintBranchMs   SprintBranch `json:"currentSprintBranchMs"`
	CurrentSprintBranchMono SprintBranch `json:"currentSprintBranchMono"`
	Profile                 string       `json:"profile"`
	Version                 string       `json:"version"`
	Gruppo                  string       `json:"gruppo"`
	ClusterDomain           string       `json:"clusterDomain"`
	GkeProjectID            string       `json:"gkeProjectID"`
	Environment             string       `json:"environment"`
	RefappCustomerID        string       `json:"refappCustomerID"`
}
type SprintBranch struct {
	CurrentBranch string
	Tipo          string
	Data          string
}
type ActionProtocol struct {
	Name        string
	Value       string
	Description string
}
type ActionBitbucket struct {
	Name        string
	Value       string
	Description string
}
type VersionModel struct {
	Tipo    string `mapstructure:"tipo"`
	Version string `mapstructure:"version"`
	Attivo  string `mapstructure:"attivo"`
	Detail  string `mapstructure:"buildDetail"`
	Enviro  string `mapstructure:"enviro"`
}
type RepoListStruct struct {
	Nome  string
	Repo  string
	SwGit string
}
