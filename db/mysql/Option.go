package mysql

import "time"

type options struct {
	dblog       bool // json 或者 stdout
	maxIdleConn int
	maxOpenConn int
	maxLifetime time.Duration
}

type Option func(*options)

func WithDbLog(dblog bool) Option {
	return func(o *options) {
		o.dblog = dblog
	}
}

func WithMaxIdleConn(max int) Option {
	return func(o *options) {
		o.maxIdleConn = max
	}
}

func WithMaxOpenConn(max int) Option {
	return func(o *options) {
		o.maxOpenConn = max
	}
}

func WithMaxLifetime(max time.Duration) Option {
	return func(o *options) {
		o.maxLifetime = max
	}
}
