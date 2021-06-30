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
	"runtime/debug"
	"testing"
	"time"
)

func TestCounter(t *testing.T) {
	const ttl = time.Second

	counter := NewCounter(ttl, func(f float64) { t.Logf("callback: %f", f) })

	lenAssert := func(l int) {
		if x := counter.Len(); x != l {
			debug.PrintStack()
			t.Fatalf("counter.Len(); expected %d, got %d", l, x)
		}
	}

	mkReq := func(remote, path, agent, httpXforwardedFor string) map[string]string {
		return map[string]string{
			"remote":               remote,
			"path":                 path,
			"agent":                agent,
			"http_x_forwarded_for": httpXforwardedFor,
		}
	}

	lenAssert(0)

	counter.Update(mkReq("::1", "/live/foo_720p/index.m3u8", "libmpv", ""))
	lenAssert(1)

	time.Sleep(ttl * 2)
	lenAssert(0)

	counter.Update(mkReq("::1", "/live/foo_720p/index.m3u8", "libmpv", ""))
	lenAssert(1)

	counter.Update(mkReq("::2", "/live/foo_720p/index.m3u8", "libmpv", ""))
	lenAssert(2)

	time.Sleep(ttl)
	counter.Update(mkReq("::2", "/live/foo_720p/index.m3u8", "libmpv", ""))
	time.Sleep(ttl)
	lenAssert(1)

	time.Sleep(ttl * 2)
	lenAssert(0)
}
