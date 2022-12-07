package routes

import (
	controller "golang-hospital-management/controllers"

	"github.com/gin-gonic/gin"
)

func PrescriptionRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.GET("/prescriptions", controller.GetPrescriptions())
	incomingRoutes.GET("/prescription/:prescription_id", controller.GetPrescription())
	incomingRoutes.POST("/precription", controller.CreatePrescription())
	incomingRoutes.PATCH("/prescription/:prescription_id", controller.UpdatePrescription())
}