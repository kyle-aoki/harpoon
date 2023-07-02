package agent

import (
	"log"
	"time"
)

type EngineStatus string

const (
	STANDBY  EngineStatus = "STANDBY"
	ACTIVE   EngineStatus = "ACTIVE"
	UPDATING EngineStatus = "UPDATING"
)

type HarpoonState struct {
	EngineStatus EngineStatus
}

var harpoonState = &HarpoonState{
	EngineStatus: STANDBY,
}

type HarpoonTarget struct {
	Repository string
	Version    string
	Port       string
}

type Deployment struct {
	ID          string
	HarpoonPort string
	Repository  string
	Version     string
	Port        string
}

var harpoonTarget *HarpoonTarget
var activeDeployment *Deployment

func ShouldUpdateDeployment() bool {
	return harpoonTarget.Repository != activeDeployment.Repository ||
		harpoonTarget.Version != activeDeployment.Version ||
		harpoonTarget.Port != activeDeployment.Port
}

func HarpoonEngine() {
	defer EngineRecover()
	for {
		EngineWait()

		UpdateContainers()
		FindActiveDeployment()
		DeleteInactive()

		if harpoonTarget == nil && activeDeployment == nil {
			log.Println("no target or deployment active, harpoon on standby")
			harpoonState.EngineStatus = STANDBY
			continue
		}

		if harpoonTarget == nil && activeDeployment != nil {
			log.Println("deployment active, harpoon is active")
			harpoonState.EngineStatus = ACTIVE
			continue
		}

		if harpoonTarget != nil && activeDeployment == nil {
			log.Println("target set and no active deployment, harpoon is updating")
			harpoonState.EngineStatus = UPDATING
			UpdateDeployment()
			forceCycle = true
			continue
		}

		if harpoonTarget != nil && activeDeployment != nil {
			if ShouldUpdateDeployment() {
				log.Println("target does not match deployment, updating deployment")
				harpoonState.EngineStatus = UPDATING
				UpdateDeployment()
			}
			log.Println("target and deployment match, harpoon is active")
			harpoonState.EngineStatus = ACTIVE
			continue
		}
	}
}

var lastCycle = time.Now()
var forceCycle = false

const harpoonEngineCadence = time.Second * 60

func EngineWait() {
	for !forceCycle && time.Since(lastCycle) < harpoonEngineCadence {
		time.Sleep(time.Second * 2)
	}
	forceCycle, lastCycle = false, time.Now()
}

func EngineRecover() {
	if r := recover(); r != nil {
		log.Println(r)
		go HarpoonEngine()
	}
}
