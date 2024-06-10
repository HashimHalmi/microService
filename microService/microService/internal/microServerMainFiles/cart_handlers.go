package microServerMainFiles

import (
	"context"
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"microService/pkg/email"
	"net/http"
	"time"
)

type CartItem struct {
	ProductID string  `bson:"product_id" json:"product_id"`
	Quantity  int     `bson:"quantity" json:"quantity"`
	Price     float64 `bson:"price" json:"price"`
}

type PaymentForm struct {
	CardNumber     string `json:"cardNumber"`
	ExpirationDate string `json:"expirationDate"`
	CVV            string `json:"cvv"`
	Name           string `json:"name"`
	Address        string `json:"address"`
}

// AddProductToCart adds a product to the user's shopping cart
func AddProductToCart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var item CartItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Log the item received
	log.Printf("Received item: %+v", item)

	if item.ProductID == "" {
		log.Println("Product ID is empty")
		http.Error(w, "Product ID is required", http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		log.Println("User ID not found in context")
		http.Error(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	log.Printf("Adding item to cart for user %s: %+v", userID, item)

	if err := AddItemToUserCart(userID, item); err != nil {
		log.Printf("Failed to add item to cart for user %s: %v", userID, err)
		http.Error(w, "Failed to add item to cart", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// AddItemToUserCart handles the logic to add an item to the cart
func AddItemToUserCart(userID string, item CartItem) error {
	collection := db.Collection("carts")

	// Add logging to check item before database operation
	log.Printf("Inserting item into cart: %+v", item)

	_, err := collection.UpdateOne(
		context.TODO(),
		bson.M{"user_id": userID},
		bson.M{"$push": bson.M{"items": item}}, // Ensure the correct nesting of maps
		options.Update().SetUpsert(true),
	)
	if err != nil {
		log.Printf("Error updating cart in database: %v", err)
	}
	return err
}

// GetCart retrieves a user's shopping cart
func GetCart(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		log.Println("User ID not found in context")
		http.Error(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	cart, err := RetrieveUserCart(userID)
	if err != nil {
		log.Printf("Unable to retrieve cart for user %s: %v", userID, err)
		http.Error(w, "Unable to retrieve cart", http.StatusInternalServerError)
		return
	}
	log.Printf("Cart for user %s: %+v", userID, cart)
	json.NewEncoder(w).Encode(cart)
}

// RetrieveUserCart retrieves the cart from the database
func RetrieveUserCart(userID string) (*Cart, error) {
	collection := db.Collection("carts")
	var cart Cart
	err := collection.FindOne(context.TODO(), bson.M{"user_id": userID}).Decode(&cart)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Printf("No cart found for user %s, returning empty cart", userID)
			return &Cart{UserID: userID, Items: []CartItem{}}, nil // Return empty cart if not found
		}
		log.Printf("Error finding cart in database for user %s: %v", userID, err)
		return nil, err
	}
	return &cart, nil
}

// ClearCart clears a user's shopping cart
func ClearCart(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		log.Println("User ID not found in context")
		http.Error(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	err := ClearUserCart(userID)
	if err != nil {
		log.Printf("Unable to clear cart for user %s: %v", userID, err)
		http.Error(w, "Unable to clear cart", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// ClearUserCart removes all items from the user's cart in the database
func ClearUserCart(userID string) error {
	collection := db.Collection("carts")
	_, err := collection.UpdateOne(
		context.TODO(),
		bson.M{"user_id": userID},
		bson.M{"$set": bson.M{"items": []CartItem{}}}, // Clear the items array
	)
	if err != nil {
		log.Printf("Error clearing cart in database for user %s: %v", userID, err)
	}
	return err
}

// Checkout creates a transaction from the cart
func Checkout(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		log.Println("User ID not found in context")
		http.Error(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	transaction, err := CreateTransactionFromCart(userID)
	if err != nil {
		log.Printf("Unable to create transaction for user %s: %v", userID, err)
		http.Error(w, "Unable to create transaction", http.StatusInternalServerError)
		return
	}
	log.Printf("Transaction created for user %s: %+v", userID, transaction)
	json.NewEncoder(w).Encode(transaction)
}

// CreateTransactionFromCart converts a cart into a transaction
func CreateTransactionFromCart(userID string) (*Transaction, error) {
	cart, err := RetrieveUserCart(userID)
	if err != nil {
		return nil, err
	}

	transaction := &Transaction{
		UserID:      userID,
		Items:       cart.Items,
		TotalAmount: CalculateTotal(cart.Items),
		Status:      "pending",
		CreatedAt:   time.Now(),
	}

	collection := db.Collection("transactions")
	_, err = collection.InsertOne(context.TODO(), transaction)
	if err != nil {
		log.Printf("Error creating transaction in database for user %s: %v", userID, err)
		return nil, err
	}

	// Optionally clear the cart here
	ClearUserCart(userID)

	return transaction, nil
}

func CalculateTotal(items []CartItem) float64 {
	var total float64
	for _, item := range items {
		total += item.Price * float64(item.Quantity)
	}
	return total
}

// ProcessPayment processes the payment for a transaction
//func ProcessPayment(w http.ResponseWriter, r *http.Request) {
//	if r.Method != http.MethodPost {
//		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
//		return
//	}
//
//	var payment PaymentForm
//	if err := json.NewDecoder(r.Body).Decode(&payment); err != nil {
//		log.Printf("Error decoding payment form: %v", err)
//		http.Error(w, "Invalid request", http.StatusBadRequest)
//		return
//	}
//
//	// Assume the user ID is retrieved from the JWT token
//	userID, ok := r.Context().Value("userID").(string)
//	if !ok {
//		log.Println("User ID not found in context")
//		http.Error(w, "User ID not found in context", http.StatusUnauthorized)
//		return
//	}
//
//	// Retrieve the transaction for the user
//	transaction, err := RetrievePendingTransaction(userID)
//	if err != nil {
//		log.Printf("Failed to retrieve pending transaction for user %s: %v", userID, err)
//		http.Error(w, "Failed to retrieve transaction", http.StatusInternalServerError)
//		return
//	}
//
//	// Process the payment (always successful for testing)
//	log.Printf("Processing payment for user %s: %+v", userID, payment)
//
//	// Update the transaction status to "paid"
//	if err := UpdateTransactionStatus(transaction.ID, "paid"); err != nil {
//		log.Printf("Failed to update transaction status for user %s: %v", userID, err)
//		http.Error(w, "Failed to update transaction status", http.StatusInternalServerError)
//		return
//	}
//
//	log.Printf("Payment processed for user %s: %+v", userID, transaction)
//	w.WriteHeader(http.StatusOK)
//}

// ProcessPayment processes the payment for a transaction
//func ProcessPayment(w http.ResponseWriter, r *http.Request) {
//	if r.Method != http.MethodPost {
//		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
//		return
//	}
//
//	var payment PaymentForm
//	if err := json.NewDecoder(r.Body).Decode(&payment); err != nil {
//		log.Printf("Error decoding payment form: %v", err)
//		http.Error(w, "Invalid request", http.StatusBadRequest)
//		return
//	}
//
//	// Assume the user ID is retrieved from the JWT token
//	userID, ok := r.Context().Value("userID").(string)
//	if !ok {
//		log.Println("User ID not found in context")
//		http.Error(w, "User ID not found in context", http.StatusUnauthorized)
//		return
//	}
//
//	// Retrieve the transaction for the user
//	transaction, err := RetrievePendingTransaction(userID)
//	if err != nil {
//		log.Printf("Failed to retrieve pending transaction for user %s: %v", userID, err)
//		http.Error(w, "Failed to retrieve transaction", http.StatusInternalServerError)
//		return
//	}
//
//	// Process the payment (always successful for testing)
//	log.Printf("Processing payment for user %s: %+v", userID, payment)
//
//	// Update the transaction status to "paid"
//	if err := UpdateTransactionStatus(transaction.ID, "paid"); err != nil {
//		log.Printf("Failed to update transaction status for user %s: %v", userID, err)
//		http.Error(w, "Failed to update transaction status", http.StatusInternalServerError)
//		return
//	}
//
//	// Clear the user's cart after successful payment
//	if err := ClearUserCart(userID); err != nil {
//		log.Printf("Failed to clear cart for user %s: %v", userID, err)
//		http.Error(w, "Failed to clear cart", http.StatusInternalServerError)
//		return
//	}
//
//	log.Printf("Payment processed for user %s: %+v", userID, transaction)
//	w.WriteHeader(http.StatusOK)
//	json.NewEncoder(w).Encode(map[string]string{"redirect": "cart.html"})
//}

// RetrievePendingTransaction retrieves the pending transaction for a user
func RetrievePendingTransaction(userID string) (*Transaction, error) {
	collection := db.Collection("transactions")
	var transaction Transaction
	err := collection.FindOne(context.TODO(), bson.M{"user_id": userID, "status": "pending"}).Decode(&transaction)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Printf("No pending transaction found for user %s", userID)
			return nil, err
		}
		log.Printf("Error finding transaction in database for user %s: %v", userID, err)
		return nil, err
	}
	return &transaction, nil
}

// UpdateTransactionStatus updates the status of a transaction
func UpdateTransactionStatus(transactionID interface{}, status string) error {
	collection := db.Collection("transactions")

	filter := bson.M{"_id": transactionID}
	update := bson.M{"$set": bson.M{"status": status}}

	result, err := collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		log.Printf("Error updating transaction status in database: %v", err)
		return err
	}

	if result.ModifiedCount == 0 {
		log.Printf("No transaction found with ID %v", transactionID)
		return mongo.ErrNoDocuments
	}

	return nil
}

// GetPendingTransaction retrieves the pending transaction for the user
func GetPendingTransaction(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		log.Println("User ID not found in context")
		http.Error(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	transaction, err := RetrievePendingTransaction(userID)
	if err != nil {
		log.Printf("Failed to retrieve pending transaction for user %s: %v", userID, err)
		http.Error(w, "Failed to retrieve transaction", http.StatusInternalServerError)
		return
	}

	log.Printf("Pending transaction for user %s: %+v", userID, transaction)
	json.NewEncoder(w).Encode(transaction)
}

// GetTransactions retrieves all transactions for the user
func GetTransactions(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		log.Println("User ID not found in context")
		http.Error(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	transactions, err := RetrieveUserTransactions(userID)
	if err != nil {
		log.Printf("Failed to retrieve transactions for user %s: %v", userID, err)
		http.Error(w, "Failed to retrieve transactions", http.StatusInternalServerError)
		return
	}

	log.Printf("Transactions for user %s: %+v", userID, transactions)
	json.NewEncoder(w).Encode(transactions)
}

// RetrieveUserTransactions retrieves all transactions for a user
func RetrieveUserTransactions(userID string) ([]Transaction, error) {
	collection := db.Collection("transactions")
	var transactions []Transaction
	cursor, err := collection.Find(context.TODO(), bson.M{"user_id": userID})
	if err != nil {
		log.Printf("Error finding transactions in database for user %s: %v", userID, err)
		return nil, err
	}
	if err = cursor.All(context.TODO(), &transactions); err != nil {
		log.Printf("Error decoding transactions for user %s: %v", userID, err)
		return nil, err
	}
	return transactions, nil
}

func DeleteLastTransaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		log.Println("User ID not found in context")
		http.Error(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	collection := db.Collection("transactions")

	// Find the last transaction for the user
	var lastTransaction Transaction
	err := collection.FindOne(context.TODO(), bson.M{"user_id": userID}, options.FindOne().SetSort(bson.D{{"created_at", -1}})).Decode(&lastTransaction)
	if err != nil {
		log.Printf("Error finding last transaction for user %s: %v", userID, err)
		http.Error(w, "Failed to find last transaction", http.StatusInternalServerError)
		return
	}

	// Delete the last transaction
	_, err = collection.DeleteOne(context.TODO(), bson.M{"_id": lastTransaction.ID})
	if err != nil {
		log.Printf("Error deleting last transaction for user %s: %v", userID, err)
		http.Error(w, "Failed to delete last transaction", http.StatusInternalServerError)
		return
	}

	log.Printf("Last transaction deleted for user %s: %+v", userID, lastTransaction)
	w.WriteHeader(http.StatusOK)
}

//func ProcessPayment(w http.ResponseWriter, r *http.Request) {
//	if r.Method != http.MethodPost {
//		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
//		return
//	}
//
//	var payment PaymentForm
//	if err := json.NewDecoder(r.Body).Decode(&payment); err != nil {
//		log.Printf("Error decoding payment form: %v", err)
//		http.Error(w, "Invalid request", http.StatusBadRequest)
//		return
//	}
//
//	userID, ok := r.Context().Value("userID").(string)
//	if !ok {
//		log.Println("User ID not found in context")
//		http.Error(w, "User ID not found in context", http.StatusUnauthorized)
//		return
//	}
//
//	transaction, err := RetrievePendingTransaction(userID)
//	if err != nil {
//		log.Printf("Failed to retrieve pending transaction for user %s: %v", userID, err)
//		http.Error(w, "Failed to retrieve transaction", http.StatusInternalServerError)
//		return
//	}
//
//	log.Printf("Processing payment for user %s: %+v", userID, payment)
//	transaction.Status = "completed"
//
//	collection := db.Collection("transactions")
//	_, err = collection.UpdateOne(
//		context.TODO(),
//		bson.M{"_id": transaction.ID},
//		bson.M{"$set": bson.M{"status": transaction.Status}},
//	)
//	if err != nil {
//		log.Printf("Error updating transaction status in database: %v", err)
//		http.Error(w, "Failed to update transaction status", http.StatusInternalServerError)
//		return
//	}
//
//	log.Printf("Payment processed for user %s: %+v", userID, transaction)
//
//	// Generate and send receipt
//	pdf, err := GenerateReceiptPDF(transaction, payment.Name)
//	if err != nil {
//		log.Printf("Failed to generate receipt PDF: %v", err)
//		http.Error(w, "Failed to generate receipt", http.StatusInternalServerError)
//		return
//	}
//
//	err = email.SendReceiptEmail(userID, "Your Receipt", "Thank you for your purchase!", pdf)
//	if err != nil {
//		log.Printf("Failed to send receipt email: %v", err)
//		http.Error(w, "Failed to send receipt email", http.StatusInternalServerError)
//		return
//	}
//
//	log.Printf("Receipt sent to user %s", userID)
//	w.WriteHeader(http.StatusOK)
//	http.Redirect(w, r, "/cart.html", http.StatusSeeOther) // Redirect to cart.html after successful payment
//}

func ProcessPayment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var payment PaymentForm
	if err := json.NewDecoder(r.Body).Decode(&payment); err != nil {
		log.Printf("Error decoding payment form: %v", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		log.Println("User ID not found in context")
		http.Error(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	transaction, err := RetrievePendingTransaction(userID)
	if err != nil {
		log.Printf("Failed to retrieve pending transaction for user %s: %v", userID, err)
		http.Error(w, "Failed to retrieve transaction", http.StatusInternalServerError)
		return
	}

	log.Printf("Processing payment for user %s: %+v", userID, payment)
	transaction.Status = "completed"

	collection := db.Collection("transactions")
	_, err = collection.UpdateOne(
		context.TODO(),
		bson.M{"_id": transaction.ID},
		bson.M{"$set": bson.M{"status": transaction.Status}},
	)
	if err != nil {
		log.Printf("Error updating transaction status in database: %v", err)
		http.Error(w, "Failed to update transaction status", http.StatusInternalServerError)
		return
	}

	log.Printf("Payment processed for user %s: %+v", userID, transaction)

	// Generate and send receipt
	pdf, err := GenerateReceiptPDF(transaction, payment.Name)
	if err != nil {
		log.Printf("Failed to generate receipt PDF: %v", err)
		http.Error(w, "Failed to generate receipt", http.StatusInternalServerError)
		return
	}

	err = email.SendReceiptEmail(userID, "Your Receipt", "Thank you for your purchase!", pdf)
	if err != nil {
		log.Printf("Failed to send receipt email: %v", err)
		http.Error(w, "Failed to send receipt email", http.StatusInternalServerError)
		return
	}

	log.Printf("Receipt sent to user %s", userID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"redirect": "/cart.html"})
}
