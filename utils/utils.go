package utils

import (
	"os"
)

func GetRootfsUrl() string {
	url := os.Getenv("FOCKER_ROOTFS_URL")
	if url == "" {
		url = "https://s3.amazonaws.com/blob.cfblob.com/fee97b71-17d7-4fab-a5b0-69d4112521e6"
	}
	return url
}
