package database

import (
	"context"
	"errors"

	"github.com/getaddrinfo/proxy-fingerprint-scraper/common"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
)

var ErrorNotListening = errors.New("database chan is not being listened to")

type Database struct {
	Conn *pgxpool.Pool
	ctx  context.Context
	log  *zap.Logger
}

func NewDatabase(ctx context.Context, url string) (*Database, error) {
	log := zap.L().Named("db")

	conn, err := pgxpool.Connect(ctx, url)

	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	log.Info("connected")

	return &Database{Conn: conn, ctx: ctx, log: log}, nil
}

func (db *Database) ListenForNewFingerprints(channel common.FingerprintResultChannel) {
Iter:
	for {
		select {
		case <-db.ctx.Done():
			close(channel)
			break Iter

		case r := <-channel:
			added, err := db.AddFingerprint(r.Fingerprint, r.ProxyIP)

			if err != nil {
				db.log.Sugar().Errorf("error: %s", err.Error())
			}

			db.log.Debug("completed process", zap.Bool("success", added))
		}
	}
}
