package lib

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
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
		Logga(ctx, os.Getenv("JsonLog"), "Getting KUBEIMICROSERV - deploy.go")
		argsImicro := make(map[string]string)
		argsImicro["source"] = "devops-8"
		argsImicro["$select"] = "XKUBEIMICROSERV04,XKUBEIMICROSERV05"
		argsImicro["center_dett"] = "dettaglio"
		argsImicro["$filter"] = "equals(XKUBEIMICROSERV03,'" + istanza + "') "

		restyKubeImicroservRes, errKubeImicroservRes := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsImicro, "ms"+devops, "/"+devops+"/KUBEIMICROSERV", devopsToken, dominio, coreApiVersion)
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

				resKubeims := ApiCallPOST(ctx, os.Getenv("RestyDebug"), keyvalueslices, "ms"+devops, "/"+devops+"/KUBEIMICROSERV", devopsToken, dominio, coreApiVersion)

				if resKubeims.Errore < 0 {
					Logga(ctx, os.Getenv("JsonLog"), "")
					Logga(ctx, os.Getenv("JsonLog"), "NON RIESCO A SCRIVRERE L'ISTANZA "+resKubeims.Log)
					Logga(ctx, os.Getenv("JsonLog"), "")

					erro = errors.New(resKubeims.Log)
					return ims, erro
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
	restyStageRes, _ := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsStage, "ms"+devops, "/"+devops+"/KUBESTAGE", devopsTokenDst, dominio, coreApiVersion)
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

	restyKubeCluRes, errKubeCluRes := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsClu, "ms"+devops, "/"+devops+"/KUBECLUSTER", devopsTokenDst, dominio, coreApiVersion)
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

			restyKubeCluEnvRes, errKubeCluEnvRes := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsCluEnv, "ms"+devops, "/"+devops+"/KUBECLUSTERENV", devopsToken, dominio, coreApiVersion)
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
		ims.DepEnv = clus[ims.Cluster].DepEnv

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
	restyAmbdomainRes, errAmbdomainRes := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsAmbdomain, "msauth", "/core/AMBDOMAIN", devopsToken, dominio, coreApiVersion)
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
	argsMS["$select"] = "XKUBEMICROSERV09,XKUBEMICROSERV15,XKUBEMICROSERV18,XKUBEMICROSERV20,XKUBEMICROSERV21,XKUBEMICROSERV22"
	argsMS["center_dett"] = "dettaglio"
	argsMS["$filter"] = "equals(XKUBEMICROSERV05,'" + microservice + "') "
	restyKubeMSRes, errKubeMSRes := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsMS, "ms"+devops, "/"+devops+"/KUBEMICROSERV", devopsToken, dominio, coreApiVersion)
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
		if isRefappFloat == 1 {
			ims.RefApp.RefAppName = strings.ToLower(slugify.Slugify(restyKubeMSRes.BodyJson["XKUBEMICROSERV22"].(string)))
			rfapp, errRefapp := fillRefapp(ctx, microservice, restyKubeMSRes.BodyJson["XKUBEMICROSERV22"].(string), devopsToken, dominio, coreApiVersion)
			ims.RefApp = rfapp
			if errRefapp != nil {
				Logga(ctx, os.Getenv("JsonLog"), errRefapp.Error())

				erro = errors.New(errRefapp.Error())
				return ims, erro
			}
		}

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
	argsAmb["cluster"] = ims.Cluster
	if ims.Monolith == 1 {
		argsAmb["refappID"] = ims.PodName // MWPO DICE CHE ANCHE SE CE SCRITTO refappID é GIUSTO PASSARE IL PODNAME
		argsAmb["customerDomain"] = customerDomain
	}
	argsAmb["monolith"] = strconv.Itoa(int(ims.Monolith))
	argsAmb["env"] = strconv.Itoa(int(ims.ProfileInt))
	argsAmb["public"] = strconv.Itoa(microservicePublic)
	//argsAmb["swMultiEnvironment"] = ims.SwMultiEnvironment

	restyKubeAmbRes, errKubeAmbRes := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsAmb, "msauth", "/"+auth+"/getAmbDomainMs", devopsTokenDst, dominio, coreApiVersion)
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

	/* ************************************************************************************************ */
	// DEPLOYLOG
	var erroIstanzaVersioni error
	ims.IstanzaMicroVersioni, erroIstanzaVersioni = GetIstanzaVersioni(ctx, iresReq, istanza, enviro, devops, devopsTokenDst, dominio, coreApiVersion)
	if erroIstanzaVersioni != nil {
		Logga(ctx, os.Getenv("JsonLog"), erroIstanzaVersioni.Error())
		return ims, erroIstanzaVersioni
	}
	// DEPLOYLOG
	/* ************************************************************************************************ */

	//Logga(ctx, os.Getenv("JsonLog"), ims)
	Logga(ctx, os.Getenv("JsonLog"), "", "getIstanceDetail end")
	Logga(ctx, os.Getenv("JsonLog"), "", " - - - - - - - - - - - - - - - - - - - ")
	Logga(ctx, os.Getenv("JsonLog"), "", "")
	//os.Exit(0)
	return ims, erro
}

// questo metodo restituisce cio che serve in caso in cui il MS e di tipo REFAPP
func fillRefapp(ctx context.Context, microservice, refappname, devopsToken, dominio, coreApiVersion string) (Refapp, error) {

	var refapp Refapp
	var dmes []BaseRoute

	Logga(ctx, os.Getenv("JsonLog"), "CERCO I MICROSERVIZI SU KUBEIMICROSERV")
	// entro su microservice per avere i ms
	argsMs := make(map[string]string)
	argsMs["source"] = "devops-8"
	argsMs["$select"] = "XKUBEMICROSERV05"
	argsMs["center_dett"] = "visualizza"

	MsRes, errMsRes := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsMs, "msdevops", "/devops/KUBEMICROSERV", devopsToken, dominio, coreApiVersion)
	if errMsRes != nil {
		Logga(ctx, os.Getenv("JsonLog"), errMsRes.Error())
		erro := errors.New(errMsRes.Error())
		return refapp, erro
	}
	var micros, microsFullQ string
	if len(MsRes.BodyArray) > 0 {
		for _, x := range MsRes.BodyArray {
			_, errcast := x["XKUBEMICROSERV05"].(string)
			if !errcast {
				Logga(ctx, os.Getenv("JsonLog"), "XKUBEMICROSERV05 no cast")
				erro := errors.New("XKUBEMICROSERV05 no cast")
				return refapp, erro
			}
			micros += x["XKUBEMICROSERV05"].(string) + ","
			microsFullQ += "'" + x["XKUBEMICROSERV05"].(string) + "', "
		}
		micros = "'" + micros[0:len(micros)-1] + "'"
		microsFullQ = microsFullQ[0 : len(microsFullQ)-2]
	}

	// ottengo il codice della refapp per agganciarmi a app
	argsRefapp := make(map[string]string)
	argsRefapp["source"] = "devops-8"
	argsRefapp["$select"] = "XREFAPPNEW03"
	argsRefapp["center_dett"] = "dettaglio"
	argsRefapp["$filter"] = "equals(XREFAPPNEW05,'" + refappname + "') "

	RefappRes, errRefappRes := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsRefapp, "msappman", "/appman/REFAPPNEW", devopsToken, dominio, coreApiVersion)
	if errRefappRes != nil {
		Logga(ctx, os.Getenv("JsonLog"), errRefappRes.Error())
		erro := errors.New(errRefappRes.Error())
		return refapp, erro
	}
	appID := ""
	if len(RefappRes.BodyJson) > 0 {
		_, errcast := RefappRes.BodyJson["XREFAPPNEW03"].(string)
		if !errcast {
			Logga(ctx, os.Getenv("JsonLog"), "XREFAPPNEW03 no cast")
			erro := errors.New("XREFAPPNEW03 no cast")
			return refapp, erro
		}
		appID = RefappRes.BodyJson["XREFAPPNEW03"].(string)
	}

	// entro su refappcustomer ed ottengo i domini
	argsRefappcustomer := make(map[string]string)
	argsRefappcustomer["source"] = "appman-8"
	argsRefappcustomer["$select"] = "XREFAPPCUSTOMER12"
	argsRefappcustomer["center_dett"] = "visualizza"
	argsRefappcustomer["$filter"] = "equals(XREFAPPCUSTOMER09,'" + refappname + "') "

	RefappcustomerRes, errRefappcustomerRes := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsRefappcustomer, "msappman", "/appman/REFAPPCUSTOMER", devopsToken, dominio, coreApiVersion)
	if errRefappcustomerRes != nil {
		Logga(ctx, os.Getenv("JsonLog"), errRefappcustomerRes.Error())
		erro := errors.New(errRefappcustomerRes.Error())
		return refapp, erro
	}
	var domini []string
	if len(RefappcustomerRes.BodyArray) > 0 {
		for _, x := range RefappcustomerRes.BodyArray {
			_, errcast := x["XREFAPPCUSTOMER12"].(string)
			if !errcast {
				Logga(ctx, os.Getenv("JsonLog"), "XREFAPPCUSTOMER12 no cast")
				erro := errors.New("XREFAPPCUSTOMER12 no cast")
				return refapp, erro
			}
			domini = append(domini, x["XREFAPPCUSTOMER12"].(string))
		}
	}

	// ottengo il nome della refapp
	argsApp := make(map[string]string)
	argsApp["source"] = "appman-8"
	argsApp["$select"] = "XAPP17"
	argsApp["center_dett"] = "dettaglio"
	argsApp["$filter"] = "equals(XAPP03,'" + appID + "') "

	AppRes, errAppRes := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsApp, "msappman", "/appman/APP", devopsToken, dominio, coreApiVersion)
	if errAppRes != nil {
		Logga(ctx, os.Getenv("JsonLog"), errAppRes.Error())
		erro := errors.New(errAppRes.Error())
		return refapp, erro
	}
	if len(AppRes.BodyJson) > 0 {
		_, errcast := AppRes.BodyJson["XAPP17"].(string)
		if !errcast {
			Logga(ctx, os.Getenv("JsonLog"), "XAPP17 no cast")
			erro := errors.New("XAP17 no cast")
			return refapp, erro
		}
		refapp.RefAppName = strings.ToLower(slugify.Slugify(AppRes.BodyJson["XAPP17"].(string)))
	}

	// entro su appbox per avere i box
	argsAppbox := make(map[string]string)
	argsAppbox["source"] = "appman-8"
	argsAppbox["$select"] = "XAPPBOX04"
	argsAppbox["center_dett"] = "visualizza"
	argsAppbox["$filter"] = "equals(XAPPBOX03,'" + appID + "') "

	AppboxRes, errAppboxRes := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsAppbox, "msappman", "/appman/APPBOX", devopsToken, dominio, coreApiVersion)
	if errAppboxRes != nil {
		Logga(ctx, os.Getenv("JsonLog"), errAppboxRes.Error())
		erro := errors.New(errAppboxRes.Error())
		return refapp, erro
	}
	var boxes string
	if len(AppboxRes.BodyArray) > 0 {
		for _, x := range AppboxRes.BodyArray {
			_, errcast := x["XAPPBOX04_COD"].(string)
			if !errcast {
				Logga(ctx, os.Getenv("JsonLog"), "XAPPBOX04_COD no cast")
				erro := errors.New("XAPPBOX04_COD no cast")
				return refapp, erro
			}
			boxes += x["XAPPBOX04_COD"].(string) + ","
		}
		boxes = "'" + boxes[0:len(boxes)-1] + "'"
	}

	// entro su boxpkg per avere i microservizi
	argsBoxpkg := make(map[string]string)

	//argsBoxpkg["debug"] = "true"
	argsBoxpkg["source"] = "appman-8"
	argsBoxpkg["$select"] = "XBOXPKG04"
	argsBoxpkg["center_dett"] = "visualizza"
	filtro := "in_s(XBOXPKG03," + boxes + ") "
	filtro += " and in_s(XBOXPKG04," + micros + ")"
	argsBoxpkg["$filter"] = filtro

	BoxpkgRes, errBoxpkgRes := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsBoxpkg, "msappman", "/appman/BOXPKG", devopsToken, dominio, coreApiVersion)
	if errBoxpkgRes != nil {
		Logga(ctx, os.Getenv("JsonLog"), errBoxpkgRes.Error())
		erro := errors.New(errBoxpkgRes.Error())
		return refapp, erro
	}
	var pkgs string
	if len(BoxpkgRes.BodyArray) > 0 {
		for _, x := range BoxpkgRes.BodyArray {
			_, errcast := x["XBOXPKG04_COD"].(string)
			if !errcast {
				Logga(ctx, os.Getenv("JsonLog"), "XBOXPKG04_COD no cast")
				erro := errors.New("XBOXPKG04_COD no cast")
				return refapp, erro
			}
			pkgs += "'" + x["XBOXPKG04_COD"].(string) + "', "
		}
		pkgs = pkgs[0 : len(pkgs)-2]
	}

	// entro su microservice per avere i ms
	argsBr := make(map[string]string)
	argsBr["source"] = "devops-8"
	argsBr["$fullquery"] = " select XKUBECLUSTER15,XKUBECLUSTER22,XKUBEIMICROSERV04,XKUBEIMICROSERV05, XKUBEIMICROSERV06, XKUBEMICROSERV07 "
	argsBr["$fullquery"] += " from TB_ANAG_KUBEIMICROSERV00 "
	argsBr["$fullquery"] += " join TB_ANAG_KUBEMICROSERV00 on (XKUBEMICROSERV05 = XKUBEIMICROSERV04) "
	argsBr["$fullquery"] += " join TB_ANAG_KUBECLUSTER00 on (XKUBEIMICROSERV05 = XKUBECLUSTER03) "
	argsBr["$fullquery"] += " where XKUBEIMICROSERV04 in (" + microsFullQ + ") "
	Logga(ctx, os.Getenv("JsonLog"), argsBr["$fullquery"])
	BrRes, errBrRes := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsBr, "msdevops", "/devops/custom/KUBEIMICROSERV/values", devopsToken, dominio, coreApiVersion)
	if errBrRes != nil {
		Logga(ctx, os.Getenv("JsonLog"), errBrRes.Error())
		erro := errors.New(errBrRes.Error())
		return refapp, erro
	}

	if len(BrRes.BodyArray) > 0 {
		for _, x := range BrRes.BodyArray {
			var dme BaseRoute
			dme.Domino = x["XKUBECLUSTER15"].(string)
			dme.Env = x["XKUBEIMICROSERV06"].(string)
			dme.Team = x["XKUBEMICROSERV07"].(string)
			dme.Ip = x["XKUBECLUSTER22"].(string)
			dme.BaseRoute = "/" + x["XKUBEIMICROSERV06"].(string) + "-" + strings.ToLower(x["XKUBEMICROSERV07"].(string)) + "/"
			dmes = append(dmes, dme)
		}
	}

	// sgrasso i doppioni
	var dmesOK []BaseRoute
	var dmeOK BaseRoute
	for _, x := range dmes {
		inArr := false
		for _, y := range dmesOK {
			if y.BaseRoute == x.BaseRoute {
				inArr = true
				break
			}
		}
		if !inArr {
			dmeOK.BaseRoute = x.BaseRoute
			dmeOK.Domino = x.Domino
			dmeOK.Env = x.Env
			dmeOK.Team = x.Team
			dmeOK.Ip = x.Ip
			dmesOK = append(dmesOK, dmeOK)
		}
	}

	// cambio il gruppo in team da GRU
	var gruppiArr []string
	gruteam := make(map[string]string)
	for idx, x := range dmesOK {
		inArr := false
		for _, y := range gruppiArr {
			if y == x.Team {
				inArr = true
				break
			}
		}
		if !inArr {
			// ottengo da gru il nome del team
			argsGru := make(map[string]string)
			argsGru["source"] = "devops-8"
			argsGru["$select"] = "XGRU05"
			argsGru["center_dett"] = "dettaglio"
			argsGru["$filter"] = "equals(XGRU03,'" + x.Team + "') "

			GruRes, errGruRes := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsGru, "msusers", "/users/GRU", devopsToken, dominio, coreApiVersion)
			if errGruRes != nil {
				Logga(ctx, os.Getenv("JsonLog"), errGruRes.Error())
				erro := errors.New(errGruRes.Error())
				return refapp, erro
			}

			if len(GruRes.BodyJson) > 0 {
				team, errcast := GruRes.BodyJson["XGRU05"].(string)
				if !errcast {
					Logga(ctx, os.Getenv("JsonLog"), "XGRU05 no cast")
					erro := errors.New("XGRU05 no cast")
					return refapp, erro
				}
				dmesOK[idx].Team = strings.ToLower(team)
				gruteam[x.Team] = strings.ToLower(team)
				dmesOK[idx].BaseRoute = "/" + x.Env + "-" + strings.ToLower(team) + "/"
				dmesOK[idx].Domino = x.Env + "-" + strings.ToLower(team) + "." + x.Domino
				gruppiArr = append(gruppiArr, x.Team)
			}

		} else {
			dmesOK[idx].Team = gruteam[x.Team]
			dmesOK[idx].BaseRoute = "/" + x.Env + "-" + gruteam[x.Team] + "/"
			dmesOK[idx].Domino = x.Env + "-" + gruteam[x.Team] + "." + x.Domino
		}
	}

	// cerco eventuali rotte esterne
	fillMarketPlaceRoute(&dmesOK)
	refapp.BaseRoute = dmesOK
	Logga(ctx, os.Getenv("JsonLog"), "FINE CERCO I MICROSERVIZI SU KUBEIMICROSERV")

	Logga(ctx, os.Getenv("JsonLog"), "CERCO LE PORTE DEI MS PER GW")

	// leggo le porte da aprire sul GW
	argsSr := make(map[string]string)
	argsSr["source"] = "appman-8"
	argsSr["$select"] = "XAPPSRV04,XAPPSRV05,XAPPSRV06"
	argsSr["center_dett"] = "visualizza"
	argsSr["$filter"] = "equals(XAPPSRV03,'" + appID + "') "

	Logga(ctx, os.Getenv("JsonLog"), argsSr["$fullquery"])
	SrRes, errSrRes := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsSr, "msappman", "/appman/APPSRV", devopsToken, dominio, coreApiVersion)
	if errSrRes != nil {
		Logga(ctx, os.Getenv("JsonLog"), errSrRes.Error())
		erro := errors.New(errSrRes.Error())
		return refapp, erro
	}

	var srvs []Server
	if len(SrRes.BodyArray) > 0 {
		for _, x := range SrRes.BodyArray {
			var srv Server
			srv.Domini = domini
			srv.Name = x["XAPPSRV04"].(string)
			srv.Number = strconv.Itoa(int(x["XAPPSRV05"].(float64)))
			srv.Protocol = x["XAPPSRV06"].(string)
			srvs = append(srvs, srv)
		}
	}
	refapp.Servers = srvs

	return refapp, nil
}

// questo medoto è un harcoded di un futuro possibile MARKET PLACE
func fillMarketPlaceRoute(dmesOK *[]BaseRoute) {
	found := false
	for _, x := range *dmesOK {
		if x.Domino == "q01.io" {
			found = true
			break
		}
	}
	if !found {
		var dme BaseRoute
		dme.BaseRoute = "prod-core.q01.io"
		dme.Domino = "q01.io"
		dme.Env = "prod"
		dme.Team = "core"
		dme.Ip = "" // lasciare vuoto indica che e un mondo esterno !!!! DNS
		*dmesOK = append(*dmesOK, dme)
	}
}
func GetMsRoutes(ctx context.Context, routeJson RouteJson) ([]Service, error) {

	Logga(ctx, os.Getenv("JsonLog"), "GetMsRoutes")
	var erro error
	var eps []Endpoint

	devopsToken := routeJson.DevopsToken
	team := strings.ToLower(routeJson.Team)
	namespace := routeJson.Enviro + "-" + team
	cluster := routeJson.Cluster

	var services []Service
	/* ************************************************************************************************ */
	// DOCKER AND PORTS

	argsDoker := make(map[string]string)
	argsDoker["source"] = "devops-8"

	argsDoker["$fullquery"] = " select XKUBEMICROSERVDKR04,XKUBESERVICEDKR06 from TB_ANAG_KUBEIMICROSERV00 "
	argsDoker["$fullquery"] += " join TB_ANAG_KUBEMICROSERVDKR00 on (XKUBEIMICROSERV04=XKUBEMICROSERVDKR03) "
	argsDoker["$fullquery"] += " join TB_ANAG_KUBESERVICEDKR00 on (XKUBESERVICEDKR04=XKUBEMICROSERVDKR04) "
	argsDoker["$fullquery"] += " where XKUBEIMICROSERV05 = '" + cluster + "'   and XKUBEIMICROSERV08 ='" + team + "' "
	Logga(ctx, os.Getenv("JsonLog"), argsDoker["$fullquery"])
	restyDokerRes, errDokerRes := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsDoker, "msdevops", "/core/custom/KUBEIMICROSERV/values", devopsToken, os.Getenv("loginApiHost"), os.Getenv("coreApiVersion"))

	if errDokerRes != nil {
		Logga(ctx, os.Getenv("JsonLog"), errDokerRes.Error())
		return services, errDokerRes
	}
	if restyDokerRes.Errore < 0 {
		Logga(ctx, os.Getenv("JsonLog"), restyDokerRes.Log)
		erro = errors.New(restyDokerRes.Log)
		return services, erro
	}

	if len(restyDokerRes.BodyArray) > 0 {
		var port, tipo string
		for _, x := range restyDokerRes.BodyArray {

			docker := x["XKUBEMICROSERVDKR04"].(string)
			port = strconv.FormatFloat(x["XKUBESERVICEDKR06"].(float64), 'f', 0, 64)

			/* ************************************************************************************************ */
			// ENDPOINTS

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

			restyKubeEndpointRes, erroEnd := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsEndpoint, "msdevops", "/core/custom/KUBEENDPOINT/values", devopsToken, os.Getenv("loginApiHost"), os.Getenv("coreApiVersion"))
			if erroEnd != nil {
				Logga(ctx, os.Getenv("JsonLog"), erroEnd.Error())
				return services, erroEnd
			}

			if len(restyKubeEndpointRes.BodyArray) > 0 {
				for _, x := range restyKubeEndpointRes.BodyArray {

					var ep Endpoint
					ep.Priority = x["priority"].(string)

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
			/* ************************************************************************************************ */
		}
		var service Service
		service.Port = port
		service.Tipo = tipo
		service.Endpoint = eps

		services = append(services, service)
	}
	return services, erro
}
func GetIstanzaVersioni(ctx context.Context, iresReq IresRequest, istanza, enviro, devops, devopsTokenDst, dominio, coreApiVersion string) ([]IstanzaMicroVersioni, error) {
	Logga(ctx, os.Getenv("JsonLog"), "Getting DEPLOYLOG")
	var erro error
	var istanzaMicroVersioni []IstanzaMicroVersioni

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

	restyDeployRes, errDeployRes := ApiCallGET(ctx, os.Getenv("RestyDebug"), argsDeploy, "ms"+devops, "/"+devops+"/DEPLOYLOG", devopsTokenDst, dominio, coreApiVersion)
	if errDeployRes != nil {
		Logga(ctx, os.Getenv("JsonLog"), errDeployRes.Error())
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
		Logga(ctx, os.Getenv("JsonLog"), "DEPLOYLOG OK")
	} else {
		Logga(ctx, os.Getenv("JsonLog"), "DEPLOYLOG MISSING")
	}
	Logga(ctx, os.Getenv("JsonLog"), "")

	return istanzaMicroVersioni, erro
}
func UpdateIstanzaMicroservice(ctx context.Context, canaryProduction, versioneMicroservizio string, istanza IstanzaMicro, micros Microservice, utente, enviro, devopsToken, dominio, coreApiVersion, microfrontendJson string) error {

	Logga(ctx, os.Getenv("JsonLog"), "")
	Logga(ctx, os.Getenv("JsonLog"), " + + + + + + + + + + + + + + + + + + + + ")
	Logga(ctx, os.Getenv("JsonLog"), "updateIstanzaMicroservice begin")
	Logga(ctx, os.Getenv("JsonLog"), "versioneMicroservizio "+versioneMicroservizio)
	for _, ccc := range istanza.IstanzaMicroVersioni {
		Logga(ctx, os.Getenv("JsonLog"), ccc.TipoVersione+" "+ccc.Versione)
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

				Logga(ctx, os.Getenv("JsonLog"), "Old canary found")

				Logga(ctx, os.Getenv("JsonLog"), "Delete canary "+istanza.Istanza+"-v"+versioni.Versione)
				Logga(ctx, os.Getenv("JsonLog"), "Make obsolete canary "+istanza.Istanza+" to version "+versioni.Versione)
				Logga(ctx, os.Getenv("JsonLog"), "New canary "+istanza.Istanza+" to version "+versioneMicroservizio)

				// rendo obsoleto il vecchio canarino
				keyvalueslice := make(map[string]interface{})
				keyvalueslice["debug"] = false
				keyvalueslice["source"] = "devops-8"
				keyvalueslice["XDEPLOYLOG06"] = "0"

				filter := "equals(XDEPLOYLOG04,'" + istanza.Istanza + "') and equals(XDEPLOYLOG03,'canary') and XDEPLOYLOG06 eq 1"

				_, erro := ApiCallPUT(ctx, os.Getenv("RestyDebug"), keyvalueslice, "ms"+devops, "/"+devops+"/DEPLOYLOG/"+filter, devopsToken, dominio, coreApiVersion)

				if erro != nil {
					return erro
				}
			}

			break

		case "production", "Production":

			Logga(ctx, os.Getenv("JsonLog"), "Delete production "+istanza.Istanza+"-v"+versioni.Versione)
			Logga(ctx, os.Getenv("JsonLog"), "Make obsolete production "+istanza.Istanza+" to version "+versioni.Versione)
			Logga(ctx, os.Getenv("JsonLog"), "Make canary the new production "+istanza.Istanza)

			// FAC-744 - rendere tutte le precedenti versioni obsolete XDEPLOYLOG07 = 1

			switch versioni.TipoVersione {
			case "production", "Production":

				// rendo obsoleto il vecchio production
				keyvalueslice := make(map[string]interface{})
				keyvalueslice["debug"] = false
				keyvalueslice["source"] = "devops-8"
				keyvalueslice["XDEPLOYLOG06"] = "0"

				filter := "equals(XDEPLOYLOG04,'" + istanza.Istanza + "') and equals(XDEPLOYLOG03,'production') and XDEPLOYLOG06 eq 1"

				_, erro := ApiCallPUT(ctx, os.Getenv("RestyDebug"), keyvalueslice, "ms"+devops, "/"+devops+"/DEPLOYLOG/"+filter, devopsToken, dominio, coreApiVersion)
				if erro != nil {
					return erro
				}

			case "canary", "Canary":

				// rendo obsoleto il canarino
				keyvalueslice := make(map[string]interface{})
				keyvalueslice["debug"] = false
				keyvalueslice["source"] = "devops-8"
				keyvalueslice["XDEPLOYLOG06"] = "0"

				filter := "equals(XDEPLOYLOG04,'" + istanza.Istanza + "') and equals(XDEPLOYLOG03,'canary')"

				_, erro := ApiCallPUT(ctx, os.Getenv("RestyDebug"), keyvalueslice, "ms"+devops, "/"+devops+"/DEPLOYLOG/"+filter, devopsToken, dominio, coreApiVersion)
				if erro != nil {
					return erro
				}

				break
			}

			break
		}

	}

	Logga(ctx, os.Getenv("JsonLog"), "Inserisco il record")

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
		Logga(ctx, os.Getenv("JsonLog"), err.Error())
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

	resPOST := ApiCallPOST(ctx, os.Getenv("RestyDebug"), keyvalueslices, "ms"+devops, "/"+devops+"/DEPLOYLOG", devopsToken, dominio, coreApiVersion)
	if resPOST.Errore < 0 {
		erro := errors.New(resPOST.Log)
		return erro
	}

	Logga(ctx, os.Getenv("JsonLog"), "updateIstanzaMicroservice end")
	Logga(ctx, os.Getenv("JsonLog"), " - - - - - - - - - - - - - - - - - - - ")

	//os.Exit(0)
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
func GetGkeDevopsToken(salt string) string {
	ct := time.Now()
	date := ct.Format("20060102")
	year := date[0:4]
	month := date[4:6]
	day := date[6:8]
	secret := year + "." + salt + "." + month + "." + day

	h := sha256.New()
	h.Write([]byte(secret))
	sha1_hash := hex.EncodeToString(h.Sum(nil))
	return base64.StdEncoding.EncodeToString([]byte(sha1_hash))
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

	Logga(ctx, os.Getenv("JsonLog"), "")
	Logga(ctx, os.Getenv("JsonLog"), "CLOUD BUILD for "+docker)
	Logga(ctx, os.Getenv("JsonLog"), "")

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
	res := ApiCallPOST(ctx, os.Getenv("RestyDebug"), keyvalueslices, "msdevops", "/devops/KUBEDKRBUILD", devopsToken, dominio, coreApiVersion)
	Logga(ctx, os.Getenv("JsonLog"), "after ApiCallPOST")
	if res.Errore != 0 {
		Logga(ctx, os.Getenv("JsonLog"), res.Log)
		erro = errors.New(res.Log)
		return erro
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
func CreaDirAndCloneDocker(ctx context.Context, dkr DockerStruct, dirToCreate, branch string, buildArgs BuildArgs) {

	Logga(ctx, os.Getenv("JsonLog"), "Work on: "+dkr.Docker)
	Logga(ctx, os.Getenv("JsonLog"), "Repo git: "+dkr.GitRepo)
	Logga(ctx, os.Getenv("JsonLog"), "Repo git branch: "+branch)

	// REPO TEMPLATE DOCKER
	repoDocker := "https://" + buildArgs.UserGit + ":" + buildArgs.TokenGit + "@" + buildArgs.UrlGit + "/" + buildArgs.ProjectGit + "/docker-tmpl.git"
	// REPO TU BUILD
	repoproject := "https://" + buildArgs.UserGit + ":" + buildArgs.TokenGit + "@" + buildArgs.UrlGit + "/" + buildArgs.ProjectGit + "/" + dkr.GitRepo + ".git"

	dir := dirToCreate + "/" + dkr.Docker
	dirSrc := dir + "/src"

	// creo la dir del docker
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		Logga(ctx, os.Getenv("JsonLog"), err.Error(), "error")
	}

	// mi porto a terra i dockerfile e tutto cio che mi serve per creare il docker
	GitClone(dir, repoDocker)
	GitCheckout(dir, dkr.Dockerfile)

	// remove .git
	err = os.RemoveAll(dir + "/.git")
	if err != nil {
		Logga(ctx, os.Getenv("JsonLog"), err.Error(), "error")
	}

	// creo la dir src
	err = os.MkdirAll(dirSrc, 0755)
	if err != nil {
		Logga(ctx, os.Getenv("JsonLog"), err.Error(), "error")
	}

	// mi porto a terra i file del progetto e mi porto al branch dichiarato
	GitClone(dirSrc, repoproject)
	GitCheckout(dirSrc, branch)

	// remove .git
	err = os.RemoveAll(dirSrc + "/.git")
	if err != nil {
		Logga(ctx, os.Getenv("JsonLog"), err.Error(), "error")
	}
}
