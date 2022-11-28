package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
	"github.com/gregszalay/ocpp-charging-station-simulator/stationmessages"
	"github.com/gregszalay/ocpp-charging-station-simulator/usecases"
)

var ocpp_host = flag.String("h", "localhost:3000", "ocpp websocket server host")
var ocpp_url = flag.String("u", "/ocpp", "ocpp URL")
var ocpp_station_id = flag.String("id", "CS001", "id of the charging station")

var simulationinprogress bool = false

func main() {
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *ocpp_host, Path: *ocpp_url + "/" + *ocpp_station_id}
	fmt.Printf("connecting to %s\n", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	// B01 - Cold boot charging station
	runBootChargingStation(c, "B01", 3)

	// E02 - Cable plugin first
	runTransactions(c, "E02", 3)

	// HeartbeatRequest time interval
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case t := <-ticker.C:
			fmt.Println("\nCreating HeartbeatRequest...")
			fmt.Println(t)
			fmt.Println()
			err := c.WriteMessage(websocket.TextMessage, stationmessages.Create_HeartbeatRequest())
			if err != nil {
				log.Println("write:", err)
				return
			}
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
			// default:
			// 	if simulationinprogress == false {
			// 		fmt.Print("Please type 'charge' to start charging simulation")
			// 		input := bufio.NewScanner(os.Stdin)
			// 		input.Scan()
			// 		user_input := input.Text()
			// 		fmt.Println(user_input)
			// 		if user_input == "charge" {
			// 			runTXsimulation(c)
			// 		}
			// 	}

		}
	}
}

func runBootChargingStation(c *websocket.Conn, usecase string, interval_sec int) bool {
	provisioning := usecases.NewProvisioning()

	reqList, err := provisioning.GetValidatorList(usecase)
	if err != nil {
		log.Println("error:", err)
		return false
	}
	return doRun(c, usecase, reqList, interval_sec)
}

func runTransactions(c *websocket.Conn, usecase string, interval_sec int) bool {
	transactions := usecases.NewTransactions()

	reqList, err := transactions.GetValidatorList(usecase)
	if err != nil {
		log.Println("error:", err)
		return false
	}
	return doRun(c, usecase, reqList, interval_sec)
}

func doRun(c *websocket.Conn, usecase string, reqList usecases.ValidatorList, interval_sec int) bool {
	for i, req := range reqList {
		// Request
		request := req.GetRequest()
		fmt.Printf("\n[%s] %d/%d, Request...\n%s\n", usecase, i+1, len(reqList), request)
		err := c.WriteMessage(websocket.TextMessage, request)
		if err != nil {
			log.Println("error:", err)
			return false
		}

		// Response
		fmt.Printf("\n[%s] %d/%d, Response...\n", usecase, i+1, len(reqList))
		_, response, err := c.ReadMessage()
		if err != nil {
			log.Println("read error:", err)
			return false
		}
		fmt.Printf("%s\n", response)

		// Validate
		fmt.Printf("\n[%s] %d/%d, Validate...\n", usecase, i+1, len(reqList))
		if req.Validate(response) == false {
			fmt.Printf("\n[%s] %d/%d, Fail...\n", usecase, i+1, len(reqList))
			return false
		}

		// Wait for next interval
		time.Sleep(time.Second * time.Duration(interval_sec))
	}
	return true
}
