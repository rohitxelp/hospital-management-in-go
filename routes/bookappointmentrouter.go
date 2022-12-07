package routes

import (
	controller "golang-hospital-management/controllers"

	"github.com/gin-gonic/gin"
)

func BookappointmentRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.GET("/appoinments", controller.GetAppoinments())
	incomingRoutes.GET("/appoinment/:appointment_id", controller.GetAppoinment())
	
	incomingRoutes.POST("/appointment", controller.CreateAppointment())
	incomingRoutes.PATCH("/appointment/:appointment_id", controller.UpdateAppointment())
}