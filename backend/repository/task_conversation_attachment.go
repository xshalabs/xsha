package repository

import (
	"xsha-backend/database"

	"gorm.io/gorm"
)

type taskConversationAttachmentRepository struct {
	db *gorm.DB
}

func NewTaskConversationAttachmentRepository(db *gorm.DB) TaskConversationAttachmentRepository {
	return &taskConversationAttachmentRepository{db: db}
}

func (r *taskConversationAttachmentRepository) Create(attachment *database.TaskConversationAttachment) error {
	return r.db.Create(attachment).Error
}

func (r *taskConversationAttachmentRepository) GetByID(id uint) (*database.TaskConversationAttachment, error) {
	var attachment database.TaskConversationAttachment
	err := r.db.Preload("Conversation").First(&attachment, id).Error
	if err != nil {
		return nil, err
	}
	return &attachment, nil
}

func (r *taskConversationAttachmentRepository) GetByConversationID(conversationID uint) ([]database.TaskConversationAttachment, error) {
	var attachments []database.TaskConversationAttachment
	err := r.db.Where("conversation_id = ?", conversationID).Order("sort_order, created_at").Find(&attachments).Error
	if err != nil {
		return nil, err
	}
	return attachments, nil
}

func (r *taskConversationAttachmentRepository) GetUnassociated(createdBy string) ([]database.TaskConversationAttachment, error) {
	var attachments []database.TaskConversationAttachment
	err := r.db.Where("conversation_id IS NULL AND created_by = ?", createdBy).Order("created_at DESC").Find(&attachments).Error
	if err != nil {
		return nil, err
	}
	return attachments, nil
}

func (r *taskConversationAttachmentRepository) Update(attachment *database.TaskConversationAttachment) error {
	return r.db.Save(attachment).Error
}

func (r *taskConversationAttachmentRepository) Delete(id uint) error {
	return r.db.Delete(&database.TaskConversationAttachment{}, id).Error
}

func (r *taskConversationAttachmentRepository) DeleteByConversationID(conversationID uint) error {
	return r.db.Where("conversation_id = ?", conversationID).Delete(&database.TaskConversationAttachment{}).Error
}
