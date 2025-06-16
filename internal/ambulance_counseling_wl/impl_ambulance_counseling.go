package ambulance_counseling_wl

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type implAmbulanceCounselingAPI struct {
}

func NewAmbulanceCounselingApi() AmbulanceCounselingAPI {
	return &implAmbulanceCounselingAPI{}
}

func (o *implAmbulanceCounselingAPI) CreateQuestion(c *gin.Context) {
	c.AbortWithStatus(http.StatusNotImplemented)
}

func (o *implAmbulanceCounselingAPI) DeleteQuestionById(c *gin.Context) {
	c.AbortWithStatus(http.StatusNotImplemented)
}

func (o *implAmbulanceCounselingAPI) DeleteReplyById(c *gin.Context) {
	c.AbortWithStatus(http.StatusNotImplemented)
}

func (o *implAmbulanceCounselingAPI) GetQuestionById(c *gin.Context) {
	c.AbortWithStatus(http.StatusNotImplemented)
}

func (o *implAmbulanceCounselingAPI) GetQuestions(c *gin.Context) {
	c.AbortWithStatus(http.StatusNotImplemented)
}

func (o *implAmbulanceCounselingAPI) GetRepliesByQuestionId(c *gin.Context) {
	c.AbortWithStatus(http.StatusNotImplemented)
}

func (o *implAmbulanceCounselingAPI) GetReplyById(c *gin.Context) {
	c.AbortWithStatus(http.StatusNotImplemented)
}

func (o *implAmbulanceCounselingAPI) ReplyToQuestion(c *gin.Context) {
	c.AbortWithStatus(http.StatusNotImplemented)
}

func (o *implAmbulanceCounselingAPI) UpdateQuestionById(c *gin.Context) {
	c.AbortWithStatus(http.StatusNotImplemented)
}

func (o *implAmbulanceCounselingAPI) UpdateReplyById(c *gin.Context) {
	c.AbortWithStatus(http.StatusNotImplemented)
}