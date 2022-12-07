package routes

import (
	controller "golang-hospital-management/controllers"

	"github.com/gin-gonic/gin"
)

func PatientRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.GET("/patients", controller.GetPatients())
	incomingRoutes.GET("/patients/:patient_id", controller.GetPatient())
	incomingRoutes.POST("/patients/signup", controller.SignUp())
	incomingRoutes.POST("/patients/login", controller.Login())
}