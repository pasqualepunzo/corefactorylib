package lib

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/manifoldco/promptui"
)

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
