package create_container_action_test

import (
	"errors"
	"github.com/cloudfoundry-incubator/executor/action_runner"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/runtime-schema/models"
	steno "github.com/cloudfoundry/gosteno"
	"github.com/vito/gordon/fake_gordon"

	. "github.com/cloudfoundry-incubator/executor/runoncehandler/create_container_action"
)

var _ = Describe("CreateContainerAction", func() {
	var action action_runner.Action

	var runOnce models.RunOnce
	var gordon *fake_gordon.FakeGordon

	BeforeEach(func() {
		gordon = fake_gordon.New()

		runOnce = models.RunOnce{
			Guid:  "totally-unique",
			Stack: "penguin",
			Actions: []models.ExecutorAction{
				{
					models.RunAction{
						Script: "sudo reboot",
					},
				},
			},

			ExecutorID: "some-executor-id",
		}

		action = New(
			&runOnce,
			steno.NewLogger("test-logger"),
			gordon,
		)
	})

	Describe("Perform", func() {
		It("creates a container and updates the RunOnce's ContainerHandle", func() {
			err := action.Perform()
			Ω(err).Should(BeNil())

			Ω(gordon.CreatedHandles()).Should(HaveLen(1))
		})

		Context("when registering fails", func() {
			disaster := errors.New("oh no!")

			BeforeEach(func() {
				gordon.CreateError = disaster
			})

			It("sends back the error", func() {
				err := action.Perform()
				Ω(err).Should(Equal(disaster))
			})
		})
	})

	Describe("Cleanup", func() {
		It("destroys the created container", func() {
			err := action.Perform()
			Ω(err).Should(BeNil())

			action.Cleanup()

			Ω(gordon.DestroyedHandles()).Should(Equal(gordon.CreatedHandles()))
		})
	})
})
