package lib

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/golang-jwt/jwt"
)

// questo è il token temporaneo dell'utente che esegue gli script
func GetGkeAccessTokenApi(writer http.ResponseWriter, request *http.Request) {

	ctx := context.Background()
	Logga(ctx, os.Getenv("JsonLog"), "getVersion", "warn")

	res := make(map[string]interface{})
	authToken64 := request.Header.Get("Auth-Token") // il token passato dal chiamante

	if authToken64 != "" {
		// devopsorc mi chiama passandomi un token che deve essere == a quello che mi calcola la
		// func GetGkeSaltedToken quindi per bucarmi devi conoscere il sale e la logica con cui si genera il token
		if GetGkeSaltedToken(os.Getenv("GkeDevopsSalt")) == string(authToken64) {

			// get JWT TOKEN
			jwt, errJWT := GetJWT(ctx)
			if errJWT != nil {
				res["err"] = true
				res["errLog"] = errJWT.Error()
				json.NewEncoder(writer).Encode(res)
			}

			// get access token
			gkeTokenRes, errGkeT := GetGkeBearerToken(ctx, jwt)
			if errGkeT != nil {
				res["err"] = true
				res["errLog"] = errGkeT.Error()
				json.NewEncoder(writer).Encode(res)
				return
			}

			// SUCCESS
			res["err"] = false
			res["errLog"] = ""
			res["token"] = gkeTokenRes
			json.NewEncoder(writer).Encode(res)

		} else {
			res["token"] = ""
			res["errLog"] = "Auth-Token not valid"
			json.NewEncoder(writer).Encode(res)
		}
	} else {
		res["token"] = ""
		res["errLog"] = "Auth-Token missing"
		json.NewEncoder(writer).Encode(res)
	}
}

/*
genera un token per poter comunicare in modo sicuro con devops
*/
func GetGkeSaltedToken(salt string) string {
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

/*
func obsoleta che sfrutta gloud per ottenere un token per le google API
attualmente si utilizza JWT e Access Token
*/
func GetGkeToken() (string, error) {
	cmd := exec.Command("bash", "-c", "gcloud config config-helper --format='value(credential.access_token)'")
	stdout, err := cmd.Output()
	if err != nil {
		return "", err
	}
	gkeToken := strings.TrimSuffix(string(stdout), "\n")
	return gkeToken, err
}
func GetJWT(ctx context.Context) (string, error) {

	now := time.Now()
	signBytes, err := ioutil.ReadFile(os.Getenv("PEM_FILE"))
	if err != nil {
		return "", err
	}

	signKey, err := jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss":   os.Getenv("SERVICE_ACCOUNT"),
		"aud":   "https://oauth2.googleapis.com/token",
		"exp":   now.Add(15 * time.Minute).Unix(),
		"iat":   now.Unix(),
		"scope": "https://www.googleapis.com/auth/devstorage.full_control https://www.googleapis.com/auth/cloud-platform",
	})
	token.Header["kid"] = os.Getenv("SERVICE_ACCOUNT_KID")
	jwtString, _ := token.SignedString(signKey)

	return jwtString, err
}
func GetGkeBearerToken(ctx context.Context, jwtString string) (string, error) {

	debool, errBool := strconv.ParseBool(os.Getenv("RestyDebug"))
	if errBool != nil {
		return "", errBool
	}

	data := "grant_type=urn%3Aietf%3Aparams%3Aoauth%3Agrant-type%3Ajwt-bearer&assertion=" + jwtString

	// lancio la BUILD
	cliB := resty.New()
	cliB.Debug = debool
	cliB.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	restyResB, errApiB := cliB.R().
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetBody(data).
		Post("https://oauth2.googleapis.com/token")

	if errApiB != nil {
		return "", errApiB
	}

	if restyResB.StatusCode() != 200 {
		erro := errors.New("Status Code: " + strconv.Itoa(restyResB.StatusCode()))
		return "", erro
	}

	var gkeT GkeToken
	errM := json.Unmarshal([]byte(restyResB.Body()), &gkeT)
	if errM != nil {
		return "", errM
	}

	if gkeT.AccessToken == "" {
		erro := errors.New("Access Token MISSING")
		return "", erro
	}

	return gkeT.AccessToken, nil
}
