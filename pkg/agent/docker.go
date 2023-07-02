package agent

import (
	"context"
	"fmt"
	"harpoon/pkg/util"
	"io"
	"log"
	"sync"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

var cli *client.Client

func ConfigureDockerClient() {
	cli = util.Must(client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation()))
}

var localContainers []types.Container

func UpdateContainers() {
	localContainers = util.Must(cli.ContainerList(context.Background(), types.ContainerListOptions{All: true}))
}

func DeleteInactive() {
	wg := &sync.WaitGroup{}
	for i := 0; i < len(localContainers); i++ {
		if activeDeployment != nil && localContainers[i].ID == activeDeployment.ID {
			continue
		}
		log.Println("deleting", localContainers[i].Image)
		wg.Add(1)
		go func(c types.Container) {
			cli.ContainerStop(context.Background(), c.ID, container.StopOptions{})
			removeOptions := types.ContainerRemoveOptions{
				Force: false,
			}
			err := cli.ContainerRemove(context.Background(), c.ID, removeOptions)
			if err != nil {
				removeOptions.Force = true
				cli.ContainerRemove(context.Background(), c.ID, removeOptions)
			}
			wg.Done()
		}(localContainers[i])
	}
	wg.Wait()
}

func FindActiveDeployment() {
	if len(localContainers) != 1 {
		return
	}
	activeContainer := &localContainers[0]
	repository, version, err := ParseImageString(activeContainer.Image)
	if err != nil {
		return
	}
	port, exists := ParsePort(activeContainer.Ports)
	if !exists {
		return
	}
	activeDeployment = &Deployment{
		ID:          activeContainer.ID,
		Repository:  repository,
		Version:     version,
		Port:        util.ToStr(port.PrivatePort),
		HarpoonPort: util.ToStr(port.PublicPort),
	}
}

func UpdateDeployment() {
	image := fmt.Sprintf("%s:%s", harpoonTarget.Repository, harpoonTarget.Version)
	log.Println("updating deployment with image", image)

	readCloser := util.Must(cli.ImagePull(context.Background(), image, types.ImagePullOptions{}))
	_ = util.Must(io.ReadAll(readCloser))
	readCloser.Close()

	// ##########################################################################

	hostPort := FindAvailablePort()
	containerPort := harpoonTarget.Port
	hostBinding := nat.PortBinding{
		HostIP:   "0.0.0.0",
		HostPort: hostPort,
	}
	formattedContainerPort := util.Must(nat.NewPort("tcp", containerPort))
	portBinding := nat.PortMap{formattedContainerPort: []nat.PortBinding{hostBinding}}
	createResponse := util.Must(cli.ContainerCreate(context.Background(),
		&container.Config{Image: image, ExposedPorts: nat.PortSet{formattedContainerPort: struct{}{}}},
		&container.HostConfig{PortBindings: portBinding},
		&network.NetworkingConfig{}, &v1.Platform{}, "",
	))
	log.Println("created", image, "container:", createResponse.ID[16:])
	log.Println("public port", hostPort, "mapped to private port", containerPort)

	// ##########################################################################

	activeDeployment = &Deployment{
		ID:          createResponse.ID,
		HarpoonPort: hostPort,
		Repository:  harpoonTarget.Repository,
		Version:     harpoonTarget.Version,
		Port:        harpoonTarget.Port,
	}

	util.Check(cli.ContainerStart(context.Background(), createResponse.ID, types.ContainerStartOptions{}))
	log.Println("started container", createResponse.ID[16:])

	log.Println("switching nginx reverse proxy to", hostPort)
	SwitchNginxReverseProxyPort(hostPort)
	DeleteInactive()
	harpoonState.EngineStatus = ACTIVE
	log.Println("finished updating deployment")
}
