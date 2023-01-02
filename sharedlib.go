package lib

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

func GetIstanceDetail(ctx context.Context, iresReq IresRequest, canaryProduction, devopsToken string) (IstanzaMicro, LoggaErrore) {

	Logga(ctx, "")
	Logga(ctx, " + + + + + + + + + + + + + + + + + + + +")
	Logga(ctx, "getIstanceDetail begin")

	var LoggaErrore LoggaErrore
	LoggaErrore.Errore = 0

	istanza := iresReq.Istanza
	enviro := iresReq.Enviro
	refAppID := iresReq.RefAppID
	refAppCustomerID := iresReq.RefAppCustomerID
	customerDomain := iresReq.CustomerDomain
	monolithArg := iresReq.Monolith
	tags := iresReq.Tags
	profileDeployStr := iresReq.Enviro

	tagsArr := []string{}
	//rendo tags un array

	tagsArrDirt := strings.Split(tags, ",")
	for _, str := range tagsArrDirt {
		if str != "" {
			tagsArr = append(tagsArr, strings.TrimSpace(str))
		}
	}

	i64, _ := strconv.ParseInt(monolithArg, 10, 32)
	monolith := int32(i64)

	profileDeploy64, _ := strconv.ParseInt(profileDeployStr, 10, 32)
	profileDeploy := int32(profileDeploy64)

	var ims IstanzaMicro
	ims.Monolith = monolith
	ims.ProfileDeploy = profileDeploy
	ims.Enviro = enviro
	ims.Tags = tagsArr
	ims.Istanza = istanza

	ims.PodName = iresReq.PodName
	ims.RefappID = refAppID

	// qui in data 19 maggio 2021
	// con davide e mauro si decide che le connessione MASTER vanno
	// definite su un config
	// devopsProfile, _ := os.LookupEnv("APP_ENV")
	//if devopsProfile == "prod" {
	ims.MasterHost = os.Getenv("hostData")
	// } else {
	// ims.MasterHost = os.Getenv("hostDataDev")
	// }
	ims.MasterName = os.Getenv("nameData")
	ims.MasterUser = os.Getenv("userData")
	ims.MasterPass = os.Getenv("passData")

	/* ************************************************************************************************ */
	// KUBEIMICROSERV
	Logga(ctx, "Getting KUBEIMICROSERV - deploy.go")
	argsImicro := make(map[string]string)
	argsImicro["source"] = "devops-8"
	argsImicro["$select"] = "XKUBEIMICROSERV04,XKUBEIMICROSERV05"
	argsImicro["center_dett"] = "dettaglio"
	argsImicro["$filter"] = "equals(XKUBEIMICROSERV03,'" + istanza + "') "

	restyKubeImicroservRes := ApiCallGET(ctx, false, argsImicro, "msdevops", "/devops/KUBEIMICROSERV", devopsToken, "")
	if restyKubeImicroservRes.Errore < 0 {
		Logga(ctx, restyKubeImicroservRes.Log)
		LoggaErrore.Errore = restyKubeImicroservRes.Errore
		LoggaErrore.Log = restyKubeImicroservRes.Log
		return ims, LoggaErrore
	}

	microservice := ""
	if len(restyKubeImicroservRes.BodyJson) > 0 {
		microservice = restyKubeImicroservRes.BodyJson["XKUBEIMICROSERV04_COD"].(string)
		ims.Cluster = restyKubeImicroservRes.BodyJson["XKUBEIMICROSERV05"].(string)
		ims.Microservice = microservice

		Logga(ctx, "KUBEIMICROSERV OK")
	} else {
		fmt.Println("aspettamma a mauro con l'orchestretor ...")

		istanzaArr := strings.Split(istanza, "-"+enviro+"-")

		if len(istanzaArr) >= 2 {

			_cluster := istanzaArr[0]

			_msDirt := istanzaArr[len(istanzaArr)-1]

			msArr := strings.Split(_msDirt, "-")
			_ms := msArr[len(msArr)-1]

			if monolith == 1 {
				_ms = "msrefappmonolith"
			}

			ims.Microservice = _ms
			microservice = _ms
			ims.Cluster = _cluster

			keyvalueslices := make([]map[string]interface{}, 0)
			keyvalueslice := make(map[string]interface{})
			keyvalueslice["debug"] = true
			keyvalueslice["source"] = "devops-8"
			keyvalueslice["XKUBEIMICROSERV03"] = istanza
			keyvalueslice["XKUBEIMICROSERV04"] = _ms
			keyvalueslice["XKUBEIMICROSERV05"] = _cluster
			keyvalueslice["XKUBEIMICROSERV06"] = enviro

			keyvalueslices = append(keyvalueslices, keyvalueslice)

			resKubeims := ApiCallPOST(ctx, false, keyvalueslices, "msdevops", "/devops/KUBEIMICROSERV", devopsToken, "")

			if resKubeims.Errore < 0 {
				Logga(ctx, "")
				Logga(ctx, "NON RIESCO A SCRIVRERE L'ISTANZA "+resKubeims.Log)
				Logga(ctx, "")

				LoggaErrore.Errore = resKubeims.Errore
				LoggaErrore.Log = resKubeims.Log
				return ims, LoggaErrore
			}

		} else {

			Logga(ctx, "KUBEIMICROSERV MISSING")
		}
	}
	Logga(ctx, "")

	/* ************************************************************************************************ */
	// KUBECLUSTER
	//
	// FAC-563
	// gestione MASTER HOST USER E PASSWD per OGNI CLUSTER
	Logga(ctx, "Getting KUBECLUSTER")

	argsClu := make(map[string]string)
	argsClu["source"] = "devops-8"
	// argsClu["$select"] = "XKUBECLUSTER03,XKUBECLUSTER05,XKUBECLUSTER06,XKUBECLUSTER08,XKUBECLUSTER09,XKUBECLUSTER10,"
	// argsClu["$select"] += "XKUBECLUSTER11,XKUBECLUSTER12,XKUBECLUSTER15,XKUBECLUSTER16,XKUBECLUSTER17,XKUBECLUSTER18,XKUBECLUSTER20,XKUBECLUSTER21,XKUBECLUSTER22"
	argsClu["center_dett"] = "allviews"
	//["$filter"] = "equals(XKUBECLUSTER03,'" + ims.cluster + "') "

	restyKubeCluRes := ApiCallGET(ctx, false, argsClu, "msdevops", "/devops/KUBECLUSTER", devopsToken, "")
	if restyKubeCluRes.Errore < 0 {
		Logga(ctx, restyKubeCluRes.Log)
		LoggaErrore.Errore = restyKubeCluRes.Errore
		LoggaErrore.Log = restyKubeCluRes.Log
		return ims, LoggaErrore
	}

	profile := ""
	var profileNum int
	var clu ClusterSt
	clus := make(map[string]ClusterSt, 0)
	if len(restyKubeCluRes.BodyArray) > 0 {

		for _, x := range restyKubeCluRes.BodyArray {

			clu.ProjectID = x["XKUBECLUSTER05"].(string)
			clu.Owner = x["XKUBECLUSTER06"].(string)

			clu.ApiHost = x["XKUBECLUSTER16"].(string)
			clu.ApiToken = x["XKUBECLUSTER18"].(string)

			profileFloat := x["XKUBECLUSTER12"].(float64)
			profileNum = int(profileFloat)

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
			clu.Profile = profile
			clu.ProfileInt = int32(profileNum)

			clu.Domain = x["XKUBECLUSTER15"].(string)
			clu.Token = x["XKUBECLUSTER20"].(string)
			clu.MasterHost = x["XKUBECLUSTER09"].(string)
			clu.MasterUser = x["XKUBECLUSTER10"].(string)
			clu.MasterPasswd = x["XKUBECLUSTER11"].(string)
			clu.AccessToken = x["XKUBECLUSTER20"].(string)

			ambienteFloat := x["XKUBECLUSTER12"].(float64)
			clu.Ambiente = int32(ambienteFloat)
			clu.RefappID = x["XKUBECLUSTER22"].(string)
			clu.SwMultiEnvironment = x["XKUBECLUSTER17"].(string)

			clu.CloudNet = x["XKUBECLUSTER24"].(string)

			clu.Autopilot = strconv.FormatFloat(x["XKUBECLUSTER14"].(float64), 'f', 0, 64)

			// PORCATA PER FATICARE AL VOLO SU KEEPUP-STAGE
			// clu.SwMultiEnvironment = "1"

			/**
			Andiamo a vedere se esiste un record in KUBECLUSTERENV che fa l'overwrite di alcune proprietà di
			KUBECLUSTER in base all'env
			**/

			argsCluEnv := make(map[string]string)
			argsCluEnv["source"] = "devops-8"
			argsCluEnv["center_dett"] = "dettaglio"
			argsCluEnv["$filter"] = "equals(XKUBECLUSTERENV03,'" + x["XKUBECLUSTER03"].(string) + "') "
			argsCluEnv["$filter"] += " and equals(XKUBECLUSTERENV04,'" + clu.Owner + "') "
			argsCluEnv["$filter"] += " and XKUBECLUSTERENV05 eq " + strconv.Itoa(int(ambienteFloat)) + " "
			argsCluEnv["$filter"] += " and equals(XKUBECLUSTERENV06,'" + enviro + "') "

			restyKubeCluEnvRes := ApiCallGET(ctx, false, argsCluEnv, "msdevops", "/devops/KUBECLUSTERENV", devopsToken, "")
			if restyKubeCluEnvRes.Errore < 0 {
				Logga(ctx, restyKubeCluEnvRes.Log)
				LoggaErrore.Errore = restyKubeCluEnvRes.Errore
				LoggaErrore.Log = restyKubeCluEnvRes.Log
				return ims, LoggaErrore
			}

			if len(restyKubeCluEnvRes.BodyJson) > 0 {
				clu.Domain = restyKubeCluEnvRes.BodyJson["XKUBECLUSTERENV08"].(string)
				clu.RefappID = restyKubeCluEnvRes.BodyJson["XKUBECLUSTERENV09"].(string)
				clu.MasterHost = restyKubeCluEnvRes.BodyJson["XKUBECLUSTERENV07"].(string)

				Logga(ctx, "KUBECLUSTERENV OK")
			}

			// if x["XKUBECLUSTER03"].(string) == "keepup-prod" && enviro == "uat" {
			// 	clu.Domain = "ms-uat.keepup-pro.it"
			// 	clu.RefappID = "00000UAT-0000-0000-0000-000CUSTOMER"
			// 	clu.MasterHost = "kauat_admin"
			// }

			clus[x["XKUBECLUSTER03"].(string)] = clu

			// Logga(ctx, x["XKUBECLUSTER03"].(string))
			// LogJson(clu)

		}

		//ims.Clusters = clus

		Logga(ctx, "KUBECLUSTER OK")
	} else {
		Logga(ctx, "KUBECLUSTER MISSING")
	}
	Logga(ctx, "")

	/* ************************************************************************************************ */

	/* ************************************************************************************************ */
	// KUBECLUSTER
	Logga(ctx, "Getting KUBECLUSTER")

	if _, ok := clus[ims.Cluster]; ok {

		Logga(ctx, " +++ KUBECLUSTER OK +++ ")

		// fmt.Println(ims.cluster)
		// LogJson(clus[ims.cluster])

		ims.ProjectID = clus[ims.Cluster].ProjectID
		ims.Owner = clus[ims.Cluster].Owner
		ims.Profile = clus[ims.Cluster].Profile
		ims.ProfileInt = clus[ims.Cluster].ProfileInt
		ims.ClusterDomain = clus[ims.Cluster].Domain
		ims.Token = clus[ims.Cluster].Token
		ims.MasterUser = clus[ims.Cluster].MasterUser
		ims.MasterPass = clus[ims.Cluster].MasterPasswd
		ims.Ambiente = clus[ims.Cluster].Ambiente
		ims.ClusterRefAppID = clus[ims.Cluster].RefappID
		ims.ClusterRefAppID = clus[ims.Cluster].RefappID
		ims.SwMultiEnvironment = clus[ims.Cluster].SwMultiEnvironment
		ims.ApiHost = clus[ims.Cluster].ApiHost
		ims.ApiToken = clus[ims.Cluster].ApiToken
		ims.Autopilot = clus[ims.Cluster].Autopilot
		ims.CloudNet = clus[ims.Cluster].CloudNet

		Logga(ctx, "KUBECLUSTER OK")
	} else {
		Logga(ctx, "KUBECLUSTER MISSING")
	}
	Logga(ctx, "")

	/* ************************************************************************************************ */
	// AMBDOMAIN

	Logga(ctx, "Getting AMBDOMAIN")

	argsAmbdomain := make(map[string]string)
	argsAmbdomain["source"] = "auth-1"
	argsAmbdomain["$select"] = "XAMBDOMAIN05,XAMBDOMAIN07,XAMBDOMAIN08,XAMBDOMAIN09,XAMBDOMAIN10,XAMBDOMAIN11"
	argsAmbdomain["center_dett"] = "dettaglio"
	if refAppCustomerID != "" { // oggi 18 maggio 2021 davide afferma che questo è un pezzotto
		argsAmbdomain["$filter"] = "equals(XAMBDOMAIN05,'" + refAppCustomerID + "') and "
	}
	argsAmbdomain["$filter"] += "  equals(XAMBDOMAIN04,'" + customerDomain + "') "
	restyAmbdomainRes := ApiCallGET(ctx, false, argsAmbdomain, "msauth", "/core/AMBDOMAIN", devopsToken, "")
	if restyAmbdomainRes.Errore < 0 {
		Logga(ctx, restyAmbdomainRes.Log)
		LoggaErrore.Errore = restyAmbdomainRes.Errore
		LoggaErrore.Log = restyAmbdomainRes.Log
		return ims, LoggaErrore
	}

	if len(restyAmbdomainRes.BodyJson) > 0 {
		ims.CustomerSalt = restyAmbdomainRes.BodyJson["XAMBDOMAIN11"].(string)
		ims.RefappCustomerID = restyAmbdomainRes.BodyJson["XAMBDOMAIN05"].(string)
		ims.MasterHostData = restyAmbdomainRes.BodyJson["XAMBDOMAIN07"].(string)
		ims.MasterHostMeta = restyAmbdomainRes.BodyJson["XAMBDOMAIN07"].(string)
		Logga(ctx, "AMBDOMAIN OK")
	} else {
		Logga(ctx, "AMBDOMAIN MISSING")
	}
	Logga(ctx, "")
	/* ************************************************************************************************ */
	// KUBEMICROSERV
	Logga(ctx, "Getting KUBEMICROSERV")

	argsMS := make(map[string]string)
	argsMS["source"] = "devops-8"
	argsMS["$select"] = "XKUBEMICROSERV09,XKUBEMICROSERV15"
	argsMS["center_dett"] = "dettaglio"
	argsMS["$filter"] = "equals(XKUBEMICROSERV05,'" + microservice + "') "
	restyKubeMSRes := ApiCallGET(ctx, false, argsMS, "msdevops", "/devops/KUBEMICROSERV", devopsToken, "")
	if restyKubeMSRes.Errore < 0 {
		Logga(ctx, restyKubeMSRes.Log)
		LoggaErrore.Errore = restyKubeMSRes.Errore
		LoggaErrore.Log = restyKubeMSRes.Log
		return ims, LoggaErrore
	}

	if len(restyKubeMSRes.BodyJson) > 0 {
		var swCoreBool bool
		swCoreFloat := restyKubeMSRes.BodyJson["XKUBEMICROSERV09"].(float64)
		if swCoreFloat == 0 {
			swCoreBool = false
		} else {
			swCoreBool = true
		}
		ims.SwCore = swCoreBool

		swDb := int(restyKubeMSRes.BodyJson["XKUBEMICROSERV15"].(float64))

		ims.SwDb = swDb
		Logga(ctx, "KUBEMICROSERV OK")
	} else {
		Logga(ctx, "KUBEMICROSERV MISSING")
	}
	Logga(ctx, "")

	/* ************************************************************************************************ */
	// AMB
	Logga(ctx, "Getting AMB")

	versionAmb := ""
	if canaryProduction == "canary" {
		versionAmb = "1"
	} else {
		versionAmb = "0"
	}

	argsAmb := make(map[string]string)
	argsAmb["microservice"] = microservice
	argsAmb["enviro"] = enviro
	argsAmb["version"] = versionAmb
	argsAmb["cluster"] = ims.Cluster
	if ims.Monolith == 1 {
		argsAmb["refappID"] = ims.PodName // MWPO DICE CHE ANCHE SE CE SCRITTO refappID é GIUSTO PASSARE IL PODNAME
		argsAmb["customerDomain"] = customerDomain
	}
	argsAmb["monolith"] = strconv.Itoa(int(ims.Monolith))
	argsAmb["env"] = strconv.Itoa(int(ims.ProfileInt))
	argsAmb["swMultiEnvironment"] = ims.SwMultiEnvironment

	restyKubeAmbRes := ApiCallGET(ctx, true, argsAmb, "msauth", "/auth/getAmbDomainMs", devopsToken, "")
	if restyKubeAmbRes.Errore < 0 {
		Logga(ctx, restyKubeAmbRes.Log)
		LoggaErrore.Errore = restyKubeAmbRes.Errore
		LoggaErrore.Log = restyKubeAmbRes.Log
		return ims, LoggaErrore
	}

	var dbMetaConnMss []DbMetaConnMs
	var dbDataConnMss []DbDataConnMs
	var attributiMSs []AttributiMS

	if len(restyKubeAmbRes.BodyArray) > 0 {
		for _, x := range restyKubeAmbRes.BodyArray {

			var attributiMS AttributiMS

			attributiMS.Customer = x["XAMB14"].(string)
			attributiMS.Market = strconv.FormatFloat(x["XAMB11"].(float64), 'f', 0, 64)
			attributiMS.Partner = x["XAMB25"].(string)

			attributiMSs = append(attributiMSs, attributiMS)

			// mster host lo ottengo da AMB uno per meta uno per data
			// ed e identico per tutti i microservizi - andrebbero prese da KUBECLUSTERDBTMPL
			ims.MasterHostMeta = x["XAMB03"].(string)
			ims.MasterHostData = x["XAMB07"].(string)

			ims.MasterUser = clus[x["CLUSTER"].(string)].MasterUser
			ims.MasterPass = clus[x["CLUSTER"].(string)].MasterPasswd

			//fmt.Println(x)
			var dbMetaConnMs DbMetaConnMs
			dbMetaConnMs.MetaHost = x["XAMB03"].(string)
			dbMetaConnMs.MetaName = x["XAMB04"].(string)
			dbMetaConnMs.MetaUser = x["XAMB05"].(string)
			dbMetaConnMs.MetaPass = x["XAMB06"].(string)
			dbMetaConnMs.MetaMicroAmb = x["XAMB19"].(string)
			dbMetaConnMs.Cluster = x["CLUSTER"].(string)

			var dbDataConnMs DbDataConnMs
			dbDataConnMs.DataHost = x["XAMB07"].(string)
			dbDataConnMs.DataName = x["XAMB08"].(string)
			dbDataConnMs.DataUser = x["XAMB09"].(string)
			dbDataConnMs.DataPass = x["XAMB10"].(string)
			dbDataConnMs.DataMicroAmb = x["XAMB19"].(string)
			dbDataConnMs.Cluster = x["CLUSTER"].(string)

			// fmt.Println(x["CLUSTER"].(string))
			// fmt.Println(ims.masterHostMeta)
			// fmt.Println(ims.masterUser)
			// fmt.Println(ims.masterPass)

			dbMetaConnMss = append(dbMetaConnMss, dbMetaConnMs)
			dbDataConnMss = append(dbDataConnMss, dbDataConnMs)
		}

		ims.DbMetaConnMs = dbMetaConnMss
		ims.DbDataConnMs = dbDataConnMss
		ims.AttributiMS = attributiMSs

		//fmt.Println(dbMetaConnMss, dbDataConnMss, attributiMSs)
		Logga(ctx, "AMB OK")
	} else {
		Logga(ctx, "AMB MISSING")
	}
	// os.Exit(0)
	Logga(ctx, "")
	/* ************************************************************************************************ */
	// DEPLOYLOG
	Logga(ctx, "Getting DEPLOYLOG")

	var istanzaMicroVersioni []IstanzaMicroVersioni

	argsDeploy := make(map[string]string)
	argsDeploy["source"] = "devops-8"
	argsDeploy["$select"] = "XDEPLOYLOG03,XDEPLOYLOG05"
	argsDeploy["center_dett"] = "visualizza"
	argsDeploy["$filter"] = "equals(XDEPLOYLOG04,'" + istanza + "') "
	argsDeploy["$filter"] += " and equals(XDEPLOYLOG06,'1') "

	if ims.SwMultiEnvironment == "1" {
		argsDeploy["$filter"] += " and equals(XDEPLOYLOG09,'" + enviro + "') "
	}

	restyDeployRes := ApiCallGET(ctx, false, argsDeploy, "msdevops", "/devops/DEPLOYLOG", devopsToken, "")
	if restyDeployRes.Errore < 0 {
		Logga(ctx, restyDeployRes.Log)
		LoggaErrore.Errore = restyDeployRes.Errore
		LoggaErrore.Log = restyDeployRes.Log
		return ims, LoggaErrore
	}

	if len(restyDeployRes.BodyArray) > 0 {
		for _, x := range restyDeployRes.BodyArray {
			var istanzaMicroVersione IstanzaMicroVersioni

			istanzaMicroVersione.TipoVersione = x["XDEPLOYLOG03"].(string)
			istanzaMicroVersione.Versione = x["XDEPLOYLOG05"].(string)

			istanzaMicroVersioni = append(istanzaMicroVersioni, istanzaMicroVersione)
		}
		Logga(ctx, "DEPLOYLOG OK")
	} else {
		Logga(ctx, "DEPLOYLOG MISSING")
	}
	Logga(ctx, "")
	/* ************************************************************************************************ */

	ims.IstanzaMicroVersioni = istanzaMicroVersioni

	//Logga(ctx, ims)
	Logga(ctx, "getIstanceDetail end")
	Logga(ctx, " - - - - - - - - - - - - - - - - - - - ")
	Logga(ctx, "")
	//os.Exit(0)
	return ims, LoggaErrore
}
func UpdateIstanzaMicroservice(ctx context.Context, canaryProduction, versioneMicroservizio string, istanza IstanzaMicro, micros Microservice, utente, enviro, devopsToken string) LoggaErrore {

	var LoggaErrore LoggaErrore
	LoggaErrore.Errore = 0

	Logga(ctx, "")
	Logga(ctx, " + + + + + + + + + + + + + + + + + + + + ")
	Logga(ctx, "updateIstanzaMicroservice begin")
	Logga(ctx, "versioneMicroservizio "+versioneMicroservizio)
	for _, ccc := range istanza.IstanzaMicroVersioni {
		Logga(ctx, ccc.TipoVersione+" "+ccc.Versione)
	}

	//var clusterContext = "gke_" + istanza.ProjectID + "_europe-west1-d_" + istanza.Cluster

	// logica:
	// se canary devo rendere obsoleto il vecchio canarino  se esiste e inserire il nuovo canarino
	// se production devo rendere obsoleto la vecchia produzione e rendere il canarino produzione

	for _, versioni := range istanza.IstanzaMicroVersioni {

		switch canaryProduction {
		case "canary", "Canary":
			if versioni.TipoVersione == "canary" || versioni.TipoVersione == "Canary" {

				Logga(ctx, "Old canary found")

				Logga(ctx, "Delete canary "+istanza.Istanza+"-v"+versioni.Versione)
				Logga(ctx, "Make obsolete canary "+istanza.Istanza+" to version "+versioni.Versione)
				Logga(ctx, "New canary "+istanza.Istanza+" to version "+versioneMicroservizio)

				// rendo obsoleto il vecchio canarino
				keyvalueslice := make(map[string]interface{})
				keyvalueslice["debug"] = false
				keyvalueslice["source"] = "devops-8"
				keyvalueslice["XDEPLOYLOG06"] = "0"

				filter := "equals(XDEPLOYLOG04,'" + istanza.Istanza + "') and equals(XDEPLOYLOG03,'canary') and XDEPLOYLOG06 eq 1"

				_, erro := ApiCallPUT(ctx, false, keyvalueslice, "msdevops", "/devops/DEPLOYLOG/"+filter, devopsToken, "")

				if erro.Errore < 0 {
					return erro
				}
			}

			break

		case "production", "Production":

			Logga(ctx, "Delete production "+istanza.Istanza+"-v"+versioni.Versione)
			Logga(ctx, "Make obsolete production "+istanza.Istanza+" to version "+versioni.Versione)
			Logga(ctx, "Make canary the new production "+istanza.Istanza)

			// FAC-744 - rendere tutte le precedenti versioni obsolete XDEPLOYLOG07 = 1

			switch versioni.TipoVersione {
			case "production", "Production":

				// rendo obsoleto il vecchio production
				keyvalueslice := make(map[string]interface{})
				keyvalueslice["debug"] = false
				keyvalueslice["source"] = "devops-8"
				keyvalueslice["XDEPLOYLOG06"] = "0"

				filter := "equals(XDEPLOYLOG04,'" + istanza.Istanza + "') and equals(XDEPLOYLOG03,'production') and XDEPLOYLOG06 eq 1"

				_, erro := ApiCallPUT(ctx, false, keyvalueslice, "msdevops", "/devops/DEPLOYLOG/"+filter, devopsToken, "")
				if erro.Errore < 0 {
					return erro
				}

			case "canary", "Canary":

				// rendo obsoleto il canarino
				keyvalueslice := make(map[string]interface{})
				keyvalueslice["debug"] = false
				keyvalueslice["source"] = "devops-8"
				keyvalueslice["XDEPLOYLOG06"] = "0"

				filter := "equals(XDEPLOYLOG04,'" + istanza.Istanza + "') and equals(XDEPLOYLOG03,'canary')"

				_, erro := ApiCallPUT(ctx, false, keyvalueslice, "msdevops", "/devops/DEPLOYLOG/"+filter, devopsToken, "")
				if erro.Errore < 0 {
					return erro
				}

				break
			}

			break
		}

	}

	Logga(ctx, "Inserisco il record")

	canaryProduction = strings.ToLower(canaryProduction)

	/* ----------------- */

	type DeployDetail struct {
		Docker       string `json:"docker"`
		Versione     string `json:"versione"`
		Merged       string `json:"merged"`
		Tag          string `json:"tag"`
		MasterDev    string `json:"masterDev"`
		ReleaseNote  string `json:"releaseNote"`
		SprintBranch string `json:"sprintBranch"`
		Sha          string `json:"sha"`
	}

	var details []DeployDetail
	var detail DeployDetail

	for _, x := range micros.Pod {
		detail.Sha = x.Branch.Sha
		detail.Docker = x.Docker
		detail.Versione = x.PodBuild.Versione
		detail.Merged = x.PodBuild.Merged
		detail.Tag = x.PodBuild.Tag
		detail.MasterDev = x.PodBuild.MasterDev
		detail.ReleaseNote = x.PodBuild.ReleaseNote
		detail.SprintBranch = x.PodBuild.SprintBranch
		details = append(details, detail)
	}

	detailJson, err := json.Marshal(details)
	if err != nil {
		Logga(ctx, err.Error())
	}

	/* -------------------------- */
	// inserisco il nuovo canarino

	keyvalueslices := make([]map[string]interface{}, 0)
	keyvalueslice := make(map[string]interface{})
	keyvalueslice["debug"] = false
	keyvalueslice["source"] = "devops-8"
	keyvalueslice["XDEPLOYLOG03"] = canaryProduction
	keyvalueslice["XDEPLOYLOG04"] = istanza.Istanza
	keyvalueslice["XDEPLOYLOG05"] = versioneMicroservizio
	keyvalueslice["XDEPLOYLOG06"] = 1
	keyvalueslice["XDEPLOYLOG07"] = 0
	keyvalueslice["XDEPLOYLOG08"] = string(detailJson)
	keyvalueslice["XDEPLOYLOG09"] = enviro
	keyvalueslices = append(keyvalueslices, keyvalueslice)

	resPOST := ApiCallPOST(ctx, false, keyvalueslices, "msdevops", "/deploy/DEPLOYLOG", devopsToken, "")
	if resPOST.Errore < 0 {
		LoggaErrore.Log = resPOST.Log
		return LoggaErrore
	}

	Logga(ctx, "updateIstanzaMicroservice end")
	Logga(ctx, " - - - - - - - - - - - - - - - - - - - ")

	//os.Exit(0)
	return LoggaErrore
}
func CloudBuils(ctx context.Context, docker, verPad, dirRepo string, swMonolith bool) string {

	Logga(ctx, "")
	Logga(ctx, "CLOUD BUILD for "+docker)
	Logga(ctx, "")

	dir := ""
	dockerName := ""
	if swMonolith == true {
		dockerName = docker + "-monolith"
		dir = dirRepo
	} else {
		dir = dirRepo + "/" + docker
		dockerName = docker
	}

	cloudNet := ""
	if os.Getenv("gcloudNetRegion") != "" {
		cloudNet = " --region " + os.Getenv("gcloudNetRegion")
	} else {
		cloudNet = " --zone " + os.Getenv("gcloudNetZone")
	}
	SwitchProject(os.Getenv("gcloudProjectID"))
	SwitchCluster(os.Getenv("clusterKube8"), cloudNet)

	fileCloudBuild := dir + "/cloudBuild.yaml"

	fmt.Println(fileCloudBuild)

	cloudBuild := "steps:\n"
	cloudBuild += "- name: 'gcr.io/cloud-builders/docker'\n"
	cloudBuild += "  args: ['build', "
	if swMonolith == true {
		cloudBuild += "'--build-arg', 'CLIENTE=" + docker + "', "
	}
	cloudBuild += "'-t', '" + os.Getenv("gcloudUrl") + "/" + os.Getenv("gcloudProjectID") + "/" + dockerName + ":" + verPad + "', '.']\n"
	cloudBuild += "- name: 'gcr.io/cloud-builders/docker'\n"
	cloudBuild += "  args: ['push', '" + os.Getenv("gcloudUrl") + "/" + os.Getenv("gcloudProjectID") + "/" + dockerName + ":" + verPad + "']\n"
	cloudBuild += "images: ['" + os.Getenv("gcloudUrl") + "/" + os.Getenv("gcloudProjectID") + "/" + dockerName + ":" + verPad + "']\n"
	cloudBuild += "tags:\n"
	cloudBuild += "- '" + dockerName + "-" + verPad + "'\n"
	cloudBuild += "options:\n"
	cloudBuild += "  machineType: 'E2_HIGHCPU_8'\n"
	//cloudBuild += "  logStreamingOption: 'STREAM_ON'\n"

	Logga(ctx, cloudBuild)

	f, errF := os.Create(fileCloudBuild)
	if errF != nil {
		Logga(ctx, errF.Error())
	}
	_, errF = f.WriteString(cloudBuild)
	if errF != nil {
		Logga(ctx, errF.Error())
		f.Close()
	}

	// RUN THE BUILD
	command := "gcloud builds submit --async --config " + fileCloudBuild + " " + dir
	Logga(ctx, command)
	ExecCommand(command, false)

	// SEEK THE BUILD ID
	command = "gcloud builds list --filter \"tags='" + dockerName + "-" + verPad + "'\" --format=\"json\""
	Logga(ctx, command)

	fmt.Println("_##START##_Build Started_##STOP##_")

	type logStruct struct {
		ID      string `json:"id"`
		LogUrl  string `json:"logUrl"`
		Status  string `json:"status"`
		Results struct {
			Images []struct {
				Digest string `json:"digest"`
			} `json:"images"`
		} `json:"results"`
	}

	sha256 := ""
	i := 0
	for {

		cmd := exec.Command("bash", "-c", command)
		out, err := cmd.CombinedOutput()
		if err != nil {
			Logga(ctx, "cmd.Run() failed with %s\n"+err.Error())
		}
		//s := strings.TrimSpace(string(out))
		//Logga(ctx, s)

		var logRes []logStruct
		errJson := json.Unmarshal(out, &logRes)
		if errJson != nil {
			Logga(ctx, errJson, "error")
		}
		// fmt.Println(len(logRes))
		// LogJson(logRes)

		if i == 0 {
			fmt.Println("_##START##_Build ID: " + logRes[0].ID + "_##STOP##_")
			fmt.Println("_##START##_Build LOG at : " + logRes[0].LogUrl + "_##STOP##_")

			JobID := ctx.Value("JobID").(string)
			telegramText := "JobID: " + JobID
			telegramText += " " + dockerName + " - " + logRes[0].LogUrl
			erroTelegram := TelegramSendMessage(os.Getenv("telegramBotToken"), os.Getenv("telegramCftoolDevopsChatID"), telegramText)

			if erroTelegram.Errore < 0 {

				Logga(ctx, "")
				Logga(ctx, "ERRORE")
				Logga(ctx, erroTelegram.Log)
			} else {
				Logga(ctx, "The job has been loaded successfully")
				Logga(ctx, "A telegram message has been sent to you")
			}
		}
		fmt.Println("_##START##_Build Status : " + logRes[0].Status + "_##STOP##_")

		if logRes[0].Status == "SUCCESS" {
			sha256 = logRes[0].Results.Images[0].Digest
			break
		}

		if logRes[0].Status == "FAILURE" ||
			logRes[0].Status == "CANCELLED" ||
			logRes[0].Status == "TIMEOUT" ||
			logRes[0].Status == "FAILED" {
			fmt.Println("_##START##_Build FAILED _##STOP##_")
			fmt.Println("_##START##_Build FAILED_##STOP##_")
			os.Exit(1)
		}

		time.Sleep(10 * time.Second)
		i++
	}

	fmt.Println("_##START##_Build Process Finished_##STOP##_")

	return sha256
}
func UpdateDockerVersion(ctx context.Context, docker, ver, user, devMaster, sha, team, newTagName, releaseNote, parentBranch, cs, merged string) {

	Logga(ctx, "Getting token")
	devopsToken, erro := GetCoreFactoryToken(ctx)
	if erro.Errore < 0 {
		Logga(ctx, erro.Log)
	} else {
		Logga(ctx, "Token OK")
	}

	/* ***************************************************** */
	Logga(ctx, "Insert TB_ANAG_KUBEDKRBUILD00 ")

	keyvalueslices := make([]map[string]interface{}, 0)
	keyvalueslice := make(map[string]interface{})
	keyvalueslice["debug"] = true
	keyvalueslice["source"] = "devops-8"
	keyvalueslice["XKUBEDKRBUILD03"] = docker
	keyvalueslice["XKUBEDKRBUILD04"] = devMaster
	keyvalueslice["XKUBEDKRBUILD05"] = user
	keyvalueslice["XKUBEDKRBUILD06"] = ver
	keyvalueslice["XKUBEDKRBUILD07"] = sha
	keyvalueslice["XKUBEDKRBUILD08"] = team
	keyvalueslice["XKUBEDKRBUILD09"] = newTagName
	keyvalueslice["XKUBEDKRBUILD10"] = parentBranch
	keyvalueslice["XKUBEDKRBUILD11"] = cs
	keyvalueslice["XKUBEDKRBUILD12"] = releaseNote
	keyvalueslice["XKUBEDKRBUILD13"] = merged

	keyvalueslices = append(keyvalueslices, keyvalueslice)

	res := ApiCallPOST(ctx, false, keyvalueslices, "msdevops", "/devops/KUBEDKRBUILD", devopsToken, "")
	if res.Errore != 0 {
		Logga(ctx, res.Log)
	}
}

/*
Fa il merge di un branch sull'altro
se ci sono conflitti li segnala

accetta branch source, branch dest, tipo (tag o branch)
ritorna un LOG
*/
func GitMergeApi(ctx context.Context, src, dst, repo, tipo string) (string, string) {

	Logga(ctx, "gitMergeApi")

	var mergeRes MergeResponse
	var erroMerge, noChanges bool
	erroMerge = false
	noChanges = false

	mergeRes.Log += "\n**********************************************************************\n"
	mergeRes.Log += "Work on repo " + repo + " -> " + src + " on " + dst + "\n\n"

	// se il tipo di merge è fra un tag e un branch
	// va prima fatto un branch partendo dal tag
	var tmpBranch string
	if tipo == "tag" {

		tmpBranch = src + "-tmp-branch"

		Logga(ctx, repo+": creo branch dal tag "+src)
		// creo un branch vivo dal tag

		body := `{"name": "` + tmpBranch + `","target": {  "hash": "` + src + `"}}`

		clientTag := resty.New()
		clientTag.Debug = false
		restyTagResponse, err := clientTag.R().
			SetHeader("Content-Type", "application/json").
			SetBasicAuth(os.Getenv("bitbucketUser"), os.Getenv("bitbucketToken")).
			SetBody(body).
			Post(os.Getenv("bitbucketHost") + "/repositories/" + os.Getenv("bitbucketProject") + "/" + repo + "/refs/branches")

		if err != nil {
			fmt.Println("_##START##_   !!! New branch on " + repo + " ERROR " + err.Error() + "_##STOP##_")

			mergeRes.Error += "Error: "
			erroMerge = true
			mergeRes.Error += err.Error()
			mergeRes.Error += "\n"
		}

		var restyRes CreateBranchResponse
		_ = json.Unmarshal(restyTagResponse.Body(), &restyRes)
		//fmt.Println(restyTagResponse, err)

		if restyRes.Type == "error" {
			fmt.Println("_##START##_   !!! New branch on " + repo + " ERROR " + restyRes.Error.Message + "_##STOP##_")
			if restyRes.Error.Data.Key != "BRANCH_ALREADY_EXISTS" {
				mergeRes.Error += "Error: "
				erroMerge = true
				mergeRes.Error += restyRes.Error.Message
				mergeRes.Error += "\n"
			}
		} else {
			fmt.Println("_##START##_New branch on " + repo + " created_##STOP##_")
		}
		// --------------------------------
	}

	// FACCIO LA PULL REQUEST PER IL MERGE
	Logga(ctx, repo+": faccio pull req di merge di "+src+" su "+dst)
	titolo := "Merge " + src + " on " + dst
	var body string
	if tipo == "tag" {
		body = `{"title": "` + titolo + `","source": {"branch": {"name": "` + tmpBranch + `"}},"destination": {"branch": {"name": "` + dst + `"}},"close_source_branch": true }`
	} else {
		body = `{"title": "` + titolo + `","source": {"branch": {"name": "` + src + `"}},"destination": {"branch": {"name": "` + dst + `"}} }`
	}

	clientPullR := resty.New()
	clientPullR.Debug = false
	restyPullReqResponse, err := clientPullR.R().
		SetHeader("Content-Type", "application/json").
		SetBasicAuth(os.Getenv("bitbucketUser"), os.Getenv("bitbucketToken")).
		SetBody(body).
		Post(os.Getenv("bitbucketHost") + "/repositories/" + os.Getenv("bitbucketProject") + "/" + repo + "/pullrequests")

	if err != nil {
		fmt.Println("_##START##_   !!! Merge di " + src + " su " + dst + " ERROR " + err.Error() + "_##STOP##_")

		mergeRes.Error += err.Error()
		mergeRes.Error += "\n"
		fmt.Println(err)
	}

	var restyRes RestyResStruct
	_ = json.Unmarshal(restyPullReqResponse.Body(), &restyRes)
	//fmt.Println(restyRes, err)
	// os.Exit(0)

	if restyRes.Error.Message != "" {
		if restyRes.Error.Message == "There are no changes to be pulled" {
			noChanges = true
		} else {
			mergeRes.Error += "Error: "
			erroMerge = true
			mergeRes.Error += restyRes.Error.Message
			mergeRes.Error += "\n"
		}
	}
	// ----------------------------

	//fmt.Println("@@@", noChanges, erroMerge)
	//os.Exit(0)

	if !noChanges {
		if !erroMerge {

			// NON HO ERRORI E QUINDI FACCIO IL MERGE
			Logga(ctx, repo+": faccio Merge di "+src+" su "+dst)
			mergeRes.Log += "Do MERGE of " + src + " on " + dst + "\n"

			clientMerge := resty.New()
			clientMerge.Debug = false
			respMerge, errMerge := clientMerge.R().
				SetBasicAuth(os.Getenv("bitbucketUser"), os.Getenv("bitbucketToken")).
				Post(os.Getenv("bitbucketHost") + "/repositories/" + os.Getenv("bitbucketProject") + "/" + repo + "/pullrequests/" + strconv.Itoa(restyRes.ID) + "/merge")
			// fmt.Println(string(respMerge.Body()), errMerge)

			if errMerge != nil {
				mergeRes.Error += "Error: "
				mergeRes.Error += errMerge.Error()
				mergeRes.Error += "\n"
			}

			var restyResMerge RestyResStruct
			_ = json.Unmarshal(respMerge.Body(), &restyResMerge)
			//fmt.Println(restyResMerge, err)

			// HO DEGLI ERRORI NEL MERGE
			if restyResMerge.Error.Message != "" {

				// MI CERCO IL DIFF DEI CONFLITTI
				clientConflict := resty.New()
				clientConflict.Debug = false
				respConflict, errConflict := clientConflict.R().
					EnableTrace().
					SetBasicAuth(os.Getenv("bitbucketUser"), os.Getenv("bitbucketToken")).
					Get(restyRes.Links.Diff.Href)

				if errConflict != nil {
				}

				mergeRes.Error += "\nError: "
				mergeRes.Error += repo + "\n"
				mergeRes.Error += "Pull Request ID #" + strconv.Itoa(restyRes.ID) + " \n"
				mergeRes.Error += "------------------------------\n"
				mergeRes.Error += string(respConflict.Body())
				mergeRes.Error += "------------------------------\n"
				mergeRes.Error += "\n"
				mergeRes.Error += restyResMerge.Error.Message
				mergeRes.Error += "\n"

			} else {

				// MERGE OK
				mergeRes.Log += repo + ": Merge of " + src + " on " + dst + " OK\n"
			}

		}
	} else {
		mergeRes.Log += repo + ": There are no changes to be merged\n"
	}

	//fmt.Println("------------------------------------------------|" + mergeRes.Error + "|")
	if mergeRes.Error != "" {
		mergeRes.Log += mergeRes.Error
		mergeRes.Log += "\n"
	}

	mergeRes.Log += "\n**********************************************************************\n"

	return mergeRes.Log, mergeRes.Error
}
