package download_action_test

import (
	"errors"
	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/runtime-schema/models"
	steno "github.com/cloudfoundry/gosteno"
	"github.com/vito/gordon/fake_gordon"

	"github.com/cloudfoundry-incubator/executor/action_runner"
	"github.com/cloudfoundry-incubator/executor/downloader/fakedownloader"
	"github.com/cloudfoundry-incubator/executor/linuxplugin"
	. "github.com/cloudfoundry-incubator/executor/runoncehandler/execute_action/download_action"
)

var _ = Describe("DownloadAction", func() {
	var action action_runner.Action
	var result chan error

	var runOnce *models.RunOnce
	var downloadAction models.DownloadAction
	var downloader *fakedownloader.FakeDownloader
	var tempDir string
	var backendPlugin *linuxplugin.LinuxPlugin
	var wardenClient *fake_gordon.FakeGordon
	var logger *steno.Logger

	BeforeEach(func() {
		var err error

		result = make(chan error)

		downloadAction = models.DownloadAction{
			From:    "http://mr_jones",
			To:      "/tmp/Antarctica",
			Extract: false,
		}

		runOnce = &models.RunOnce{
			ContainerHandle: "some-container-handle",
		}

		downloader = &fakedownloader.FakeDownloader{}

		tempDir, err = ioutil.TempDir("", "download-action-tmpdir")
		Ω(err).ShouldNot(HaveOccurred())

		wardenClient = fake_gordon.New()

		backendPlugin = linuxplugin.New()

		logger = steno.NewLogger("test-logger")
	})

	JustBeforeEach(func() {
		action = New(
			runOnce,
			downloadAction,
			downloader,
			tempDir,
			backendPlugin,
			wardenClient,
			logger,
		)
	})

	Describe("Perform", func() {
		It("downloads the file from the given URL", func() {
			err := action.Perform()
			Ω(err).ShouldNot(HaveOccurred())

			Ω(downloader.DownloadedUrls).ShouldNot(BeEmpty())
			Ω(downloader.DownloadedUrls[0].Host).To(ContainSubstring("mr_jones"))
		})

		It("places the file in the container", func() {
			err := action.Perform()
			Ω(err).ShouldNot(HaveOccurred())

			Ω(wardenClient.ThingsCopiedIn()).ShouldNot(BeEmpty())

			copiedFile := wardenClient.ThingsCopiedIn()[0]
			Ω(copiedFile.Handle).To(Equal("some-container-handle"))
			Ω(copiedFile.Dst).To(Equal("/tmp/Antarctica"))
		})

		It("creates the parent of the destination directory", func() {
			err := action.Perform()
			Ω(err).ShouldNot(HaveOccurred())

			Ω(wardenClient.ScriptsThatRan()).ShouldNot(BeEmpty())

			scriptThatRun := wardenClient.ScriptsThatRan()[0]
			Ω(scriptThatRun.Handle).To(Equal("some-container-handle"))
			Ω(scriptThatRun.Script).To(Equal("mkdir -p /tmp"))
		})

		Context("when there is an error copying the file in", func() {
			disaster := errors.New("oh no!")

			BeforeEach(func() {
				wardenClient.SetCopyInErr(disaster)
			})

			It("sends back the error", func() {
				err := action.Perform()
				Ω(err).Should(Equal(disaster))
			})
		})
	})
})
