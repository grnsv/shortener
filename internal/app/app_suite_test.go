package app_test

import (
	"context"
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/grnsv/shortener/internal/app"
)

func TestApp(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "App Suite")
}

var _ = Describe("Application", func() {
	var (
		ctx    context.Context
		cancel context.CancelFunc
		osArgs []string
	)

	BeforeEach(func() {
		ctx, cancel = context.WithCancel(context.Background())
		osArgs = os.Args
		os.Args = os.Args[:1]
	})

	AfterEach(func() {
		cancel()
		os.Args = osArgs
	})

	It("should create Application successfully", func() {
		application, err := app.NewApplication(ctx)
		Expect(err).To(BeNil())
		Expect(application).NotTo(BeNil())
		Expect(application.Config).NotTo(BeNil())
		Expect(application.Logger).NotTo(BeNil())
		Expect(application.Storage).NotTo(BeNil())
		Expect(application.Shortener).NotTo(BeNil())
		Expect(application.HTTPServer).NotTo(BeNil())
		Expect(application.GRPCServer).NotTo(BeNil())
	})

	It("should run and shutdown gracefully", func() {
		application, err := app.NewApplication(ctx)
		Expect(err).To(BeNil())
		Expect(application).NotTo(BeNil())

		application.Run()
		time.Sleep(time.Second)

		err = application.Shutdown(ctx)
		Expect(err).To(BeNil())
	})
})
