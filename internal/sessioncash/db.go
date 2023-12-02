package sessioncash

import (
	"context"

	"accelerator/internal/session"
	"github.com/tarantool/go-tarantool/v2"
	"github.com/tarantool/go-tarantool/v2/datetime"
)

type CashDb interface {
	StoreSession(s *session.Session) error
	FindSession(token string) (session.Session, error)
	DeleteSession(token string) error
}

type SessionSerialized struct {
	_msgpack struct{} `msgpack:",asArray"`
	Token    string
	ExpTime  datetime.Datetime
}

type TarantoolCashDb struct {
	space *tarantool.Space
	conn  *tarantool.Connection
}

func NewTarantoolCashDB(host, user, password, name string, opts tarantool.Opts) (CashDb, error) {
	dialer := tarantool.NetDialer{Address: host, User: user, Password: password}
	conn, err := tarantool.Connect(context.Background(), dialer, opts)
	if err != nil {
		return nil, err
	}
	space := tarantool.Space{Name: name, Id: 1}
	return &TarantoolCashDb{conn: conn, space: &space}, nil
}

func (t *TarantoolCashDb) StoreSession(s *session.Session) error {
	dt, err := datetime.MakeDatetime(s.ExpTime)
	if err != nil {
		return err
	}
	toinsert := SessionSerialized{Token: s.Token, ExpTime: dt}
	_, err = t.conn.Do(tarantool.NewInsertRequest(t.space.Id).Tuple(toinsert)).Get()
	return err
}

func (t *TarantoolCashDb) FindSession(token string) (session.Session, error) {
	resp, err := t.conn.Do(tarantool.NewSelectRequest(t.space.Id).Key(token)).Get()
	if err != nil {
		return session.Session{}, err
	}
	d := resp.Data[1].(datetime.Datetime)
	res := session.Session{Token: resp.Data[0].(string), ExpTime: d.ToTime()}
	return res, nil
}

func (t *TarantoolCashDb) DeleteSession(token string) error {
	_, err := t.conn.Do(tarantool.NewDeleteRequest(t.space.Id).Key(token)).Get()
	return err
}
