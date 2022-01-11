package server

import (
	"encoding/json"
	"net/http"

	kg "github.com/kubearmor/KVMService/src/log"
	tp "github.com/kubearmor/KVMService/src/types"
)

var (
	policyEventCb tp.KubeArmorHostPolicyEventCallback
	vmEventCb     tp.HandleVmCallback
	labelEventCb  tp.HandleLabelCallback
	vmListCb      tp.ListVmCallback
)

func HandleVm(w http.ResponseWriter, r *http.Request) {

	vmEvent := tp.KubeArmorVirtualMachinePolicyEvent{}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&vmEvent)
	if err != nil {
		kg.Err(err.Error())
		if _, err = w.Write([]byte("Failed to decode data")); err != nil {
			return
		}
		return
	}

	kg.Printf("Received onboarding/offboarding request for VM : vm name [%s]", vmEvent.Object.Metadata.Name)
	vmEventCb(vmEvent)
}

func HandlePolicies(w http.ResponseWriter, r *http.Request) {

	policyEvent := tp.KubeArmorHostPolicyEvent{}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&policyEvent)
	if err != nil {
		kg.Err(err.Error())
		if _, err = w.Write([]byte("Failed to decode data")); err != nil {
			return
		}
		return
	}

	kg.Printf("Received policy request for VM : policy name [%s]", policyEvent.Object.Metadata.Name)
	policyEventCb(policyEvent)
}

func HandleLabels(w http.ResponseWriter, r *http.Request) {
	labelEvent := tp.KubeArmorVirtualMachineLabel{}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&labelEvent)
	if err != nil {
		kg.Err(err.Error())
		if _, err = w.Write([]byte("Failed to decode data")); err != nil {
			return
		}
		return
	}

	kg.Printf("Received label management request for VM : %s", labelEvent.Name)
	labelList := labelEventCb(labelEvent)

	if labelEvent.Type == "LIST" {
		if _, err := w.Write([]byte(labelList)); err != nil {
			return
		}
	}
}

func ListVms(w http.ResponseWriter, r *http.Request) {
	kg.Printf("Received vm-list request")

	vmList := vmListCb()
	if _, err := w.Write([]byte(vmList)); err != nil {
		return
	}
}

func InitHttpServer(policyCbFunc tp.KubeArmorHostPolicyEventCallback,
	vmCBFunc tp.HandleVmCallback, vmListCbFunc tp.ListVmCallback,
	labelCbFunc tp.HandleLabelCallback) {

	// Set routing rule for vm handling
	http.HandleFunc("/vm", HandleVm)
	vmEventCb = vmCBFunc

	// Set routing rule for vm handling
	http.HandleFunc("/vmlist", ListVms)
	vmListCb = vmListCbFunc

	// Set routing rule for policy handling
	http.HandleFunc("/policy", HandlePolicies)
	policyEventCb = policyCbFunc

	// Set routing rule for label handling
	http.HandleFunc("/label", HandleLabels)
	labelEventCb = labelCbFunc

	//Use the default DefaultServeMux.
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		kg.Err(err.Error())
		return
	}
}
