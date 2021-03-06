package rainmaker_test

import (
	"fmt"
	"time"
	"net/http"
	"net/http/httptest"

	"github.com/pivotal-cf-experimental/rainmaker"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("BuildpacksService", func() {
	var (
		token   string
		service rainmaker.BuildpacksService
	)

	BeforeEach(func() {
		token = "token"
		service = rainmaker.NewBuildpacksService(rainmaker.Config{
			Host: fakeCloudController.URL(),
		})
	})

	Describe("Create", func() {
		It("creates a buildpack with the given name", func() {
			buildpack, err := service.Create("my-buildpack", token, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(buildpack.GUID).NotTo(BeEmpty())
			Expect(buildpack.URL).To(Equal(fmt.Sprintf("/v2/buildpacks/%s", buildpack.GUID)))
			Expect(buildpack.CreatedAt).To(Equal(time.Time{}.UTC()))
			Expect(buildpack.UpdatedAt).To(Equal(time.Time{}))
			Expect(buildpack.Name).To(Equal("my-buildpack"))
			Expect(buildpack.Position).To(Equal(0))
			Expect(buildpack.Enabled).To(BeFalse())
			Expect(buildpack.Locked).To(BeFalse())
			Expect(buildpack.Filename).To(BeEmpty())
		})

		It("creates a buildpack with optional values", func() {
			buildpack, err := service.Create("my-buildpack", token, &rainmaker.CreateBuildpackOptions{
				Position: rainmaker.IntPtr(3),
				Enabled:  rainmaker.BoolPtr(true),
				Locked:   rainmaker.BoolPtr(true),
				Filename: rainmaker.StringPtr("some-file"),
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(buildpack.GUID).NotTo(BeEmpty())
			Expect(buildpack.URL).To(Equal(fmt.Sprintf("/v2/buildpacks/%s", buildpack.GUID)))
			Expect(buildpack.CreatedAt).To(Equal(time.Time{}.UTC()))
			Expect(buildpack.UpdatedAt).To(Equal(time.Time{}))
			Expect(buildpack.Name).To(Equal("my-buildpack"))
			Expect(buildpack.Position).To(Equal(3))
			Expect(buildpack.Enabled).To(BeTrue())
			Expect(buildpack.Locked).To(BeTrue())
			Expect(buildpack.Filename).To(Equal("some-file"))
		})

		Context("when the request errors", func() {
			PIt("returns the error", func() {})
		})

		Context("when the response cannot be unmarshalled", func() {
			PIt("returns the error", func() {})
		})
	})

	Describe("Get", func() {
		var (
			bp rainmaker.Buildpack
		)

		BeforeEach(func() {
			var err error
			bp, err = service.Create("my-buildpack", token, &rainmaker.CreateBuildpackOptions{
				Position: rainmaker.IntPtr(1),
				Enabled:  rainmaker.BoolPtr(true),
			})
			Expect(err).NotTo(HaveOccurred())
		})

		It("finds the buildpack", func() {
			buildpack, err := service.Get(bp.GUID, token)
			Expect(err).NotTo(HaveOccurred())
			Expect(buildpack).To(Equal(rainmaker.Buildpack{
				GUID:      bp.GUID,
				URL:       fmt.Sprintf("/v2/buildpacks/%s", bp.GUID),
				CreatedAt: time.Time{}.UTC(),
				Name:      "my-buildpack",
				Position:  1,
				Enabled:   true,
			}))
		})

		Context("when the buildpack does not exist", func() {
			It("returns an error", func() {
				_, err := service.Get("does-not-exist", token)
				Expect(err).To(BeAssignableToTypeOf(rainmaker.NotFoundError{}))
			})
		})

		Context("when the request errors", func() {
			PIt("returns the error", func() {})
		})

		Context("when the response cannot be unmarshalled", func() {
			PIt("returns the error", func() {})
		})
	})

	Describe("Delete", func() {
		var (
			buildpack rainmaker.Buildpack
		)

		BeforeEach(func() {
			var err error
			buildpack, err = service.Create("my-buildpack", token, &rainmaker.CreateBuildpackOptions{
				Position: rainmaker.IntPtr(1),
				Enabled:  rainmaker.BoolPtr(true),
			})
			Expect(err).NotTo(HaveOccurred())
		})

		It("deletes the buildpack", func() {
			err := service.Delete(buildpack.GUID, token)
			Expect(err).NotTo(HaveOccurred())

			_, err = service.Get(buildpack.GUID, token)
			Expect(err).To(BeAssignableToTypeOf(rainmaker.NotFoundError{}))
		})
	})

	Describe("Update", func() {
		var (
			buildpack rainmaker.Buildpack
		)

		BeforeEach(func() {
			var err error
			buildpack, err = service.Create("my-buildpack", token, &rainmaker.CreateBuildpackOptions{
				Position: rainmaker.IntPtr(1),
				Enabled:  rainmaker.BoolPtr(true),
				Locked:   rainmaker.BoolPtr(false),
				Filename: rainmaker.StringPtr("some-file"),
			})
			Expect(err).NotTo(HaveOccurred())
		})

		It("modifies the buildpack with the updated values", func() {
			buildpack.Name = "updated-buildpack"
			buildpack.Enabled = false
			buildpack.Position = 3
			buildpack.Locked = true
			buildpack.Filename = "some-other-file"

			updatedBuildpack, err := service.Update(buildpack, token)
			Expect(err).NotTo(HaveOccurred())
			Expect(updatedBuildpack.Name).To(Equal("updated-buildpack"))
			Expect(updatedBuildpack.Enabled).To(BeFalse())
			Expect(updatedBuildpack.Position).To(Equal(3))
			Expect(updatedBuildpack.Locked).To(BeTrue())
			Expect(updatedBuildpack.Filename).To(Equal("some-other-file"))

			fetchedBuildpack, err := service.Get(buildpack.GUID, token)
			Expect(err).NotTo(HaveOccurred())
			Expect(fetchedBuildpack.Name).To(Equal("updated-buildpack"))
			Expect(fetchedBuildpack.Enabled).To(BeFalse())
			Expect(fetchedBuildpack.Position).To(Equal(3))
			Expect(fetchedBuildpack.Locked).To(BeTrue())
			Expect(fetchedBuildpack.Filename).To(Equal("some-other-file"))
		})

		Context("when the buildpack does not exist", func() {
			It("returns an error", func() {
				buildpack.GUID = "some-missing-guid"

				_, err := service.Update(buildpack, token)
				Expect(err).To(BeAssignableToTypeOf(rainmaker.NotFoundError{}))
			})
		})

		Context("when the API returns some malformed JSON", func() {
			It("returns an error", func() {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusCreated)
					w.Write([]byte(`%%%%`))
				}))
				defer server.Close()

				service = rainmaker.NewBuildpacksService(rainmaker.Config{
					Host: server.URL,
				})

				_, err := service.Update(buildpack, token)
				Expect(err).To(BeAssignableToTypeOf(rainmaker.Error{}))
			})
		})
	})
})
