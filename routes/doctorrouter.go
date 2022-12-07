package routes

import (
	controller "golang-hospital-management/controllers"

	"github.com/gin-gonic/gin"
)

func DoctorRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.GET("/doctors", controller.GetDoctors())
	incomingRoutes.GET("/doctors/:doctor_id", controller.GetDoctor())
	incomingRoutes.POST("/doctors", controller.CreateDoctor())
	incomingRoutes.PATCH("/doctors/:doctor_id", controller.UpdateDoctor())
}