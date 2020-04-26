package main

//go:generate errorgen

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
  //"regexp"

	"v2ray.com/core"
  //"v2ray.com/core/common/serial"
	"v2ray.com/core/common/cmdarg"
	"v2ray.com/core/common/platform"
	_ "v2ray.com/core/main/distro/all"
  socks5 "github.com/armon/go-socks5"
)

var (
	configFiles cmdarg.Arg // "Config file for V2Ray.", the option is customed type, parse in main
	configDir   string
	version     = flag.Bool("version", false, "Show current version of V2Ray.")
	test        = flag.Bool("test", false, "Test config file only, without launching V2Ray server.")
	format      = flag.String("format", "json", "Format of input file.")

	/*  We have to do this here because Golang's Test will also need to parse flag, before
		main func in this file is run.
	*/
	_ = func() error {

		flag.Var(&configFiles, "config", "Config file for V2Ray. Multiple assign is accepted (only json). Latter ones overrides the former ones.")
		flag.Var(&configFiles, "c", "Short alias of -config")
		flag.StringVar(&configDir, "confdir", "", "A dir with multiple json config")

		return nil
	}()
)

func fileExists(file string) bool {
	info, err := os.Stat(file)
	return err == nil && !info.IsDir()
}

func dirExists(file string) bool {
	if file == "" {
		return false
	}
	info, err := os.Stat(file)
	return err == nil && info.IsDir()
}

func readConfDir(dirPath string) {
	confs, err := ioutil.ReadDir(dirPath)
	if err != nil {
		log.Fatalln(err)
	}
	for _, f := range confs {
		if strings.HasSuffix(f.Name(), ".json") {
			configFiles.Set(path.Join(dirPath, f.Name()))
		}
	}
}

func getConfigFilePath() (cmdarg.Arg, error) {
	if dirExists(configDir) {
		log.Println("Using confdir from arg:", configDir)
		readConfDir(configDir)
	} else {
		if envConfDir := platform.GetConfDirPath(); dirExists(envConfDir) {
			log.Println("Using confdir from env:", envConfDir)
			readConfDir(envConfDir)
		}
	}

	if len(configFiles) > 0 {
		return configFiles, nil
	}

	if workingDir, err := os.Getwd(); err == nil {
		configFile := filepath.Join(workingDir, "config.json")
		if fileExists(configFile) {
			log.Println("Using default config: ", configFile)
			return cmdarg.Arg{configFile}, nil
		}
	}

	if configFile := platform.GetConfigurationPath(); fileExists(configFile) {
		log.Println("Using config from env: ", configFile)
		return cmdarg.Arg{configFile}, nil
	}

	log.Println("Using config from STDIN")
	return cmdarg.Arg{"stdin:"}, nil
}

func GetConfigFormat() string {
	switch strings.ToLower(*format) {
	case "pb", "protobuf":
		return "protobuf"
	default:
		return "json"
	}
}

func startV2Ray(config *core.Config) (core.Server, error) {
	server, err := core.New(config)
	if err != nil {
		return nil, newError("failed to create server").Base(err)
	}

	return server, nil
}

func printVersion() {
	version := core.VersionStatement()
	for _, s := range version {
		fmt.Println(s)
	}
}

func handleHybridConfig() ([]*core.Config, error) {
	configFiles, err := getConfigFilePath()
	if err != nil {
		return nil, err
	}

	config, err := core.LoadConfig(GetConfigFormat(), configFiles[0], configFiles)
	if err != nil {
		return nil, newError("failed to read config files: [", configFiles.String(), "]").Base(err)
	}
  //pinstance, _ := config.Inbound[0].ProxySettings.GetInstance()
  //rinstance, _ := config.Inbound[0].ReceiverSettings.GetInstance()
  //newDebugMsg("Main: Config: ProxySettings " + StructString(pinstance))
  //newDebugMsg("Main: Config: ReceiverSettings " + StructString(rinstance))

	ret := make([]*core.Config, 0, 3)
	if len(config.Inbound) == 3 { // hybrid
		_config1 := *config
		config1 := &_config1
		config1.Inbound = config.Inbound[:1]
		config1.Outbound = config.Outbound[:1]
		_config2 := *config
		config2 := &_config2
		config2.Inbound = config.Inbound[1:2]
		config2.Outbound = config.Outbound[1:2]
    _config3 := *config
    config3 := &_config3
    config3.Inbound = config.Inbound[2:3]
    config3.Outbound = config.Outbound[2:3]
		ret = append(ret, config1)
		ret = append(ret, config2)
    ret = append(ret, config3)
	} else { // v2ray default
		ret = append(ret, config)
	}
	return ret, nil
}

func startV2RayWrapper(config *core.Config) (core.Server) {
	server, err := startV2Ray(config)
	if err != nil {
		fmt.Println(err)
		// Configuration error. Exit with a special value to prevent systemd from restarting.
		os.Exit(23)
	}

	if *test {
		fmt.Println("Configuration OK.")
		os.Exit(0)
	}

	if err := server.Start(); err != nil {
		fmt.Println("Failed to start", err)
		os.Exit(-1)
	}
	return server
}

func runSOCKS5Server(addr string) {
  conf := &socks5.Config{}
  server, err := socks5.New(conf)
  if err != nil {
    panic(err)
  }

  if err := server.ListenAndServe("tcp", addr); err != nil {
    panic(err)
  }
}

func main() {

	flag.Parse()

	printVersion()

	if *version {
		return
	}
	configs, err := handleHybridConfig()
  if err != nil {
    panic("Load config failed")
  }

  if len(configs) == 1 {
    server := startV2RayWrapper(configs[0])
    defer server.Close()
  } else if len(configs) == 3 {

    controlServer := startV2RayWrapper(configs[0])
    freeServer := startV2RayWrapper(configs[1])
    relayServer := startV2RayWrapper(configs[2])
    //re := regexp.MustCompile(`From:([0-9]+)`)
    //config1, _ := configs[0].Inbound[0].ReceiverSettings.GetInstance()
    //port1 := string(re.FindSubmatch([]byte(config1.String()))[1])
    //config2, _ := configs[1].Inbound[0].ReceiverSettings.GetInstance()
    //port2 := string(re.FindSubmatch([]byte(config2.String()))[1])
    //config3, _ := configs[2].Inbound[0].ReceiverSettings.GetInstance()
    //port3 := string(re.FindSubmatch([]byte(config3.String()))[1])
    //newDebugMsg("Main: port " + port1 + ", " + port2 + ", " + port3)

    defer controlServer.Close()
    defer freeServer.Close()
    defer relayServer.Close()

  }

	// Explicitly triggering GC to remove garbage from config loading.
	runtime.GC()

	{
		osSignals := make(chan os.Signal, 1)
		signal.Notify(osSignals, os.Interrupt, os.Kill, syscall.SIGTERM)
		<-osSignals
	}
}
