package lowhttp

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"net/textproto"
	"strings"
)

//func FixMultipartPacket(i []byte) ([]byte, bool /* fixed */) {
//	haveMultipart := false
//	header, body := SplitHTTPHeadersAndBodyFromPacket(i, func(line string) {
//		if haveMultipart {
//			return
//		}
//		if strings.Contains(strings.ToLower(line), "content-type: multipart/form-data") {
//			haveMultipart = true
//		}
//	})
//	if haveMultipart {
//		var boundary string
//		boundary, body = FixMultipartBody(body)
//		if boundary == "" {
//			return i, false
//		}
//		headerBytes := ReplaceMIMEType([]byte(header), "multipart/form-data; boundary="+boundary)
//		return ReplaceHTTPPacketBody(headerBytes, body, false), true
//	}
//	return i, false
//}
//
//func FixMultipartBodyLegacy(i []byte) (string, []byte) {
//	// Remove leading and trailing spaces
//	i = bytes.TrimSpace(i)
//
//	// The beginning and end must be --, otherwise it is wrong, not Multipart Body
//	if !(bytes.HasPrefix(i, []byte{'-', '-'}) && bytes.HasSuffix(i, []byte{'-', '-'})) {
//		return "", nil
//	}
//
//	i = i[2 : len(i)-2]
//	scanner := bufio.NewScanner(bytes.NewBuffer(i))
//	scanner.Split(bufio.ScanBytes)
//
//	var boundary string
//	var splitBoundary string
//	var blockContent bytes.Buffer
//
//	var fullBody bytes.Buffer
//	var current byte
//	var lastByte byte
//	_ = lastByte
//	for scanner.Scan() {
//		current = scanner.Bytes()[0]
//
//		if boundary == "" {
//			// If boudnary is empty, start looking for the first boundary as the true boundary
//			switch current {
//			case '\n':
//				boundary = strings.TrimSpace(blockContent.String())
//				if boundary == "" {
//					return "", nil
//				}
//				splitBoundary = "--" + boundary
//				blockContent.Reset()
//				fullBody.WriteString(splitBoundary)
//				fullBody.WriteByte('\r')
//				fullBody.WriteByte('\n')
//				lastByte = current
//				continue
//			default:
//				blockContent.WriteByte(current)
//				lastByte = current
//				continue
//			}
//		}
//
//		// will not be repaired. The current character is \n, and other characters are not \n
//		if current == '\n' {
//			if strings.HasPrefix(blockContent.String(), "--"+boundary) {
//				// Parse the boundary line
//				fullBody.WriteString(splitBoundary)
//				fullBody.WriteByte('\r')
//				fullBody.WriteByte('\n')
//				blockContent.Reset()
//				lastByte = current
//				continue
//			} else {
//				// Parse textproto
//				r := textproto.NewReader(bufio.NewReader(bytes.NewBufferString(blockContent.String())))
//				var headers []string
//				for {
//					line, _ := r.ReadLine()
//					line = strings.TrimSpace(line)
//					if line == "" {
//						// Empty lines should exit the loop line
//						break
//					}
//					headers = append(headers, line)
//				}
//				rawBody, _ := ioutil.ReadAll(r.R)
//				if len(rawBody) > 0 {
//					headers = append(headers, string(rawBody))
//				}
//
//				fullBody.WriteString(strings.Join(headers, CRLF))
//				fullBody.WriteString(CRLF)
//				blockContent.Reset()
//				lastByte = current
//				continue
//			}
//		}
//
//		blockContent.WriteByte(current)
//		lastByte = current
//	}
//	fullBody.WriteString("--")
//	fullBody.WriteString(boundary)
//	fullBody.WriteString("--")
//	fullBody.WriteString(CRLF)
//	return boundary, fullBody.Bytes()
//}

// State machine
func FixMultipartBody(i []byte) (string, []byte) {
	var lineBuffer bytes.Buffer
	var blockBuffer bytes.Buffer
	var boundary string
	var splitBoundary string

	const (
		FINDING_BOUNDARY = 1
		PARSING_BLOCK    = 2
		BLOCK_FINISHED   = 3
		FINISHED         = 4
	)
	var state = FINDING_BOUNDARY

	i = bytes.TrimSpace(i)
	var blocks []bytes.Buffer

	raw := bytes.NewBuffer(i)
	raw.WriteByte('\n')
	scanner := bufio.NewScanner(raw)
	scanner.Split(bufio.ScanBytes)

	handleBlockBuffer := func() bytes.Buffer {
		if blockBuffer.Len() <= 0 {
			return blockBuffer
		}
		// Parse textproto
		var rawBody []byte
		r := textproto.NewReader(bufio.NewReader(bytes.NewBufferString(blockBuffer.String())))
		var blockFixedBuffer bytes.Buffer

		fastFail := false
		for {
			line, err := r.ReadLine()
			if err != nil {
				// Reading error
				fastFail = true
				break
			}
			line = strings.TrimSpace(line)
			if line == "" {
				blockFixedBuffer.WriteString(CRLF)
				rawBody, _ = ioutil.ReadAll(r.R)
				if len(rawBody) <= 0 {
					fastFail = true
				}
				break
			}
			blockFixedBuffer.WriteString(line)
			blockFixedBuffer.WriteString(CRLF)
		}

		if !fastFail {
			if bytes.HasSuffix(rawBody, []byte{'\r', '\n'}) {
				// block ends with CRLF,
				blockFixedBuffer.Write(rawBody)
			} else if bytes.HasSuffix(rawBody, []byte{'\n'}) {
				blockFixedBuffer.Write(rawBody[:len(rawBody)-1])
				blockFixedBuffer.WriteString(CRLF)
			} else {
				blockFixedBuffer.Write(rawBody)
				blockFixedBuffer.WriteString(CRLF)
			}
		}

		return blockFixedBuffer
	}

	for scanner.Scan() {
		current := scanner.Bytes()[0]
		lineBuffer.WriteByte(current)
		if current != '\n' {
			continue
		}

		// State flow
	NEXT_STATE:
		switch state {
		case FINDING_BOUNDARY:
			firstLine := strings.TrimSpace(lineBuffer.String())
			if firstLine == "" {
				return "", nil
			}
			if !strings.HasPrefix(firstLine, "--") {
				return "", nil
			}
			splitBoundary = firstLine
			boundary = splitBoundary[2:]
			lineBuffer.Reset()
			state = PARSING_BLOCK
			continue
		case PARSING_BLOCK:
			if lineBuffer.Len() < 3*len(splitBoundary) {
				trimed := strings.TrimSpace(lineBuffer.String())
				if strings.HasPrefix(trimed, splitBoundary) {
					if strings.HasPrefix(trimed, splitBoundary+"--") {
						state = FINISHED
					} else {
						state = BLOCK_FINISHED
					}
					lineBuffer.Reset()
					goto NEXT_STATE
				}
			}
			blockBuffer.Write(lineBuffer.Bytes())
			lineBuffer.Reset()
			continue
		case BLOCK_FINISHED, FINISHED:
			blockFixedBuffer := handleBlockBuffer()
			blocks = append(blocks, blockFixedBuffer)

			lineBuffer.Reset()
			blockBuffer.Reset()
			if state == FINISHED {
				break
			}
			state = PARSING_BLOCK
			continue
		}
	}

	if blockBuffer.Len() > 0 {
		blocks = append(blocks, handleBlockBuffer())
	}

	var body bytes.Buffer
	body.WriteString(splitBoundary)
	for _, b := range blocks {
		body.WriteString(CRLF)
		body.Write(b.Bytes())
		body.WriteString(splitBoundary)
	}
	body.WriteString("--")
	body.WriteString(CRLF)
	return boundary, body.Bytes()
}
