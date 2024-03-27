package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	//modulo locale
	"util"

	quic "github.com/quic-go/quic-go"
)

// Sinks
var d []string

// rtt
var rtt []float64

// Average service time
var service time.Duration

// Number of packets sent back
var pkts int64

// For computing moving average
var stime = time.Time{}

type application struct {
	destinationAddresses []string
	serviceTime          time.Duration
	serviceTimeType      string
	pkts                 int64 // number of pakets generated by each receved one
	obswindow            time.Duration
	sent                 int64
	received             int64
}

var app application = application{
	destinationAddresses: []string{},
	serviceTime:          0,
	serviceTimeType:      "",
	pkts:                 1,
	obswindow:            10 * time.Second,
	sent:                 0,
	received:             0,
}

// const tw = 10*time.Second

// func InitReply(workTime interface{}, dests interface{}, rtts interface{}, packets interface{}) {
func InitReply(workTime interface{}, workTimeDistribution interface{}, dests interface{}, packets interface{}) {
	// Initialize app configuration
	fmt.Println("Policy 'reply'")
	app.serviceTimeType = workTimeDistribution.(string)
	fmt.Println("\tService time type: ", app.serviceTimeType)
	app.serviceTime, _ = time.ParseDuration(workTime.(string))
	fmt.Println("\tAverage service time: ", app.serviceTime)

	app.pkts = packets.(int64)
	if dests != nil {
		// rs := rtts.([]interface{})
		ds := dests.([]interface{})
		fmt.Println("\tDestinations:")
		for i, di := range ds {
			// fmt.Printf("%d - %s:\t %f", i, di, rs[i])
			fmt.Printf("\t\t%d - %s", i, di)
			// rtt = append(rtt, rs[i].(float64))
			app.destinationAddresses = append(app.destinationAddresses, ds[i].(string))
		}
	}
}

func ReplyDecision(req *util.RoPEMessage, session *map[string]quic.EarlyConnection, i int64) bool {
	// Emulate processing time
	if app.serviceTimeType == "exp" {
		var slp time.Duration = time.Duration(math.Round(rand.ExpFloat64() * float64(app.serviceTime.Nanoseconds())))
		time.Sleep(slp)
	} else {
		time.Sleep(app.serviceTime)
	}

	req.Type = util.Response
	tmp := req.Source
	req.Source = req.Destination
	req.Destination = tmp
	app.sent = app.sent + 1
	t := time.Now()
	encoding_tmestp, _ := t.MarshalBinary()
	var body_start []byte = req.Body[:(len(req.Body) - 15)]
	req.Body = append(body_start, encoding_tmestp...)
	return false
}

func ReplySetLastResponse(lastResp *util.RoPEMessage) {
	if lastResp.Type == util.Response {
		app.received = app.received + 1
		if time.Now().After(stime.Add(app.obswindow)) {
			stime = time.Now()
			fmt.Printf("Uplink (req/s):%.6f\tDownlink (req/s):%.6f\n",
				float64(app.sent)/float64(app.obswindow.Seconds()),
				float64(app.received)/float64(app.obswindow.Seconds()))
			app.sent = 0
			app.received = 0
		}
	}
}
