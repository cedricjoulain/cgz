// Modified gzip reader without header & footer based on
// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package gzip implements reading and writing of gzip format compressed files,
// as specified in RFC 1952.
package main

import (
	"bufio"
	"compress/flate"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"io"
)

const (
	gzipID1     = 0x1f
	gzipID2     = 0x8b
	gzipDeflate = 8
	flagText    = 1 << 0
	flagHdrCrc  = 1 << 1
	flagExtra   = 1 << 2
	flagName    = 1 << 3
	flagComment = 1 << 4
)

var (
	// ErrChecksum is returned when reading GZIP data that has an invalid checksum.
	ErrChecksum = errors.New("gzip: invalid checksum")
	// ErrHeader is returned when reading GZIP data that has an invalid header.
	ErrHeader = errors.New("gzip: invalid header")
)

var le = binary.LittleEndian

// noEOF converts io.EOF to io.ErrUnexpectedEOF.
func noEOF(err error) error {
	if err == io.EOF {
		return io.ErrUnexpectedEOF
	}
	return err
}

// A Reader is an io.Reader that can be read to retrieve
// uncompressed data from a gzip-format compressed file.
//
// In general, a gzip file can be a concatenation of gzip files,
// each with its own header. Reads from the Reader
// return the concatenation of the uncompressed data of each.
// Only the first header is recorded in the Reader fields.
//
// Gzip files store a length and checksum of the uncompressed data.
// The Reader will return an ErrChecksum when Read
// reaches the end of the uncompressed data if it does not
// have the expected length or checksum. Clients should treat data
// returned by Read as tentative until they receive the io.EOF
// marking the end of the data.
type Reader struct {
	r            flate.Reader
	decompressor io.ReadCloser
	digest       uint32 // CRC-32, IEEE polynomial (section 8)
	size         uint32 // Uncompressed size (section 2.3.1)
	buf          [512]byte
	err          error
}

// NewReader creates a new Reader reading the given reader.
// If r does not also implement io.ByteReader,
// the decompressor may read more data than necessary from r.
//
// It is the caller's responsibility to call Close on the Reader when done.
//
// The Reader.Header fields will be valid in the Reader returned.
func NewReader(r io.Reader) (*Reader, error) {
	z := new(Reader)
	if err := z.Reset(r); err != nil {
		return nil, err
	}
	return z, nil
}

// Reset discards the Reader z's state and makes it equivalent to the
// result of its original state from NewReader, but reading from r instead.
// This permits reusing a Reader rather than allocating a new one.
func (z *Reader) Reset(r io.Reader) error {
	*z = Reader{
		decompressor: z.decompressor,
	}
	if rr, ok := r.(flate.Reader); ok {
		z.r = rr
	} else {
		z.r = bufio.NewReader(r)
	}
	z.decompressor = flate.NewReader(z.r)
	return z.err
}

// readString reads a NUL-terminated string from z.r.
// It treats the bytes read as being encoded as ISO 8859-1 (Latin-1) and
// will output a string encoded using UTF-8.
// This method always updates z.digest with the data read.
func (z *Reader) readString() (string, error) {
	var err error
	needConv := false
	for i := 0; ; i++ {
		if i >= len(z.buf) {
			return "", ErrHeader
		}
		z.buf[i], err = z.r.ReadByte()
		if err != nil {
			return "", err
		}
		if z.buf[i] > 0x7f {
			needConv = true
		}
		if z.buf[i] == 0 {
			// Digest covers the NUL terminator.
			z.digest = crc32.Update(z.digest, crc32.IEEETable, z.buf[:i+1])

			// Strings are ISO 8859-1, Latin-1 (RFC 1952, section 2.3.1).
			if needConv {
				s := make([]rune, 0, i)
				for _, v := range z.buf[:i] {
					s = append(s, rune(v))
				}
				return string(s), nil
			}
			return string(z.buf[:i]), nil
		}
	}
}

// Read implements io.Reader, reading uncompressed bytes from its underlying Reader.
func (z *Reader) Read(p []byte) (n int, err error) {
	if z.err != nil {
		return 0, z.err
	}

	n, z.err = z.decompressor.Read(p)
	err = z.err
	return
}

// Close closes the Reader. It does not close the underlying io.Reader.
// In order for the GZIP checksum to be verified, the reader must be
// fully consumed until the io.EOF.
func (z *Reader) Close() error { return z.decompressor.Close() }
