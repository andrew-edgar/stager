package main_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/stager/cmd/stager/testrunner"
	"github.com/onsi/gomega/gexec"
)

func TestStager(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Stager Suite")
}

var stagerPath string
var runner *testrunner.StagerRunner

var _ = SynchronizedBeforeSuite(func() []byte {
	stager, err := gexec.Build("github.com/cloudfoundry-incubator/stager/cmd/stager", "-race")
	Expect(err).NotTo(HaveOccurred())
	return []byte(stager)
}, func(stager []byte) {
	stagerPath = string(stager)
})

var _ = SynchronizedAfterSuite(func() {
	if runner != nil {
		runner.Stop()
	}
}, func() {
	gexec.CleanupBuildArtifacts()
})
