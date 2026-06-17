package model

import "time"

type Status string

const (
	StatusOpen       Status = "open"
	StatusInProgress Status = "in_progress"
	StatusClosed     Status = "closed"
)

type Ticket struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	Status        Status    `json:"status"`
	CurrentNodeID string    `json:"current_node_id"`
	Solution      string    `json:"solution,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type CreateTicketRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type UpdateTicketRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      Status `json:"status"`
}

type ChooseProblemRequest struct {
	NextNodeID string `json:"next_node_id"`
}

type CreateDecisionNodeRequest struct {
	ID        string `json:"id"`
	Text      string `json:"text"`
	Solution  string `json:"solution,omitempty"`
	ParentID  string `json:"parent_id,omitempty"`
	OptionText string `json:"option_text,omitempty"`
}

type AddDecisionOptionRequest struct {
	Text       string `json:"text"`
	NextNodeID string `json:"next_node_id"`
}

type UpdateDecisionSolutionRequest struct {
	Solution string `json:"solution"`
}

type DecisionResponse struct {
	Ticket *Ticket `json:"ticket"`
	Node   Node    `json:"node"`
	Done   bool    `json:"done"`
}

type Option struct {
	Text       string `json:"text"`
	NextNodeID string `json:"next_node_id"`
}

type Node struct {
	ID       string   `json:"id"`
	Text     string   `json:"text"`
	Options  []Option `json:"options,omitempty"`
	Solution string   `json:"solution,omitempty"`
}
