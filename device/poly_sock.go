package device

import (
	"encoding/binary"
	"errors"
	"github.com/encodeous/polyamide/conn"
)

type outboundElement struct {
	buffer *[MaxMessageSize]byte // slice holding the packet data
	packet []byte                // slice of "buffer" (always!)
	ep     conn.Endpoint
}

type PolySock struct {
	recv     PolyReceiver
	outQueue chan *outboundElement
	device   *Device
}

type PolyReceiver interface {
	// Receive takes in PolyReceiver packets from the Polyamide listener. It must not block, and the packet bytes are not owned by the Receive function.
	Receive(packet []byte, endpoint conn.Endpoint)
}

func (s *PolySock) Send(packet []byte, endpoint conn.Endpoint) {
	elem := &outboundElement{}
	elem.buffer = s.device.GetMessageBuffer()
	copy(elem.buffer[4:], packet)
	binary.LittleEndian.PutUint32(packet[:4], PolySockType)
	elem.packet = elem.buffer[:4+len(packet)]
	elem.ep = endpoint
	s.outQueue <- elem
}

func newPolySock(dev *Device) *PolySock {
	return &PolySock{
		recv:     nil,
		outQueue: make(chan *outboundElement),
		device:   dev,
	}
}

func (s *PolySock) stop() {
	s.outQueue <- nil
}

func (s *PolySock) routinePolySender(maxBatchSize int) {
	defer func() {
		defer s.device.log.Verbosef("Routine: PolySock sender - stopped")
	}()
	s.device.log.Verbosef("Routine: PolySock sender - started")

	bufs := make([][]byte, 0, maxBatchSize)

	// could probably group endpoints together, but whatever.
	for elemsContainer := range s.outQueue {
		bufs = bufs[:0]
		if elemsContainer == nil {
			return
		}
		err := s.device.net.bind.Send([][]byte{elemsContainer.packet}, elemsContainer.ep)

		if err != nil {
			var errGSO conn.ErrUDPGSODisabled
			if errors.As(err, &errGSO) {
				s.device.log.Verbosef(err.Error())
				err = errGSO.RetryErr
			}
		}
		if err != nil {
			s.device.log.Errorf("Failed to send PolySock packets: %v", err)
			continue
		}
		s.device.PutMessageBuffer(elemsContainer.buffer)
	}
}
