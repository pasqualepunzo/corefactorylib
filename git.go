package lib

import (
	"log"
)

func GitClone(dir, repo string) {
	comando := "cd " + dir + " && git clone " + repo + " ."
	ExecCommand(comando, true)
}

func GitAdd(dir string) {
	comando := "cd " + dir + " && git add . "
	ExecCommand(comando, true)
}

func GitNewbranch(dir, tag string) {
	comando := "cd " + dir + " && git  checkout -q -b " + tag
	ExecCommand(comando, true)
}

func GitCheckout(dir, tag string) {
	comando := "cd " + dir + " && git checkout -q " + tag
	ExecCommand(comando, true)
}

func GitCommit(dir, message string) {
	comando := "cd " + dir + " && git commit -m '" + message + "'"
	ExecCommand(comando, true)
}

func GitPush(dir, branch string) {
	log.Println(branch)
	comando := "cd " + dir + "; git push -u origin '" + branch + "'"
	ExecCommand(comando, true)
}
