package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/segmentio/kafka-go"
	"gitlab.com/g6834/team26/analytic/internal/domain/models"
	"gitlab.com/g6834/team26/analytic/internal/ports"
	"gitlab.com/g6834/team26/analytic/pkg/config"
	"time"
)

type Consumer struct {
	KafkaRead *kafka.Reader
	ac        ports.Analytic
	l         *zerolog.Logger
}

func New(logger *zerolog.Logger, analytic ports.Analytic, c *config.Config) (*Consumer, error) {
	if len(c.Server.KafkaAnalyticTopic) > 0 && len(c.Server.KafkaUrl) > 0 && len(c.Server.KafkaGroupId) > 0 {
		r := kafka.NewReader(kafka.ReaderConfig{
			Brokers:     []string{c.Server.KafkaUrl},
			Topic:       c.Server.KafkaAnalyticTopic,
			GroupID:     c.Server.KafkaGroupId,
			StartOffset: kafka.FirstOffset,
		})

		return &Consumer{KafkaRead: r, ac: analytic, l: logger}, nil
	}

	return nil, fmt.Errorf("empty config var")
}

func (ks *Consumer) StartRead() error {
	for {
		ctx := context.Background()

		m, err := ks.KafkaRead.FetchMessage(ctx)
		if err != nil {
			return err
		}

		messKafka := models.KafkaMessage{}

		err = json.Unmarshal(m.Value, &messKafka)
		if err != nil {
			ks.l.Error().Msgf("cannot unmarshal message: %s", err)
		}

		mess := models.Message{
			UUIDMessage: string(m.Key),
			UUID:        messKafka.UUID,
			Timestamp:   time.Unix(messKafka.Timestamp, 0),
			Type:        messKafka.Type,
			Value:       messKafka.Value,
		}

		err = ks.ac.ActionTask(ctx, &mess)
		if err != nil {
			ks.l.Error().Msgf("error action: %s", err)
		}

		if err = ks.KafkaRead.CommitMessages(ctx, m); err != nil {
			ks.l.Error().Msgf("error commit message: %s", err)
		}
	}
}

func (ks *Consumer) StopRead(ctx context.Context) error {
	return ks.KafkaRead.Close()
}
