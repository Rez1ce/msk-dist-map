package main

import (
	"database/sql"
	"time"
)

type Promo struct {
	ID                int64
	StationName       string
	Lat               *float64
	Lon               *float64
	CreatedAt         time.Time
	AgeMinutes        int64
	SourceMessageID   *string
}

func getActivePromos(db *sql.DB) ([]Promo, error) {
	rows, err := db.Query(`
		SELECT
			p.id,
			s.NameOfStation,
			s.Latitude_WGS84,
			s.Longitude_WGS84,
			p.created_at,
			p.source_message_id
		FROM promos p
		JOIN sub_view s ON s.ID = p.station_id
		WHERE p.expires_at IS NULL
		ORDER BY p.created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []Promo
	now := time.Now().UTC()

	for rows.Next() {
		var p Promo
		var createdAtStr string
		var sourceMsgID sql.NullString
		if err := rows.Scan(&p.ID, &p.StationName, &p.Lat, &p.Lon, &createdAtStr, &sourceMsgID); err != nil {
			return nil, err
		}
		t, err := time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			return nil, err
		}
		p.CreatedAt = t
		p.AgeMinutes = int64(now.Sub(t).Minutes())
		if sourceMsgID.Valid {
			s := sourceMsgID.String
			p.SourceMessageID = &s
		}
		res = append(res, p)
	}
	return res, nil
}

func insertPromo(db *sql.DB, stationID int64, sourceMessageID *string) (int64, time.Time, error) {
    now := time.Now().UTC()
    var res sql.Result
    var err error

    if sourceMessageID != nil {
        res, err = db.Exec(`
            INSERT INTO promos (station_id, created_at, source_message_id)
            VALUES (?, ?, ?)
        `, stationID, now.Format("2006-01-02 15:04:05"), *sourceMessageID)
    } else {
        res, err = db.Exec(`
            INSERT INTO promos (station_id, created_at)
            VALUES (?, ?)
        `, stationID, now.Format("2006-01-02 15:04:05"))
    }

    if err != nil {
        return 0, time.Time{}, err
    }

    id, err := res.LastInsertId()
    if err != nil {
        return 0, time.Time{}, err
    }

    return id, now, nil
}

func updatePromo(db *sql.DB, id int64) error {
    now := time.Now().UTC()
    _, err := db.Exec(`
        UPDATE promos SET expires_at = ? WHERE id = ?
    `, now.Format("2006-01-02 15:04:05"), id)
    return err
}

func deletePromo(db *sql.DB, id int64) error {
    _, err := db.Exec(`DELETE FROM promos WHERE id = ?`, id)
    return err
}

func getStationIDByName(db *sql.DB, stationName string) (int64, error) {
    var id int64
    err := db.QueryRow(`SELECT ID FROM sub_view WHERE NameOfStation = ?`, stationName).Scan(&id)
    return id, err
}
