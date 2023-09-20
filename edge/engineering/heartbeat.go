package engineering

import (
	"edge/utils"
	"log"
	"time"
)

func heartbeatFunction() {
	heartbeatMessage := HeartbeatMessage{EdgePeer: SelfPeer}
	if registryClient == nil {
		newRegistryConnection, err := ConnectToNode("registry:1234")
		if err != nil {
			utils.PrintEvent("REGISTRY_UNREACHABLE", "Impossibile stabilire connessione con il Registry")
			return
		}
		registryClient = newRegistryConnection
	}
	call := registryClient.Go("RegistryService.Heartbeat", heartbeatMessage, new(int), nil)
	select {
	case <-call.Done:
		if call.Error != nil {
			utils.PrintEvent("HEARTBEAT_ERROR", "Invio di heartbeat al Registry fallito")
			log.Println(call.Error.Error())
			registryClient.Close()
			registryClient = nil
		}
	case <-time.After(time.Second * time.Duration(utils.GetInt64EnvironmentVariable("MAX_WAITING_TIME_FOR_EDGE"))):
		utils.PrintEvent("HEARTBEAT_ERROR", "Timer scaduto.. Impossibile contattare il registry")
		log.Println(call.Error.Error())
		registryClient.Close()
		registryClient = nil
	}

}
