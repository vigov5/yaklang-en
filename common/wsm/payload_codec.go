package wsm

type PayloadCodecI interface {
	// EchoResultEncodeFormYak payload internally encodes the echo result, hybrid programming, and executes yaklang
	EchoResultEncodeFormYak(raw []byte) ([]byte, error)
	// EchoResultDecodeFormYak decodes the echo result of the payload
	EchoResultDecodeFormYak(raw []byte) ([]byte, error)
	SetPayloadScriptContent(content string)
}
