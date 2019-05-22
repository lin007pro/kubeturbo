package executor

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/turbonomic/kubeturbo/pkg/turbostore"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
)

type MachineActionExecutor struct {
	executor TurboK8sActionExecutor
	cache    *turbostore.Cache
}

func NewMachineActionExecutor(ae TurboK8sActionExecutor) *MachineActionExecutor {
	return &MachineActionExecutor{
		executor: ae,
		cache:    turbostore.NewCache(),
	}
}

func (s *MachineActionExecutor) unlock(key string) {
	err := s.cache.Delete(key)
	if err != nil {
		glog.Errorf("Error unlocking action %v", err)
	}
}

// Execute : executes the scale action.
func (s *MachineActionExecutor) Execute(vmDTO *TurboActionExecutorInput) (*TurboActionExecutorOutput, error) {
	nodeName := vmDTO.ActionItem.GetCurrentSE().GetDisplayName()
	var actionType ActionType
	switch vmDTO.ActionItem.GetActionType() {
	case proto.ActionItemDTO_PROVISION:
		actionType = ProvisionAction
		break
	case proto.ActionItemDTO_SUSPEND:
		actionType = SuspendAction
		break
	default:
		return nil, fmt.Errorf("Unsupported action type %v", vmDTO.ActionItem.GetActionType())
	}
	// Get on with it.
	controller, key, err := newController(nodeName, 1, actionType, s.executor.cApiClient, s.executor.kubeClient)
	if err != nil {
		return nil, err
	} else if key == nil {
		return nil, fmt.Errorf("the target machine deployment has no name")
	}
	// See if we already have this.
	_, ok := s.cache.Get(*key)
	if ok {
		return nil, fmt.Errorf("the action against the %s is already running", *key)
	}
	s.cache.Add(*key, key)
	defer s.unlock(*key)
	// Check other preconditions.
	err = controller.checkPreconditions()
	if err != nil {
		return nil, err
	}
	err = controller.executeAction()
	if err != nil {
		return nil, err
	}
	err = controller.checkSuccess()
	if err != nil {
		return nil, err
	}
	return &TurboActionExecutorOutput{Succeeded: true}, nil
}
