package ed2ksrv

import (
	"regexp"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/monkeyWie/goed2k/protocol"
)

func TestNewCatalogStoreRejectsMissingDatabaseDSN(t *testing.T) {
	_, err := newCatalogStore(Config{StorageBackend: storageBackendMySQL})
	if err == nil {
		t.Fatalf("expected missing dsn error")
	}
}

func TestSQLCatalogStoreLoadAndSave(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer db.Close()

	store := &sqlCatalogStore{db: db, backend: storageBackendMySQL, tableName: "shared_files"}
	file := FileRecord{
		Hash:            protocol.MustHashFromString("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"),
		Name:            "runtime-demo.iso",
		Size:            4096,
		FileType:        "Iso",
		Extension:       "iso",
		MediaCodec:      "",
		MediaLength:     0,
		MediaBitrate:    0,
		Sources:         1,
		CompleteSources: 1,
		Endpoints:       []SourceEntry{{Host: "127.0.0.1", Port: 4662}},
	}

	rows := sqlmock.NewRows([]string{"hash", "name", "size", "file_type", "extension", "media_codec", "media_length", "media_bitrate", "sources", "complete_sources", "endpoints_json"}).
		AddRow(file.Hash.String(), file.Name, file.Size, file.FileType, file.Extension, file.MediaCodec, file.MediaLength, file.MediaBitrate, file.Sources, file.CompleteSources, `[{"host":"127.0.0.1","port":4662}]`)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT hash, name, size, file_type, extension, media_codec, media_length, media_bitrate, sources, complete_sources, endpoints_json FROM shared_files ORDER BY name ASC")).WillReturnRows(rows)

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(loaded) != 1 || loaded[0].Hash.String() != file.Hash.String() || loaded[0].Endpoints[0].Port != 4662 {
		t.Fatalf("unexpected loaded files: %+v", loaded)
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM shared_files")).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO shared_files (hash, name, size, file_type, extension, media_codec, media_length, media_bitrate, sources, complete_sources, endpoints_json) VALUES (?,?,?,?,?,?,?,?,?,?,?)")).
		WithArgs(file.Hash.String(), file.Name, file.Size, file.FileType, file.Extension, file.MediaCodec, file.MediaLength, file.MediaBitrate, file.Sources, file.CompleteSources, `[{"host":"127.0.0.1","port":4662}]`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	if err := store.Save([]FileRecord{file}); err != nil {
		t.Fatalf("save: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestSQLCatalogStoreInsertSQLForPgSQL(t *testing.T) {
	store := &sqlCatalogStore{backend: storageBackendPgSQL, tableName: "shared_files"}
	got := store.insertSQL()
	want := "INSERT INTO shared_files (hash, name, size, file_type, extension, media_codec, media_length, media_bitrate, sources, complete_sources, endpoints_json) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)"
	if got != want {
		t.Fatalf("unexpected pgsql insert sql: %s", got)
	}
}
