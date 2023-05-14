package sampler

import (
	"fmt"
	"math"
	"os"
	"sync"
	"time"

	"github.com/grant-carpenter/go-ads"
)

// sample records the current sample readings
type sample struct {
	sum   int
	count int
}

// Result returns the averaged reading from the ADS since the last read
func (s *sample) Result() float64 {
	if s.count == 0 || s.sum == 0 {
		return 0
	}
	return float64(s.sum / s.count)
}

// Sampler defines a sampler
type Sampler struct {
	sync.Mutex
	currentSample *sample
	ads           *ads.ADS
	stopSignal    chan bool
	pollRate      time.Duration
}

// New returns a configured Sampler
func New() *Sampler {
	s := &Sampler{
		currentSample: &sample{},
		stopSignal:    make(chan bool),
		pollRate:      time.Millisecond * 5,
	}

	// initialise the ADS
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

	return s
}

// Start starts the sampler
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

// Stop stops the sampler reading the ADS
func (s *Sampler) Stop() {
	s.stopSignal <- true
}

// sample reads the ADS value and adds it to the sample data
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

// Read returns the samples data
func (s *Sampler) Read() float64 {
	s.Lock()
	defer s.Unlock()

	result := s.currentSample.Result()
	s.currentSample = &sample{}
	return result
}
