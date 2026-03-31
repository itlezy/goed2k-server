package ed2ksrv

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/monkeyWie/goed2k/protocol"
	serverproto "github.com/monkeyWie/goed2k/protocol/server"
)

const (
	opOfferFiles byte = 0x15
	opDisconnect byte = 0x18
)

// OfferFiles represents the ED2K server-side shared file announcement packet.
type OfferFiles struct {
	Entries []serverproto.SharedFileEntry
}

func (o *OfferFiles) Get(src *bytes.Reader) error {
	count, err := protocol.ReadUInt32(src)
	if err != nil {
		return err
	}
	o.Entries = make([]serverproto.SharedFileEntry, int(count))
	for idx := range o.Entries {
		if err := o.Entries[idx].Get(src); err != nil {
			return err
		}
	}
	return nil
}

func (o OfferFiles) Put(dst *bytes.Buffer) error {
	if err := protocol.WriteUInt32(dst, uint32(len(o.Entries))); err != nil {
		return err
	}
	for idx := range o.Entries {
		if err := o.Entries[idx].Put(dst); err != nil {
			return err
		}
	}
	return nil
}

func (o OfferFiles) BytesCount() int {
	size := 4
	for idx := range o.Entries {
		size += o.Entries[idx].BytesCount()
	}
	return size
}

func fileRecordFromSharedEntry(entry serverproto.SharedFileEntry) (FileRecord, error) {
	record := FileRecord{Hash: entry.Hash}
	if name, ok := entry.StringTag(protocol.FTFilename); ok {
		record.Name = strings.TrimSpace(name)
	}
	if record.Name == "" {
		return FileRecord{}, fmt.Errorf("shared file name is required")
	}
	if size, ok := entry.UIntTag(protocol.FTFileSize); ok {
		record.Size = int64(size)
	}
	if sizeHi, ok := entry.UIntTag(protocol.FTFileSizeHi); ok {
		record.Size += int64(sizeHi << 32)
	}
	if record.Size <= 0 {
		return FileRecord{}, fmt.Errorf("shared file size must be positive")
	}
	if sources, ok := entry.UIntTag(protocol.FTSources); ok {
		record.Sources = int(sources)
	}
	if complete, ok := entry.UIntTag(protocol.FTCompleteSources); ok {
		record.CompleteSources = int(complete)
	}
	if bitrate, ok := entry.UIntTag(protocol.FTMediaBitrate); ok {
		record.MediaBitrate = int(bitrate)
	}
	if length, ok := entry.UIntTag(protocol.FTMediaLength); ok {
		record.MediaLength = int(length)
	}
	if codec, ok := entry.StringTag(protocol.FTMediaCodec); ok {
		record.MediaCodec = codec
	}
	if ext, ok := entry.StringTag(protocol.FTFileFormat); ok {
		record.Extension = ext
	}
	if fileType, ok := entry.StringTag(protocol.FTFileType); ok {
		record.FileType = fileType
	} else if fileTypeID, ok := entry.UIntTag(protocol.FTFileType); ok {
		record.FileType = ed2kFileTypeName(fileTypeID)
	}
	if record.Extension == "" {
		record.Extension = strings.TrimPrefix(strings.ToLower(filepath.Ext(record.Name)), ".")
	}
	if record.FileType == "" {
		record.FileType = inferFileTypeFromExtension(record.Extension)
	}
	return normalizeFileRecord(record)
}

func ed2kFileTypeName(value uint64) string {
	switch value {
	case 1:
		return "Audio"
	case 2:
		return "Video"
	case 3:
		return "Image"
	case 4:
		return "Program"
	case 5:
		return "Document"
	case 6:
		return "Archive"
	case 7:
		return "Iso"
	default:
		return ""
	}
}

func inferFileTypeFromExtension(ext string) string {
	switch strings.ToLower(strings.TrimSpace(ext)) {
	case "mp3", "flac", "aac", "ogg", "wav", "m4a":
		return "Audio"
	case "mp4", "mkv", "avi", "wmv", "mov", "mpeg":
		return "Video"
	case "jpg", "jpeg", "png", "gif", "bmp", "webp":
		return "Image"
	case "zip", "rar", "7z", "tar", "gz":
		return "Archive"
	case "iso", "bin", "nrg":
		return "Iso"
	case "pdf", "doc", "docx", "txt", "epub":
		return "Document"
	case "exe", "msi", "apk", "deb", "rpm":
		return "Program"
	default:
		return ""
	}
}
