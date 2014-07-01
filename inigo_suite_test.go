package inigo_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"github.com/cloudfoundry-incubator/inigo/world"
	"github.com/cloudfoundry/gosteno"
)

var SHORT_TIMEOUT = 5 * time.Second
var LONG_TIMEOUT = 15 * time.Second

const StackName = "lucid64"

var builtArtifacts world.BuiltArtifacts
var componentMaker world.ComponentMaker

func TestInigo(t *testing.T) {
	RegisterFailHandler(Fail)

	SynchronizedBeforeSuite(func() []byte {
		payload, err := json.Marshal(world.BuiltArtifacts{
			Executables: CompileTestedExecutables(),
			Circuses:    CompileAndZipUpCircuses(),
		})
		Ω(err).ShouldNot(HaveOccurred())

		return payload
	}, func(encodedBuiltArtifacts []byte) {
		var err error

		if os.Getenv("SHORT_TIMEOUT") != "" {
			SHORT_TIMEOUT, err = time.ParseDuration(os.Getenv("SHORT_TIMEOUT"))
			Ω(err).ShouldNot(HaveOccurred())
		}

		if os.Getenv("LONG_TIMEOUT") != "" {
			LONG_TIMEOUT, err = time.ParseDuration(os.Getenv("LONG_TIMEOUT"))
			Ω(err).ShouldNot(HaveOccurred())
		}

		err = json.Unmarshal(encodedBuiltArtifacts, &builtArtifacts)
		Ω(err).ShouldNot(HaveOccurred())

		addresses := world.ComponentAddresses{
			WardenLinux:    fmt.Sprintf("127.0.0.1:%d", 10000+config.GinkgoConfig.ParallelNode),
			NATS:           fmt.Sprintf("127.0.0.1:%d", 11000+config.GinkgoConfig.ParallelNode),
			Etcd:           fmt.Sprintf("127.0.0.1:%d", 12000+config.GinkgoConfig.ParallelNode),
			EtcdPeer:       fmt.Sprintf("127.0.0.1:%d", 12500+config.GinkgoConfig.ParallelNode),
			Executor:       fmt.Sprintf("127.0.0.1:%d", 13000+config.GinkgoConfig.ParallelNode),
			Rep:            fmt.Sprintf("127.0.0.1:%d", 14000+config.GinkgoConfig.ParallelNode),
			LoggregatorIn:  fmt.Sprintf("127.0.0.1:%d", 15000+config.GinkgoConfig.ParallelNode),
			LoggregatorOut: fmt.Sprintf("127.0.0.1:%d", 16000+config.GinkgoConfig.ParallelNode),
			FileServer:     fmt.Sprintf("127.0.0.1:%d", 17000+config.GinkgoConfig.ParallelNode),
			Router:         fmt.Sprintf("127.0.0.1:%d", 18000+config.GinkgoConfig.ParallelNode),
			TPS:            fmt.Sprintf("127.0.0.1:%d", 19000+config.GinkgoConfig.ParallelNode),
			FakeCC:         fmt.Sprintf("127.0.0.1:%d", 20000+config.GinkgoConfig.ParallelNode),
		}

		wardenBinPath := os.Getenv("WARDEN_BINPATH")
		wardenRootFSPath := os.Getenv("WARDEN_ROOTFS")
		externalAddress := os.Getenv("EXTERNAL_ADDRESS")

		Ω(wardenBinPath).ShouldNot(BeEmpty(), "must provide $WARDEN_BINPATH")
		Ω(wardenRootFSPath).ShouldNot(BeEmpty(), "must provide $WARDEN_ROOTFS")
		Ω(externalAddress).ShouldNot(BeEmpty(), "must provide $EXTERNAL_ADDRESS")

		componentMaker = world.ComponentMaker{
			Artifacts: builtArtifacts,
			Addresses: addresses,

			Stack: StackName,

			ExternalAddress: externalAddress,

			WardenBinPath:    wardenBinPath,
			WardenRootFSPath: wardenRootFSPath,
		}

		// set up steno for testing
		logSink := gosteno.NewTestingSink()
		gosteno.Init(&gosteno.Config{Sinks: []gosteno.Sink{logSink}})
		gosteno.EnterTestMode()
	})

	BeforeEach(func() {
		currentTestDescription := CurrentGinkgoTestDescription()
		fmt.Fprintf(GinkgoWriter, "\n%s\n%s\n\n", strings.Repeat("~", 50), currentTestDescription.FullTestText)
	})

	RunSpecs(t, "Inigo Integration Suite")
}

func CompileTestedExecutables() world.BuiltExecutables {
	var err error

	builtExecutables := world.BuiltExecutables{}

	builtExecutables["warden-linux"], err = gexec.BuildIn(os.Getenv("WARDEN_LINUX_GOPATH"), "github.com/cloudfoundry-incubator/warden-linux", "-race")
	Ω(err).ShouldNot(HaveOccurred())

	builtExecutables["loggregator"], err = gexec.BuildIn(os.Getenv("LOGGREGATOR_GOPATH"), "loggregator/loggregator")
	Ω(err).ShouldNot(HaveOccurred())

	builtExecutables["auctioneer"], err = gexec.BuildIn(os.Getenv("AUCTIONEER_GOPATH"), "github.com/cloudfoundry-incubator/auctioneer", "-race")
	Ω(err).ShouldNot(HaveOccurred())

	builtExecutables["exec"], err = gexec.BuildIn(os.Getenv("EXECUTOR_GOPATH"), "github.com/cloudfoundry-incubator/executor", "-race")
	Ω(err).ShouldNot(HaveOccurred())

	builtExecutables["converger"], err = gexec.BuildIn(os.Getenv("CONVERGER_GOPATH"), "github.com/cloudfoundry-incubator/converger", "-race")
	Ω(err).ShouldNot(HaveOccurred())

	builtExecutables["rep"], err = gexec.BuildIn(os.Getenv("REP_GOPATH"), "github.com/cloudfoundry-incubator/rep", "-race")
	Ω(err).ShouldNot(HaveOccurred())

	builtExecutables["stager"], err = gexec.BuildIn(os.Getenv("STAGER_GOPATH"), "github.com/cloudfoundry-incubator/stager", "-race")
	Ω(err).ShouldNot(HaveOccurred())

	builtExecutables["nsync-listener"], err = gexec.BuildIn(os.Getenv("NSYNC_GOPATH"), "github.com/cloudfoundry-incubator/nsync/listener", "-race")
	Ω(err).ShouldNot(HaveOccurred())

	builtExecutables["nsync-bulker"], err = gexec.BuildIn(os.Getenv("NSYNC_GOPATH"), "github.com/cloudfoundry-incubator/nsync/bulker", "-race")
	Ω(err).ShouldNot(HaveOccurred())

	builtExecutables["app-manager"], err = gexec.BuildIn(os.Getenv("APP_MANAGER_GOPATH"), "github.com/cloudfoundry-incubator/app-manager", "-race")
	Ω(err).ShouldNot(HaveOccurred())

	builtExecutables["file-server"], err = gexec.BuildIn(os.Getenv("FILE_SERVER_GOPATH"), "github.com/cloudfoundry-incubator/file-server", "-race")
	Ω(err).ShouldNot(HaveOccurred())

	builtExecutables["route-emitter"], err = gexec.BuildIn(os.Getenv("ROUTE_EMITTER_GOPATH"), "github.com/cloudfoundry-incubator/route-emitter", "-race")
	Ω(err).ShouldNot(HaveOccurred())

	builtExecutables["router"], err = gexec.BuildIn(os.Getenv("ROUTER_GOPATH"), "github.com/cloudfoundry/gorouter", "-race")
	Ω(err).ShouldNot(HaveOccurred())

	builtExecutables["tps"], err = gexec.BuildIn(os.Getenv("TPS_GOPATH"), "github.com/cloudfoundry-incubator/tps", "-race")
	Ω(err).ShouldNot(HaveOccurred())

	return builtExecutables
}

func CompileAndZipUpCircuses() world.BuiltCircuses {
	builtCircuses := world.BuiltCircuses{}

	tailorPath, err := gexec.BuildIn(os.Getenv("LINUX_CIRCUS_GOPATH"), "github.com/cloudfoundry-incubator/linux-circus/tailor", "-race")
	Ω(err).ShouldNot(HaveOccurred())

	spyPath, err := gexec.BuildIn(os.Getenv("LINUX_CIRCUS_GOPATH"), "github.com/cloudfoundry-incubator/linux-circus/spy", "-race")
	Ω(err).ShouldNot(HaveOccurred())

	soldierPath, err := gexec.BuildIn(os.Getenv("LINUX_CIRCUS_GOPATH"), "github.com/cloudfoundry-incubator/linux-circus/soldier", "-race")
	Ω(err).ShouldNot(HaveOccurred())

	circusDir, err := ioutil.TempDir("", "circus-dir")
	Ω(err).ShouldNot(HaveOccurred())

	err = os.Rename(tailorPath, filepath.Join(circusDir, "tailor"))
	Ω(err).ShouldNot(HaveOccurred())

	err = os.Rename(spyPath, filepath.Join(circusDir, "spy"))
	Ω(err).ShouldNot(HaveOccurred())

	err = os.Rename(soldierPath, filepath.Join(circusDir, "soldier"))
	Ω(err).ShouldNot(HaveOccurred())

	cmd := exec.Command("zip", "-v", "circus.zip", "tailor", "soldier", "spy")
	cmd.Stderr = GinkgoWriter
	cmd.Stdout = GinkgoWriter
	cmd.Dir = circusDir
	err = cmd.Run()
	Ω(err).ShouldNot(HaveOccurred())

	builtCircuses[StackName] = filepath.Join(circusDir, "circus.zip")

	return builtCircuses
}
