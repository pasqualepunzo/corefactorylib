package lib

import (
	"bytes"
	"context"
	"crypto/sha1"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	//"log"
	"math"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
)

func Logga(ctx context.Context, jsonLog string, i interface{}, level ...string) {

	JobID := ""
	if ctx != nil {
		if ctx.Value("JobID") != nil {
			JobID = ctx.Value("JobID").(string)
		}
	}
	caller := ""

	_, file, line, ok := runtime.Caller(1)
	if ok {
		caller = file + ":" + strconv.Itoa(line)
	}

	text := ""
	switch v := i.(type) {
	case int:
		// v is an int here, so e.g. v + 1 is possible.
		fmt.Printf("Integer: %v", v)
	case float64:
		// v is a float64 here, so e.g. v + 1.0 is possible.
		fmt.Printf("Float64: %v", v)
	case string:
		text = i.(string)
	default:
		var b []byte
		b, _ = json.MarshalIndent(i, "", "\t")
		text = string(b)
		//fmt.Println(text)
	}

	if jsonLog == "false" {
		fmt.Println(text)
	} else {

		log.SetFormatter(&log.JSONFormatter{})
		log.SetOutput(os.Stdout)

		if len(level) > 0 {
			switch level[0] {
			case "info":
				log.WithFields(log.Fields{
					"JobID":  JobID,
					"pid":    os.Getpid(),
					"caller": caller,
				}).Info(text)

			case "error":
				log.WithFields(log.Fields{
					"JobID":  JobID,
					"pid":    os.Getpid(),
					"caller": caller,
				}).Error(text)

			case "warn":
				log.WithFields(log.Fields{
					"JobID":  JobID,
					"pid":    os.Getpid(),
					"caller": caller,
				}).Warn(text)
			}
		} else {
			log.WithFields(log.Fields{
				"JobID":  JobID,
				"pid":    os.Getpid(),
				"caller": caller,
			}).Info(text)
		}
	}
}
func Base64decode(str string) string {

	if str != "" {
		decoded, err := base64.StdEncoding.DecodeString(str)
		if err != nil {
			fmt.Println("decode error:", err)
		}
		return string(decoded)
	} else {
		return ""
	}
}

/*
	text += "\033[1;31m *********************************************************************************\n"
	text += " *  _____\n"
	text += " * | ____|_ __ _ __ ___  _ __ \n"
	text += " * |  _| | '__| '__/ _ \\| '__|\n"
	text += " * | |___| |  | | | (_) | |   \n"
	text += " * |_____|_|  |_|  \\___/|_|  \n"
	text += " * \033[1;0m\n"
	text += " *\n"
	text += " * " + errorLabel

per i comandi che inviano in streamig al cf-tool del testo
la convenzione è: se la stringa inizia con ### vuol dire che va printato qualcosa se dopo ### abbiamo Error è un errore
altrimenti e una comunicazione
*/
func PrintaErroreStream(errorLabel, log string, flagDie bool) {
	fmt.Println("_##ERROR##_", errorLabel, " - ", log)
	if flagDie {
		os.Exit(1)
	}
}
func PrintaErroreStreamText(errorLabel, log string) string {
	text := "_##START##_\n"
	text += "\n" + "\n"
	text += " *\n"
	text += " *\n"
	text += " * " + log + "\n"
	text += " *\n"
	text += " *\n"
	text += "\033[0;31m *********************************************************************************\033[1;0m\n"
	text += "\n"
	text += "_##STOP##_\n"

	return text
}

// SwitchCluster ...
func SwitchCluster(clusterName, cloudNet string) {
	comando := "gcloud container clusters get-credentials " + clusterName + cloudNet
	ExecCommand(comando, true)
}

// SwitchProject ...
func SwitchProject(clusterProject string) {

	comando := "gcloud config set project  " + clusterProject + " --quiet "
	ExecCommand(comando, true)
}
func ExecCommand(command string, printOutput bool) bool {

	println()
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", command)
	} else {
		cmd = exec.Command("bash", "-c", command)
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	stdoutIn, _ := cmd.StdoutPipe()
	stderrIn, _ := cmd.StderrPipe()

	var errStdout, errStderr error
	stdout := io.MultiWriter(os.Stdout, &stdoutBuf)
	stderr := io.MultiWriter(os.Stderr, &stderrBuf)
	err := cmd.Start()
	if err != nil {
		log.Fatalf("cmd.Start() failed with '%s'\n", err)
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		_, errStdout = io.Copy(stdout, stdoutIn)
		wg.Done()
	}()

	_, errStderr = io.Copy(stderr, stderrIn)
	wg.Wait()

	err = cmd.Wait()
	FataleErrore := false
	if err != nil {
		if printOutput == true {
			PrintaErroreStream("Exec command errors: "+command+" -> ", err.Error(), true)
			PrintaErrore("Exec command errors: "+command+" -> ", err.Error(), "fix errors and try again")
		}
		FataleErrore = true
	}
	if errStdout != nil || errStderr != nil {
		if printOutput == true {
			log.Fatal("failed to capture stdout or stderr\n")
		}
	}
	outStr := string(stdoutBuf.Bytes())
	if printOutput == true {
		fmt.Printf("%s", outStr)
	}
	return FataleErrore
}
func PrintaErrore(errorLabel, log, errorSuggest string) {
	fmt.Println()
	fmt.Println("*!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!*")
	fmt.Println(" _____")
	fmt.Println("| ____|_ __ _ __ ___  _ __ ")
	fmt.Println("|  _| | '__| '__/ _ \\| '__|")
	fmt.Println("| |___| |  | | | (_) | |   ")
	fmt.Println("|_____|_|  |_|  \\___/|_|  ")
	fmt.Println("*                                                                                *")
	fmt.Println("*  " + errorLabel)
	fmt.Println("*                                                                                *")
	fmt.Println("*  " + log)
	fmt.Println("*                                                                                *")
	fmt.Println("*                                                                                *")
	fmt.Println("*  " + errorSuggest)
	fmt.Println("*                                                                                *")
	fmt.Println("*!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!*")
	fmt.Println()
}
func LogJson(msg interface{}) {

	//MarshalIndent
	empJSON, _ := json.MarshalIndent(msg, "", "  ")
	fmt.Printf("%s\n", string(empJSON))
}
func StrPad(input string, padLength int, padString string, padType string) string {
	var output string

	inputLength := len(input)
	padStringLength := len(padString)

	if inputLength >= padLength {
		return input
	}

	repeat := math.Ceil(float64(1) + (float64(padLength-padStringLength))/float64(padStringLength))

	switch padType {
	case "RIGHT":
		output = input + strings.Repeat(padString, int(repeat))
		output = output[:padLength]
	case "LEFT":
		output = strings.Repeat(padString, int(repeat)) + input
		output = output[len(output)-padLength:]
	case "BOTH":
		length := (float64(padLength - inputLength)) / float64(2)
		repeat = math.Ceil(length / float64(padStringLength))
		output = strings.Repeat(padString, int(repeat))[:int(math.Floor(float64(length)))] + input + strings.Repeat(padString, int(repeat))[:int(math.Ceil(float64(length)))]
	}

	return output
}
func GetUserGroup(ctx context.Context, token, gruppo, dominio, coreApiVersion string) (map[string]string, error) {

	Logga(ctx, os.Getenv("JsonLog"), "", "Getting GRU +")

	var erro error

	args := make(map[string]string)
	args["center_dett"] = "dettaglio"
	args["source"] = "users-3"
	args["$select"] = "XGRU05,XGRU06"

	gruRes, _ := ApiCallGET(ctx, os.Getenv("RestyDebug"), args, "msusers", "/api/"+os.Getenv("coreApiVersion")+"/users/GRU/equals(XGRU03,'"+gruppo+"')", token, dominio, coreApiVersion)
	gru := make(map[string]string)

	if len(gruRes.BodyJson) > 0 {
		gru["gruppo"] = gruRes.BodyJson["XGRU05"].(string)
		gru["stage"] = gruRes.BodyJson["XGRU06"].(string)
		Logga(ctx, os.Getenv("JsonLog"), "GRU OK")
	} else {
		Logga(ctx, os.Getenv("JsonLog"), "GRU MISSING")
		erro = errors.New("GRU MISSING")
	}

	return gru, erro
}
func GetNextVersion(ctx context.Context, masterBranch, nomeDocker, tenant, accessToken, loginApiDomain, coreApiVersion, resource, devopsToken string) (string, error) {

	var erro error

	ct := time.Now()
	dateVers := ct.Format("060102")
	ver := ""
	/* ************************************************************************************************ */
	// KUBEIMICROSERV
	Logga(ctx, os.Getenv("JsonLog"), "Getting KUBEDKRBUILD - func.go 1")
	argsImicro := make(map[string]string)
	argsImicro["source"] = "devops-8"
	argsImicro["$select"] = "XKUBEDKRBUILD06"
	argsImicro["center_dett"] = "dettaglio"
	argsImicro["$filter"] = "startwith(XKUBEDKRBUILD06,'" + dateVers + "') "
	argsImicro["$filter"] += " and equals(XKUBEDKRBUILD03,'" + nomeDocker + "') "

	restyKubeImicroservRes, errImicroservRes := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsImicro, "msdevops", "/api/"+os.Getenv("coreApiVersion")+"/devops/KUBEDKRBUILD", devopsToken, loginApiDomain, coreApiVersion)
	if errImicroservRes != nil {
		Logga(ctx, os.Getenv("JsonLog"), errImicroservRes.Error())
		return "", errImicroservRes
	}

	if len(restyKubeImicroservRes.BodyJson) > 0 {
		ver = restyKubeImicroservRes.BodyJson["XKUBEDKRBUILD06"].(string)
		Logga(ctx, os.Getenv("JsonLog"), "KUBEDKRBUILD OK")
	} else {
		Logga(ctx, os.Getenv("JsonLog"), "KUBEDKRBUILD MISSING")
	}
	Logga(ctx, os.Getenv("JsonLog"), "")

	if ver == "" {
		ver = dateVers + "01"
	} else {

		prefix := string(ver[0:6])
		lastDigit := string(ver[6:8])

		i, _ := strconv.Atoi(lastDigit)
		i++

		var num string
		if i < 10 {
			num = "0" + strconv.Itoa(i)
		} else {
			num = strconv.Itoa(i)
		}

		ver = prefix + num
	}

	return ver, erro
}
func Times(str string, n int) string {
	if n <= 0 {
		return ""
	}
	return strings.Repeat(str, n)
}

// func GetMicroserviceDetail(ctx context.Context, team, ims, gitDevMaster, buildVersion, devopsToken, autopilot, enviro, dominio, coreApiVersion string) (Microservice, error) {

// 	Logga(ctx, os.Getenv("JsonLog"), "")
// 	Logga(ctx, os.Getenv("JsonLog"), " + + + + + + + + + + + + + + + + + + + + ")
// 	Logga(ctx, os.Getenv("JsonLog"), "TEAM "+team)
// 	Logga(ctx, os.Getenv("JsonLog"), "IMS "+ims)
// 	Logga(ctx, os.Getenv("JsonLog"), "ENVIRO "+enviro)
// 	Logga(ctx, os.Getenv("JsonLog"), "gitDevMaster "+gitDevMaster)
// 	Logga(ctx, os.Getenv("JsonLog"), "BUILDVERSION "+buildVersion)
// 	Logga(ctx, os.Getenv("JsonLog"), "getMicroserviceDetail begin")

// 	var erro error

// 	devops := "devops"
// 	if strings.Contains(ims, "p2rpowerna-monolith") {
// 		devops = "devopsmono"
// 	}

// 	versioneArr := strings.Split(buildVersion, ".")
// 	versione := ""

// 	if len(versioneArr) > 1 {
// 		versione = versioneArr[0] + Times("0", 2-len(versioneArr[1])) + versioneArr[1] + Times("0", 2-len(versioneArr[2])) + versioneArr[2] + Times("0", 2-len(versioneArr[3])) + versioneArr[3]
// 	} else {
// 		versione = buildVersion
// 	}

// 	var microservices Microservice

// 	/* ************************************************************************************************ */
// 	// KUBEIMICROSERV
// 	Logga(ctx, os.Getenv("JsonLog"), "Getting KUBEIMICROSERV")
// 	argsImicro := make(map[string]string)
// 	argsImicro["source"] = "devops-8"
// 	argsImicro["$select"] = "XKUBEIMICROSERV04,XKUBEIMICROSERV05,XKUBEIMICROSERV07"
// 	argsImicro["center_dett"] = "dettaglio"
// 	argsImicro["$filter"] = "equals(XKUBEIMICROSERV03,'" + ims + "') "

// 	restyKubeImicroservRes, errKubeImicroservRes := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsImicro, "ms"+devops, "/api/"+os.Getenv("coreApiVersion")+"/"+devops+"/KUBEIMICROSERV", devopsToken, dominio, coreApiVersion)
// 	if errKubeImicroservRes != nil {
// 		Logga(ctx, os.Getenv("JsonLog"), errKubeImicroservRes.Error())
// 		return microservices, errKubeImicroservRes
// 	}

// 	microservice := ""
// 	cluster := ""
// 	if len(restyKubeImicroservRes.BodyJson) > 0 {

// 		microservice = restyKubeImicroservRes.BodyJson["XKUBEIMICROSERV04_COD"].(string)
// 		microservices.BuildVersione = restyKubeImicroservRes.BodyJson["XKUBEIMICROSERV07"].(string)

// 		cluster = restyKubeImicroservRes.BodyJson["XKUBEIMICROSERV05"].(string)
// 		Logga(ctx, os.Getenv("JsonLog"), "KUBEIMICROSERV OK")
// 	} else {
// 		Logga(ctx, os.Getenv("JsonLog"), "   !!!   KUBEIMICROSERV MISSING")
// 		erro := errors.New("KUBEIMICROSERV MISSING")
// 		return microservices, erro
// 	}
// 	Logga(ctx, os.Getenv("JsonLog"), "")

// 	/* ************************************************************************************************ */
// 	// KUBEMICROSERV
// 	Logga(ctx, os.Getenv("JsonLog"), "Getting KUBEMICROSERV")

// 	argsMS := make(map[string]string)
// 	argsMS["source"] = "devops-8"
// 	argsMS["$select"] = "XKUBEMICROSERV03,XKUBEMICROSERV04,XKUBEMICROSERV05,XKUBEMICROSERV08,XKUBEMICROSERV15,XKUBEMICROSERV16,XKUBEMICROSERV17,XKUBEMICROSERV18"
// 	argsMS["center_dett"] = "dettaglio"
// 	argsMS["$filter"] = "equals(XKUBEMICROSERV05,'" + microservice + "') "
// 	restyKubeMSRes, errKubeMSRes := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsMS, "ms"+devops, "/api/"+os.Getenv("coreApiVersion")+"/"+devops+"/KUBEMICROSERV", devopsToken, dominio, coreApiVersion)
// 	if errKubeMSRes != nil {
// 		Logga(ctx, os.Getenv("JsonLog"), errKubeMSRes.Error())
// 		return microservices, errKubeMSRes
// 	}

// 	hpaTmpl := ""
// 	if len(restyKubeMSRes.BodyJson) > 0 {
// 		microservices.Nome = restyKubeMSRes.BodyJson["XKUBEMICROSERV05"].(string)
// 		microservices.Descrizione = restyKubeMSRes.BodyJson["XKUBEMICROSERV03"].(string)
// 		microservices.Public = int(restyKubeMSRes.BodyJson["XKUBEMICROSERV18"].(float64))
// 		//microservices.Namespace = restyKubeMSRes.BodyJson["XKUBEMICROSERV04_COD"].(string)
// 		microservices.Virtualservice = strconv.FormatFloat(restyKubeMSRes.BodyJson["XKUBEMICROSERV08"].(float64), 'f', 0, 64)

// 		microservices.DatabasebEnable = strconv.FormatFloat(restyKubeMSRes.BodyJson["XKUBEMICROSERV15"].(float64), 'f', 0, 64)

// 		hpaTmpl = restyKubeMSRes.BodyJson["XKUBEMICROSERV16_COD"].(string)
// 		Logga(ctx, os.Getenv("JsonLog"), "KUBEMICROSERV OK")
// 	} else {
// 		Logga(ctx, os.Getenv("JsonLog"), "   !!!   KUBEMICROSERV MISSING")
// 		erro := errors.New("KUBEIMICROSERV MISSING")
// 		return microservices, erro
// 	}
// 	Logga(ctx, os.Getenv("JsonLog"), "")

// 	if autopilot != "1" {
// 		/* ************************************************************************************************ */
// 		// KUBEMICROSERVHPA
// 		Logga(ctx, os.Getenv("JsonLog"), "Getting KUBEMICROSERVHPA")
// 		argsHpa := make(map[string]string)
// 		argsHpa["source"] = "devops-8"
// 		argsHpa["$select"] = "XKUBEMICROSERVHPA04,XKUBEMICROSERVHPA05,XKUBEMICROSERVHPA06,XKUBEMICROSERVHPA07,XKUBEMICROSERVHPA08,XKUBEMICROSERVHPA09,XKUBEMICROSERVHPA10"
// 		argsHpa["center_dett"] = "dettaglio"
// 		argsHpa["$filter"] = "equals(XKUBEMICROSERVHPA03,'" + hpaTmpl + "') "

// 		restyKubeHpaRes, errKubeHpaRes := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsHpa, "ms"+devops, "/api/"+os.Getenv("coreApiVersion")+"/"+devops+"/KUBEMICROSERVHPA", devopsToken, dominio, coreApiVersion)
// 		if errKubeHpaRes != nil {
// 			Logga(ctx, os.Getenv("JsonLog"), errKubeHpaRes.Error())
// 			return microservices, errKubeHpaRes
// 		}

// 		if len(restyKubeHpaRes.BodyJson) > 0 {

// 			// In XKUBEMICROSERVHPA10 salviamo la mappa per personalizzare l'HPA in ogni environment
// 			hpaString, _ := restyKubeHpaRes.BodyJson["XKUBEMICROSERVHPA10"].(string)

// 			checkHpaEnviro := false
// 			var hpaEnviro Hpa
// 			if hpaString != "" {

// 				var hpaMap map[string]Hpa
// 				json.Unmarshal([]byte(hpaString), &hpaMap)

// 				hpaEnviro, checkHpaEnviro = hpaMap[enviro]
// 			}

// 			// Se esiste la personalizzazione per environment, prendo quella, altrimenti il default delle altri colonne
// 			if checkHpaEnviro {
// 				microservices.Hpa = hpaEnviro
// 			} else {
// 				var hpa Hpa
// 				hpa.MinReplicas = strconv.FormatFloat(restyKubeHpaRes.BodyJson["XKUBEMICROSERVHPA04"].(float64), 'f', 0, 64)
// 				hpa.MaxReplicas = strconv.FormatFloat(restyKubeHpaRes.BodyJson["XKUBEMICROSERVHPA05"].(float64), 'f', 0, 64)
// 				hpa.CpuTipoTarget = restyKubeHpaRes.BodyJson["XKUBEMICROSERVHPA06"].(string)
// 				hpa.CpuTarget = restyKubeHpaRes.BodyJson["XKUBEMICROSERVHPA07"].(string)
// 				hpa.MemTipoTarget = restyKubeHpaRes.BodyJson["XKUBEMICROSERVHPA08"].(string)
// 				hpa.MemTarget = restyKubeHpaRes.BodyJson["XKUBEMICROSERVHPA09"].(string)
// 				microservices.Hpa = hpa
// 			}
// 			Logga(ctx, os.Getenv("JsonLog"), "KUBEMICROSERVHPA OK")
// 		} else {
// 			Logga(ctx, os.Getenv("JsonLog"), "   !!!   KUBEMICROSERVHPA MISSING")
// 			erro := errors.New("KUBEMICROSERVHPA MISSING")
// 			return microservices, erro
// 		}
// 		Logga(ctx, os.Getenv("JsonLog"), "")

// 		/* ************************************************************************************************ */
// 	}

// 	/* ************************************************************************************************ */
// 	// SELKUBEDKRLIST
// 	Logga(ctx, os.Getenv("JsonLog"), "Getting SELKUBEDKRLIST")
// 	argsDkr := make(map[string]string)
// 	argsDkr["center_dett"] = "visualizza"
// 	argsDkr["$filter"] = "equals(XSELKUBEDKRLIST10,'" + microservices.Nome + "') "

// 	restyDkrLstRes, errDkrLstRes := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsDkr, "ms"+devops, "/api/"+os.Getenv("coreApiVersion")+"/devops/custom/SELKUBEDKRLIST/values", devopsToken, dominio, coreApiVersion)
// 	if errDkrLstRes != nil {
// 		Logga(ctx, os.Getenv("JsonLog"), errDkrLstRes.Error())
// 		return microservices, errDkrLstRes
// 	}

// 	if len(restyDkrLstRes.BodyArray) > 0 {
// 		var pods []Pod
// 		for _, x := range restyDkrLstRes.BodyArray {

// 			/* ************************************************************************************************ */

// 			var pod Pod

// 			pod.Docker = x["XSELKUBEDKRLIST03"].(string)
// 			docker := pod.Docker
// 			pod.GitRepo = x["XSELKUBEDKRLIST04"].(string)
// 			resourceTmpl := x["XSELKUBEDKRLIST05"].(string)
// 			pod.Descr = x["XSELKUBEDKRLIST06"].(string)
// 			pod.Dockerfile = x["XSELKUBEDKRLIST07"].(string)
// 			pod.Tipo = x["XSELKUBEDKRLIST08"].(string)
// 			pod.Vpn = int(x["XSELKUBEDKRLIST09"].(float64))
// 			pod.Workdir = x["XSELKUBEDKRLIST11"].(string)

// 			// --------------------------------
// 			// CERCHIAMO LE VERSIONI E GLI SHA
// 			//
// 			// in caso di promote leggo da deploylog
// 			// in caso di build leggo da KUBEDKRBUILD
// 			/* ************************************************************************************************ */

// 			Logga(ctx, os.Getenv("JsonLog"), " *** SCEGLIAMO SE PRENDERE I DETTAGLI DA DEPLOYLOG08 o KUBEDKRBUILD")
// 			Logga(ctx, os.Getenv("JsonLog"), "versione: "+versione)
// 			Logga(ctx, os.Getenv("JsonLog"), "enviro: "+enviro)
// 			Logga(ctx, os.Getenv("JsonLog"), "se enviro == int e versione == \"\" siamo in BUILD, DELPOYLOG NON ESISTE E QUINDI VAI DI KUBEDKRBUILD")

// 			if versione != "" && enviro == "int" {
// 				argsDeploy := make(map[string]string)
// 				argsDeploy["$fullquery"] = " select XDEPLOYLOG08 "
// 				argsDeploy["$fullquery"] += " from TB_ANAG_KUBEIMICROSERV00 "
// 				argsDeploy["$fullquery"] += " join TB_ANAG_DEPLOYLOG00 tad ON (XKUBEIMICROSERV03 = XDEPLOYLOG04) "
// 				argsDeploy["$fullquery"] += " where XKUBEIMICROSERV04 = '" + microservices.Nome + "' "
// 				argsDeploy["$fullquery"] += " and XDEPLOYLOG05 = '" + versione + "' "
// 				argsDeploy["$fullquery"] += " AND XDEPLOYLOG06 = '1' "

// 				restyDeployRes, err := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsDeploy, "msdevops", "/api/"+os.Getenv("coreApiVersion")+"/devops/custom/KUBEIMICROSERV/values", devopsToken, os.Getenv("apiDomain"), os.Getenv("coreApiVersion"))
// 				if err != nil {
// 					Logga(ctx, os.Getenv("JsonLog"), err.Error())
// 				}
// 				if restyDeployRes.Errore < 0 {
// 					Logga(ctx, os.Getenv("JsonLog"), restyDeployRes.Log)
// 				}

// 				type DockerShaVersion struct {
// 					Docker       string `json:"docker"`
// 					Versione     string `json:"versione"`
// 					Merged       string `json:"merged"`
// 					Tag          string `json:"tag"`
// 					MasterDev    string `json:"masterDev"`
// 					ReleaseNote  string `json:"releaseNote"`
// 					SprintBranch string `json:"sprintBranch"`
// 					Sha          string `json:"sha"`
// 				}

// 				if len(restyDeployRes.BodyArray) > 0 {
// 					for _, x := range restyDeployRes.BodyArray {
// 						var dsvs []DockerShaVersion

// 						tipoDeploy := x["XDEPLOYLOG08"].(string)
// 						_ = json.Unmarshal([]byte(tipoDeploy), &dsvs)

// 						for _, dsv := range dsvs {
// 							if dsv.Docker == docker {
// 								var branchs Branch
// 								branchs.Branch = dsv.SprintBranch
// 								branchs.Version = dsv.Versione
// 								branchs.Sha = dsv.Sha

// 								podBuild := &PodBuild{
// 									Versione:     dsv.Versione,
// 									Merged:       dsv.Merged,
// 									Tag:          dsv.Tag,
// 									MasterDev:    dsv.MasterDev,
// 									ReleaseNote:  dsv.ReleaseNote,
// 									SprintBranch: dsv.SprintBranch,
// 								}
// 								pod.PodBuild = podBuild
// 								// var podBuild PodBuild
// 								// podBuild.Versione = dsv.Versione
// 								// podBuild.Merged = dsv.Merged
// 								// podBuild.Tag = dsv.Tag
// 								// podBuild.MasterDev = dsv.MasterDev
// 								// podBuild.ReleaseNote = dsv.ReleaseNote
// 								// podBuild.SprintBranch = dsv.SprintBranch

// 								pod.PodBuild = podBuild
// 								pod.Branch = branchs

// 							}
// 						}
// 					}
// 				}

// 			} else { // MANCA LA VERSIONE ERGO SIAMO IN POST BUILD E LA CERCHIAMO DA KUBEDKRBUILD

// 				/* ************************************************************************************************ */
// 				// KUBEDKRBUILD
// 				Logga(ctx, os.Getenv("JsonLog"), "Getting KUBEDKRBUILD func.go 2")
// 				argsBld := make(map[string]string)
// 				argsBld["$fullquery"] = "select XKUBEDKRBUILD06,XKUBEDKRBUILD04,XKUBEDKRBUILD07,XKUBEDKRBUILD09,XKUBEDKRBUILD10,XKUBEDKRBUILD12,XKUBEDKRBUILD13 "
// 				argsBld["$fullquery"] += "from TB_ANAG_KUBEDKRBUILD00 "
// 				argsBld["$fullquery"] += "where 1>0 "
// 				argsBld["$fullquery"] += "AND XKUBEDKRBUILD03 = '" + docker + "' "
// 				argsBld["$fullquery"] += "AND XKUBEDKRBUILD08 = '" + team + "' "
// 				argsBld["$fullquery"] += " order by cast(XKUBEDKRBUILD06 as unsigned) DESC "
// 				argsBld["$fullquery"] += " limit 1 "
// 				fmt.Println(argsBld["$fullquery"])

// 				restyKubeBldRes, errKubeBldRes := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsBld, "ms"+devops, "/api/"+os.Getenv("coreApiVersion")+"/"+devops+"/custom/KUBEDKRBUILD/values", devopsToken, dominio, coreApiVersion)

// 				if errKubeBldRes != nil {
// 					//fmt.Println("A")
// 					Logga(ctx, os.Getenv("JsonLog"), errKubeBldRes.Error())
// 					return microservices, errKubeBldRes
// 				}
// 				if len(restyKubeBldRes.BodyArray) > 0 {

// 					// fmt.Println("B")
// 					var branchs Branch
// 					branchs.Branch = restyKubeBldRes.BodyArray[0]["XKUBEDKRBUILD04"].(string)
// 					branchs.Version = restyKubeBldRes.BodyArray[0]["XKUBEDKRBUILD06"].(string)
// 					branchs.Sha = restyKubeBldRes.BodyArray[0]["XKUBEDKRBUILD07"].(string)

// 					podBuild := &PodBuild{
// 						Versione:     restyKubeBldRes.BodyArray[0]["XKUBEDKRBUILD06"].(string),
// 						Merged:       restyKubeBldRes.BodyArray[0]["XKUBEDKRBUILD13"].(string),
// 						Tag:          restyKubeBldRes.BodyArray[0]["XKUBEDKRBUILD09"].(string),
// 						MasterDev:    restyKubeBldRes.BodyArray[0]["XKUBEDKRBUILD04"].(string),
// 						ReleaseNote:  restyKubeBldRes.BodyArray[0]["XKUBEDKRBUILD12"].(string),
// 						SprintBranch: restyKubeBldRes.BodyArray[0]["XKUBEDKRBUILD10"].(string),
// 					}
// 					// var podBuild PodBuild
// 					// podBuild.Versione = restyKubeBldRes.BodyArray[0]["XKUBEDKRBUILD06"].(string)
// 					// podBuild.Merged = restyKubeBldRes.BodyArray[0]["XKUBEDKRBUILD13"].(string)
// 					// podBuild.Tag = restyKubeBldRes.BodyArray[0]["XKUBEDKRBUILD09"].(string)
// 					// podBuild.MasterDev = restyKubeBldRes.BodyArray[0]["XKUBEDKRBUILD04"].(string)
// 					// podBuild.ReleaseNote = restyKubeBldRes.BodyArray[0]["XKUBEDKRBUILD12"].(string)
// 					// podBuild.SprintBranch = restyKubeBldRes.BodyArray[0]["XKUBEDKRBUILD10"].(string)

// 					pod.PodBuild = podBuild
// 					pod.Branch = branchs
// 					Logga(ctx, os.Getenv("JsonLog"), "KUBEDKRBUILD OK")
// 				}
// 			}

// 			/* ************************************************************************************************ */

// 			/* ************************************************************************************************ */
// 			// KUBEDKRMOUNT
// 			Logga(ctx, os.Getenv("JsonLog"), "Getting KUBEDKRMOUNT")
// 			argsMnt := make(map[string]string)
// 			argsMnt["source"] = "devops-8"
// 			argsMnt["$select"] = "XKUBEDKRMOUNT04,XKUBEDKRMOUNT05,XKUBEDKRMOUNT06,XKUBEDKRMOUNT07,XKUBEDKRMOUNT08"
// 			argsMnt["center_dett"] = "visualizza"
// 			argsMnt["$filter"] = "equals(XKUBEDKRMOUNT03,'" + docker + "') "

// 			restyKubeMntRes, errKubeMntRes := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsMnt, "ms"+devops, "/api/"+os.Getenv("coreApiVersion")+"/"+devops+"/KUBEDKRMOUNT", devopsToken, dominio, coreApiVersion)
// 			if errKubeMntRes != nil {
// 				Logga(ctx, os.Getenv("JsonLog"), errKubeMntRes.Error())
// 				return microservices, errKubeMntRes
// 			}

// 			if len(restyKubeMntRes.BodyArray) > 0 {
// 				var mounts []Mount
// 				for _, xMnt := range restyKubeMntRes.BodyArray {

// 					var mount Mount
// 					mount.Nome = xMnt["XKUBEDKRMOUNT04"].(string)
// 					mount.Mount = xMnt["XKUBEDKRMOUNT05"].(string)
// 					mount.Subpath = xMnt["XKUBEDKRMOUNT06"].(string)
// 					mount.ClaimName = xMnt["XKUBEDKRMOUNT07"].(string)

// 					if xMnt["XKUBEDKRMOUNT08"] != nil {
// 						fromSecretFloat := xMnt["XKUBEDKRMOUNT08"].(float64)
// 						mount.FromSecret = fromSecretFloat == 1
// 					}

// 					mounts = append(mounts, mount)
// 				}
// 				pod.Mount = mounts
// 				Logga(ctx, os.Getenv("JsonLog"), "KUBEDKRMOUNT OK")
// 			} else {
// 				Logga(ctx, os.Getenv("JsonLog"), "KUBEDKRMOUNT MISSING")
// 			}
// 			Logga(ctx, os.Getenv("JsonLog"), "")

// 			/* ************************************************************************************************ */
// 			// GET FUTURE TOGGLE

// 			var cfgMaps []ConfigMap
// 			argsCfgMap := make(map[string]string)
// 			argsCfgMap["source"] = "devops-9"
// 			argsCfgMap["center_dett"] = "allviews"
// 			argsCfgMap["$select"] = "XKUBEDKRCONFIGMAP06,XKUBEDKRCONFIGMAP07,XKUBEDKRCONFIGMAP08,XKUBEDKRCONFIGMAP09,XKUBEDKRCONFIGMAP10"
// 			argsCfgMap["$filter"] = "equals(XKUBEDKRCONFIGMAP03,'" + cluster + "')"
// 			argsCfgMap["$filter"] += " AND equals(XKUBEDKRCONFIGMAP04,'" + enviro + "')"
// 			argsCfgMap["$filter"] += " AND equals(XKUBEDKRCONFIGMAP05,'" + docker + "')"

// 			restyKUBEDKRCONFIGMAPRes, errCfgMap := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsCfgMap, "msdevops", "/api/"+os.Getenv("coreApiVersion")+"/devops/KUBEDKRCONFIGMAP", devopsToken, dominio, coreApiVersion)
// 			if errCfgMap != nil {
// 				Logga(ctx, os.Getenv("JsonLog"), errCfgMap.Error())
// 				return microservices, errCfgMap
// 			}
// 			if restyKUBEDKRCONFIGMAPRes.Errore < 0 {
// 				erro := errors.New(restyKUBEDKRCONFIGMAPRes.Log)
// 				Logga(ctx, os.Getenv("JsonLog"), erro.Error())
// 				return microservices, erro
// 			}

// 			if len(restyKUBEDKRCONFIGMAPRes.BodyArray) > 0 {
// 				for _, xCfg := range restyKUBEDKRCONFIGMAPRes.BodyArray {
// 					var cfgMap ConfigMap
// 					cfgMap.Name = xCfg["XKUBEDKRCONFIGMAP06"].(string)
// 					cfgMap.ConfigType = xCfg["XKUBEDKRCONFIGMAP07"].(string)
// 					cfgMap.MountType = xCfg["XKUBEDKRCONFIGMAP08"].(string)
// 					cfgMap.MountPath = xCfg["XKUBEDKRCONFIGMAP09"].(string)
// 					cfgMap.Content = xCfg["XKUBEDKRCONFIGMAP10"].(string)
// 					cfgMaps = append(cfgMaps, cfgMap)
// 				}
// 				pod.ConfigMap = cfgMaps
// 				Logga(ctx, os.Getenv("JsonLog"), "KUBEDKRCONFIGMAP OK")
// 			} else {
// 				Logga(ctx, os.Getenv("JsonLog"), "KUBEDKRCONFIGMAP MISSING")
// 			}

// 			/* ************************************************************************************************ */
// 			// KUBEDKRRESOURCE
// 			Logga(ctx, os.Getenv("JsonLog"), "Getting KUBEDKRRESOURCE")
// 			argsSrc := make(map[string]string)
// 			argsSrc["source"] = "devops-8"
// 			argsSrc["$select"] = "XKUBEDKRRESOURCE04,XKUBEDKRRESOURCE05,XKUBEDKRRESOURCE06,XKUBEDKRRESOURCE07"
// 			argsSrc["center_dett"] = "dettaglio"
// 			argsSrc["$filter"] = "equals(XKUBEDKRRESOURCE03,'" + resourceTmpl + "') "

// 			restyKubeSrcRes, errKubeSrcRes := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsSrc, "ms"+devops, "/api/"+os.Getenv("coreApiVersion")+"/"+devops+"/KUBEDKRRESOURCE", devopsToken, dominio, coreApiVersion)
// 			if errKubeSrcRes != nil {
// 				Logga(ctx, os.Getenv("JsonLog"), errKubeSrcRes.Error())
// 				return microservices, errKubeSrcRes
// 			}

// 			if len(restyKubeSrcRes.BodyJson) > 0 {
// 				var resource Resource

// 				resource.CpuReq = restyKubeSrcRes.BodyJson["XKUBEDKRRESOURCE04"].(string) //   -- cpu res
// 				resource.MemReq = restyKubeSrcRes.BodyJson["XKUBEDKRRESOURCE05"].(string) //   -- mem res
// 				resource.CpuLim = restyKubeSrcRes.BodyJson["XKUBEDKRRESOURCE06"].(string) //   -- cpu limit
// 				resource.MemLim = restyKubeSrcRes.BodyJson["XKUBEDKRRESOURCE07"].(string) //   -- mem limit

// 				pod.Resource = resource
// 				Logga(ctx, os.Getenv("JsonLog"), "KUBEDKRRESOURCE OK")
// 			} else {
// 				erro := errors.New("KUBEDKRRESOURCE MISSING")
// 				return microservices, erro
// 			}
// 			Logga(ctx, os.Getenv("JsonLog"), "")

// 			/* ************************************************************************************************ */

// 			/* ************************************************************************************************ */

// 			// KUBEDKRPROBE
// 			Logga(ctx, os.Getenv("JsonLog"), "Getting KUBEDKRPROBE")
// 			argsProbes := make(map[string]string)
// 			argsProbes["source"] = "devops-8"
// 			//argsProbes["$select"] = "XKUBEDKRPROBE04,XKUBEDKRPROBE05,XKUBEDKRPROBE06,XKUBEDKRPROBE07,XKUBEDKRPROBE08,XKUBEDKRPROBE09,XKUBEDKRPROBE10"
// 			//argsProbes["$select"] += "XKUBEDKRPROBE11,XKUBEDKRPROBE12,XKUBEDKRPROBE13,XKUBEDKRPROBE14,XKUBEDKRPROBE15,XKUBEDKRPROBE16,XKUBEDKRPROBE17,XKUBEDKRPROBE18,XKUBEDKRPROBE19"
// 			argsProbes["center_dett"] = "allviews"
// 			argsProbes["$filter"] = "equals(XKUBEDKRPROBE03,'" + docker + "') "

// 			restyKubePrbRes, errKubePrbRes := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsProbes, "ms"+devops, "/api/"+os.Getenv("coreApiVersion")+"/"+devops+"/KUBEDKRPROBE", devopsToken, dominio, coreApiVersion)
// 			if errKubePrbRes != nil {
// 				Logga(ctx, os.Getenv("JsonLog"), errKubePrbRes.Error())
// 				return microservices, errKubePrbRes
// 			}

// 			if len(restyKubePrbRes.BodyArray) > 0 {

// 				var probes []Probes
// 				for _, xPrb := range restyKubePrbRes.BodyArray {

// 					var elem Probes

// 					elem.Category = xPrb["XKUBEDKRPROBE04"].(string)
// 					elem.Type = xPrb["XKUBEDKRPROBE05"].(string)
// 					if xPrb["XKUBEDKRPROBE06"] == nil {
// 						elem.Command = ""
// 					} else {
// 						elem.Command = xPrb["XKUBEDKRPROBE06"].(string)
// 					}
// 					elem.HttpHost = xPrb["XKUBEDKRPROBE07"].(string)
// 					elem.HttpPort = int(xPrb["XKUBEDKRPROBE08"].(float64))
// 					elem.HttpPath = xPrb["XKUBEDKRPROBE09"].(string)
// 					if xPrb["XKUBEDKRPROBE10"] == nil {
// 						elem.HttpHeaders = ""
// 					} else {
// 						elem.HttpHeaders = xPrb["XKUBEDKRPROBE10"].(string)
// 					}
// 					elem.HttpScheme = xPrb["XKUBEDKRPROBE11"].(string)
// 					elem.TcpHost = xPrb["XKUBEDKRPROBE12"].(string)
// 					elem.TcpPort = int(xPrb["XKUBEDKRPROBE13"].(float64))
// 					elem.GrpcPort = int(xPrb["XKUBEDKRPROBE14"].(float64))
// 					elem.InitialDelaySeconds = int(xPrb["XKUBEDKRPROBE15"].(float64))
// 					elem.PeriodSeconds = int(xPrb["XKUBEDKRPROBE16"].(float64))
// 					elem.TimeoutSeconds = int(xPrb["XKUBEDKRPROBE17"].(float64))
// 					elem.SuccessThreshold = int(xPrb["XKUBEDKRPROBE18"].(float64))
// 					elem.FailureThreshold = int(xPrb["XKUBEDKRPROBE19"].(float64))
// 					probes = append(probes, elem)

// 				}
// 				pod.Probes = probes

// 				//Logga(ctx, os.Getenv("JsonLog"), probes)
// 				Logga(ctx, os.Getenv("JsonLog"), "KUBEDKRPROBE OK")
// 			} else {
// 				Logga(ctx, os.Getenv("JsonLog"), "KUBEDKRPROBE MISSING")
// 			}

// 			Logga(ctx, os.Getenv("JsonLog"), "")

// 			/* ************************************************************************************************ */

// 			/* ************************************************************************************************ */
// 			// KUBESERVICEDKR
// 			Logga(ctx, os.Getenv("JsonLog"), "Getting KUBESERVICEDKR")
// 			argsSrvDkr := make(map[string]string)
// 			argsSrvDkr["source"] = "devops-8"
// 			argsSrvDkr["$select"] = "XKUBESERVICEDKR06,XKUBESERVICEDKR05"
// 			argsSrvDkr["center_dett"] = "visualizza"
// 			argsSrvDkr["$filter"] = "equals(XKUBESERVICEDKR04,'" + docker + "') "

// 			restyKubeSrvDkrRes, errKubeSrvDkrRes := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsSrvDkr, "ms"+devops, "/api/"+os.Getenv("coreApiVersion")+"/"+devops+"/KUBESERVICEDKR", devopsToken, dominio, coreApiVersion)
// 			if errKubeSrvDkrRes != nil {
// 				Logga(ctx, os.Getenv("JsonLog"), errKubeSrvDkrRes.Error())
// 				return microservices, errKubeSrvDkrRes
// 			}

// 			if len(restyKubeSrvDkrRes.BodyArray) > 0 {
// 				var port, tipo string
// 				var services []Service
// 				for _, xSrv := range restyKubeSrvDkrRes.BodyArray {

// 					port = strconv.FormatFloat(xSrv["XKUBESERVICEDKR06"].(float64), 'f', 0, 64)
// 					tipo = xSrv["XKUBESERVICEDKR05"].(string)

// 					var service Service
// 					service.Port = port
// 					service.Tipo = tipo
// 					// service.Endpoint = endpoints

// 					services = append(services, service)

// 				}

// 				pod.Service = services

// 				Logga(ctx, os.Getenv("JsonLog"), "KUBESERVICEDKR OK")
// 			} else {
// 				erro := errors.New("KUBESERVICEDKR MISSING")
// 				return microservices, erro
// 			}
// 			// aggiungo pod corrente a pods
// 			pods = append(pods, pod)

// 			microservices.Pod = pods

// 		}

// 		Logga(ctx, os.Getenv("JsonLog"), "SELKUBEDKRLIST OK")
// 	} else {
// 		erro := errors.New("SELKUBEDKRLIST MISSING")
// 		return microservices, erro
// 	}
// 	Logga(ctx, os.Getenv("JsonLog"), "")

//		//LogJson(microservices)
//		Logga(ctx, os.Getenv("JsonLog"), "Seek Microservice details ok")
//		Logga(ctx, os.Getenv("JsonLog"), " - - - - - - - - - - - - - - -  ")
//		Logga(ctx, os.Getenv("JsonLog"), "")
//		// fmt.Println(microservices)
//		// LogJson(microservices)
//		// os.Exit(0)
//		return microservices, erro
//	}
func GetTenant(ctx context.Context, token, dominio, coreApiVersion string) ([]Tenant, error) {
	//Logga(ctx, os.Getenv("JsonLog"), "Get TENANT")

	var erro error
	var tenants []Tenant
	var tenant Tenant

	args := make(map[string]string)

	tenantRes, errtenantRes := ApiCallGET(ctx, os.Getenv("RestyDebug"), args, "msauth", "/api/"+os.Getenv("coreApiVersion")+"/auth/tenants", token, dominio, coreApiVersion)
	//LogJson(tenantRes)
	if errtenantRes != nil {
		Logga(ctx, os.Getenv("JsonLog"), errtenantRes.Error(), "error")
		erro = errors.New(errtenantRes.Error())
		return tenants, erro
	}

	if len(tenantRes.BodyArray) > 0 {
		for _, y := range tenantRes.BodyArray {
			tenant.Tenant = y["tenantId"].(string)
			tenant.Master = strconv.FormatFloat(y["isDefault"].(float64), 'f', 0, 64)
			tenant.Descrizione = y["tenantDescription"].(string)
			tenants = append(tenants, tenant)
		}
	}
	return tenants, erro
}
func GetProfileInfo(ctx context.Context, token, dominio, coreApiVersion string) (map[string]interface{}, error) {

	Logga(ctx, os.Getenv("JsonLog"), "Getting getProfileInfo")

	var erro error
	info := make(map[string]interface{})

	args := make(map[string]string)
	infoRes, _ := ApiCallGET(ctx, os.Getenv("RestyDebug"), args, "msauth", "/api/"+os.Getenv("coreApiVersion")+"/auth/getProfileInfo?getDbInfo=true", token, dominio, coreApiVersion)

	if len(infoRes.BodyJson) > 0 {
		restyProfileInfoResponse := ProfileInfo{}

		b, err := json.Marshal(infoRes.BodyJson)
		if err != nil {
			Logga(ctx, os.Getenv("JsonLog"), err.Error())
			return info, err
		}
		json.Unmarshal(b, &restyProfileInfoResponse)

		info["market"] = restyProfileInfoResponse.Session.Market.Decval
		info["gruppo"] = restyProfileInfoResponse.Session.GrantSession.Gruppo
		info["nome"] = restyProfileInfoResponse.Session.GrantSession.NomeCognome
		info["email"] = restyProfileInfoResponse.Session.GrantSession.Email
		info["vsessionMicroservice"] = restyProfileInfoResponse.Session.Vsession.KUBEMICROSERV

		Logga(ctx, os.Getenv("JsonLog"), "GetProfileInfo OK")
		return info, nil
	} else {
		erro = errors.New("GetProfileInfo MISSING")
		Logga(ctx, os.Getenv("JsonLog"), "GetProfileInfo MISSING")
		return info, erro
	}
}
func GetBuildLastTag(ctx context.Context, team, docker, tipo, tenant, accessToken, loginApiDomain, coreApiVersion, devopsToken string) (string, error) {

	//var erro error

	sprint, erroGCBS := GetCurrentBranchSprint(ctx, team, tipo, tenant, accessToken, loginApiDomain, coreApiVersion, devopsToken)
	if erroGCBS != nil {
		Logga(ctx, os.Getenv("JsonLog"), erroGCBS.Error())
		return "", erroGCBS
	}

	/* ************************************************************************************************ */
	// KUBEIMICROSERV
	Logga(ctx, os.Getenv("JsonLog"), "Getting KUBEDKRBUILD - func.go 1")
	argsImicro := make(map[string]string)
	argsImicro["source"] = "devops-8"
	argsImicro["$select"] = "XKUBEDKRBUILD09"
	argsImicro["center_dett"] = "dettaglio"
	argsImicro["$filter"] = "equals(XKUBEDKRBUILD03,'" + docker + "') "
	argsImicro["$filter"] += " and equals(XKUBEDKRBUILD08,'" + team + "') "
	argsImicro["$filter"] += " and equals(XKUBEDKRBUILD10,'" + sprint + "') "
	argsImicro["$order"] = "CDATA desc"
	argsImicro["num_rows"] = " 1 "

	restyKubeImicroservRes, errKubeImicroservRes := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsImicro, "msdevops", "/api/"+os.Getenv("coreApiVersion")+"/devops/KUBEDKRBUILD", devopsToken, loginApiDomain, coreApiVersion)
	if errKubeImicroservRes != nil {
		Logga(ctx, os.Getenv("JsonLog"), errKubeImicroservRes.Error())
		return "", errKubeImicroservRes
	}

	tag := ""
	if len(restyKubeImicroservRes.BodyJson) > 0 {
		tag = restyKubeImicroservRes.BodyJson["XKUBEDKRBUILD09"].(string)
		Logga(ctx, os.Getenv("JsonLog"), "KUBEDKRBUILD OK")
	} else {
		Logga(ctx, os.Getenv("JsonLog"), "KUBEDKRBUILD MISSING")
	}
	Logga(ctx, os.Getenv("JsonLog"), "")
	//	fmt.Println(tag)
	/* ************************************************************************************************ */

	return tag, nil
}
func GetCurrentBranchSprint(ctx context.Context, team, tipo, tenant, accessToken, loginApiDomain, coreApiVersion, devopsToken string) (string, error) {

	// cerco il token di Corefactory
	Logga(ctx, os.Getenv("JsonLog"), "Getting token")

	/* ************************************************************************************************ */
	// KUBEIMICROSERV
	Logga(ctx, os.Getenv("JsonLog"), "Getting KUBETEAMBRANCH - func.go")
	argsImicro := make(map[string]string)
	argsImicro["source"] = "devops-8"
	argsImicro["$select"] = "XKUBETEAMBRANCH05"
	argsImicro["center_dett"] = "dettaglio"
	argsImicro["$filter"] = "equals(XKUBETEAMBRANCH03,'" + team + "') "
	argsImicro["$filter"] += " and equals(XKUBETEAMBRANCH04,'" + tipo + "') "

	restyKubeImicroservRes, errKubeImicroservRes := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsImicro, "msdevops", "/api/"+os.Getenv("coreApiVersion")+"/devops/KUBETEAMBRANCH", devopsToken, loginApiDomain, coreApiVersion)
	if errKubeImicroservRes != nil {
		return "", errKubeImicroservRes
	}

	sprintBranch := ""
	if len(restyKubeImicroservRes.BodyJson) > 0 {
		sprintBranch = restyKubeImicroservRes.BodyJson["XKUBETEAMBRANCH05"].(string)
		Logga(ctx, os.Getenv("JsonLog"), "KUBETEAMBRANCH OK")
	} else {
		Logga(ctx, os.Getenv("JsonLog"), "KUBETEAMBRANCH MISSING - getCurrentBranchSprint")
	}
	Logga(ctx, os.Getenv("JsonLog"), "")
	/* ************************************************************************************************ */

	return sprintBranch, nil
}
func CreateTag(ctx context.Context, buildArgs BuildArgs, tag, repo string) error {

	Logga(ctx, os.Getenv("JsonLog"), "Create tag: "+tag)
	Logga(ctx, os.Getenv("JsonLog"), "git repo: "+repo)

	// OTTENGO L' HASH del branch vivo
	clientBranch := resty.New()
	clientBranch.Debug = os.Getenv("RestyDebug") == "true"
	var respBranch, restyResponse *resty.Response
	var errBranch, errTag error
	switch buildArgs.TypeGit {
	case "", "bitbucket":
		respBranch, errBranch = clientBranch.R().
			EnableTrace().
			SetBasicAuth(buildArgs.UserGit, buildArgs.TokenGit).
			Get(buildArgs.ApiHostGit + "/repositories/" + buildArgs.WorkspaceGit + "/" + repo + "/refs/branches/" + buildArgs.SprintBranch)
	case "gitub":
		respBranch, errBranch = clientBranch.R().
			EnableTrace().
			SetHeader("Accept", "application/json").
			SetHeader("Accept", "application/vnd.github+json").
			SetBasicAuth(buildArgs.UserGit, buildArgs.TokenGit).
			Get(buildArgs.ApiHostGit + "/repos/" + buildArgs.UserGit + "/" + repo + "/git/refs/heads/" + buildArgs.SprintBranch)
	}

	if errBranch != nil {
		Logga(ctx, os.Getenv("JsonLog"), errBranch.Error())
		return errBranch
	}

	body := ""
	if respBranch.StatusCode() != 200 {

	} else {
		switch buildArgs.TypeGit {
		case "", "bitbucket":
			var branchRes BranchResStruct
			json.Unmarshal(respBranch.Body(), &branchRes)
			// STACCO IL TAG
			body = `{"name": "` + tag + `","target": {  "hash": "` + branchRes.Target.Hash + `"}}`
			client := resty.New()
			client.Debug = os.Getenv("RestyDebug") == "true"
			restyResponse, errTag = client.R().
				SetHeader("Content-Type", "application/json").
				SetBasicAuth(buildArgs.UserGit, buildArgs.TokenGit).
				SetBody(body).
				Post(buildArgs.ApiHostGit + "/repositories/" + buildArgs.WorkspaceGit + "/" + repo + "/refs/tags")
		case "github":
			type ResFrom struct {
				Object struct {
					Sha string `json:"sha"`
				} `json:"object"`
			}
			var branchRes ResFrom
			json.Unmarshal(respBranch.Body(), &branchRes)

			body = `{"tag":"` + tag + `","message":"","object":"` + branchRes.Object.Sha + `","type":"commit"}`

			client := resty.New()
			client.Debug = os.Getenv("RestyDebug") == "true"
			restyResponse, errTag = client.R().
				SetHeader("Content-Type", "application/json").
				SetHeader("Accept", "application/vnd.github+json").
				SetBasicAuth(buildArgs.UserGit, buildArgs.TokenGit).
				SetBody(body).
				Post(buildArgs.ApiHostGit + "/repos/" + buildArgs.UserGit + "/" + repo + "/git/tags")
		}
	}

	/*
		{
		   "type": "error",
		   "error": {
		      "message": "tag \"CORE-202201023-tag\" already exists"
		   }
		}

	*/

	if errTag != nil {
		Logga(ctx, os.Getenv("JsonLog"), errTag.Error())
		return errTag
	}
	if restyResponse.StatusCode() == 201 || restyResponse.StatusCode() == 200 {
		fmt.Println("Tag created")
	} else {
		fmt.Println("Error CODE: " + strconv.Itoa(restyResponse.StatusCode()))
	}
	return nil
}

func GetEnvironmentStatus(ctx context.Context, cluster, enviro, microserice, customer, tenant, accessToken, loginApiDomain, coreApiVersion, resource string) error {

	status := ""

	// cerco il token di Corefactory
	devopsToken, erroT := GetCoreFactoryToken(ctx, tenant, accessToken, loginApiDomain, coreApiVersion, resource, os.Getenv("RestyDebug"))
	if erroT != nil {
		Logga(ctx, os.Getenv("JsonLog"), erroT.Error())
	}

	/* ************************************************************************************************ */
	// KUBEENVSTATUS
	Logga(ctx, os.Getenv("JsonLog"), "Getting KUBEENVSTATUS")
	args := make(map[string]string)
	args["source"] = "devops-8"

	/* ************************************************************************************************ */

	Logga(ctx, os.Getenv("JsonLog"), "Getting KUBEENVSTATUS")

	argsEs := make(map[string]string)
	argsEs["source"] = "devops-8"
	argsEs["$select"] = "XKUBEENVSTATUS07"
	argsEs["center_dett"] = "dettaglio"
	argsEs["$filter"] = "equals(XKUBEENVSTATUS03,'" + cluster + "') "
	argsEs["$filter"] += " and equals(XKUBEENVSTATUS04,'" + enviro + "') "
	argsEs["$filter"] += " and equals(XKUBEENVSTATUS05,'" + microserice + "') "
	if customer != "" {
		args["$filter"] += " and equals(XKUBEENVSTATUS06,'" + customer + "') "
	}

	restyEsRes, errEsRes := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsEs, "msdevops", "/api/"+os.Getenv("coreApiVersion")+"/devops/KUBEENVSTATUS", devopsToken, loginApiDomain, coreApiVersion)
	if errEsRes != nil {
		Logga(ctx, os.Getenv("JsonLog"), restyEsRes.Log)
		return errEsRes
	}

	var erro error
	if len(restyEsRes.BodyJson) > 0 {
		status = strconv.Itoa(int(restyEsRes.BodyJson["XKUBEENVSTATUS07"].(float64)))
		erro = errors.New(status)
		Logga(ctx, os.Getenv("JsonLog"), "KUBEENVSTATUS OK")
	} else {
		Logga(ctx, os.Getenv("JsonLog"), "KUBEENVSTATUS MISSING")
	}
	Logga(ctx, os.Getenv("JsonLog"), "")
	/* ************************************************************************************************ */

	return erro
}
func SetEnvironmentStatus(ctx context.Context, cluster, enviro, microserice, customer, user, toggle, tenant, accessToken, loginApiDomain, coreApiVersion, resource string) error {

	var erro error
	// cerco il token di Corefactory
	devopsToken, erroT := GetCoreFactoryToken(ctx, tenant, accessToken, loginApiDomain, coreApiVersion, resource, os.Getenv("RestyDebug"))
	if erroT != nil {
		Logga(ctx, os.Getenv("JsonLog"), erroT.Error())
		return erroT
	}

	/* ************************************************************************************************ */
	// KUBEENVSTATUS
	Logga(ctx, os.Getenv("JsonLog"), "Setting KUBEENVSTATUS")
	args := make(map[string]string)
	args["source"] = "devops-8"

	/* ************************************************************************************************ */
	// KUBEENVSTATUS

	keyvalueslices := make([]map[string]interface{}, 0)
	keyvalueslice := make(map[string]interface{})
	keyvalueslice["debug"] = true
	keyvalueslice["source"] = "devops-8"
	keyvalueslice["XKUBEENVSTATUS03"] = cluster
	keyvalueslice["XKUBEENVSTATUS04"] = enviro
	keyvalueslice["XKUBEENVSTATUS05"] = microserice
	keyvalueslice["XKUBEENVSTATUS06"] = customer
	if toggle == "ON" {
		keyvalueslice["XKUBEENVSTATUS07"] = "1"
	} else {
		keyvalueslice["XKUBEENVSTATUS07"] = "0"
	}
	keyvalueslice["XKUBEENVSTATUS08"] = user

	keyvalueslices = append(keyvalueslices, keyvalueslice)

	_, erroPost := ApiCallPOST(ctx, os.Getenv("RestyDebug"), keyvalueslices, "msdevops", "/api/"+os.Getenv("coreApiVersion")+"/devops/KUBEENVSTATUS", devopsToken, loginApiDomain, coreApiVersion)

	if erroPost != nil {
		Logga(ctx, os.Getenv("JsonLog"), erroPost.Error())
		return erroPost
	} else {
		if toggle == "ON" {
			erro = errors.New("Environment set to LOCK")
		} else {
			erro = errors.New("Environment set to UNLOCK")
		}
	}

	/* ************************************************************************************************ */

	return erro
}

func GetAccessCluster(ctx context.Context, cluster, devopsToken, loginApiDomain, coreApiVersion string, monolith bool) ClusterAccess {
	/* ************************************************************************************************ */
	// KUBECLUSTER
	Logga(ctx, os.Getenv("JsonLog"), "Getting KUBECLUSTER")

	devops := "devops"
	if monolith {
		devops = "devopsmono"
	}

	argsClu := make(map[string]string)
	argsClu["source"] = "devops-8"
	argsClu["$select"] = "XKUBECLUSTER15,XKUBECLUSTER20,XKUBECLUSTER22"
	argsClu["center_dett"] = "dettaglio"
	argsClu["$filter"] = "equals(XKUBECLUSTER03,'" + cluster + "') "

	restyKubeCluRes, _ := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsClu, "ms"+devops, "/api/"+os.Getenv("coreApiVersion")+"/"+devops+"/KUBECLUSTER", devopsToken, loginApiDomain, coreApiVersion)
	if restyKubeCluRes.Errore < 0 {
		Logga(ctx, os.Getenv("JsonLog"), restyKubeCluRes.Log)
	}

	var cluAcc ClusterAccess
	if len(restyKubeCluRes.BodyJson) > 0 {

		cluAcc.Domain = restyKubeCluRes.BodyJson["XKUBECLUSTER15"].(string)
		cluAcc.AccessToken = restyKubeCluRes.BodyJson["XKUBECLUSTER20"].(string)
		cluAcc.ReffappCustomerID = restyKubeCluRes.BodyJson["XKUBECLUSTER22"].(string)

		Logga(ctx, os.Getenv("JsonLog"), "KUBECLUSTER OK")
	} else {
		Logga(ctx, os.Getenv("JsonLog"), "   !!!   KUBECLUSTER MISSING")
	}
	Logga(ctx, os.Getenv("JsonLog"), "")
	/* ************************************************************************************************ */

	return cluAcc
}
func GetJsonDatabases(ctx context.Context, stage, developer string, market int32, arrConn MasterConn, tenant, accessToken, loginApiDomain, coreApiVersion, resource string, monolith bool) (map[string]interface{}, error) {
	Logga(ctx, os.Getenv("JsonLog"), "Getting Json Db")

	var erro error
	callResponse := map[string]interface{}{}

	if monolith {
		callResponse["monolith"] = "true"
		return callResponse, erro
	}

	devopsToken, erroT := GetCoreFactoryToken(ctx, tenant, accessToken, loginApiDomain, coreApiVersion, resource, os.Getenv("RestyDebug"))
	if erroT != nil {
		Logga(ctx, os.Getenv("JsonLog"), erroT.Error())
	} else {
		Logga(ctx, os.Getenv("JsonLog"), "Token OK")
	}

	clusterDett := GetAccessCluster(ctx, stage, devopsToken, loginApiDomain, coreApiVersion, monolith)
	clusterToken, erro := GetCustomerToken(ctx, clusterDett.AccessToken, clusterDett.ReffappCustomerID, clusterDett.Domain, clusterDett.Domain, coreApiVersion)

	dominio := loginApiDomain

	keyvalueslice := make(map[string]interface{})
	keyvalueslice["debug"] = true
	keyvalueslice["clusterId"] = stage
	keyvalueslice["customerId"] = developer
	keyvalueslice["enableMonolith"] = true
	keyvalueslice["host"] = arrConn.Host
	keyvalueslice["apiUrl"] = "ms-int." + arrConn.Domain
	keyvalueslice["market"] = market
	keyvalueslice["appID"] = "1"
	keyvalueslice["refappCustomerID"] = developer
	keyvalueslice["platformUrl"] = "developer." + arrConn.Domain

	debool, errBool := strconv.ParseBool(os.Getenv("RestyDebug"))
	if errBool != nil {
		return nil, errBool
	}

	client := resty.New()
	client.Debug = debool
	client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	res, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetHeader("microservice", "msappman").
		SetAuthToken(clusterToken).
		SetBody(keyvalueslice).
		Post(dominio + "/api/" + os.Getenv("coreApiVersion") + "/appman/getDeveloperMsList")

	if err != nil { // HTTP ERRORE
		return callResponse, err
	} else {
		if res.StatusCode() != 200 {
			erro = errors.New("ERROR CODE: " + strconv.Itoa(res.StatusCode()))
			return callResponse, erro
		} else {

			err1 := json.Unmarshal(res.Body(), &callResponse)
			if err1 != nil {
				return callResponse, err1
			}
		}
	}

	return callResponse, erro
}
func GetCustomerToken(ctx context.Context, accessToken, refappCustomer, resource, dominio, coreApiVersion string) (string, error) {

	Logga(ctx, os.Getenv("JsonLog"), "getCustomerToken")
	Logga(ctx, os.Getenv("JsonLog"), "Customer Token "+dominio)
	Logga(ctx, os.Getenv("JsonLog"), "refappCustomer "+refappCustomer)
	Logga(ctx, os.Getenv("JsonLog"), "resource "+resource)
	Logga(ctx, os.Getenv("JsonLog"), "accessToken "+accessToken)

	var erro error

	ct := time.Now()
	now := ct.Format("20060102150405")
	h := sha1.New()
	h.Write([]byte(now))
	sha := hex.EncodeToString(h.Sum(nil))

	argsAuth := make(map[string]interface{})
	argsAuth["access_token"] = accessToken
	argsAuth["refappCustomer"] = refappCustomer
	argsAuth["resource"] = resource
	argsAuth["uuid"] = "devops-" + sha

	restyAuthResponse, restyAuthErr := ApiCallLOGIN(ctx, os.Getenv("RestyDebug"), argsAuth, "msauth", "/api/"+os.Getenv("coreApiVersion")+"/auth/login", dominio, coreApiVersion)
	if restyAuthErr != nil {
		return "", restyAuthErr
	}
	if len(restyAuthResponse) > 0 {
		return restyAuthResponse["idToken"].(string), erro
	} else {
		erro = errors.New("token MISSING")
		return "", erro
	}
}

func GetCfToolEnv(ctx context.Context, token, dominio, tenant, coreApiVersion, enviro string, monolith bool) (TenantEnv, error) {

	Logga(ctx, os.Getenv("JsonLog"), "Getting KUBETENANTENV")

	var erro error
	devops := "devops"
	if monolith {
		devops = "devopsmono"
	}

	args := make(map[string]string)
	args["center_dett"] = "dettaglio"
	args["source"] = "devops-8"
	args["$filter"] = "equals(XKUBETENANTENV03,'" + strings.Replace(dominio, "https://", "", -1) + "') "
	args["$filter"] += " and equals(XKUBETENANTENV19,'" + tenant + "') "
	args["$filter"] += " and equals(XKUBETENANTENV18,'" + enviro + "') "

	envRes, _ := ApiCallGET(ctx, os.Getenv("RestyDebug"), args, "ms"+devops, "/api/"+os.Getenv("coreApiVersion")+"/"+devops+"/KUBETENANTENV", token, dominio, coreApiVersion)

	var tntEnv TenantEnv

	if len(envRes.BodyJson) > 0 {

		tntEnv.TelegramKey = envRes.BodyJson["XKUBETENANTENV04"].(string)
		tntEnv.TelegramID = envRes.BodyJson["XKUBETENANTENV05"].(string)
		tntEnv.CoreApiVersion = envRes.BodyJson["XKUBETENANTENV06"].(string)
		tntEnv.CoreApiPort = envRes.BodyJson["XKUBETENANTENV07"].(string)
		tntEnv.CoreAccessToken = envRes.BodyJson["XKUBETENANTENV08"].(string)
		tntEnv.AtlassianHost = envRes.BodyJson["XKUBETENANTENV09"].(string)
		tntEnv.AtlassianUser = envRes.BodyJson["XKUBETENANTENV10"].(string)
		tntEnv.AtlassianToken = envRes.BodyJson["XKUBETENANTENV11"].(string)
		tntEnv.ApiHostGit = envRes.BodyJson["XKUBETENANTENV12"].(string)
		tntEnv.UrlGit = envRes.BodyJson["XKUBETENANTENV20"].(string)
		tntEnv.TypeGit = envRes.BodyJson["XKUBETENANTENV21"].(string)
		tntEnv.UserGit = envRes.BodyJson["XKUBETENANTENV13"].(string)
		tntEnv.TokenGit = envRes.BodyJson["XKUBETENANTENV14"].(string)
		tntEnv.WorkspaceGit = envRes.BodyJson["XKUBETENANTENV15"].(string)
		tntEnv.CoreGkeProject = envRes.BodyJson["XKUBETENANTENV16"].(string)
		tntEnv.CoreGkeUrl = envRes.BodyJson["XKUBETENANTENV17"].(string)
		tntEnv.CoreApiDominio = envRes.BodyJson["XKUBETENANTENV18"].(string)
		tntEnv.WorkspaceKey = envRes.BodyJson["XKUBETENANTENV22"].(string)
		tntEnv.WorkspaceSecret = envRes.BodyJson["XKUBETENANTENV23"].(string)
		tntEnv.WorkspaceRefreshToken = envRes.BodyJson["XKUBETENANTENV24"].(string)

		Logga(ctx, os.Getenv("JsonLog"), "KUBETENANTENV OK")
	} else {
		Logga(ctx, os.Getenv("JsonLog"), "KUBETENANTENV MISSING")
		erro = errors.New("KUBETENANTENV MISSING")
	}

	return tntEnv, erro

}
func GetDeploymentApi(microservice, namespace, apiHost, apiToken string, scaleToZero, debug bool) (DeploymntStatus, error) {

	var erro error

	var deploy DeploymntStatus

	clientKUBE := resty.New()
	clientKUBE.Debug = debug
	clientKUBE.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	endPoint := "https://" + apiHost + "/apis/apps/v1/namespaces/" + namespace + "/deployments?labelSelector=app=" + microservice
	if scaleToZero {
		endPoint = "https://" + apiHost + "/apis/serving.knative.dev/v1/namespaces/" + namespace + "/services"
	}
	resKUBE, errKUBE := clientKUBE.R().
		SetHeader("Content-Type", "application/json").
		SetAuthToken(apiToken).
		Get(endPoint)

	if errKUBE != nil {
		return deploy, errKUBE
	}

	if resKUBE.StatusCode() != 200 {
		erro = errors.New("API Res Status: " + resKUBE.Status())
		return deploy, erro
	}

	a := map[string]interface{}{}
	errUm := json.Unmarshal(resKUBE.Body(), &a)
	if errUm != nil {
		return deploy, errUm
	}

	jsonStr, errMa := json.Marshal(a)
	if errMa != nil {
		return deploy, errMa
	}

	errUm2 := json.Unmarshal(jsonStr, &deploy)
	if errUm2 != nil {
		return deploy, errUm2
	}

	return deploy, erro
}

func CheckPodHealth(microservice, versione, namespace, apiHost, apiToken string, scaleToZero bool, debug string) (bool, error) {

	var erro error
	var c context.Context

	Logga(c, os.Getenv("JsonLog"), "Checkpodhealth: "+microservice+" "+versione)

	debool, errBool := strconv.ParseBool(debug)
	if errBool != nil {
		return false, errBool
	}

	msDeploy := microservice + "-v" + versione
	if scaleToZero {
		msDeploy = microservice
	}
	msMatch := false
	i := 0
	for {
		item, errDpl := GetDeploymentApi(microservice, namespace, apiHost, apiToken, scaleToZero, debool)
		if errDpl != nil {
			return false, errDpl
		} else {

			if len(item.Items) == 0 {
				erro = errors.New("no Deployment Found in Namespace")
				return false, erro
			}

			for _, item := range item.Items {
				Logga(c, os.Getenv("JsonLog"), "LISTA DEPLOY: "+item.Metadata.Name+"-"+msDeploy)
				if item.Metadata.Name == msDeploy {
					msMatch = true
					if !scaleToZero {
						var ctx context.Context
						Logga(ctx, os.Getenv("JsonLog"), item.Metadata.Name+" desired: "+strconv.Itoa(item.Spec.Replicas)+" - aviable: "+strconv.Itoa(item.Status.ReadyReplicas))

						if item.Spec.Replicas == item.Status.ReadyReplicas {
							return true, erro
						}
					} else {
						return true, erro
					}
				}

				// sto girando a vuoto perche nessun item risponde a cio che cerco
				if i >= 1 && !msMatch {
					Logga(c, os.Getenv("JsonLog"), "nessun item risponde a cio che cerco")
					erro = errors.New("nessun item risponde a cio che cerco")
					return false, erro
				}
			}

			i++
			time.Sleep(10 * time.Second)
			if i > 150 { // superati 5 minuti direi che stann e plobble'
				erro = errors.New("time Out. Pod is not Running")
				return false, erro
			}
		}
	}
}
func DeleteObjectsApi(namespace, apiHost, apiToken, object, kind string, debug bool) error {

	/* *************************************** */
	// LEGGIMI
	//
	// PER AVERE TUTTE LE API :
	//
	// k api-resources

	//http://localhost:8080/apis/autoscaling/v1/namespaces/uat-powerna/horizontalpodautoscalers/
	/*
	   NAME                               SHORTNAMES      APIVERSION                             NAMESPACED   KIND
	   deployments                        deploy          apps/v1                                true         Deployment
	   horizontalpodautoscalers           hpa             autoscaling/v1                         true         HorizontalPodAutoscaler
	*/
	apiversion := ""
	name := ""
	switch kind {
	case "hpa":
		apiversion = "autoscaling/v1"
		name = "horizontalpodautoscalers"
	case "deployment":
		apiversion = "apps/v1"
		name = "deployments"
	case "knservice":
		apiversion = "serving.knative.dev/v1"
		name = "services"
	}

	var erro error

	args := make(map[string]string)
	args["kind"] = "DeleteOptions"
	args["apiVersion"] = apiversion
	args["apiVerpropagationPolicysion"] = "Foreground"

	clientKUBE := resty.New()
	clientKUBE.Debug = debug
	clientKUBE.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	resKUBE, errKUBE := clientKUBE.R().
		SetHeader("Content-Type", "application/json").
		SetAuthToken(apiToken).
		SetQueryParams(args).
		Delete("https://" + apiHost + "/apis/" + apiversion + "/namespaces/" + namespace + "/" + name + "/" + object)

	if errKUBE != nil {
		return errKUBE
	}

	if resKUBE.StatusCode() != 200 {
		erro = errors.New("API Res Status: " + resKUBE.Status())
		return erro
	}

	a := map[string]interface{}{}
	errUm := json.Unmarshal(resKUBE.Body(), &a)
	if errUm != nil {
		return errUm
	}

	return erro
}
func GetOverrideTenantEnv(ctx context.Context, bearerToken, team string, tntEnv TenantEnv, dominio, coreApiVersion string) (TenantEnv, string, error) {
	// ho le env di default
	// cerco per ogni valore su KUBETEAMBRANCH se trovo sovrascrivo
	var erro error

	var sprintBranch string

	args := make(map[string]string)
	args["center_dett"] = "dettaglio"
	args["source"] = "devops-8"
	args["$filter"] = "equals(XKUBETEAMBRANCH03,'" + team + "') "

	envRes, _ := ApiCallGET(ctx, os.Getenv("RestyDebug"), args, "msdevops", "/api/"+os.Getenv("coreApiVersion")+"/devops/KUBETEAMBRANCH", bearerToken, dominio, coreApiVersion)

	if len(envRes.BodyJson) > 0 {

		sprintBranch = envRes.BodyJson["XKUBETEAMBRANCH05"].(string)

		if envRes.BodyJson["XKUBETEAMBRANCH07"].(string) != "" {
			Logga(ctx, os.Getenv("JsonLog"), "overrdide 07")
			tntEnv.TelegramKey = envRes.BodyJson["XKUBETEAMBRANCH07"].(string)
		}

		if envRes.BodyJson["XKUBETEAMBRANCH08"].(string) != "" {
			Logga(ctx, os.Getenv("JsonLog"), "overrdide 08")
			tntEnv.TelegramID = envRes.BodyJson["XKUBETEAMBRANCH08"].(string)
		}

		if envRes.BodyJson["XKUBETEAMBRANCH09"].(string) != "" {
			Logga(ctx, os.Getenv("JsonLog"), "overrdide 09")
			tntEnv.AtlassianHost = envRes.BodyJson["XKUBETEAMBRANCH09"].(string)
		}

		if envRes.BodyJson["XKUBETEAMBRANCH10"].(string) != "" {
			Logga(ctx, os.Getenv("JsonLog"), "overrdide 10")
			tntEnv.AtlassianUser = envRes.BodyJson["XKUBETEAMBRANCH10"].(string)
		}

		if envRes.BodyJson["XKUBETEAMBRANCH11"].(string) != "" {
			Logga(ctx, os.Getenv("JsonLog"), "overrdide 11")
			tntEnv.AtlassianToken = envRes.BodyJson["XKUBETEAMBRANCH11"].(string)
		}

		if envRes.BodyJson["XKUBETEAMBRANCH12"].(string) != "" {
			Logga(ctx, os.Getenv("JsonLog"), "overrdide 12")
			tntEnv.ApiHostGit = envRes.BodyJson["XKUBETEAMBRANCH12"].(string)
		}

		if envRes.BodyJson["XKUBETEAMBRANCH13"].(string) != "" {
			Logga(ctx, os.Getenv("JsonLog"), "overrdide 13")
			tntEnv.UserGit = envRes.BodyJson["XKUBETEAMBRANCH13"].(string)
		}

		if envRes.BodyJson["XKUBETEAMBRANCH14"].(string) != "" {
			Logga(ctx, os.Getenv("JsonLog"), "overrdide 14")
			tntEnv.TokenGit = envRes.BodyJson["XKUBETEAMBRANCH14"].(string)
		}

		if envRes.BodyJson["XKUBETEAMBRANCH15"].(string) != "" {
			Logga(ctx, os.Getenv("JsonLog"), "overrdide 15")
			tntEnv.WorkspaceGit = envRes.BodyJson["XKUBETEAMBRANCH15"].(string)
		}

		if envRes.BodyJson["XKUBETEAMBRANCH16"].(string) != "" {
			Logga(ctx, os.Getenv("JsonLog"), "overrdide 16")
			tntEnv.UrlGit = envRes.BodyJson["XKUBETEAMBRANCH16"].(string)
		}

		if envRes.BodyJson["XKUBETEAMBRANCH17"].(string) != "" {
			Logga(ctx, os.Getenv("JsonLog"), "overrdide 17")
			tntEnv.TypeGit = envRes.BodyJson["XKUBETEAMBRANCH17"].(string)
		}

		if envRes.BodyJson["XKUBETEAMBRANCH18"].(string) != "" {
			Logga(ctx, os.Getenv("JsonLog"), "overrdide 18")
			tntEnv.WorkspaceKey = envRes.BodyJson["XKUBETEAMBRANCH18"].(string)
		}
		if envRes.BodyJson["XKUBETEAMBRANCH19"].(string) != "" {
			Logga(ctx, os.Getenv("JsonLog"), "overrdide 19")
			tntEnv.WorkspaceSecret = envRes.BodyJson["XKUBETEAMBRANCH19"].(string)
		}
		if envRes.BodyJson["XKUBETEAMBRANCH20"].(string) != "" {
			Logga(ctx, os.Getenv("JsonLog"), "overrdide 20")
			tntEnv.WorkspaceRefreshToken = envRes.BodyJson["XKUBETEAMBRANCH20"].(string)
		}
		if envRes.BodyJson["XKUBETEAMBRANCH21"].(string) != "" {
			Logga(ctx, os.Getenv("JsonLog"), "overrdide 21")
			tntEnv.ProjectGit = envRes.BodyJson["XKUBETEAMBRANCH21"].(string)
		}

		Logga(ctx, os.Getenv("JsonLog"), "KUBETEAMBRANCH OK")
	} else {
		Logga(ctx, os.Getenv("JsonLog"), "KUBETEAMBRANCH MISSING")
		erro = errors.New("KUBETEAMBRANCH MISSING")
	}
	return tntEnv, sprintBranch, erro
}
func GetApiHostAndToken(ctx context.Context, enviro, cluster, token, apiDomain, coreApiVersion, swmono string) (string, string, error) {
	var erro error

	args := make(map[string]string)
	args["source"] = "devops-8"
	args["$select"] = "XKUBECLUSTER16,XKUBECLUSTER18"
	args["center_dett"] = "dettaglio"
	args["$filter"] = "equals(XKUBECLUSTER03,'" + cluster + "') "

	var endpointRes CallGetResponse
	var errendpointRes error
	if swmono == "mono" {
		endpointRes, errendpointRes = ApiCallGET(ctx, os.Getenv("RestyDebug"), args, "msdevopsmono", "/api/"+os.Getenv("coreApiVersion")+"/devopsmono/KUBECLUSTER", token, apiDomain, coreApiVersion)
	} else {
		endpointRes, errendpointRes = ApiCallGET(ctx, os.Getenv("RestyDebug"), args, "msdevops", "/api/"+os.Getenv("coreApiVersion")+"/devops/KUBECLUSTER", token, apiDomain, coreApiVersion)
	}

	if errendpointRes != nil {
		erro := errors.New(errendpointRes.Error())
		return "", "", erro
	}

	apiHost, apiToken := "", ""
	if len(endpointRes.BodyJson) > 0 {
		apiHost = endpointRes.BodyJson["XKUBECLUSTER16"].(string)
		apiToken = endpointRes.BodyJson["XKUBECLUSTER18"].(string)
	} else {
		erro := errors.New("nohost")
		return "", "", erro
	}

	return apiHost, apiToken, erro
}
func GeneraBuildDirectory(ctx context.Context, microservice, jobID string) (string, error) {

	/* sezione filesystem */

	Logga(ctx, os.Getenv("JsonLog"), "Sezione filesystem")

	var erro error

	ct := time.Now()
	now := ct.Format("20060102150405")
	nowNum, _ := strconv.Atoi(now)

	files, erro := os.ReadDir("/tmp/")
	if erro != nil {
		return "", erro
	}

	// verifico se devo cancellare 	qualche TEMP DIR
	for _, f := range files {
		if len(f.Name()) >= 6 {
			if f.Name()[:6] == "build_" {
				arr := strings.Split(f.Name(), "_")
				num, _ := strconv.Atoi(arr[1])
				diff := nowNum - num
				//fmt.Println(ctx, nowNum, num, diff)

				// se la diff >= 2 significa che la tmp dir in oggetto e piu vecchia di 2 ore e quindi va cancellata
				if diff >= 200 {
					erro = os.RemoveAll("/tmp/" + f.Name())
					Logga(ctx, os.Getenv("JsonLog"), "Delete /tmp/"+f.Name())
					if erro != nil {
						Logga(ctx, os.Getenv("JsonLog"), erro.Error(), "error")
						return "", erro
					}
				}
			}
		}
	}

	dirToCreate := "/tmp/build_" + now + "_" + microservice + "_" + jobID
	Logga(ctx, os.Getenv("JsonLog"), "Create "+dirToCreate)
	erro = os.MkdirAll(dirToCreate, 0755)
	if erro != nil {
		Logga(ctx, os.Getenv("JsonLog"), erro.Error())
		return "", erro
	}

	return dirToCreate, erro
}
func GetTeamFromGroup(ctx context.Context, devopsToken, dominio, group string) (string, error) {

	Logga(ctx, os.Getenv("JsonLog"), "Getting GRU ++")
	// ottengo da gru il nome del team
	argsGru := make(map[string]string)
	argsGru["source"] = "devops-8"
	argsGru["$select"] = "XGRU05"
	argsGru["center_dett"] = "dettaglio"
	argsGru["$filter"] = "equals(XGRU03,'" + group + "') "

	GruRes, errGruRes := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsGru, "msusers", "/api/"+os.Getenv("coreApiVersion")+"/users/GRU", devopsToken, dominio, os.Getenv("coreApiVersion"))
	if errGruRes != nil {
		Logga(ctx, os.Getenv("JsonLog"), errGruRes.Error())
		erro := errors.New(errGruRes.Error())
		return " ", erro
	}

	var team string
	var errcast bool
	if len(GruRes.BodyJson) > 0 {
		team, errcast = GruRes.BodyJson["XGRU05"].(string)
		if !errcast {
			Logga(ctx, os.Getenv("JsonLog"), "XGRU05 no cast")
			erro := errors.New("XGRU05 no cast")
			return " ", erro
		}
	}
	return strings.ToLower(team), nil
}
func GetSingleGroup(grpArrDirt []string) []string {
	var grpArrClean []string
	for _, grp := range grpArrDirt {
		inArr := false
		for _, y := range grpArrClean {
			if y == grp {
				inArr = true
				break
			}
		}
		if !inArr {
			grpArrClean = append(grpArrClean, grp)
		}
	}
	return grpArrClean
}
