package lib

import (
	"context"
	"crypto/sha1"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"errors"
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
		Logga(ctx, os.Getenv("JsonLog"), m, "info")
	}
}

func (cliLogger *RestyClientLogger) Warnf(format string, v ...interface{}) {
	for _, m := range v {
		ctx := context.Background()
		Logga(ctx, os.Getenv("JsonLog"), m, "warn")
	}
}

func (cliLogger *RestyClientLogger) Errorf(format string, v ...interface{}) {
	for _, m := range v {
		ctx := context.Background()
		Logga(ctx, os.Getenv("JsonLog"), m, "error")
	}
}

func ApiCallPOST(ctx context.Context, debug string, args []map[string]interface{}, microservice, routing, token, dominio, coreApiVersion string) CallGetResponse {

	Logga(ctx, os.Getenv("JsonLog"), "apiCallPOST")
	if !strings.Contains(dominio, "http") {
		dominio = "https://" + dominio
	}

	type restyPOSTStruct []struct {
		Code   int         `json:"code"`
		Errors interface{} `json:"errors"`
		Debug  struct {
			SQL    string      `json:"sql"`
			Fields interface{} `json:"fields"`
			Body   interface{} `json:"body"`
		} `json:"debug"`
		InsertedID string `json:"insertedId"`

		RowCount int `json:"rowCount"`
	}

	var resStruct CallGetResponse

	Logga(ctx, os.Getenv("JsonLog"), dominio+"/api/"+coreApiVersion+routing+" - "+microservice)

	var LoggaErrore LoggaErrore
	LoggaErrore.Errore = 0

	client := resty.New()

	// Set logrus as Logger
	restyLogger := RestyClientLogger{}
	client.SetLogger(&restyLogger)

	debool, errBool := strconv.ParseBool(debug)
	if errBool != nil {
		resStruct.Errore = -1
		resStruct.Log = errBool.Error()
		return resStruct
	}
	client.Debug = debool
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
		Post(dominio + "/api/" + coreApiVersion + routing)

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
				resStruct.DebugFields = restyPOSTRes[0].Debug.Fields
				resStruct.DebugBody = restyPOSTRes[0].Debug.Body
			}
		}
	}

	return resStruct
}
func ApiCallGET(ctx context.Context, debug string, args map[string]string, microservice, routing, token, dominio, coreApiVersion string) (CallGetResponse, error) {

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

	var resStruct CallGetResponse

	debool, errBool := strconv.ParseBool(debug)
	if errBool != nil {
		return resStruct, errBool
	}

	if debool {
		Logga(ctx, os.Getenv("JsonLog"), "apiCallGET")
	}
	if !strings.Contains(dominio, "http") {
		dominio = "https://" + dominio
	}

	JobID := ""
	if ctx.Value("JobID") != nil {
		JobID = ctx.Value("JobID").(string)
	}

	args["JobID"] = JobID

	if debool {
		Logga(ctx, os.Getenv("JsonLog"), dominio+"/api/"+coreApiVersion+routing+" - "+microservice)
	}

	client := resty.New()

	// Set logrus as Logger
	restyLogger := RestyClientLogger{}
	client.SetLogger(&restyLogger)

	client.Debug = debool

	// client.Debug = true
	client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	// Set retry count to non zero to enable retries
	client.SetRetryCount(1)
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
		SetHeader("microservice", microservice).
		SetAuthToken(token).
		SetQueryParams(args).
		Get(dominio + "/api/" + coreApiVersion + routing)

	if err != nil { // HTTP ERRORE
		return resStruct, err
	} else {

		switch res.StatusCode() {
		case 200:
			switch res.Body()[0] {
			case '{':
				callResponse := map[string]interface{}{}
				err1 := json.Unmarshal(res.Body(), &callResponse)
				if err1 != nil {
					return resStruct, err1
				} else {

					val, ok := callResponse["data"]
					if ok {
						if val == nil {
							erro := errors.New(dominio + "/api/" + coreApiVersion + routing + " -> ***** NO CONTENT *****")
							return resStruct, erro
						}
					}
					resStruct.Kind = "Json"
					resStruct.BodyJson = callResponse

					bodyCode, _ := resStruct.BodyJson["code"].(float64)
					if bodyCode != 0 && bodyCode != 200 {
						errString, _ := resStruct.BodyJson["error_msg"].(string)
						erro := errors.New(errString)
						return resStruct, erro
					}
				}
			case '[':
				callResponse := []map[string]interface{}{}
				err1 := json.Unmarshal(res.Body(), &callResponse)
				if err1 != nil {
					return resStruct, err1
				} else {
					resStruct.Kind = "Array"
					resStruct.BodyArray = callResponse
				}
			}
		case 400, 401, 404:

			callResponse := map[string]interface{}{}
			err1 := json.Unmarshal(res.Body(), &callResponse)
			if err1 != nil {
				return resStruct, err1
			} else {

				val, ok := callResponse["message"].(map[string]interface{})
				if ok {

					message, ok2 := val["message"].(string)
					if ok2 {
						erro := errors.New(dominio + "/api/" + coreApiVersion + routing + " -> " + message)
						return resStruct, erro
					}
				}
			}
			erro := errors.New(res.Status())
			return resStruct, erro
		case 500, 501, 502, 503, 504:
			erro := errors.New(res.Status())
			return resStruct, erro
		}

	}
	//LogJson(resStruct)
	return resStruct, nil
}

func ApiCallLOGIN(ctx context.Context, debug string, args map[string]interface{}, microservice, routing, dominio, coreApiVersion string) (map[string]interface{}, LoggaErrore) {

	if !strings.Contains(dominio, "http") {
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

	callResponse := map[string]interface{}{}

	var LoggaErrore LoggaErrore
	LoggaErrore.Errore = 0

	debool, errBool := strconv.ParseBool(debug)
	if errBool != nil {
		LoggaErrore.Errore = -1
		LoggaErrore.Log = errBool.Error()
		return callResponse, LoggaErrore
	}
	if debool {
		Logga(ctx, os.Getenv("JsonLog"), "")
		Logga(ctx, os.Getenv("JsonLog"), "apiCallLOGIN")
		Logga(ctx, os.Getenv("JsonLog"), "Args : ")
	}
	jsonString, _ := json.Marshal(args)
	if debool {
		Logga(ctx, os.Getenv("JsonLog"), string(jsonString))

		Logga(ctx, os.Getenv("JsonLog"), "Microservice : "+microservice)
		Logga(ctx, os.Getenv("JsonLog"), "Url : "+dominio+"/api/"+coreApiVersion+routing)
	}

	client := resty.New()

	// Set logrus as Logger
	restyLogger := RestyClientLogger{}
	client.SetLogger(&restyLogger)

	//client.Debug = debug
	client.Debug = debool
	client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	res, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		//SetHeader("canary-mode", "on").
		SetHeader("microservice", microservice).
		SetBody(args).
		Post(dominio + "/api/" + coreApiVersion + routing)

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
func ApiCallPUT(ctx context.Context, debug string, args map[string]interface{}, microservice, routing, token, dominio, coreApiVersion string) ([]byte, error) {

	var erro error
	if !strings.Contains(dominio, "http") {
		dominio = "https://" + dominio
	}

	debool, errBool := strconv.ParseBool(debug)
	if errBool != nil {
		return nil, errBool
	}

	client := resty.New()
	// Set logrus as Logger
	restyLogger := RestyClientLogger{}
	client.SetLogger(&restyLogger)
	client.Debug = debool
	client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	res, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		//SetHeader("canary-mode", "on").
		SetHeader("microservice", microservice).
		SetAuthToken(token).
		SetBody(args).
		Put(dominio + "/api/" + coreApiVersion + routing)

	if res.StatusCode() != 200 {
		erro = errors.New("Update error - CODE: " + strconv.Itoa(res.StatusCode()))
		return nil, erro
	}
	if err != nil {
		return nil, err
	}
	return res.Body(), erro
}
func GetCoreFactoryToken(ctx context.Context, tenant, accessToken, loginApiDomain, coreApiVersion string, debug string) (string, error) {
	/* ************************************************************************************************ */
	// cerco il token di devops
	var erro error

	debool, errBool := strconv.ParseBool(debug)
	if errBool != nil {
		return "", errBool
	}

	if debool {
		Logga(ctx, os.Getenv("JsonLog"), "Core factory Token")
	}

	urlDevops := loginApiDomain
	urlDevopsStripped := strings.Replace(urlDevops, "https://", "", -1)

	ct := time.Now()
	now := ct.Format("20060102150405")
	h := sha1.New()
	h.Write([]byte(now))
	sha := hex.EncodeToString(h.Sum(nil))

	argsAuth := make(map[string]interface{})
	argsAuth["access_token"] = accessToken
	argsAuth["refappCustomer"] = tenant
	argsAuth["resource"] = urlDevopsStripped
	argsAuth["uuid"] = "devops-" + sha

	restyAuthResponse, restyAuthErr := ApiCallLOGIN(ctx, debug, argsAuth, "msauth", "/auth/login", loginApiDomain, coreApiVersion)
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
func ApiCallDELETE(ctx context.Context, debug string, args map[string]string, microservice, routing, token, dominio, coreApiVersion string) CallGetResponse {

	JobID := ""
	if ctx.Value("JobID") != nil {
		JobID = ctx.Value("JobID").(string)
	}

	if !strings.Contains(dominio, "http") {
		dominio = "https://" + dominio
	}

	args["JobID"] = JobID

	type restyPOSTStruct struct {
		Code   int         `json:"code"`
		Errors interface{} `json:"errors"`
	}

	var resStruct CallGetResponse

	Logga(ctx, os.Getenv("JsonLog"), dominio+"/api/"+coreApiVersion+routing+" - "+microservice)

	debool, errBool := strconv.ParseBool(debug)
	if errBool != nil {
		resStruct.Errore = -1
		resStruct.Log = errBool.Error()
	}

	//fmt.Println("apiCallDELETE", debug)
	client := resty.New()
	// Set logrus as Logger
	restyLogger := RestyClientLogger{}
	client.SetLogger(&restyLogger)
	client.Debug = debool
	client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	res, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		//SetHeader("canary-mode", "on").
		SetHeader("microservice", microservice).
		SetAuthToken(token).
		SetQueryParams(args).
		Delete(dominio + "/api/" + coreApiVersion + routing)

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
