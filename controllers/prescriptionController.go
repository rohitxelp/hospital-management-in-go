package controller

import (
	"context"
	"fmt"
	"golang-hospital-management/database"
	"golang-hospital-management/models"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var prescriptionCollection *mongo.Collection = database.OpenCollection(database.Client, "prescription")

func GetPrescriptions() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		result, err := prescriptionCollection.Find(context.TODO(), bson.M{})
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing the prescription"})
		}
		var allPrescriptions []bson.M
		if err = result.All(ctx, &allPrescriptions); err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, allPrescriptions)
	}
}

func GetPrescription() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		prescriptionId := c.Param("prescription_id")
		var prescription models.Prescription

		err := doctorCollection.FindOne(ctx, bson.M{"prescription_id": prescriptionId}).Decode(&prescription)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while fetching the prescription"})
		}
		c.JSON(http.StatusOK, prescription)
	}
}

func CreatePrescription() gin.HandlerFunc {
	return func(c *gin.Context) {
		var prescription models.Prescription
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		if err := c.BindJSON(&prescription); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(prescription)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		prescription.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		prescription.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		prescription.ID = primitive.NewObjectID()
		prescription.Prescription_id = prescription.ID.Hex()

		result, insertErr := prescriptionCollection.InsertOne(ctx, prescription)
		if insertErr != nil {
			msg := fmt.Sprintf("prescription was not created")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		defer cancel()
		c.JSON(http.StatusOK, result)
		defer cancel()
	}
}

func inTimeSpan(start, end, check time.Time) bool {
	return start.After(time.Now()) && end.After(start)
}

func UpdatePrescription() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var prescription models.Prescription

		if err := c.BindJSON(&prescription); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		prescriptionId := c.Param("prescription_id")
		filter := bson.M{"prescription_id": prescriptionId}

		var updateObj primitive.D

		if prescription.Start_Date != nil && prescription.End_Date != nil {
			if !inTimeSpan(*prescription.Start_Date, *prescription.End_Date, time.Now()) {
				msg := "kindly retype the time"
				c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
				defer cancel()
				return
			}

			updateObj = append(updateObj, bson.E{"start_date", prescription.Start_Date})
			updateObj = append(updateObj, bson.E{"end_date", prescription.End_Date})

			if prescription.Drugs != "" {
				updateObj = append(updateObj, bson.E{"Drugs", prescription.Drugs})
			}
			if prescription.Dosage != "" {
				updateObj = append(updateObj, bson.E{"Dosage", prescription.Dosage})
			}

			prescription.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
			updateObj = append(updateObj, bson.E{"updated_at", prescription.Updated_at})

			upsert := true

			opt := options.UpdateOptions{
				Upsert: &upsert,
			}

			result, err := prescriptionCollection.UpdateOne(
				ctx,
				filter,
				bson.D{
					{"$set", updateObj},
				},
				&opt,
			)

			if err != nil {
				msg := "prescription update failed"
				c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			}

			defer cancel()
			c.JSON(http.StatusOK, result)
		}
	}
}
