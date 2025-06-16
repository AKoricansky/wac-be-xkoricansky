package ambulance_counseling_wl

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type implAmbulanceCounselingAuthAPI struct {
}

func NewAmbulanceCounselingAuthApi() AmbulanceCounselingAuthAPI {
	return &implAmbulanceCounselingAuthAPI{}
}

func (o *implAmbulanceCounselingAuthAPI) UserLogin(c *gin.Context) {
	c.AbortWithStatus(http.StatusNotImplemented)
}

func (o *implAmbulanceCounselingAuthAPI) UserRegister(c *gin.Context) {
	c.AbortWithStatus(http.StatusNotImplemented)
}