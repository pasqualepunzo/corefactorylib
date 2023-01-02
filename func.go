package lib

import (
	"bytes"
	"context"
	"crypto/sha1"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
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

func Logga(ctx context.Context, i interface{}, level ...string) {

	// QUI SE VUOI VEDERE IL TESTO IN CHIARO
	//logtext := true
	logtext := false

	JobID := ""
	if ctx.Value("JobID") != nil {
		JobID = ctx.Value("JobID").(string)
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

	if logtext == true {
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
				}).Fatal(text)

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
	text += "\n"
	text += "\033[1;31m *********************************************************************************\n"
	text += " *  _____\n"
	text += " * | ____|_ __ _ __ ___  _ __ \n"
	text += " * |  _| | '__| '__/ _ \\| '__|\n"
	text += " * | |___| |  | | | (_) | |   \n"
	text += " * |_____|_|  |_|  \\___/|_|  \n"
	text += " * \033[1;0m\n"
	text += " *\n"
	text += " * " + errorLabel + "\n"
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

	comando := "gcloud config set project  " + clusterProject
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
func GetUserGroup(ctx context.Context, token, gruppo string) (map[string]string, LoggaErrore) {

	Logga(ctx, "Getting GRU")

	var loggaErrore LoggaErrore
	loggaErrore.Errore = 0

	args := make(map[string]string)
	args["center_dett"] = "dettaglio"
	args["source"] = "users-3"
	args["$select"] = "XGRU05,XGRU06"

	gruRes := ApiCallGET(ctx, false, args, "msusers", "/users/GRU/equals(XGRU03,'"+gruppo+"')", token, "")

	gru := make(map[string]string)

	if len(gruRes.BodyJson) > 0 {
		gru["gruppo"] = gruRes.BodyJson["XGRU05"].(string)
		gru["stage"] = gruRes.BodyJson["XGRU06"].(string)
		Logga(ctx, "GRU OK")
	} else {
		Logga(ctx, "GRU MISSING")
		loggaErrore.Errore = -1
		loggaErrore.Log = "GRU MISSING"
	}

	return gru, loggaErrore
}
func GetNextVersion(ctx context.Context, masterBranch, nomeDocker string) (string, LoggaErrore) {

	var loggaErrore LoggaErrore
	loggaErrore.Errore = 0

	// cerco il token di Corefactory
	Logga(ctx, "Getting token")
	devopsToken, erro := GetCoreFactoryToken(ctx)
	if erro.Errore < 0 {
		Logga(ctx, erro.Log)
	} else {
		Logga(ctx, "Token OK")
	}

	ct := time.Now()
	dateVers := ct.Format("060102")
	ver := ""
	/* ************************************************************************************************ */
	// KUBEIMICROSERV
	Logga(ctx, "Getting KUBEDKRBUILD - func.go 1")
	argsImicro := make(map[string]string)
	argsImicro["source"] = "devops-8"
	argsImicro["$select"] = "XKUBEDKRBUILD06"
	argsImicro["center_dett"] = "dettaglio"
	argsImicro["$filter"] = "startwith(XKUBEDKRBUILD06,'" + dateVers + "') "
	argsImicro["$filter"] += " and equals(XKUBEDKRBUILD03,'" + nomeDocker + "') "

	restyKubeImicroservRes := ApiCallGET(ctx, false, argsImicro, "msdevops", "/devops/KUBEDKRBUILD", devopsToken, "")
	if restyKubeImicroservRes.Errore < 0 {
		Logga(ctx, restyKubeImicroservRes.Log)
		loggaErrore.Errore = restyKubeImicroservRes.Errore
		loggaErrore.Log = restyKubeImicroservRes.Log
		return "", loggaErrore
	}

	if len(restyKubeImicroservRes.BodyJson) > 0 {
		ver = restyKubeImicroservRes.BodyJson["XKUBEDKRBUILD06"].(string)
		Logga(ctx, "KUBEDKRBUILD OK")
	} else {
		Logga(ctx, "KUBEDKRBUILD MISSING")
	}
	Logga(ctx, "")

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

	return ver, loggaErrore
}
func Times(str string, n int) string {
	if n <= 0 {
		return ""
	}
	return strings.Repeat(str, n)
}
func GetMicroserviceDetail(ctx context.Context, team, ims, gitDevMaster, buildVersion, devopsToken, autopilot string) (Microservice, LoggaErrore) {

	Logga(ctx, "")
	Logga(ctx, " + + + + + + + + + + + + + + + + + + + + ")
	Logga(ctx, "TEAM "+team)
	Logga(ctx, "IMS "+ims)
	Logga(ctx, "gitDevMaster "+gitDevMaster)
	Logga(ctx, "BUILDVERSION "+buildVersion)
	Logga(ctx, "getMicroserviceDetail begin")

	versioneArr := strings.Split(buildVersion, ".")
	versione := ""

	if len(versioneArr) > 1 {
		versione = versioneArr[0] + Times("0", 2-len(versioneArr[1])) + versioneArr[1] + Times("0", 2-len(versioneArr[2])) + versioneArr[2] + Times("0", 2-len(versioneArr[3])) + versioneArr[3]
	} else {
		versione = buildVersion
	}

	var microservices Microservice

	var loggaErrore LoggaErrore
	loggaErrore.Errore = 0
	loggaErrore.Log = ""

	/* ************************************************************************************************ */
	// KUBEIMICROSERV
	Logga(ctx, "Getting KUBEIMICROSERV")
	argsImicro := make(map[string]string)
	argsImicro["source"] = "devops-8"
	argsImicro["$select"] = "XKUBEIMICROSERV04,XKUBEIMICROSERV05,XKUBEIMICROSERV07"
	argsImicro["center_dett"] = "dettaglio"
	argsImicro["$filter"] = "equals(XKUBEIMICROSERV03,'" + ims + "') "

	restyKubeImicroservRes := ApiCallGET(ctx, false, argsImicro, "msdevops", "/devops/KUBEIMICROSERV", devopsToken, "")
	if restyKubeImicroservRes.Errore != 0 {
		Logga(ctx, restyKubeImicroservRes.Log)
		loggaErrore.Errore = restyKubeImicroservRes.Errore
		loggaErrore.Log = restyKubeImicroservRes.Log
		return microservices, loggaErrore
	}

	microservice := ""
	cluster := ""
	if len(restyKubeImicroservRes.BodyJson) > 0 {

		microservice = restyKubeImicroservRes.BodyJson["XKUBEIMICROSERV04_COD"].(string)
		microservices.VersMicroservice = restyKubeImicroservRes.BodyJson["XKUBEIMICROSERV07"].(string)
		cluster = restyKubeImicroservRes.BodyJson["XKUBEIMICROSERV05"].(string)
		Logga(ctx, "KUBEIMICROSERV OK")
	} else {
		Logga(ctx, "   !!!   KUBEIMICROSERV MISSING")
		loggaErrore.Errore = -1
		loggaErrore.Log = "KUBEIMICROSERV MISSING"
		return microservices, loggaErrore
	}
	Logga(ctx, "")

	/* ************************************************************************************************ */
	// KUBECLUSTER
	Logga(ctx, "Getting KUBECLUSTER")

	argsClu := make(map[string]string)
	argsClu["source"] = "devops-8"
	argsClu["$select"] = "XKUBECLUSTER06,XKUBECLUSTER12,XKUBECLUSTER15,XKUBECLUSTER17"
	argsClu["center_dett"] = "dettaglio"
	argsClu["$filter"] = "equals(XKUBECLUSTER03,'" + cluster + "') "

	restyKubeCluRes := ApiCallGET(ctx, false, argsClu, "msdevops", "/devops/KUBECLUSTER", devopsToken, "")
	if restyKubeCluRes.Errore < 0 {
		Logga(ctx, restyKubeCluRes.Log)
		loggaErrore.Errore = restyKubeCluRes.Errore
		loggaErrore.Log = restyKubeCluRes.Log
		//return ims, loggaErrore
	}

	profile := ""
	clusterOwner := ""
	clusterHost := ""
	// swMultiEnviro := ""
	var profileNum int
	if len(restyKubeCluRes.BodyJson) > 0 {

		clusterHost = restyKubeCluRes.BodyJson["XKUBECLUSTER15"].(string)
		clusterOwner = restyKubeCluRes.BodyJson["XKUBECLUSTER06"].(string)

		profileFloat := restyKubeCluRes.BodyJson["XKUBECLUSTER12"].(float64)
		profileNum = int(profileFloat)

		profile = strconv.Itoa(profileNum)

		// swMultiEnviro = restyKubeCluRes.BodyJson["XKUBECLUSTER17"].(string)

		/*
			il 24 / 11 / 2021 modifico la variabile profile da stringa a numero e nessuno di noi si ricorda perche con la FAC-442 è stata cambiata da int a string
			switch profileNum {
			case 0:
				profile = "dev"
				break
			case 1:
				profile = "qa"
				break
			case 2:
				profile = "prod"
				break
			}
		*/
		//profile = profile
		//profileInt = int32(profileNum)

		Logga(ctx, "KUBECLUSTER OK")
	} else {
		Logga(ctx, "   !!!   KUBECLUSTER MISSING")
	}
	Logga(ctx, "")
	/* ************************************************************************************************ */

	/* ************************************************************************************************ */
	// KUBEMICROSERV
	Logga(ctx, "Getting KUBEMICROSERV")

	argsMS := make(map[string]string)
	argsMS["source"] = "devops-8"
	argsMS["$select"] = "XKUBEMICROSERV03,XKUBEMICROSERV04,XKUBEMICROSERV05,XKUBEMICROSERV08,XKUBEMICROSERV16,XKUBEMICROSERV17,XKUBEMICROSERV18"
	argsMS["center_dett"] = "dettaglio"
	argsMS["$filter"] = "equals(XKUBEMICROSERV05,'" + microservice + "') "
	restyKubeMSRes := ApiCallGET(ctx, false, argsMS, "msdevops", "/devops/KUBEMICROSERV", devopsToken, "")
	if restyKubeMSRes.Errore != 0 {
		Logga(ctx, restyKubeMSRes.Log)
		loggaErrore.Errore = restyKubeMSRes.Errore
		loggaErrore.Log = restyKubeMSRes.Log
		return microservices, loggaErrore
	}

	hpaTmpl := ""
	if len(restyKubeMSRes.BodyJson) > 0 {
		microservices.Nome = restyKubeMSRes.BodyJson["XKUBEMICROSERV05"].(string)
		microservices.Descrizione = restyKubeMSRes.BodyJson["XKUBEMICROSERV03"].(string)
		microservices.Replicas = restyKubeMSRes.BodyJson["XKUBEMICROSERV18"].(string)
		microservices.Namespace = restyKubeMSRes.BodyJson["XKUBEMICROSERV04_COD"].(string)
		microservices.Virtualservice = strconv.FormatFloat(restyKubeMSRes.BodyJson["XKUBEMICROSERV08"].(float64), 'f', 0, 64)
		hpaTmpl = restyKubeMSRes.BodyJson["XKUBEMICROSERV16_COD"].(string)
		Logga(ctx, "KUBEMICROSERV OK")
	} else {
		Logga(ctx, "   !!!   KUBEMICROSERV MISSING")
		loggaErrore.Errore = 0
		loggaErrore.Log = "KUBEMICROSERV MISSING"
		//return microservices, loggaErrore
	}
	Logga(ctx, "")

	if autopilot != "1" {
		/* ************************************************************************************************ */
		// KUBEMICROSERVHPA
		Logga(ctx, "Getting KUBEMICROSERVHPA")
		argsHpa := make(map[string]string)
		argsHpa["source"] = "devops-8"
		argsHpa["$select"] = "XKUBEMICROSERVHPA04,XKUBEMICROSERVHPA05,XKUBEMICROSERVHPA06,XKUBEMICROSERVHPA07,XKUBEMICROSERVHPA08,XKUBEMICROSERVHPA09"
		argsHpa["center_dett"] = "dettaglio"
		argsHpa["$filter"] = "equals(XKUBEMICROSERVHPA03,'" + hpaTmpl + "') "

		restyKubeHpaRes := ApiCallGET(ctx, false, argsHpa, "msdevops", "/devops/KUBEMICROSERVHPA", devopsToken, "")
		if restyKubeHpaRes.Errore < 0 {
			Logga(ctx, restyKubeHpaRes.Log)
			loggaErrore.Errore = restyKubeHpaRes.Errore
			loggaErrore.Log = restyKubeHpaRes.Log
			return microservices, loggaErrore
		}

		if len(restyKubeHpaRes.BodyJson) > 0 {
			var hpa Hpa
			hpa.MinReplicas = strconv.FormatFloat(restyKubeHpaRes.BodyJson["XKUBEMICROSERVHPA04"].(float64), 'f', 0, 64)
			hpa.MaxReplicas = strconv.FormatFloat(restyKubeHpaRes.BodyJson["XKUBEMICROSERVHPA05"].(float64), 'f', 0, 64)
			hpa.CpuTipoTarget = restyKubeHpaRes.BodyJson["XKUBEMICROSERVHPA06"].(string)
			hpa.CpuTarget = restyKubeHpaRes.BodyJson["XKUBEMICROSERVHPA07"].(string)
			hpa.MemTipoTarget = restyKubeHpaRes.BodyJson["XKUBEMICROSERVHPA08"].(string)
			hpa.MemTarget = restyKubeHpaRes.BodyJson["XKUBEMICROSERVHPA09"].(string)
			microservices.Hpa = hpa
			Logga(ctx, "KUBEMICROSERVHPA OK")
		} else {
			Logga(ctx, "   !!!   KUBEMICROSERVHPA MISSING")
			loggaErrore.Errore = 0
			loggaErrore.Log = "KUBEMICROSERVHPA MISSING"
			//return microservices, loggaErrore
		}
		Logga(ctx, "")

		/* ************************************************************************************************ */
	}

	/* ************************************************************************************************ */
	// SELKUBEDKRLIST
	Logga(ctx, "Getting SELKUBEDKRLIST")
	argsDkr := make(map[string]string)
	argsDkr["center_dett"] = "visualizza"
	argsDkr["$filter"] = "equals(XSELKUBEDKRLIST10,'" + microservices.Nome + "') "

	restyDkrLstRes := ApiCallGET(ctx, false, argsDkr, "msdevops", "/core/custom/SELKUBEDKRLIST/values", devopsToken, "")
	if restyDkrLstRes.Errore != 0 {
		Logga(ctx, restyDkrLstRes.Log)
		loggaErrore.Errore = restyDkrLstRes.Errore
		loggaErrore.Log = restyDkrLstRes.Log
		return microservices, loggaErrore
	}

	if len(restyDkrLstRes.BodyArray) > 0 {
		var pods []Pod
		for _, x := range restyDkrLstRes.BodyArray {

			/* ************************************************************************************************ */

			var pod Pod

			pod.Docker = x["XSELKUBEDKRLIST03"].(string)
			docker := pod.Docker
			pod.GitRepo = x["XSELKUBEDKRLIST04"].(string)
			resourceTmpl := x["XSELKUBEDKRLIST05"].(string)
			pod.Descr = x["XSELKUBEDKRLIST06"].(string)
			pod.Dockerfile = x["XSELKUBEDKRLIST07"].(string)
			pod.Tipo = x["XSELKUBEDKRLIST08"].(string)
			pod.Vpn = int(x["XSELKUBEDKRLIST09"].(float64))
			pod.Workdir = x["XSELKUBEDKRLIST11"].(string)

			/* ************************************************************************************************ */
			// KUBEDKRBUILD
			Logga(ctx, "Getting KUBEDKRBUILD func.go 2")
			argsBld := make(map[string]string)

			argsDeploy := make(map[string]string)
			argsDeploy["source"] = "devops-8"

			argsBld["$fullquery"] = "select XKUBEDKRBUILD06,XKUBEDKRBUILD04,XKUBEDKRBUILD07,XKUBEDKRBUILD09,XKUBEDKRBUILD10,XKUBEDKRBUILD12,XKUBEDKRBUILD13 "
			argsBld["$fullquery"] += "from TB_ANAG_KUBEDKRBUILD00 "
			argsBld["$fullquery"] += "where 1>0 "
			argsBld["$fullquery"] += "AND XKUBEDKRBUILD03 = '" + docker + "' "
			argsBld["$fullquery"] += "AND XKUBEDKRBUILD08 = '" + team + "' "
			// if ftNewStageProcess_FAC530 {
			// 	argsBld["$fullquery"] += "AND XKUBEDKRBUILD15 = '" + enviro + "' "
			// } else {
			//argsBld["$fullquery"] += "AND XKUBEDKRBUILD04 = '" + gitDevMaster + "' "
			// }
			// FAC-462 argsBld["$fullquery"] += "AND XKUBEDKRBUILD08 = '" + team + "' "
			if versione != "" {
				argsBld["$fullquery"] += " AND XKUBEDKRBUILD06 = '" + versione + "' "
			}
			argsBld["$fullquery"] += " order by (case when XKUBEDKRBUILD04 = 'master' then 0 else 1 end ) desc,cast(XKUBEDKRBUILD06 as unsigned) DESC "
			argsBld["$fullquery"] += " limit 1 "
			fmt.Println(argsBld["$fullquery"])

			restyKubeBldRes := ApiCallGET(ctx, false, argsBld, "msdevops", "/core/custom/KUBEDKRBUILD/values", devopsToken, "")

			//fmt.Println(restyKubeBldRes)
			if restyKubeBldRes.Errore < 0 {
				//fmt.Println("A")
				Logga(ctx, restyKubeBldRes.Log)
				loggaErrore.Errore = restyKubeBldRes.Errore
				loggaErrore.Log = restyKubeBldRes.Log
				return microservices, loggaErrore
			}

			if len(restyKubeBldRes.BodyArray) > 0 {

				// fmt.Println("B")
				var branchs Branch
				branchs.Branch = restyKubeBldRes.BodyArray[0]["XKUBEDKRBUILD04"].(string)
				branchs.Version = restyKubeBldRes.BodyArray[0]["XKUBEDKRBUILD06"].(string)
				branchs.Sha = restyKubeBldRes.BodyArray[0]["XKUBEDKRBUILD07"].(string)

				var podBuild PodBuild
				podBuild.Versione = restyKubeBldRes.BodyArray[0]["XKUBEDKRBUILD06"].(string)
				podBuild.Merged = restyKubeBldRes.BodyArray[0]["XKUBEDKRBUILD13"].(string)
				podBuild.Tag = restyKubeBldRes.BodyArray[0]["XKUBEDKRBUILD09"].(string)
				podBuild.MasterDev = restyKubeBldRes.BodyArray[0]["XKUBEDKRBUILD04"].(string)
				podBuild.ReleaseNote = restyKubeBldRes.BodyArray[0]["XKUBEDKRBUILD12"].(string)
				podBuild.SprintBranch = restyKubeBldRes.BodyArray[0]["XKUBEDKRBUILD10"].(string)

				pod.PodBuild = podBuild
				pod.Branch = branchs
				Logga(ctx, "KUBEDKRBUILD OK")
			} else {
				// se manca la build alla versione indicata proviamo a cercare l'ultima
				// se manca anche questa allora errore mai fatta una build !!!!!

				Logga(ctx, "KUBEDKRBUILD MISSING ON "+versione+" seek for latest")
				argsBld := make(map[string]string)

				argsBld["$fullquery"] = "select XKUBEDKRBUILD06,XKUBEDKRBUILD04,XKUBEDKRBUILD07,XKUBEDKRBUILD09,XKUBEDKRBUILD10,XKUBEDKRBUILD12,XKUBEDKRBUILD13 "
				argsBld["$fullquery"] += "from TB_ANAG_KUBEDKRBUILD00 "
				argsBld["$fullquery"] += "where 1>0 "
				argsBld["$fullquery"] += "AND XKUBEDKRBUILD03 = '" + docker + "' "
				argsBld["$fullquery"] += "AND XKUBEDKRBUILD08 = '" + team + "' "
				argsBld["$fullquery"] += " order by (case when XKUBEDKRBUILD04 = 'master' then 0 else 1 end ) desc,cast(XKUBEDKRBUILD06 as unsigned) DESC "
				argsBld["$fullquery"] += " limit 1 "
				fmt.Println(argsBld["$fullquery"])

				restyKubeBldRes := ApiCallGET(ctx, false, argsBld, "msdevops", "/core/custom/KUBEDKRBUILD/values", devopsToken, "")

				//fmt.Println(restyKubeBldRes)
				if restyKubeBldRes.Errore < 0 {
					//fmt.Println("A")
					Logga(ctx, restyKubeBldRes.Log)
					loggaErrore.Errore = restyKubeBldRes.Errore
					loggaErrore.Log = restyKubeBldRes.Log
					return microservices, loggaErrore
				}

				if len(restyKubeBldRes.BodyArray) > 0 {

					// fmt.Println("B")
					var branchs Branch
					branchs.Branch = restyKubeBldRes.BodyArray[0]["XKUBEDKRBUILD04"].(string)
					branchs.Version = restyKubeBldRes.BodyArray[0]["XKUBEDKRBUILD06"].(string)
					branchs.Sha = restyKubeBldRes.BodyArray[0]["XKUBEDKRBUILD07"].(string)

					var podBuild PodBuild
					podBuild.Versione = restyKubeBldRes.BodyArray[0]["XKUBEDKRBUILD06"].(string)
					podBuild.Merged = restyKubeBldRes.BodyArray[0]["XKUBEDKRBUILD13"].(string)
					podBuild.Tag = restyKubeBldRes.BodyArray[0]["XKUBEDKRBUILD09"].(string)
					podBuild.MasterDev = restyKubeBldRes.BodyArray[0]["XKUBEDKRBUILD04"].(string)
					podBuild.ReleaseNote = restyKubeBldRes.BodyArray[0]["XKUBEDKRBUILD12"].(string)
					podBuild.SprintBranch = restyKubeBldRes.BodyArray[0]["XKUBEDKRBUILD10"].(string)

					pod.PodBuild = podBuild
					pod.Branch = branchs
					Logga(ctx, "KUBEDKRBUILD LATEST OK")
				} else {

					Logga(ctx, "   !!! "+docker+"  KUBEDKRBUILD MISSING")
					loggaErrore.Errore = -1
					loggaErrore.Log = "!!! ERROR !!!\n\nThe component " + docker + " of the microservice " + microservices.Nome + " is MISSING,\nyou have to build it first.\nbye\n\n"
					return microservices, loggaErrore
				}
			}

			Logga(ctx, "")

			/* ************************************************************************************************ */

			/* ************************************************************************************************ */
			// KUBEDKRMOUNT
			Logga(ctx, "Getting KUBEDKRMOUNT")
			argsMnt := make(map[string]string)
			argsMnt["source"] = "devops-8"
			argsMnt["$select"] = "XKUBEDKRMOUNT04,XKUBEDKRMOUNT05,XKUBEDKRMOUNT06,XKUBEDKRMOUNT07"
			argsMnt["center_dett"] = "visualizza"
			argsMnt["$filter"] = "equals(XKUBEDKRMOUNT03,'" + docker + "') "

			restyKubeMntRes := ApiCallGET(ctx, false, argsMnt, "msdevops", "/devops/KUBEDKRMOUNT", devopsToken, "")
			if restyKubeMntRes.Errore != 0 {
				Logga(ctx, restyKubeMntRes.Log)
				loggaErrore.Errore = restyKubeMntRes.Errore
				loggaErrore.Log = restyKubeMntRes.Log
				return microservices, loggaErrore
			}

			if len(restyKubeMntRes.BodyArray) > 0 {
				var mounts []Mount
				for _, x := range restyKubeMntRes.BodyArray {

					var mount Mount
					mount.Nome = x["XKUBEDKRMOUNT04"].(string)
					mount.Mount = x["XKUBEDKRMOUNT05"].(string)
					mount.Subpath = x["XKUBEDKRMOUNT06"].(string)
					mount.ClaimName = x["XKUBEDKRMOUNT07"].(string)

					mounts = append(mounts, mount)
				}
				pod.Mount = mounts
				Logga(ctx, "KUBEDKRMOUNT OK")
			} else {
				loggaErrore.Errore = 0
				loggaErrore.Log = "KUBEDKRMOUNT MISSING"
				//return microservices, loggaErrore
			}
			Logga(ctx, "")

			/* ************************************************************************************************ */

			/* ************************************************************************************************ */
			// KUBEDKRRESOURCE
			Logga(ctx, "Getting KUBEDKRRESOURCE")
			argsSrc := make(map[string]string)
			argsSrc["source"] = "devops-8"
			argsSrc["$select"] = "XKUBEDKRRESOURCE04,XKUBEDKRRESOURCE05,XKUBEDKRRESOURCE06,XKUBEDKRRESOURCE07"
			argsSrc["center_dett"] = "dettaglio"
			argsSrc["$filter"] = "equals(XKUBEDKRRESOURCE03,'" + resourceTmpl + "') "

			restyKubeSrcRes := ApiCallGET(ctx, false, argsSrc, "msdevops", "/devops/KUBEDKRRESOURCE", devopsToken, "")
			if restyKubeSrcRes.Errore < 0 {
				Logga(ctx, restyKubeSrcRes.Log)
				loggaErrore.Errore = restyKubeSrcRes.Errore
				loggaErrore.Log = restyKubeSrcRes.Log
				return microservices, loggaErrore
			}

			if len(restyKubeSrcRes.BodyJson) > 0 {
				var resource Resource

				resource.CpuReq = restyKubeSrcRes.BodyJson["XKUBEDKRRESOURCE04"].(string) //   -- cpu res
				resource.MemReq = restyKubeSrcRes.BodyJson["XKUBEDKRRESOURCE05"].(string) //   -- mem res
				resource.CpuLim = restyKubeSrcRes.BodyJson["XKUBEDKRRESOURCE06"].(string) //   -- cpu limit
				resource.MemLim = restyKubeSrcRes.BodyJson["XKUBEDKRRESOURCE07"].(string) //   -- mem limit

				pod.Resource = resource
				Logga(ctx, "KUBEDKRRESOURCE OK")
			} else {
				loggaErrore.Errore = -1
				loggaErrore.Log = "KUBEDKRRESOURCE MISSING"
				return microservices, loggaErrore
			}
			Logga(ctx, "")

			/* ************************************************************************************************ */

			/* ************************************************************************************************ */

			// KUBEDKRPROBE
			Logga(ctx, "Getting KUBEDKRPROBE")
			argsProbes := make(map[string]string)
			argsProbes["source"] = "devops-8"
			//argsProbes["$select"] = "XKUBEDKRPROBE04,XKUBEDKRPROBE05,XKUBEDKRPROBE06,XKUBEDKRPROBE07,XKUBEDKRPROBE08,XKUBEDKRPROBE09,XKUBEDKRPROBE10"
			//argsProbes["$select"] += "XKUBEDKRPROBE11,XKUBEDKRPROBE12,XKUBEDKRPROBE13,XKUBEDKRPROBE14,XKUBEDKRPROBE15,XKUBEDKRPROBE16,XKUBEDKRPROBE17,XKUBEDKRPROBE18,XKUBEDKRPROBE19"
			argsProbes["center_dett"] = "allviews"
			argsProbes["$filter"] = "equals(XKUBEDKRPROBE03,'" + docker + "') "

			restyKubePrbRes := ApiCallGET(ctx, false, argsProbes, "msdevops", "/devops/KUBEDKRPROBE", devopsToken, "")
			if restyKubePrbRes.Errore < 0 {
				Logga(ctx, restyKubePrbRes.Log)
				loggaErrore.Errore = restyKubePrbRes.Errore
				loggaErrore.Log = restyKubePrbRes.Log
				return microservices, loggaErrore
			}

			if len(restyKubePrbRes.BodyArray) > 0 {

				var probes []Probes
				for _, x := range restyKubePrbRes.BodyArray {

					var elem Probes

					elem.Category = x["XKUBEDKRPROBE04"].(string)
					elem.Type = x["XKUBEDKRPROBE05"].(string)
					if x["XKUBEDKRPROBE06"] == nil {
						elem.Command = ""
					} else {
						elem.Command = x["XKUBEDKRPROBE06"].(string)
					}
					elem.HttpHost = x["XKUBEDKRPROBE07"].(string)
					elem.HttpPort = int(x["XKUBEDKRPROBE08"].(float64))
					elem.HttpPath = x["XKUBEDKRPROBE09"].(string)
					if x["XKUBEDKRPROBE10"] == nil {
						elem.HttpHeaders = ""
					} else {
						elem.HttpHeaders = x["XKUBEDKRPROBE10"].(string)
					}
					elem.HttpScheme = x["XKUBEDKRPROBE11"].(string)
					elem.TcpHost = x["XKUBEDKRPROBE12"].(string)
					elem.TcpPort = int(x["XKUBEDKRPROBE13"].(float64))
					elem.GrpcPort = int(x["XKUBEDKRPROBE14"].(float64))
					elem.InitialDelaySeconds = int(x["XKUBEDKRPROBE15"].(float64))
					elem.PeriodSeconds = int(x["XKUBEDKRPROBE16"].(float64))
					elem.TimeoutSeconds = int(x["XKUBEDKRPROBE17"].(float64))
					elem.SuccessThreshold = int(x["XKUBEDKRPROBE18"].(float64))
					elem.FailureThreshold = int(x["XKUBEDKRPROBE19"].(float64))
					probes = append(probes, elem)

				}
				pod.Probes = probes

				//Logga(ctx, probes)
				Logga(ctx, "KUBEDKRPROBE OK")
			} else {
				Logga(ctx, "KUBEDKRPROBE MISSING")
			}

			Logga(ctx, "")

			/* ************************************************************************************************ */

			/* ************************************************************************************************ */
			// KUBESERVICEDKR
			Logga(ctx, "Getting KUBESERVICEDKR")
			argsSrvDkr := make(map[string]string)
			argsSrvDkr["source"] = "devops-8"
			argsSrvDkr["$select"] = "XKUBESERVICEDKR06,XKUBESERVICEDKR05"
			argsSrvDkr["center_dett"] = "visualizza"
			argsSrvDkr["$filter"] = "equals(XKUBESERVICEDKR04,'" + docker + "') "

			restyKubeSrvDkrRes := ApiCallGET(ctx, false, argsSrvDkr, "msdevops", "/devops/KUBESERVICEDKR", devopsToken, "")
			if restyKubeSrvDkrRes.Errore != 0 {
				Logga(ctx, restyKubeSrvDkrRes.Log)
				loggaErrore.Errore = restyKubeSrvDkrRes.Errore
				loggaErrore.Log = restyKubeSrvDkrRes.Log
				return microservices, loggaErrore
			}

			if len(restyKubeSrvDkrRes.BodyArray) > 0 {
				var port, tipo string
				var services []Service
				for _, x := range restyKubeSrvDkrRes.BodyArray {

					port = strconv.FormatFloat(x["XKUBESERVICEDKR06"].(float64), 'f', 0, 64)
					tipo = x["XKUBESERVICEDKR05"].(string)

					/* ************************************************************************************************ */
					// ENDPOINTS
					Logga(ctx, "ENDPOINTS")
					sqlEndpoint := ""

					// per ogni servizio cerco gli endpoints
					sqlEndpoint += "select * from ( "

					sqlEndpoint += "select "
					sqlEndpoint += "ifnull(aa.XKUBEENDPOINT05,'') as microservice_src, "
					sqlEndpoint += "ifnull(cc.XKUBESERVICEDKR04,'') as docker_src, "
					sqlEndpoint += "ifnull(aa.XKUBEENDPOINT10,'') as type_src, "
					sqlEndpoint += "ifnull(aa.XKUBEENDPOINT09,'') as route_src, "
					sqlEndpoint += "ifnull(aa.XKUBEENDPOINT11,'') as rewrite_src, "
					sqlEndpoint += "ifnull(cc.XKUBESERVICEDKR06,'') as port_src, "
					sqlEndpoint += "ifnull(aa.XKUBEENDPOINT12,'') as priority, "
					sqlEndpoint += "ifnull(bb.XKUBEENDPOINT05,'') as microservice_dst, "
					sqlEndpoint += "ifnull(bb.XKUBESERVICEDKR03,'') as docker_dst, "
					sqlEndpoint += "ifnull(bb.XKUBEENDPOINT10,'') as type_dst, "
					sqlEndpoint += "ifnull(bb.XKUBEENDPOINT09,'') as route_dst, "
					sqlEndpoint += "ifnull(bb.XKUBEENDPOINT11,'') as rewrite_dst, "
					sqlEndpoint += "ifnull(bb.XKUBEMICROSERV04,'') as namespace_dst, "
					sqlEndpoint += "ifnull(bb.XDEPLOYLOG05,'') as version_dst, "
					sqlEndpoint += "ifnull(bb.XKUBECLUSTER15,'') as domain, "
					sqlEndpoint += "ifnull(bb.XKUBEENDPOINTOVR06,'') as market, "
					sqlEndpoint += "ifnull(bb.XKUBEENDPOINTOVR07,'') as partner, "
					sqlEndpoint += "ifnull(bb.XKUBEENDPOINTOVR08,'') as customer "
					sqlEndpoint += "from TB_ANAG_KUBEENDPOINT00 aa "
					sqlEndpoint += "JOIN TB_ANAG_KUBESERVICEDKR00 cc on (cc.XKUBESERVICEDKR03=aa.XKUBEENDPOINT06) "
					sqlEndpoint += "left join "
					sqlEndpoint += "( "
					sqlEndpoint += "select XKUBECLUSTER15, XKUBEENDPOINT03,XKUBEENDPOINT09,XKUBEENDPOINT10,XKUBEENDPOINT11,XKUBEENDPOINT05,XKUBEMICROSERV05, "
					sqlEndpoint += "XKUBEMICROSERV04,XKUBEENDPOINTOVR03,XKUBESERVICEDKR04,XKUBESERVICEDKR03,XDEPLOYLOG05, "
					sqlEndpoint += "XKUBEENDPOINTOVR06,XKUBEENDPOINTOVR07,XKUBEENDPOINTOVR08 "
					sqlEndpoint += "from TB_ANAG_KUBEENDPOINT00 a "
					sqlEndpoint += "join TB_ANAG_KUBEENDPOINTOVR00 b on (a.XKUBEENDPOINT03=b.XKUBEENDPOINTOVR04) "
					sqlEndpoint += "join TB_ANAG_KUBEMICROSERV00 on (XKUBEMICROSERV05=XKUBEENDPOINT05) "
					sqlEndpoint += "join TB_ANAG_KUBESERVICEDKR00 on (XKUBESERVICEDKR03=XKUBEENDPOINT06) "
					sqlEndpoint += "JOIN TB_ANAG_KUBEIMICROSERV00 on (XKUBEENDPOINT05=XKUBEIMICROSERV04  and XKUBEIMICROSERV05 = '" + cluster + "' ) "
					sqlEndpoint += "JOIN TB_ANAG_KUBECLUSTER00 on(XKUBECLUSTER03 = XKUBEIMICROSERV05 and XKUBECLUSTER12 = '" + profile + "') "
					sqlEndpoint += "JOIN TB_ANAG_DEPLOYLOG00 on (XDEPLOYLOG04=XKUBEIMICROSERV03 "
					sqlEndpoint += "and XDEPLOYLOG03='production' "
					sqlEndpoint += "and XDEPLOYLOG06=1 and XDEPLOYLOG07=0) ) bb "
					sqlEndpoint += "on (aa.XKUBEENDPOINT03 = bb.XKUBEENDPOINTOVR03 ) "
					sqlEndpoint += "having 1>0 "
					sqlEndpoint += "and docker_src = '" + docker + "'"
					sqlEndpoint += "and port_src = '" + port + "'"
					sqlEndpoint += "union "
					sqlEndpoint += "select "
					sqlEndpoint += "ifnull(aa.XKUBEENDPOINT05,'') as microservice_src, "
					sqlEndpoint += "ifnull(cc.XKUBESERVICEDKR04,'') as docker_src, "
					sqlEndpoint += "ifnull(aa.XKUBEENDPOINT10,'') as type_src, "
					sqlEndpoint += "ifnull(aa.XKUBEENDPOINT09,'') as route_src, "
					sqlEndpoint += "ifnull(aa.XKUBEENDPOINT11,'') as rewrite_src, "
					sqlEndpoint += "ifnull(cc.XKUBESERVICEDKR06,'') as port_src, "
					sqlEndpoint += "ifnull(aa.XKUBEENDPOINT12,'') as priority, "
					sqlEndpoint += "ifnull(bb.XKUBEENDPOINT05,'') as microservice_dst, "
					sqlEndpoint += "ifnull(bb.XKUBESERVICEDKR03,'') as docker_dst, "
					sqlEndpoint += "ifnull(bb.XKUBEENDPOINT10,'') as type_dst, "
					sqlEndpoint += "ifnull(bb.XKUBEENDPOINT09,'') as route_dst, "
					sqlEndpoint += "ifnull(bb.XKUBEENDPOINT11,'') as rewrite_dst, "
					sqlEndpoint += "ifnull(bb.XKUBEMICROSERV04,'') as namespace_dst, "
					sqlEndpoint += "ifnull(bb.XDEPLOYLOG05,'') as version_dst, "
					sqlEndpoint += "ifnull(bb.XKUBECLUSTER15,'') as domain, "
					sqlEndpoint += "ifnull(bb.XKUBEENDPOINTOVR06,'') as market, "
					sqlEndpoint += "ifnull(bb.XKUBEENDPOINTOVR07,'') as partner, "
					sqlEndpoint += "ifnull(bb.XKUBEENDPOINTOVR08,'') as customer "
					sqlEndpoint += "from TB_ANAG_KUBEENDPOINT00 aa "
					sqlEndpoint += "JOIN TB_ANAG_KUBESERVICEDKR00 cc on (cc.XKUBESERVICEDKR03=aa.XKUBEENDPOINT06) "
					sqlEndpoint += "join "
					sqlEndpoint += "( "
					sqlEndpoint += "select XKUBECLUSTER15, XKUBEENDPOINT03,XKUBEENDPOINT09,XKUBEENDPOINT10,XKUBEENDPOINT11,XKUBEENDPOINT05,XKUBEMICROSERV05, "
					sqlEndpoint += "XKUBEMICROSERV04,XKUBEENDPOINTOVR03,XKUBESERVICEDKR04,XKUBESERVICEDKR03,XDEPLOYLOG05, "
					sqlEndpoint += "XKUBEENDPOINTOVR06,XKUBEENDPOINTOVR07,XKUBEENDPOINTOVR08 "
					sqlEndpoint += "from TB_ANAG_KUBEENDPOINT00 a "
					sqlEndpoint += "join TB_ANAG_KUBEENDPOINTOVR00 b on (a.XKUBEENDPOINT03=b.XKUBEENDPOINTOVR04) "
					sqlEndpoint += "join TB_ANAG_KUBEMICROSERV00 on (XKUBEMICROSERV05=XKUBEENDPOINT05) "
					sqlEndpoint += "join TB_ANAG_KUBESERVICEDKR00 on (XKUBESERVICEDKR03=XKUBEENDPOINT06) "
					sqlEndpoint += "JOIN TB_ANAG_KUBEIMICROSERV00 on (XKUBEENDPOINT05=XKUBEIMICROSERV04) "
					sqlEndpoint += "JOIN TB_ANAG_KUBECLUSTER00 on(XKUBECLUSTER03 = XKUBEIMICROSERV05 and XKUBECLUSTER12 = '2' and XKUBECLUSTER06 != '" + clusterOwner + "') "
					sqlEndpoint += "JOIN TB_ANAG_DEPLOYLOG00 on (XDEPLOYLOG04=XKUBEIMICROSERV03 "
					sqlEndpoint += "and XDEPLOYLOG03='production' "
					sqlEndpoint += "and XDEPLOYLOG06=1 and XDEPLOYLOG07=0) ) bb "
					sqlEndpoint += "on (aa.XKUBEENDPOINT03 = bb.XKUBEENDPOINTOVR03 ) "
					sqlEndpoint += "having 1>0 "
					sqlEndpoint += "and docker_src = '" + docker + "'"
					sqlEndpoint += "and port_src = '" + port + "'"
					sqlEndpoint += ") as tbl "
					sqlEndpoint += "order by length(priority), priority, route_src, customer desc , partner desc , market desc, "
					sqlEndpoint += "(case when domain = '" + clusterHost + "' then 0 else 1 end) asc "

					//fmt.Println(sqlEndpoint)
					//	os.Exit(0)

					argsEndpoint := make(map[string]string)
					argsEndpoint["source"] = "devops-8"
					argsEndpoint["$fullquery"] = sqlEndpoint

					restyKubeEndpointRes := ApiCallGET(ctx, false, argsEndpoint, "msdevops", "/core/custom/KUBEENDPOINT/values", devopsToken, "")
					if restyKubeEndpointRes.Errore < 0 {

					}

					var endpoints []Endpoint
					if len(restyKubeEndpointRes.BodyArray) > 0 {
						for _, x := range restyKubeEndpointRes.BodyArray {

							var endpoint Endpoint

							if (x["domain"].(string) == "ketch-app.com" || x["domain"].(string) == "labketch-app.it") && x["microservice_dst"].(string) == "mscoreservice" {

							} else {

								endpoint.MicroserviceDst = x["microservice_dst"].(string)
								endpoint.DockerDst = x["docker_dst"].(string)
								endpoint.TypeSrvDst = x["type_dst"].(string)
								endpoint.RouteDst = x["route_dst"].(string)
								endpoint.RewriteDst = x["rewrite_dst"].(string)
								endpoint.NamespaceDst = x["namespace_dst"].(string)
								endpoint.VersionDst = x["version_dst"].(string)

								endpoint.MicroserviceSrc = x["microservice_src"].(string)
								endpoint.DockerSrc = x["docker_src"].(string)
								endpoint.TypeSrvSrc = x["type_src"].(string)
								endpoint.RouteSrc = x["route_src"].(string)
								endpoint.RewriteSrc = x["rewrite_src"].(string)
								endpoint.NamespaceSrc = microservices.Namespace
								endpoint.VersionSrc = ""

								endpoint.Domain = x["domain"].(string)
								endpoint.Market = x["market"].(string)
								endpoint.Partner = x["partner"].(string)
								endpoint.Customer = x["customer"].(string)

								endpoints = append(endpoints, endpoint)

							}
						}
						Logga(ctx, "ENDPOINTS OK")
					} else {
						// loggaErrore.Errore = 0
						// loggaErrore.Log = "ENDPOINTS MISSING"
						Logga(ctx, "ENDPOINTS MISSING")
					}
					Logga(ctx, "")

					/* ************************************************************************************************ */

					var service Service
					service.Port = port
					service.Tipo = tipo
					service.Endpoint = endpoints

					services = append(services, service)

				}

				pod.Service = services

				Logga(ctx, "KUBESERVICEDKR OK")
			} else {
				loggaErrore.Errore = 0
				loggaErrore.Log = "KUBESERVICEDKR MISSING"
				// return microservices, loggaErrore
			}
			Logga(ctx, "")

			// aggiungo pod corrente a pods
			pods = append(pods, pod)

			microservices.Pod = pods

		}

		Logga(ctx, "SELKUBEDKRLIST OK")
	} else {
		loggaErrore.Errore = 0
		loggaErrore.Log = "SELKUBEDKRLIST MISSING"
		return microservices, loggaErrore
	}
	Logga(ctx, "")

	//LogJson(microservices)
	Logga(ctx, "Seek Microservice details ok")
	Logga(ctx, " - - - - - - - - - - - - - - -  ")
	Logga(ctx, "")
	// fmt.Println(microservices)
	// LogJson(microservices)
	// os.Exit(0)
	return microservices, loggaErrore
}
func GetPrifileInfo(ctx context.Context, token string) (map[string]interface{}, string) {

	Logga(ctx, "Getting getProfileInfo")

	info := make(map[string]interface{})

	args := make(map[string]string)
	infoRes := ApiCallGET(ctx, false, args, "mscore", "/core/getProfileInfo", token, "")

	erro := ""

	if len(infoRes.BodyJson) > 0 {
		restyProfileInfoResponse := ProfileInfo{}

		b, _ := json.Marshal(infoRes.BodyJson)
		json.Unmarshal(b, &restyProfileInfoResponse)

		info["market"] = restyProfileInfoResponse.Session.Market.Decval
		info["gruppo"] = restyProfileInfoResponse.Session.GrantSession.Gruppo
		info["nome"] = restyProfileInfoResponse.Session.GrantSession.NomeCognome
		info["email"] = restyProfileInfoResponse.Session.GrantSession.Email

		Logga(ctx, "GetProfileInfo OK")
	} else {
		erro = "-1"
		Logga(ctx, "GetProfileInfo MISSING")
	}

	return info, erro
}
func GetBuildLastTag(ctx context.Context, team, docker, tipo string) (string, LoggaErrore) {

	sprint, erro := GetCurrentBranchSprint(ctx, team, tipo)
	if erro.Errore < 0 {
		Logga(ctx, erro.Log)
	}

	var loggaErrore LoggaErrore
	loggaErrore.Errore = 0

	// cerco il token di Corefactory
	Logga(ctx, "Getting token")
	devopsToken, erro := GetCoreFactoryToken(ctx)
	if erro.Errore < 0 {
		Logga(ctx, erro.Log)
	} else {
		Logga(ctx, "Token OK")
	}

	/* ************************************************************************************************ */
	// KUBEIMICROSERV
	Logga(ctx, "Getting KUBEDKRBUILD - func.go 1")
	argsImicro := make(map[string]string)
	argsImicro["source"] = "devops-8"
	argsImicro["$select"] = "XKUBEDKRBUILD09"
	argsImicro["center_dett"] = "dettaglio"
	argsImicro["$filter"] = "equals(XKUBEDKRBUILD03,'" + docker + "') "
	argsImicro["$filter"] += " and equals(XKUBEDKRBUILD08,'" + team + "') "
	argsImicro["$filter"] += " and equals(XKUBEDKRBUILD10,'" + sprint + "') "
	argsImicro["$order"] = "CDATA desc"
	argsImicro["num_rows"] = " 1 "

	restyKubeImicroservRes := ApiCallGET(ctx, false, argsImicro, "msdevops", "/devops/KUBEDKRBUILD", devopsToken, "")
	if restyKubeImicroservRes.Errore < 0 {
		Logga(ctx, restyKubeImicroservRes.Log)
		loggaErrore.Errore = restyKubeImicroservRes.Errore
		loggaErrore.Log = restyKubeImicroservRes.Log
		return "", loggaErrore
	}

	tag := ""
	if len(restyKubeImicroservRes.BodyJson) > 0 {

		tag = restyKubeImicroservRes.BodyJson["XKUBEDKRBUILD09"].(string)
		Logga(ctx, "KUBEDKRBUILD OK")
	} else {
		Logga(ctx, "KUBEDKRBUILD MISSING")
	}
	Logga(ctx, "")
	//	fmt.Println(tag)
	/* ************************************************************************************************ */

	return tag, loggaErrore
}
func GetCurrentBranchSprint(ctx context.Context, team, tipo string) (string, LoggaErrore) {

	var loggaErrore LoggaErrore
	loggaErrore.Errore = 0

	// cerco il token di Corefactory
	Logga(ctx, "Getting token")
	devopsToken, erro := GetCoreFactoryToken(ctx)
	if erro.Errore < 0 {
		Logga(ctx, erro.Log)
	} else {
		Logga(ctx, "Token OK")
	}

	/* ************************************************************************************************ */
	// KUBEIMICROSERV
	Logga(ctx, "Getting KUBETEAMBRANCH - func.go")
	argsImicro := make(map[string]string)
	argsImicro["source"] = "devops-8"
	argsImicro["$select"] = "XKUBETEAMBRANCH05"
	argsImicro["center_dett"] = "dettaglio"
	argsImicro["$filter"] = "equals(XKUBETEAMBRANCH03,'" + team + "') "
	argsImicro["$filter"] += " and equals(XKUBETEAMBRANCH04,'" + tipo + "') "

	restyKubeImicroservRes := ApiCallGET(ctx, false, argsImicro, "msdevops", "/devops/KUBETEAMBRANCH", devopsToken, "")
	if restyKubeImicroservRes.Errore < 0 {
		loggaErrore.Errore = restyKubeImicroservRes.Errore
		loggaErrore.Log = restyKubeImicroservRes.Log
		return "", loggaErrore
	}

	sprintBranch := ""
	if len(restyKubeImicroservRes.BodyJson) > 0 {
		sprintBranch = restyKubeImicroservRes.BodyJson["XKUBETEAMBRANCH05"].(string)
		Logga(ctx, "KUBETEAMBRANCH OK")
	} else {
		Logga(ctx, "KUBETEAMBRANCH MISSING - getCurrentBranchSprint")
	}
	Logga(ctx, "")
	/* ************************************************************************************************ */

	return sprintBranch, loggaErrore
}
func CreateTag(ctx context.Context, branch, tag, repo string) {

	// OTTENGO L' HASH del branch vivo
	clientBranch := resty.New()
	respBranch, errBranch := clientBranch.R().
		EnableTrace().
		SetBasicAuth(os.Getenv("bitbucketUser"), os.Getenv("bitbucketToken")).
		Get(os.Getenv("bitbucketHost") + "/repositories/" + os.Getenv("bitbucketProject") + "/" + repo + "/refs/branches/" + branch)

	if errBranch != nil {
		Logga(ctx, errBranch.Error())
	}

	var branchRes BranchResStruct
	err := json.Unmarshal(respBranch.Body(), &branchRes)
	if err != nil {
		fmt.Println(err.Error())
	}
	//fmt.Println(branchRes.Target.Hash)

	// STACCO IL TAG
	body := `{"name": "` + tag + `","target": {  "hash": "` + branchRes.Target.Hash + `"}}`

	client := resty.New()
	client.Debug = false
	restyResponse, errTag := client.R().
		SetHeader("Content-Type", "application/json").
		SetBasicAuth(os.Getenv("bitbucketUser"), os.Getenv("bitbucketToken")).
		SetBody(body).
		Post(os.Getenv("bitbucketHost") + "/repositories/" + os.Getenv("bitbucketProject") + "/" + repo + "/refs/tags")

	if errTag != nil {
		Logga(ctx, errTag.Error())
	}
	//fmt.Println(restyResponse)

	var tagCreateRes TagCreateResStruct
	err = json.Unmarshal(restyResponse.Body(), &tagCreateRes)
	if err != nil {
		fmt.Println(err.Error())
	}
	if tagCreateRes.Type == "error" {
		fmt.Println(repo, tagCreateRes.Error.Message)
	}
}
func GetFutureToggle(ctx context.Context, cluster string) (bool, LoggaErrore) {

	var loggaErrore LoggaErrore
	loggaErrore.Errore = 0

	// cerco il token di Corefactory
	Logga(ctx, "Getting token")
	devopsToken, erro := GetCoreFactoryToken(ctx)
	if erro.Errore < 0 {
		Logga(ctx, erro.Log)
	} else {
		Logga(ctx, "Token OK")
	}

	/* ************************************************************************************************ */
	// CLUSTER

	Logga(ctx, "Getting CLUSTER")

	argsCLUSTER := make(map[string]string)
	argsCLUSTER["source"] = "devops-8"
	argsCLUSTER["$select"] = "XKUBECLUSTER17"
	argsCLUSTER["center_dett"] = "dettaglio"
	argsCLUSTER["$filter"] = "equals(XKUBECLUSTER03,'" + cluster + "') "
	restyCLUSTERRes := ApiCallGET(ctx, false, argsCLUSTER, "msdevops", "/core/KUBECLUSTER", devopsToken, "")
	if restyCLUSTERRes.Errore < 0 {
		Logga(ctx, restyCLUSTERRes.Log)
		loggaErrore.Errore = restyCLUSTERRes.Errore
		loggaErrore.Log = restyCLUSTERRes.Log
		return false, loggaErrore
	}

	var sw bool
	if len(restyCLUSTERRes.BodyJson) > 0 {
		swStr := restyCLUSTERRes.BodyJson["XKUBECLUSTER17"].(string)
		sw, _ = strconv.ParseBool(swStr)

		Logga(ctx, "CLUSTER OK")
	} else {
		Logga(ctx, "CLUSTER MISSING")
	}
	Logga(ctx, "")
	/* ************************************************************************************************ */

	// PORCATA PER FATICARE AL VOLO SU KEEPUP-STAGE
	// sw = true

	return sw, loggaErrore
}
func GetEnvironmentStatus(ctx context.Context, cluster, enviro, microserice, customer string) LoggaErrore {

	var loggaErrore LoggaErrore
	loggaErrore.Errore = 0

	status := ""

	// cerco il token di Corefactory
	devopsToken, erro := GetCoreFactoryToken(ctx)
	if erro.Errore < 0 {
		Logga(ctx, erro.Log)
	}

	/* ************************************************************************************************ */
	// KUBEENVSTATUS
	Logga(ctx, "Getting KUBEENVSTATUS")
	args := make(map[string]string)
	args["source"] = "devops-8"

	/* ************************************************************************************************ */

	Logga(ctx, "Getting KUBEENVSTATUS")

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

	restyEsRes := ApiCallGET(ctx, false, argsEs, "msdevops", "/core/KUBEENVSTATUS", devopsToken, "")
	if restyEsRes.Errore < 0 {
		Logga(ctx, restyEsRes.Log)
		loggaErrore.Errore = restyEsRes.Errore
		loggaErrore.Log = restyEsRes.Log
	}

	if len(restyEsRes.BodyJson) > 0 {
		status = strconv.Itoa(int(restyEsRes.BodyJson["XKUBEENVSTATUS07"].(float64)))
		loggaErrore.Log = status
		Logga(ctx, "KUBEENVSTATUS OK")
	} else {
		loggaErrore.Log = ""
		Logga(ctx, "KUBEENVSTATUS MISSING")
	}
	Logga(ctx, "")
	/* ************************************************************************************************ */

	return loggaErrore
}
func SetEnvironmentStatus(ctx context.Context, cluster, enviro, microserice, customer, user, toggle string) LoggaErrore {

	var loggaErrore LoggaErrore
	loggaErrore.Errore = 0

	// cerco il token di Corefactory
	devopsToken, erro := GetCoreFactoryToken(ctx)
	if erro.Errore < 0 {
		Logga(ctx, erro.Log)
		loggaErrore.Errore = erro.Errore
		loggaErrore.Log = erro.Log
	}

	/* ************************************************************************************************ */
	// KUBEENVSTATUS
	Logga(ctx, "Setting KUBEENVSTATUS")
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

	res := ApiCallPOST(ctx, false, keyvalueslices, "msdevops", "/devops/KUBEENVSTATUS", devopsToken, "")

	if res.Errore < 0 {
		Logga(ctx, res.Log)
		loggaErrore.Errore = res.Errore
		loggaErrore.Log = res.Log
	} else {
		if toggle == "ON" {
			loggaErrore.Log = "Environment set to LOCK"
		} else {
			loggaErrore.Log = "Environment set to UNLOCK"
		}
	}

	/* ************************************************************************************************ */

	return loggaErrore
}
func GetVersionPreviousStage(ctx context.Context, cluster, enviro, istanza, devopsToken, clusterDev string) (string, string) {

	nomeStage := ""
	versione := ""
	istanzaOld := ""

	// cerco per il cluster lo stage number di dove vuoi andare
	/* ************************************************************************************************ */
	// KUBESTAGE
	Logga(ctx, "Getting KUBESTAGE grpcserver")

	argsStage := make(map[string]string)
	argsStage["source"] = "devops-8"
	argsStage["$select"] = "XKUBESTAGE05"
	argsStage["center_dett"] = "dettaglio"
	argsStage["$filter"] = "equals(XKUBESTAGE03,'" + cluster + "') "
	argsStage["$filter"] += " and equals(XKUBESTAGE04,'" + enviro + "') "

	//$filter=contains(XART20,'(kg)') or contains(XART20,'pizza')
	restyStageRes := ApiCallGET(ctx, false, argsStage, "msdevops", "/devops/KUBESTAGE", devopsToken, "")
	if restyStageRes.Errore < 0 {
		Logga(ctx, restyStageRes.Log)
	}

	stageNum := 0.00
	if len(restyStageRes.BodyJson) > 0 {
		stageNum = restyStageRes.BodyJson["XKUBESTAGE05"].(float64)
		Logga(ctx, "KUBESTAGE: OK")
	} else {
		Logga(ctx, "KUBESTAGE: MISSING")
	}
	/* ************************************************************************************************ */

	Logga(ctx, "")
	Logga(ctx, "FOUND stage num: "+strconv.Itoa(int(stageNum)))

	stageNum = stageNum - 1
	stageNumStr := strconv.Itoa(int(stageNum))

	Logga(ctx, "Stage Name to find: "+stageNumStr)
	Logga(ctx, "")

	changeCluster := false
	// cerco lo stage number - 1
	/* ************************************************************************************************ */
	// KUBESTAGE
	Logga(ctx, "Getting KUBESTAGE grpcserver")

	argsStage2 := make(map[string]string)
	argsStage2["source"] = "devops-8"
	argsStage2["$select"] = "XKUBESTAGE04"
	argsStage2["center_dett"] = "dettaglio"
	argsStage2["$filter"] = "equals(XKUBESTAGE03,'" + cluster + "') "
	argsStage2["$filter"] += " and XKUBESTAGE05 eq " + stageNumStr + " "

	restyStage2Res := ApiCallGET(ctx, false, argsStage2, "msdevops", "/devops/KUBESTAGE", devopsToken, "")
	if restyStage2Res.Errore < 0 {
		Logga(ctx, restyStage2Res.Log)
	}

	if len(restyStage2Res.BodyJson) > 0 {
		nomeStage = restyStage2Res.BodyJson["XKUBESTAGE04_COD"].(string)
		Logga(ctx, "Stage Name: "+nomeStage)
		Logga(ctx, "KUBESTAGE: OK")
	} else {

		Logga(ctx, "")
		Logga(ctx, "Non trovo lo stage nel cluster di prod e lo cerco nel relativo cluster di DEV")
		Logga(ctx, "Getting KUBESTAGE grpcserver secondo round")
		Logga(ctx, "")
		Logga(ctx, "")
		argsStage3 := make(map[string]string)
		argsStage3["source"] = "devops-8"
		argsStage3["$select"] = "XKUBESTAGE04"
		argsStage3["center_dett"] = "dettaglio"
		argsStage3["$filter"] = "equals(XKUBESTAGE03,'" + clusterDev + "') "
		argsStage3["$filter"] += " and XKUBESTAGE05 eq " + stageNumStr + " "

		restyStage3Res := ApiCallGET(ctx, false, argsStage3, "msdevops", "/devops/KUBESTAGE", devopsToken, "")
		if restyStage3Res.Errore < 0 {
			Logga(ctx, restyStage3Res.Log)
		}

		if len(restyStage3Res.BodyJson) > 0 {
			changeCluster = true
			nomeStage = restyStage3Res.BodyJson["XKUBESTAGE04_COD"].(string)
			Logga(ctx, "Stage Name: "+nomeStage)
			Logga(ctx, "KUBESTAGE: OK")
		}

		Logga(ctx, "KUBESTAGE: MISSING")
	}
	/* ************************************************************************************************ */

	Logga(ctx, "")
	Logga(ctx, "THE STAGE IS: "+nomeStage)
	Logga(ctx, "")

	// cerco la versione con lo stage number n-1
	/* ************************************************************************************************ */
	// KUBESTAGE
	Logga(ctx, "Getting DEPLOYLOG grpcserver 1")

	// NORMALIZZO IL NOME PER LA VECCHIA GESTIONE
	istanzaSplit := strings.Split(istanza, "-")

	if changeCluster {
		istanzaOld = clusterDev + "-" + istanzaSplit[len(istanzaSplit)-1]
	} else {
		istanzaOld = cluster + "-" + istanzaSplit[len(istanzaSplit)-1]
	}

	Logga(ctx, "")
	Logga(ctx, "THE ISTANZA IS: "+istanza+"|"+istanzaOld)
	Logga(ctx, "")

	argsDeploy := make(map[string]string)
	argsDeploy["source"] = "devops-8"
	argsDeploy["$select"] = "XDEPLOYLOG05"
	argsDeploy["center_dett"] = "dettaglio"
	argsDeploy["$filter"] = " ( equals(XDEPLOYLOG04,'" + istanza + "') OR equals(XDEPLOYLOG04,'" + istanzaOld + "') ) "
	argsDeploy["$filter"] += " and equals(XDEPLOYLOG09,'" + nomeStage + "') "
	argsDeploy["$filter"] += " and equals(XDEPLOYLOG06,'1') "

	//$filter=contains(XART20,'(kg)') or contains(XART20,'pizza')
	restyDeployRes := ApiCallGET(ctx, true, argsDeploy, "msdevops", "/devops/DEPLOYLOG", devopsToken, "")
	if restyDeployRes.Errore < 0 {
		Logga(ctx, restyDeployRes.Log)
	}

	if len(restyDeployRes.BodyJson) > 0 {
		versione = restyDeployRes.BodyJson["XDEPLOYLOG05"].(string)
		Logga(ctx, "Versione: "+versione)
		Logga(ctx, "DEPLOYLOG 1: OK")
	} else {
		Logga(ctx, "DEPLOYLOG 1: MISSING")
	}
	/* ************************************************************************************************ */

	return versione, istanzaOld
}
func GetAccessCluster(ctx context.Context, cluster, devopsToken string) ClusterAccess {
	/* ************************************************************************************************ */
	// KUBECLUSTER
	Logga(ctx, "Getting KUBECLUSTER")

	argsClu := make(map[string]string)
	argsClu["source"] = "devops-8"
	argsClu["$select"] = "XKUBECLUSTER15,XKUBECLUSTER20,XKUBECLUSTER22"
	argsClu["center_dett"] = "dettaglio"
	argsClu["$filter"] = "equals(XKUBECLUSTER03,'" + cluster + "') "

	restyKubeCluRes := ApiCallGET(ctx, false, argsClu, "msdevops", "/devops/KUBECLUSTER", devopsToken, "")
	if restyKubeCluRes.Errore < 0 {
		Logga(ctx, restyKubeCluRes.Log)
	}

	var cluAcc ClusterAccess
	if len(restyKubeCluRes.BodyJson) > 0 {

		cluAcc.Domain = restyKubeCluRes.BodyJson["XKUBECLUSTER15"].(string)
		cluAcc.AccessToken = restyKubeCluRes.BodyJson["XKUBECLUSTER20"].(string)
		cluAcc.ReffappCustomerID = restyKubeCluRes.BodyJson["XKUBECLUSTER22"].(string)

		Logga(ctx, "KUBECLUSTER OK")
	} else {
		Logga(ctx, "   !!!   KUBECLUSTER MISSING")
	}
	Logga(ctx, "")
	/* ************************************************************************************************ */

	return cluAcc
}
func GetJsonDatabases(ctx context.Context, stage, developer string, market int32, arrConn MasterConn) (map[string]interface{}, LoggaErrore) {
	Logga(ctx, "Getting Json Db")

	var erro LoggaErrore
	erro.Errore = 0

	devopsToken, erro := GetCoreFactoryToken(ctx)
	if erro.Errore < 0 {
		Logga(ctx, erro.Log)
	} else {
		Logga(ctx, "Token OK")
	}

	clusterDett := GetAccessCluster(ctx, stage, devopsToken)
	clusterToken, erro := GetCustomerToken(ctx, clusterDett.AccessToken, clusterDett.ReffappCustomerID, clusterDett.Domain, clusterDett.Domain)

	dominio := GetApiHost()

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

	client := resty.New()
	client.Debug = false
	client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	res, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetHeader("microservice", "msappman").
		SetAuthToken(clusterToken).
		SetBody(keyvalueslice).
		Post(dominio + "/api/" + os.Getenv("coreapiVersion") + "/appman/getDeveloperMsList")

	callResponse := map[string]interface{}{}
	if err != nil { // HTTP ERRORE
		erro.Errore = -1
		erro.Log = err.Error()
	} else {
		if res.StatusCode() != 200 {
			erro.Errore = -1
			erro.Log = err.Error()
		} else {

			err1 := json.Unmarshal(res.Body(), &callResponse)
			if err1 != nil {
				erro.Errore = -1
				erro.Log = err1.Error()
			}
		}
	}
	return callResponse, erro
}
func GetCustomerToken(ctx context.Context, accessToken, refappCustomer, resource, dominio string) (string, LoggaErrore) {

	Logga(ctx, "getCustomerToken")
	Logga(ctx, "Customer Token "+dominio)
	Logga(ctx, "refappCustomer "+refappCustomer)
	Logga(ctx, "resource "+resource)
	Logga(ctx, "accessToken "+accessToken)

	var erro LoggaErrore
	erro.Errore = 0

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

	restyAuthResponse, restyAuthErr := ApiCallLOGIN(ctx, false, argsAuth, "msauth", "/auth/login", dominio)
	if restyAuthErr.Errore < 0 {
		// QUI ERRORE
		erro.Errore = -1
		erro.Log = restyAuthErr.Log
		return "", erro
	}
	if len(restyAuthResponse) > 0 {
		return restyAuthResponse["idToken"].(string), erro
	} else {
		erro.Errore = -1
		erro.Log = "token MISSING"
		return "", erro
	}
}
func GetArrRepo(ctx context.Context, team, customSettings string) map[int]Repos {

	// cerco il token di Corefactory
	devopsToken, erro := GetCoreFactoryToken(ctx)
	if erro.Errore < 0 {
		Logga(ctx, erro.Log)
	}

	/* ************************************************************************************************ */
	// KUBEMONOLITHREPO
	Logga(ctx, "Getting KUBEMONOLITHREPO")
	args := make(map[string]string)
	args["source"] = "devops-9"
	args["center_dett"] = "visualizza"
	args["$select"] = "XKUBEMONOLITHREPO04,XKUBEMONOLITHREPO05,XKUBEMONOLITHREPO06"
	args["$filter"] = "equals(XKUBEMONOLITHREPO03,'" + team + "')"
	args["$order"] += "XKUBEMONOLITHREPO05"

	restyKUBEMONOLITHREPORes := ApiCallGET(ctx, false, args, "msdevops", "/devops/KUBEMONOLITHREPO", devopsToken, "")
	if restyKUBEMONOLITHREPORes.Errore < 0 {
		Logga(ctx, restyKUBEMONOLITHREPORes.Log)
	}

	i := 0
	arrRepo := make(map[int]Repos)
	var repo Repos
	if len(restyKUBEMONOLITHREPORes.BodyArray) > 0 {
		for _, x := range restyKUBEMONOLITHREPORes.BodyArray {

			repo.Nome = x["XKUBEMONOLITHREPO04"].(string)
			repo.Repo = x["XKUBEMONOLITHREPO05"].(string)
			repo.Sw = int(x["XKUBEMONOLITHREPO06"].(float64))

			arrRepo[i] = repo
			i++

		}
		Logga(ctx, "KUBEMONOLITHREPO OK")
	} else {
		Logga(ctx, "KUBEMONOLITHREPO MISSING")
	}

	/* ************************************************************************************************ */

	// se alla func non passo un custom setting
	// vuol dire che li voglio tutti
	fmt.Println(customSettings)
	if customSettings == "" {

		// qui ottengo i CS del team

		/* ************************************************************************************************ */
		// KUBEMONOLITHCS
		Logga(ctx, "Getting KUBEMONOLITHCS")
		args := make(map[string]string)
		args["source"] = "devops-9"
		args["center_dett"] = "visualizza"
		args["$select"] = "XKUBEMONOLITHCS06"
		args["$filter"] = "equals(XKUBEMONOLITHCS03,'" + team + "')"
		args["$filter"] += " and equals(XKUBEMONOLITHCS07,'1')"

		restyKUBEMONOLITHCSRes := ApiCallGET(ctx, false, args, "msdevops", "/devops/KUBEMONOLITHCS", devopsToken, "")
		if restyKUBEMONOLITHCSRes.Errore < 0 {
			Logga(ctx, restyKUBEMONOLITHCSRes.Log)
		}

		if len(restyKUBEMONOLITHCSRes.BodyArray) > 0 {
			for _, x := range restyKUBEMONOLITHCSRes.BodyArray {

				repo.Nome = x["XKUBEMONOLITHCS06"].(string)
				repo.Repo = x["XKUBEMONOLITHCS06"].(string)
				repo.Tipo = "CS"
				arrRepo[i] = repo
				i++

			}
			Logga(ctx, "KUBEMONOLITHCS OK")
		} else {
			Logga(ctx, "KUBEMONOLITHCS MISSING")
		}

		/* ************************************************************************************************ */

	} else {
		fmt.Println("customSettings: " + customSettings)
		repo.Nome = "customSettings"
		repo.Repo = customSettings
		arrRepo[i] = repo

	}

	return arrRepo
}
func GetDeploymentApi(namespace, apiHost, apiToken string) (DeploymntStatus, LoggaErrore) {

	var erro LoggaErrore
	erro.Errore = 0

	var deploy DeploymntStatus

	clientKUBE := resty.New()
	clientKUBE.Debug = false
	clientKUBE.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	resKUBE, errKUBE := clientKUBE.R().
		SetHeader("Content-Type", "application/json").
		SetAuthToken(apiToken).
		Get("https://" + apiHost + "/apis/apps/v1/namespaces/" + namespace + "/deployments")

	if errKUBE != nil {
		erro.Errore = -1
		erro.Log = errKUBE.Error()
		return deploy, erro
	}

	if resKUBE.StatusCode() != 200 {
		erro.Errore = -1
		erro.Log = "API Res Status: " + resKUBE.Status()
		return deploy, erro
	}

	a := map[string]interface{}{}
	errUm := json.Unmarshal(resKUBE.Body(), &a)
	if errUm != nil {
		erro.Errore = -1
		erro.Log = errUm.Error()
		return deploy, erro
	}

	jsonStr, errMa := json.Marshal(a)
	if errMa != nil {
		erro.Errore = -1
		erro.Log = errMa.Error()
		return deploy, erro
	}

	errUm2 := json.Unmarshal(jsonStr, &deploy)
	if errUm2 != nil {
		erro.Errore = -1
		erro.Log = errUm2.Error()
		return deploy, erro
	}

	return deploy, erro
}
func CheckPodHealth(microservice, versione, namespace, apiHost, apiToken string) (bool, LoggaErrore) {

	var erro LoggaErrore
	erro.Errore = 0

	msDeploy := microservice + "-v" + versione
	msMatch := false
	i := 0
	for {
		item, err := GetDeploymentApi(namespace, apiHost, apiToken)
		if err.Errore < 0 {
			erro.Errore = -1
			erro.Log = err.Log
			return false, erro
		} else {

			if len(item.Items) == 0 {
				erro.Errore = -1
				erro.Log = "No Deployment Found in Namespace"
				return false, erro
			}

			for _, item := range item.Items {

				//fmt.Println(item.Metadata.Name, "-", msDeploy)
				if item.Metadata.Name == msDeploy {
					msMatch = true

					fmt.Println(item.Metadata.Name+" desired: ", item.Spec.Replicas, " - aviable: ", item.Status.ReadyReplicas)

					if item.Spec.Replicas == item.Status.ReadyReplicas {
						return true, erro
					}
				}

				// sto girando a vuoto perche nessun item risponde a cio che cerco
				if i >= 1 && !msMatch {
					erro.Errore = -1
					erro.Log = "nessun item risponde a cio che cerco"
					return false, erro
				}

			}

			time.Sleep(10 * time.Second)
			if i > 25 {
				erro.Errore = -1
				erro.Log = "Time Out. Pod is not Running"
				return false, erro
			}
		}
	}
}
func DeleteObsoleteObjects(ctx context.Context, ires IstanzaMicro, versione, canaryProduction, namespace, enviro string) LoggaErrore {

	var erro LoggaErrore
	erro.Errore = 0

	istanza := ires.Istanza
	microservice := ""
	if ires.Monolith == 0 {
		microservice = ires.Microservice
	} else {
		microservice = ires.PodName
	}
	Logga(ctx, "****************************************************************************************")
	Logga(ctx, "DELETING OBSOLETE PODS")

	// cerco il token di Corefactory
	devopsToken, erro := GetCoreFactoryToken(ctx)
	if erro.Errore < 0 {
		Logga(ctx, erro.Log)
	}
	/* ************************************************************************************************ */
	// DEPLOYLOG
	Logga(ctx, "Getting DEPLOYLOG - deleteObsoleteMonolith")

	argsDeploy := make(map[string]string)
	argsDeploy["source"] = "devops-8"
	argsDeploy["$select"] = "XDEPLOYLOG03,XDEPLOYLOG05"
	argsDeploy["center_dett"] = "visualizza"
	argsDeploy["$filter"] = "equals(XDEPLOYLOG04,'" + istanza + "') "
	argsDeploy["$filter"] += " and (equals(XDEPLOYLOG03,'canary') OR equals(XDEPLOYLOG03,'production'))  "
	argsDeploy["$filter"] += " and equals(XDEPLOYLOG06,'1') "
	if ires.SwMultiEnvironment == "1" {
		argsDeploy["$filter"] += " and equals(XDEPLOYLOG09,'" + enviro + "') "
	}

	restyDeployRes := ApiCallGET(ctx, true, argsDeploy, "msdevops", "/devops/DEPLOYLOG", devopsToken, "")
	if restyDeployRes.Errore < 0 {
		Logga(ctx, restyDeployRes.Log)
		return erro
	}

	versioneProductionDb := ""
	versioneCanaryDb := ""
	if len(restyDeployRes.BodyArray) > 0 {
		for _, x := range restyDeployRes.BodyArray {

			if x["XDEPLOYLOG03"].(string) == "canary" {
				versioneCanaryDb = "v" + x["XDEPLOYLOG05"].(string)
			}
			if x["XDEPLOYLOG03"].(string) == "production" {
				versioneProductionDb = "v" + x["XDEPLOYLOG05"].(string)
			}

		}
		Logga(ctx, "DEPLOYLOG OK")
	} else {
		Logga(ctx, "DEPLOYLOG MISSING", "warn")
	}
	Logga(ctx, "=== NEVER DELETE INNOCENT DEPLOYMENTS === ")
	Logga(ctx, "eventually to kill: "+microservice)
	Logga(ctx, "Never kill Canary: "+versioneCanaryDb)
	Logga(ctx, "Never kill Production: "+versioneProductionDb)

	/* ************************************************************************************************ */

	// ho recuperato le versioni canary e production che NON cancellero MAI :D

	//msDeploy := microservice + "-v" + versione

	//LogJson(ires)
	item, err := GetDeploymentApi(namespace, ires.ApiHost, ires.ApiToken)
	//LogJson(item)
	if err.Errore < 0 {
		erro.Errore = -1
		erro.Log = err.Log
		return erro
	} else {

		if len(item.Items) == 0 {
			erro.Errore = -1
			erro.Log = "No Deployment Found in Namespace"
			return erro
		}

		Logga(ctx, "API returns : "+strconv.Itoa(len(item.Items))+" ITEMS")

		for _, item := range item.Items {

			Logga(ctx, "item yaml: "+item.Spec.Selector.MatchLabels.App)
			Logga(ctx, "version: "+item.Spec.Selector.MatchLabels.Version)
			Logga(ctx, "item ires: "+microservice)
			// primo filtro sulla refapp giusta
			if item.Spec.Selector.MatchLabels.App == microservice {

				Logga(ctx, "Kill everything with different version of canary: "+versioneCanaryDb+" or production: "+versioneProductionDb+" - Current version: "+item.Spec.Selector.MatchLabels.Version)
				// secondo filtro sulle versione
				if item.Spec.Selector.MatchLabels.Version == versioneCanaryDb || item.Spec.Selector.MatchLabels.Version == versioneProductionDb {
					// SKIP
				} else {

					deployment := item.Spec.Selector.MatchLabels.App + "-" + item.Spec.Selector.MatchLabels.Version
					Logga(ctx, "KILL-KILL-KILL-KILL-KILL-KILL-KILL-KILL-KILL-KILL-KILL-KILL-KILL-KILL-KILL-KILL-KILL-KILL-KILL")

					Logga(ctx, "I DO KILL: "+deployment)
					// delete deployment
					DeleteObjectsApi(namespace, ires.ApiHost, ires.ApiToken, deployment, "deployment")

					// delete HPA
					DeleteObjectsApi(namespace, ires.ApiHost, ires.ApiToken, deployment, "hpa")

				}
			}
		}

	}
	return erro
}
func DeleteObjectsApi(namespace, apiHost, apiToken, object, kind string) LoggaErrore {

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
	}

	var erro LoggaErrore
	erro.Errore = 0

	args := make(map[string]string)
	args["kind"] = "DeleteOptions"
	args["apiVersion"] = "v1"
	args["apiVerpropagationPolicysion"] = "Foreground"

	clientKUBE := resty.New()
	clientKUBE.Debug = true
	clientKUBE.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	resKUBE, errKUBE := clientKUBE.R().
		SetHeader("Content-Type", "application/json").
		SetAuthToken(apiToken).
		SetQueryParams(args).
		Delete("https://" + apiHost + "/apis/" + apiversion + "/namespaces/" + namespace + "/" + name + "/" + object)

	if errKUBE != nil {
		erro.Errore = -1
		erro.Log = errKUBE.Error()
		return erro
	}

	if resKUBE.StatusCode() != 200 {
		erro.Errore = -1
		erro.Log = "API Res Status: " + resKUBE.Status()
		return erro
	}

	a := map[string]interface{}{}
	errUm := json.Unmarshal(resKUBE.Body(), &a)
	if errUm != nil {
		erro.Errore = -1
		erro.Log = errUm.Error()
		return erro
	}

	return erro
}
