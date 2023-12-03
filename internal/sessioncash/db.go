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
	_, err = t.conn.Do(tarantool.NewInsertRequest(t.space.Name).Tuple(toinsert)).Get()
	return err
}

func (t *TarantoolCashDb) FindSession(token string) (session.Session, error) {
	const index = "primary"
	resp, err := t.conn.Do(tarantool.NewSelectRequest(t.space.Name).Index(index).Iterator(tarantool.IterEq).Key(tarantool.StringKey{S: token})).Get()
	if err != nil || len(resp.Data) < 1 || len(resp.Data[0].([]interface{})) < 2 {
		return session.Session{}, err
	}
	values := resp.Data[0].([]interface{})
	d := values[1].(datetime.Datetime)
	res := session.Session{Token: values[0].(string), ExpTime: d.ToTime()}
	return res, nil
}

func (t *TarantoolCashDb) DeleteSession(token string) error {
	const index = "primary"
	_, err := t.conn.Do(tarantool.NewDeleteRequest(t.space.Name).Index(index).Key(tarantool.StringKey{S: token})).Get()
	return err
}
