package controller

import (
	"fmt"
	"math"
	"os"
	"sync"
	"time"

	"github.com/grant-carpenter/go-ads"
)

type sample struct {
	sum   int
	count int
}

func (s *sample) Result() float64 {
	if s.count == 0 || s.sum == 0 {
		return 0
	}
	return float64(s.sum / s.count)
}

type Sampler struct {
	sync.Mutex
	currentSample *sample
	ads           *ads.ADS
	stopSignal    chan bool
	pollRate      time.Duration
}

func NewSampler() *Sampler {
	s := &Sampler{
		currentSample: &sample{},
		stopSignal:    make(chan bool),
		pollRate:      time.Millisecond * 2,
	}

	err := ads.HostInit()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	s.ads, err = ads.NewADS("I2C1", 0x48, "")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	s.ads.SetConfigGain(ads.ConfigGain2_3)

	s.Start()

	return s
}

func (s *Sampler) Start() {
	ticker := time.NewTicker(s.pollRate)
	go func() {
		for {
			select {
			case <-ticker.C:
				s.sample()
			case <-s.stopSignal:
				ticker.Stop()
				return
			}
		}
	}()
}

func (s *Sampler) Stop() {
	s.stopSignal <- true
}

func (s *Sampler) sample() {
	s.Lock()
	defer s.Unlock()
	// read retry from ads chip
	keyResult, err := s.ads.ReadRetry(5)
	if err != nil {
		s.ads.Close()
		fmt.Println(err)
		os.Exit(1)
	}

	s.currentSample.sum += int(math.Round(float64(keyResult) / 32767.0 * 1000.0))
	s.currentSample.count++
}

func (s *Sampler) Read() float64 {
	s.Lock()
	defer s.Unlock()
	result := s.currentSample.Result()
	s.currentSample = &sample{}
	return result
}
