package models_test

import (
	"encoding/json"

	"github.com/cloudfoundry-incubator/candiedyaml"
	. "github.com/cloudfoundry-incubator/runtime-schema/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("StagingMessages", func() {
	Describe("StagingRequestFromCC", func() {
		ccJSON := `{
           "app_id" : "fake-app_id",
           "task_id" : "fake-task_id",
           "memory_mb" : 1024,
           "disk_mb" : 10000,
           "file_descriptors" : 3,
           "environment" : [{"name": "FOO", "value":"BAR"}],
           "stack" : "fake-stack",
           "app_bits_download_uri" : "http://fake-download_uri",
           "build_artifacts_cache_download_uri" : "http://a-nice-place-to-get-valuable-artifacts.com",
           "buildpacks" : [{"name":"fake-buildpack-name", "key":"fake-buildpack-key" ,"url":"fake-buildpack-url"}]
        }`

		It("should be mapped to the CC's staging request JSON", func() {
			var stagingRequest StagingRequestFromCC
			err := json.Unmarshal([]byte(ccJSON), &stagingRequest)
			Ω(err).ShouldNot(HaveOccurred())

			Ω(stagingRequest).Should(Equal(StagingRequestFromCC{
				AppId:                          "fake-app_id",
				TaskId:                         "fake-task_id",
				Stack:                          "fake-stack",
				AppBitsDownloadUri:             "http://fake-download_uri",
				BuildArtifactsCacheDownloadUri: "http://a-nice-place-to-get-valuable-artifacts.com",
				MemoryMB:                       1024,
				FileDescriptors:                3,
				DiskMB:                         10000,
				Buildpacks: []Buildpack{
					{
						Name: "fake-buildpack-name",
						Key:  "fake-buildpack-key",
						Url:  "fake-buildpack-url",
					},
				},
				Environment: []EnvironmentVariable{
					{Name: "FOO", Value: "BAR"},
				},
			}))
		})
	})

	Describe("Buildpack", func() {
		ccJSONFragment := `{
						"name": "ocaml-buildpack",
            "key": "ocaml-buildpack-guid",
            "url": "http://ocaml.org/buildpack.zip"
          }`

		It("extracts key and url", func() {
			var buildpack Buildpack

			err := json.Unmarshal([]byte(ccJSONFragment), &buildpack)
			Ω(err).ShouldNot(HaveOccurred())

			Ω(buildpack).To(Equal(Buildpack{
				Name: "ocaml-buildpack",
				Key:  "ocaml-buildpack-guid",
				Url:  "http://ocaml.org/buildpack.zip",
			}))
		})
	})

	Describe("StagingInfo", func() {
		Context("when yaml", func() {
			stagingYAML := `---
detected_buildpack: yaml-buildpack
start_command: yaml-ize -d`

			It("exposes an extracted `detected_buildpack` property", func() {
				var stagingInfo StagingInfo

				err := candiedyaml.Unmarshal([]byte(stagingYAML), &stagingInfo)
				Ω(err).ShouldNot(HaveOccurred())

				Ω(stagingInfo).Should(Equal(StagingInfo{
					DetectedBuildpack:    "yaml-buildpack",
					DetectedStartCommand: "yaml-ize -d",
				}))
			})
		})
	})

	Describe("StagingResponseForCC", func() {
		Context("with a detected buildpack", func() {
			It("generates valid JSON with the buildpack", func() {
				stagingResponseForCC := StagingResponseForCC{
					DetectedBuildpack: "ocaml-buildpack",
				}

				Ω(json.Marshal(stagingResponseForCC)).Should(MatchJSON(`{"detected_buildpack": "ocaml-buildpack"}`))
			})
		})

		Context("with an admin buildpack key", func() {
			It("generates valid JSON with the buildpack key", func() {
				stagingResponseForCC := StagingResponseForCC{
					BuildpackKey: "admin-buildpack-key",
				}

				Ω(json.Marshal(stagingResponseForCC)).Should(MatchJSON(`{"buildpack_key": "admin-buildpack-key"}`))
			})
		})

		Context("without an admin buildpack key", func() {
			It("generates valid JSON and omits the buildpack key", func() {
				stagingResponseForCC := StagingResponseForCC{}

				Ω(json.Marshal(stagingResponseForCC)).Should(MatchJSON(`{}`))
			})
		})

		Context("with an error", func() {
			It("generates valid JSON with the error", func() {
				stagingResponseForCC := StagingResponseForCC{
					Error: "FAIL, missing camels!",
				}

				Ω(json.Marshal(stagingResponseForCC)).Should(MatchJSON(`{"error": "FAIL, missing camels!"}`))
			})
		})
	})
})
