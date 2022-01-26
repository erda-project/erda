package bootstrap

import (
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	containerruntime "github.com/erda-project/erda/modules/core/monitor/diagnotor/agent/container-runtime"
)

// Run .
func Run(fn func()) {
	if os.Getenv("DIAGNOTOR_AGENT_RUN_MODE") == "program" {
		if os.Getenv("DIAGNOTOR_AGENT_DISABLE_INIT_NS") != "true" {
			err := initNamespace()
			if err != nil {
				log.Fatalf("failed to initNamespace: %s", err)
				return
			}
		}
		fn()
	} else { // DIAGNOTOR_AGENT_RUN_MODE=bootstrap
		bootstrap()
	}
}

func bootstrap() {
	targetPodUID := os.Getenv("TARGET_POD_UID")
	targetContainerID := os.Getenv("TARGET_CONTAINER_ID")
	if len(targetPodUID) <= 0 || len(targetContainerID) <= 0 {
		log.Fatalf("TARGET_POD_UID or TARGET_CONTAINER_ID is empty")
		return
	}
	targetPID, err := containerruntime.FindPidByPodContainer(targetPodUID, targetContainerID)
	if err != nil {
		log.Fatalf("failed to pid by pod + container: %s", err)
		return
	}
	runInTargetNamespace(targetPID)
}

func runInTargetNamespace(targetPID string) {
	go func() {
		args := []string{
			"--target", targetPID, "--uts", "--ipc", "--net", "--pid", // "--user", // nsenter: setns(): can't reassociate to namespace 'user': Invalid argument
		}
		args = append(args, os.Args...)
		cmd := exec.Command("nsenter", args...)
		cmd.Env = append(os.Environ(), "DIAGNOTOR_AGENT_RUN_MODE=program")
		cmd.Stdin = os.Stdin
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		err := cmd.Run()
		if err != nil {
			log.Fatalf("failed to run program with namespaces: %s", err)
			return
		}
	}()

	log.Println("wait to exit ...")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-sc
}

func initNamespace() error {
	cmd := exec.Command("mount", "-t", "proc", "/proc", "/proc")
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	listProcesses()
	return err
}

func listProcesses() {
	cmd := exec.Command("ps", "-ef")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Printf("[ERROR] failed to ps -ef: %s", err)
	}
}
