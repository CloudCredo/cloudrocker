package models_test

import (
	. "github.com/amitkgupta/match_array_or_slice"
	. "github.com/cloudfoundry-incubator/runtime-schema/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CircusTailorConfig", func() {
	var tailorConfig CircusTailorConfig

	BeforeEach(func() {
		tailorConfig = NewCircusTailorConfig([]string{"ocaml-buildpack", "haskell-buildpack", "bash-buildpack"})
	})

	Context("with defaults", func() {
		It("generates a script for running its tailor", func() {
			commandFlags := []string{
				"-appDir=/app",
				"-buildpackOrder=ocaml-buildpack,haskell-buildpack,bash-buildpack",
				"-buildpacksDir=/tmp/buildpacks",
				"-buildArtifactsCacheDir=/tmp/cache",
				"-outputDropletDir=/tmp/droplet",
				"-outputMetadataDir=/tmp/result",
			}

			Ω(tailorConfig.Path()).Should(Equal("/tmp/circus/tailor"))
			Ω(tailorConfig.Args()).Should(MatchArrayOrSlice(commandFlags))
		})
	})

	Context("with overrides", func() {
		BeforeEach(func() {
			tailorConfig.Set("appDir", "/some/app/dir")
			tailorConfig.Set("outputDropletDir", "/some/droplet/dir")
			tailorConfig.Set("outputMetadataDir", "/some/result/dir")
			tailorConfig.Set("buildpacksDir", "/some/buildpacks/dir")
			tailorConfig.Set("buildArtifactsCacheDir", "/some/cache/dir")
		})

		It("generates a script for running its tailor", func() {
			commandFlags := []string{
				"-appDir=/some/app/dir",
				"-buildpackOrder=ocaml-buildpack,haskell-buildpack,bash-buildpack",
				"-buildpacksDir=/some/buildpacks/dir",
				"-buildArtifactsCacheDir=/some/cache/dir",
				"-outputDropletDir=/some/droplet/dir",
				"-outputMetadataDir=/some/result/dir",
			}

			Ω(tailorConfig.Path()).Should(Equal("/tmp/circus/tailor"))
			Ω(tailorConfig.Args()).Should(MatchArrayOrSlice(commandFlags))
		})
	})

	It("returns the path to the app bits", func() {
		Ω(tailorConfig.AppDir()).To(Equal("/app"))
	})

	It("returns the path to a given buildpack", func() {
		key := "my-buildpack/key/::"
		Ω(tailorConfig.BuildpackPath(key)).To(Equal("/tmp/buildpacks/8b2f72a0702aed614f8b5d8f7f5b431b"))
	})

	It("returns the path to the staging metadata", func() {
		Ω(tailorConfig.OutputMetadataPath()).To(Equal("/tmp/result/result.json"))
	})
})
