package models_test

import (
	"encoding/json"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry-incubator/runtime-schema/models"
)

var _ = Describe("ExecutorAction", func() {
	Describe("With an invalid action", func() {
		It("should fail to marshal", func() {
			invalidAction := []string{"aliens", "from", "mars"}
			payload, err := json.Marshal(&ExecutorAction{Action: invalidAction})
			Ω(payload).Should(BeZero())
			Ω(err.(*json.MarshalerError).Err).Should(Equal(InvalidActionConversion))
		})

		It("should fail to unmarshal", func() {
			var unmarshalledAction *ExecutorAction
			actionPayload := `{"action":"alienz","args":{"from":"space"}}`
			err := json.Unmarshal([]byte(actionPayload), &unmarshalledAction)
			Ω(err).Should(Equal(InvalidActionConversion))
		})
	})

	itSerializesAndDeserializes := func(actionPayload string, action interface{}) {
		Describe("Converting to JSON", func() {
			It("creates a json representation of the object", func() {
				marshalledAction := action

				json, err := json.Marshal(&marshalledAction)
				Ω(err).Should(BeNil())
				Ω(json).Should(MatchJSON(actionPayload))
			})
		})

		Describe("Converting from JSON", func() {
			It("constructs an object from the json string", func() {
				var unmarshalledAction *ExecutorAction
				err := json.Unmarshal([]byte(actionPayload), &unmarshalledAction)
				Ω(err).Should(BeNil())
				Ω(*unmarshalledAction).Should(Equal(action))
			})
		})
	}

	Describe("Download", func() {
		itSerializesAndDeserializes(
			`{
				"action": "download",
				"args": {
					"from": "web_location",
					"to": "local_location",
					"cache_key": "elephant",
					"extract": true
				}
			}`,
			ExecutorAction{
				Action: DownloadAction{
					From:     "web_location",
					To:       "local_location",
					Extract:  true,
					CacheKey: "elephant",
				},
			},
		)
	})

	Describe("Upload", func() {
		itSerializesAndDeserializes(
			`{
				"action": "upload",
				"args": {
					"from": "local_location",
					"to": "web_location",
					"compress": true
				}
			}`,
			ExecutorAction{
				Action: UploadAction{
					From:     "local_location",
					To:       "web_location",
					Compress: true,
				},
			},
		)
	})

	Describe("Run", func() {
		itSerializesAndDeserializes(
			`{
				"action": "run",
				"args": {
					"path": "rm",
					"args": ["-rf", "/"],
					"timeout": 10000000,
					"env": [
						{"name":"FOO", "value":"1"},
						{"name":"BAR", "value":"2"}
					],
					"resource_limits":{}
				}
			}`,
			ExecutorAction{
				Action: RunAction{
					Path:    "rm",
					Args:    []string{"-rf", "/"},
					Timeout: 10 * time.Millisecond,
					Env: []EnvironmentVariable{
						{"FOO", "1"},
						{"BAR", "2"},
					},
				},
			},
		)
	})

	Describe("FetchResult", func() {
		itSerializesAndDeserializes(
			`{
				"action": "fetch_result",
				"args": {
					"file": "/tmp/foo"
				}
			}`,
			ExecutorAction{
				FetchResultAction{
					File: "/tmp/foo",
				},
			},
		)
	})

	Describe("EmitProgressAction", func() {
		itSerializesAndDeserializes(
			`{
				"action": "emit_progress",
				"args": {
					"start_message": "reticulating splines",
					"success_message": "reticulated splines",
					"failure_message": "reticulation failed",
					"action": {
						"action": "run",
						"args": {
							"path": "echo",
							"args": null,
							"timeout": 0,
							"env": null,
							"resource_limits":{}
						}
					}
				}
			}`,
			EmitProgressFor(
				ExecutorAction{
					RunAction{
						Path: "echo",
					},
				}, "reticulating splines", "reticulated splines", "reticulation failed"),
		)
	})

	Describe("Try", func() {
		itSerializesAndDeserializes(
			`{
				"action": "try",
				"args": {
					"action": {
						"action": "run",
						"args": {
							"path": "echo",
							"args": null,
							"timeout": 0,
							"env": null,
							"resource_limits":{}
						}
					}
				}
			}`,
			Try(ExecutorAction{
				RunAction{Path: "echo"},
			}),
		)
	})

	Describe("Monitor", func() {
		itSerializesAndDeserializes(
			`{
				"action": "monitor",
				"args": {
					"action": {
						"action": "run",
						"args": {
							"resource_limits": {},
							"env": null,
							"timeout": 0,
							"path": "echo",
							"args": null
						}
					},
					"healthy_hook": {
						"method": "POST",
						"url": "bogus_healthy_hook"
					},
					"unhealthy_hook": {
						"method": "DELETE",
						"url": "bogus_unhealthy_hook"
					},
					"healthy_threshold": 2,
					"unhealthy_threshold": 5
				}
			}`,
			ExecutorAction{
				MonitorAction{
					Action: ExecutorAction{RunAction{Path: "echo"}},
					HealthyHook: HealthRequest{
						Method: "POST",
						URL:    "bogus_healthy_hook",
					},
					UnhealthyHook: HealthRequest{
						Method: "DELETE",
						URL:    "bogus_unhealthy_hook",
					},
					HealthyThreshold:   2,
					UnhealthyThreshold: 5,
				},
			},
		)
	})

	Describe("Parallel", func() {
		itSerializesAndDeserializes(
			`{
        "action": "parallel",
        "args": {
          "actions": [
            {
              "action": "download",
              "args": {
                "extract": true,
                "cache_key": "elephant",
                "to": "local_location",
                "from": "web_location"
              }
            },
            {
              "action": "run",
              "args": {
                "resource_limits": {},
                "env": null,
                "timeout": 0,
                "path": "echo",
                "args": null
              }
            }
          ]
        }
      }`,
			Parallel(
				ExecutorAction{
					DownloadAction{
						From:     "web_location",
						To:       "local_location",
						Extract:  true,
						CacheKey: "elephant",
					},
				},
				ExecutorAction{
					RunAction{Path: "echo"},
				},
			),
		)
	})
})
