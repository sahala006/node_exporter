//go:build !windows
// +build !windows

package reporter

import (
	"fmt"
	"os"
	"strings"
)

func GetDeviceNumber(deviceName string) string {
	fi, err := os.Readlink(fmt.Sprintf("/sys/block/%s", deviceName))
	if err != nil {
		fmt.Println(err)
		return ""
	}
	// segments := strings.Split(strings.TrimPrefix(fi, "/"), "/")
	for _, v := range strings.Split(strings.TrimPrefix(fi, "/"), "/") {
		if strings.HasPrefix(v, "ata") {
			return v
		}
	}
	return ""
	// linkedName := From(segments).FirstWithT(func(s string) bool { return strings.HasPrefix(s, "ata") })
	// fmt.Println(linkedName)
	// return linkedName.data
	// return ""
}
