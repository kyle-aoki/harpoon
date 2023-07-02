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

var lastCycle = time.Now()
var jam = false

const harpoonEngineCadence = time.Second * 60

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
			harpoonState.EngineStatus = STANDBY
			continue
		}

		if harpoonTarget == nil && activeDeployment != nil {
			harpoonState.EngineStatus = ACTIVE
			continue
		}

		if harpoonTarget != nil && activeDeployment == nil {
			harpoonState.EngineStatus = UPDATING
			UpdateDeployment()
			harpoonState.EngineStatus = ACTIVE
			continue
		}

		if harpoonTarget != nil && activeDeployment != nil {
			if ShouldUpdateDeployment() {
				harpoonState.EngineStatus = UPDATING
				UpdateDeployment()
			}
			harpoonState.EngineStatus = ACTIVE
			continue
		}
	}
}

func EngineWait() {
	for !jam && time.Since(lastCycle) < harpoonEngineCadence {
		time.Sleep(time.Second * 2)
	}
	jam, lastCycle = false, time.Now()
}

func EngineRecover() {
	if r := recover(); r != nil {
		log.Println(r)
		go HarpoonEngine()
	}
}
