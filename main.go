package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Constants
const DBNAME = "catfacts"
const COLLNAME = "facts"
const MONGOURL = "mongodb:27017"

// Server
type Server struct {
	client *mongo.Client
}

func NewServer(c *mongo.Client) *Server {
	return &Server{
		client: c,
	}
}


func (s *Server) handleGetAllFacts(w http.ResponseWriter, r *http.Request) {
	coll := s.client.Database(DBNAME).Collection(COLLNAME)

	query := bson.M{}
	cursor, err := coll.Find(context.TODO(), query)
	if err != nil {
		log.Fatal(err)
	}

	results := []bson.M{}
	if err = cursor.All(context.TODO(), &results); err != nil {
		log.Fatal(err)
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}


// Cat Fact worker
type CatFactWorker struct {
	client *mongo.Client
}

func NewCatFactWorker (c *mongo.Client) *CatFactWorker {
	return &CatFactWorker{
		client: c,
	}
}

func (cfw *CatFactWorker) start() error {
	coll := cfw.client.Database(DBNAME).Collection(COLLNAME)
	ticker := time.NewTicker(2 * time.Second)

	for {
		resp, err := http.Get("https://catfact.ninja/fact")
		if err != nil {
			return err
		}

		fmt.Println(resp.Body)

		var CatFact bson.M
		if err := json.NewDecoder(resp.Body).Decode(&CatFact); err != nil {
			return err
		}

		_, err = coll.InsertOne(context.TODO(), CatFact)
		if err != nil {
			return err
		}

		<-ticker.C

	}

}

func main() {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(fmt.Sprintf("mongodb://%s", MONGOURL)))
	if err != nil {
		panic(err)
	}

	worker := NewCatFactWorker(client)
	go worker.start()

	fmt.Println(client)

	server := NewServer(client)
	http.HandleFunc("/facts", server.handleGetAllFacts)
	http.ListenAndServe(":3000", nil)

}
