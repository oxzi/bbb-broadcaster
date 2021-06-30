/*
 * Copyright (c) 2021 Alvar Penning
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package main

import (
	"crypto/sha256"
	"sync"
	"time"
)

// Counter for users accessing live web ressources, like /live/ID_SEL/index.m3u8
//
// Users will be grouped based on their remote and agent together with the
// requested path and an optional http_x_forwarded_for header. If such a request
// have not occurred since a given TTL, it will be removed.
type Counter struct {
	entries  map[[sha256.Size]byte]time.Time
	mutex    sync.Mutex
	ttl      time.Duration
	callback func(float64)
}

// NewCounter for a TTL after which users will be removed. The callback function
// is called at after every change.
func NewCounter(ttl time.Duration, callback func(float64)) (counter *Counter) {
	counter = &Counter{
		entries:  make(map[[sha256.Size]byte]time.Time),
		ttl:      ttl,
		callback: callback,
	}

	go func() {
		for now := range time.Tick(time.Second) {
			counter.mutex.Lock()
			for k, v := range counter.entries {
				if delta := now.Sub(v); delta > counter.ttl {
					delete(counter.entries, k)
				}
			}
			counter.callback(float64(len(counter.entries)))
			counter.mutex.Unlock()
		}
	}()

	return
}

// Update the Counter with a nginx log entry, as a map.
//
// The current time stamp is used instead of the log's date. Thus, it is only
// intended for "real time" usage.
func (counter *Counter) Update(request map[string]string) {
	h := sha256.New()
	for _, f := range []string{"remote", "path", "agent", "http_x_forwarded_for"} {
		h.Write([]byte(request[f]))
	}
	var k [sha256.Size]byte
	copy(k[:], h.Sum(nil))

	counter.mutex.Lock()
	defer counter.mutex.Unlock()

	counter.entries[k] = time.Now()
	counter.callback(float64(len(counter.entries)))
}

// Len returns the amount of currently active users.
func (counter *Counter) Len() int {
	counter.mutex.Lock()
	defer counter.mutex.Unlock()
	return len(counter.entries)
}
