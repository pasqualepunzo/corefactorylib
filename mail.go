package lib

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"mime/quotedprintable"
	"net/smtp"

	"github.com/go-resty/resty/v2"
)

func LoginAuth(username, password string) smtp.Auth {
	return &loginAuth{username, password}
}
func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte{}, nil
}
func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch string(fromServer) {
		case "Username:":
			return []byte(a.username), nil
		case "Password:":
			return []byte(a.password), nil
		default:
			return nil, errors.New("Unkown fromServer")
		}
	}
	return nil, nil
}
func SendMail(subject, testo string, receivers []string) {
	from_email := "devops@custom.it"
	host := "relay.custom.it:465"
	//auth := smtp.PlainAuth("", from_email, password, "custom.relay.it")
	// auth := LoginAuth(from_email, password)

	header := make(map[string]string)
	var to_email []string
	if len(receivers) > 0 {
		to_email = receivers
	}

	// a prescindere mando a me e betto
	to_email = append(to_email, "e.disanto@custom.it")

	header["From"] = from_email
	header["To"] = to_email[0]
	header["Subject"] = subject

	header["MIME-Version"] = "1.0"
	header["Content-Disposition"] = "inline"
	header["Content-Transfer-Encoding"] = "quoted-printable"

	header_message := ""
	for key, value := range header {
		header_message += fmt.Sprintf("%s: %s\r\n", key, value)
	}

	body := testo
	var body_message bytes.Buffer
	temp := quotedprintable.NewWriter(&body_message)
	temp.Write([]byte(body))
	temp.Close()

	final_message := header_message + "\r\n" + body_message.String()

	status := smtp.SendMail(host, nil, from_email, to_email, []byte(final_message))
	if status != nil {
		log.Printf("Error from SMTP Server: %s", status)
	}
	log.Print("Email Sent Successfully")
}
func TelegramSendMessage(botToken, cftoolDevopsChatID, text string) LoggaErrore {
	type telegramResStruct struct {
		Ok     bool `json:"ok"`
		Result struct {
			MessageID  int `json:"message_id"`
			SenderChat struct {
				ID       int64  `json:"id"`
				Title    string `json:"title"`
				Username string `json:"username"`
				Type     string `json:"type"`
			} `json:"sender_chat"`
			Chat struct {
				ID       int64  `json:"id"`
				Title    string `json:"title"`
				Username string `json:"username"`
				Type     string `json:"type"`
			} `json:"chat"`
			Date int    `json:"date"`
			Text string `json:"text"`
		} `json:"result"`
	}

	var erro LoggaErrore
	erro.Errore = 0

	clientTelegram := resty.New()
	clientTelegram.Debug = false
	resTelegram, errTelegram := clientTelegram.R().
		SetHeader("Content-Type", "application/json").
		Post("https://api.telegram.org/bot" + botToken + "/sendMessage?chat_id=" + cftoolDevopsChatID + "&text=" + text)

	var telegramRes telegramResStruct
	if errTelegram != nil {
		erro.Errore = -1
		erro.Log = errTelegram.Error()
	} else {
		err1 := json.Unmarshal(resTelegram.Body(), &telegramRes)
		if err1 != nil {
			fmt.Println(err1.Error())
		}
	}

	// LogJson(telegramRes)

	if !telegramRes.Ok {
		erro.Errore = -1
		erro.Log = telegramRes.Result.Text
	}

	return erro
}
