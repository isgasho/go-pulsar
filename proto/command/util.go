package command

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
)

func NewSizeFrame(size int) (frame []byte, err error) {
	b := new(bytes.Buffer)
	err = binary.Write(b, binary.BigEndian, uint32(size))
	if err != nil {
		err = errors.Wrap(err, "failed to write as binary")
		return
	}

	frame = b.Bytes()
	return
}

func MarshalMessage(msg proto.Message) (
	size int, data []byte, err error,
) {
	msgFrame, err := proto.Marshal(msg)
	if err != nil {
		err = errors.Wrap(err, "failed to proto.Marshal message")
		return
	}

	msgLength := len(msgFrame)
	sizeFrame, err := NewSizeFrame(msgLength)
	if err != nil {
		err = errors.Wrap(err, "failed to create size frame")
		return
	}

	size = msgLength + int(FrameSizeFieldSize)
	data = append(sizeFrame, msgFrame...)
	return
}

func HasChecksum(frame []byte) (r bool) {
	nextBytes := frame[0:FrameMagicNumberFieldSize]
	r = bytes.Equal(nextBytes, FrameMagicNumber)
	return
}

func VerifyChecksum(data []byte) (msgAndPayload []byte, err error) {
	pos := FrameMagicNumberFieldSize + FrameChecksumSize
	checksum := data[FrameMagicNumberFieldSize:pos]
	msgAndPayload = data[pos:]

	calcChecksum := CalculateChecksum(msgAndPayload)
	if !bytes.Equal(checksum, calcChecksum) {
		s := "unmatch checksum: received: %s, calculate %s"
		err = errors.Errorf(s, checksum, calcChecksum)
		return
	}

	return
}

var crc32cTable = crc32.MakeTable(crc32.Castagnoli)

func CalculateChecksum(data []byte) (checksum []byte) {
	checksum = make([]byte, FrameChecksumSize)
	value := crc32.Checksum(data, crc32cTable)
	binary.BigEndian.PutUint32(checksum, value)
	return
}
