package agent

import (
	"errors"
	"harpoon/pkg/util"
	"math/rand"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types"
)

func AllActivePorts() []int {
	var ports []int
	for i := 0; i < len(localContainers); i++ {
		for j := 0; j < len(localContainers[i].Ports); j++ {
			ports = append(ports, int(localContainers[i].Ports[j].PublicPort))
		}
	}
	return ports
}

func FindAvailablePort() string {
	activePorts := AllActivePorts()
	var port int
	for port == 0 || util.Contains(activePorts, port) {
		port = 10_000 + rand.Intn(10_000)
	}
	return strconv.FormatInt(int64(port), 10)
}

func ParsePort(ports []types.Port) (*types.Port, bool) {
	var ipv4Only []types.Port
	for i := 0; i < len(ports); i++ {
		if !strings.Contains(ports[i].IP, "::") {
			ipv4Only = append(ipv4Only, ports[i])
		}
	}
	if len(ipv4Only) > 1 {
		return nil, false
	}
	if len(ipv4Only) == 0 {
		return nil, false
	}
	return &ipv4Only[0], true
}

func ParseImageString(image string) (repository string, version string, err error) {
	parts := strings.Split(image, ":")
	if len(parts) != 2 {
		return "", "", errors.New("something wrong with image string")
	}
	return parts[0], parts[1], nil
}
