// Package bandplan defines frequency band and channel data structures for RF
// spectrum analysis, along with built-in band sets for common RF standards.
package bandplan

import (
	"encoding/json"
	"errors"
	"fmt"
)

// MaxFrequency is the largest center or width Validate accepts: 2^53-1 Hz,
// the largest integer a JSON number holds exactly in IEEE 754 doubles, so
// JavaScript and other float-based decoders read it without loss.
const MaxFrequency Frequency = 1<<53 - 1

// Channel is a single RF channel defined by its center frequency and bandwidth.
// In JSON, frequencies are integer hertz in centerHz and widthHz.
type Channel struct {
	Name   string
	Center Frequency
	Width  Frequency
}

type channelJSON struct {
	Name     string `json:"name"`
	CenterHz int64  `json:"centerHz"`
	WidthHz  int64  `json:"widthHz"`
}

// MarshalJSON implements json.Marshaler.
func (c Channel) MarshalJSON() ([]byte, error) {
	return json.Marshal(channelJSON{Name: c.Name, CenterHz: int64(c.Center), WidthHz: int64(c.Width)})
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *Channel) UnmarshalJSON(b []byte) error {
	var j channelJSON
	if err := json.Unmarshal(b, &j); err != nil {
		return err
	}
	*c = Channel{Name: j.Name, Center: Frequency(j.CenterHz), Width: Frequency(j.WidthHz)}
	return nil
}

// Low returns the lower edge of the channel.
func (c Channel) Low() Frequency {
	return c.Center - c.Width/2
}

// High returns the upper edge of the channel.
func (c Channel) High() Frequency {
	return c.Low() + c.Width
}

// Contains reports whether f lies within the channel edges, inclusive.
func (c Channel) Contains(f Frequency) bool {
	return f >= c.Low() && f <= c.High()
}

// Band is a named group of channels.
type Band struct {
	Name     string    `json:"name"`
	Channels []Channel `json:"channels"`
}

// BandSet is a named collection of bands covering a theme or standard.
// Bands and channels within a set may overlap in frequency.
type BandSet struct {
	Name  string `json:"name"`
	Bands []Band `json:"bands"`
}

// Span returns the lowest and highest channel edge across all bands. It
// returns (0, 0) if the set has no channels.
func (bs *BandSet) Span() (lo, hi Frequency) {
	var first = true
	for _, band := range bs.Bands {
		for _, ch := range band.Channels {
			if first {
				lo, hi = ch.Low(), ch.High()
				first = false
				continue
			}
			lo = min(lo, ch.Low())
			hi = max(hi, ch.High())
		}
	}
	return lo, hi
}

// ChannelRef identifies a channel within a set by its band and channel
// indexes.
type ChannelRef struct {
	Band, Channel int
}

// Lookup returns the channels whose edges contain f, in set order.
func (bs *BandSet) Lookup(f Frequency) []ChannelRef {
	var refs []ChannelRef
	for bi, band := range bs.Bands {
		for ci, ch := range band.Channels {
			if ch.Contains(f) {
				refs = append(refs, ChannelRef{Band: bi, Channel: ci})
			}
		}
	}
	return refs
}

// Validate checks that the set, its bands and its channels are named and
// that every channel has a positive center and width no larger than
// MaxFrequency.
func (bs *BandSet) Validate() error {
	if bs.Name == "" {
		return errors.New("band set has no name")
	}
	for _, band := range bs.Bands {
		if band.Name == "" {
			return fmt.Errorf("band set %q: band has no name", bs.Name)
		}
		for _, ch := range band.Channels {
			if ch.Name == "" {
				return fmt.Errorf("band set %q, band %q: channel has no name", bs.Name, band.Name)
			}
			if ch.Center <= 0 {
				return fmt.Errorf("band set %q, band %q, channel %q: center must be positive", bs.Name, band.Name, ch.Name)
			}
			if ch.Width <= 0 {
				return fmt.Errorf("band set %q, band %q, channel %q: width must be positive", bs.Name, band.Name, ch.Name)
			}
			if ch.Center > MaxFrequency || ch.Width > MaxFrequency {
				return fmt.Errorf("band set %q, band %q, channel %q: frequency exceeds %d Hz", bs.Name, band.Name, ch.Name, int64(MaxFrequency))
			}
		}
	}
	return nil
}

// Clone returns a deep copy of the set.
func (bs *BandSet) Clone() *BandSet {
	c := &BandSet{Name: bs.Name, Bands: make([]Band, len(bs.Bands))}
	for i, band := range bs.Bands {
		c.Bands[i] = Band{Name: band.Name, Channels: append([]Channel(nil), band.Channels...)}
	}
	return c
}

// Find returns the set named name, or nil if there is none.
func Find(sets []*BandSet, name string) *BandSet {
	for _, bs := range sets {
		if bs.Name == name {
			return bs
		}
	}
	return nil
}
