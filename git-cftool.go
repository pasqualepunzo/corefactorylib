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

func CloneRepo(dirRepo, newBranchName, userBitbucket, passBitbucket, bitbucketProject, actionBitbucket string, repo RepoListStruct) {
	err := os.RemoveAll(dirRepo)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println(" + Creating: " + repo.Repo + " in " + repo.Nome)
	err = os.MkdirAll(dirRepo, 0755)

	fmt.Println(" + Cloning: " + repo.Repo)
	GitCloneCfTool(dirRepo, repo.Repo, userBitbucket, passBitbucket, bitbucketProject, actionBitbucket)
	fmt.Println("\n + Checking out: " + repo.Repo + " to " + newBranchName)
	GitCheckout(dirRepo, newBranchName)

	writeStIgnore(dirRepo)
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

func GitCloneCfTool(dir, repo, userBitbucket, passBitbucket, bitbucketProject, actionBitbucket string) {

	//fmt.Println(dir)
	var cmd *exec.Cmd

	s := spinner.New(spinner.CharSets[36], 100*time.Millisecond)
	s.Start() // Start the spinner
	s.Color("green", "bold")
	s.Prefix = "Waiting: "

	//fmt.Print("Cloning " + repo + " ")
	var comando string

	if runtime.GOOS == "windows" {
		if actionBitbucket == "ssh" {
			aaa := "cd " + dir + " && git config --global core.autocrlf false && " + " git clone git@bitbucket.org:" + bitbucketProject + "/" + repo + " . "
			//fmt.Println(aaa)

			cmd = exec.Command("cmd", "/C", aaa)
		} else {
			aaa := "cd " + dir + " && git config --global core.autocrlf false && " + " git clone https://" + userBitbucket + ":" + passBitbucket + "@bitbucket.org/" + bitbucketProject + "/" + repo + " . "
			//fmt.Println(aaa)
			cmd = exec.Command("cmd", "/C", aaa)
		}
	} else {
		if actionBitbucket == "ssh" {
			comando = "cd " + dir + " && git clone git@bitbucket.org:" + bitbucketProject + "/" + repo + " . "
		} else {
			comando = "cd " + dir + " && git clone https://" + userBitbucket + ":" + passBitbucket + "@bitbucket.org/" + bitbucketProject + "/" + repo + " . "
		}
		//fmt.Println(comando)
		cmd = exec.Command("bash", "-c", comando)
	}

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
