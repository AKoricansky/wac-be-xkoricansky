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
	existingQuestion, err := o.questionDbService.FindDocument(ctx, id)
	if err != nil {
		if err == db_service.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Question not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Delete the question
	err = o.questionDbService.DeleteDocument(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete question"})
		return
	}

	c.Status(http.StatusNoContent)
}

func (o *implAmbulanceCounselingAPI) ReplyToQuestion(c *gin.Context) {
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
	reply.CreatedAt = time.Now()
	reply.RepliedTo = false

	err = o.replyDbService.CreateDocument(ctx, reply.Id, &reply)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create reply"})
		return
	}

	question.RepliedTo = true
	question.LastUpdated = time.Now()
	question.Replies = append(question.Replies, reply)

	err = o.questionDbService.UpdateDocument(ctx, questionId, question)
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

	questions, err := o.questionDbService.FindDocumentsByField(ctx, "replies.id", replyId)
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

	if existingReply.RepliedTo {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot delete reply that has been replied to"})
		return
	}

	questions, err := o.questionDbService.FindDocumentsByField(ctx, "replies.id", replyId)

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
