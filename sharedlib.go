package lib

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-resty/resty/v2"

	"github.com/mozillazg/go-slugify"
)

func GetIstanceDetail(ctx context.Context, iresReq IresRequest, canaryProduction, dominio, coreApiVersion string) (IstanzaMicro, error) {

	Logga(ctx, os.Getenv("JsonLog"), "")
	Logga(ctx, os.Getenv("JsonLog"), " + + + + + + + + + + + + + + + + + + + +")
	Logga(ctx, os.Getenv("JsonLog"), "getIstanceDetail begin")

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
		Logga(ctx, os.Getenv("JsonLog"), "Getting KUBEIMICROSERV - deploy.go")
		argsImicro := make(map[string]string)
		argsImicro["source"] = "devops-8"
		argsImicro["$select"] = "XKUBEIMICROSERV04,XKUBEIMICROSERV05"
		argsImicro["center_dett"] = "dettaglio"
		argsImicro["$filter"] = "equals(XKUBEIMICROSERV03,'" + istanza + "') "

		restyKubeImicroservRes, errKubeImicroservRes := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsImicro, "ms"+devops, "/api/"+os.Getenv("coreApiVersion")+"/"+devops+"/KUBEIMICROSERV", devopsToken, dominio, coreApiVersion)
		if errKubeImicroservRes != nil {
			Logga(ctx, os.Getenv("JsonLog"), errKubeImicroservRes.Error())
			erro = errors.New(errKubeImicroservRes.Error())
			return ims, erro
		}

		if len(restyKubeImicroservRes.BodyJson) > 0 {
			microservice = restyKubeImicroservRes.BodyJson["XKUBEIMICROSERV04_COD"].(string)
			ims.Cluster = restyKubeImicroservRes.BodyJson["XKUBEIMICROSERV05"].(string)
			ims.Microservice = microservice

			Logga(ctx, os.Getenv("JsonLog"), "KUBEIMICROSERV OK")
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

				_, erroPost := ApiCallPOST(ctx, os.Getenv("RestyDebug"), keyvalueslices, "ms"+devops, "/api/"+os.Getenv("coreApiVersion")+"/"+devops+"/KUBEIMICROSERV", devopsToken, dominio, coreApiVersion)

				if erroPost != nil {
					Logga(ctx, os.Getenv("JsonLog"), "")
					Logga(ctx, os.Getenv("JsonLog"), "NON RIESCO A SCRIVRERE L'ISTANZA "+erroPost.Error())
					Logga(ctx, os.Getenv("JsonLog"), "")
					return ims, erroPost
				}

			} else {

				Logga(ctx, os.Getenv("JsonLog"), "KUBEIMICROSERV MISSING")
			}
		}
		Logga(ctx, os.Getenv("JsonLog"), "")
	} else {

		Logga(ctx, os.Getenv("JsonLog"), "CLUSTER PRESO DA iresReq "+iresReq.ClusterDst)
		// se siamo in migrazione non applichiamo questo metodo ma abbiamo necessita di avere il cluster valorizzato
		ims.Cluster = iresReq.ClusterDst
	}

	/* ************************************************************************************************ */
	// KUBESTAGE
	Logga(ctx, os.Getenv("JsonLog"), "Getting KUBESTAGE sharedlib")

	argsStage := make(map[string]string)
	argsStage["source"] = "devops-8"
	argsStage["$select"] = "XKUBESTAGE08,XKUBESTAGE09"
	argsStage["center_dett"] = "dettaglio"
	argsStage["$filter"] = "equals(XKUBESTAGE03,'" + ims.Cluster + "') "
	argsStage["$filter"] += " and equals(XKUBESTAGE04,'" + enviro + "') "

	//$filter=contains(XART20,'(kg)') or contains(XART20,'pizza')
	restyStageRes, _ := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsStage, "ms"+devops, "/api/"+os.Getenv("coreApiVersion")+"/"+devops+"/KUBESTAGE", devopsTokenDst, dominio, coreApiVersion)
	if restyStageRes.Errore < 0 {
		Logga(ctx, os.Getenv("JsonLog"), restyStageRes.Log)
	}

	var swProdStage int
	var depEnv string
	if len(restyStageRes.BodyJson) > 0 {
		depEnv, _ = restyStageRes.BodyJson["XKUBESTAGE09"].(string)
		swProdStage = int(restyStageRes.BodyJson["XKUBESTAGE08"].(float64))
		Logga(ctx, os.Getenv("JsonLog"), "KUBESTAGE: OK")
	} else {
		Logga(ctx, os.Getenv("JsonLog"), "KUBESTAGE: MISSING")
	}

	/* ************************************************************************************************ */
	// KUBECLUSTER

	// il 21 04 2023 mepo laszlo e frnc non si addoneno del motivo di una array bidimens
	// e a puorc schiattano un filtro sul cluster lasciando tutto invariato
	// ma ovviamente la matrice avra una sola KEY
	Logga(ctx, os.Getenv("JsonLog"), "Getting KUBECLUSTER")

	argsClu := make(map[string]string)
	argsClu["source"] = "devops-8"
	argsClu["center_dett"] = "allviews"
	argsClu["$filter"] = " equals(XKUBECLUSTER03,'" + ims.Cluster + "')"

	restyKubeCluRes, errKubeCluRes := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsClu, "ms"+devops, "/api/"+os.Getenv("coreApiVersion")+"/"+devops+"/KUBECLUSTER", devopsTokenDst, dominio, coreApiVersion)
	if errKubeCluRes != nil {
		Logga(ctx, os.Getenv("JsonLog"), errKubeCluRes.Error())

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
			case 1:
				profile = "qa"
			case 2:
				profile = "prod"
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

			restyKubeCluEnvRes, errKubeCluEnvRes := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsCluEnv, "ms"+devops, "/api/"+os.Getenv("coreApiVersion")+"/"+devops+"/KUBECLUSTERENV", devopsToken, dominio, coreApiVersion)
			if errKubeCluEnvRes != nil {
				Logga(ctx, os.Getenv("JsonLog"), errKubeCluEnvRes.Error())

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

				Logga(ctx, os.Getenv("JsonLog"), "KUBECLUSTERENV OK")
			}

			clus[x["XKUBECLUSTER03"].(string)] = clu

			// Logga(ctx, os.Getenv("JsonLog"), x["XKUBECLUSTER03"].(string))
			// LogJson(clu)

		}

		//ims.Clusters = clus

		Logga(ctx, os.Getenv("JsonLog"), "KUBECLUSTER OK")
	} else {
		Logga(ctx, os.Getenv("JsonLog"), "KUBECLUSTER MISSING")
	}
	Logga(ctx, os.Getenv("JsonLog"), "")

	/* ************************************************************************************************ */

	/* ************************************************************************************************ */
	// KUBECLUSTER
	Logga(ctx, os.Getenv("JsonLog"), "Getting KUBECLUSTER")

	if _, ok := clus[ims.Cluster]; ok {

		Logga(ctx, os.Getenv("JsonLog"), " +++ KUBECLUSTER OK +++ ")

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
		// ims.DepEnv = clus[ims.Cluster].DepEnv

		Logga(ctx, os.Getenv("JsonLog"), "KUBECLUSTER OK")
	} else {
		Logga(ctx, os.Getenv("JsonLog"), "KUBECLUSTER MISSING")
	}
	Logga(ctx, os.Getenv("JsonLog"), "")

	/* ************************************************************************************************ */
	// AMBDOMAIN

	Logga(ctx, os.Getenv("JsonLog"), "Getting AMBDOMAIN")

	argsAmbdomain := make(map[string]string)
	argsAmbdomain["source"] = "auth-1"
	argsAmbdomain["$select"] = "XAMBDOMAIN05,XAMBDOMAIN07,XAMBDOMAIN08,XAMBDOMAIN09,XAMBDOMAIN10,XAMBDOMAIN11"
	argsAmbdomain["center_dett"] = "dettaglio"
	argsAmbdomain["$filter"] += "  equals(XAMBDOMAIN04,'" + customerDomain + "') "
	restyAmbdomainRes, errAmbdomainRes := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsAmbdomain, "msauth", "/api/"+os.Getenv("coreApiVersion")+"/core/AMBDOMAIN", devopsToken, dominio, coreApiVersion)
	if errAmbdomainRes != nil {
		Logga(ctx, os.Getenv("JsonLog"), errAmbdomainRes.Error())

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
		Logga(ctx, os.Getenv("JsonLog"), "AMBDOMAIN OK")
	} else {
		Logga(ctx, os.Getenv("JsonLog"), "AMBDOMAIN MISSING")
	}
	Logga(ctx, os.Getenv("JsonLog"), "")

	/* ************************************************************************************************ */
	// KUBEMICROSERV
	Logga(ctx, os.Getenv("JsonLog"), "Getting KUBEMICROSERV")

	argsMS := make(map[string]string)
	argsMS["source"] = "devops-8"
	argsMS["$select"] = "XKUBEMICROSERV07,XKUBEMICROSERV09,XKUBEMICROSERV15,XKUBEMICROSERV18,XKUBEMICROSERV20,XKUBEMICROSERV21,XKUBEMICROSERV22"
	argsMS["center_dett"] = "dettaglio"
	argsMS["$filter"] = "equals(XKUBEMICROSERV05,'" + microservice + "') "
	restyKubeMSRes, errKubeMSRes := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsMS, "ms"+devops, "/api/"+os.Getenv("coreApiVersion")+"/"+devops+"/KUBEMICROSERV", devopsToken, dominio, coreApiVersion)
	if errKubeMSRes != nil {
		Logga(ctx, os.Getenv("JsonLog"), errKubeMSRes.Error())

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

		isRefappFloat := restyKubeMSRes.BodyJson["XKUBEMICROSERV21"].(float64)
		if isRefappFloat == 0 {
			ims.IsApp = false
		} else {
			ims.IsApp = true
		}

		ims.RefAppCode = strings.ToLower(slugify.Slugify(restyKubeMSRes.BodyJson["XKUBEMICROSERV22"].(string)))

		var scaleToZero bool
		scaleToZeroFloat := restyKubeMSRes.BodyJson["XKUBEMICROSERV20"].(float64)
		if scaleToZeroFloat == 0 {
			scaleToZero = false
		} else {
			scaleToZero = true
		}
		ims.ScaleToZero = scaleToZero

		swDb := int(restyKubeMSRes.BodyJson["XKUBEMICROSERV15"].(float64))
		microservicePublic = int(restyKubeMSRes.BodyJson["XKUBEMICROSERV18"].(float64))

		ims.SwDb = swDb
		Logga(ctx, os.Getenv("JsonLog"), "KUBEMICROSERV OK")
	} else {
		Logga(ctx, os.Getenv("JsonLog"), "KUBEMICROSERV MISSING")
	}
	Logga(ctx, os.Getenv("JsonLog"), "")

	/* ************************************************************************************************ */
	// AMB
	Logga(ctx, os.Getenv("JsonLog"), "Getting AMB")

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
	if iresReq.Istanza != iresReq.IstanzaDst && iresReq.IstanzaDst != "" {
		argsAmb["cluster"] = iresReq.ClusterDst
	} else {
		argsAmb["cluster"] = ims.Cluster
	}
	argsAmb["monolith"] = strconv.Itoa(int(ims.Monolith))
	argsAmb["env"] = strconv.Itoa(int(ims.ProfileInt))
	argsAmb["public"] = strconv.Itoa(microservicePublic)
	//argsAmb["swMultiEnvironment"] = ims.SwMultiEnvironment

	restyKubeAmbRes, errKubeAmbRes := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsAmb, "msauth", "/api/"+os.Getenv("coreApiVersion")+"/"+auth+"/getAmbDomainMs", devopsTokenDst, dominio, coreApiVersion)
	if errKubeAmbRes != nil {
		Logga(ctx, os.Getenv("JsonLog"), errKubeAmbRes.Error())

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
		Logga(ctx, os.Getenv("JsonLog"), "AMB OK")
	} else {
		Logga(ctx, os.Getenv("JsonLog"), "AMB MISSING")
	}
	// os.Exit(0)
	Logga(ctx, os.Getenv("JsonLog"), "")

	//Logga(ctx, os.Getenv("JsonLog"), ims)
	Logga(ctx, os.Getenv("JsonLog"), "", "getIstanceDetail end")
	Logga(ctx, os.Getenv("JsonLog"), "", " - - - - - - - - - - - - - - - - - - - ")
	Logga(ctx, os.Getenv("JsonLog"), "", "")
	//os.Exit(0)
	return ims, erro
}

// questo metodo restituisce cio che serve in caso in cui il MS e di tipo REFAPP
// calcola il GW, SE e VS di tutti i MS della APP
func GetLayerDueDetails(ctx context.Context, refappname, enviro, team, devopsToken, dominio, coreApiVersion string) (LayerMesh, error) {

	var layerDue LayerMesh

	Logga(ctx, os.Getenv("JsonLog"), "Get Layer Due Start")
	// entro su microservice per avere i ms

	/* ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++ */
	// CERCO TUTTI I MS che filtrero successivamente per avere solo i MS del tenant
	argsMs := make(map[string]string)
	argsMs["source"] = "devops-8"
	argsMs["$select"] = "XKUBEMICROSERV05"
	argsMs["center_dett"] = "visualizza"
	argsMs["$offset"] = "all"

	MsRes, errMsRes := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsMs, "msdevops", "/api/"+os.Getenv("coreApiVersion")+"/devops/KUBEMICROSERV", devopsToken, dominio, coreApiVersion)
	if errMsRes != nil {
		Logga(ctx, os.Getenv("JsonLog"), errMsRes.Error())
		erro := errors.New(errMsRes.Error())
		return layerDue, erro
	}
	// var micros, microsFullQ string
	var micros string
	if len(MsRes.BodyArray) > 0 {
		for _, x := range MsRes.BodyArray {
			_, errcast := x["XKUBEMICROSERV05"].(string)
			if !errcast {
				Logga(ctx, os.Getenv("JsonLog"), "XKUBEMICROSERV05 no cast")
				erro := errors.New("XKUBEMICROSERV05 no cast")
				return layerDue, erro
			}
			micros += "'" + x["XKUBEMICROSERV05"].(string) + "',"
		}
		micros = micros[0 : len(micros)-1]
	}
	/* ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++ */

	/* ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++ */
	// qui cerco solo i MS relativi all' APP

	argsApp := make(map[string]string)
	argsApp["$fullquery"] = " select XAPP04,XBOXPKG04,XBOXPKG03 "
	argsApp["$fullquery"] += " from TB_ANAG_REFAPPNEW00 "
	//argsApp["$fullquery"] += " join TB_ANAG_REFAPPCUSTOMER00 on (XREFAPPCUSTOMER09 = '" + refappname + "') "
	argsApp["$fullquery"] += " join TB_ANAG_APP00 on (XAPP03=XREFAPPNEW03) "
	argsApp["$fullquery"] += " join TB_ANAG_APPBOX00 on (XAPPBOX03=XREFAPPNEW03) "
	argsApp["$fullquery"] += " join TB_ANAG_BOXPKG00 on (XBOXPKG03=XAPPBOX04 and XBOXPKG04 in (" + micros + ")) "
	argsApp["$fullquery"] += " where 1 "
	argsApp["$fullquery"] += " and XREFAPPNEW05 = '" + refappname + "' "

	Logga(ctx, os.Getenv("JsonLog"), argsApp["$fullquery"])
	AppRes, errAppRes := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsApp, "msappman", "/api/"+os.Getenv("coreApiVersion")+"/appman/custom/REFAPPNEW/values", devopsToken, dominio, coreApiVersion)
	if errAppRes != nil {
		Logga(ctx, os.Getenv("JsonLog"), errAppRes.Error())
		erro := errors.New(errAppRes.Error())
		return layerDue, erro
	}

	var pkgs, appName string
	if len(AppRes.BodyArray) > 0 {
		//var mms []string
		for _, x := range AppRes.BodyArray {
			_, errcast := x["XBOXPKG04"].(string)
			if !errcast {
				Logga(ctx, os.Getenv("JsonLog"), "XBOXPKG04 no cast")
				erro := errors.New("XBOXPKG04 no cast")
				return layerDue, erro
			}

			appName = strings.ToLower(slugify.Slugify(x["XAPP04"].(string)))
			pkgs += "'" + x["XBOXPKG04"].(string) + "', "
			//mms = append(mms, x["XBOXPKG04"].(string))
		}
		pkgs = pkgs[0 : len(pkgs)-2]
	}
	/* ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++ */

	layerDue.AppName = appName

	// QUESTA QUERY MI DA TUTTE LE INFO PER CREARE IL LAYER 2
	// quindi le rotte per accedere al layer 3
	argsBr := make(map[string]string)
	argsBr["source"] = "devops-8"

	argsBr["$fullquery"] = "  select XKUBEIMICROSERV04,XKUBEENDPOINT09,XKUBECLUSTER15,XKUBECLUSTER22, XKUBEMICROSERV07, XKUBEMICROSERV20 "
	argsBr["$fullquery"] += " from TB_ANAG_KUBEIMICROSERV00 "
	argsBr["$fullquery"] += " join TB_ANAG_KUBEMICROSERV00 on (XKUBEMICROSERV05 = XKUBEIMICROSERV04) "
	argsBr["$fullquery"] += " join TB_ANAG_KUBECLUSTER00 on (XKUBEIMICROSERV05 = XKUBECLUSTER03) "
	argsBr["$fullquery"] += " join TB_ANAG_KUBEENDPOINT00 on (XKUBEENDPOINT05 = XKUBEIMICROSERV04 and XKUBEENDPOINT12 = 100 and XKUBEENDPOINT09 != '') "
	argsBr["$fullquery"] += " where 1>0 "
	argsBr["$fullquery"] += " AND XKUBEIMICROSERV04 in (" + pkgs + ") "
	argsBr["$fullquery"] += " AND XKUBEIMICROSERV06  = '" + enviro + "' "
	argsBr["$fullquery"] += " order by XKUBEIMICROSERV04"

	Logga(ctx, os.Getenv("JsonLog"), argsBr["$fullquery"])
	BrRes, errBrRes := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsBr, "msdevops", "/api/"+os.Getenv("coreApiVersion")+"/devops/custom/KUBEIMICROSERV/values", devopsToken, dominio, coreApiVersion)
	if errBrRes != nil {
		Logga(ctx, os.Getenv("JsonLog"), errBrRes.Error())
		erro := errors.New(errBrRes.Error())
		return layerDue, erro
	}

	var Ip string
	var gws []Gw
	var gw Gw

	// mi conservo IP e DOMINIO del cluster
	// popolo le rotte con il gruppo al posto del team che calcolero successivamente
	var se Se
	var vs Vs
	var InternalHostArr []string
	InternalHostArr = append(InternalHostArr, enviro+"-"+layerDue.AppName+".local")
	vs.InternalHost = InternalHostArr
	msOld := ""

	// azzecco il dominio interno del layer uno
	se.Hosts = append(se.Hosts, enviro+"-"+appName+".local")

	if len(BrRes.BodyArray) > 0 {
		for _, x := range BrRes.BodyArray {

			Ip = x["XKUBECLUSTER22"].(string)

			// SE
			if msOld == "" || msOld != x["XKUBEIMICROSERV04"].(string) {
				se.Hosts = append(se.Hosts, enviro+"-"+x["XKUBEIMICROSERV04"].(string)+".local")
			}
			msOld = x["XKUBEIMICROSERV04"].(string)

			// VS
			var v VsDetails
			v.DestinationHost = enviro + "-" + x["XKUBEIMICROSERV04"].(string) + ".local"
			v.Authority = enviro + "-" + x["XKUBEIMICROSERV04"].(string) + ".local"

			// SCALE TO ZERO
			stz := int(x["XKUBEMICROSERV20"].(float64))
			if stz == 1 {
				stzTeam, errStzTeam := GetTeamFromGroup(ctx, devopsToken, dominio, x["XKUBEMICROSERV07"].(string))
				if errStzTeam != nil {
					Logga(ctx, os.Getenv("JsonLog"), errStzTeam.Error())
					erro := errors.New(errStzTeam.Error())
					return layerDue, erro
				}
				v.DestinationHost = "istio-ingressgateway.istio-system.svc.cluster.local"
				v.Authority = x["XKUBEIMICROSERV04"].(string) + "." + enviro + "-" + stzTeam + "." + x["XKUBECLUSTER15"].(string)
			}

			v.Prefix = x["XKUBEENDPOINT09"].(string)
			vs.VsDetails = append(vs.VsDetails, v)
		}
	}

	// azzecco i GW
	var intDominioArr []string
	intDominioArr = append(intDominioArr, enviro+"-"+layerDue.AppName+".local")
	gw.IntDominio = intDominioArr
	gw.Protocol = "HTTP"
	gw.Name = "http"
	gw.Number = "80"
	gws = append(gws, gw)

	layerDue.Gw = gws
	// azzecco SE
	se.Ip = Ip
	layerDue.Se = se
	// azzecco VS
	layerDue.Vs = vs

	// +++++++++++++++++++++++++++++++++

	// cerco eventuali rotte esterne
	fillMarketPlaceRoute(&layerDue)

	Logga(ctx, os.Getenv("JsonLog"), "Get Layer Due END")

	return layerDue, nil
}
func GetLayerTreDetails(ctx context.Context, tenant, DominioCluster, microservice, enviro, team, devopsToken, dominio, coreApiVersion string) (LayerMesh, error) {

	var layerTre LayerMesh

	Logga(ctx, os.Getenv("JsonLog"), "Get Layer Tre Start")
	// entro su microservice per avere i ms

	/* ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++ */
	// qui cerco il dominio
	argsApp := make(map[string]string)

	argsApp["$fullquery"] = "  select XREFAPPCUSTOMER12,XREFAPPNEW04 "
	argsApp["$fullquery"] += " from TB_ANAG_REFAPPCUSTOMER00 "
	argsApp["$fullquery"] += " join TB_ANAG_REFAPPNEW00 on (XREFAPPNEW03 = XREFAPPCUSTOMER04) "
	argsApp["$fullquery"] += " where 1>0 "
	argsApp["$fullquery"] += " AND XREFAPPCUSTOMER03  = '" + tenant + "' "

	AppRes, errAppRes := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsApp, "msappman", "/api/"+os.Getenv("coreApiVersion")+"/appman/custom/REFAPPCUSTOMER/values", devopsToken, dominio, coreApiVersion)
	if errAppRes != nil {
		Logga(ctx, os.Getenv("JsonLog"), errAppRes.Error())
		erro := errors.New(errAppRes.Error())
		return layerTre, erro
	}

	var dominiCustomer, appID string

	if len(AppRes.BodyArray) > 0 {
		for _, x := range AppRes.BodyArray {
			dominiCustomer = x["XREFAPPCUSTOMER12"].(string)
			appID = x["XREFAPPNEW04"].(string)
		}
	}

	/* ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++ */
	fmt.Println(dominiCustomer)
	type Domini struct {
		Env string `json:"env"`
		Dom string `json:"dom"`
	}
	var dd []Domini
	// domini, errM := json.Marshal(dominiCustomer)
	// if errM != nil {
	// 	Logga(ctx, os.Getenv("JsonLog"), errM.Error())
	// 	erro := errors.New(errM.Error())
	// 	return layerTre, erro
	// }
	json.Unmarshal([]byte(dominiCustomer), &dd)
	var extDominio string

	LogJson(dd)

	for _, dom := range dd {
		fmt.Println(dom.Env, enviro)
		// qui prendo il dominio esterno
		if dom.Env == enviro {
			extDominio = dom.Dom
		}
	}

	/* ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++ */
	argsSr := make(map[string]string)
	argsSr["source"] = "appman-8"
	argsSr["$select"] = "XAPPSRV04,XAPPSRV05,XAPPSRV06"
	argsSr["center_dett"] = "visualizza"
	argsSr["$filter"] = "equals(XAPPSRV03,'" + appID + "') "

	Logga(ctx, os.Getenv("JsonLog"), argsSr["$fullquery"])
	SrRes, errSrRes := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsSr, "msappman", "/api/"+os.Getenv("coreApiVersion")+"/appman/APPSRV", devopsToken, dominio, coreApiVersion)
	if errSrRes != nil {
		Logga(ctx, os.Getenv("JsonLog"), errSrRes.Error())
		erro := errors.New(errSrRes.Error())
		return layerTre, erro
	}

	// POPOLO UN ARR con tutte le porte HTTP da mappare sul GW
	var gws []Gw
	if len(SrRes.BodyArray) > 0 {
		for _, x := range SrRes.BodyArray {
			var gw Gw

			if microservice == "msdevops" && x["XAPPSRV04"].(string) == "grpc-"+enviro {

				var extDominioArr []string
				extDominioArr = append(extDominioArr, extDominio)
				gw.ExtDominio = extDominioArr
				gw.Name = x["XAPPSRV04"].(string)
				gw.Number = strconv.Itoa(int(x["XAPPSRV05"].(float64)))
				gw.Protocol = x["XAPPSRV06"].(string)
				gws = append(gws, gw)

			} else {

				// per i ms che non sono msdevops basta il dominio interno e la porta 80
				if x["XAPPSRV06"].(string) == "HTTP" {
					var intDominioArr []string
					intDominioArr = append(intDominioArr, enviro+"-"+microservice+".local")
					gw.IntDominio = intDominioArr
					gw.Name = x["XAPPSRV04"].(string)
					gw.Number = strconv.Itoa(int(x["XAPPSRV05"].(float64)))
					gw.Protocol = x["XAPPSRV06"].(string)
					gws = append(gws, gw)
				}
			}
		}
	}
	/* ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++ */

	// gw OK
	layerTre.Gw = gws

	var vs Vs
	layerFt := LayerFt(tenant, enviro)

	// se la combo tenant e enviro e true lavoro il gateway nel nuovo modo
	if layerFt {
		if microservice == "msdevops" {
			var ExternalHostArr []string
			ExternalHostArr = append(ExternalHostArr, extDominio)
			vs.ExternalHost = ExternalHostArr
		}
	} else {

		hostEnv := ""

		dmns := ctx.Value("dominiEnv").(map[string]interface{})

		ClusterDomainOvr := dmns["ClusterDomainOvr"].(bool)
		ClusterDomainEnv := dmns["ClusterDomainEnv"].(string)
		ClusterDomain := dmns["ClusterDomain"].(string)

		// Se il dominio arriva da override messo in KUBECLUSTERENV allora non va messo il prefisso
		if ClusterDomainOvr {
			hostEnv = ClusterDomainEnv
		}

		var ExternalHostArr []string
		host := ""
		if enviro != "prod" {
			if strings.HasPrefix(ClusterDomain, "ms-"+enviro+".") {
				host = ClusterDomain
			} else {
				host = "ms-" + enviro + "." + ClusterDomain
			}
		} else {
			host = ClusterDomain
		}
		if hostEnv != "" {
			ExternalHostArr = append(ExternalHostArr, hostEnv)
		}
		ExternalHostArr = append(ExternalHostArr, host)
		vs.ExternalHost = ExternalHostArr

	}

	var InternalHostArr []string
	InternalHostArr = append(InternalHostArr, enviro+"-"+microservice+".local")
	vs.InternalHost = InternalHostArr
	layerTre.Vs = vs

	Logga(ctx, os.Getenv("JsonLog"), "Get Layer Tre END")

	return layerTre, nil
}
func LayerFt(tnt, enviro string) bool {

	type FutToggle struct {
		Int  bool `json:"int"`
		Qa   bool `json:"qa"`
		Uat  bool `json:"uat"`
		Demo bool `json:"demo"`
		Prod bool `json:"prod"`
	}
	type FutToggles map[string]FutToggle
	var fToogles FutToggles

	futureToggleFile, _ := os.ReadFile("layerFT.json")
	_ = json.Unmarshal([]byte(futureToggleFile), &fToogles)

	switch enviro {
	case "int":
		if fToogles[tnt].Int {
			return true
		} else {
			return false
		}
	case "qa":
		if fToogles[tnt].Qa {
			return true
		} else {
			return false
		}
	case "uat":
		if fToogles[tnt].Uat {
			return true
		} else {
			return false
		}
	case "demo":
		if fToogles[tnt].Demo {
			return true
		} else {
			return false
		}
	case "prod":
		if fToogles[tnt].Prod {
			return true
		} else {
			return false
		}
	}
	return false
}

// questo medoto è un harcoded di un futuro possibile MARKET PLACE
// le cose sono cambiate e quindo va fatto ex novo ( monodominio a multidominio per env ..... il plasma a terra)
func fillMarketPlaceRoute(layerDue *LayerMesh) {
	found := false
	for _, x := range layerDue.Vs.VsDetails {
		if strings.Contains(x.DestinationHost, "mscoreservice") {
			found = true
			break
		}
	}
	if !found {
		var vsD VsDetails
		vsD.DestinationHost = "prod-mscoreservice.local"
		vsD.Authority = "prod-mscoreservice.local"
		vsD.Prefix = "/api/" + os.Getenv("coreApiVersion") + "/core"
		layerDue.Vs.VsDetails = append(layerDue.Vs.VsDetails, vsD)

		vsD.DestinationHost = "prod-msauth.local"
		vsD.Authority = "prod-msauth.local"
		vsD.Prefix = "/api/" + os.Getenv("coreApiVersion") + "/auth"
		layerDue.Vs.VsDetails = append(layerDue.Vs.VsDetails, vsD)

		vsD.DestinationHost = "prod-msusers.local"
		vsD.Authority = "prod-msusers.local"
		vsD.Prefix = "/api/" + os.Getenv("coreApiVersion") + "/users"
		layerDue.Vs.VsDetails = append(layerDue.Vs.VsDetails, vsD)

		vsD.DestinationHost = "prod-mscoreservice.local"
		vsD.Authority = "prod-mscoreservice.local"
		vsD.Prefix = "/api/" + os.Getenv("coreApiVersion") + "/document"
		layerDue.Vs.VsDetails = append(layerDue.Vs.VsDetails, vsD)

		vsD.DestinationHost = "prod-mscoreservice.local"
		vsD.Authority = "prod-mscoreservice.local"
		vsD.Prefix = "/ref-app-cs"
		layerDue.Vs.VsDetails = append(layerDue.Vs.VsDetails, vsD)

		vsD.DestinationHost = "prod-msauth.local"
		vsD.Authority = "prod-msauth.local"
		vsD.Prefix = "/ref-app-login"
		layerDue.Vs.VsDetails = append(layerDue.Vs.VsDetails, vsD)

	}
}
func GetMsRoutes(ctx context.Context, routeJson RouteJson) (RouteMs, error) {

	Logga(ctx, os.Getenv("JsonLog"), "GetMsRoutes")
	var erro error
	var ms RouteMs

	team := strings.ToLower(routeJson.Team)
	cluster := routeJson.Cluster
	istanza := routeJson.Istanza

	var eps []Endpoint
	var services []Service
	var dkrs []RouteDocker

	devopsToken := routeJson.DevopsToken
	namespace := routeJson.Enviro + "-" + team

	nomeMicroservice := ""
	/* ************************************************************************************************ */
	// DOCKER AND PORTS

	argsDoker := make(map[string]string)
	argsDoker["source"] = "devops-8"

	argsDoker["$fullquery"] = " select XKUBEIMICROSERV04,XKUBEMICROSERVDKR03,XKUBEMICROSERVDKR04,XKUBESERVICEDKR05,XKUBESERVICEDKR06"
	argsDoker["$fullquery"] += " from TB_ANAG_KUBEIMICROSERV00 "
	argsDoker["$fullquery"] += " join TB_ANAG_KUBEMICROSERVDKR00 on (XKUBEIMICROSERV04=XKUBEMICROSERVDKR03) "
	argsDoker["$fullquery"] += " join TB_ANAG_KUBESERVICEDKR00 on (XKUBESERVICEDKR04=XKUBEMICROSERVDKR04) "
	argsDoker["$fullquery"] += " where XKUBEIMICROSERV03 = '" + istanza + "'  "

	Logga(ctx, os.Getenv("JsonLog"), argsDoker["$fullquery"])
	restyDokerRes, errDokerRes := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsDoker, "msdevops", "/api/"+os.Getenv("coreApiVersion")+"/devops/custom/KUBEIMICROSERV/values", devopsToken, os.Getenv("apiDomain"), os.Getenv("coreApiVersion"))

	if errDokerRes != nil {
		Logga(ctx, os.Getenv("JsonLog"), errDokerRes.Error())
		return ms, errDokerRes
	}
	if restyDokerRes.Errore < 0 {
		Logga(ctx, os.Getenv("JsonLog"), restyDokerRes.Log)
		erro = errors.New(restyDokerRes.Log)
		return ms, erro
	}

	if len(restyDokerRes.BodyArray) > 0 {

		for _, x := range restyDokerRes.BodyArray {

			nomeMicroservice = x["XKUBEIMICROSERV04"].(string)
			porta := strconv.FormatFloat(x["XKUBESERVICEDKR06"].(float64), 'f', 0, 64)
			/* ************************************************************************************************ */
			// ENDPOINTS

			sqlEndpoint := ""

			// per ogni servizio cerco gli endpoints
			sqlEndpoint += "select "
			sqlEndpoint += "ifnull(aa.XKUBEENDPOINT05, '') as microservice_src, "
			sqlEndpoint += "ifnull(aa.XKUBEENDPOINT07, '') as allowed_method, "
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
			sqlEndpoint += "XKUBEENDPOINT07, "
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
			sqlEndpoint += "XKUBEENDPOINT07, "
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
			sqlEndpoint += "and docker_src = '" + x["XKUBEMICROSERVDKR04"].(string) + "' "
			sqlEndpoint += "and port_src = '" + porta + "' "
			sqlEndpoint += "order by "
			sqlEndpoint += "length(priority), "
			sqlEndpoint += "priority, "
			sqlEndpoint += "route_src, "
			sqlEndpoint += "customer desc , "
			sqlEndpoint += "partner desc , "
			sqlEndpoint += "market desc "

			Logga(ctx, os.Getenv("JsonLog"), sqlEndpoint)
			argsEndpoint := make(map[string]string)
			argsEndpoint["source"] = "devops-8"
			argsEndpoint["$fullquery"] = sqlEndpoint

			restyKubeEndpointRes, erroEnd := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsEndpoint, "msdevops", "/api/"+os.Getenv("coreApiVersion")+"/devops/custom/KUBEENDPOINT/values", devopsToken, os.Getenv("apiDomain"), os.Getenv("coreApiVersion"))
			if erroEnd != nil {
				Logga(ctx, os.Getenv("JsonLog"), erroEnd.Error())
				return ms, erroEnd
			}

			if len(restyKubeEndpointRes.BodyArray) > 0 {
				for _, x := range restyKubeEndpointRes.BodyArray {

					var ep Endpoint
					ep.Priority = x["priority"].(string)

					ep.AllowedMethod = x["allowed_method"].(string)
					ep.MicroserviceDst = x["microservice_dst"].(string)
					ep.DockerDst = x["docker_dst"].(string)
					ep.TypeSrvDst = x["type_dst"].(string)
					ep.RouteDst = x["route_dst"].(string)
					ep.RewriteDst = x["rewrite_dst"].(string)
					ep.NamespaceDst = x["namespace_dst"].(string)
					ep.VersionDst = x["version_dst"].(string)

					ep.MicroserviceSrc = x["microservice_src"].(string)
					ep.DockerSrc = x["docker_src"].(string)
					ep.TypeSrvSrc = x["type_src"].(string)
					ep.RouteSrc = x["route_src"].(string)
					ep.RewriteSrc = x["rewrite_src"].(string)
					ep.NamespaceSrc = namespace
					ep.VersionSrc = ""

					ep.Domain = x["domain"].(string)
					ep.Market = x["market"].(string)
					ep.Partner = x["partner"].(string)
					ep.Customer = x["customer"].(string)
					if x["use_current_cluster_domain"].(string) == "1" {
						ep.ClusterDomain = routeJson.ClusterHost
					} else {
						ep.ClusterDomain = ""
					}
					eps = append(eps, ep)
				}
			}

			var service Service
			service.Port = porta
			service.Tipo = x["XKUBESERVICEDKR05"].(string) // http grpc
			service.Endpoint = eps
			services = append(services, service)
			eps = nil

			var dkr RouteDocker
			dkr.Docker = x["XKUBEMICROSERVDKR04"].(string)
			dkr.Service = services
			dkrs = append(dkrs, dkr)
			services = nil

		}

		ms.Microservice = nomeMicroservice
		ms.Istanza = istanza
		ms.Docker = dkrs

		// CERCA LE VERSIONI
		allMsVers, errV := GetIstanzaVersioni(ctx, istanza, routeJson.Enviro, devopsToken)
		if errV != nil {
			Logga(ctx, os.Getenv("JsonLog"), errV.Error())
			return ms, errV
		}
		ms.Version = allMsVers
	}
	return ms, erro
}

func GetIstanzaVersioni(ctx context.Context, istanza, enviro, devopsToken string) ([]RouteVersion, error) {

	Logga(ctx, os.Getenv("JsonLog"), "GetIstanzaVersioni - Getting DEPLOYLOG")
	var erro error
	var vrss []RouteVersion
	argsDeploy := make(map[string]string)
	argsDeploy["source"] = "devops-8"
	argsDeploy["center_dett"] = "visualizza"
	argsDeploy["$select"] = "XDEPLOYLOG03,XDEPLOYLOG04,XDEPLOYLOG05,XDEPLOYLOG07"
	argsDeploy["$filter"] = "equals(XDEPLOYLOG04,'" + istanza + "') "
	argsDeploy["$filter"] += " and equals(XDEPLOYLOG09,'" + enviro + "') "
	argsDeploy["$filter"] += " and equals(XDEPLOYLOG06,'1') "

	restyDeployRes, errDeployRes := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsDeploy, "msdevops", "/api/"+os.Getenv("coreApiVersion")+"/devops/DEPLOYLOG", devopsToken, os.Getenv("apiDomain"), os.Getenv("coreApiVersion"))
	if errDeployRes != nil {
		Logga(ctx, os.Getenv("JsonLog"), errDeployRes.Error())
		erro = errors.New(errDeployRes.Error())
		return vrss, erro
	}
	if len(restyDeployRes.BodyArray) > 0 {
		for _, x := range restyDeployRes.BodyArray {
			var vrs RouteVersion
			vrs.CanaryProduction = x["XDEPLOYLOG03"].(string) // canary production
			vrs.Versione = x["XDEPLOYLOG05"].(string)
			vrss = append(vrss, vrs)
		}
		Logga(ctx, os.Getenv("JsonLog"), "DEPLOYLOG OK")
	} else {
		Logga(ctx, os.Getenv("JsonLog"), "DEPLOYLOG MISSING")
	}
	return vrss, erro
}

func UploadFileToBucket(bucket, tarPathFilename, fileBucket, accessToken string) error {

	fileBytes, _ := os.ReadFile(tarPathFilename)
	cliUp := resty.New()
	cliUp.Debug = true
	cliUp.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	restyResUp, errApiUp := cliUp.R().
		SetHeader("Content-Type", "application/tar+gzip").
		SetAuthToken(accessToken).
		SetBody(fileBytes).
		Post("https://storage.googleapis.com/upload/storage/v1/b/" + bucket + "/o?uploadType=media&name=" + fileBucket)
	fmt.Println(restyResUp)
	if errApiUp != nil {
		return errApiUp
	}
	if restyResUp.StatusCode() != 200 {
		erro := errors.New("STATUS CODE: " + strconv.Itoa(restyResUp.StatusCode()))
		return erro
	}
	return nil
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

func CloudBuils(ctx context.Context, docker, verPad string, bArgs []string, cftoolenv TenantEnv, gkeToken string) (BuildRes, error) {

	Logga(ctx, os.Getenv("JsonLog"), "")
	Logga(ctx, os.Getenv("JsonLog"), "CLOUD BUILD for "+docker)
	Logga(ctx, os.Getenv("JsonLog"), "")

	var errBuild error

	nomeBucket := "q01io-325908_cloudbuild"

	tarFileName := docker + "_" + verPad + ".tar.gz"

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

	var bres BuildRes

	debool, errBool := strconv.ParseBool(os.Getenv("RestyDebug"))
	if errBool != nil {
		return bres, errBool
	}

	// lancio la BUILD
	cliB := resty.New()
	cliB.Debug = debool
	cliB.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	restyResB, errApiB := cliB.R().
		SetAuthToken(gkeToken).
		SetBody(cb).
		Post("https://cloudbuild.googleapis.com/v1/projects/q01io-325908/locations/global/builds")
	if errApiB != nil {

	}

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
	var bStatus BuildStatus
	cli := resty.New()
	cli.Debug = true
	cli.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	restyRes, err := cli.R().
		SetAuthToken(token).
		Get("https://cloudbuild.googleapis.com/v1/projects/q01io-325908/builds/" + ID)
	if err != nil {
		return bStatus, err
	}
	json.Unmarshal([]byte(restyRes.Body()), &bStatus)
	return bStatus, nil
}
func UpdateDockerVersion(ctx context.Context, docker, ver, user, devMaster, sha, team, newTagName, releaseNote, parentBranch, cs, merged, tenant, devopsToken, dominio, coreApiVersion string) error {

	var erro error
	Logga(ctx, os.Getenv("JsonLog"), "Insert TB_ANAG_KUBEDKRBUILD00 ")

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

	Logga(ctx, os.Getenv("JsonLog"), "beore ApiCallPOST")

	_, erroPost := ApiCallPOST(ctx, os.Getenv("RestyDebug"), keyvalueslices, "msdevops", "/api/"+os.Getenv("coreApiVersion")+"/devops/KUBEDKRBUILD", devopsToken, dominio, coreApiVersion)
	Logga(ctx, os.Getenv("JsonLog"), "after ApiCallPOST")
	if erroPost != nil {
		Logga(ctx, os.Getenv("JsonLog"), erroPost.Error())
		return erroPost
	}
	Logga(ctx, os.Getenv("JsonLog"), "Insert TB_ANAG_KUBEDKRBUILD00 DONE")
	return erro
}

/*
Fa il merge di un branch sull'altro
se ci sono conflitti li segnala

accetta branch source, branch dest, tipo (tag o branch)
ritorna un LOG
*/
func GitMergeApi(ctx context.Context, src, dst, repo, tipo string, bitbucketEnv MergeToMaster) (string, string) {

	Logga(ctx, os.Getenv("JsonLog"), "gitMergeApi")

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

		Logga(ctx, os.Getenv("JsonLog"), repo+": creo branch dal tag "+src)
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
	Logga(ctx, os.Getenv("JsonLog"), repo+": faccio pull req di merge di "+src+" su "+dst)
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
			Logga(ctx, os.Getenv("JsonLog"), repo+": faccio Merge di "+src+" su "+dst)
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
func CreaDirAndCloneDocker(ctx context.Context, dkr DockerStruct, dirToCreate, branch string, buildArgs BuildArgs) error {

	Logga(ctx, os.Getenv("JsonLog"), "Work on: "+dkr.Docker)
	Logga(ctx, os.Getenv("JsonLog"), "Repo git: "+dkr.GitRepo)
	Logga(ctx, os.Getenv("JsonLog"), "Repo git branch: "+branch)

	// REPO TEMPLATE DOCKER
	// faccio override se il TMPL non e presente sul DB del tenant
	repoDocker := ""
	if dkr.ProjectGit != "" {
		repoDocker = "https://" + dkr.UserGit + ":" + dkr.TokenGit + "@" + dkr.UrlGit + "/" + dkr.ProjectGit + "/docker-tmpl.git"
	} else {
		repoDocker = "https://" + buildArgs.UserGit + ":" + buildArgs.TokenGit + "@" + buildArgs.UrlGit + "/" + buildArgs.ProjectGit + "/docker-tmpl.git"
	}
	// REPO TU BUILD
	repoproject := "https://" + buildArgs.UserGit + ":" + buildArgs.TokenGit + "@" + buildArgs.UrlGit + "/" + buildArgs.ProjectGit + "/" + dkr.GitRepo + ".git"

	dirTmpl := dirToCreate + "/" + dkr.Docker
	dirSrc := dirTmpl + "/src"

	// creo la dir del docker
	err := os.MkdirAll(dirTmpl, 0755)
	if err != nil {
		return err
	}

	fmt.Println("Cloning " + repoDocker + " on " + dirTmpl)
	// mi porto a terra i dockerfile e tutto cio che mi serve per creare il docker
	_, errClone := git.PlainClone(dirTmpl, false, &git.CloneOptions{
		// Progress:      os.Stdout,
		URL:           repoDocker,
		Auth:          &http.BasicAuth{Username: buildArgs.UserGit, Password: buildArgs.TokenGit},
		ReferenceName: plumbing.NewBranchReferenceName(dkr.Dockerfile),
		SingleBranch:  true,
	})
	if errClone != nil {
		return err
	}

	// GitClone(dir, repoDocker)
	// GitCheckout(dir, dkr.Dockerfile)

	// remove .git
	err = os.RemoveAll(dirTmpl + "/.git")
	if err != nil {
		return err
	}

	// creo la dir src
	err = os.MkdirAll(dirSrc, 0755)
	if err != nil {
		return err
	}

	// mi porto a terra i file del progetto e mi porto al branch dichiarato
	_, errClone = git.PlainClone(dirSrc, false, &git.CloneOptions{
		// Progress:      os.Stdout,
		URL:           repoproject,
		Auth:          &http.BasicAuth{Username: buildArgs.UserGit, Password: buildArgs.TokenGit},
		ReferenceName: plumbing.NewBranchReferenceName(branch),
		SingleBranch:  true,
	})
	if errClone != nil {
		return err
	}
	// GitClone(dirSrc, repoproject)
	// GitCheckout(dirSrc, branch)

	// remove .git
	err = os.RemoveAll(dirSrc + "/.git")
	if err != nil {
		return err
	}
	return nil
}

func GetCurrentBranchSprintApi(ctx context.Context, token, team string) (CurrentSprintBranch, error) {

	var erro error

	// io e ciccio il 29 05 2024 decidiamo di accannare il monolite
	//

	/* ************************************************************************************************ */
	// KUBETEAMBRANCH
	Logga(ctx, os.Getenv("JsonLog"), "Getting KUBETEAMBRANCH")
	args := make(map[string]string)
	args["center_dett"] = "allviews"
	args["source"] = "devops-18"
	args["$filter"] = "equals(XKUBETEAMBRANCH03,'" + team + "')"
	args["$filter"] += "equals(XKUBETEAMBRANCH04,'MICROSERVICE')"
	args["$order"] = "XKUBETEAMBRANCH04"

	restyKubeTeamBranchRes, err := ApiCallGET(ctx, os.Getenv("RestyDebug"), args, "msdevops", "/api/"+os.Getenv("coreApiVersion")+"/devops/KUBETEAMBRANCH", token, os.Getenv("apiDomain"), os.Getenv("coreApiVersion"))
	if err != nil {
		Logga(ctx, os.Getenv("JsonLog"), err.Error())
	}
	if restyKubeTeamBranchRes.Errore != 0 {
		Logga(ctx, os.Getenv("JsonLog"), restyKubeTeamBranchRes.Log)
	}

	var CSB CurrentSprintBranch

	if len(restyKubeTeamBranchRes.BodyArray) > 0 {
		for _, x := range restyKubeTeamBranchRes.BodyArray {

			CSB.CurrentBranch = x["XKUBETEAMBRANCH05"].(string)
			CSB.Tipo = x["XKUBETEAMBRANCH04"].(string)
			CSB.Data = x["XKUBETEAMBRANCH06"].(string)

		}
		Logga(ctx, os.Getenv("JsonLog"), "KUBETEAMBRANCH OK")
	} else {
		erro = errors.New("KUBETEAMBRANCH MISSING - getCurrentBranchSprintApi")
		Logga(ctx, os.Getenv("JsonLog"), "KUBETEAMBRANCH MISSING - getCurrentBranchSprintApi")
	}

	return CSB, erro
}
