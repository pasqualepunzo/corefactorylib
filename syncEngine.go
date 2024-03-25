package lib

import (
	"context"
	"encoding/json"
	"os"
)

func FailOnError(ctx context.Context, err error, msg string) {
	if err != nil {
		Logga(ctx, os.Getenv("JsonLog"), msg+" "+err.Error(), "error")

		destEmail := []string{
			"p@lanificiodigitale.com",
			"f@lanificiodigitale.com",
		}
		SyncSendMail("Sync - ERROR", msg+" - "+err.Error(), destEmail, "", nil)
	}
}
func GetConfigFile() ([]ConfigMPQ, error) {

	cfgFile, errR := os.ReadFile(".cfg")
	if errR != nil {
		return nil, errR
	}

	var cfgs []ConfigMPQ
	err := json.Unmarshal([]byte(cfgFile), &cfgs)
	if err != nil {
		return cfgs, err
	}
	return cfgs, nil
}
func SyncSendMail(subject, msg string, receiver []string, attachName string, attach []byte) error {

	// sendmail
	var mailer Mailer
	mailer.Sender = os.Getenv("smtpSender")
	mailer.SmtpHost = os.Getenv("smtpHost")
	mailer.Passwd = os.Getenv("smtpPass")
	mailer.Port = os.Getenv("smtpPort")
	mailer.Receivers = receiver
	mailer.Attach = attach
	mailer.AttachName = attachName
	mailer.Msg = msg
	mailer.Subject = subject

	erro := SendMail(mailer)
	return erro
}
func GetMsgID(Body []byte) (MsgDetails, error) {

	var body []OutboxBody
	var msg MsgDetails
	errB := json.Unmarshal(Body, &body)
	if errB != nil {
		return msg, errB
	}

	msg.UniquID = body[0].XOUTBOX03
	msg.Dim = body[0].XOUTBOX04
	msg.Microservice = body[0].XOUTBOX11
	msg.Tenant = body[0].XOUTBOX12
	msg.Resource = body[0].XOUTBOX13
	msg.Action = body[0].XOUTBOX05
	return msg, nil
}
