package sender

import (
	"errors"
	"time"
)

// ErrSendBufferFull the send service's request channel is full
var (
	ErrSendBufferFull = errors.New("Send service's request queue is full")
	ErrServiceClosed  = errors.New("Send service closed")
)

// Sender provids helper function to send coins with send service
type Sender struct {
	s *SendService
}

// NewSender creates new sender
func NewSender(s *SendService) *Sender {
	return &Sender{
		s: s,
	}
}

// SendOption send option struct
type SendOption struct {
	Timeout time.Duration
}

// Response send response
type Response struct {
	Err  string
	Txid string
}

// SendAsync send coins to dest address, should return immediately or timeout
func (s *Sender) SendAsync(destAddr string, coins int64, opt *SendOption) (<-chan interface{}, error) {
	rspC := make(chan interface{}, 1)
	req := Request{
		Address: destAddr,
		Coins:   coins,
		RspC:    rspC,
	}

	if opt != nil {
		select {
		case s.s.reqChan <- req:
			return rspC, nil
		case <-time.After(opt.Timeout):
			return rspC, ErrSendBufferFull
		}
	}

	go func() { s.s.reqChan <- req }()
	return rspC, nil
}

// Send send coins to dest address, won't return until the tx is confirmed
func (s *Sender) Send(destAddr string, coins int64, opt *SendOption) (string, error) {
	c, err := s.SendAsync(destAddr, coins, opt)
	if err != nil {
		return "", err
	}

	rsp := (<-c).(Response)

	if rsp.Err != "" {
		return "", errors.New(rsp.Err)
	}

	return rsp.Txid, nil
}
