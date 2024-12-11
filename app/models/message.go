package models

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

type Message struct {
	Header         Header   `json:"header"`
	Question       Question `json:"question"`
	Answer         Answer   `json:"answer"`
	Authority      string   `json:"authority"`
	AdditionalData string   `json:"additional"`
}

type Header struct {
	PacketIdentifier      uint16 `json:"packet_identifier"`
	Flags                 uint16 `json:"flags"`
	QuestionCount         uint16 `json:"question_count"`
	AnswerRecordCount     uint16 `json:"answer_record_count"`
	AuthorityRecordCount  uint16 `json:"authority_record_count"`
	AdditionalRecordCount uint16 `json:"additional_record_count"`
}

// 15 14 13 12 11 10 9 8 7 6 5 4 3 2 1 0
func (h *Header) SetFlags(QueryIndicator uint16, OperationCode uint16, AuthoritativeAnswer uint16, Truncation uint16, RecursionDesired uint16, RecursionAvailable uint16, Reserved uint16, ResponseCode uint16) {
	//start appending flags
	h.Flags |= (QueryIndicator << 15)
	h.Flags |= (OperationCode << 11)
	h.Flags |= (AuthoritativeAnswer << 10)
	h.Flags |= (Truncation << 9)
	h.Flags |= (RecursionDesired << 8)
	h.Flags |= (RecursionAvailable << 7)
	h.Flags |= (Reserved << 4)
	h.Flags |= (ResponseCode << 0)
}

func (h *Header) Bytes(PacketIdentifier, Flags, QuestionCount, AnswerRecordCount, AuthorityRecordCount, AdditionalRecordCount uint16) []byte {
	//start appending flags
	buf := make([]byte, 12)
	binary.BigEndian.PutUint16(buf[0:2], PacketIdentifier)
	binary.BigEndian.PutUint16(buf[2:4], Flags)
	binary.BigEndian.PutUint16(buf[4:6], QuestionCount)
	binary.BigEndian.PutUint16(buf[6:8], AnswerRecordCount)
	binary.BigEndian.PutUint16(buf[8:10], AuthorityRecordCount)
	binary.BigEndian.PutUint16(buf[10:12], AdditionalRecordCount)
	return buf
}

func (h *Header) SetFlagsWithResponseBytes(responseBytes []byte) []byte {
	flags := make([]byte, 2)
	//flags will contain byte1 and 2 of response byte
	flags[0] = responseBytes[2]
	flags[1] = responseBytes[3]
	// 15 14 13 12 11 10 9 8 7 6 5 4 3 2 1 0
	flagsToBeReturned := uint16(0)
	flagsToBeReturned |= (uint16(1) << 15)

	//mimic next 4 bits
	opcode := false
	for _, val := range []int{14, 13, 12, 11} {
		if binary.BigEndian.Uint16(flags)&(uint16(1)<<val) != 0 {
			flagsToBeReturned |= (uint16(1) << val)
			opcode = true
		}
	}

	//mimic 7th bit
	if binary.BigEndian.Uint16(flags)&(uint16(1)<<8) != 0 {
		flagsToBeReturned |= (uint16(1) << 8)
	}

	if opcode {
		//make last 4 bits as 0100 --> 4
		flagsToBeReturned |= (uint16(1) << 2)
	}
	// 00001000 00000000
	commonFlagsBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(commonFlagsBytes, flagsToBeReturned)

	return commonFlagsBytes
}

// , answerSize int, authorityRecord int, additionalRecord int
func (h *Header) SetRemainingDataAndReturnBytes(responseBytes []byte, questionSize int) []byte {
	returnResponseBytes := make([]byte, 12)
	returnResponseBytes[0] = responseBytes[0]
	returnResponseBytes[1] = responseBytes[1]

	flagResponse := h.SetFlagsWithResponseBytes(responseBytes)
	returnResponseBytes[2] = flagResponse[0]
	returnResponseBytes[3] = flagResponse[1]

	questionBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(questionBytes, uint16(questionSize))
	returnResponseBytes[4] = questionBytes[0]
	returnResponseBytes[5] = questionBytes[1]

	answerRecordBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(answerRecordBytes, uint16(questionSize))
	returnResponseBytes[6] = answerRecordBytes[0]
	returnResponseBytes[7] = answerRecordBytes[1]

	authorityRecordBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(authorityRecordBytes, 0)
	returnResponseBytes[8] = authorityRecordBytes[0]
	returnResponseBytes[9] = authorityRecordBytes[1]

	additionalRecordBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(additionalRecordBytes, 0)
	returnResponseBytes[10] = additionalRecordBytes[0]
	returnResponseBytes[11] = additionalRecordBytes[1]

	return returnResponseBytes
}

func (m *Message) Bytes(headerBytes []byte) []byte {
	b := new(bytes.Buffer)
	b.Write(headerBytes)
	return b.Bytes()
}

type Question struct {
	Name  string `json:"name"`
	Type  uint16 `json:"type"`
	Class uint16 `json:"class"`
}

func SetName(name string) []byte {
	namee := []byte{} //the name which we gonna assign to the question
	url := strings.Split(name, ".")
	for _, val := range url {
		namee = append(namee, byte(len(val)))
		namee = append(namee, []byte(val)...)
	}
	namee = append(namee, 0x00)

	return namee
}

func (q *Question) SetAllDataAndReturnQuestionBytes(name string, typ uint16, clas uint16) []byte {
	nameByte := SetName(name)

	typeBuf := make([]byte, 2)
	binary.BigEndian.PutUint16(typeBuf, typ)

	classBuf := make([]byte, 2)
	binary.BigEndian.PutUint16(classBuf, clas)

	commonBuf := []byte{}
	commonBuf = append(commonBuf, nameByte...)
	commonBuf = append(commonBuf, typeBuf...)

	commonBuf = append(commonBuf, classBuf...)
	return commonBuf
}

func ConvertNumToHexString(num uint8) string {
	buf := new(bytes.Buffer)

	// Convert the number to a byte slice (big-endian)
	err := binary.Write(buf, binary.BigEndian, num)
	if err != nil {
		logrus.Infoln("Error:", err)
		return ""
	}

	// Get the byte slice
	byteSlice := buf.Bytes()

	str := ``
	// Print the byte slice with \x format
	for _, b := range byteSlice {
		str = fmt.Sprintf(`\x%02x`, b)
	}
	return str
}

type Answer struct {
	Name   string `json:"name"`
	Type   uint16 `json:"type"`
	Class  uint16 `json:"class"`
	TTL    uint32 `json:"ttl"`
	Length uint16 `json:"length"`
	Data   uint32 `json:"data"`
}

func (a *Answer) FillAnswerAndReturnBytes(domainName string, typ int, class int, ttl int, length int, rdata string) []byte {
	nameBytes := SetName(domainName)

	typeBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(typeBytes, uint16(typ))

	classBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(classBytes, uint16(class))

	ttlBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(ttlBytes, uint32(ttl))

	lengthBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(lengthBytes, uint16(length))

	dataBytes := make([]byte, 4)

	//number to hex , then pass
	ip := rdata
	urlSplit := strings.Split(ip, ".")
	for ind, val := range urlSplit {
		num, _ := strconv.Atoi(val)
		dataBytes[ind] = byte(uint8(num))
	}

	logrus.Info(dataBytes)

	commonBytes := []byte{}
	commonBytes = append(commonBytes, nameBytes...)
	commonBytes = append(commonBytes, typeBytes...)
	commonBytes = append(commonBytes, classBytes...)
	commonBytes = append(commonBytes, ttlBytes...)
	commonBytes = append(commonBytes, lengthBytes...)
	commonBytes = append(commonBytes, dataBytes...)

	return commonBytes

}
