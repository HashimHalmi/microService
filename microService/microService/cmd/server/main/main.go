package main

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"microService/internal/microServerMainFiles"
	"net/http"
	"time"
)

func setupRoutes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/api/cart/add", microServerMainFiles.JWTMiddleware(http.HandlerFunc(microServerMainFiles.AddProductToCart)))
	mux.Handle("/api/cart", microServerMainFiles.JWTMiddleware(http.HandlerFunc(microServerMainFiles.GetCart)))
	mux.Handle("/api/cart/clear", microServerMainFiles.JWTMiddleware(http.HandlerFunc(microServerMainFiles.ClearCart)))
	mux.Handle("/api/transaction/checkout", microServerMainFiles.JWTMiddleware(http.HandlerFunc(microServerMainFiles.Checkout)))
	mux.Handle("/api/transaction/deleteLast", microServerMainFiles.JWTMiddleware(http.HandlerFunc(microServerMainFiles.DeleteLastTransaction)))
	mux.Handle("/api/transaction/pay", microServerMainFiles.JWTMiddleware(http.HandlerFunc(microServerMainFiles.ProcessPayment)))
	mux.Handle("/api/transaction/pending", microServerMainFiles.JWTMiddleware(http.HandlerFunc(microServerMainFiles.GetPendingTransaction)))
	mux.Handle("/api/transactions", microServerMainFiles.JWTMiddleware(http.HandlerFunc(microServerMainFiles.GetTransactions)))
	mux.HandleFunc("/signup", microServerMainFiles.SignUp)
	mux.HandleFunc("/login", microServerMainFiles.Login)
	mux.Handle("/", http.FileServer(http.Dir("web")))
	return mux
}

func main() {
	client, err := connectToMongoDB()
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
		return
	}
	defer client.Disconnect(context.Background())

	microServerMainFiles.SetDatabase(client.Database("microServiceDB"))

	mux := setupRoutes()
	log.Println("Server is running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func connectToMongoDB() (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		client.Disconnect(ctx)
		return nil, err
	}

	return client, nil
}
