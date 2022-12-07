package controller

import (
	"context"
	"fmt"
	"golang-hospital-management/database"
	helper "golang-hospital-management/helpers"
	"golang-hospital-management/models"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var patientCollection *mongo.Collection = database.OpenCollection(database.Client, "patient")

func GetPatients() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		recordPerPage, err := strconv.Atoi(c.Query("recordPerPage"))
		if err != nil || recordPerPage < 1 {
			recordPerPage = 10
		}

		page, err1 := strconv.Atoi(c.Query("page"))
		if err1 != nil || page < 1 {
			page = 1
		}

		startIndex := (page - 1) * recordPerPage
		startIndex, err = strconv.Atoi(c.Query("startIndex"))

		matchStage := bson.D{{"$match", bson.D{{}}}}
		projectStage := bson.D{
			{"$project", bson.D{
				{"_id", 0},
				{"total_count", 1},
				{"PATIENT", bson.D{{"$slice", []interface{}{"$data", startIndex, recordPerPage}}}},
			}}}

		result, err := patientCollection.Aggregate(ctx, mongo.Pipeline{
			matchStage, projectStage})
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing patient"})
		}

		var allpatients []bson.M
		if err = result.All(ctx, &allpatients); err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, allpatients[0])

	}
}

func GetPatient() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		patientId := c.Param("patient_id")

		var patient models.Patient

		err := patientCollection.FindOne(ctx, bson.M{"patient_id": patientId}).Decode(&patient)

		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing patient"})
		}
		c.JSON(http.StatusOK, patient)
	}
}

func SignUp() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var patient models.Patient

		//convert the JSON data coming from postman to something that golang understands
		if err := c.BindJSON(&patient); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		//validate the data based on user struct

		validationErr := validate.Struct(patient)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}
		//you'll check if the email has already been used by another user

		count, err := patientCollection.CountDocuments(ctx, bson.M{"email": patient.Email})
		defer cancel()
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while checking for the email"})
			return
		}
		//hash password

		password := HashPassword(*patient.Password)
		patient.Password = &password

		//you'll also check if the phone no. has already been used by another user

		count, err = patientCollection.CountDocuments(ctx, bson.M{"phone": patient.Phone})
		defer cancel()
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while checking for the phone number"})
			return
		}

		if count > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "this email or phone number already exsits"})
			return
		}

		//create some extra details for the user object - created_at, updated_at, ID

		patient.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		patient.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		patient.ID = primitive.NewObjectID()
		patient.Patient_id = patient.ID.Hex()

		//generate token and refersh token (generate all tokens function from helper)

		token, refreshToken, _ := helper.GenerateAllTokens(*patient.Email, *patient.First_name, *patient.Last_name, patient.Patient_id)
		patient.Token = &token
		patient.Refresh_Token = &refreshToken
		//if all ok, then you insert this new user into the user collection

		resultInsertionNumber, insertErr := patientCollection.InsertOne(ctx, patient)
		if insertErr != nil {
			msg := fmt.Sprintf("patient was not created")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		defer cancel()
		//return status OK and send the result back

		c.JSON(http.StatusOK, resultInsertionNumber)
	}
}

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var patient models.Patient
		var foundPatient models.Patient

		//convert the login data from postman which is in JSON to golang readable format

		if err := c.BindJSON(&patient); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		//find a user with that email and see if that user even exists

		err := patientCollection.FindOne(ctx, bson.M{"email": patient.Email}).Decode(&foundPatient)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "patient not found, login seems to be incorrect"})
			return
		}

		//then you will verify the password

		passwordIsValid, msg := VerifyPassword(*patient.Password, *foundPatient.Password)
		defer cancel()
		if passwordIsValid != true {
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		//if all goes well, then you'll generate tokens

		token, refreshToken, _ := helper.GenerateAllTokens(*foundPatient.Email, *foundPatient.First_name, *foundPatient.Last_name, foundPatient.Patient_id)

		//update tokens - token and refersh token
		helper.UpdateAllTokens(token, refreshToken, foundPatient.Patient_id)

		//return statusOK
		c.JSON(http.StatusOK, foundPatient)
	}
}

func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}

	return string(bytes)
}

func VerifyPassword(patientPassword string, providedPassword string) (bool, string) {

	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(patientPassword))
	check := true
	msg := ""

	if err != nil {
		msg = fmt.Sprintf("login or password is incorrect")
		check = false
	}
	return check, msg
}
