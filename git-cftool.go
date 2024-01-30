package lib

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/briandowns/spinner"
)

func CloneRepo(loginRes LoginRes, dirRepo, newBranchName, actionGit string, repo RepoListStruct, swmono bool) {
	err := os.RemoveAll(dirRepo)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println(" + Creating: " + repo.Repo + " in " + repo.Nome)
	err = os.MkdirAll(dirRepo, 0755)

	fmt.Println(" + Cloning: " + repo.Repo)
	GitCloneCfTool(loginRes, dirRepo, repo.Repo, actionGit)
	fmt.Println(" + Checking out: " + repo.Repo + " to " + newBranchName)
	GitCheckout(dirRepo, newBranchName)

	if swmono {
		writeStIgnore(dirRepo)
	}
}
func writeStIgnore(dirRepo string) {

	ignore := ".git\n"
	ignore += "**/p2rSocket\n"
	ignore += "**/node_modules\n"
	ignore += "**/dist\n"
	ignore += "**/vendor\n"
	ignore += "**/svn.simple\n"
	ignore += "**/.git\n"
	//ignore += "**/*.gif\n"
	//ignore += "**/*.png\n"
	//ignore += "**/*.jpeg\n"
	//ignore += "**/*.jpg\n"
	ignore += "**/*.pdf\n"
	//ignore += "**/*.svg\n"
	ignore += "**/*.bat\n"
	ignore += "**/*.exe\n"
	ignore += "**/*.sh\n"
	//ignore += "**/*.eot\n"
	//ignore += "**/*.ttf\n"
	//ignore += "**/*.woff\n"
	ignore += "**/*.zip\n"
	ignore += "**/*.tar.*\n"
	ignore += "package/InStoreClient/scripts/wkhtmltopdf-i386\n"

	// fmt.Println("-------------------------------------------------------------------------")
	// fmt.Println()
	//fmt.Println(dirRepo + string(os.PathSeparator) + ".stignore")
	file, err := os.Create(dirRepo + string(os.PathSeparator) + ".stignore")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	defer file.Close()

	_, err = io.WriteString(file, ignore)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func GitCloneCfTool(loginRes LoginRes, dir, repo, actionBitbucket string) {

	//fmt.Println(dir)
	var cmd *exec.Cmd

	s := spinner.New(spinner.CharSets[36], 100*time.Millisecond)
	s.Start() // Start the spinner
	s.Color("green", "bold")
	s.Prefix = "Waiting: "

	//fmt.Print("Cloning " + repo + " ")
	var comando, aaa string

	if runtime.GOOS == "windows" {
		if actionBitbucket == "ssh" {
			aaa = "cd " + dir + " && git config --global core.autocrlf false && " + " git clone git@" + loginRes.UrlGit + ":" + loginRes.ProjectGit + "/" + repo + " . "
			//fmt.Println(aaa)

			cmd = exec.Command("cmd", "/C", aaa)
		} else {

			switch loginRes.TypeGit {
			case "bitbucket", "":
				aaa = "cd " + dir + " && git config --global core.autocrlf false && " + " git clone https://" + loginRes.UserGit + ":" + loginRes.TokenGit + "@" + loginRes.UrlGit + "/" + loginRes.ProjectGit + "/" + repo + " . "
			case "github":
				aaa = "cd " + dir + " && git config --global core.autocrlf false && " + " git clone https://" + loginRes.UserGit + ":" + loginRes.TokenGit + "@" + loginRes.UrlGit + "/" + loginRes.UserGit + "/" + repo + " . "
			}

			//fmt.Println(aaa)
			cmd = exec.Command("cmd", "/C", aaa)
		}
	} else {
		if actionBitbucket == "ssh" {
			comando = "cd " + dir + " && git clone git@" + loginRes.UrlGit + ":" + loginRes.ProjectGit + "/" + repo + " . "
		} else {
			// github                    git clone https://github.com/pasqualepunzo/btcpayserver.git . git clone https://username:password@github.com/username/repository.git
			switch loginRes.TypeGit {
			case "bitbucket", "":
				comando = "cd " + dir + " && git clone https://" + loginRes.UserGit + ":" + loginRes.TokenGit + "@" + loginRes.UrlGit + "/" + loginRes.ProjectGit + "/" + repo + " . "
			case "github":
				comando = "cd " + dir + " && git clone https://" + loginRes.UserGit + ":" + loginRes.TokenGit + "@" + loginRes.UrlGit + "/" + loginRes.UserGit + "/" + repo + " . "
			}
		}
		//fmt.Println(comando)
		cmd = exec.Command("bash", "-c", comando)
	}

	//fmt.Println(comando)

	err := cmd.Run()
	if err != nil {
		s.Stop()
		fmt.Println("Error in clonig "+repo, err.Error())
		os.Exit(1)
	}
	s.Stop()
	//fmt.Println(" " + taskDone)
}
func GitCheckoutCfTool(dir, branch string) {

	// fmt.Println("la dir per la checkout e:" + dir)
	// fmt.Println("il banch: " + branch)
	//fmt.Println()

	var cmd *exec.Cmd

	comando := "cd " + dir + " &&  git checkout " + branch

	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", comando)
	} else {
		cmd = exec.Command("bash", "-c", comando)
	}

	//	fmt.Println(cmd)
	out, err := cmd.CombinedOutput()
	fmt.Println(" + " + string(out))

	if err != nil {
		fmt.Println(err)
		fmt.Println(taskError)
	}
}
func GitStash(dir string) {
	err := os.Chdir(dir)
	if err != nil {
		fmt.Println(err)
	}
	var cmd *exec.Cmd

	comando := " git stash --include-untracked "

	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", comando)
	} else {
		cmd = exec.Command("bash", "-c", comando)
	}
	out, err := cmd.CombinedOutput()
	fmt.Println(" + " + string(out))
	//os.Exit(0)
	// if err != nil {
	// 	fmt.Println(err)
	// 	fmt.Println(taskError)
	// }
}
func GitStashApply(dir string) {
	err := os.Chdir(dir)
	if err != nil {
		fmt.Println(err)
	}
	var cmd *exec.Cmd

	comando := " git stash apply "

	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", comando)
	} else {
		cmd = exec.Command("bash", "-c", comando)
	}
	out, err := cmd.CombinedOutput()
	fmt.Println(" + " + string(out))
	//os.Exit(0)
	// if err != nil {
	// 	fmt.Println(err)
	// 	fmt.Println(taskError)
	// }
}
func GitMerge(dir, branchToMerge string) {
	err := os.Chdir(dir)
	if err != nil {
		fmt.Println(err)
	}
	var cmd *exec.Cmd

	comando := " git merge origin/" + branchToMerge
	//fmt.Println(dir, comando)

	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", comando)
	} else {
		cmd = exec.Command("bash", "-c", comando)
	}
	out, err := cmd.CombinedOutput()
	fmt.Println(" + " + string(out))

	if err != nil {
		fmt.Println(err)
		fmt.Println(taskError)
		os.Exit(0)
	}
}
func GitFetch(dir string) {

	err := os.Chdir(dir)
	if err != nil {
		fmt.Println(err)
	}
	var cmd *exec.Cmd

	comando := " git fetch "

	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", comando)
	} else {
		cmd = exec.Command("bash", "-c", comando)
	}
	_, err = cmd.CombinedOutput()

	if err != nil {
		fmt.Println(err)
		fmt.Println(taskError)
	}
}
func GitPull(dir string) {

	err := os.Chdir(dir)
	if err != nil {
		fmt.Println(err)
	}
	var cmd *exec.Cmd

	comando := " git pull "

	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", comando)
	} else {
		cmd = exec.Command("bash", "-c", comando)
	}
	out, err := cmd.CombinedOutput()
	fmt.Println(" + " + string(out))
	//os.Exit(0)
	// if err != nil {
	// 	fmt.Println(err)
	// 	fmt.Println(taskError)
	// }
}
