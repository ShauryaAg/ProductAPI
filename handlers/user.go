package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/ShauryaAg/ProductAPI/models"
	"github.com/ShauryaAg/ProductAPI/models/db"
	"github.com/ShauryaAg/ProductAPI/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func Register(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		fmt.Println("err", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	ct := r.Header.Get("content-type")
	if !strings.Contains(ct, "application/json") {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		w.Write([]byte(fmt.Sprintf("Need content-type: 'application/json', but got %s", ct)))
		return
	}

	var user models.User
	err = json.Unmarshal(bodyBytes, &user)
	if err != nil {
		fmt.Println("err", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	user.Id = primitive.NewObjectID()
	user.HashPassword()
	result, err := db.Models["user"].InsertOne(r.Context(), user)
	if err != nil {
		fmt.Println("err", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	token, err := utils.CreateToken(user)
	if err != nil {
		fmt.Println(err)
		return
	}
	jsonBytes, err := json.Marshal(struct {
		Id    primitive.ObjectID
		Email string
		Token string
	}{result.InsertedID.(primitive.ObjectID), user.Email, token})
	if err != nil {
		fmt.Println("err", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Add("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func Login(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		fmt.Println("err", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	ct := r.Header.Get("content-type")
	if !strings.Contains(ct, "application/json") {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		w.Write([]byte(fmt.Sprintf("Need content-type: 'application/json', but got %s", ct)))
		return
	}

	var data map[string]string
	err = json.Unmarshal(bodyBytes, &data)
	if err != nil {
		fmt.Println("err", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	var user models.User
	err = db.Models["user"].FindOne(r.Context(), bson.M{"email": data["email"]}).Decode(&user)
	if err != nil {
		fmt.Println("err", err)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.Error()))
		return
	}

	valid := user.VerifyPassword(data["password"])
	var token string
	if valid {
		token, err = utils.CreateToken(user)
		if err != nil {
			fmt.Println(err)
			return
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Email/Password is incorrect"))
		return
	}

	jsonBytes, err := json.Marshal(struct {
		Id    primitive.ObjectID
		Email string
		Token string
	}{user.Id, user.Email, token})
	if err != nil {
		fmt.Println("err", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Add("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

// Get user details using JWT
func GetUser(w http.ResponseWriter, r *http.Request) {
	id := r.Header.Get("decoded")
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		fmt.Println("err", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	var user models.User
	err = db.Models["user"].FindOne(r.Context(), bson.M{"_id": objectId}).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	jsonBytes, err := json.Marshal(user)
	if err != nil {
		fmt.Println("err", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Add("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}