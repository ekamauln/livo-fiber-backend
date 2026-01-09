package models

import "time"

type Channel struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	ChannelCode string    `gorm:"uniqueIndex;not null;type:varchar(50)" json:"channel_code"`
	ChannelName string    `gorm:"not null;type:varchar(100)" json:"channel_name"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ChannelResponse represents the channel data returned in API responses
type ChannelResponse struct {
	ID          uint   `json:"id"`
	ChannelCode string `json:"channelCode"`
	ChannelName string `json:"channelName"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

// ToResponse converts a Channel model to a ChannelResponse
func (ch *Channel) ToResponse() *ChannelResponse {
	return &ChannelResponse{
		ID:          ch.ID,
		ChannelCode: ch.ChannelCode,
		ChannelName: ch.ChannelName,
		CreatedAt:   ch.CreatedAt.Format("02-01-2006 15:04:05"),
		UpdatedAt:   ch.UpdatedAt.Format("02-01-2006 15:04:05"),
	}
}
