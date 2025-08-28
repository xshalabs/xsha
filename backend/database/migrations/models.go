package migrations

import (
	"time"
)

// Migration represents a database migration record
type Migration struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"column:name;unique;not null" json:"name"`
	AppliedAt time.Time `gorm:"column:applied_at;not null" json:"applied_at"`
}

// Task represents a task in the system
type Task struct {
	ID            uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	WorkspacePath string    `gorm:"column:workspace_path" json:"workspace_path"`
	CreatedAt     time.Time `gorm:"column:created_at;not null" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at;not null" json:"updated_at"`
}

// DevEnvironment represents a development environment
type DevEnvironment struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	SessionDir string    `gorm:"column:session_dir" json:"session_dir"`
	CreatedBy  string    `gorm:"column:created_by" json:"created_by"`
	AdminID    *uint     `gorm:"column:admin_id" json:"admin_id"`
	CreatedAt  time.Time `gorm:"column:created_at;not null" json:"created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at;not null" json:"updated_at"`
}

// SystemConfig represents a system configuration entry
type SystemConfig struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	ConfigKey   string    `gorm:"column:config_key;unique;not null" json:"config_key"`
	ConfigValue string    `gorm:"column:config_value" json:"config_value"`
	CreatedAt   time.Time `gorm:"column:created_at;not null" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;not null" json:"updated_at"`
}

// TaskConversationAttachment represents an attachment in a task conversation
type TaskConversationAttachment struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedBy string    `gorm:"column:created_by" json:"created_by"`
	AdminID   *uint     `gorm:"column:admin_id" json:"admin_id"`
	CreatedAt time.Time `gorm:"column:created_at;not null" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null" json:"updated_at"`
}

// AdminOperationLog represents an admin operation log entry
type AdminOperationLog struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Username  string    `gorm:"column:username" json:"username"`
	AdminID   *uint     `gorm:"column:admin_id" json:"admin_id"`
	CreatedAt time.Time `gorm:"column:created_at;not null" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null" json:"updated_at"`
}

// Admin represents an admin user
type Admin struct {
	ID       uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Username string `gorm:"column:username;unique;not null" json:"username"`
}