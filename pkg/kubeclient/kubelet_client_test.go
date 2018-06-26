package kubeclient

import (
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/pkg/api/v1"
	"testing"
)

func constuctNodes(ip string) []*v1.Node {
	var nodes [1]*v1.Node
	n1 := &v1.Node{}
	address := &v1.NodeAddress{
		Type:    v1.NodeInternalIP,
		Address: ip,
	}
	n1.Status.Addresses = []v1.NodeAddress{*address}
	nodes[0] = n1
	return nodes[:]
}

func TestKubeletClientCleanupCacheCompletely(t *testing.T) {
	kc := &KubeletClient{
		cache: make(map[string]*CacheEntry),
	}
	entry := &CacheEntry{}
	kc.cache["host_1"] = entry
	kc.cache["host_2"] = entry
	assert.Equal(t, len(kc.cache), 2)
	// Test cleanup of all
	nodes := constuctNodes("host_3")
	assert.Equal(t, 2, kc.CleanupCache(nodes))
	assert.Equal(t, 0, len(kc.cache))
}

func TestKubeletClientCleanupCacheOne(t *testing.T) {
	kc := &KubeletClient{
		cache: make(map[string]*CacheEntry),
	}
	entry := &CacheEntry{}
	kc.cache["host_1"] = entry
	kc.cache["host_2"] = entry
	assert.Equal(t, len(kc.cache), 2)
	// Test cleanup of one
	nodes := constuctNodes("host_1")
	assert.Equal(t, 1, kc.CleanupCache(nodes))
	assert.Equal(t, 1, len(kc.cache))
	// See if we have the right one remained
	_, ok := kc.cache["host_1"]
	assert.True(t, ok)
}
