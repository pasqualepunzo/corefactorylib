package lib

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/antelman107/net-wait-go/wait"
	amqp "github.com/rabbitmq/amqp091-go"
)

func FailOnError(ctx context.Context, err error, msg string) {
	if err != nil {
		Logga(ctx, os.Getenv("JsonLog"), msg+" "+err.Error(), "error")

		destEmail := []string{
			"platform@q01.io",
		}

		hostname, _ := os.Hostname()
		message := hostname + " " + os.Getenv("tenant") + " " + os.Getenv("hostMQ") + " - " + err.Error()
		SyncSendMail("Sync - ERROR", message, destEmail, "", nil)
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
	msg.ReferenceTenant = body[0].XOUTBOX15
	msg.Resource = body[0].XOUTBOX13
	msg.Env = body[0].XOUTBOX14
	msg.Action = body[0].XOUTBOX05
	return msg, nil
}
func NetAlive(host, port string) bool {
	if !wait.New(
		wait.WithProto("tcp"),
		wait.WithWait(200*time.Millisecond),
		wait.WithBreak(50*time.Millisecond),
		wait.WithDeadline(25*time.Second),
		wait.WithDebug(true),
	).Do([]string{host + ":" + port}) {
		return false
	} else {
		return true
	}
}
func PublishQueueError(ctx context.Context, d amqp.Delivery, msgDtl MsgDetails, exchange, env string) {

	conn, err := amqp.Dial("amqp://" + os.Getenv("userMQ") + ":" + os.Getenv("passwdMQ") + "@" + os.Getenv("hostMQ") + ":" + os.Getenv("portMQ") + "/" + env)
	FailOnError(ctx, err, "Failed to connect to rabbitmq")
	defer conn.Close()

	ch, err := conn.Channel()
	FailOnError(ctx, err, "Failed to open a channel")
	defer ch.Close()

	err = ch.ExchangeDeclare(
		env+"-"+exchange, // name
		"fanout",         // type
		true,             // durable
		false,            // auto-deleted
		false,            // internal
		false,            // no-wait
		nil,              // arguments
	)
	FailOnError(ctx, err, "Failed to declare an exchange")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	hd := make(map[string]interface{})
	hd["Slave"] = msgDtl.Microservice

	errP := ch.PublishWithContext(ctx,
		exchange, // exchange
		"",       // routing key
		false,    // mandatory
		false,    // immediate
		amqp.Publishing{
			Headers:     hd,
			ContentType: "text/plain",
			Body:        []byte(string(d.Body)),
		})
	if errP != nil {
		Logga(ctx, os.Getenv("JsonLog"), " * ERROR Pubblish a message on "+env+"-"+exchange, "error")
	}
	Logga(ctx, os.Getenv("JsonLog"), " * Sent a message on exchange: "+env+"-"+exchange, "warn")
}
