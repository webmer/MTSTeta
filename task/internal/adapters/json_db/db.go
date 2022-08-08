package json_db

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	e "gitlab.com/g6834/team26/task/internal/domain/errors"
	"gitlab.com/g6834/team26/task/internal/domain/models"
)

type JsonDatabase struct {
	File *os.File
}

func New(filepath string) (*JsonDatabase, error) {
	if _, err := os.Stat(filepath); errors.Is(err, os.ErrNotExist) {
		f, err := os.Create(filepath)
		if err != nil {
			return nil, err
		}
		return &JsonDatabase{File: f}, nil
	} else {
		f, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
		if err != nil {
			return nil, err
		}
		return &JsonDatabase{File: f}, nil
	}
	// defer f.Close()

}

func (jdb *JsonDatabase) clearFile() error {
	defer jdb.File.Seek(0, io.SeekStart)
	f, err := os.Create(jdb.File.Name())
	if err != nil {
		return err
	}
	jdb.File = f
	return nil
}

func (jdb *JsonDatabase) List(login string) ([]*models.Task, error) {
	defer jdb.File.Seek(0, io.SeekStart)

	ts := []*models.Task{}
	fileScanner := bufio.NewScanner(jdb.File)
	fileScanner.Split(bufio.ScanLines)
	for fileScanner.Scan() {
		t := models.Task{}
		json.Unmarshal([]byte(fileScanner.Text()), &t)
		if t.InitiatorLogin == login {
			ts = append(ts, &t)
		}
	}
	return ts, nil
}

func (jdb *JsonDatabase) Run(t *models.Task) error {
	defer jdb.File.Seek(0, io.SeekStart)

	newLine, err := json.Marshal(t)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(jdb.File, string(newLine))
	if err != nil {
		return err
	}
	return nil
}

func (jdb *JsonDatabase) Delete(login, id string) error {
	defer jdb.File.Seek(0, io.SeekStart)
	var idFound bool

	ts, err := jdb.List("")
	if err != nil {
		return err
	}
	for idx, t := range ts {
		if t.UUID == id {
			idFound = true
			if idx == len(ts) {
				ts = ts[:idx]
			} else {
				ts = append(ts[:idx], ts[idx+1:]...)
			}
			break
		}
	}

	if !idFound {
		return e.ErrIdNotFound
	}

	err = jdb.clearFile()
	if err != nil {
		return err
	}

	for _, t := range ts {
		newLine, err := json.Marshal(t)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(jdb.File, string(newLine))
		if err != nil {
			return err
		}
	}

	return nil
}

func (jdb *JsonDatabase) Approve(login, id, approvalLogin string) error {
	defer jdb.File.Seek(0, io.SeekStart)
	var idFound bool
	var loginFound bool

	ts, err := jdb.List("")
	if err != nil {
		return err
	}

LOOP:
	for _, t := range ts {
		log.Println(t)
		if t.UUID == id {
			idFound = true
			for _, a := range t.Approvals {
				if a.ApprovalLogin == approvalLogin {
					loginFound = true
					a.ChangeApprovedStatus(true)
					break LOOP
				}
			}
		}
	}

	if !idFound {
		return e.ErrIdNotFound
	}
	if !loginFound {
		return e.ErrLoginNotFoundInApprovals
	}

	err = jdb.clearFile()
	if err != nil {
		return err
	}

	for _, t := range ts {
		newLine, err := json.Marshal(t)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(jdb.File, string(newLine))
		if err != nil {
			return err
		}
	}

	return nil
}

func (jdb *JsonDatabase) Decline(login, id, approvalLogin string) error {
	defer jdb.File.Seek(0, io.SeekStart)
	var idFound bool
	var loginFound bool

	ts, err := jdb.List("")
	if err != nil {
		return err
	}
LOOP:
	for _, t := range ts {
		if t.UUID == id {
			idFound = true
			for _, a := range t.Approvals {
				if a.ApprovalLogin == approvalLogin {
					loginFound = true
					a.ChangeApprovedStatus(false)
					break LOOP
				}
			}
		}
	}

	if !idFound {
		return e.ErrIdNotFound
	}
	if !loginFound {
		return e.ErrLoginNotFoundInApprovals
	}

	err = jdb.clearFile()
	if err != nil {
		return err
	}

	for _, t := range ts {
		newLine, err := json.Marshal(t)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(jdb.File, string(newLine))
		if err != nil {
			return err
		}
	}

	return nil
}
