// Copyright 2015 Dmitry Vyukov. All rights reserved.
// Use of this source code is governed by Apache 2 LICENSE that can be found in the LICENSE file.

package png

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"image/png"
	"io"
	"reflect"
)

func fixChecksums(data []byte) ([]byte, error) {
	if len(data) < 8 {
		return nil, errors.New("data too short")
	}
	var buf bytes.Buffer
	buf.Write(data[:8])
	r := bytes.NewReader(data[8:])
	for {
		chunk, err := readChunk(r)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		binary.Write(&buf, binary.BigEndian, uint32(len(chunk)-4))
		buf.Write(chunk)
		binary.Write(&buf, binary.BigEndian, crc32.ChecksumIEEE(chunk))
		r.Seek(4, io.SeekCurrent)
	}
	return buf.Bytes(), nil
}

func readChunk(r io.Reader) ([]byte, error) {
	var length uint32
	err := binary.Read(r, binary.BigEndian, &length)
	if err != nil {
		return nil, err
	}
	length += 4 // Chunk type code not included in length
	if length > 1<<20 {
		return nil, errors.New("chunk too big")
	}
	buf := make([]byte, int(length))
	_, err = io.ReadFull(r, buf)
	if err == io.EOF {
		err = io.ErrUnexpectedEOF
	}
	return buf, err
}

func Fuzz(data []byte) int {
	data, err := fixChecksums(data)
	if err != nil {
		return 0
	}
	cfg, err := png.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return 0
	}
	if cfg.Width*cfg.Height > 1e6 {
		return 0
	}
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return 0
	}
	for _, c := range []png.CompressionLevel{png.DefaultCompression, png.NoCompression, png.BestSpeed, png.BestCompression} {
		var w bytes.Buffer
		e := &png.Encoder{CompressionLevel: c}
		err = e.Encode(&w, img)
		if err != nil {
			panic(err)
		}
		img1, err := png.Decode(&w)
		if err != nil {
			panic(err)
		}
		if !reflect.DeepEqual(img.Bounds(), img1.Bounds()) {
			fmt.Printf("bounds0: %#v\n", img.Bounds())
			fmt.Printf("bounds1: %#v\n", img1.Bounds())
			panic("bounds have changed")
		}
	}
	return 1
}
