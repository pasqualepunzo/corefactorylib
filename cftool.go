package lib

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/manifoldco/promptui"
)

// questa e una murzetta per rendere i flag usabili le package lib
var fflag map[string]interface{}

func SetFlags(flags map[string]interface{}) {
	fflag = flags
}

func ChooseTipo() int32 {

	type branchOptionSt struct {
		Name        string
		Value       int32
		Description string
	}

	branchOption := []branchOptionSt{
		{Name: "Sprint Branch", Value: 0, Description: "Sprint Branch"},
		{Name: "Hotfix", Value: 1, Description: "Hotfix"},
	}

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "* {{ .Name | cyan }} ",
		Inactive: "  {{ .Name | cyan }} ",
		Selected: `* {{ "Profile:" | faint }}             {{ .Name | green  }}`,
		Details: `
		--------- Options ----------
		{{ "Name:" | faint }}	{{ .Name | yellow}}
		{{ "Description:" | faint }}	{{ .Description | yellow}}
		`,
	}

	prompt := promptui.Select{
		Label:     "Choose between branch or hotfix",
		Items:     branchOption,
		Templates: templates,
	}

	numtipoBuild, _, _ := prompt.Run()
	tipoBuildRes := branchOption[numtipoBuild].Value

	return tipoBuildRes
}
func CleanLogin(login string) string {
	userLogin := strings.Replace(login, "@", "", -1)
	userLogin = strings.Replace(userLogin, ".", "", -1)
	return userLogin
}
func ChooseDeploymentType(swBoth bool) string {

	type canProd struct {
		Tipo string
	}

	canProds := []canProd{}
	if swBoth == true {
		canProds = []canProd{
			{Tipo: "Canary"},
			{Tipo: "Production"},
			//{Tipo: "Canary-Production"},
		}
	} else {
		canProds = []canProd{
			{Tipo: "Canary"},
			{Tipo: "Production"},
		}
	}

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "* {{ .Tipo | cyan }} ",
		Inactive: "  {{ .Tipo | cyan }} ",
		Selected: `* {{ "Type:" | faint }}               {{ .Tipo | green  }}`,
		Details: `
		--------- Type ----------
		{{ "Type:" | faint }}	{{ .Tipo | yellow }}
		`,
	}

	promptCanary := promptui.Select{
		Label:     "Select kind of deploy",
		Items:     canProds,
		Templates: templates,
	}

	resultCanaryNum, _, _ := promptCanary.Run()

	return strings.ToLower(canProds[resultCanaryNum].Tipo)
}
func ChooseActionOnDocker() string {

	actionDockers := []ActionDocker{
		//{Name: "ALL", Value: 0, Binary: "11111"},
		{Name: "Continue job", Value: "continue", Description: "Continue to work the issue"},
		{Name: "Init job", Value: "init", Description: "Drop and create git repos and local directories"},
		{Name: "Hotfix", Value: "hotfix", Description: "Start and work on a hotfix"},
	}

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "* {{ .Name | cyan }} ",
		Inactive: "  {{ .Name | cyan }} ",
		Selected: `* {{ "Action Docker:" | faint }}             {{ .Name | green  }}`,
		Details: `
		--------- Action Docker ----------
		{{ "Name:" | faint }}	{{ .Name | yellow}}
		{{ "Description:" | faint }}	{{ .Description | yellow}}
		`,
	}

	prompt := promptui.Select{
		Label:     "Choose Action for Docker",
		Items:     actionDockers,
		Templates: templates,
	}

	numResultActionDocker, _, _ := prompt.Run()
	actionDockerRes := actionDockers[numResultActionDocker].Value
	return actionDockerRes
}
func ChooseActionOnDb() string {

	actionDbs := []ActionDb{
		//{Name: "ALL", Value: 0, Binary: "11111"},
		{Name: "Skip", Value: "skip", Description: "If database exists, skip creation"},
		{Name: "Build", Value: "build", Description: "Build or delete and rebuild database"},
	}

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "* {{ .Name | cyan }} ",
		Inactive: "  {{ .Name | cyan }} ",
		Selected: `* {{ "Action for Database:" | faint }}             {{ .Name | green  }}`,
		Details: `
		--------- Action Database ----------
		{{ "Name:" | faint }}	{{ .Name | yellow}}
		{{ "Description:" | faint }}	{{ .Description | yellow}}
		`,
	}

	prompt := promptui.Select{
		Label:     "Choose Action for Database",
		Items:     actionDbs,
		Templates: templates,
	}

	numResultActionDb, _, _ := prompt.Run()
	actionDbRes := actionDbs[numResultActionDb].Value
	return actionDbRes
}
func CheckIssue(resultIssue, userLogin, atlassianHost, atlassianUser, atlassianToken string, loginRes LoginRes) string {

	// cerco la issue su jira per vedere se esiste
	client := resty.New()
	client.Debug = false
	resp, _ := client.R().
		EnableTrace().
		SetHeader("Accept", "application/json").
		SetBasicAuth(atlassianUser, atlassianToken).
		Get(atlassianHost + "/rest/api/latest/issue/" + resultIssue)

	var issue JiraIssue
	err := json.Unmarshal(resp.Body(), &issue)
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println("Bye")
		fmt.Println()
		os.Exit(0)
	}

	re, err := regexp.Compile(`[^\w]`)
	if err != nil {
		fmt.Println(err.Error())
	}
	s := re.ReplaceAllString(issue.Fields.Summary, "-")

	s = issue.Key + "-" + userLogin + "-" + s

	if issue.Key == "" {
		fmt.Println()
		fmt.Println("The issue does not exists")
		fmt.Println("Bye")
		fmt.Println()

		os.Exit(0)
	}

	return s
}
func ChooseProtocol() string {

	actionProtocol := []ActionProtocol{
		//{Name: "ALL", Value: 0, Binary: "11111"},
		{Name: "Http", Value: "http", Description: "Work on a HTTP cluster"},
		{Name: "Https", Value: "https", Description: "Work on a HTTPS cluster"},
	}

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "* {{ .Name | cyan }} ",
		Inactive: "  {{ .Name | cyan }} ",
		Selected: `* {{ "Action Docker:" | faint }}             {{ .Name | green  }}`,
		Details: `
		--------- Action Docker ----------
		{{ "Name:" | faint }}	{{ .Name | yellow}}
		{{ "Description:" | faint }}	{{ .Description | yellow}}
		`,
	}

	prompt := promptui.Select{
		Label:     "Choose Protocol",
		Items:     actionProtocol,
		Templates: templates,
	}

	numResultActionProtocol, _, _ := prompt.Run()
	actionProtocolRes := actionProtocol[numResultActionProtocol].Value
	return actionProtocolRes
}
func ChooseActionBitbucket() string {
	actionBitbuckets := []ActionBitbucket{
		{Name: "Ssh", Value: "ssh", Description: "Use your ssh credentials to connect to Bitbucket"},
		{Name: "User and password", Value: "build", Description: "Use Login and Password to connect to Bitbucket"},
	}

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "* {{ .Name | cyan }} ",
		Inactive: "  {{ .Name | cyan }} ",
		Selected: `* {{ "Action for Bitbucket:" | faint }}             {{ .Name | green  }}`,
		Details: `
		--------- Action Bitbucket ----------
		{{ "Name:" | faint }}	{{ .Name | yellow}}
		{{ "Description:" | faint }}	{{ .Description | yellow}}
		`,
	}

	prompt := promptui.Select{
		Label:     "Choose Action for Bitbucket",
		Items:     actionBitbuckets,
		Templates: templates,
	}

	numResultActionBitbucket, _, _ := prompt.Run()
	actionBitbucketRes := actionBitbuckets[numResultActionBitbucket].Value
	return actionBitbucketRes
}
func ChooseDevMaster() string {
	type DevMaster struct {
		Name  string
		Value string
		Descr string
	}
	devMaster := []DevMaster{
		{Name: "Already Promoted", Value: "master", Descr: "Deploy a Microservice Already promoted to Master"},
		{Name: "Not Promoted yet", Value: "dev", Descr: "Deploy a Microservice Still in Development Mode"},
	}

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   " {{ .Name | cyan }} ",
		Inactive: "  {{ .Name | cyan }} ",
		Selected: `* {{ "Master/Dev:" | faint }}         {{ .Name | green  }}`,
		Details: `
		--------- Master/Dev ----------
		{{ "Name:" | faint }}	{{ .Name | yellow}}
		{{ "Descr:" | faint }}	{{ .Descr | yellow}}
		`,
	}

	prompt := promptui.Select{
		Label:     "Choose git branch Master or Development (sprint branch) ",
		Items:     devMaster,
		Templates: templates,
	}

	numResultDevMaster, _, _ := prompt.Run()
	DevMasterRes := devMaster[numResultDevMaster].Value
	return DevMasterRes
}
func GetNomeIstanza(cluster, enviro, microservice, team string) string {
	return cluster + "-" + enviro + "-" + strings.ToLower(team) + "-" + microservice
}
func CheckExecutable() string {

	var errore string = ""

	fmt.Print("Checking VSCode ")
	_, err := exec.LookPath("code")
	if err != nil {
		errore = "The component \"VSCode\" is missing\n"
	} else {
		fmt.Println("OK")
	}

	fmt.Print("Checking Git ")
	_, err2 := exec.LookPath("git")
	if err2 != nil {
		errore = "The component \"git\" is missing\n"
	} else {
		fmt.Println("OK")
	}

	fmt.Print("Checking Curl ")
	_, err = exec.LookPath("curl")
	if err != nil {
		errore = "The component \"curl\" is missing\n"
	} else {
		fmt.Println("OK")
	}
	return errore
}

/* ***************************** */
// Graphix
const angleUpSx = "*"
const angleUpDx = "*"
const angleDwSx = "*"
const angleDwDx = "*"
const grpxLine = "*"
const grpxCol = "*"

var Reset = "\033[0m"
var Red = "\033[31m"
var Green = "\033[32m"
var Yellow = "\033[33m"
var Blue = "\033[34m"
var Purple = "\033[35m"
var Cyan = "\033[36m"
var Gray = "\033[37m"
var White = "\033[97m"

var taskDone = Green + "OK" + Reset
var taskError = Red + "KO" + Reset

func Header(nome, descr, cfToolVers, swMono string) {
	if runtime.GOOS == "windows" {
		Reset = ""
		Red = ""
		Green = ""
		Yellow = ""
		Blue = ""
		Purple = ""
		Cyan = ""
		Gray = ""
		White = ""
	}

	cftool := "cftool"
	y := 1 + 19 + 7 + 3 + len(nome) + len(cfToolVers) + 3 + 1
	x := 81 - y
	if swMono == "mono" {
		cftool = "cftoolmono"
		x = 72 - y
	}

	// 10 spazi iniziali
	// 7 cf-tool
	// 3 -
	// len nome

	str1 := "*" + WriteGraphixEmptySpaces(19) + Yellow + cftool + Reset + " > " + cfToolVers + " - " + Blue + nome + Reset + WriteGraphixEmptySpaces(x) + "*"

	y = 1 + 19 + len(descr) + 1
	x = 80 - y
	str2 := "*" + WriteGraphixEmptySpaces(19) + White + descr + Reset + WriteGraphixEmptySpaces(x) + "*"

	WriteGraphixTop()
	WriteGraphixEmptyLine()
	fmt.Println(str1)
	WriteGraphixEmptyLine()
	fmt.Println(str2)
	WriteGraphixEmptyLine()

	WriteGraphixBottom()
}
func WriteGraphixCol() {
	fmt.Print(grpxCol)
}
func WriteGraphixEmptyLine() {
	fmt.Print(grpxCol)
	for i := 0; i < 78; i++ {
		fmt.Print(" ")
	}
	fmt.Print(grpxCol)
	fmt.Println()
}
func WriteFinalLine() {
	fmt.Print(grpxCol)
	for i := 0; i < 51; i++ {
		fmt.Print(" ")
	}
	fmt.Print(grpxCol)
	for i := 0; i < 26; i++ {
		fmt.Print(" ")
	}
	fmt.Print(grpxCol)

	fmt.Println()
}
func WriteGraphixEmptySpaces(num int) string {
	str := ""
	for i := 0; i < num; i++ {
		str += " "
	}
	return str
}
func WriteGraphixTop() {

	fmt.Print(angleUpSx)
	for i := 0; i < 78; i++ {
		fmt.Print(grpxLine)
	}
	fmt.Print(angleUpDx)
	fmt.Println()
}
func WriteGraphixBottom() {

	fmt.Print(angleDwSx)
	for i := 0; i < 78; i++ {
		fmt.Print(grpxLine)
	}
	fmt.Print(angleDwDx)
	fmt.Println()
}
func Footer(loginres LoginRes, fflag map[string]interface{}) {
	//LogJson(loginres)

	if runtime.GOOS == "windows" {
		Reset = ""
		Red = ""
		Green = ""
		Yellow = ""
		Blue = ""
		Purple = ""
		Cyan = ""
		Gray = ""
		White = ""
	}
	var spMs, spMono string

	if loginres.CurrentSprintBranchMs.CurrentBranch != "" {
		spMs = loginres.CurrentSprintBranchMs.CurrentBranch
	}
	if loginres.CurrentSprintBranchMono.CurrentBranch != "" {
		spMono = loginres.CurrentSprintBranchMono.CurrentBranch
	}

	canaryProduction := ""
	if fflag["Canary"] == false {
		canaryProduction = "Prod"
	} else {
		canaryProduction = "Canary"
	}

	line1 := "*" + WriteGraphixEmptySpaces(1) + "User " + Yellow + loginres.Nome + Reset + WriteGraphixEmptySpaces(52-(2+5+len(loginres.Nome)))
	line1 += "*" + WriteGraphixEmptySpaces(1) + "Version " + Yellow + loginres.Version + "/" + canaryProduction + Reset + WriteGraphixEmptySpaces(28-(2+8+8+len(canaryProduction)+2)) + "*"
	fmt.Println(line1)

	line2 := "*" + WriteGraphixEmptySpaces(1) + "Sprint " + Yellow + spMs + Reset + WriteGraphixEmptySpaces(52-(2+7+len(spMs)))
	line2 += "*" + WriteGraphixEmptySpaces(1) + "Cluster " + Yellow + loginres.Profile + Reset + WriteGraphixEmptySpaces(28-(2+8+3+1)) + "*"
	fmt.Println(line2)

	line3 := "*" + WriteGraphixEmptySpaces(1) + "Sprint Mono " + Yellow + spMono + Reset + WriteGraphixEmptySpaces(52-(2+12+len(spMono)))
	line3 += "*" + WriteGraphixEmptySpaces(1) + "Environment " + Yellow + loginres.Environment + Reset + WriteGraphixEmptySpaces(52-(2+12+len(loginres.Environment)))
	fmt.Println(line3)
}

// Graphix
/* ***************************** */
