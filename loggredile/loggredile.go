package loggredile

import (
	"fmt"
	"net/http"
	"time"

	"code.google.com/p/gogoprotobuf/proto"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	"github.com/gorilla/websocket"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

func StreamIntoGBuffer(port int, path string, sourceName string, outBuf, errBuf *gbytes.Buffer) chan<- bool {
	outMessages := make(chan *logmessage.LogMessage)
	errMessages := make(chan *logmessage.LogMessage)

	stop := streamMessages(port, path, outMessages, errMessages)

	go func() {
		outOpen := true
		errOpen := true

		for outOpen || errOpen {
			var message *logmessage.LogMessage

			select {
			case message, outOpen = <-outMessages:
				if !outOpen {
					outMessages = nil
					break
				}

				if message.GetSourceName() != sourceName {
					continue
				}

				outBuf.Write([]byte(string(message.GetMessage()) + "\n"))
			case message, errOpen = <-errMessages:
				if !errOpen {
					errMessages = nil
					break
				}

				if message.GetSourceName() != sourceName {
					continue
				}

				errBuf.Write([]byte(string(message.GetMessage()) + "\n"))
			}
		}
	}()

	return stop
}

func streamMessages(port int, path string, outMessages, errMessages chan<- *logmessage.LogMessage) chan<- bool {
	stop := make(chan bool, 1)

	var ws *websocket.Conn
	i := 0
	for {
		var err error
		ws, _, err = websocket.DefaultDialer.Dial(
			fmt.Sprintf("ws://127.0.0.1:%d%s", port, path),
			http.Header{},
		)
		if err != nil {
			i++
			if i > 10 {
				fmt.Printf("Unable to connect to Server in 100ms, giving up.\n")
				return nil
			}

			time.Sleep(10 * time.Millisecond)
			continue
		} else {
			break
		}
	}

	go func() {
		for {
			_, data, err := ws.ReadMessage()

			if err != nil {
				close(outMessages)
				close(errMessages)
				return
			}

			receivedMessage := &logmessage.LogMessage{}
			err = proto.Unmarshal(data, receivedMessage)
			Ω(err).ShouldNot(HaveOccurred())

			if receivedMessage.GetMessageType() == logmessage.LogMessage_OUT {
				outMessages <- receivedMessage
			} else {
				errMessages <- receivedMessage
			}
		}
	}()

	go func() {
		for {
			err := ws.WriteMessage(websocket.BinaryMessage, []byte{42})
			if err != nil {
				break
			}

			select {
			case <-stop:
				ws.Close()
				return

			case <-time.After(1 * time.Second):
				// keep-alive
			}
		}
	}()

	return stop
}
