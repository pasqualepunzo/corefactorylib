package lib

import (
	"context"
	"crypto/sha1"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

type RestyClientLogger struct{}

func (cliLogger *RestyClientLogger) Debugf(format string, v ...interface{}) {
	for _, m := range v {
		ctx := context.Background()
		Logga(ctx, m, "info")
	}
}

func (cliLogger *RestyClientLogger) Warnf(format string, v ...interface{}) {
	for _, m := range v {
		ctx := context.Background()
		Logga(ctx, m, "warn")
	}
}

func (cliLogger *RestyClientLogger) Errorf(format string, v ...interface{}) {
	for _, m := range v {
		ctx := context.Background()
		Logga(ctx, m, "error")
	}
}

func ApiCallPOST(ctx context.Context, debug bool, args []map[string]interface{}, microservice, routing, token, dominio string) CallGetResponse {

	Logga(ctx, "apiCallPOST")

	type restyPOSTStruct []struct {
		Code   int         `json:"code"`
		Errors interface{} `json:"errors"`
		Debug  struct {
			SQL string `json:"sql"`
		} `json:"debug"`
		InsertedID string `json:"insertedId"`
		RowCount   int    `json:"rowCount"`
	}

	if dominio == "" {
		dominio = GetApiHost()
	} else {
		dominio = "https://" + dominio
	}

	var resStruct CallGetResponse

	Logga(ctx, dominio+"/api/"+os.Getenv("coreapiVersion")+routing+" - "+microservice)

	var LoggaErrore LoggaErrore
	LoggaErrore.Errore = 0

	client := resty.New()

	// Set logrus as Logger
	restyLogger := RestyClientLogger{}
	client.SetLogger(&restyLogger)

	client.Debug = true
	// Set retry count to non zero to enable retries
	client.SetRetryCount(2)
	// You can override initial retry wait time.
	// Default is 100 milliseconds.
	client.SetRetryWaitTime(1 * time.Second)
	// MaxWaitTime can be overridden as well.
	// Default is 2 seconds.
	client.AddRetryCondition(
		// RetryConditionFunc type is for retry condition function
		// input: non-nil Response OR request execution error
		func(r *resty.Response, err error) bool {

			acceptedStatus := make(map[int]bool)
			acceptedStatus[200] = true //BadRequest
			acceptedStatus[201] = true //
			acceptedStatus[202] = true
			acceptedStatus[203] = true
			acceptedStatus[204] = true
			acceptedStatus[401] = true

			if ok := acceptedStatus[r.StatusCode()]; ok {
				return false
			} else {
				return true // ho ricevuto uno status che non è tra quelli accepted, quindi faccio retry
			}

		},
	)

	client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	res, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		//SetHeader("canary-mode", "on").
		SetHeader("microservice", microservice).
		SetAuthToken(token).
		SetBody(args).
		Post(dominio + "/api/" + os.Getenv("coreapiVersion") + routing)

	// fmt.Println(res)
	// LogJson(res)

	if err != nil { // HTTP ERRORE
		resStruct.Errore = -1
		resStruct.Log = err.Error()
	} else {
		if res.StatusCode() != 201 {
			resStruct.Errore = -1
			resStruct.Log = "Cannot update the record"

		} else {

			var restyPOSTRes restyPOSTStruct
			errJson := json.Unmarshal(res.Body(), &restyPOSTRes)
			callResponse := map[string]interface{}{}
			if errJson != nil {
				resStruct.Errore = -1
				resStruct.Log = errJson.Error()

			} else {
				callResponse["InsertID"] = restyPOSTRes[0].InsertedID
				resStruct.Kind = "Json"
				resStruct.BodyJson = callResponse
			}
		}
	}

	return resStruct
}
func ApiCallGET(ctx context.Context, debug bool, args map[string]string, microservice, routing, token, dominio string) CallGetResponse {

	Logga(ctx, "apiCallGET")

	JobID := ""
	if ctx.Value("JobID") != nil {
		JobID = ctx.Value("JobID").(string)
	}

	args["JobID"] = JobID

	type restyStruct struct {
		Data    string      `json:"data"`
		Errors  interface{} `json:"errors"`
		Message struct {
			Data    string        `json:"data"`
			Errors  []interface{} `json:"errors"`
			Message string        `json:"message"`
			Status  int           `json:"status"`
		} `json:"message"`
		Status int `json:"status"`
	}

	type restyError struct {
		Status  int32  `json:"status"`
		Message string `json:"message"`
	}

	if dominio == "" {
		dominio = GetApiHost()
	} else {
		dominio = "https://" + dominio
	}

	Logga(ctx, dominio+"/api/"+os.Getenv("coreapiVersion")+routing+" - "+microservice)

	var resStruct CallGetResponse

	client := resty.New()

	// Set logrus as Logger
	restyLogger := RestyClientLogger{}
	client.SetLogger(&restyLogger)

	client.Debug = true

	// client.Debug = true
	client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	// Set retry count to non zero to enable retries
	client.SetRetryCount(2)
	// You can override initial retry wait time.
	// Default is 100 milliseconds.
	client.SetRetryWaitTime(1 * time.Second)
	// MaxWaitTime can be overridden as well.
	// Default is 2 seconds.
	client.AddRetryCondition(
		// RetryConditionFunc type is for retry condition function
		// input: non-nil Response OR request execution error
		func(r *resty.Response, err error) bool {

			acceptedStatus := make(map[int]bool)
			acceptedStatus[200] = true //BadRequest
			acceptedStatus[201] = true //
			acceptedStatus[202] = true
			acceptedStatus[203] = true
			acceptedStatus[204] = true
			acceptedStatus[401] = true

			if ok := acceptedStatus[r.StatusCode()]; ok {
				return false
			} else {
				return true // ho ricevuto uno status che non è tra quelli accepted, quindi faccio retry
			}

		},
	)
	res, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		//SetHeader("canary-mode", "on").
		SetHeader("microservice", microservice).
		SetAuthToken(token).
		SetQueryParams(args).
		Get(dominio + "/api/" + os.Getenv("coreapiVersion") + routing)

	if err != nil { // HTTP ERRORE
		resStruct.Errore = -1
		resStruct.Log = "0 " + err.Error()
	} else {

		// se status ERROR
		if res.StatusCode() != 200 && res.StatusCode() != 201 && res.StatusCode() != 204 {
			var restyErr restyError
			errJson := json.Unmarshal(res.Body(), &restyErr)
			if errJson != nil {
				resStruct.Errore = -1
				resStruct.Log = errJson.Error()
			} else {

				resStruct.Errore = -2
				resStruct.Log = strconv.Itoa(res.StatusCode()) + " - " + restyErr.Message

			}
		} else {
			switch res.Body()[0] {
			case '{':
				callResponse := map[string]interface{}{}
				err1 := json.Unmarshal(res.Body(), &callResponse)
				if err1 != nil {
					resStruct.Errore = -1
					resStruct.Log = err1.Error()
				} else {
					resStruct.Kind = "Json"
					resStruct.BodyJson = callResponse
				}
			case '[':
				callResponse := []map[string]interface{}{}
				err1 := json.Unmarshal(res.Body(), &callResponse)
				if err1 != nil {
					resStruct.Errore = -1
					resStruct.Log = err1.Error()
				} else {
					resStruct.Kind = "Array"
					resStruct.BodyArray = callResponse
				}
			}
		}
	}
	//LogJson(resStruct)
	return resStruct
}
func GetApiHost() string {
	// devopsProfile, _ := os.LookupEnv("APP_ENV")
	// urlDevops := ""
	// if devopsProfile == "prod" {
	urlDevops := os.Getenv("cfDomain")
	// fmt.Println("dominoi..." + urlDevops)
	// os.Exit(0)
	// } else {
	// urlDevops = os.Getenv("cfDomainDev")
	// }

	return "https://" + urlDevops
}
func ApiCallLOGIN(ctx context.Context, debug bool, args map[string]interface{}, microservice, routing, dominio string) (map[string]interface{}, LoggaErrore) {

	//debug = true

	if dominio == "" {
		dominio = GetApiHost()
	} else {
		dominio = "https://" + dominio
	}

	JobID := ""
	if ctx.Value("JobID") != nil {
		JobID = ctx.Value("JobID").(string)
	}

	args["JobID"] = JobID

	// dominio := getApiHost()

	rand.Seed(time.Now().UnixNano())
	rnd := strconv.Itoa(rand.Intn(1000000))

	args["uuid"] = args["uuid"].(string) + "-" + rnd

	Logga(ctx, "")
	Logga(ctx, "apiCallLOGIN")
	Logga(ctx, "Args : ")
	jsonString, _ := json.Marshal(args)
	Logga(ctx, string(jsonString))

	Logga(ctx, "Microservice : "+microservice)
	Logga(ctx, "Url : "+dominio+"/api/"+os.Getenv("coreapiVersion")+routing)

	var LoggaErrore LoggaErrore
	LoggaErrore.Errore = 0

	callResponse := map[string]interface{}{}

	client := resty.New()

	// Set logrus as Logger
	restyLogger := RestyClientLogger{}
	client.SetLogger(&restyLogger)

	//client.Debug = debug
	client.Debug = true
	client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	res, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		//SetHeader("canary-mode", "on").
		SetHeader("microservice", microservice).
		SetBody(args).
		Post(dominio + "/api/" + os.Getenv("coreapiVersion") + routing)

	if err != nil { // HTTP ERRORE
		LoggaErrore.Errore = -1
		LoggaErrore.Log = err.Error()
	} else {

		if res.StatusCode() != 200 {
			LoggaErrore.Errore = -1
			LoggaErrore.Log = "Login failed - Access Denied"

		} else {

			err1 := json.Unmarshal(res.Body(), &callResponse)
			if err1 != nil {
				LoggaErrore.Errore = -1
				LoggaErrore.Log = err1.Error()
			}
		}

	}
	return callResponse, LoggaErrore
}
func ApiCallPUT(ctx context.Context, debug bool, args map[string]interface{}, microservice, routing, token, dominio string) ([]byte, LoggaErrore) {

	if dominio == "" {
		dominio = GetApiHost()
	} else {
		dominio = "https://" + dominio
	}

	var LoggaErrore LoggaErrore
	LoggaErrore.Errore = 0

	client := resty.New()
	// Set logrus as Logger
	restyLogger := RestyClientLogger{}
	client.SetLogger(&restyLogger)
	client.Debug = true
	client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	res, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		//SetHeader("canary-mode", "on").
		SetHeader("microservice", microservice).
		SetAuthToken(token).
		SetBody(args).
		Put(dominio + "/api/" + os.Getenv("coreapiVersion") + routing)

	if res.StatusCode() != 200 {
		LoggaErrore.Errore = -1
		LoggaErrore.Log = "Token error Cannot get a valid token"

	}
	if err != nil {

	}
	return res.Body(), LoggaErrore
}
func GetCoreFactoryToken(ctx context.Context) (string, LoggaErrore) {
	/* ************************************************************************************************ */
	// cerco il token di devops

	Logga(ctx, "Core factory Token")

	var erro LoggaErrore
	erro.Errore = 0

	urlDevops := GetApiHost()
	urlDevopsStripped := strings.Replace(urlDevops, "https://", "", -1)

	ct := time.Now()
	now := ct.Format("20060102150405")
	h := sha1.New()
	h.Write([]byte(now))
	sha := hex.EncodeToString(h.Sum(nil))

	argsAuth := make(map[string]interface{})
	argsAuth["access_token"] = os.Getenv("cfToken")
	argsAuth["refappCustomer"] = os.Getenv("refappCustomer")
	argsAuth["resource"] = urlDevopsStripped
	argsAuth["uuid"] = "devops-" + sha

	restyAuthResponse, restyAuthErr := ApiCallLOGIN(ctx, false, argsAuth, "msauth", "/auth/login", "")
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

	return "", erro
}
func ApiCallDELETE(ctx context.Context, debug bool, args map[string]string, microservice, routing, token, dominio string) CallGetResponse {

	if dominio == "" {
		dominio = GetApiHost()
	} else {
		dominio = "https://" + dominio
	}

	JobID := ""
	if ctx.Value("JobID") != nil {
		JobID = ctx.Value("JobID").(string)
	}

	args["JobID"] = JobID

	type restyPOSTStruct struct {
		Code   int         `json:"code"`
		Errors interface{} `json:"errors"`
	}

	var resStruct CallGetResponse

	Logga(ctx, dominio+"/api/"+os.Getenv("coreapiVersion")+routing+" - "+microservice)

	//fmt.Println("apiCallDELETE", debug)
	client := resty.New()
	// Set logrus as Logger
	restyLogger := RestyClientLogger{}
	client.SetLogger(&restyLogger)
	client.Debug = true
	client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	res, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		//SetHeader("canary-mode", "on").
		SetHeader("microservice", microservice).
		SetAuthToken(token).
		SetQueryParams(args).
		Delete(dominio + "/api/" + os.Getenv("coreapiVersion") + routing)

	if err != nil { // HTTP ERRORE
		resStruct.Errore = -1
		resStruct.Log = err.Error()
	} else {
		if res.StatusCode() != 205 {
			resStruct.Errore = -1
			resStruct.Log = "Cannot delete the record"

		} else {

			var restyPOSTRes restyPOSTStruct
			errJson := json.Unmarshal(res.Body(), &restyPOSTRes)
			callResponse := map[string]interface{}{}
			if errJson != nil {
				resStruct.Errore = -1
				resStruct.Log = errJson.Error()

			} else {
				callResponse["code"] = restyPOSTRes.Code
				resStruct.Kind = "Json"
				resStruct.BodyJson = callResponse
			}
		}
	}

	return resStruct
}
