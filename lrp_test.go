package inigo_test

import (
	"fmt"
	"syscall"

	"github.com/cloudfoundry-incubator/inigo/helpers"
	Bbs "github.com/cloudfoundry-incubator/runtime-schema/bbs"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/cloudfoundry/gunk/timeprovider"
	"github.com/nu7hatch/gouuid"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager/lagertest"
)

var _ = FDescribe("Starting an arbitrary LRP", func() {
	var (
		processGroup ifrit.Process
		processGuid  string
		tpsAddr      string
		bbs          *Bbs.BBS
	)

	BeforeEach(func() {
		bbs = Bbs.NewBBS(
			suiteContext.EtcdRunner.Adapter(),
			timeprovider.NewTimeProvider(),
			lagertest.NewTestLogger("test"),
		)

		guid, err := uuid.NewV4()
		if err != nil {
			panic("Failed to generate AppID Guid")
		}

		processGuid = guid.String()

		suiteContext.ExecutorRunner.Start()
		suiteContext.RepRunner.Start()
		suiteContext.AuctioneerRunner.Start(AUCTION_MAX_ROUNDS)

		processes := grouper.RunGroup{
			"tps": suiteContext.TPSRunner,
		}

		processGroup = ifrit.Envoke(processes)

		tpsAddr = fmt.Sprintf("http://%s", suiteContext.TPSAddress)
	})

	AfterEach(func() {
		processGroup.Signal(syscall.SIGKILL)
		Eventually(processGroup.Wait()).Should(Receive())
	})

	It("eventually runs on an executor", func() {
		err := bbs.DesireLRP(models.DesiredLRP{
			ProcessGuid: processGuid,
			Stack:       suiteContext.RepStack,
			MemoryMB:    128,
			DiskMB:      1024,
			Ports: []models.PortMapping{
				{ContainerPort: 8080},
			},

			Actions: []models.ExecutorAction{
				models.ExecutorAction{
					models.RunAction{
						Path: "bash",
						Args: []string{
							"-c",
							"while true; do sleep 2; done",
						},
					},
				},
			},
		})
		Ω(err).ShouldNot(HaveOccurred())

		Eventually(helpers.RunningLRPInstancesPoller(tpsAddr, processGuid)).Should(HaveLen(1))
	})
})
