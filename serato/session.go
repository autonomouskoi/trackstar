package serato

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"time"

	"golang.org/x/text/encoding/unicode"
)

var (
	ErrShortRead = errors.New("short read")

	utf16Decoder = unicode.UTF16(unicode.BigEndian, unicode.UseBOM).NewDecoder()

	tagADAT    = [4]byte{'a', 'd', 'a', 't'}
	tagOENT    = [4]byte{'o', 'e', 'n', 't'}
	tagVersion = [4]byte{'v', 'r', 's', 'n'}
)

type sessionReader struct {
	decoders map[[4]byte]func(rec *Record) error
}

func (sr sessionReader) getTracks(r io.Reader) error {
	err := sr.ReadRecords(r, func(rec *Record) error {
		return nil
	})
	if err != nil {
		return fmt.Errorf("reading records: %w", err)
	}
	return nil
}

func ReadSession(r io.Reader, trackCB func(Track)) error {
	sr := sessionReader{
		decoders: map[[4]byte]func(rec *Record) error{},
	}
	sr.adatHandler(trackCB)
	sr.decoders[tagOENT] = sr.decodeRecurse
	return sr.getTracks(r)
}

type Tag [4]byte

func ReadTag(r io.Reader) (Tag, error) {
	t := Tag{}
	n, err := r.Read(t[:])
	if err != nil {
		return t, err
	}
	if n != 4 {
		return t, ErrShortRead
	}
	return t, nil
}

func (t Tag) String() string {
	return string(t[:])
}

type Record struct {
	Tag     Tag
	Length  int
	Data    []byte
	Decoded any
}

func (rec Record) MustVersion() string {
	if rec.Tag != tagVersion {
		panic("wrong tag type")
	}
	return rec.Decoded.(string)
}

func decodeVRSN(rec *Record) error {
	str, err := utf16Decoder.String(string(rec.Data))
	if err != nil {
		return err
	}
	rec.Decoded = str
	return nil
}

func (sr sessionReader) ReadRecord(r io.Reader) (Record, error) {
	rec := Record{}
	t, err := ReadTag(r)
	if err != nil {
		return rec, fmt.Errorf("reading tag: %w", err)
	}
	var length uint32
	if err := binary.Read(r, binary.BigEndian, &length); err != nil {
		return rec, fmt.Errorf("reading length: %w", err)
	}
	l := int(length)
	b := make([]byte, l)
	if _, err := io.ReadFull(r, b); err != nil {
		return rec, fmt.Errorf("reading data: %w", err)
	}
	rec.Tag = t
	rec.Length = l
	rec.Data = b
	if decode, present := sr.decoders[t]; present {
		if err := decode(&rec); err != nil {
			return rec, fmt.Errorf("decoding value: %w", err)
		}
	}
	return rec, nil
}

func (rec Record) String() string {
	return fmt.Sprintf("%s %d", rec.Tag.String(), rec.Length)
}

func (sr sessionReader) decodeRecurse(rec *Record) error {
	r := bytes.NewReader(rec.Data)
	return sr.ReadRecords(r, nil)
}

type Track struct {
	Artist string
	Title  string
	When   time.Time
}

func (sr sessionReader) adatHandler(trackCB func(Track)) {
	sr.decoders[tagADAT] = func(rec *Record) error {
		r := bytes.NewReader(rec.Data)
		var t Track
		cb := func(rec *Record) error {
			switch rec.Tag {
			case [4]byte{0, 0, 0, 6}:
				str, err := utf16Decoder.String(string(rec.Data[:len(rec.Data)-2]))
				if err != nil {
					return fmt.Errorf("decoding title: %w", err)
				}
				t.Title = str
			case [4]byte{0, 0, 0, 7}:
				str, err := utf16Decoder.String(string(rec.Data[:len(rec.Data)-2]))
				if err != nil {
					return fmt.Errorf("decoding artist: %w", err)
				}
				t.Artist = str
			case [4]byte{0, 0, 0, 0x1c}:
				var start uint32
				err := binary.Read(bytes.NewReader(rec.Data), binary.BigEndian, &start)
				if err != nil {
					return fmt.Errorf("reading start time: %w", err)
				}
				t.When = time.Unix(int64(start), 0)
			}
			return nil
		}
		if err := sr.ReadRecords(r, cb); err != nil {
			return fmt.Errorf("reading records: %w", err)
		}
		trackCB(t)
		return nil
	}
}

func (sr sessionReader) ReadRecords(r io.Reader, cb func(rec *Record) error) error {
	for {
		rec, err := sr.ReadRecord(r)
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		if cb != nil {
			if err := cb(&rec); err != nil {
				return err
			}
		}
	}
}

func PrintRecord(rec *Record) error {
	switch v := rec.Decoded.(type) {
	case string:
		fmt.Println(v)
	case fmt.Stringer:
		fmt.Println(v)
	default:
		fmt.Println(rec)
	}
	return nil
}
