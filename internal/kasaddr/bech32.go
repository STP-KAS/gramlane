package kasaddr

import (
	"fmt"
	"strings"
)

const charset = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"
const checksumLength = 8

var generator = []int{0x98f2bc8e61, 0x79b76d99e2, 0xf33e5fb3c4, 0xae2eabe2a8, 0x1e4f43e470}

const (
	VersionPubKey     byte = 0x00
	VersionScriptHash byte = 0x08
)

func Encode(prefix string, payload []byte, version byte) string {
	data := make([]byte, len(payload)+1)
	data[0] = version
	copy(data[1:], payload)
	converted := convertBits(data, 8, 5, true)
	checksum := calculateChecksum(prefix, converted)
	combined := append(converted, checksum...)
	return fmt.Sprintf("%s:%s", prefix, encodeToBase32(combined))
}

func encodeToBase32(data []byte) string {
	result := make([]byte, 0, len(data))
	for _, b := range data {
		if int(b) >= len(charset) {
			return ""
		}
		result = append(result, charset[b])
	}
	return string(result)
}

func convertBits(data []byte, fromBits, toBits uint8, pad bool) []byte {
	var regrouped []byte
	nextByte := byte(0)
	filledBits := uint8(0)
	for _, b := range data {
		b = b << (8 - fromBits)
		remainingFromBits := fromBits
		for remainingFromBits > 0 {
			remainingToBits := toBits - filledBits
			toExtract := remainingFromBits
			if remainingToBits < toExtract {
				toExtract = remainingToBits
			}
			nextByte = (nextByte << toExtract) | (b >> (8 - toExtract))
			b = b << toExtract
			remainingFromBits -= toExtract
			filledBits += toExtract
			if filledBits == toBits {
				regrouped = append(regrouped, nextByte)
				filledBits = 0
				nextByte = 0
			}
		}
	}
	if pad && filledBits > 0 {
		nextByte = nextByte << (toBits - filledBits)
		regrouped = append(regrouped, nextByte)
	}
	return regrouped
}

func calculateChecksum(prefix string, payload []byte) []byte {
	prefixLower5Bits := prefixToUint5Array(prefix)
	payloadInts := ints(payload)
	templateZeroes := []int{0, 0, 0, 0, 0, 0, 0, 0}
	concat := append(prefixLower5Bits, 0)
	concat = append(concat, payloadInts...)
	concat = append(concat, templateZeroes...)
	polyModResult := polyMod(concat)
	res := make([]byte, checksumLength)
	for i := 0; i < checksumLength; i++ {
		res[i] = byte((polyModResult >> uint(5*(checksumLength-1-i))) & 31)
	}
	return res
}

func prefixToUint5Array(prefix string) []int {
	out := make([]int, len(prefix))
	for i := 0; i < len(prefix); i++ {
		out[i] = int(prefix[i] & 31)
	}
	return out
}

func ints(payload []byte) []int {
	out := make([]int, len(payload))
	for i, b := range payload {
		out[i] = int(b)
	}
	return out
}

func polyMod(values []int) int {
	checksum := 1
	for _, value := range values {
		topBits := checksum >> 35
		checksum = ((checksum & 0x07ffffffff) << 5) ^ value
		for i := 0; i < len(generator); i++ {
			if ((topBits >> uint(i)) & 1) == 1 {
				checksum ^= generator[i]
			}
		}
	}
	return checksum ^ 1
}

func DecodePayload(encoded string) (payload []byte, version byte, err error) {
	encoded = strings.ToLower(encoded)
	colon := strings.LastIndexByte(encoded, ':')
	if colon < 1 {
		return nil, 0, fmt.Errorf("no prefix")
	}
	prefix := encoded[:colon]
	data := encoded[colon+1:]
	decoded := make([]byte, 0, len(data))
	for i := 0; i < len(data); i++ {
		idx := strings.IndexByte(charset, data[i])
		if idx < 0 {
			return nil, 0, fmt.Errorf("bad char")
		}
		decoded = append(decoded, byte(idx))
	}
	if len(decoded) < checksumLength {
		return nil, 0, fmt.Errorf("short")
	}
	body := decoded[:len(decoded)-checksumLength]
	converted := convertBits(body, 5, 8, false)
	if len(converted) < 1 {
		return nil, 0, fmt.Errorf("empty")
	}
	_ = prefix
	return converted[1:], converted[0], nil
}
