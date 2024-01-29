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

	Logga(ctx, os.Getenv("JsonLog"), "", "Getting GRU")

	var erro error

	args := make(map[string]string)
	args["center_dett"] = "dettaglio"
	args["source"] = "users-3"
	args["$select"] = "XGRU05,XGRU06"

	gruRes, _ := ApiCallGET(ctx, os.Getenv("RestyDebug"), args, "msusers", "/users/GRU/equals(XGRU03,'"+gruppo+"')", token, dominio, coreApiVersion)
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
func GetNextVersion(ctx context.Context, masterBranch, nomeDocker, tenant, accessToken, loginApiDomain, coreApiVersion string) (string, error) {

	var erro error

	// cerco il token di Corefactory
	Logga(ctx, os.Getenv("JsonLog"), "", "Getting token")
	devopsToken, erro := GetCoreFactoryToken(ctx, tenant, accessToken, loginApiDomain, coreApiVersion, os.Getenv("RestyDebug"))
	if erro != nil {
		Logga(ctx, os.Getenv("JsonLog"), "", erro.Error())
	} else {
		Logga(ctx, os.Getenv("JsonLog"), "", "Token OK")
	}

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

	restyKubeImicroservRes, errImicroservRes := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsImicro, "msdevops", "/devops/KUBEDKRBUILD", devopsToken, loginApiDomain, coreApiVersion)
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
func GetMicroserviceDetail(ctx context.Context, team, ims, gitDevMaster, buildVersion, devopsToken, autopilot, enviro, dominio, coreApiVersion string) (Microservice, error) {

	Logga(ctx, os.Getenv("JsonLog"), "")
	Logga(ctx, os.Getenv("JsonLog"), " + + + + + + + + + + + + + + + + + + + + ")
	Logga(ctx, os.Getenv("JsonLog"), "TEAM "+team)
	Logga(ctx, os.Getenv("JsonLog"), "IMS "+ims)
	Logga(ctx, os.Getenv("JsonLog"), "gitDevMaster "+gitDevMaster)
	Logga(ctx, os.Getenv("JsonLog"), "BUILDVERSION "+buildVersion)
	Logga(ctx, os.Getenv("JsonLog"), "getMicroserviceDetail begin")

	var erro error

	devops := "devops"
	if strings.Contains(ims, "p2rpowerna-monolith") {
		devops = "devopsmono"
	}

	versioneArr := strings.Split(buildVersion, ".")
	versione := ""

	if len(versioneArr) > 1 {
		versione = versioneArr[0] + Times("0", 2-len(versioneArr[1])) + versioneArr[1] + Times("0", 2-len(versioneArr[2])) + versioneArr[2] + Times("0", 2-len(versioneArr[3])) + versioneArr[3]
	} else {
		versione = buildVersion
	}

	var microservices Microservice

	/* ************************************************************************************************ */
	// KUBEIMICROSERV
	Logga(ctx, os.Getenv("JsonLog"), "Getting KUBEIMICROSERV")
	argsImicro := make(map[string]string)
	argsImicro["source"] = "devops-8"
	argsImicro["$select"] = "XKUBEIMICROSERV04,XKUBEIMICROSERV05,XKUBEIMICROSERV07"
	argsImicro["center_dett"] = "dettaglio"
	argsImicro["$filter"] = "equals(XKUBEIMICROSERV03,'" + ims + "') "

	restyKubeImicroservRes, errKubeImicroservRes := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsImicro, "ms"+devops, "/"+devops+"/KUBEIMICROSERV", devopsToken, dominio, coreApiVersion)
	if errKubeImicroservRes != nil {
		Logga(ctx, os.Getenv("JsonLog"), errKubeImicroservRes.Error())
		return microservices, errKubeImicroservRes
	}

	microservice := ""
	cluster := ""
	if len(restyKubeImicroservRes.BodyJson) > 0 {

		microservice = restyKubeImicroservRes.BodyJson["XKUBEIMICROSERV04_COD"].(string)
		microservices.VersMicroservice = restyKubeImicroservRes.BodyJson["XKUBEIMICROSERV07"].(string)
		cluster = restyKubeImicroservRes.BodyJson["XKUBEIMICROSERV05"].(string)
		Logga(ctx, os.Getenv("JsonLog"), "KUBEIMICROSERV OK")
	} else {
		Logga(ctx, os.Getenv("JsonLog"), "   !!!   KUBEIMICROSERV MISSING")
		erro := errors.New("KUBEIMICROSERV MISSING")
		return microservices, erro
	}
	Logga(ctx, os.Getenv("JsonLog"), "")

	/* ************************************************************************************************ */
	// KUBECLUSTER
	Logga(ctx, os.Getenv("JsonLog"), "Getting KUBECLUSTER")

	argsClu := make(map[string]string)
	argsClu["source"] = "devops-8"
	argsClu["$select"] = "XKUBECLUSTER06,XKUBECLUSTER12,XKUBECLUSTER15"
	argsClu["center_dett"] = "dettaglio"
	argsClu["$filter"] = "equals(XKUBECLUSTER03,'" + cluster + "') "

	restyKubeCluRes, errKubeCluRes := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsClu, "ms"+devops, "/"+devops+"/KUBECLUSTER", devopsToken, dominio, coreApiVersion)
	if errKubeCluRes != nil {
		Logga(ctx, os.Getenv("JsonLog"), errKubeCluRes.Error())
		return microservices, errKubeCluRes
	}

	clusterHost := ""
	if len(restyKubeCluRes.BodyJson) > 0 {

		clusterHost = restyKubeCluRes.BodyJson["XKUBECLUSTER15"].(string)

		Logga(ctx, os.Getenv("JsonLog"), "KUBECLUSTER OK")
	} else {
		Logga(ctx, os.Getenv("JsonLog"), "   !!!   KUBECLUSTER MISSING")
	}
	Logga(ctx, os.Getenv("JsonLog"), "")
	/* ************************************************************************************************ */

	/* ************************************************************************************************ */
	// KUBEMICROSERV
	Logga(ctx, os.Getenv("JsonLog"), "Getting KUBEMICROSERV")

	argsMS := make(map[string]string)
	argsMS["source"] = "devops-8"
	argsMS["$select"] = "XKUBEMICROSERV03,XKUBEMICROSERV04,XKUBEMICROSERV05,XKUBEMICROSERV08,XKUBEMICROSERV16,XKUBEMICROSERV17,XKUBEMICROSERV18"
	argsMS["center_dett"] = "dettaglio"
	argsMS["$filter"] = "equals(XKUBEMICROSERV05,'" + microservice + "') "
	restyKubeMSRes, errKubeMSRes := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsMS, "ms"+devops, "/"+devops+"/KUBEMICROSERV", devopsToken, dominio, coreApiVersion)
	if errKubeMSRes != nil {
		Logga(ctx, os.Getenv("JsonLog"), errKubeMSRes.Error())
		return microservices, errKubeMSRes
	}

	hpaTmpl := ""
	if len(restyKubeMSRes.BodyJson) > 0 {
		microservices.Nome = restyKubeMSRes.BodyJson["XKUBEMICROSERV05"].(string)
		microservices.Descrizione = restyKubeMSRes.BodyJson["XKUBEMICROSERV03"].(string)
		//microservices.Replicas = restyKubeMSRes.BodyJson["XKUBEMICROSERV18"].(string)
		microservices.Public = int(restyKubeMSRes.BodyJson["XKUBEMICROSERV18"].(float64))
		microservices.Namespace = restyKubeMSRes.BodyJson["XKUBEMICROSERV04_COD"].(string)
		microservices.Virtualservice = strconv.FormatFloat(restyKubeMSRes.BodyJson["XKUBEMICROSERV08"].(float64), 'f', 0, 64)
		hpaTmpl = restyKubeMSRes.BodyJson["XKUBEMICROSERV16_COD"].(string)
		Logga(ctx, os.Getenv("JsonLog"), "KUBEMICROSERV OK")
	} else {
		Logga(ctx, os.Getenv("JsonLog"), "   !!!   KUBEMICROSERV MISSING")
		erro := errors.New("KUBEIMICROSERV MISSING")
		return microservices, erro
	}
	Logga(ctx, os.Getenv("JsonLog"), "")

	if autopilot != "1" {
		/* ************************************************************************************************ */
		// KUBEMICROSERVHPA
		Logga(ctx, os.Getenv("JsonLog"), "Getting KUBEMICROSERVHPA")
		argsHpa := make(map[string]string)
		argsHpa["source"] = "devops-8"
		argsHpa["$select"] = "XKUBEMICROSERVHPA04,XKUBEMICROSERVHPA05,XKUBEMICROSERVHPA06,XKUBEMICROSERVHPA07,XKUBEMICROSERVHPA08,XKUBEMICROSERVHPA09,XKUBEMICROSERVHPA10"
		argsHpa["center_dett"] = "dettaglio"
		argsHpa["$filter"] = "equals(XKUBEMICROSERVHPA03,'" + hpaTmpl + "') "

		restyKubeHpaRes, errKubeHpaRes := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsHpa, "ms"+devops, "/"+devops+"/KUBEMICROSERVHPA", devopsToken, dominio, coreApiVersion)
		if errKubeHpaRes != nil {
			Logga(ctx, os.Getenv("JsonLog"), errKubeHpaRes.Error())
			return microservices, errKubeHpaRes
		}

		if len(restyKubeHpaRes.BodyJson) > 0 {

			// In XKUBEMICROSERVHPA10 salviamo la mappa per personalizzare l'HPA in ogni environment
			hpaString, _ := restyKubeHpaRes.BodyJson["XKUBEMICROSERVHPA10"].(string)

			checkHpaEnviro := false
			var hpaEnviro Hpa
			if hpaString != "" {

				var hpaMap map[string]Hpa
				json.Unmarshal([]byte(hpaString), &hpaMap)

				hpaEnviro, checkHpaEnviro = hpaMap[enviro]
			}

			// Se esiste la personalizzazione per environment, prendo quella, altrimenti il default delle altri colonne
			if checkHpaEnviro {
				microservices.Hpa = hpaEnviro
			} else {
				var hpa Hpa
				hpa.MinReplicas = strconv.FormatFloat(restyKubeHpaRes.BodyJson["XKUBEMICROSERVHPA04"].(float64), 'f', 0, 64)
				hpa.MaxReplicas = strconv.FormatFloat(restyKubeHpaRes.BodyJson["XKUBEMICROSERVHPA05"].(float64), 'f', 0, 64)
				hpa.CpuTipoTarget = restyKubeHpaRes.BodyJson["XKUBEMICROSERVHPA06"].(string)
				hpa.CpuTarget = restyKubeHpaRes.BodyJson["XKUBEMICROSERVHPA07"].(string)
				hpa.MemTipoTarget = restyKubeHpaRes.BodyJson["XKUBEMICROSERVHPA08"].(string)
				hpa.MemTarget = restyKubeHpaRes.BodyJson["XKUBEMICROSERVHPA09"].(string)
				microservices.Hpa = hpa
			}
			Logga(ctx, os.Getenv("JsonLog"), "KUBEMICROSERVHPA OK")
		} else {
			Logga(ctx, os.Getenv("JsonLog"), "   !!!   KUBEMICROSERVHPA MISSING")
			erro := errors.New("KUBEMICROSERVHPA MISSING")
			return microservices, erro
		}
		Logga(ctx, os.Getenv("JsonLog"), "")

		/* ************************************************************************************************ */
	}

	/* ************************************************************************************************ */
	// SELKUBEDKRLIST
	Logga(ctx, os.Getenv("JsonLog"), "Getting SELKUBEDKRLIST")
	argsDkr := make(map[string]string)
	argsDkr["center_dett"] = "visualizza"
	argsDkr["$filter"] = "equals(XSELKUBEDKRLIST10,'" + microservices.Nome + "') "

	restyDkrLstRes, errDkrLstRes := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsDkr, "ms"+devops, "/core/custom/SELKUBEDKRLIST/values", devopsToken, dominio, coreApiVersion)
	if errDkrLstRes != nil {
		Logga(ctx, os.Getenv("JsonLog"), errDkrLstRes.Error())
		return microservices, errDkrLstRes
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
			Logga(ctx, os.Getenv("JsonLog"), "Getting KUBEDKRBUILD func.go 2")
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

			// 2023 04 13 - laszlo mwpo e scarp non ricordano il perche della esclusione dei master e la mandano a fanculo, cosi a puorc
			// perche a prescindere riteniamo che in caso di build dobbiamo prendere
			//argsBld["$fullquery"] += " order by (case when XKUBEDKRBUILD04 = 'master' then 0 else 1 end ) desc,cast(XKUBEDKRBUILD06 as unsigned) DESC "
			argsBld["$fullquery"] += " order by cast(XKUBEDKRBUILD06 as unsigned) DESC "
			argsBld["$fullquery"] += " limit 1 "
			fmt.Println(argsBld["$fullquery"])

			restyKubeBldRes, errKubeBldRes := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsBld, "ms"+devops, "/"+devops+"/custom/KUBEDKRBUILD/values", devopsToken, dominio, coreApiVersion)

			//fmt.Println(restyKubeBldRes)
			if errKubeBldRes != nil {
				//fmt.Println("A")
				Logga(ctx, os.Getenv("JsonLog"), errKubeBldRes.Error())
				return microservices, errKubeBldRes
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
				Logga(ctx, os.Getenv("JsonLog"), "KUBEDKRBUILD OK")
			} else {
				// se manca la build alla versione indicata proviamo a cercare l'ultima
				// se manca anche questa allora errore mai fatta una build !!!!!

				Logga(ctx, os.Getenv("JsonLog"), "KUBEDKRBUILD MISSING ON "+versione+" seek for latest")
				argsBld := make(map[string]string)

				argsBld["$fullquery"] = "select XKUBEDKRBUILD06,XKUBEDKRBUILD04,XKUBEDKRBUILD07,XKUBEDKRBUILD09,XKUBEDKRBUILD10,XKUBEDKRBUILD12,XKUBEDKRBUILD13 "
				argsBld["$fullquery"] += "from TB_ANAG_KUBEDKRBUILD00 "
				argsBld["$fullquery"] += "where 1>0 "
				argsBld["$fullquery"] += "AND XKUBEDKRBUILD03 = '" + docker + "' "
				argsBld["$fullquery"] += "AND XKUBEDKRBUILD08 = '" + team + "' "
				argsBld["$fullquery"] += " order by (case when XKUBEDKRBUILD04 = 'master' then 0 else 1 end ) desc,cast(XKUBEDKRBUILD06 as unsigned) DESC "
				argsBld["$fullquery"] += " limit 1 "
				fmt.Println(argsBld["$fullquery"])

				restyKubeBldRes, errKubeBldRes := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsBld, "ms"+devops, "/core/custom/KUBEDKRBUILD/values", devopsToken, dominio, coreApiVersion)

				//fmt.Println(restyKubeBldRes)
				if errKubeBldRes != nil {
					//fmt.Println("A")
					Logga(ctx, os.Getenv("JsonLog"), errKubeBldRes.Error())
					return microservices, errKubeBldRes
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
					Logga(ctx, os.Getenv("JsonLog"), "KUBEDKRBUILD LATEST OK")
				} else {

					Logga(ctx, os.Getenv("JsonLog"), "   !!! "+docker+"  KUBEDKRBUILD MISSING")
					erro := errors.New("The component " + docker + " of the microservice " + microservices.Nome + " is MISSING - you have to build it first.")
					return microservices, erro
				}
			}

			Logga(ctx, os.Getenv("JsonLog"), "")

			/* ************************************************************************************************ */

			/* ************************************************************************************************ */
			// KUBEDKRMOUNT
			Logga(ctx, os.Getenv("JsonLog"), "Getting KUBEDKRMOUNT")
			argsMnt := make(map[string]string)
			argsMnt["source"] = "devops-8"
			argsMnt["$select"] = "XKUBEDKRMOUNT04,XKUBEDKRMOUNT05,XKUBEDKRMOUNT06,XKUBEDKRMOUNT07,XKUBEDKRMOUNT08"
			argsMnt["center_dett"] = "visualizza"
			argsMnt["$filter"] = "equals(XKUBEDKRMOUNT03,'" + docker + "') "

			restyKubeMntRes, errKubeMntRes := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsMnt, "ms"+devops, "/"+devops+"/KUBEDKRMOUNT", devopsToken, dominio, coreApiVersion)
			if errKubeMntRes != nil {
				Logga(ctx, os.Getenv("JsonLog"), errKubeMntRes.Error())
				return microservices, errKubeMntRes
			}

			if len(restyKubeMntRes.BodyArray) > 0 {
				var mounts []Mount
				for _, x := range restyKubeMntRes.BodyArray {

					var mount Mount
					mount.Nome = x["XKUBEDKRMOUNT04"].(string)
					mount.Mount = x["XKUBEDKRMOUNT05"].(string)
					mount.Subpath = x["XKUBEDKRMOUNT06"].(string)
					mount.ClaimName = x["XKUBEDKRMOUNT07"].(string)

					if x["XKUBEDKRMOUNT08"] != nil {
						fromSecretFloat := x["XKUBEDKRMOUNT08"].(float64)
						mount.FromSecret = fromSecretFloat == 1
					}

					mounts = append(mounts, mount)
				}
				pod.Mount = mounts
				Logga(ctx, os.Getenv("JsonLog"), "KUBEDKRMOUNT OK")
			} else {
				Logga(ctx, os.Getenv("JsonLog"), "KUBEDKRMOUNT MISSING")
			}
			Logga(ctx, os.Getenv("JsonLog"), "")

			/* ************************************************************************************************ */

			/* ************************************************************************************************ */
			// KUBEDKRRESOURCE
			Logga(ctx, os.Getenv("JsonLog"), "Getting KUBEDKRRESOURCE")
			argsSrc := make(map[string]string)
			argsSrc["source"] = "devops-8"
			argsSrc["$select"] = "XKUBEDKRRESOURCE04,XKUBEDKRRESOURCE05,XKUBEDKRRESOURCE06,XKUBEDKRRESOURCE07"
			argsSrc["center_dett"] = "dettaglio"
			argsSrc["$filter"] = "equals(XKUBEDKRRESOURCE03,'" + resourceTmpl + "') "

			restyKubeSrcRes, errKubeSrcRes := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsSrc, "ms"+devops, "/"+devops+"/KUBEDKRRESOURCE", devopsToken, dominio, coreApiVersion)
			if errKubeSrcRes != nil {
				Logga(ctx, os.Getenv("JsonLog"), errKubeSrcRes.Error())
				return microservices, errKubeSrcRes
			}

			if len(restyKubeSrcRes.BodyJson) > 0 {
				var resource Resource

				resource.CpuReq = restyKubeSrcRes.BodyJson["XKUBEDKRRESOURCE04"].(string) //   -- cpu res
				resource.MemReq = restyKubeSrcRes.BodyJson["XKUBEDKRRESOURCE05"].(string) //   -- mem res
				resource.CpuLim = restyKubeSrcRes.BodyJson["XKUBEDKRRESOURCE06"].(string) //   -- cpu limit
				resource.MemLim = restyKubeSrcRes.BodyJson["XKUBEDKRRESOURCE07"].(string) //   -- mem limit

				pod.Resource = resource
				Logga(ctx, os.Getenv("JsonLog"), "KUBEDKRRESOURCE OK")
			} else {
				erro := errors.New("KUBEDKRRESOURCE MISSING")
				return microservices, erro
			}
			Logga(ctx, os.Getenv("JsonLog"), "")

			/* ************************************************************************************************ */

			/* ************************************************************************************************ */

			// KUBEDKRPROBE
			Logga(ctx, os.Getenv("JsonLog"), "Getting KUBEDKRPROBE")
			argsProbes := make(map[string]string)
			argsProbes["source"] = "devops-8"
			//argsProbes["$select"] = "XKUBEDKRPROBE04,XKUBEDKRPROBE05,XKUBEDKRPROBE06,XKUBEDKRPROBE07,XKUBEDKRPROBE08,XKUBEDKRPROBE09,XKUBEDKRPROBE10"
			//argsProbes["$select"] += "XKUBEDKRPROBE11,XKUBEDKRPROBE12,XKUBEDKRPROBE13,XKUBEDKRPROBE14,XKUBEDKRPROBE15,XKUBEDKRPROBE16,XKUBEDKRPROBE17,XKUBEDKRPROBE18,XKUBEDKRPROBE19"
			argsProbes["center_dett"] = "allviews"
			argsProbes["$filter"] = "equals(XKUBEDKRPROBE03,'" + docker + "') "

			restyKubePrbRes, errKubePrbRes := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsProbes, "ms"+devops, "/"+devops+"/KUBEDKRPROBE", devopsToken, dominio, coreApiVersion)
			if errKubePrbRes != nil {
				Logga(ctx, os.Getenv("JsonLog"), errKubePrbRes.Error())
				return microservices, errKubePrbRes
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

				//Logga(ctx, os.Getenv("JsonLog"), probes)
				Logga(ctx, os.Getenv("JsonLog"), "KUBEDKRPROBE OK")
			} else {
				Logga(ctx, os.Getenv("JsonLog"), "KUBEDKRPROBE MISSING")
			}

			Logga(ctx, os.Getenv("JsonLog"), "")

			/* ************************************************************************************************ */

			/* ************************************************************************************************ */
			// KUBESERVICEDKR
			Logga(ctx, os.Getenv("JsonLog"), "Getting KUBESERVICEDKR")
			argsSrvDkr := make(map[string]string)
			argsSrvDkr["source"] = "devops-8"
			argsSrvDkr["$select"] = "XKUBESERVICEDKR06,XKUBESERVICEDKR05"
			argsSrvDkr["center_dett"] = "visualizza"
			argsSrvDkr["$filter"] = "equals(XKUBESERVICEDKR04,'" + docker + "') "

			restyKubeSrvDkrRes, errKubeSrvDkrRes := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsSrvDkr, "ms"+devops, "/"+devops+"/KUBESERVICEDKR", devopsToken, dominio, coreApiVersion)
			if errKubeSrvDkrRes != nil {
				Logga(ctx, os.Getenv("JsonLog"), errKubeSrvDkrRes.Error())
				return microservices, errKubeSrvDkrRes
			}

			if len(restyKubeSrvDkrRes.BodyArray) > 0 {
				var port, tipo string
				var services []Service
				for _, x := range restyKubeSrvDkrRes.BodyArray {

					port = strconv.FormatFloat(x["XKUBESERVICEDKR06"].(float64), 'f', 0, 64)
					tipo = x["XKUBESERVICEDKR05"].(string)

					/* ************************************************************************************************ */
					// ENDPOINTS
					Logga(ctx, os.Getenv("JsonLog"), "ENDPOINTS")
					sqlEndpoint := ""

					// per ogni servizio cerco gli endpoints
					sqlEndpoint += "select "
					sqlEndpoint += "ifnull(aa.XKUBEENDPOINT05, '') as microservice_src, "
					sqlEndpoint += "ifnull(cc.XKUBESERVICEDKR04, '') as docker_src, "
					sqlEndpoint += "ifnull(aa.XKUBEENDPOINT10, '') as type_src, "
					sqlEndpoint += "ifnull(aa.XKUBEENDPOINT09, '') as route_src, "
					sqlEndpoint += "ifnull(aa.XKUBEENDPOINT11, '') as rewrite_src, "
					sqlEndpoint += "ifnull(cc.XKUBESERVICEDKR06, '') as port_src, "
					sqlEndpoint += "ifnull(aa.XKUBEENDPOINT12, '') as priority, "
					sqlEndpoint += "(case when ifnull(bb_ext.XKUBEENDPOINT05, '') != '' THEN ifnull(bb_ext.XKUBEENDPOINT05, '') ELSE ifnull(bb.XKUBEENDPOINT05, '') END) as microservice_dst, "
					sqlEndpoint += "(case when ifnull(bb_ext.XKUBESERVICEDKR03, '') != '' THEN ifnull(bb_ext.XKUBESERVICEDKR03, '') ELSE ifnull(bb.XKUBESERVICEDKR03, '') END) as docker_dst, "
					sqlEndpoint += "(case when ifnull(bb_ext.XKUBEENDPOINT10, '') != '' THEN ifnull(bb_ext.XKUBEENDPOINT10, '') ELSE ifnull(bb.XKUBEENDPOINT10, '') END) as type_dst, "
					sqlEndpoint += "(case when ifnull(bb_ext.XKUBEENDPOINT09, '') != '' THEN ifnull(bb_ext.XKUBEENDPOINT09, '') ELSE ifnull(bb.XKUBEENDPOINT09, '') END) as route_dst, "
					sqlEndpoint += "(case when ifnull(bb_ext.XKUBEENDPOINT11, '') != '' THEN ifnull(bb_ext.XKUBEENDPOINT11, '') ELSE ifnull(bb.XKUBEENDPOINT11, '') END) as rewrite_dst, "
					sqlEndpoint += "(case when ifnull(bb_ext.XKUBEMICROSERV04, '') != '' THEN ifnull(bb_ext.XKUBEMICROSERV04, '') ELSE ifnull(bb.XKUBEMICROSERV04, '') END) as namespace_dst, "
					sqlEndpoint += "(case when ifnull(bb_ext.XDEPLOYLOG05, '') != '' THEN ifnull(bb_ext.XDEPLOYLOG05, '') ELSE ifnull(bb.XDEPLOYLOG05, '') END) as version_dst, "
					sqlEndpoint += "(case when ifnull(bb_ext.XKUBECLUSTER15, '') != '' THEN ifnull(bb_ext.XKUBECLUSTER15, '') ELSE ifnull(bb.XKUBECLUSTER15, '') END) as domain, "
					sqlEndpoint += "(case when ifnull(bb_ext.XKUBEENDPOINTOVR06, '') != '' THEN ifnull(bb_ext.XKUBEENDPOINTOVR06, '') ELSE ifnull(bb.XKUBEENDPOINTOVR06, '') END) as market, "
					sqlEndpoint += "(case when ifnull(bb_ext.XKUBEENDPOINTOVR07, '') != '' THEN ifnull(bb_ext.XKUBEENDPOINTOVR07, '') ELSE ifnull(bb.XKUBEENDPOINTOVR07, '') END) as partner, "
					sqlEndpoint += "(case when ifnull(bb_ext.XKUBEENDPOINTOVR08, '') != '' THEN ifnull(bb_ext.XKUBEENDPOINTOVR08, '') ELSE ifnull(bb.XKUBEENDPOINTOVR08, '') END) as customer, "
					sqlEndpoint += "(case when ifnull(bb_ext.XDEPLOYLOG05_CURRENT,'') != '' THEN '1' ELSE '0' END) as use_current_cluster_domain "
					sqlEndpoint += "from "
					sqlEndpoint += "TB_ANAG_KUBEENDPOINT00 aa "
					sqlEndpoint += "JOIN TB_ANAG_KUBESERVICEDKR00 cc on "
					sqlEndpoint += "(cc.XKUBESERVICEDKR03 = aa.XKUBEENDPOINT06) "
					sqlEndpoint += "left join ( "
					sqlEndpoint += "select "
					sqlEndpoint += "XKUBECLUSTER15, "
					sqlEndpoint += "XKUBEENDPOINT03, "
					sqlEndpoint += "XKUBEENDPOINT09, "
					sqlEndpoint += "XKUBEENDPOINT10, "
					sqlEndpoint += "XKUBEENDPOINT11, "
					sqlEndpoint += "XKUBEENDPOINT05, "
					sqlEndpoint += "XKUBEMICROSERV05, "
					sqlEndpoint += "XKUBEMICROSERV04, "
					sqlEndpoint += "XKUBEENDPOINTOVR03, "
					sqlEndpoint += "XKUBESERVICEDKR04, "
					sqlEndpoint += "XKUBESERVICEDKR03, "
					sqlEndpoint += "XDEPLOYLOG05, "
					sqlEndpoint += "'' as XDEPLOYLOG05_CURRENT, "
					sqlEndpoint += "XKUBEENDPOINTOVR06, "
					sqlEndpoint += "XKUBEENDPOINTOVR07, "
					sqlEndpoint += "XKUBEENDPOINTOVR08 "
					sqlEndpoint += "from "
					sqlEndpoint += "TB_ANAG_KUBEENDPOINT00 a "
					sqlEndpoint += "join TB_ANAG_KUBEENDPOINTOVR00 b on "
					sqlEndpoint += "(a.XKUBEENDPOINT03 = b.XKUBEENDPOINTOVR04) "
					sqlEndpoint += "join TB_ANAG_KUBEMICROSERV00 on "
					sqlEndpoint += "(XKUBEMICROSERV05 = XKUBEENDPOINT05) "
					sqlEndpoint += "join TB_ANAG_KUBESERVICEDKR00 on "
					sqlEndpoint += "(XKUBESERVICEDKR03 = XKUBEENDPOINT06) "
					sqlEndpoint += "JOIN TB_ANAG_KUBEIMICROSERV00 on "
					sqlEndpoint += "(XKUBEENDPOINT05 = XKUBEIMICROSERV04 and XKUBEIMICROSERV05 = '" + cluster + "' ) "
					sqlEndpoint += "JOIN TB_ANAG_KUBECLUSTER00 on "
					sqlEndpoint += "(XKUBECLUSTER03 = XKUBEIMICROSERV05) "
					sqlEndpoint += "JOIN TB_ANAG_DEPLOYLOG00 on "
					sqlEndpoint += "(XDEPLOYLOG04 = XKUBEIMICROSERV03 "
					sqlEndpoint += "and XDEPLOYLOG09 = 'prod' "
					sqlEndpoint += "and XDEPLOYLOG03 = 'production' "
					sqlEndpoint += "and XDEPLOYLOG06 = 1 "
					sqlEndpoint += "and XDEPLOYLOG07 = 0) ) bb on "
					sqlEndpoint += "(aa.XKUBEENDPOINT03 = bb.XKUBEENDPOINTOVR03 ) "
					sqlEndpoint += "left join ( "
					sqlEndpoint += "select "
					sqlEndpoint += "XKUBECLUSTER15, "
					sqlEndpoint += "XKUBEENDPOINT03, "
					sqlEndpoint += "XKUBEENDPOINT09, "
					sqlEndpoint += "XKUBEENDPOINT10, "
					sqlEndpoint += "XKUBEENDPOINT11, "
					sqlEndpoint += "XKUBEENDPOINT05, "
					sqlEndpoint += "XKUBEMICROSERV05, "
					sqlEndpoint += "XKUBEMICROSERV04, "
					sqlEndpoint += "XKUBEENDPOINTOVR03, "
					sqlEndpoint += "XKUBESERVICEDKR04, "
					sqlEndpoint += "XKUBESERVICEDKR03, "
					sqlEndpoint += "Q01_DEPLOYLOG.XDEPLOYLOG05, "
					sqlEndpoint += "CURRENT_DEPLOYLOG.XDEPLOYLOG05 as XDEPLOYLOG05_CURRENT, "
					sqlEndpoint += "XKUBEENDPOINTOVR06, "
					sqlEndpoint += "XKUBEENDPOINTOVR07, "
					sqlEndpoint += "XKUBEENDPOINTOVR08 "
					sqlEndpoint += "from "
					sqlEndpoint += "TB_ANAG_KUBEENDPOINT00 a_ext "
					sqlEndpoint += "join TB_ANAG_KUBEENDPOINTOVR00 b_ext on "
					sqlEndpoint += "(a_ext.XKUBEENDPOINT03 = b_ext.XKUBEENDPOINTOVR04) "
					sqlEndpoint += "join devops_data.TB_ANAG_KUBEMICROSERV00 on "
					sqlEndpoint += "(XKUBEMICROSERV05 = XKUBEENDPOINT05) "
					sqlEndpoint += "join devops_data.TB_ANAG_KUBESERVICEDKR00 on "
					sqlEndpoint += "(XKUBESERVICEDKR03 = XKUBEENDPOINT06) "
					sqlEndpoint += "JOIN devops_data.TB_ANAG_KUBEIMICROSERV00 on "
					sqlEndpoint += "(XKUBEENDPOINT05 = XKUBEIMICROSERV04 and XKUBEIMICROSERV05 = '" + os.Getenv("clusterKube8") + "' ) "
					sqlEndpoint += "JOIN devops_data.TB_ANAG_KUBECLUSTER00 on "
					sqlEndpoint += "(XKUBECLUSTER03 = XKUBEIMICROSERV05) "
					sqlEndpoint += "JOIN devops_data.TB_ANAG_DEPLOYLOG00 Q01_DEPLOYLOG on "
					sqlEndpoint += "(Q01_DEPLOYLOG.XDEPLOYLOG04 = XKUBEIMICROSERV03 "
					sqlEndpoint += "and Q01_DEPLOYLOG.XDEPLOYLOG09 = 'prod' "
					sqlEndpoint += "and Q01_DEPLOYLOG.XDEPLOYLOG03 = 'production' "
					sqlEndpoint += "and Q01_DEPLOYLOG.XDEPLOYLOG06 = 1 "
					sqlEndpoint += "and Q01_DEPLOYLOG.XDEPLOYLOG07 = 0)  "
					sqlEndpoint += "LEFT JOIN TB_ANAG_DEPLOYLOG00 CURRENT_DEPLOYLOG on "
					sqlEndpoint += "(CURRENT_DEPLOYLOG.XDEPLOYLOG04 = REPLACE(XKUBEIMICROSERV03, '" + os.Getenv("clusterKube8") + "', '" + cluster + "') "
					sqlEndpoint += "and CURRENT_DEPLOYLOG.XDEPLOYLOG09 = 'prod' "
					sqlEndpoint += "and CURRENT_DEPLOYLOG.XDEPLOYLOG03 = 'production' "
					sqlEndpoint += "and CURRENT_DEPLOYLOG.XDEPLOYLOG06 = 1 "
					sqlEndpoint += "and CURRENT_DEPLOYLOG.XDEPLOYLOG07 = 0)  "
					sqlEndpoint += ") bb_ext on "
					sqlEndpoint += "(aa.XKUBEENDPOINT03 = bb_ext.XKUBEENDPOINTOVR03 ) "
					sqlEndpoint += "having "
					sqlEndpoint += "1>0 "
					sqlEndpoint += "and docker_src = '" + docker + "' "
					sqlEndpoint += "and port_src = '" + port + "' "
					sqlEndpoint += "order by "
					sqlEndpoint += "length(priority), "
					sqlEndpoint += "priority, "
					sqlEndpoint += "route_src, "
					sqlEndpoint += "customer desc , "
					sqlEndpoint += "partner desc , "
					sqlEndpoint += "market desc "

					argsEndpoint := make(map[string]string)
					argsEndpoint["source"] = "devops-8"
					argsEndpoint["$fullquery"] = sqlEndpoint

					restyKubeEndpointRes, _ := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsEndpoint, "ms"+devops, "/core/custom/KUBEENDPOINT/values", devopsToken, dominio, coreApiVersion)
					if restyKubeEndpointRes.Errore < 0 {

					}

					var endpoints []Endpoint
					if len(restyKubeEndpointRes.BodyArray) > 0 {
						for _, x := range restyKubeEndpointRes.BodyArray {

							var endpoint Endpoint

							endpoint.Priority = x["priority"].(string)

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
							if x["use_current_cluster_domain"].(string) == "1" {
								endpoint.ClusterDomain = clusterHost
							} else {
								endpoint.ClusterDomain = ""
							}

							endpoints = append(endpoints, endpoint)

						}
						Logga(ctx, os.Getenv("JsonLog"), "ENDPOINTS OK")
					} else {
						Logga(ctx, os.Getenv("JsonLog"), "ENDPOINTS MISSING")
					}
					Logga(ctx, os.Getenv("JsonLog"), "")

					/* ************************************************************************************************ */

					var service Service
					service.Port = port
					service.Tipo = tipo
					service.Endpoint = endpoints

					services = append(services, service)

				}

				pod.Service = services

				Logga(ctx, os.Getenv("JsonLog"), "KUBESERVICEDKR OK")
			} else {
				erro := errors.New("KUBESERVICEDKR MISSING")
				return microservices, erro
			}
			Logga(ctx, os.Getenv("JsonLog"), "")

			// aggiungo pod corrente a pods
			pods = append(pods, pod)

			microservices.Pod = pods

		}

		Logga(ctx, os.Getenv("JsonLog"), "SELKUBEDKRLIST OK")
	} else {
		erro := errors.New("SELKUBEDKRLIST MISSING")
		return microservices, erro
	}
	Logga(ctx, os.Getenv("JsonLog"), "")

	//LogJson(microservices)
	Logga(ctx, os.Getenv("JsonLog"), "Seek Microservice details ok")
	Logga(ctx, os.Getenv("JsonLog"), " - - - - - - - - - - - - - - -  ")
	Logga(ctx, os.Getenv("JsonLog"), "")
	// fmt.Println(microservices)
	// LogJson(microservices)
	// os.Exit(0)
	return microservices, erro
}
func GetTenant(ctx context.Context, token, dominio, coreApiVersion string) ([]Tenant, error) {
	//Logga(ctx, os.Getenv("JsonLog"), "Get TENANT")

	var erro error
	var tenants []Tenant
	var tenant Tenant

	args := make(map[string]string)

	tenantRes, errtenantRes := ApiCallGET(ctx, os.Getenv("RestyDebug"), args, "msauth", "/auth/tenants", token, dominio, coreApiVersion)
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
	infoRes, _ := ApiCallGET(ctx, os.Getenv("RestyDebug"), args, "msauth", "/auth/getProfileInfo", token, dominio, coreApiVersion)

	if len(infoRes.BodyJson) > 0 {
		restyProfileInfoResponse := ProfileInfo{}

		b, _ := json.Marshal(infoRes.BodyJson)
		json.Unmarshal(b, &restyProfileInfoResponse)

		info["market"] = restyProfileInfoResponse.Session.Market.Decval
		info["gruppo"] = restyProfileInfoResponse.Session.GrantSession.Gruppo
		info["nome"] = restyProfileInfoResponse.Session.GrantSession.NomeCognome
		info["email"] = restyProfileInfoResponse.Session.GrantSession.Email

		Logga(ctx, os.Getenv("JsonLog"), "GetProfileInfo OK")
	} else {
		erro = errors.New("GetProfileInfo MISSING")
		Logga(ctx, os.Getenv("JsonLog"), "GetProfileInfo MISSING")
	}

	return info, erro
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

	restyKubeImicroservRes, errKubeImicroservRes := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsImicro, "msdevops", "/devops/KUBEDKRBUILD", devopsToken, loginApiDomain, coreApiVersion)
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

	restyKubeImicroservRes, errKubeImicroservRes := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsImicro, "msdevops", "/devops/KUBETEAMBRANCH", devopsToken, loginApiDomain, coreApiVersion)
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
	Logga(ctx, os.Getenv("JsonLog"), "git repo: "+buildArgs.TypeGit)

	// OTTENGO L' HASH del branch vivo
	clientBranch := resty.New()
	clientBranch.Debug = true
	var respBranch, restyResponse *resty.Response
	var errBranch, errTag error
	switch buildArgs.TypeGit {
	case "", "bitbucket":
		respBranch, errBranch = clientBranch.R().
			EnableTrace().
			SetBasicAuth(buildArgs.UserGit, buildArgs.TokenGit).
			Get(buildArgs.ApiHostGit + "/repositories/" + buildArgs.ProjectGit + "/" + repo + "/refs/branches/" + buildArgs.SprintBranch)
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
			client.Debug = false
			restyResponse, errTag = client.R().
				SetHeader("Content-Type", "application/json").
				SetBasicAuth(buildArgs.UserGit, buildArgs.TokenGit).
				SetBody(body).
				Post(buildArgs.ApiHostGit + "/repositories/" + buildArgs.ProjectGit + "/" + repo + "/refs/tags")
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
			client.Debug = true
			restyResponse, errTag = client.R().
				SetHeader("Content-Type", "application/json").
				SetHeader("Accept", "application/vnd.github+json").
				SetBasicAuth(buildArgs.UserGit, buildArgs.TokenGit).
				SetBody(body).
				Post(buildArgs.ApiHostGit + "/repos/" + buildArgs.UserGit + "/" + repo + "/git/tags")
		}
	}

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

func GetEnvironmentStatus(ctx context.Context, cluster, enviro, microserice, customer, tenant, accessToken, loginApiDomain, coreApiVersion string) error {

	status := ""

	// cerco il token di Corefactory
	devopsToken, erroT := GetCoreFactoryToken(ctx, tenant, accessToken, loginApiDomain, coreApiVersion, os.Getenv("RestyDebug"))
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

	restyEsRes, errEsRes := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsEs, "msdevops", "/devops/KUBEENVSTATUS", devopsToken, loginApiDomain, coreApiVersion)
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
func SetEnvironmentStatus(ctx context.Context, cluster, enviro, microserice, customer, user, toggle, tenant, accessToken, loginApiDomain, coreApiVersion string) error {

	var erro error
	// cerco il token di Corefactory
	devopsToken, erroT := GetCoreFactoryToken(ctx, tenant, accessToken, loginApiDomain, coreApiVersion, os.Getenv("RestyDebug"))
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

	res := ApiCallPOST(ctx, os.Getenv("RestyDebug"), keyvalueslices, "msdevops", "/devops/KUBEENVSTATUS", devopsToken, loginApiDomain, coreApiVersion)

	if res.Errore < 0 {
		Logga(ctx, os.Getenv("JsonLog"), res.Log)
		erro = errors.New(res.Log)
		return erro

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

	restyKubeCluRes, _ := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsClu, "ms"+devops, "/"+devops+"/KUBECLUSTER", devopsToken, loginApiDomain, coreApiVersion)
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
func GetJsonDatabases(ctx context.Context, stage, developer string, market int32, arrConn MasterConn, tenant, accessToken, loginApiDomain, coreApiVersion string, monolith bool) (map[string]interface{}, error) {
	Logga(ctx, os.Getenv("JsonLog"), "Getting Json Db")

	var erro error
	callResponse := map[string]interface{}{}

	if monolith {
		callResponse["monolith"] = "true"
		return callResponse, erro
	}

	devopsToken, erroT := GetCoreFactoryToken(ctx, tenant, accessToken, loginApiDomain, coreApiVersion, os.Getenv("RestyDebug"))
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

	client := resty.New()
	client.Debug = true
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

	restyAuthResponse, restyAuthErr := ApiCallLOGIN(ctx, os.Getenv("RestyDebug"), argsAuth, "msauth", "/auth/login", dominio, coreApiVersion)
	if restyAuthErr.Errore < 0 {
		erro = errors.New(restyAuthErr.Log)
		return "", erro
	}
	if len(restyAuthResponse) > 0 {
		return restyAuthResponse["idToken"].(string), erro
	} else {
		erro = errors.New("token MISSING")
		return "", erro
	}
}

func GetCfToolEnv(ctx context.Context, token, dominio, tenant, coreApiVersion string, monolith bool) (TenantEnv, error) {

	Logga(ctx, os.Getenv("JsonLog"), "Getting KUBECFTOOLENV")

	var erro error
	devops := "devops"
	if monolith {
		devops = "devopsmono"
	}

	args := make(map[string]string)
	args["center_dett"] = "dettaglio"
	args["source"] = "devops-8"
	args["$filter"] = "equals(XKUBECFTOOLENV03,'" + dominio + "') "
	args["$filter"] += " and equals(XKUBECFTOOLENV19,'" + tenant + "') "

	envRes, _ := ApiCallGET(ctx, os.Getenv("RestyDebug"), args, "ms"+devops, "/"+devops+"/KUBECFTOOLENV", token, dominio, coreApiVersion)

	var tntEnv TenantEnv

	if len(envRes.BodyJson) > 0 {

		tntEnv.TelegramKey = envRes.BodyJson["XKUBECFTOOLENV04"].(string)
		tntEnv.TelegramID = envRes.BodyJson["XKUBECFTOOLENV05"].(string)
		tntEnv.CoreApiVersion = envRes.BodyJson["XKUBECFTOOLENV06"].(string)
		tntEnv.CoreApiPort = envRes.BodyJson["XKUBECFTOOLENV07"].(string)
		tntEnv.CoreAccessToken = envRes.BodyJson["XKUBECFTOOLENV08"].(string)
		tntEnv.AtlassianHost = envRes.BodyJson["XKUBECFTOOLENV09"].(string)
		tntEnv.AtlassianUser = envRes.BodyJson["XKUBECFTOOLENV10"].(string)
		tntEnv.AtlassianToken = envRes.BodyJson["XKUBECFTOOLENV11"].(string)
		tntEnv.ApiHostGit = envRes.BodyJson["XKUBECFTOOLENV12"].(string)
		tntEnv.UrlGit = envRes.BodyJson["XKUBECFTOOLENV20"].(string)
		tntEnv.TypeGit = envRes.BodyJson["XKUBECFTOOLENV21"].(string)
		tntEnv.UserGit = envRes.BodyJson["XKUBECFTOOLENV13"].(string)
		tntEnv.TokenGit = envRes.BodyJson["XKUBECFTOOLENV14"].(string)
		tntEnv.ProjectGit = envRes.BodyJson["XKUBECFTOOLENV15"].(string)
		tntEnv.CoreGkeProject = envRes.BodyJson["XKUBECFTOOLENV16"].(string)
		tntEnv.CoreGkeUrl = envRes.BodyJson["XKUBECFTOOLENV17"].(string)
		tntEnv.CoreApiDominio = envRes.BodyJson["XKUBECFTOOLENV18"].(string)

		Logga(ctx, os.Getenv("JsonLog"), "KUBECFTOOLENV OK")
	} else {
		Logga(ctx, os.Getenv("JsonLog"), "KUBECFTOOLENV MISSING")
		erro = errors.New("KUBECFTOOLENV MISSING")
	}

	return tntEnv, erro

}
func GetDeploymentApi(namespace, apiHost, apiToken string, scaleToZero, debug bool) (DeploymntStatus, error) {

	var erro error

	var deploy DeploymntStatus

	clientKUBE := resty.New()
	clientKUBE.Debug = debug
	clientKUBE.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	endPoint := "https://" + apiHost + "/apis/apps/v1/namespaces/" + namespace + "/deployments"
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

	debool, errBool := strconv.ParseBool(debug)
	if errBool != nil {
		return false, errBool
	}

	msDeploy := microservice + "-v" + versione
	msMatch := false
	i := 0
	for {
		item, errDpl := GetDeploymentApi(namespace, apiHost, apiToken, scaleToZero, debool)
		if errDpl != nil {
			return false, errDpl
		} else {

			if len(item.Items) == 0 {
				erro = errors.New("No Deployment Found in Namespace")
				return false, erro
			}

			for _, item := range item.Items {

				//fmt.Println(item.Metadata.Name, "-", msDeploy)
				if item.Metadata.Name == msDeploy {
					msMatch = true

					if !scaleToZero {
						var ctx context.Context
						Logga(ctx, os.Getenv("JsonLog"), item.Metadata.Name+" desired: ", strconv.Itoa(item.Spec.Replicas), " - aviable: ", strconv.Itoa(item.Status.ReadyReplicas))

						if item.Spec.Replicas == item.Status.ReadyReplicas {
							return true, erro
						}
					} else {
						return true, erro
					}
				}

				// sto girando a vuoto perche nessun item risponde a cio che cerco
				if i >= 1 && !msMatch {
					erro = errors.New("nessun item risponde a cio che cerco")
					return false, erro
				}

			}

			i++
			time.Sleep(10 * time.Second)
			if i > 150 { // superati 5 minuti direi che stann e plobble'
				erro = errors.New("Time Out. Pod is not Running")
				return false, erro
			}
		}
	}
}
func DeleteObsoleteObjects(ctx context.Context, ires IstanzaMicro, versione, canaryProduction, namespace, enviro, tenant, devopsToken, dominio, coreApiVersion string, debug string) error {

	var erro error
	istanza := ires.Istanza
	microservice := ""
	if ires.Monolith == 0 {
		microservice = ires.Microservice
	} else {
		microservice = ires.PodName
	}

	debool, errBool := strconv.ParseBool(debug)
	if errBool != nil {
		return errBool
	}

	Logga(ctx, os.Getenv("JsonLog"), "****************************************************************************************")
	Logga(ctx, os.Getenv("JsonLog"), "DELETING OBSOLETE PODS")

	/* ************************************************************************************************ */
	// DEPLOYLOG
	Logga(ctx, os.Getenv("JsonLog"), "Getting DEPLOYLOG - deleteObsoleteMonolith")

	devops := "devops"
	if ires.Monolith == 1 {
		devops = "devopsmono"
	}

	argsDeploy := make(map[string]string)
	argsDeploy["source"] = "devops-8"
	argsDeploy["$select"] = "XDEPLOYLOG03,XDEPLOYLOG05"
	argsDeploy["center_dett"] = "visualizza"
	argsDeploy["$filter"] = "equals(XDEPLOYLOG04,'" + istanza + "') "
	argsDeploy["$filter"] += " and (equals(XDEPLOYLOG03,'canary') OR equals(XDEPLOYLOG03,'production'))  "
	argsDeploy["$filter"] += " and equals(XDEPLOYLOG06,'1') "
	argsDeploy["$filter"] += " and equals(XDEPLOYLOG09,'" + enviro + "') "

	restyDeployRes, errDeployRes := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsDeploy, "ms"+devops, "/"+devops+"/DEPLOYLOG", devopsToken, dominio, coreApiVersion)
	if errDeployRes != nil {
		Logga(ctx, os.Getenv("JsonLog"), errDeployRes.Error())
		erro = errors.New(errDeployRes.Error())
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
		Logga(ctx, os.Getenv("JsonLog"), "DEPLOYLOG OK")
	} else {
		Logga(ctx, os.Getenv("JsonLog"), "DEPLOYLOG MISSING", "warn")
	}
	Logga(ctx, os.Getenv("JsonLog"), "=== NEVER DELETE INNOCENT DEPLOYMENTS === ")
	Logga(ctx, os.Getenv("JsonLog"), "eventually to kill: "+microservice)
	Logga(ctx, os.Getenv("JsonLog"), "Never kill Canary: "+versioneCanaryDb)
	Logga(ctx, os.Getenv("JsonLog"), "Never kill Production: "+versioneProductionDb)

	/* ************************************************************************************************ */

	// ho recuperato le versioni canary e production che NON cancellero MAI :D

	//msDeploy := microservice + "-v" + versione

	//LogJson(ires)
	item, errDepl := GetDeploymentApi(namespace, ires.ApiHost, ires.ApiToken, ires.ScaleToZero, debool)
	//LogJson(item)
	if errDepl != nil {
		return errDepl
	} else {

		if len(item.Items) == 0 {

			erro = errors.New("No Deployment Found in Namespace")
			return erro
		}

		Logga(ctx, os.Getenv("JsonLog"), "API returns : "+strconv.Itoa(len(item.Items))+" ITEMS")

		// se non abbiamo ne la versione canary ne quella production
		// stiamo al primo giro di giostra e non sevo cancellare anything
		if versioneCanaryDb == "" && versioneProductionDb == "" {
			Logga(ctx, os.Getenv("JsonLog"), "First Deploy - no items to kill, of course ")
			return nil
		}
		for _, item := range item.Items {

			Logga(ctx, os.Getenv("JsonLog"), "item yaml: "+item.Spec.Template.Metadata.Labels.App)
			Logga(ctx, os.Getenv("JsonLog"), "version: "+item.Spec.Template.Metadata.Labels.Version)
			Logga(ctx, os.Getenv("JsonLog"), "item ires: "+microservice)
			// primo filtro sulla refapp giusta
			if item.Spec.Template.Metadata.Labels.App == microservice {

				Logga(ctx, os.Getenv("JsonLog"), "Kill everything with different version of canary: "+versioneCanaryDb+" or production: "+versioneProductionDb+" - Current version: "+item.Spec.Selector.MatchLabels.Version)
				// secondo filtro sulle versione
				if item.Spec.Template.Metadata.Labels.Version == versioneCanaryDb || item.Spec.Template.Metadata.Labels.Version == versioneProductionDb {
					// SKIP
				} else {

					deployment := item.Spec.Template.Metadata.Labels.App + "-" + item.Spec.Template.Metadata.Labels.Version
					Logga(ctx, os.Getenv("JsonLog"), "KILL-KILL-KILL-KILL-KILL-KILL-KILL-KILL-KILL-KILL-KILL-KILL-KILL-KILL-KILL-KILL-KILL-KILL-KILL")

					Logga(ctx, os.Getenv("JsonLog"), "I DO KILL: "+deployment)

					if ires.ScaleToZero {
						DeleteObjectsApi(namespace, ires.ApiHost, ires.ApiToken, deployment, "knservice", debool)
					} else {
						// delete deployment
						DeleteObjectsApi(namespace, ires.ApiHost, ires.ApiToken, deployment, "deployment", debool)
						// delete HPA
						DeleteObjectsApi(namespace, ires.ApiHost, ires.ApiToken, deployment, "hpa", debool)
					}

				}
			}

		}

	}
	return erro
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
func GetOverrideTenantEnv(ctx context.Context, bearerToken, team string, tntEnv TenantEnv, dominio, coreApiVersion string) (TenantEnv, error) {
	// ho le env di default
	// cerco per ogni valore su KUBETEAMBRANCH se trovo sovrascrivo
	var erro error

	args := make(map[string]string)
	args["center_dett"] = "dettaglio"
	args["source"] = "devops-8"
	args["$filter"] = "equals(XKUBETEAMBRANCH03,'" + team + "') "

	envRes, _ := ApiCallGET(ctx, os.Getenv("RestyDebug"), args, "msdevops", "/devops/KUBETEAMBRANCH", bearerToken, dominio, coreApiVersion)

	if len(envRes.BodyJson) > 0 {

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
			tntEnv.ProjectGit = envRes.BodyJson["XKUBETEAMBRANCH15"].(string)
		}

		if envRes.BodyJson["XKUBETEAMBRANCH16"].(string) != "" {
			Logga(ctx, os.Getenv("JsonLog"), "overrdide 16")
			tntEnv.UrlGit = envRes.BodyJson["XKUBETEAMBRANCH16"].(string)
		}

		if envRes.BodyJson["XKUBETEAMBRANCH17"].(string) != "" {
			Logga(ctx, os.Getenv("JsonLog"), "overrdide 17")
			tntEnv.TypeGit = envRes.BodyJson["XKUBETEAMBRANCH17"].(string)
		}

		Logga(ctx, os.Getenv("JsonLog"), "KUBETEAMBRANCH OK")
	} else {
		Logga(ctx, os.Getenv("JsonLog"), "KUBETEAMBRANCH MISSING")
		erro = errors.New("KUBETEAMBRANCH MISSING")
	}
	return tntEnv, erro
}
func GetApiHostAndToken(ctx context.Context, enviro, cluster, token, loginApiHost, coreApiVersion, swmono string) (string, string, error) {
	var erro error

	args := make(map[string]string)
	args["source"] = "devops-8"
	args["$select"] = "XKUBECLUSTER16,XKUBECLUSTER18"
	args["center_dett"] = "dettaglio"
	args["$filter"] = "equals(XKUBECLUSTER03,'" + cluster + "') "

	var endpointRes CallGetResponse
	var errendpointRes error
	if swmono == "mono" {
		endpointRes, errendpointRes = ApiCallGET(ctx, os.Getenv("RestyDebug"), args, "msdevopsmono", "/devopsmono/KUBECLUSTER", token, loginApiHost, coreApiVersion)
	} else {
		endpointRes, errendpointRes = ApiCallGET(ctx, os.Getenv("RestyDebug"), args, "msdevops", "/devops/KUBECLUSTER", token, loginApiHost, coreApiVersion)
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
