package main

import (
	"context"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"math/rand"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/dgraph-io/badger/pb"
)

type Edge struct {
	From int64 `json:"from"`
	To   int64 `json:"to"`
}

type Event struct {
	From int64     `json:"from"`
	To   int64     `json:"to"`
	Time time.Time `json:"time"`
}

func main() {

	var edgedir = flag.String("folder", "/tmp/badger", "folder where your badger db lives")
	flag.Parse()

	txndir, err := ioutil.TempDir("", "badger")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("writing transactions to", txndir)
	txndb, err := badger.Open(badger.DefaultOptions(txndir))
	if err != nil {
		log.Fatal(err)
	}

	defer txndb.Close()

	var simulationDuration time.Duration
	simulationDuration = time.Duration(24 * time.Hour)

	var startTime time.Time
	startTime = time.Now()

	log.Println("using db at", *edgedir)

	db, err := badger.Open(badger.DefaultOptions(*edgedir))
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	stream := db.NewStream()

	stream.Send = func(list *pb.KVList) error {
		for _, kv := range list.GetKv() {
			var e Edge
			err = json.Unmarshal(kv.GetValue(), &e)
			if err != nil {
				log.Fatal(err)
			}

			wb := txndb.NewWriteBatch()
			defer wb.Cancel()

			// pick a rate
			λ := 1 / 60.0

			// while time is less than the simulation duration, keep emitting
			t := time.Duration(0)
			for {
				if t > simulationDuration {
					break
				}

				// sample a duration
				ti := time.Duration(int64(rand.ExpFloat64()/λ) * 1000000000)

				// update the total duration
				t += ti

				// emit
				eventTime := startTime.Add(t)
				event := Event{e.From, e.To, eventTime}
				eventBytes, err := json.Marshal(event)
				if err != nil {
					log.Fatal(err)
				}

				key, err := json.Marshal(eventTime)
				if err != nil {
					log.Fatal(err)
				}

				err = wb.Set(key, eventBytes)
				if err != nil {
					log.Fatal(err)
				}

			}

			err := wb.Flush()
			if err != nil {
				log.Fatal(err)
			}
		}

		return nil
	}

	if err := stream.Orchestrate(context.Background()); err != nil {
		log.Fatal(err)
	}

}
