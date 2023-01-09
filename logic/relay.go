package logic

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gravitl/netmaker/database"
	"github.com/gravitl/netmaker/logger"
	"github.com/gravitl/netmaker/models"
)

// CreateRelay - creates a relay
func CreateRelay(relay models.RelayRequest) ([]models.Node, models.Node, error) {
	var returnnodes []models.Node

	node, err := GetNodeByID(relay.NodeID)
	if err != nil {
		return returnnodes, models.Node{}, err
	}
	host, err := GetHost(node.HostID.String())
	if err != nil {
		return returnnodes, models.Node{}, err
	}
	if host.OS != "linux" {
		return returnnodes, models.Node{}, fmt.Errorf("only linux machines can be relay nodes")
	}
	err = ValidateRelay(relay)
	if err != nil {
		return returnnodes, models.Node{}, err
	}
	node.IsRelay = true
	node.RelayAddrs = relay.RelayAddrs

	node.SetLastModified()
	nodeData, err := json.Marshal(&node)
	if err != nil {
		return returnnodes, node, err
	}
	if err = database.Insert(node.ID.String(), string(nodeData), database.NODES_TABLE_NAME); err != nil {
		return returnnodes, models.Node{}, err
	}
	returnnodes, err = SetRelayedNodes(true, node.Network, node.RelayAddrs)
	if err != nil {
		return returnnodes, node, err
	}
	return returnnodes, node, nil
}

// SetRelayedNodes- set relayed nodes
func SetRelayedNodes(setRelayed bool, networkName string, addrs []string) ([]models.Node, error) {
	var returnnodes []models.Node
	networkNodes, err := GetNetworkNodes(networkName)
	if err != nil {
		return returnnodes, err
	}
	for _, node := range networkNodes {
		for _, addr := range addrs {
			if addr == node.Address.IP.String() || addr == node.Address6.IP.String() {
				if setRelayed {
					node.IsRelayed = true
				} else {
					node.IsRelayed = false
				}
				data, err := json.Marshal(&node)
				if err != nil {
					return returnnodes, err
				}
				database.Insert(node.ID.String(), string(data), database.NODES_TABLE_NAME)
				returnnodes = append(returnnodes, node)
			}
		}
	}
	return returnnodes, nil
}
func GetRelayedNodes(relayNode *models.Node) ([]models.Node, error) {
	var returnnodes []models.Node
	networkNodes, err := GetNetworkNodes(relayNode.Network)
	if err != nil {
		return returnnodes, err
	}
	for _, node := range networkNodes {
		for _, addr := range relayNode.RelayAddrs {
			if addr == node.Address.IP.String() || addr == node.Address6.IP.String() {
				returnnodes = append(returnnodes, node)
			}
		}
	}
	return returnnodes, nil
}

// ValidateRelay - checks if relay is valid
func ValidateRelay(relay models.RelayRequest) error {
	var err error
	//isIp := functions.IsIpCIDR(gateway.RangeString)
	empty := len(relay.RelayAddrs) == 0
	if empty {
		err = errors.New("IP Ranges Cannot Be Empty")
	}
	return err
}

// UpdateRelay - updates a relay
func UpdateRelay(network string, oldAddrs []string, newAddrs []string) []models.Node {
	var returnnodes []models.Node
	time.Sleep(time.Second / 4)
	_, err := SetRelayedNodes(false, network, oldAddrs)
	if err != nil {
		logger.Log(1, err.Error())
	}
	returnnodes, err = SetRelayedNodes(true, network, newAddrs)
	if err != nil {
		logger.Log(1, err.Error())
	}
	return returnnodes
}

// DeleteRelay - deletes a relay
func DeleteRelay(network, nodeid string) ([]models.Node, models.Node, error) {
	var returnnodes []models.Node
	node, err := GetNodeByID(nodeid)
	if err != nil {
		return returnnodes, models.Node{}, err
	}
	returnnodes, err = SetRelayedNodes(false, node.Network, node.RelayAddrs)
	if err != nil {
		return returnnodes, node, err
	}

	node.IsRelay = false
	node.RelayAddrs = []string{}
	node.SetLastModified()

	data, err := json.Marshal(&node)
	if err != nil {
		return returnnodes, models.Node{}, err
	}
	if err = database.Insert(nodeid, string(data), database.NODES_TABLE_NAME); err != nil {
		return returnnodes, models.Node{}, err
	}
	return returnnodes, node, nil
}