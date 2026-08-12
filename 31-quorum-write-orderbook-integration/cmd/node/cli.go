package main

import (
	"bufio"
	"fmt"
	"log/slog"
	"multi_node_p2p/cluster"
	"multi_node_p2p/orderbook"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// no cmd/node ou um arquivo cli.go separado
func startCLI(cl *cluster.Cluster) {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("commands: order <side:bid|ask> <price> <qty>  |  cancel <orderID>")

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) == 0 {
			continue
		}

		switch fields[0] {
		case "order":
			if len(fields) != 4 {
				fmt.Println("usage: order <bid|ask> <price> <qty>")
				continue
			}

			side := orderbook.Bid
			if fields[1] == "ask" {
				side = orderbook.Ask
			}

			price, err1 := strconv.Atoi(fields[2])
			qty, err2 := strconv.Atoi(fields[3])
			if err1 != nil || err2 != nil {
				fmt.Println("price and qty must be integers")
				continue
			}

			order := orderbook.NewOrder(uuid.NewString(), "cli-user", side, price, qty)
			op := orderbook.Operation{Kind: orderbook.OpAddOrder, Order: order}

			payload, err := orderbook.EncodeOperation(op)
			if err != nil {
				slog.Error("failed to encode operation", "error", err)
				continue
			}

			id := cl.Propose(payload)
			fmt.Printf("proposed order %s, proposal id %s\n", order.ID, id)

		case "cancel":
			if len(fields) != 2 {
				fmt.Println("usage: cancel <orderID>")
				continue
			}

			op := orderbook.Operation{Kind: orderbook.OpCancel, OrderID: fields[1]}
			payload, err := orderbook.EncodeOperation(op)
			if err != nil {
				slog.Error("failed to encode operation", "error", err)
				continue
			}

			id := cl.Propose(payload)
			fmt.Printf("proposed cancel %s, proposal id %s\n", fields[1], id)

		case "list":

			fmt.Println(cl.GetOrders())

		default:
			fmt.Println("unknown command")
		}
	}
}
