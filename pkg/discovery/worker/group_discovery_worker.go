package worker

import (
	"github.com/turbonomic/kubeturbo/pkg/discovery/dtofactory"
	"github.com/turbonomic/kubeturbo/pkg/discovery/repository"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
)

const (
	k8sGroupWorkerID string = "GroupsDiscoveryWorker"
)

// Converts the cluster group objects to Group DTOs
type k8sEntityGroupDiscoveryWorker struct {
	id       string
	targetId string
	cluster  *repository.ClusterSummary
}

func Newk8sEntityGroupDiscoveryWorker(cluster *repository.ClusterSummary,
	targetId string) *k8sEntityGroupDiscoveryWorker {
	return &k8sEntityGroupDiscoveryWorker{
		cluster:  cluster,
		id:       k8sGroupWorkerID,
		targetId: targetId,
	}
}

// Group discovery worker collects pod and container groups discovered by different discovery workers.
// It merges the group members belonging to the same group but discovered by different discovery workers.
// Then it creates DTOs for the pod/container groups to be sent to the server.
func (worker *k8sEntityGroupDiscoveryWorker) Do(entityGroupList []*repository.EntityGroup,
) ([]*proto.GroupDTO, error) {
	var groupDTOs []*proto.GroupDTO

	// Entity groups per Owner type and instance
	entityGroupMap := make(map[string]*repository.EntityGroup)

	// Combine pod and container members discovered by different
	// discovery workers but belong to the same parent
	for _, entityGroup := range entityGroupList {
		groupId := entityGroup.GroupId
		existingGroup, groupExists := entityGroupMap[groupId]

		if groupExists {
			for etype, newmembers := range entityGroup.Members {
				members, hasMembers := existingGroup.Members[etype]
				if hasMembers {
					existingGroup.Members[etype] = append(members, newmembers...)
				}
			}
		} else {
			entityGroupMap[groupId] = entityGroup
		}
	}

	// Create DTOs for each group entity
	groupDtoBuilder, err := dtofactory.NewGroupDTOBuilder(entityGroupMap, worker.targetId)
	if err != nil {
		return nil, err
	}
	groupDTOs = groupDtoBuilder.BuildGroupDTOs()

	// Create DTO for cluster
	clusterDTO := dtofactory.NewClusterDTOBuilder(worker.cluster, worker.targetId).Build()
	if clusterDTO != nil {
		groupDTOs = append(groupDTOs, clusterDTO)
	}

	// Create a static group for Master Nodes
	masterNodesGroupDTO := dtofactory.NewMasterNodesGroupDTOBuilder(worker.cluster, worker.targetId).Build()
	if masterNodesGroupDTO != nil {
		groupDTOs = append(groupDTOs, masterNodesGroupDTO)
	}

	return groupDTOs, nil
}
