package backend_test

import (
	"github.com/cloudfoundry-incubator/bbs/models"
	"github.com/cloudfoundry-incubator/receptor"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func actionsFromDesiredTask(desiredTask receptor.TaskCreateRequest) []*models.Action {
	timeoutAction := desiredTask.Action.GetTimeoutAction()
	Expect(timeoutAction).NotTo(BeNil())
	serialAction := timeoutAction.Action.GetSerialAction()
	Expect(serialAction).NotTo(BeNil())

	return serialAction.Actions
}

func TestBackend(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Backend Suite")
}
