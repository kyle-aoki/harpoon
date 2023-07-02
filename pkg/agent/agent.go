package agent

import (
	"harpoon/pkg/util"
	"log"
	"math/rand"
	"net/http"
	"time"
)

// ############################################################################
// ############################################################################

func Agent() {
	log.Println("harpoon agent started")
	rand.Seed(time.Now().UnixMicro())
	ConfigureDockerClient()
	go HarpoonEngine()
	util.Register(AgentDeploy)
	util.Register(AgentStatus)
	util.Check(http.ListenAndServe(":7227", http.DefaultServeMux))
}

// ############################################################################
// ############################################################################

type AgentDeployInput struct {
	Repository string
	Version    string
	Port       string
}
type AgentDeployOutput struct{}

func AgentDeploy(adi *AgentDeployInput) *AgentDeployOutput {
	log.Printf("AgentDeploy -- %+v\n", adi)
	if util.IfOneTrue(
		adi.Repository == "",
		adi.Version == "",
		adi.Port == "",
	) {
		panic("invalid input")
	}
	harpoonTarget = &HarpoonTarget{
		Repository: adi.Repository,
		Version:    adi.Version,
		Port:       adi.Port,
	}
	forceCycle = true
	return &AgentDeployOutput{}
}

// ############################################################################
// ############################################################################

type AgentStatusInput struct{}
type AgentStatusOutput struct {
	HarpoonState     *HarpoonState
	ActiveDeployment *Deployment
	HarpoonTarget    *HarpoonTarget
}

func AgentStatus(asi *AgentStatusInput) *AgentStatusOutput {
	return &AgentStatusOutput{
		HarpoonState:     harpoonState,
		ActiveDeployment: activeDeployment,
		HarpoonTarget:    harpoonTarget,
	}
}

// ############################################################################
// ############################################################################
