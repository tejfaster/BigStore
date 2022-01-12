package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

const URL = "mongodb://localhost:27017/?readPreference=primary&appname=MongoDB%20Compass&directConnection=true&ssl=false"
const lengthmessage = "length of name must be above 2 and less than 20 letter"

var IsLetter = regexp.MustCompile(`^[a-zA-Z]+$`).MatchString
var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

type User struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	FirstName string             `json:"firstname" bson:"firstname" binding:"required"`
	LastName  string             `json:"lastname" bson:"lastname" binding:"required"`
	Phone     string             `json:"phone" bson:"phone" binding:"required"`
	Email     string             `json:"email" bson:"email" binding:"required"`
	Address   string             `json:"address" bson:"address" binding:"required"`
	Password  string             `json:"password" bson:"password" binding:"required"`
}

func ConnectDB() *mongo.Collection {
	clientOption := options.Client().ApplyURI(URL)
	client, err := mongo.Connect(context.TODO(), clientOption)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connection is stablish")
	return client.Database("Bigstore").Collection("User")
}

func getUser(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Contect-Type", "application/json")

	user := User{}

	userParams := mux.Vars(request)

	id, _ := primitive.ObjectIDFromHex(userParams["id"])

	filter := bson.M{"_id": id}
	err := ConnectDB().FindOne(context.TODO(), filter).Decode(&user)
	if err != nil {
		log.Fatal(err.Error())
	}
	json.NewEncoder(response).Encode(user)
}

func createUser(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Contect-Type", "application/json")

	user := User{}
	var data bson.M
	_ = json.NewDecoder(request.Body).Decode(&user)
	filter := bson.M{"phone": user.Phone}
	_ = ConnectDB().FindOne(context.TODO(), filter).Decode(&data)
	if data["phone"] == user.Phone {
		json.NewEncoder(response).Encode("User Already Existed")
	} else {

		if len(user.FirstName) < 2 || len(user.FirstName) > 20 {
			json.NewEncoder(response).Encode(lengthmessage)
		} else if len(user.LastName) < 2 || len(user.LastName) > 20 {
			json.NewEncoder(response).Encode(lengthmessage)
		} else if len(user.Phone) < 10 || len(user.Phone) > 10 {
			json.NewEncoder(response).Encode("Please Enter valid phone number")
		} else if !emailRegex.MatchString(user.Email) {
			json.NewEncoder(response).Encode("Your Email is not valid")
		} else if len(user.Address) < 3 {
			json.NewEncoder(response).Encode("Please enter a valid Address")
		} else if len(user.Password) < 8 {
			json.NewEncoder(response).Encode("above or equal to 8 letter is required")
		} else {
			newpassword, _ := bcrypt.GenerateFromPassword([]byte(user.Password), 14)
			fmt.Println(string(newpassword))

			user.Password = string(newpassword)
			_, err := ConnectDB().InsertOne(context.TODO(), user)

			if err != nil {
				log.Fatal(err.Error())
			}
			json.NewEncoder(response).Encode(user)
		}

	}
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/api/store/user", createUser).Methods("Post")
	router.HandleFunc("/api/store/{id}", getUser).Methods("Get")
	log.Fatal(http.ListenAndServe("127.0.0.1:8080", router))
}
