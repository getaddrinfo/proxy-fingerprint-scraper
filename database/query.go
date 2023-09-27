package database

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v4"
)

const defaultToken = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

var ErrDefaultToken = errors.New("change the default admin token")

func (db *Database) AddFingerprint(fp string, ip string) (bool, error) {
	result, err := db.Conn.Exec(context.Background(), "INSERT INTO fingerprints (fingerprint, proxy_ip) VALUES ($1, $2);", fp, ip)

	if err != nil {
		return false, err
	}

	return result.RowsAffected() > 0, err
}

func (db *Database) CountFingerprints() (uint64, error) {
	var out uint64
	err := db.Conn.QueryRow(context.Background(), "SELECT COUNT(*) FROM fingerprints;").Scan(&out)

	if err != nil {
		return 0, err
	}

	return out, nil
}

func (db *Database) GetRandomFingerprint() (GetFingerprintResult, error) {
	var out GetFingerprintResult

	err := db.Conn.QueryRow(context.Background(), "SELECT id, fingerprint, proxy_ip FROM fingerprints ORDER BY random() LIMIT 1").
		Scan(&out.ID, &out.Fingerprint, &out.ProxyIP)

	return out, err
}

func (db *Database) GetSpecificFingerprint(id uint64) (GetFingerprintResult, error) {
	var out GetFingerprintResult

	err := db.Conn.QueryRow(context.Background(), "SELECT id, fingerprint, proxy_ip FROM fingerprints WHERE id = $1 LIMIT 1;", id).
		Scan(&out.ID, &out.Fingerprint, &out.ProxyIP)

	return out, err
}

func (db *Database) GetAllFingerprints() ([]string, error) {
	var out []string

	rows, err := db.Conn.Query(context.Background(), "SELECT fingerprint FROM fingerprints ORDER BY id DESC")

	if err != nil {
		return out, err
	}

	for rows.Next() {
		var data string

		err = rows.Scan(&data)

		if err != nil {
			return out, err
		}

		out = append(out, data)
	}

	if rows.Err() != nil {
		return out, rows.Err()
	}

	return out, err
}

func (db *Database) CheckAuthValid(token string) (GetAuthResult, error) {
	var out GetAuthResult

	if len(token) != 32 {
		return GetAuthResult{Valid: false}, nil
	}

	err := db.Conn.QueryRow(context.Background(), "SELECT user_id, permissions FROM auth WHERE token = $1 LIMIT 1", token).
		Scan(&out.UserId, &out.Permissions)

	if err != nil && errors.Is(err, pgx.ErrNoRows) {
		return GetAuthResult{Valid: false, Permissions: 0}, nil
	}

	if err != nil {
		return GetAuthResult{}, err
	}

	out.Valid = true

	if out.UserId == 0 && token == defaultToken {
		return GetAuthResult{}, ErrDefaultToken
	}

	return out, nil
}

func (db *Database) GetAllUsers() ([]GetUserResult, error) {
	var out []GetUserResult

	rows, err := db.Conn.Query(context.Background(), "SELECT user_id, permissions, token FROM auth")

	if err != nil {
		return out, err
	}

	for rows.Next() {
		var data GetUserResult

		err = rows.Scan(&data.UserId, &data.Permissions, &data.Token)

		if err != nil {
			return out, err
		}

		out = append(out, data)
	}

	if rows.Err() != nil {
		return out, rows.Err()
	}

	return out, err
}
