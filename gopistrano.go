package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/BurntSushi/toml"
)

var (
	deployConfig *Config
	remotePath   Path
	deployCh     chan *deploy
)

type Path struct {
	deployment string
	backup     string
	utils      string
}

type Config struct {
	Login   loginInfo
	Servers map[string]serverInfo
	Deploy  deployInfo
}

type loginInfo struct {
	User    string `toml:"username"`
	Pwd     string `toml:"password"`
	SShPath string `toml:"ssh_path"`
}

type serverInfo struct {
	Ip   []string
	Port string
}

type deployInfo struct {
	Repository    string
	GoProjectPath string `toml:"go_project_path"`
	Package       string `toml:"package_name"`
	App           string `toml:"app_name"`
	KeepRelease   string `toml:"keep_releases"`
	UseSudo       bool   `toml:"use_sudo"`
	WebUser       string `toml:"webserver_user"`
}

func ReadConfig(configfile string) (*Config, error) {
	_, err := os.Stat(configfile)
	if err != nil {
		log.Fatal("Config file is missing: ", configfile)
	}

	var config *Config
	if _, err := toml.DecodeFile(configfile, &config); err != nil {
		log.Fatal(err)

		return config, err
	}
	return config, nil
}

func init() {
	deployCh = make(chan *deploy)
}

func main() {
	// 모든 CPU 사용
	runtime.GOMAXPROCS(runtime.NumCPU())
	// 현재 설정값 리턴
	fmt.Println(runtime.GOMAXPROCS(0))

	configFile := flag.String("config", "", "")
	deployAction := flag.String("action", "", "")
	serverEnv := flag.String("env", "", "")

	flag.Parse()

	if *configFile == "" || *deployAction == "" || *serverEnv == "" {
		fmt.Println("Usage:", os.Args[0], "-config=file -action=[deploy|setup] -env=[development|staging|production]")
		os.Exit(1)
	}

	var err error
	deployConfig, err = ReadConfig(*configFile)

	if err != nil {
		fmt.Println("Failed to read config")
		os.Exit(1)
	}

	remotePath.deployment = deployConfig.Deploy.GoProjectPath + "/src/" + deployConfig.Deploy.Package
	remotePath.backup = "/home/" + deployConfig.Login.User + "/backup"
	remotePath.utils = "/home/" + deployConfig.Login.User + "/utils"

	// spawn worker goroutines
	// http://stackoverflow.com/questions/18405023/how-would-you-define-a-pool-of-goroutines-to-be-executed-at-once-in-golang
	var wg sync.WaitGroup
	for range deployConfig.Servers[*serverEnv].Ip {
		// Add adds delta, which may be negative, to the WaitGroup counter.
		// If the counter becomes zero, all goroutines blocked on Wait are released.
		// If the counter goes negative, Add panics.
		//
		// the calls to Add should execute before the statement creating the goroutine or other event to be waited for.
		// If a WaitGroup is reused to wait for several independent sets of events,
		// new Add calls must happen after all previous Wait calls have returned.
		wg.Add(1)
		go func() {
			for ch := range deployCh {
				switch strings.ToLower(ch.Action()) {
				case "setup":
					err = ch.Setup()
				case "deploy":
					err = ch.Run()
				default:
					fmt.Println("Invalid command!")
				}

				if err != nil {
					fmt.Println(err.Error())
				}
			}
			// Done decrements the WaitGroup counter.
			wg.Done()
		}()
	}

	// generate some tasks
	for _, ip := range deployConfig.Servers[*serverEnv].Ip {
		deploy, err := newDeploy(
			deployConfig.Login.User,
			deployConfig.Login.Pwd,
			ip,
			deployConfig.Servers[*serverEnv].Port,
			deployConfig.Login.SShPath,
			*deployAction,
		)

		if err != nil {
			fmt.Println("Failed to start: " + err.Error())
			return
		}

		deployCh <- deploy
	}
	close(deployCh)

	// Wait blocks until the WaitGroup counter is zero.
	// wait for the workers to finish
	wg.Wait()
}
