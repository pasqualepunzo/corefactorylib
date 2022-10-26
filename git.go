package lib

import (
	"fmt"
	"log"
	"os"
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
	comando := "cd " + dir + " && git checkout -b " + tag
	ExecCommand(comando, true)
}

func GitCheckout(dir, tag string) {

	comando := "cd " + dir + " && git checkout " + tag
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

func GitInitRepo(nomeRepo string) {
	comando := " curl -X POST -v -u \"laszlo72:2TvddWPjJaSdJFTqhUdD\" -H \"Content-Type: application/json\" "
	comando += " " + os.Getenv("bitbucketHost") + "/repositories/" + os.Getenv("bitbucketProject") + "/" + nomeRepo
	comando += " -d '{\"scm\": \"git\", \"is_private\": \"true\",\"project\": {\"key\": \"MSF\"}, \"name\":\"" + nomeRepo + "\"}'"
	ExecCommand(comando, true)
}

func GitCreateNewBranchApi(repo, branch string) {
	comando := "curl -X POST -is -u \"laszlo72:2TvddWPjJaSdJFTqhUdD\" -H \"Content-Type: application/json\" "
	comando += " " + os.Getenv("bitbucketHost") + "/repositories/" + os.Getenv("bitbucketProject") + "/" + repo + "/refs/branches "
	comando += " -d '{ \"name\": \"" + branch + "\", \"target\": { \"hash\": \"master\" } }'"
	ExecCommand(comando, true)
}

func GitInit(dir, nomeRepo, GitSrcTipo, Namespace string) {

	// // troppo a majale
	// fmt.Println()
	// fmt.Println()
	// fmt.Println()
	// fmt.Println()
	// fmt.Println()
	// fmt.Println()
	// fmt.Println()
	// fmt.Println("sudo chown laszlo:laszlo " + dir + " -R")
	// fmt.Println()
	// fmt.Println("sudo chmod 777 " + dir + " -R")
	// fmt.Println()
	// fmt.Println()
	// fmt.Println()
	// fmt.Println()
	// fmt.Println()
	// fmt.Println()
	// fmt.Println()

	// comando := "sudo chown laszlo:laszlo " + dir + " -R"
	// ExecCommand(comando)
	// comando = "sudo chmod 777 " + dir + " -R"
	// ExecCommand(comando)

	// git init
	comando := "cd " + dir + "; git init "
	ExecCommand(comando, true)

	// scrivo il README.md
	f, err := os.Create(dir + "/README.md")
	if err != nil {
		fmt.Println(err)
		return
	}
	l, err := f.WriteString("##README##\n")
	if err != nil {
		fmt.Println(err)
		f.Close()
		return
	}
	fmt.Println(l, "bytes written successfully")
	err = f.Close()
	if err != nil {
		fmt.Println(err)
		return
	}

	err = os.Chdir(dir)
	if err != nil {
		panic(err)
	}

	comando = "git config --global user.email \"p.punzo@custom.it\""
	ExecCommand(comando, true)
	comando = "git config --global user.name \"devops-operator\""
	ExecCommand(comando, true)

	// git add .
	comando = "cd " + dir + "; git add ."
	ExecCommand(comando, true)

	// git commit .
	comando = "cd " + dir + "; git commit -m 'first commit'"
	ExecCommand(comando, true)

	// punto a bitbucket
	comando = "cd " + dir + "; git remote add origin https://" + os.Getenv("bitbucketUser") + ":" + os.Getenv("bitbucketToken") + "@bitbucket.org/" + os.Getenv("bitbucketProject") + "/" + nomeRepo
	ExecCommand(comando, true)

	// // pull senno si incazza al secondo giro
	// comando = "cd " + dir + "; git pull origin master --allow-unrelated-histories"
	// ExecCommand(comando)

	// push in remoto
	comando = "cd " + dir + "; git push -u origin master"
	ExecCommand(comando, true)
}
