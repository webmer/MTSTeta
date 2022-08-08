package analytic

import (
	"context"
	"database/sql"
	"gitlab.com/g6834/team26/analytic/internal/domain/models"
	"gitlab.com/g6834/team26/analytic/internal/ports"
	"gitlab.com/g6834/team26/analytic/pkg/api"
	"strconv"
)

type Service struct {
	db   ports.AnalyticDB
	grpc ports.GrpcAuth
}

func New(db ports.AnalyticDB, grpc ports.GrpcAuth) *Service {
	return &Service{
		db:   db,
		grpc: grpc,
	}
}

func (s *Service) ApprovedTasks(ctx context.Context, login string) (int, error) {
	t, err := s.db.GetCountTasksByUserStatus(ctx, login, true)
	if err != nil {
		return 0, err
	}
	return t, nil
}

func (s *Service) DeclinedTasks(ctx context.Context, login string) (int, error) {
	t, err := s.db.GetCountTasksByUserStatus(ctx, login, false)
	if err != nil {
		return 0, err
	}
	return t, nil
}

func (s *Service) TotalTimeTasks(ctx context.Context, login string) ([]*models.Task, error) {
	t, err := s.db.GetTotalTimeTasks(ctx, login)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (s *Service) Validate(tokens models.TokenPair) (*api.AuthResponse, error) {
	grpcResponse, err := s.grpc.Validate(tokens)
	if err != nil {
		return nil, err
	}
	return grpcResponse, nil
}

func (s *Service) MessageIsExist(ctx context.Context, m *models.Message) (bool, error) {
	tm, err := s.db.GetMessage(ctx, m.UUIDMessage)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}

	if tm == nil {
		return false, nil
	}

	return true, nil
}

func (s *Service) ActionTask(ctx context.Context, m *models.Message) error {
	types := map[string]struct{}{"run": {}, "delete": {}, "approve": {}, "complete": {}, "send": {}}

	if _, ok := types[m.Type]; ok {
		ok, err := s.MessageIsExist(ctx, m)
		if err != nil {
			return err
		}

		if !ok {
			switch m.Type {
			case "run":
				t := &models.TaskDB{
					UUID:       m.UUID,
					Login:      m.Value,
					DateCreate: m.Timestamp,
				}
				err := s.db.AddTask(ctx, t)
				if err != nil {
					return err
				}
			case "send", "approve":
				t := &models.TaskDB{
					UUID:       m.UUID,
					DateAction: m.Timestamp,
				}
				err := s.db.DateActionTask(ctx, t)
				if err != nil {
					return err
				}
				return nil
			case "complete", "delete":
				t := &models.TaskDB{
					UUID:       m.UUID,
					DateAction: m.Timestamp,
					//Status:     m.Value,
				}

				v, err := strconv.ParseBool(m.Value)
				if err == nil {
					t.Status = sql.NullBool{
						Bool:  v,
						Valid: true,
					}
				}

				err = s.db.CompleteTask(ctx, t)
				if err != nil {
					return err
				}
			}

			msg := &models.MessageDB{
				UUID:       m.UUIDMessage,
				TaskUUID:   m.UUID,
				DateCreate: m.Timestamp,
				Type:       m.Type,
				Value:      m.Value,
			}
			err := s.db.AddMessage(ctx, msg)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
