package lib

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
)

func Base64decode(str string) string {

	if str != "" {
		decoded, err := base64.StdEncoding.DecodeString(str)
		if err != nil {
			fmt.Println("decode error:", err)
		}
		return string(decoded)
	} else {
		return ""
	}
}

/*
 per i comandi che inviano in streamig al cf-tool del testo
 la convenzione è: se la stringa inizia con ### vuol dire che va printato qualcosa se dopo ### abbiamo Error è un errore
 altrimenti e una comunicazione
*/
func PrintaErroreStream(errorLabel, log string, flagDie bool) {
	fmt.Println("_##ERROR##_", errorLabel, " - ", log)
	if flagDie {
		os.Exit(1)
	}
}
func PrintaErroreStreamText(errorLabel, log string) string {
	text := "_##START##_\n"
	text += "\n"
	text += "\033[1;31m *********************************************************************************\n"
	text += " *  _____\n"
	text += " * | ____|_ __ _ __ ___  _ __ \n"
	text += " * |  _| | '__| '__/ _ \\| '__|\n"
	text += " * | |___| |  | | | (_) | |   \n"
	text += " * |_____|_|  |_|  \\___/|_|  \n"
	text += " * \033[1;0m\n"
	text += " *\n"
	text += " * " + errorLabel + "\n"
	text += " *\n"
	text += " *\n"
	text += " * " + log + "\n"
	text += " *\n"
	text += " *\n"
	text += "\033[0;31m *********************************************************************************\033[1;0m\n"
	text += "\n"
	text += "_##STOP##_\n"

	return text
}

// SwitchCluster ...
func SwitchCluster(clusterName string) {

	comando := "gcloud container clusters get-credentials " + clusterName
	ExecCommand(comando, true)
}

// SwitchProject ...
func SwitchProject(clusterProject string) {

	comando := "gcloud config set project  " + clusterProject
	ExecCommand(comando, true)
}
func ExecCommand(command string, printOutput bool) bool {

	println()
	cmd := exec.Command("bash", "-c", command)
	//println(command)

	var stdoutBuf, stderrBuf bytes.Buffer
	stdoutIn, _ := cmd.StdoutPipe()
	stderrIn, _ := cmd.StderrPipe()

	var errStdout, errStderr error
	stdout := io.MultiWriter(os.Stdout, &stdoutBuf)
	stderr := io.MultiWriter(os.Stderr, &stderrBuf)
	err := cmd.Start()
	if err != nil {
		log.Fatalf("cmd.Start() failed with '%s'\n", err)
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		_, errStdout = io.Copy(stdout, stdoutIn)
		wg.Done()
	}()

	_, errStderr = io.Copy(stderr, stderrIn)
	wg.Wait()

	err = cmd.Wait()
	FataleErrore := false
	if err != nil {
		if printOutput == true {
			PrintaErroreStream("Exec command errors: "+command+" -> ", err.Error(), true)
			PrintaErrore("Exec command errors: "+command+" -> ", err.Error(), "fix errors and try again")
		}
		FataleErrore = true
	}
	if errStdout != nil || errStderr != nil {
		if printOutput == true {
			log.Fatal("failed to capture stdout or stderr\n")
		}
	}
	outStr := string(stdoutBuf.Bytes())
	if printOutput == true {
		fmt.Printf("%s", outStr)
	}
	return FataleErrore
}
func PrintaErrore(errorLabel, log, errorSuggest string) {
	fmt.Println()
	fmt.Println("*!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!*")
	fmt.Println(" _____")
	fmt.Println("| ____|_ __ _ __ ___  _ __ ")
	fmt.Println("|  _| | '__| '__/ _ \\| '__|")
	fmt.Println("| |___| |  | | | (_) | |   ")
	fmt.Println("|_____|_|  |_|  \\___/|_|  ")
	fmt.Println("*                                                                                *")
	fmt.Println("*  " + errorLabel)
	fmt.Println("*                                                                                *")
	fmt.Println("*  " + log)
	fmt.Println("*                                                                                *")
	fmt.Println("*                                                                                *")
	fmt.Println("*  " + errorSuggest)
	fmt.Println("*                                                                                *")
	fmt.Println("*!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!*")
	fmt.Println()
}
func LogJson(msg interface{}) {

	//MarshalIndent
	empJSON, _ := json.MarshalIndent(msg, "", "  ")
	fmt.Printf("%s\n", string(empJSON))
}
func Logga(i interface{}) {

	switch v := i.(type) {
	case int:
		// v is an int here, so e.g. v + 1 is possible.
		fmt.Printf("Integer: %v", v)
	case float64:
		// v is a float64 here, so e.g. v + 1.0 is possible.
		fmt.Printf("Float64: %v", v)
	case string:
		fmt.Println(i)
	default:
		var b []byte
		b, _ = json.MarshalIndent(i, "", "\t")
		text := string(b)
		fmt.Println(text)
	}

	/*
		fmt.Println(strings.Trim(text, "\""))

		pid := strconv.Itoa(os.Getpid())
		hostname, _ := os.Hostname()

		log.SetFormatter(&log.JSONFormatter{})
		log.SetOutput(os.Stdout)

		logLevel, err := log.ParseLevel(os.Getenv("LOG_LEVEL"))
		if err != nil {
			logLevel = log.InfoLevel
		}

		log.SetLevel(logLevel)

		log.WithFields(log.Fields{
			"hostname": hostname,
			"pid":      pid,
		}).Info(text)
	*/

	/*
		// open a file
		f, err := os.OpenFile("/tmp/devops.log", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
		if err != nil {
			fmt.Printf("error opening file: %v", err)
		}

		// don't forget to close it
		defer f.Close()

		// Log as JSON instead of the default ASCII formatter.
		log.SetFormatter(&log.TextFormatter{
			DisableColors:   true,
			TimestampFormat: "2006-01-02 15:04:05",
			FullTimestamp:   true,
		})

		log.SetFormatter(&log.JSONFormatter{})

		// Output to stderr instead of stdout, could also be a file.
		log.SetOutput(f)
		log.SetFormatter(&log.JSONFormatter{})
		log.WithFields(log.Fields{"pid": pid, "hostname": hostname}).Println(text)
	*/
}
