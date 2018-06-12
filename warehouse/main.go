package main

import (
    "fmt"
    "log"
    "net/http"

    "github.com/danielcuervo/kafka-101/kafka"
)

var ordersReceived []kafka.Message

func main() {
    client, err := kafka.NewClient("kafka:9092")
    if err != nil {
        log.Println(err.Error())
    }

    ctxt, cancelFunc := context.WithCancel(context.Background())
    defer cancelFunc()

    groupID := "singlegroup"
    go client.Consume("order.received", groupID, ctxt)

    go setUpServer()

    for {
        select {
        case <-ctxt.Done():
            return;
        case msg := <-client.Receive():
            ordersReceived = append(ordersReceived, msg)

            log.Println("Order received")
        }
    }

}

func setUpServer() {
    http.HandleFunc("/orders-received", listOrders)
    _ = http.ListenAndServe(":80", nil)
}

func listOrders(rw http.ResponseWriter, r *http.Request) {
    for _, msg := range ordersReceived {
        fmt.Fprintf(rw, "%#v", msg.Payload())
        fmt.Fprint(rw, "\r\n")
    }
}