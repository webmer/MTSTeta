package task

import (
	"context"
	"sync"
	"time"

	"gitlab.com/g6834/team26/task/internal/domain/models"
	"gitlab.com/g6834/team26/task/internal/ports"
	"gitlab.com/g6834/team26/task/pkg/api"
)

type Service struct {
	db             ports.TaskDB
	grpcAuth       ports.GrpcAuth
	analyticSender ports.TaskAnalyticSender
	emailSender    ports.EmailSender
	wg             *sync.WaitGroup
}

func New(db ports.TaskDB, grpcAuth ports.GrpcAuth, analyticSender ports.TaskAnalyticSender, emailSender ports.EmailSender) *Service {
	return &Service{
		db:             db,
		grpcAuth:       grpcAuth,
		analyticSender: analyticSender,
		emailSender:    emailSender,
		wg:             &sync.WaitGroup{},
	}
}

func (s *Service) Stop() error {
	s.wg.Wait()
	return nil
}

func (s *Service) ListTasks(ctx context.Context, login string) ([]*models.Task, error) {
	t, err := s.db.List(ctx, login)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (s *Service) RunTask(ctx context.Context, createdTask *models.Task) error {
	err := s.db.Run(ctx, createdTask)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) UpdateTask(ctx context.Context, id, login, name, text string) error {
	err := s.db.Update(ctx, id, login, name, text)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) DeleteTask(ctx context.Context, login, id string) error {
	err := s.db.Delete(ctx, login, id)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) ApproveTask(ctx context.Context, login, id, approvalLogin string) error {
	err := s.db.Approve(ctx, login, id, approvalLogin)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) DeclineTask(ctx context.Context, login, id, approvalLogin string) error {
	err := s.db.Decline(ctx, login, id, approvalLogin)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) Validate(ctx context.Context, tokens ports.TokenPair) (*api.AuthResponse, error) {
	grpcResponse, err := s.grpcAuth.Validate(ctx, tokens)
	if err != nil {
		return nil, err
	}
	return grpcResponse, nil
}

func (s *Service) StartMessageSender(ctx context.Context) {
	for {
		// log.Println("sleeping for 60 seconds")
		time.Sleep(60 * time.Second) // TODO: уточнить, может есть возможность получения уведомления от postgresql о внесении новых данных в БД
		// log.Println("looking for messages to send")
		messages, _ := s.db.GetMessagesToSend(ctx)
		// log.Println(messages)
		for id, message := range messages {
			err := s.analyticSender.ActionTask(ctx, message)
			if err != nil {
				continue
			}
			err = s.db.UpdateMessageStatus(ctx, id)
			if err != nil {
				continue
			}
		}
	}
}

func (s *Service) GetResultOfEmailSending(ctx context.Context) {
	defer s.wg.Done()
	// log.Println("Reader Started!!!")
	resChan := s.emailSender.GetEmailResultChan()
	for result := range resChan {
		// log.Printf("Get result of sending - %v", result)
		// time.Sleep(5 * time.Second)
		for email, res := range result {
			err := s.db.ChangeEmailStatusAndSendMessage(ctx, email, res)
			if err != nil {
				continue
			}
		}
	}
	// log.Println("result channel closed")
}

func (s *Service) StartEmailSender(ctx context.Context) {
	s.emailSender.StartEmailWorkers(ctx)
	s.wg.Add(1)
	go s.GetResultOfEmailSending(ctx)

	for {
		time.Sleep(60 * time.Second) // TODO: уточнить, может есть возможность получения уведомления от postgresql о внесении новых данных в БД
		emails, _ := s.db.GetEmailsToSend(ctx)
		for _, email := range emails {
			s.emailSender.PushEmailToChan(email)
		}
		// log.Println(emails)
	}
}
