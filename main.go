package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"os"
)

var client *mongo.Client
var collection *mongo.Collection

type Student struct {
	ID       string `json:"id" bson:"_id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func main() { 
	router := mux.NewRouter()
	 

	// Get MongoDB URI from environment variable
	mongodbURI := os.Getenv("MONGODB_URI")

	clientOptions := options.Client().ApplyURI(mongodbURI)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())

	// Initialize the "students" collection
	collection = client.Database("pbscybsec").Collection("students")

	// Define API routes
	router.HandleFunc("/students", getStudents).Methods("GET")
	router.HandleFunc("/students/{id}", getStudent).Methods("GET")
	router.HandleFunc("/students", createStudent).Methods("POST")
	router.HandleFunc("/students/{id}", updateStudent).Methods("PUT")
	router.HandleFunc("/students/{id}", deleteStudent).Methods("DELETE")

	fmt.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func getStudents(w http.ResponseWriter, r *http.Request) {
	var students []Student

	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var student Student
		if err := cursor.Decode(&student); err != nil {
			log.Println(err)
			continue
		}
		students = append(students, student)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(students)
}

func getStudent(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	studentID := params["id"]

	var student Student
	err := collection.FindOne(context.Background(), bson.M{"_id": studentID}).Decode(&student)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(student)
}

func createStudent(w http.ResponseWriter, r *http.Request) {
	var student Student
	if err := json.NewDecoder(r.Body).Decode(&student); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Generate a new ObjectID for the student
	student.ID = primitive.NewObjectID().Hex()

	_, err := collection.InsertOne(context.Background(), student)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func updateStudent(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	studentID := params["id"]

	var studentUpdates Student
	if err := json.NewDecoder(r.Body).Decode(&studentUpdates); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Convert student ID string to ObjectID
	objectID, err := primitive.ObjectIDFromHex(studentID)
	if err != nil {
		http.Error(w, "Invalid student ID", http.StatusBadRequest)
		return
	}

	// Define an update filter based on the student ID
	filter := bson.M{"_id": objectID}

	// Create an update document that sets the new values for the student
	update := bson.M{
		"$set": bson.M{
			"name":     studentUpdates.Name,
			"email":    studentUpdates.Email,
			"password": studentUpdates.Password,
		},
	}

	// Perform the update operation
	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func deleteStudent(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	studentID := params["id"]

	_, err := collection.DeleteOne(context.Background(), bson.M{"_id": studentID})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
