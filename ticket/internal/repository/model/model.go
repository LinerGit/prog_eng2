package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type Ticket struct {
	ID            string     `gorm:"column:id;primaryKey"`
	Title         string     `gorm:"column:title"`
	Description   string     `gorm:"column:description"`
	Status        string     `gorm:"column:status"`
	CurrentNodeID string     `gorm:"column:current_node_id"`
	Solution      string     `gorm:"column:solution"`
	CreatedAt     time.Time  `gorm:"column:created_at"`
	UpdatedAt     time.Time  `gorm:"column:updated_at"`
	DeletedAt     *time.Time `gorm:"column:deleted_at"`
}

func (Ticket) TableName() string {
	return "tickets"
}

type Option struct {
	Text       string `json:"text"`
	NextNodeID string `json:"next_node_id"`
}

type Options []Option

func (o Options) Value() (driver.Value, error) {
	if o == nil {
		return "[]", nil
	}
	data, err := json.Marshal(o)
	if err != nil {
		return nil, err
	}
	return string(data), nil
}

func (o *Options) Scan(value any) error {
	if value == nil {
		*o = Options{}
		return nil
	}

	var data []byte
	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return fmt.Errorf("unsupported options type %T", value)
	}

	if len(data) == 0 {
		*o = Options{}
		return nil
	}
	return json.Unmarshal(data, o)
}

type DecisionNode struct {
	ID        string    `gorm:"column:id;primaryKey"`
	Text      string    `gorm:"column:text"`
	Options   Options   `gorm:"column:options;type:jsonb"`
	Solution  string    `gorm:"column:solution"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (DecisionNode) TableName() string {
	return "decision_nodes"
}
