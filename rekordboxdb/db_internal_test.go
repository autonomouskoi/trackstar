package rekordboxdb

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

const (
	createTableDjmdContent = `
CREATE TABLE DjmdContent (
	ID       VARCHAR(255),
	Title    VARCHAR(255),
	ArtistID VARCHAR(255)
)`
)

// https://github.com/autonomouskoi/trackstar/issues/4
// "getting track: getting content: sql: Scan error on column index 1, name \"ArtistID\": converting NULL to string is unsupported"
func TestGetTrackNullArtistID(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// connect
	conn, err := sqlx.Open("sqlite3", ":memory:")
	require.NoError(t, err, "opening database")

	// initialize
	_, err = conn.Exec(createTableDjmdContent)
	require.NoError(t, err, "creating table")
	defer conn.Close()

	contentID := "test123"
	title := "test-title"
	_, err = conn.Exec(`INSERT INTO DjmdContent (ID, Title) VALUES (?,?)`, contentID, title)
	require.NoError(t, err, "inserting")

	db := db{conn}
	track, err := db.GetTrack(ctx, contentID)
	require.NoError(t, err, "getting track")
	_ = track
}
