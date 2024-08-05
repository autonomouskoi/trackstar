package rekordboxdb

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"

	"github.com/autonomouskoi/trackstar"
)

const (
	// DB key protected appropriately
	protectedDBKey = "QC/UgsOIF8Nf+o/7jH2TFDt0nn0xXfeoFzKh/0NghJc="
)

type db struct {
	*sqlx.DB
}

func getOptionsPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting user homedir: %w", err)
	}
	return filepath.Join(homeDir, optionsRelPath), nil
}

func newDB() (db, error) {
	optionsPath, err := getOptionsPath()
	if err != nil {
		return db{}, fmt.Errorf("getting options path: %w", err)
	}
	b, err := os.ReadFile(optionsPath)
	if err != nil {
		return db{}, fmt.Errorf("reading options file: %w", err)
	}
	options := struct {
		Options [][]string `json:"options"`
	}{}
	if err := json.Unmarshal(b, &options); err != nil {
		return db{}, fmt.Errorf("parsing options from %s: %w", optionsPath, err)
	}

	dbPath := ""
	for _, option := range options.Options {
		if len(option) < 2 {
			continue
		}
		if option[0] == "db-path" {
			dbPath = option[1]
			break
		}
	}
	if dbPath == "" {
		return db{}, errors.New("couldn't determine DB path")
	}

	b, err = base64.StdEncoding.DecodeString(protectedDBKey)
	if err != nil {
		return db{}, fmt.Errorf("decoding key: %w", err)
	}
	key := hex.EncodeToString(b)
	dsn := fmt.Sprintf("file:%s?mode=ro&_key=%s", dbPath, key)
	dbx, err := sqlx.Open("sqlite3", dsn)
	return db{dbx}, err
}

func (db db) latestHistoryID(ctx context.Context) (string, error) {
	var historyID string
	err := db.GetContext(ctx, &historyID, `
SELECT ID FROM DjmdHistory
	WHERE DateCreated = (SELECT max(DateCreated) FROM DjmdHistory)
		AND attribute = '0'`)
	return historyID, err
}

/*
ID: VARCHAR(255)
Seq: INTEGER
Name: VARCHAR(255)
Attribute: INTEGER
ParentID: VARCHAR(255)
DateCreated: VARCHAR(255)
UUID: VARCHAR(255)
rb_data_status: INTEGER
rb_local_data_status: INTEGER
rb_local_deleted: TINYINT(1)
rb_local_synced: TINYINT(1)
usn: BIGINT
rb_local_usn: BIGINT
created_at: DATETIME
updated_at: DATETIME
*/
type dbHistoryEntry struct {
	ID          string `db:"ID"`
	Name        string `db:"Name"`
	DateCreated string `db:"DateCreated"`
}

/*
ID: VARCHAR(255)
HistoryID: VARCHAR(255)
ContentID: VARCHAR(255)
TrackNo: INTEGER
UUID: VARCHAR(255)
rb_data_status: INTEGER
rb_local_data_status: INTEGER
rb_local_deleted: TINYINT(1)
rb_local_synced: TINYINT(1)
usn: BIGINT
rb_local_usn: BIGINT
created_at: DATETIME
updated_at: DATETIME
*/
type dbSongHistoryEntry struct {
	ContentID string `db:"ContentID"`
	TrackNo   int64  `db:"TrackNo"`
}

func (db db) GetSongHistory(ctx context.Context, historyID string) ([]*dbSongHistoryEntry, error) {
	var entries []*dbSongHistoryEntry
	query := `
SELECT ContentID, TrackNo FROM DjmdSongHistory
	WHERE HistoryID = ?
	ORDER BY TrackNo
`
	//query = `SELECT HistoryID, ContentID, TrackNo FROM DjmdSongHistory`
	err := db.Select(&entries, query, historyID)
	return entries, err
}

/*
ID: VARCHAR(255)
FolderPath: VARCHAR(255)
FileNameL: VARCHAR(255)
FileNameS: VARCHAR(255)
Title: VARCHAR(255)
ArtistID: VARCHAR(255)
AlbumID: VARCHAR(255)
GenreID: VARCHAR(255)
BPM: INTEGER
Length: INTEGER
TrackNo: INTEGER
BitRate: INTEGER
BitDepth: INTEGER
Commnt: TEXT
FileType: INTEGER
Rating: INTEGER
ReleaseYear: INTEGER
RemixerID: VARCHAR(255)
LabelID: VARCHAR(255)
OrgArtistID: VARCHAR(255)
KeyID: VARCHAR(255)
StockDate: VARCHAR(255)
ColorID: VARCHAR(255)
DJPlayCount: INTEGER
ImagePath: VARCHAR(255)
MasterDBID: VARCHAR(255)
MasterSongID: VARCHAR(255)
AnalysisDataPath: VARCHAR(255)
SearchStr: VARCHAR(255)
FileSize: INTEGER
DiscNo: INTEGER
ComposerID: VARCHAR(255)
Subtitle: VARCHAR(255)
SampleRate: INTEGER
DisableQuantize: INTEGER
Analysed: INTEGER
ReleaseDate: VARCHAR(255)
DateCreated: VARCHAR(255)
ContentLink: INTEGER
Tag: VARCHAR(255)
ModifiedByRBM: VARCHAR(255)
HotCueAutoLoad: VARCHAR(255)
DeliveryControl: VARCHAR(255)
DeliveryComment: VARCHAR(255)
CueUpdated: VARCHAR(255)
AnalysisUpdated: VARCHAR(255)
TrackInfoUpdated: VARCHAR(255)
Lyricist: VARCHAR(255)
ISRC: VARCHAR(255)
SamplerTrackInfo: INTEGER
SamplerPlayOffset: INTEGER
SamplerGain: FLOAT
VideoAssociate: VARCHAR(255)
LyricStatus: INTEGER
ServiceID: INTEGER
OrgFolderPath: VARCHAR(255)
Reserved1: TEXT
Reserved2: TEXT
Reserved3: TEXT
Reserved4: TEXT
ExtInfo: TEXT
rb_file_id: VARCHAR(255)
DeviceID: VARCHAR(255)
rb_LocalFolderPath: VARCHAR(255)
SrcID: VARCHAR(255)
SrcTitle: VARCHAR(255)
SrcArtistName: VARCHAR(255)
SrcAlbumName: VARCHAR(255)
SrcLength: INTEGER
UUID: VARCHAR(255)
rb_data_status: INTEGER
rb_local_data_status: INTEGER
rb_local_deleted: TINYINT(1)
rb_local_synced: TINYINT(1)
usn: BIGINT
rb_local_usn: BIGINT
created_at: DATETIME
updated_at: DATETIME
*/
type dbContent struct {
	Title    string `db:"Title"`
	ArtistID string `db:"ArtistID"`
}

func (db db) GetTrack(ctx context.Context, contentID string) (*trackstar.Track, error) {
	var content dbContent
	query := `
SELECT Title, ArtistID FROM DjmdContent
	WHERE ID = ?
`
	err := db.GetContext(ctx, &content, query, contentID)
	if err != nil {
		return nil, fmt.Errorf("getting content: %w", err)
	}

	track := &trackstar.Track{
		Title: content.Title,
	}
	err = db.GetContext(ctx, &track.Artist, `SELECT Name FROM DjmdArtist WHERE ID = ?`, content.ArtistID)
	return track, err
}

func discoverTable(dbx *sqlx.DB, table string) error {
	rows, err := dbx.Query("SELECT * FROM " + table + " LIMIT 1")
	if err != nil {
		return err
	}
	names, err := rows.Columns()
	if err != nil {
		return err
	}
	types, err := rows.ColumnTypes()
	if err != nil {
		return err
	}
	for i := range names {
		fmt.Printf("%s: %s\n", names[i], types[i].DatabaseTypeName())
	}
	return nil
}
