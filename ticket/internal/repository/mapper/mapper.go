package mapper

import (
	"ticket/internal/model"
	repositorymodel "ticket/internal/repository/model"
)

func ToRepo(ticket *model.Ticket) *repositorymodel.Ticket {
	return &repositorymodel.Ticket{
		ID:            ticket.ID,
		Title:         ticket.Title,
		Description:   ticket.Description,
		Status:        string(ticket.Status),
		CurrentNodeID: ticket.CurrentNodeID,
		Solution:      ticket.Solution,
		CreatedAt:     ticket.CreatedAt,
		UpdatedAt:     ticket.UpdatedAt,
	}
}

func ToDomain(ticket *repositorymodel.Ticket) *model.Ticket {
	return &model.Ticket{
		ID:            ticket.ID,
		Title:         ticket.Title,
		Description:   ticket.Description,
		Status:        model.Status(ticket.Status),
		CurrentNodeID: ticket.CurrentNodeID,
		Solution:      ticket.Solution,
		CreatedAt:     ticket.CreatedAt,
		UpdatedAt:     ticket.UpdatedAt,
	}
}

func ToDomainList(tickets []repositorymodel.Ticket) []model.Ticket {
	result := make([]model.Ticket, 0, len(tickets))
	for i := range tickets {
		result = append(result, *ToDomain(&tickets[i]))
	}
	return result
}
