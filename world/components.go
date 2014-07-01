package world

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/cloudfoundry-incubator/inigo/fake_cc"
	"github.com/cloudfoundry-incubator/inigo/loggregator_runner"
	wardenrunner "github.com/cloudfoundry-incubator/warden-linux/integration/runner"
	gorouterconfig "github.com/cloudfoundry/gorouter/config"
	"github.com/fraenkel/candiedyaml"
	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"
)

type BuiltExecutables map[string]string
type BuiltCircuses map[string]string

const CircusZipFilename = "some-circus.tgz"

const CONVERGE_REPEAT_INTERVAL = time.Second
const KICK_PENDING_TASK_DURATION = time.Second
const KICK_PENDING_LRP_START_AUCTION_DURATION = time.Second
const EXPIRE_CLAIMED_TASK_DURATION = 5 * time.Second
const EXPIRE_CLAIMED_LRP_START_AUCTION_DURATION = 5 * time.Second

type BuiltArtifacts struct {
	Executables BuiltExecutables
	Circuses    BuiltCircuses
}

type ComponentAddresses struct {
	NATS           string
	Etcd           string
	EtcdPeer       string
	LoggregatorIn  string
	LoggregatorOut string
	Executor       string
	Rep            string
	FakeCC         string
	FileServer     string
	Router         string
	TPS            string
	WardenLinux    string
}

type ComponentMaker struct {
	Artifacts BuiltArtifacts
	Addresses ComponentAddresses

	ExternalAddress string

	Stack string

	WardenBinPath    string
	WardenRootFSPath string
}

func (maker ComponentMaker) NATS(argv ...string) ifrit.Runner {
	host, port, err := net.SplitHostPort(maker.Addresses.NATS)
	Ω(err).ShouldNot(HaveOccurred())

	return &ginkgomon.Runner{
		Name:          "gnatsd",
		BinPath:       "gnatsd",
		AnsiColorCode: "30",
		// StartCheck:        "gnatsd is ready",
		// StartCheckTimeout: 5 * time.Second,
		Args: append([]string{
			"--addr", host,
			"--port", port,
		}, argv...),
	}
}

func (maker ComponentMaker) Etcd(argv ...string) ifrit.Runner {
	nodeName := fmt.Sprintf("etcd_%d", ginkgo.GinkgoParallelNode())
	dataDir := path.Join(os.TempDir(), nodeName)

	return &ginkgomon.Runner{
		Name:              "etcd",
		BinPath:           "etcd",
		AnsiColorCode:     "30",
		StartCheck:        "leader changed",
		StartCheckTimeout: 5 * time.Second,
		Args: append([]string{
			"-data-dir", dataDir,
			"-addr", maker.Addresses.Etcd,
			"-peer-addr", maker.Addresses.EtcdPeer,
			"-name", nodeName,
		}, argv...),
		Cleanup: func() {
			err := os.RemoveAll(dataDir)
			Ω(err).ShouldNot(HaveOccurred())
		},
	}
}

func (maker ComponentMaker) WardenLinux(argv ...string) *wardenrunner.Runner {
	return wardenrunner.New(
		maker.Addresses.WardenLinux,
		maker.Artifacts.Executables["warden-linux"],
		maker.WardenBinPath,
		maker.WardenRootFSPath,
		argv...,
	)
}

func (maker ComponentMaker) Executor(argv ...string) ifrit.Runner {
	return &ginkgomon.Runner{
		Name:              "exec",
		BinPath:           maker.Artifacts.Executables["exec"],
		AnsiColorCode:     "31",
		StartCheck:        "executor.started",
		StartCheckTimeout: 5 * time.Second,
		Args: append([]string{
			"-listenAddr", maker.Addresses.Executor,
			"-wardenNetwork", "tcp",
			"-wardenAddr", maker.Addresses.WardenLinux,
			"-loggregatorServer", maker.Addresses.LoggregatorIn,
			"-loggregatorSecret", "loggregator-secret",
			"-containerMaxCpuShares", "1024",
		}, argv...),
	}
}

func (maker ComponentMaker) Rep(argv ...string) ifrit.Runner {
	return &ginkgomon.Runner{
		Name:              "rep",
		BinPath:           maker.Artifacts.Executables["rep"],
		AnsiColorCode:     "32",
		StartCheck:        "rep.started",
		StartCheckTimeout: 5 * time.Second,
		Args: append([]string{
			"-stack", maker.Stack,
			"-lrpHost", maker.ExternalAddress,
			"-listenAddr", maker.Addresses.Rep,
			"-etcdCluster", "http://" + maker.Addresses.Etcd,
			"-natsAddresses", maker.Addresses.NATS,
			"-repID", "the-rep-id",
			"-executorURL", "http://" + maker.Addresses.Executor,
			// "-heartbeatInterval", "TODO",
		}, argv...),
	}
}

func (maker ComponentMaker) Converger(argv ...string) ifrit.Runner {

	return &ginkgomon.Runner{

		Name:              "converger",
		BinPath:           maker.Artifacts.Executables["converger"],
		AnsiColorCode:     "32",
		StartCheck:        "converger.started",
		StartCheckTimeout: 5 * time.Second,

		Args: append([]string{
			"-etcdCluster", "http://" + maker.Addresses.Etcd,

			// shorter base intervals so tests don't have to wait so long
			"-convergeRepeatInterval", CONVERGE_REPEAT_INTERVAL.String(),
			"-kickPendingTaskDuration", KICK_PENDING_TASK_DURATION.String(),
			"-kickPendingLRPStartAuctionDuration", KICK_PENDING_LRP_START_AUCTION_DURATION.String(),
			"-expireClaimedTaskDuration", EXPIRE_CLAIMED_TASK_DURATION.String(),
			"-expireClaimedLRPStartAuctionDuration", EXPIRE_CLAIMED_LRP_START_AUCTION_DURATION.String(),
		}, argv...),
	}
}

func (maker ComponentMaker) Auctioneer(argv ...string) ifrit.Runner {
	return &ginkgomon.Runner{
		Name:          "auctioneer",
		BinPath:       maker.Artifacts.Executables["auctioneer"],
		AnsiColorCode: "33",
		StartCheck:    "auctioneer.started",
		Args: append([]string{
			"-etcdCluster", "http://" + maker.Addresses.Etcd,
			"-natsAddresses", maker.Addresses.NATS,

			// we limit this to prevent overwhelming numbers of auctioneer logs.  it
			// should not impact the behavior of the tests.
			"-maxRounds", "3",
		}, argv...),
	}
}

func (maker ComponentMaker) AppManager(argv ...string) ifrit.Runner {
	return &ginkgomon.Runner{
		Name:          "app-manager",
		BinPath:       maker.Artifacts.Executables["app-manager"],
		AnsiColorCode: "34",
		StartCheck:    "app_manager.started",
		Args: append([]string{
			"-etcdCluster", "http://" + maker.Addresses.Etcd,
			"-repAddrRelativeToExecutor", maker.Addresses.Rep,
			"-circuses", fmt.Sprintf(`{"%s": "%s"}`, maker.Stack, CircusZipFilename),
		}, argv...),
	}
}

func (maker ComponentMaker) RouteEmitter(argv ...string) ifrit.Runner {
	return &ginkgomon.Runner{
		Name:          "route-emitter",
		BinPath:       maker.Artifacts.Executables["route-emitter"],
		AnsiColorCode: "35",
		StartCheck:    "route-emitter.started",
		Args: append([]string{
			"-etcdCluster", "http://" + maker.Addresses.Etcd,
			"-natsAddresses", maker.Addresses.NATS,
		}, argv...),
	}
}

func (maker ComponentMaker) TPS(argv ...string) ifrit.Runner {
	return &ginkgomon.Runner{
		Name:          "tps",
		BinPath:       maker.Artifacts.Executables["tps"],
		AnsiColorCode: "36",
		StartCheck:    "tps.started",
		Args: append([]string{
			"-etcdCluster", "http://" + maker.Addresses.Etcd,
			"-natsAddresses", maker.Addresses.NATS,
			"-listenAddr", maker.Addresses.TPS,
		}, argv...),
	}
}

func (maker ComponentMaker) NsyncListener(argv ...string) ifrit.Runner {
	return &ginkgomon.Runner{
		Name:          "nsync-listener",
		BinPath:       maker.Artifacts.Executables["nsync-listener"],
		AnsiColorCode: "37",
		StartCheck:    "nsync.listener.started",
		Args: append([]string{
			"-etcdCluster", "http://" + maker.Addresses.Etcd,
			"-natsAddresses", maker.Addresses.NATS,
		}, argv...),
	}
}

func (maker ComponentMaker) FileServer(argv ...string) (ifrit.Runner, string) {
	servedFilesDir, err := ioutil.TempDir("", "file-server-files")
	Ω(err).ShouldNot(HaveOccurred())

	host, port, err := net.SplitHostPort(maker.Addresses.FileServer)
	Ω(err).ShouldNot(HaveOccurred())

	return &ginkgomon.Runner{
		Name:          "file-server",
		BinPath:       maker.Artifacts.Executables["file-server"],
		AnsiColorCode: "31",
		StartCheck:    "file-server.ready",
		Args: append([]string{
			"-address", host,
			"-port", port,
			"-etcdCluster", "http://" + maker.Addresses.Etcd,
			"-ccAddress", "http://" + maker.Addresses.FakeCC,
			"-ccJobPollingInterval", "100ms",
			"-ccUsername", fake_cc.CC_USERNAME,
			"-ccPassword", fake_cc.CC_PASSWORD,
			"-staticDirectory", servedFilesDir,
		}, argv...),
		Cleanup: func() {
			err := os.RemoveAll(servedFilesDir)
			Ω(err).ShouldNot(HaveOccurred())
		},
	}, servedFilesDir
}

func (maker ComponentMaker) Router() ifrit.Runner {
	_, routerPort, err := net.SplitHostPort(maker.Addresses.Router)
	Ω(err).ShouldNot(HaveOccurred())

	routerPortInt, err := strconv.Atoi(routerPort)
	Ω(err).ShouldNot(HaveOccurred())

	natsHost, natsPort, err := net.SplitHostPort(maker.Addresses.NATS)
	Ω(err).ShouldNot(HaveOccurred())

	natsPortInt, err := strconv.Atoi(natsPort)
	Ω(err).ShouldNot(HaveOccurred())

	routerConfig := &gorouterconfig.Config{
		Port: uint16(routerPortInt),

		PruneStaleDropletsIntervalInSeconds: 5,
		DropletStaleThresholdInSeconds:      10,
		PublishActiveAppsIntervalInSeconds:  0,
		StartResponseDelayIntervalInSeconds: 1,

		Nats: []gorouterconfig.NatsConfig{
			{
				Host: natsHost,
				Port: uint16(natsPortInt),
			},
		},
		Logging: gorouterconfig.LoggingConfig{
			File:  "/dev/stdout",
			Level: "info",
		},
	}

	configFile, err := ioutil.TempFile(os.TempDir(), "router-config")
	Ω(err).ShouldNot(HaveOccurred())

	defer configFile.Close()

	err = candiedyaml.NewEncoder(configFile).Encode(routerConfig)
	Ω(err).ShouldNot(HaveOccurred())

	return &ginkgomon.Runner{
		Name:              "router",
		BinPath:           maker.Artifacts.Executables["router"],
		AnsiColorCode:     "32",
		StartCheck:        "router.started",
		StartCheckTimeout: 5 * time.Second, // it waits 1 second before listening. yep.
		Args:              []string{"-c", configFile.Name()},
		Cleanup: func() {
			err := os.Remove(configFile.Name())
			Ω(err).ShouldNot(HaveOccurred())
		},
	}
}

func (maker ComponentMaker) Loggregator() ifrit.Runner {
	_, inPort, err := net.SplitHostPort(maker.Addresses.LoggregatorIn)
	Ω(err).ShouldNot(HaveOccurred())

	_, outPort, err := net.SplitHostPort(maker.Addresses.LoggregatorOut)
	Ω(err).ShouldNot(HaveOccurred())

	inPortInt, err := strconv.Atoi(inPort)
	Ω(err).ShouldNot(HaveOccurred())

	outPortInt, err := strconv.Atoi(outPort)
	Ω(err).ShouldNot(HaveOccurred())

	natsHost, natsPort, err := net.SplitHostPort(maker.Addresses.NATS)
	Ω(err).ShouldNot(HaveOccurred())

	natsPortInt, err := strconv.Atoi(natsPort)
	Ω(err).ShouldNot(HaveOccurred())

	loggregatorConfig := loggregator_runner.Config{
		IncomingPort:           inPortInt,
		OutgoingPort:           outPortInt,
		MaxRetainedLogMessages: 1000,
		SharedSecret:           "loggregator-secret",
		NatsHost:               natsHost,
		NatsPort:               natsPortInt,
	}

	configFile, err := ioutil.TempFile(os.TempDir(), "loggregator-config")
	Ω(err).ShouldNot(HaveOccurred())

	defer configFile.Close()

	err = json.NewEncoder(configFile).Encode(loggregatorConfig)
	Ω(err).ShouldNot(HaveOccurred())

	return &ginkgomon.Runner{
		Name:              "loggregator",
		BinPath:           maker.Artifacts.Executables["loggregator"],
		AnsiColorCode:     "33",
		StartCheck:        "Listening on port",
		StartCheckTimeout: 5 * time.Second, // it waits 1 second before listening. yep.
		Args:              []string{"-config", configFile.Name()},
		Cleanup: func() {
			err := os.Remove(configFile.Name())
			Ω(err).ShouldNot(HaveOccurred())
		},
	}
}

func (maker ComponentMaker) FakeCC() *fake_cc.FakeCC {
	return fake_cc.New(maker.Addresses.FakeCC)
}

func (maker ComponentMaker) Stager(argv ...string) ifrit.Runner {
	return &ginkgomon.Runner{
		Name:          "stager",
		BinPath:       maker.Artifacts.Executables["stager"],
		AnsiColorCode: "34",
		StartCheck:    "Listening for staging requests!",
		Args: append([]string{
			"-etcdCluster", "http://" + maker.Addresses.Etcd,
			"-natsAddresses", maker.Addresses.NATS,
			"-circuses", fmt.Sprintf(`{"%s": "%s"}`, maker.Stack, CircusZipFilename),
		}, argv...),
	}
}
