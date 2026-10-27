package integrity

import (
	"encoding/binary"
	"errors"
	"unicode/utf8"
)

const cborMaxDepth = 4

const (
	cborUnsigned = 0
	cborBytes    = 2
	cborText     = 3
	cborArray    = 4
	cborMap      = 5
)

var errMalformedCBOR = errors.New("integrity: malformed cbor")

type cborDecoder struct {
	data   []byte
	offset int
}

func decodeCBOR(data []byte) (any, error) {
	decoder := cborDecoder{data: data}
	value, err := decoder.value(0)
	if err != nil {
		return nil, err
	}
	if decoder.offset != len(data) {
		return nil, errMalformedCBOR
	}
	return value, nil
}

func (d *cborDecoder) value(depth int) (any, error) {
	if depth > cborMaxDepth {
		return nil, errMalformedCBOR
	}
	major, argument, err := d.head()
	if err != nil {
		return nil, err
	}
	switch major {
	case cborUnsigned:
		return argument, nil
	case cborBytes:
		return d.take(argument)
	case cborText:
		return d.text(argument)
	case cborArray:
		return d.array(argument, depth)
	case cborMap:
		return d.textMap(argument, depth)
	}
	return nil, errMalformedCBOR
}

func (d *cborDecoder) head() (byte, uint64, error) {
	if d.offset >= len(d.data) {
		return 0, 0, errMalformedCBOR
	}
	initial := d.data[d.offset]
	d.offset++
	major, info := initial>>5, initial&0x1f
	if info < 24 {
		return major, uint64(info), nil
	}
	if info > 27 {
		return 0, 0, errMalformedCBOR
	}
	raw, err := d.take(1 << (info - 24))
	if err != nil {
		return 0, 0, err
	}
	var buffer [8]byte
	copy(buffer[8-len(raw):], raw)
	return major, binary.BigEndian.Uint64(buffer[:]), nil
}

func (d *cborDecoder) remaining() uint64 {
	return uint64(len(d.data) - d.offset)
}

func (d *cborDecoder) take(length uint64) ([]byte, error) {
	if length > d.remaining() {
		return nil, errMalformedCBOR
	}
	start := d.offset
	d.offset += int(length)
	return d.data[start:d.offset], nil
}

func (d *cborDecoder) text(length uint64) (string, error) {
	raw, err := d.take(length)
	if err != nil || !utf8.Valid(raw) {
		return "", errMalformedCBOR
	}
	return string(raw), nil
}

func (d *cborDecoder) array(length uint64, depth int) ([]any, error) {
	if length > d.remaining() {
		return nil, errMalformedCBOR
	}
	items := make([]any, 0, length)
	for range length {
		item, err := d.value(depth + 1)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (d *cborDecoder) textMap(length uint64, depth int) (map[string]any, error) {
	if length > d.remaining()/2 {
		return nil, errMalformedCBOR
	}
	entries := make(map[string]any, length)
	for range length {
		key, err := d.value(depth + 1)
		if err != nil {
			return nil, err
		}
		name, isText := key.(string)
		if !isText {
			return nil, errMalformedCBOR
		}
		if _, duplicate := entries[name]; duplicate {
			return nil, errMalformedCBOR
		}
		entry, err := d.value(depth + 1)
		if err != nil {
			return nil, err
		}
		entries[name] = entry
	}
	return entries, nil
}
