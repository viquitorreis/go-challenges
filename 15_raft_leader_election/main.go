package main

import (
	"fmt"
	"time"
)

func main() {
	cluster := NewCluster(3)
	cluster.Start()

	fmt.Println("Cluster started, watching for leader election...")
	time.Sleep(2 * time.Second)

	leader := cluster.GetLeader()
	if leader != nil {
		fmt.Printf("Leader elected: Node %d (term %d)\n", leader.id, leader.currentTerm)
	}

	// Simula falha do líder
	fmt.Println("Killing the leader...")
	cluster.KillLeader()

	time.Sleep(2 * time.Second)

	newLeader := cluster.GetLeader()
	if newLeader != nil {
		fmt.Printf("New leader elected: Node %d (term %d)\n", newLeader.id, newLeader.currentTerm)
	}

	cluster.Stop()
}
