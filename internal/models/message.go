package models

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Message struct {
	ChatID    int64
	MessageID int
}

type MessageModel struct {
	Pool *pgxpool.Pool
}

func (m *MessageModel) AddMessage(chatID int64, messageID int) error {
	rows, err := m.Pool.Query(context.Background(),
		`insert into messages (chat_id, message_id) values ($1, $2);`,
		chatID, messageID,
	)
	if err != nil {
		return err
	}
	defer rows.Close()
	return nil
}

func (m *MessageModel) DeleteExpired() ([]*Message, error) {
	rows, err := m.Pool.Query(context.Background(),
		`delete from messages
		where created < now() - interval '5 minutes'
		returning chat_id, message_id;`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var messages []*Message
	for rows.Next() {
		var chatID int64
		var messageID int
		if err := rows.Scan(&chatID, &messageID); err != nil {
			return nil, err
		}
		messages = append(messages, &Message{ChatID: chatID, MessageID: messageID})
	}
	return messages, nil
}
