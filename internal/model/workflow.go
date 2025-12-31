// Package model defines the workflow models for the application.
package model

import (
	"time"

	"gorm.io/gorm"
)

// Workflow represents a workflow definition.
type Workflow struct {
	BaseModel
	Name        string `gorm:"size:100;not null;uniqueIndex" json:"name"`
	DisplayName string `gorm:"size:200" json:"display_name"`
	Description string `gorm:"type:text" json:"description"`
	Type        string `gorm:"size:50;not null" json:"type"`     // e.g., approval, provisioning
	Status      int    `gorm:"not null;default:1" json:"status"` // 0: disabled, 1: enabled
	Steps       string `gorm:"type:text" json:"steps"`           // JSON array of workflow steps
	CreatedBy   string `gorm:"size:36" json:"created_by"`
	Creator     *User  `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

// TableName specifies the table name for Workflow.
func (Workflow) TableName() string {
	return "workflows"
}

// BeforeCreate hook for Workflow.
func (w *Workflow) BeforeCreate(tx *gorm.DB) error {
	return w.BaseModel.BeforeCreate(tx)
}

// WorkflowInstance represents an instance of a running workflow.
type WorkflowInstance struct {
	BaseModel
	WorkflowID  string     `gorm:"size:36;not null;index" json:"workflow_id"`
	Workflow    *Workflow  `gorm:"foreignKey:WorkflowID" json:"workflow,omitempty"`
	Status      string     `gorm:"size:20;not null;default:'pending'" json:"status"` // pending, running, completed, failed, canceled
	CurrentStep int        `gorm:"not null;default:0" json:"current_step"`
	TotalSteps  int        `gorm:"not null" json:"total_steps"`
	Input       string     `gorm:"type:text" json:"input"`  // JSON
	Output      string     `gorm:"type:text" json:"output"` // JSON
	InitiatorID string     `gorm:"size:36;not null;index" json:"initiator_id"`
	Initiator   *User      `gorm:"foreignKey:InitiatorID" json:"initiator,omitempty"`
	StartedAt   *time.Time `gorm:"" json:"started_at"`
	CompletedAt *time.Time `gorm:"" json:"completed_at"`
	Error       string     `gorm:"type:text" json:"error"`
}

// TableName specifies the table name for WorkflowInstance.
func (WorkflowInstance) TableName() string {
	return "workflow_instances"
}

// BeforeCreate hook for WorkflowInstance.
func (wi *WorkflowInstance) BeforeCreate(tx *gorm.DB) error {
	return wi.BaseModel.BeforeCreate(tx)
}

// WorkflowStep represents a step in a workflow instance.
type WorkflowStep struct {
	BaseModel
	InstanceID  string            `gorm:"size:36;not null;index" json:"instance_id"`
	Instance    *WorkflowInstance `gorm:"foreignKey:InstanceID" json:"instance,omitempty"`
	StepNumber  int               `gorm:"not null" json:"step_number"`
	StepName    string            `gorm:"size:100;not null" json:"step_name"`
	StepType    string            `gorm:"size:50;not null" json:"step_type"`                // approval, action, notification
	Status      string            `gorm:"size:20;not null;default:'pending'" json:"status"` // pending, in_progress, completed, failed, skipped
	AssigneeID  string            `gorm:"size:36" json:"assignee_id"`
	Assignee    *User             `gorm:"foreignKey:AssigneeID" json:"assignee,omitempty"`
	Input       string            `gorm:"type:text" json:"input"`  // JSON
	Output      string            `gorm:"type:text" json:"output"` // JSON
	StartedAt   *time.Time        `gorm:"" json:"started_at"`
	CompletedAt *time.Time        `gorm:"" json:"completed_at"`
	Comment     string            `gorm:"type:text" json:"comment"`
}

// TableName specifies the table name for WorkflowStep.
func (WorkflowStep) TableName() string {
	return "workflow_steps"
}

// BeforeCreate hook for WorkflowStep.
func (ws *WorkflowStep) BeforeCreate(tx *gorm.DB) error {
	return ws.BaseModel.BeforeCreate(tx)
}
