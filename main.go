package main

import (
	"os"

	"golang-hospital-management/database"
	middleware "golang-hospital-management/middleware"
	routes "golang-hospital-management/routes"

	"github.com/gin-gonic/gin"

	"go.mongodb.org/mongo-driver/mongo"
)

var patientCollection *mongo.Collection = database.OpenCollection(database.Client, "patient")

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		port = "8001"
	}
	router := gin.New()
	router.Use(gin.Logger())
	routes.PatientRoutes(router)
	router.Use(middleware.Authentication())

	routes.DoctorRoutes(router)
	routes.PrescriptionRoutes(router)
	routes.BookappointmentRoutes(router)
	routes.InvoiceRoutes(router)

	router.Run(":" + port)
}
