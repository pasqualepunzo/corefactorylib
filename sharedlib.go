package lib

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/go-resty/resty/v2"
)

func GetIstanceDetail(ctx context.Context, iresReq IresRequest, canaryProduction, dominio, coreApiVersion string) (IstanzaMicro, error) {

	Logga(ctx, ctx.Value("JsonLog").(bool), "")
	Logga(ctx, ctx.Value("JsonLog").(bool), " + + + + + + + + + + + + + + + + + + + +")
	Logga(ctx, ctx.Value("JsonLog").(bool), "getIstanceDetail begin")
	restyDebug := false
	if os.Getenv("restyDebug") == "true" {
		restyDebug = true
	}

	var LoggaErrore LoggaErrore
	LoggaErrore.Errore = 0

	var erro error

	devopsToken := iresReq.TokenSrc
	devopsTokenDst := iresReq.TokenDst

	if devopsTokenDst == "" {
		devopsTokenDst = devopsToken
	}

	istanza := iresReq.Istanza
	//istanzaDst := iresReq.IstanzaDst
	microservice := iresReq.Microservice
	enviro := iresReq.Enviro
	refAppID := iresReq.RefAppID
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

	devops := "devops"
	auth := "auth"
	if ims.Monolith == 1 {
		devops = "devopsmono"
		auth = "adpauth"
	}

	// qui in data 19 maggio 2021
	// con davide e mauro si decide che le connessione MASTER vanno
	// definite su un config
	// devopsProfile, _ := os.LookupEnv("APP_ENV")
	//if devopsProfile == "prod" {
	//ims.MasterHost = os.Getenv("hostData").

	// } else {
	// ims.MasterHost = os.Getenv("hostDataDev")
	// }
	// ims.MasterName = os.Getenv("nameData")
	// ims.MasterUser = os.Getenv("userData")
	// ims.MasterPass = os.Getenv("passData")

	/* ************************************************************************************************ */
	// KUBEIMICROSERV

	if !iresReq.SwDest { // se stiamo in migrazione non server fare questa chiamata perche nella destinazione non esiste e non deve esistere
		Logga(ctx, ctx.Value("JsonLog").(bool), "Getting KUBEIMICROSERV - deploy.go")
		argsImicro := make(map[string]string)
		argsImicro["source"] = "devops-8"
		argsImicro["$select"] = "XKUBEIMICROSERV04,XKUBEIMICROSERV05"
		argsImicro["center_dett"] = "dettaglio"
		argsImicro["$filter"] = "equals(XKUBEIMICROSERV03,'" + istanza + "') "

		restyKubeImicroservRes, errKubeImicroservRes := ApiCallGET(ctx, restyDebug, argsImicro, "ms"+devops, "/"+devops+"/KUBEIMICROSERV", devopsToken, dominio, coreApiVersion)
		if errKubeImicroservRes != nil {
			Logga(ctx, ctx.Value("JsonLog").(bool), errKubeImicroservRes.Error())
			erro = errors.New(errKubeImicroservRes.Error())
			return ims, erro
		}

		if len(restyKubeImicroservRes.BodyJson) > 0 {
			microservice = restyKubeImicroservRes.BodyJson["XKUBEIMICROSERV04_COD"].(string)
			ims.Cluster = restyKubeImicroservRes.BodyJson["XKUBEIMICROSERV05"].(string)
			ims.Microservice = microservice

			Logga(ctx, ctx.Value("JsonLog").(bool), "KUBEIMICROSERV OK")
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
				keyvalueslice["XKUBEIMICROSERV03"] = strings.ToLower(istanza)
				keyvalueslice["XKUBEIMICROSERV04"] = _ms
				keyvalueslice["XKUBEIMICROSERV05"] = _cluster
				keyvalueslice["XKUBEIMICROSERV06"] = enviro

				keyvalueslices = append(keyvalueslices, keyvalueslice)

				resKubeims := ApiCallPOST(ctx, restyDebug, keyvalueslices, "ms"+devops, "/"+devops+"/KUBEIMICROSERV", devopsToken, dominio, coreApiVersion)

				if resKubeims.Errore < 0 {
					Logga(ctx, ctx.Value("JsonLog").(bool), "")
					Logga(ctx, ctx.Value("JsonLog").(bool), "NON RIESCO A SCRIVRERE L'ISTANZA "+resKubeims.Log)
					Logga(ctx, ctx.Value("JsonLog").(bool), "")

					erro = errors.New(resKubeims.Log)
					return ims, erro
				}

			} else {

				Logga(ctx, ctx.Value("JsonLog").(bool), "KUBEIMICROSERV MISSING")
			}
		}
		Logga(ctx, ctx.Value("JsonLog").(bool), "")
	} else {

		Logga(ctx, ctx.Value("JsonLog").(bool), "CLUSTER PRESO DA iresReq "+iresReq.ClusterDst)
		// se siamo in migrazione non applichiamo questo metodo ma abbiamo necessita di avere il cluster valorizzato
		ims.Cluster = iresReq.ClusterDst
	}

	/* ************************************************************************************************ */
	// KUBESTAGE
	Logga(ctx, ctx.Value("JsonLog").(bool), "Getting KUBESTAGE sharedlib")

	argsStage := make(map[string]string)
	argsStage["source"] = "devops-8"
	argsStage["$select"] = "XKUBESTAGE08,XKUBESTAGE09"
	argsStage["center_dett"] = "dettaglio"
	argsStage["$filter"] = "equals(XKUBESTAGE03,'" + ims.Cluster + "') "
	argsStage["$filter"] += " and equals(XKUBESTAGE04,'" + enviro + "') "

	//$filter=contains(XART20,'(kg)') or contains(XART20,'pizza')
	restyStageRes, _ := ApiCallGET(ctx, restyDebug, argsStage, "ms"+devops, "/"+devops+"/KUBESTAGE", devopsTokenDst, dominio, coreApiVersion)
	if restyStageRes.Errore < 0 {
		Logga(ctx, ctx.Value("JsonLog").(bool), restyStageRes.Log)
	}

	var swProdStage int
	var depEnv string
	if len(restyStageRes.BodyJson) > 0 {
		depEnv, _ = restyStageRes.BodyJson["XKUBESTAGE09"].(string)
		swProdStage = int(restyStageRes.BodyJson["XKUBESTAGE08"].(float64))
		Logga(ctx, ctx.Value("JsonLog").(bool), "KUBESTAGE: OK")
	} else {
		Logga(ctx, ctx.Value("JsonLog").(bool), "KUBESTAGE: MISSING")
	}

	/* ************************************************************************************************ */
	// KUBECLUSTER

	// il 21 04 2023 mepo laszlo e frnc non si addoneno del motivo di una array bidimens
	// e a puorc schiattano un filtro sul cluster lasciando tutto invariato
	// ma ovviamente la matrice avra una sola KEY
	Logga(ctx, ctx.Value("JsonLog").(bool), "Getting KUBECLUSTER")

	argsClu := make(map[string]string)
	argsClu["source"] = "devops-8"
	argsClu["center_dett"] = "allviews"
	argsClu["$filter"] = " equals(XKUBECLUSTER03,'" + ims.Cluster + "')"

	restyKubeCluRes, errKubeCluRes := ApiCallGET(ctx, restyDebug, argsClu, "ms"+devops, "/"+devops+"/KUBECLUSTER", devopsTokenDst, dominio, coreApiVersion)
	if errKubeCluRes != nil {
		Logga(ctx, ctx.Value("JsonLog").(bool), errKubeCluRes.Error())

		erro = errors.New(errKubeCluRes.Error())
		return ims, erro
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
			clu.DepEnv = depEnv

			clu.Domain = x["XKUBECLUSTER15"].(string)
			if swProdStage == 0 && x["XKUBECLUSTER17"].(string) != "" {
				clu.Domain = x["XKUBECLUSTER17"].(string)
			}
			// a prescindere mi porto tutti i domini
			clu.DomainProd = x["XKUBECLUSTER15"].(string)
			clu.DomainStage = x["XKUBECLUSTER17"].(string)

			clu.Token = x["XKUBECLUSTER20"].(string)

			// TOLGO QUESTE PERCHE PRENDO DA AMBDOMAIN 2023 05 02
			// clu.MasterHost = x["XKUBECLUSTER09"].(string)
			// clu.MasterUser = x["XKUBECLUSTER10"].(string)
			// clu.MasterPasswd = x["XKUBECLUSTER11"].(string)
			clu.AccessToken = x["XKUBECLUSTER20"].(string)

			ambienteFloat := x["XKUBECLUSTER12"].(float64)
			clu.Ambiente = int32(ambienteFloat)
			clu.RefappID = x["XKUBECLUSTER22"].(string)

			clu.CloudNet = x["XKUBECLUSTER24"].(string)

			clu.Autopilot = strconv.FormatFloat(x["XKUBECLUSTER14"].(float64), 'f', 0, 64)
			clu.DomainOvr = false

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

			restyKubeCluEnvRes, errKubeCluEnvRes := ApiCallGET(ctx, restyDebug, argsCluEnv, "ms"+devops, "/"+devops+"/KUBECLUSTERENV", devopsToken, dominio, coreApiVersion)
			if errKubeCluEnvRes != nil {
				Logga(ctx, ctx.Value("JsonLog").(bool), errKubeCluEnvRes.Error())

				erro = errors.New(errKubeCluEnvRes.Error())
				return ims, erro
			}

			if len(restyKubeCluEnvRes.BodyJson) > 0 {
				domainCluEnv, _ := restyKubeCluEnvRes.BodyJson["XKUBECLUSTERENV08"].(string)
				if domainCluEnv != "" {
					clu.DomainEnv = restyKubeCluEnvRes.BodyJson["XKUBECLUSTERENV08"].(string)
					clu.DomainOvr = true
				}
				refAppIDCluEnv, _ := restyKubeCluEnvRes.BodyJson["XKUBECLUSTERENV09"].(string)
				if refAppIDCluEnv != "" {
					clu.RefappID = restyKubeCluEnvRes.BodyJson["XKUBECLUSTERENV09"].(string)
				}

				// TOLGO QUESTE PERCHE PRENDO DA AMBDOMAIN 2023 05 02
				// clu.MasterHost = restyKubeCluEnvRes.BodyJson["XKUBECLUSTERENV07"].(string)

				Logga(ctx, ctx.Value("JsonLog").(bool), "KUBECLUSTERENV OK")
			}

			clus[x["XKUBECLUSTER03"].(string)] = clu

			// Logga(ctx, ctx.Value("JsonLog").(bool), x["XKUBECLUSTER03"].(string))
			// LogJson(clu)

		}

		//ims.Clusters = clus

		Logga(ctx, ctx.Value("JsonLog").(bool), "KUBECLUSTER OK")
	} else {
		Logga(ctx, ctx.Value("JsonLog").(bool), "KUBECLUSTER MISSING")
	}
	Logga(ctx, ctx.Value("JsonLog").(bool), "")

	/* ************************************************************************************************ */

	/* ************************************************************************************************ */
	// KUBECLUSTER
	Logga(ctx, ctx.Value("JsonLog").(bool), "Getting KUBECLUSTER")

	if _, ok := clus[ims.Cluster]; ok {

		Logga(ctx, ctx.Value("JsonLog").(bool), " +++ KUBECLUSTER OK +++ ")

		// fmt.Println(ims.cluster)
		// LogJson(clus[ims.cluster])

		ims.ProjectID = clus[ims.Cluster].ProjectID
		ims.Owner = clus[ims.Cluster].Owner
		ims.Profile = clus[ims.Cluster].Profile
		ims.ProfileInt = clus[ims.Cluster].ProfileInt
		ims.ClusterDomain = clus[ims.Cluster].Domain
		ims.ClusterDomainOvr = clus[ims.Cluster].DomainOvr
		ims.ClusterDomainProd = clus[ims.Cluster].DomainProd
		ims.ClusterDomainStage = clus[ims.Cluster].DomainStage
		ims.ClusterDomainEnv = clus[ims.Cluster].DomainEnv
		ims.Token = clus[ims.Cluster].Token
		ims.MasterUser = clus[ims.Cluster].MasterUser
		ims.MasterPass = clus[ims.Cluster].MasterPasswd
		ims.Ambiente = clus[ims.Cluster].Ambiente
		ims.ClusterRefAppID = clus[ims.Cluster].RefappID
		ims.ApiHost = clus[ims.Cluster].ApiHost
		ims.ApiToken = clus[ims.Cluster].ApiToken
		ims.Autopilot = clus[ims.Cluster].Autopilot
		ims.CloudNet = clus[ims.Cluster].CloudNet
		ims.DepEnv = clus[ims.Cluster].DepEnv

		Logga(ctx, ctx.Value("JsonLog").(bool), "KUBECLUSTER OK")
	} else {
		Logga(ctx, ctx.Value("JsonLog").(bool), "KUBECLUSTER MISSING")
	}
	Logga(ctx, ctx.Value("JsonLog").(bool), "")

	/* ************************************************************************************************ */
	// AMBDOMAIN

	Logga(ctx, ctx.Value("JsonLog").(bool), "Getting AMBDOMAIN")

	argsAmbdomain := make(map[string]string)
	argsAmbdomain["source"] = "auth-1"
	argsAmbdomain["$select"] = "XAMBDOMAIN05,XAMBDOMAIN07,XAMBDOMAIN08,XAMBDOMAIN09,XAMBDOMAIN10,XAMBDOMAIN11"
	argsAmbdomain["center_dett"] = "dettaglio"
	argsAmbdomain["$filter"] += "  equals(XAMBDOMAIN04,'" + customerDomain + "') "
	restyAmbdomainRes, errAmbdomainRes := ApiCallGET(ctx, restyDebug, argsAmbdomain, "msauth", "/core/AMBDOMAIN", devopsToken, dominio, coreApiVersion)
	if errAmbdomainRes != nil {
		Logga(ctx, ctx.Value("JsonLog").(bool), errAmbdomainRes.Error())

		erro = errors.New(errAmbdomainRes.Error())
		return ims, erro
	}

	if len(restyAmbdomainRes.BodyJson) > 0 {
		ims.CustomerSalt = restyAmbdomainRes.BodyJson["XAMBDOMAIN11"].(string)

		// AGGIUNGO QUESTE PERCHE PRENDO DA AMBDOMAIN 2023 05 02
		ims.MasterHostData = restyAmbdomainRes.BodyJson["XAMBDOMAIN07"].(string)
		ims.MasterHostMeta = restyAmbdomainRes.BodyJson["XAMBDOMAIN07"].(string)
		ims.MasterUser = restyAmbdomainRes.BodyJson["XAMBDOMAIN09"].(string)
		ims.MasterPass = restyAmbdomainRes.BodyJson["XAMBDOMAIN10"].(string)
		Logga(ctx, ctx.Value("JsonLog").(bool), "AMBDOMAIN OK")
	} else {
		Logga(ctx, ctx.Value("JsonLog").(bool), "AMBDOMAIN MISSING")
	}
	Logga(ctx, ctx.Value("JsonLog").(bool), "")
	/* ************************************************************************************************ */
	// KUBEMICROSERV
	Logga(ctx, ctx.Value("JsonLog").(bool), "Getting KUBEMICROSERV")

	argsMS := make(map[string]string)
	argsMS["source"] = "devops-8"
	argsMS["$select"] = "XKUBEMICROSERV09,XKUBEMICROSERV15,XKUBEMICROSERV18"
	argsMS["center_dett"] = "dettaglio"
	argsMS["$filter"] = "equals(XKUBEMICROSERV05,'" + microservice + "') "
	restyKubeMSRes, errKubeMSRes := ApiCallGET(ctx, restyDebug, argsMS, "ms"+devops, "/"+devops+"/KUBEMICROSERV", devopsToken, dominio, coreApiVersion)
	if errKubeMSRes != nil {
		Logga(ctx, ctx.Value("JsonLog").(bool), errKubeMSRes.Error())

		erro = errors.New(errKubeMSRes.Error())
		return ims, erro
	}

	microservicePublic := 0
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
		microservicePublic = int(restyKubeMSRes.BodyJson["XKUBEMICROSERV18"].(float64))

		ims.SwDb = swDb
		Logga(ctx, ctx.Value("JsonLog").(bool), "KUBEMICROSERV OK")
	} else {
		Logga(ctx, ctx.Value("JsonLog").(bool), "KUBEMICROSERV MISSING")
	}
	Logga(ctx, ctx.Value("JsonLog").(bool), "")

	/* ************************************************************************************************ */
	// AMB
	Logga(ctx, ctx.Value("JsonLog").(bool), "Getting AMB")

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
	argsAmb["public"] = strconv.Itoa(microservicePublic)
	//argsAmb["swMultiEnvironment"] = ims.SwMultiEnvironment

	restyKubeAmbRes, errKubeAmbRes := ApiCallGET(ctx, restyDebug, argsAmb, "msauth", "/"+auth+"/getAmbDomainMs", devopsTokenDst, dominio, coreApiVersion)
	if errKubeAmbRes != nil {
		Logga(ctx, ctx.Value("JsonLog").(bool), errKubeAmbRes.Error())

		erro = errors.New(errKubeAmbRes.Error())
		return ims, erro
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

			// TOLGO QUESTE PERCHE PRENDO DA AMBDOMAIN 2023 05 02
			//ims.MasterUser = clus[x["CLUSTER"].(string)].MasterUser
			//ims.MasterPass = clus[x["CLUSTER"].(string)].MasterPasswd

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
		Logga(ctx, ctx.Value("JsonLog").(bool), "AMB OK")
	} else {
		Logga(ctx, ctx.Value("JsonLog").(bool), "AMB MISSING")
	}
	// os.Exit(0)
	Logga(ctx, ctx.Value("JsonLog").(bool), "")

	/* ************************************************************************************************ */
	// DEPLOYLOG
	var erroIstanzaVersioni error
	ims.IstanzaMicroVersioni, erroIstanzaVersioni = GetIstanzaVersioni(ctx, iresReq, istanza, enviro, devops, devopsTokenDst, dominio, coreApiVersion)
	if erroIstanzaVersioni != nil {
		Logga(ctx, ctx.Value("JsonLog").(bool), erroIstanzaVersioni.Error())
		return ims, erroIstanzaVersioni
	}
	// DEPLOYLOG
	/* ************************************************************************************************ */

	//Logga(ctx, ctx.Value("JsonLog").(bool), ims)
	Logga(ctx, ctx.Value("JsonLog").(bool), "", "getIstanceDetail end")
	Logga(ctx, ctx.Value("JsonLog").(bool), "", " - - - - - - - - - - - - - - - - - - - ")
	Logga(ctx, ctx.Value("JsonLog").(bool), "", "")
	//os.Exit(0)
	return ims, erro
}
func GetIstanzaVersioni(ctx context.Context, iresReq IresRequest, istanza, enviro, devops, devopsTokenDst, dominio, coreApiVersion string) ([]IstanzaMicroVersioni, error) {
	Logga(ctx, ctx.Value("JsonLog").(bool), "Getting DEPLOYLOG")
	var erro error
	var istanzaMicroVersioni []IstanzaMicroVersioni

	restyDebug := false
	if os.Getenv("restyDebug") == "true" {
		restyDebug = true
	}

	argsDeploy := make(map[string]string)
	argsDeploy["source"] = "devops-8"
	argsDeploy["$select"] = "XDEPLOYLOG03,XDEPLOYLOG05"
	argsDeploy["center_dett"] = "visualizza"
	if iresReq.SwDest { // MIGRAZIONE MS
		argsDeploy["$filter"] = "equals(XDEPLOYLOG04,'" + iresReq.IstanzaDst + "') "
	} else {
		argsDeploy["$filter"] = "equals(XDEPLOYLOG04,'" + istanza + "') "
	}
	argsDeploy["$filter"] += " and equals(XDEPLOYLOG09,'" + enviro + "') "
	argsDeploy["$filter"] += " and equals(XDEPLOYLOG06,'1') "

	restyDeployRes, errDeployRes := ApiCallGET(ctx, restyDebug, argsDeploy, "ms"+devops, "/"+devops+"/DEPLOYLOG", devopsTokenDst, dominio, coreApiVersion)
	if errDeployRes != nil {
		Logga(ctx, ctx.Value("JsonLog").(bool), errDeployRes.Error())
		erro = errors.New(errDeployRes.Error())
		return istanzaMicroVersioni, erro
	}

	if len(restyDeployRes.BodyArray) > 0 {
		for _, x := range restyDeployRes.BodyArray {
			var istanzaMicroVersione IstanzaMicroVersioni

			istanzaMicroVersione.TipoVersione = x["XDEPLOYLOG03"].(string)
			istanzaMicroVersione.Versione = x["XDEPLOYLOG05"].(string)

			istanzaMicroVersioni = append(istanzaMicroVersioni, istanzaMicroVersione)
		}
		Logga(ctx, ctx.Value("JsonLog").(bool), "DEPLOYLOG OK")
	} else {
		Logga(ctx, ctx.Value("JsonLog").(bool), "DEPLOYLOG MISSING")
	}
	Logga(ctx, ctx.Value("JsonLog").(bool), "")

	return istanzaMicroVersioni, erro
}
func UpdateIstanzaMicroservice(ctx context.Context, canaryProduction, versioneMicroservizio string, istanza IstanzaMicro, micros Microservice, utente, enviro, devopsToken, dominio, coreApiVersion, microfrontendJson string) LoggaErrore {

	var LoggaErrore LoggaErrore
	LoggaErrore.Errore = 0

	restyDebug := false
	if os.Getenv("restyDebug") == "true" {
		restyDebug = true
	}

	Logga(ctx, ctx.Value("JsonLog").(bool), "")
	Logga(ctx, ctx.Value("JsonLog").(bool), " + + + + + + + + + + + + + + + + + + + + ")
	Logga(ctx, ctx.Value("JsonLog").(bool), "updateIstanzaMicroservice begin")
	Logga(ctx, ctx.Value("JsonLog").(bool), "versioneMicroservizio "+versioneMicroservizio)
	for _, ccc := range istanza.IstanzaMicroVersioni {
		Logga(ctx, ctx.Value("JsonLog").(bool), ccc.TipoVersione+" "+ccc.Versione)
	}

	devops := "devops"
	if istanza.Monolith == 1 {
		devops = "devopsmono"
	}

	//var clusterContext = "gke_" + istanza.ProjectID + "_europe-west1-d_" + istanza.Cluster

	// logica:
	// se canary devo rendere obsoleto il vecchio canarino  se esiste e inserire il nuovo canarino
	// se production devo rendere obsoleto la vecchia produzione e rendere il canarino produzione

	for _, versioni := range istanza.IstanzaMicroVersioni {

		switch canaryProduction {
		case "canary", "Canary":
			if versioni.TipoVersione == "canary" || versioni.TipoVersione == "Canary" {

				Logga(ctx, ctx.Value("JsonLog").(bool), "Old canary found")

				Logga(ctx, ctx.Value("JsonLog").(bool), "Delete canary "+istanza.Istanza+"-v"+versioni.Versione)
				Logga(ctx, ctx.Value("JsonLog").(bool), "Make obsolete canary "+istanza.Istanza+" to version "+versioni.Versione)
				Logga(ctx, ctx.Value("JsonLog").(bool), "New canary "+istanza.Istanza+" to version "+versioneMicroservizio)

				// rendo obsoleto il vecchio canarino
				keyvalueslice := make(map[string]interface{})
				keyvalueslice["debug"] = false
				keyvalueslice["source"] = "devops-8"
				keyvalueslice["XDEPLOYLOG06"] = "0"

				filter := "equals(XDEPLOYLOG04,'" + istanza.Istanza + "') and equals(XDEPLOYLOG03,'canary') and XDEPLOYLOG06 eq 1"

				_, erro := ApiCallPUT(ctx, restyDebug, keyvalueslice, "ms"+devops, "/"+devops+"/DEPLOYLOG/"+filter, devopsToken, dominio, coreApiVersion)

				if erro.Errore < 0 {
					return erro
				}
			}

			break

		case "production", "Production":

			Logga(ctx, ctx.Value("JsonLog").(bool), "Delete production "+istanza.Istanza+"-v"+versioni.Versione)
			Logga(ctx, ctx.Value("JsonLog").(bool), "Make obsolete production "+istanza.Istanza+" to version "+versioni.Versione)
			Logga(ctx, ctx.Value("JsonLog").(bool), "Make canary the new production "+istanza.Istanza)

			// FAC-744 - rendere tutte le precedenti versioni obsolete XDEPLOYLOG07 = 1

			switch versioni.TipoVersione {
			case "production", "Production":

				// rendo obsoleto il vecchio production
				keyvalueslice := make(map[string]interface{})
				keyvalueslice["debug"] = false
				keyvalueslice["source"] = "devops-8"
				keyvalueslice["XDEPLOYLOG06"] = "0"

				filter := "equals(XDEPLOYLOG04,'" + istanza.Istanza + "') and equals(XDEPLOYLOG03,'production') and XDEPLOYLOG06 eq 1"

				_, erro := ApiCallPUT(ctx, restyDebug, keyvalueslice, "ms"+devops, "/"+devops+"/DEPLOYLOG/"+filter, devopsToken, dominio, coreApiVersion)
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

				_, erro := ApiCallPUT(ctx, restyDebug, keyvalueslice, "ms"+devops, "/"+devops+"/DEPLOYLOG/"+filter, devopsToken, dominio, coreApiVersion)
				if erro.Errore < 0 {
					return erro
				}

				break
			}

			break
		}

	}

	Logga(ctx, ctx.Value("JsonLog").(bool), "Inserisco il record")

	canaryProduction = strings.ToLower(canaryProduction)
	ista := strings.ToLower(istanza.Istanza)

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
		Logga(ctx, ctx.Value("JsonLog").(bool), err.Error())
	}

	/* -------------------------- */
	// inserisco il nuovo canarino

	keyvalueslices := make([]map[string]interface{}, 0)
	keyvalueslice := make(map[string]interface{})
	keyvalueslice["debug"] = false
	keyvalueslice["source"] = "devops-8"
	keyvalueslice["XDEPLOYLOG03"] = canaryProduction
	keyvalueslice["XDEPLOYLOG04"] = ista
	keyvalueslice["XDEPLOYLOG05"] = versioneMicroservizio
	keyvalueslice["XDEPLOYLOG06"] = 1
	keyvalueslice["XDEPLOYLOG07"] = 0
	keyvalueslice["XDEPLOYLOG08"] = string(detailJson)
	keyvalueslice["XDEPLOYLOG09"] = enviro
	keyvalueslice["XDEPLOYLOG10"] = microfrontendJson
	keyvalueslices = append(keyvalueslices, keyvalueslice)

	resPOST := ApiCallPOST(ctx, restyDebug, keyvalueslices, "ms"+devops, "/"+devops+"/DEPLOYLOG", devopsToken, dominio, coreApiVersion)
	if resPOST.Errore < 0 {
		LoggaErrore.Log = resPOST.Log
		return LoggaErrore
	}

	Logga(ctx, ctx.Value("JsonLog").(bool), "updateIstanzaMicroservice end")
	Logga(ctx, ctx.Value("JsonLog").(bool), " - - - - - - - - - - - - - - - - - - - ")

	//os.Exit(0)
	return LoggaErrore
}
func UploadFileBucket(bucket, object, filename string) error {
	// bucket := "bucket-name"
	// object := "object-name"
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	// Open local file.
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	o := client.Bucket(bucket).Object(object)

	// Optional: set a generation-match precondition to avoid potential race
	// conditions and data corruptions. The request to upload is aborted if the
	// object's generation number does not match your precondition.
	// For an object that does not yet exist, set the DoesNotExist precondition.
	// o = o.If(storage.Conditions{DoesNotExist: true})
	// If the live object already exists in your bucket, set instead a
	// generation-match precondition using the live object's generation number.
	// attrs, err := o.Attrs(ctx)
	// if err != nil {
	// 	return err
	// }
	// o = o.If(storage.Conditions{GenerationMatch: attrs.Generation})

	// Upload an object with storage.Writer.
	wc := o.NewWriter(ctx)
	if _, err = io.Copy(wc, f); err != nil {
		return err
	}
	if err := wc.Close(); err != nil {
		return err
	}
	fmt.Println(filename + " uploaded in " + object)
	return nil
}
func GetGkeToken() (string, error) {
	cmd := exec.Command("bash", "-c", "gcloud config config-helper --format='value(credential.access_token)'")
	stdout, err := cmd.Output()
	if err != nil {
		return "", err
	}
	gkeToken := strings.TrimSuffix(string(stdout), "\n")
	return gkeToken, err
}
func CloudBuils(ctx context.Context, docker, verPad, dirRepo string, bArgs []string, swMonolith bool, cftoolenv TenantEnv) (BuildRes, error) {

	Logga(ctx, ctx.Value("JsonLog").(bool), "")
	Logga(ctx, ctx.Value("JsonLog").(bool), "CLOUD BUILD for "+docker)
	Logga(ctx, ctx.Value("JsonLog").(bool), "")

	var errBuild error
	fmt.Println(dirRepo + "-" + docker + "-" + verPad)

	nomeBucket := "q01io-325908_cloudbuild"

	tarFileName := docker + "_" + verPad + ".tar.gz"

	// ottengo un token
	gkeToken, errToken := GetGkeToken()
	if errToken != nil {
	}

	// Prepariamo la struct per fare la BUILD
	var cb CBuild
	var step1 BuildStep
	var step2 BuildStep
	cb.Source.StorageSource.Bucket = nomeBucket
	cb.Source.StorageSource.Object = "buildTgz/" + tarFileName
	cb.Options.MachineType = "E2_HIGHCPU_8"

	var img []string
	img = append(img, cftoolenv.CoreGkeUrl+"/"+cftoolenv.CoreGkeProject+"/"+docker+":"+verPad)
	cb.Images = img

	var args1 []string
	args1 = append(args1, "build")
	for _, ar := range bArgs {
		args1 = append(args1, ar)
	}
	args1 = append(args1, "-t")
	args1 = append(args1, cftoolenv.CoreGkeUrl+"/"+cftoolenv.CoreGkeProject+"/"+docker+":"+verPad)
	args1 = append(args1, ".")

	step1.Name = "gcr.io/cloud-builders/docker"
	step1.Args = args1

	var args2 []string
	args2 = append(args2, "push")
	args2 = append(args2, cftoolenv.CoreGkeUrl+"/"+cftoolenv.CoreGkeProject+"/"+docker+":"+verPad)

	step2.Name = "gcr.io/cloud-builders/docker"
	step2.Args = args2

	cb.Steps = append(cb.Steps, step1)
	cb.Steps = append(cb.Steps, step2)

	// lancio la BUILD
	cliB := resty.New()
	cliB.Debug = true
	cliB.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	restyResB, errApiB := cliB.R().
		SetAuthToken(gkeToken).
		SetBody(cb).
		Post("https://cloudbuild.googleapis.com/v1/projects/q01io-325908/locations/global/builds")
	if errApiB != nil {

	}

	var bres BuildRes
	if restyResB.StatusCode() != 200 {
		var brerr BuildError
		json.Unmarshal([]byte(restyResB.Body()), &brerr)
		errBuild = errors.New(brerr.Error.Message)
		return bres, errBuild
	}

	// code 200
	json.Unmarshal([]byte(restyResB.Body()), &bres)
	return bres, errBuild
}
func GetBuildStatus(ID string, cftoolenv TenantEnv, token string) (BuildStatus, error) {

	cli := resty.New()
	cli.Debug = true
	cli.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	restyRes, err := cli.R().
		SetAuthToken(token).
		Get("https://cloudbuild.googleapis.com/v1/projects/q01io-325908/builds/" + ID)
	if err != nil {

	}
	var bStatus BuildStatus
	json.Unmarshal([]byte(restyRes.Body()), &bStatus)

	return bStatus, err
}
func UpdateDockerVersion(ctx context.Context, docker, ver, user, devMaster, sha, team, newTagName, releaseNote, parentBranch, cs, merged, tenant, devopsToken, dominio, coreApiVersion string) error {

	var erro error
	Logga(ctx, ctx.Value("JsonLog").(bool), "Insert TB_ANAG_KUBEDKRBUILD00 ")

	restyDebug := false
	if os.Getenv("restyDebug") == "true" {
		restyDebug = true
	}
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

	Logga(ctx, ctx.Value("JsonLog").(bool), "beore ApiCallPOST")
	res := ApiCallPOST(ctx, restyDebug, keyvalueslices, "msdevops", "/devops/KUBEDKRBUILD", devopsToken, dominio, coreApiVersion)
	Logga(ctx, ctx.Value("JsonLog").(bool), "after ApiCallPOST")
	if res.Errore != 0 {
		Logga(ctx, ctx.Value("JsonLog").(bool), res.Log)
		erro = errors.New(res.Log)
		return erro
	}
	Logga(ctx, ctx.Value("JsonLog").(bool), "Insert TB_ANAG_KUBEDKRBUILD00 DONE")
	return erro
}

/*
Fa il merge di un branch sull'altro
se ci sono conflitti li segnala

accetta branch source, branch dest, tipo (tag o branch)
ritorna un LOG
*/
func GitMergeApi(ctx context.Context, src, dst, repo, tipo string, bitbucketEnv MergeToMaster) (string, string) {

	Logga(ctx, ctx.Value("JsonLog").(bool), "gitMergeApi")

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

		clientTagDel := resty.New()
		clientTagDel.Debug = false
		restyTagResponseDel, errDel := clientTagDel.R().
			SetHeader("Content-Type", "application/json").
			SetBasicAuth(bitbucketEnv.UserGit, bitbucketEnv.TokenGit).
			Delete(bitbucketEnv.ApiHostGit + "/repositories/" + bitbucketEnv.ProjectGit + "/" + repo + "/refs/branches/" + tmpBranch)

		if errDel != nil {
			fmt.Println("_##START##_   !!! Cannot delete previous temporary branch " + tmpBranch + " ERROR " + errDel.Error() + "_##STOP##_")
		}

		if restyTagResponseDel.StatusCode() == 404 {
			fmt.Println("_##START##_   *WARNING* Cannot delete previous temporary branch " + tmpBranch + " _##STOP##_")
		}

		Logga(ctx, ctx.Value("JsonLog").(bool), repo+": creo branch dal tag "+src)
		// creo un branch vivo dal tag

		body := `{"name": "` + tmpBranch + `","target": {  "hash": "` + src + `"}}`

		clientTag := resty.New()
		clientTag.Debug = false
		restyTagResponse, err := clientTag.R().
			SetHeader("Content-Type", "application/json").
			SetBasicAuth(bitbucketEnv.UserGit, bitbucketEnv.TokenGit).
			SetBody(body).
			Post(bitbucketEnv.ApiHostGit + "/repositories/" + bitbucketEnv.ProjectGit + "/" + repo + "/refs/branches")

		if err != nil {
			fmt.Println("_##START##_   !ERROR! New branch on " + repo + " ERROR " + err.Error() + "_##STOP##_")

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
	Logga(ctx, ctx.Value("JsonLog").(bool), repo+": faccio pull req di merge di "+src+" su "+dst)
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
		SetBasicAuth(bitbucketEnv.UserGit, bitbucketEnv.TokenGit).
		SetBody(body).
		Post(bitbucketEnv.ApiHostGit + "/repositories/" + bitbucketEnv.ProjectGit + "/" + repo + "/pullrequests")

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
			Logga(ctx, ctx.Value("JsonLog").(bool), repo+": faccio Merge di "+src+" su "+dst)
			mergeRes.Log += "Do MERGE of " + src + " on " + dst + "\n"

			clientMerge := resty.New()
			clientMerge.Debug = false
			respMerge, errMerge := clientMerge.R().
				SetBasicAuth(bitbucketEnv.UserGit, bitbucketEnv.TokenGit).
				Post(bitbucketEnv.ApiHostGit + "/repositories/" + bitbucketEnv.ProjectGit + "/" + repo + "/pullrequests/" + strconv.Itoa(restyRes.ID) + "/merge")
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
					SetBasicAuth(bitbucketEnv.UserGit, bitbucketEnv.TokenGit).
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
func CreaDirAndCloneDocker(ctx context.Context, dkr DockerStruct, dirToCreate, branch string, buildArgs BuildArgs) {

	Logga(ctx, ctx.Value("JsonLog").(bool), "Work on: "+dkr.Docker)
	Logga(ctx, ctx.Value("JsonLog").(bool), "Repo git: "+dkr.GitRepo)
	Logga(ctx, ctx.Value("JsonLog").(bool), "Repo git branch: "+branch)

	// REPO TEMPLATE DOCKER
	repoDocker := "https://" + buildArgs.UserGit + ":" + buildArgs.TokenGit + "@" + buildArgs.UrlGit + "/" + buildArgs.ProjectGit + "/docker-tmpl.git"
	// REPO TU BUILD
	repoproject := "https://" + buildArgs.UserGit + ":" + buildArgs.TokenGit + "@" + buildArgs.UrlGit + "/" + buildArgs.ProjectGit + "/" + dkr.GitRepo + ".git"

	dir := dirToCreate + "/" + dkr.Docker
	dirSrc := dir + "/src"

	// creo la dir del docker
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		Logga(ctx, ctx.Value("JsonLog").(bool), err.Error(), "error")
	}

	// mi porto a terra i dockerfile e tutto cio che mi serve per creare il docker
	GitClone(dir, repoDocker)
	GitCheckout(dir, dkr.Dockerfile)

	// remove .git
	err = os.RemoveAll(dir + "/.git")
	if err != nil {
		Logga(ctx, ctx.Value("JsonLog").(bool), err.Error(), "error")
	}

	// creo la dir src
	err = os.MkdirAll(dirSrc, 0755)
	if err != nil {
		Logga(ctx, ctx.Value("JsonLog").(bool), err.Error(), "error")
	}

	// mi porto a terra i file del progetto e mi porto al branch dichiarato
	GitClone(dirSrc, repoproject)
	GitCheckout(dirSrc, branch)

	// remove .git
	err = os.RemoveAll(dirSrc + "/.git")
	if err != nil {
		Logga(ctx, ctx.Value("JsonLog").(bool), err.Error(), "error")
	}
}
