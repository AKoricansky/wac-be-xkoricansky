package ambulance_counseling_wl

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"time"

	"github.com/AKoricansky/wac-be-xkoricansky/internal/db_service"
	"github.com/gin-gonic/gin"
)

// Helper function to check if user is a doctor
func isDoctor(c *gin.Context) bool {
	userType, exists := c.Get("userType")
	if !exists {
		return false
	}
	return userType == "doctor"
}

// Helper function to check if user is creator of content
func isCreator(c *gin.Context, creatorId string) bool {
	userId, exists := c.Get("userId")
	if !exists {
		return false
	}
	return userId == creatorId
}

type implAmbulanceCounselingAPI struct {
	questionDbService db_service.DbService[Question]
	replyDbService    db_service.DbService[Reply]
}

func NewAmbulanceCounselingApi(questionDbService db_service.DbService[Question], replyDbService db_service.DbService[Reply]) AmbulanceCounselingAPI {
	return &implAmbulanceCounselingAPI{
		questionDbService: questionDbService,
		replyDbService:    replyDbService,
	}
}

func (o *implAmbulanceCounselingAPI) generateDocumentID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (o *implAmbulanceCounselingAPI) CreateQuestion(c *gin.Context) {
	var question Question
	if err := c.ShouldBindJSON(&question); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid question data"})
		return
	}

	id, err := o.generateDocumentID()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate question ID"})
		return
	}

	question.Id = id
	question.CreatedAt = time.Now()
	question.LastUpdated = time.Now()
	question.RepliedTo = false
	question.Replies = []Reply{}

	ctx := context.Background()
	err = o.questionDbService.CreateDocument(ctx, question.Id, &question)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create question"})
		return
	}

	c.JSON(http.StatusCreated, question)
}

func (o *implAmbulanceCounselingAPI) GetQuestions(c *gin.Context) {
	ctx := context.Background()

	questions, err := o.questionDbService.FindAllDocuments(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve questions"})
		return
	}

	c.JSON(http.StatusOK, questions)
}

func (o *implAmbulanceCounselingAPI) GetQuestionById(c *gin.Context) {
	id := c.Param("questionId")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Question ID is required"})
		return
	}

	ctx := context.Background()
	question, err := o.questionDbService.FindDocument(ctx, id)
	if err != nil {
		if err == db_service.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Question not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, question)
}

func (o *implAmbulanceCounselingAPI) UpdateQuestionById(c *gin.Context) {
	id := c.Param("questionId")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Question ID is required"})
		return
	}

	ctx := context.Background()
	existingQuestion, err := o.questionDbService.FindDocument(ctx, id)
	if err != nil {
		if err == db_service.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Question not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Verify if user is the creator
	if !isCreator(c, existingQuestion.PatientId) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the creator can update this question"})
		return
	}

	if existingQuestion.RepliedTo {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot update question that has been replied to"})
		return
	}

	var updateData Question
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid question data"})
		return
	}

	existingQuestion.Summary = updateData.Summary
	existingQuestion.Question = updateData.Question
	existingQuestion.LastUpdated = time.Now()

	err = o.questionDbService.UpdateDocument(ctx, id, existingQuestion)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update question"})
		return
	}

	c.JSON(http.StatusOK, existingQuestion)
}

func (o *implAmbulanceCounselingAPI) DeleteQuestionById(c *gin.Context) {
	id := c.Param("questionId")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Question ID is required"})
		return
	}

	ctx := context.Background()
	question, err := o.questionDbService.FindDocument(ctx, id)
	if err != nil {
		if err == db_service.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Question not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Verify if user is authorized (must be a doctor or the creator)
	if !isDoctor(c) && !isCreator(c, question.PatientId) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only doctors and the question creator can delete this question"})
		return
	}

	// First, delete all associated replies from the reply collection
	deleteErrors := false
	for _, reply := range question.Replies {
		err = o.replyDbService.DeleteDocument(ctx, reply.Id)
		if err != nil {
			log.Printf("Failed to delete reply %s when deleting question %s: %v", reply.Id, id, err)
			deleteErrors = true
		}
	}

	// Delete the question
	err = o.questionDbService.DeleteDocument(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete question"})
		return
	}

	if deleteErrors {
		c.JSON(http.StatusOK, gin.H{"message": "Question deleted, but some replies could not be deleted"})
		return
	}

	c.Status(http.StatusNoContent)
}

func (o *implAmbulanceCounselingAPI) ReplyToQuestion(c *gin.Context) {
	id := c.Param("questionId")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Question ID is required"})
		return
	}

	// Check if user is authorized (must be a doctor or the creator of the question)
	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	ctx := context.Background()
	question, err := o.questionDbService.FindDocument(ctx, id)
	if err != nil {
		if err == db_service.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Question not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Verify if user is authorized to reply
	if !isDoctor(c) && !isCreator(c, question.PatientId) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only doctors and the question creator can reply"})
		return
	}

	var reply Reply
	if err := c.ShouldBindJSON(&reply); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reply data"})
		return
	}

	replyId, err := o.generateDocumentID()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate reply ID"})
		return
	}

	reply.Id = replyId
	reply.UserId = userId.(string)
	reply.CreatedAt = time.Now()
	reply.RepliedTo = false

	// If replier is a doctor, add their name
	if isDoctor(c) {
		// Use the user's name from the JWT if available, or a default
		doctorName, exists := c.Get("userName")
		if exists {
			reply.DoctorName = doctorName.(string)
		} else {
			reply.DoctorName = "Doctor"
		}
	}

	err = o.replyDbService.CreateDocument(ctx, reply.Id, &reply)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create reply"})
		return
	}

	// Set repliedTo flag on the question
	question.RepliedTo = true
	question.LastUpdated = time.Now()

	// Mark previous replies as replied to (for conversation threading)
	for i := range question.Replies {
		question.Replies[i].RepliedTo = true
	}

	// Add the new reply
	question.Replies = append(question.Replies, reply)

	err = o.questionDbService.UpdateDocument(ctx, id, question)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update question with reply"})
		return
	}

	c.JSON(http.StatusCreated, reply)
}

func (o *implAmbulanceCounselingAPI) GetRepliesByQuestionId(c *gin.Context) {
	questionId := c.Param("questionId")
	if questionId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Question ID is required"})
		return
	}

	ctx := context.Background()
	question, err := o.questionDbService.FindDocument(ctx, questionId)
	if err != nil {
		if err == db_service.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Question not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, question.Replies)
}

func (o *implAmbulanceCounselingAPI) GetReplyById(c *gin.Context) {
	replyId := c.Param("replyId")
	if replyId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Reply ID is required"})
		return
	}

	ctx := context.Background()
	reply, err := o.replyDbService.FindDocument(ctx, replyId)
	if err != nil {
		if err == db_service.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Reply not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, reply)
}

func (o *implAmbulanceCounselingAPI) UpdateReplyById(c *gin.Context) {
	replyId := c.Param("replyId")
	if replyId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Reply ID is required"})
		return
	}

	ctx := context.Background()
	existingReply, err := o.replyDbService.FindDocument(ctx, replyId)
	if err != nil {
		if err == db_service.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Reply not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Verify if user is the creator of the reply
	if !isCreator(c, existingReply.UserId) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the creator can update this reply"})
		return
	}

	if existingReply.RepliedTo {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot update reply that has been replied to"})
		return
	}

	var updateData Reply
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reply data"})
		return
	}

	existingReply.Text = updateData.Text

	err = o.replyDbService.UpdateDocument(ctx, replyId, existingReply)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update reply"})
		return
	}

	questions, err := o.questionDbService.FindDocumentsByField(ctx, "replies._id", replyId)
	if err == nil && len(questions) > 0 {
		for _, question := range questions {
			for i, reply := range question.Replies {
				if reply.Id == replyId {
					question.Replies[i] = *existingReply
					err = o.questionDbService.UpdateDocument(ctx, question.Id, question)
					if err != nil {
						log.Printf("Failed to update parent question %s: %v", question.Id, err)
					}
					break
				}
			}
		}
	}

	c.JSON(http.StatusOK, existingReply)
}

func (o *implAmbulanceCounselingAPI) DeleteReplyById(c *gin.Context) {
	replyId := c.Param("replyId")
	if replyId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Reply ID is required"})
		return
	}

	ctx := context.Background()
	existingReply, err := o.replyDbService.FindDocument(ctx, replyId)
	if err != nil {
		if err == db_service.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Reply not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Verify if user is the creator of the reply
	if !isCreator(c, existingReply.UserId) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the creator can delete this reply"})
		return
	}

	if existingReply.RepliedTo {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot delete reply that has been replied to"})
		return
	}

	questions, err := o.questionDbService.FindDocumentsByField(ctx, "replies._id", replyId)

	err = o.replyDbService.DeleteDocument(ctx, replyId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete reply"})
		return
	}

	if err == nil && len(questions) > 0 {
		for _, question := range questions {
			var updatedReplies []Reply
			for _, reply := range question.Replies {
				if reply.Id != replyId {
					updatedReplies = append(updatedReplies, reply)
				}
			}

			question.Replies = updatedReplies

			err = o.questionDbService.UpdateDocument(ctx, question.Id, question)
			if err != nil {
				log.Printf("Failed to update parent question %s after reply deletion: %v", question.Id, err)
			}
		}
	}

	c.Status(http.StatusNoContent)
}
