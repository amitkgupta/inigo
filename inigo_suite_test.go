package inigo_test

import (
	"fmt"
	"os"
	"os/signal"
	"testing"

	"github.com/cloudfoundry/gunk/natsrunner"

	"github.com/cloudfoundry/storeadapter/storerunner/etcdstorerunner"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/vito/cmdtest"
	"github.com/vito/gordon"

	"github.com/cloudfoundry-incubator/inigo/executor_runner"
	"github.com/cloudfoundry-incubator/inigo/garden_runner"
	"github.com/cloudfoundry-incubator/inigo/stager_runner"

	_ "github.com/cloudfoundry-incubator/executor"
	_ "github.com/cloudfoundry-incubator/stager"
	_ "github.com/pivotal-cf-experimental/garden"
)

var etcdRunner *etcdstorerunner.ETCDClusterRunner
var wardenClient gordon.Client
var executor *cmdtest.Session

var gardenRunner *garden_runner.GardenRunner
var executorRunner *executor_runner.ExecutorRunner
var natsRunner *natsrunner.NATSRunner
var stagerRunner *stager_runner.StagerRunner

var wardenNetwork, wardenAddr string

func TestRun_once(t *testing.T) {
	registerSignalHandler()
	RegisterFailHandler(Fail)

	etcdRunner = etcdstorerunner.NewETCDClusterRunner(5001, 1)
	etcdRunner.Start()

	wardenNetwork = os.Getenv("WARDEN_NETWORK")
	wardenAddr = os.Getenv("WARDEN_ADDR")

	gardenRoot := os.Getenv("GARDEN_ROOT")
	gardenRootfs := os.Getenv("GARDEN_ROOTFS")

	if (wardenNetwork == "" || wardenAddr == "") && (gardenRoot == "" || gardenRootfs == "") {
		println("Please define either WARDEN_NETWORK and WARDEN_ADDR (for a running Warden), or")
		println("GARDEN_ROOT and GARDEN_ROOTFS (for the tests to start it)")
		println("")
		println("Skipping!")
		return
	}

	if gardenRoot != "" && gardenRootfs != "" {
		var err error

		gardenRunner, err = garden_runner.New(
			gardenRoot,
			gardenRootfs,
		)
		if err != nil {
			panic(err.Error())
		}

		gardenRunner.SnapshotsPath = ""

		err = gardenRunner.Start()
		if err != nil {
			panic(err.Error())
		}

		wardenClient = gardenRunner.NewClient()

		wardenNetwork = "tcp"
		wardenAddr = fmt.Sprintf("127.0.0.1:%d", gardenRunner.Port)
	} else {
		wardenClient = gordon.NewClient(&gordon.ConnectionInfo{
			Network: wardenNetwork,
			Addr:    wardenAddr,
		})
	}

	err := wardenClient.Connect()
	if err != nil {
		println("warden is not up!")
		os.Exit(1)
		return
	}

	executorPath, err := cmdtest.Build("github.com/cloudfoundry-incubator/executor")
	if err != nil {
		println("failed to compile executor!")
		os.Exit(1)
		return
	}

	executorRunner = executor_runner.New(
		executorPath,
		wardenNetwork,
		wardenAddr,
		etcdRunner.NodeURLS(),
	)

	stagerPath, err := cmdtest.Build("github.com/cloudfoundry-incubator/stager")
	if err != nil {
		println("failed to compile stager!")
		os.Exit(1)
		return
	}

	stagerRunner = stager_runner.New(
		stagerPath,
		etcdRunner.NodeURLS(),
	)

	natsRunner = natsrunner.NewNATSRunner(4222)

	RunSpecs(t, "RunOnce Suite")

	etcdRunner.Stop()

	if gardenRunner != nil {
		gardenRunner.Stop()
	}

	natsRunner.Stop()
}

var _ = BeforeEach(func() {
	etcdRunner.Reset()

	if gardenRunner != nil {
		// local
		gardenRunner.DestroyContainers()
	} else {
		// remote
		nukeAllWardenContainers()
	}
})

func nukeAllWardenContainers() {
	listResponse, err := wardenClient.List()
	Ω(err).ShouldNot(HaveOccurred())

	handles := listResponse.GetHandles()
	for _, handle := range handles {
		_, err := wardenClient.Destroy(handle)
		Ω(err).ShouldNot(HaveOccurred())
	}
}

func registerSignalHandler() {
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, os.Kill)

		select {
		case <-c:
			etcdRunner.Stop()
			gardenRunner.Stop()
			stagerRunner.Stop()
			os.Exit(1)
		}
	}()
}