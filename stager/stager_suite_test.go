package stager_test

import (
	steno "github.com/cloudfoundry/gosteno"
	"github.com/onsi/ginkgo/config"
	"os"
	"os/signal"
	"testing"

	"github.com/cloudfoundry/storeadapter/storerunner/etcdstorerunner"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var etcdRunner *etcdstorerunner.ETCDClusterRunner

func TestExecutor(t *testing.T) {
	registerSignalHandler()
	RegisterFailHandler(Fail)

	etcdRunner = etcdstorerunner.NewETCDClusterRunner(5001+config.GinkgoConfig.ParallelNode, 1)
	etcdRunner.Start()

	RunSpecs(t, "Stager Suite")

	etcdRunner.Stop()
}

var _ = BeforeEach(func() {
	steno.EnterTestMode(steno.LOG_DEBUG)

	etcdRunner.Stop()
	etcdRunner.Start()
})

func registerSignalHandler() {
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, os.Kill)

		select {
		case <-c:
			etcdRunner.Stop()
			os.Exit(0)
		}
	}()
}
