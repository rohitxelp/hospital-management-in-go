package controller

import (
	"context"
	"fmt"
	"golang-hospital-management/database"
	"golang-hospital-management/models"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var doctorCollection *mongo.Collection = database.OpenCollection(database.Client, "doctor")
var validate = validator.New()

func GetDoctors() gin.HandlerFunc {
	return func(c *gin.Context) {

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		recordPerPage, err := strconv.Atoi(c.Query("recordPerPage"))
		if err != nil || recordPerPage < 1 {
			recordPerPage = 10
		}

		page, err := strconv.Atoi(c.Query("page"))
		if err != nil || page < 1 {
			page = 1
		}

		startIndex := (page - 1) * recordPerPage
		startIndex, err = strconv.Atoi(c.Query("startIndex"))

		matchStage := bson.D{{"$match", bson.D{{}}}}
		groupStage := bson.D{{"$group", bson.D{{"_id", bson.D{{"_id", "null"}}}, {"total_count", bson.D{{"$sum", 1}}}, {"data", bson.D{{"$push", "$$ROOT"}}}}}}
		projectStage := bson.D{
			{
				"$project", bson.D{
					{"_id", 0},
					{"total_doctors", 1},
					{"doctors", bson.D{{"$slice", []interface{}{"$data", startIndex, recordPerPage}}}},
				}}}

		result, err := doctorCollection.Aggregate(ctx, mongo.Pipeline{
			matchStage, groupStage, projectStage})
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing doctors"})
		}
		var allDoctors []bson.M
		if err = result.All(ctx, &allDoctors); err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, allDoctors[0])
	}
}

func GetDoctor() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		doctorId := c.Param("doctor_id")
		var doctor models.Doctor

		err := doctorCollection.FindOne(ctx, bson.M{"doctor_id": doctorId}).Decode(&doctor)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while fetching the doctor"})
		}
		c.JSON(http.StatusOK, doctor)
	}
}

func CreateDoctor() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		var doctor models.Doctor

		if err := c.BindJSON(&doctor); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(doctor)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}
		err := doctorCollection.FindOne(ctx, bson.M{"Doctor_id": doctor.Doctor_id}).Decode(&doctor)
		defer cancel()
		if err != nil {
			msg := fmt.Sprintf("doctor_id was not found")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		doctor.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		doctor.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		doctor.ID = primitive.NewObjectID()
		doctor.Doctor_id = doctor.ID.Hex()

		result, insertErr := doctorCollection.InsertOne(ctx, doctor)
		if insertErr != nil {
			msg := fmt.Sprintf("doctor was not created")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		defer cancel()
		c.JSON(http.StatusOK, result)
	}
}

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func toFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}

func UpdateDoctor() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var Speciality models.Doctor
		var doctor models.Doctor

		doctorId := c.Param("doctor_id")

		if err := c.BindJSON(&doctor); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var updateObj primitive.D

		if doctor.Name != nil {
			updateObj = append(updateObj, bson.E{"name", doctor.Name})
		}

		if doctor.Speciality != nil {
			err := doctorCollection.FindOne(ctx, bson.M{"speciality": doctor.Speciality}).Decode(&Speciality)
			defer cancel()
			if err != nil {
				msg := fmt.Sprintf("message:Speciality was not found")
				c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
				return
			}

		}

		doctor.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		updateObj = append(updateObj, bson.E{"updated_at", doctor.Updated_at})

		upsert := true
		filter := bson.M{"doctor_id": doctorId}

		opt := options.UpdateOptions{
			Upsert: &upsert,
		}

		result, err := doctorCollection.UpdateOne(
			ctx,
			filter,
			bson.D{
				{"$set", updateObj},
			},
			&opt,
		)

		if err != nil {
			msg := fmt.Sprint("Doctor update failed")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		c.JSON(http.StatusOK, result)
	}
}
