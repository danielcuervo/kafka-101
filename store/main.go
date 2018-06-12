package main

import (
    "fmt"
    "log"
    "math/rand"
    "net/http"
    "strconv"

    "github.com/Pallinder/go-randomdata"
    "github.com/danielcuervo/kafka-101/kafka"
)

func main() {
    http.HandleFunc("/receive-order", receiveOrder)
    _ = http.ListenAndServe(":80", nil)
}

func receiveOrder(rw http.ResponseWriter, r *http.Request) {
    client, err := kafka.NewClient("kafka:9092")
    if err != nil {
        log.Println(err.Error())
    }

    orderReceived := &OrderReceived{}
    err = client.Dispatch(orderReceived)
    if err != nil {
        log.Println(err.Error())
    }
    fmt.Fprintf(rw,"Order dispatched")
    fmt.Fprintf(rw,"\r\n")
    fmt.Fprintf(rw,orderReceived.Topic())
    fmt.Fprintf(rw,"\r\n")
    fmt.Fprintf(rw,"User id:" +  strconv.Itoa(orderReceived.Payload()["userId"].(int)))
    fmt.Fprintf(rw,"\r\n")
    for index, value := range orderReceived.Payload()["itemsSKU"].([]string) {
        fmt.Fprintf(rw,"Item SKU " + strconv.Itoa(index) + ":" +  value)
        fmt.Fprintf(rw,"\r\n")
    }
    fmt.Fprintf(rw,"Comment:" +  orderReceived.Payload()["comment"].(string))
    fmt.Fprintf(rw,"\r\n")
}

type OrderReceived struct {
    payload map[string]interface{}
}

func (or *OrderReceived) Topic() string {
    return "order.received"
}

func (or *OrderReceived) Payload() map[string]interface{} {
    if len(or.payload) == 0 {
        or.payload = map[string]interface{}{
            "userId": rand.Intn(10),
            "itemsSKU": []string{
                randomdata.RandStringRunes(10),
                randomdata.RandStringRunes(10),
            },
            "comment": "This is an order",
        }
    }

    return or.payload
}
