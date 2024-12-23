// This file is part of DiceDB.
// Copyright (C) 2024 DiceDB (dicedb.io).
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package websocket

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestExists(t *testing.T) {
	exec := NewWebsocketCommandExecutor()

	testCases := []struct {
		name     string
		commands []string
		expected []interface{}
		delay    []time.Duration
	}{
		{
			name:     "Test EXISTS command",
			commands: []string{"SET key value", "EXISTS key", "EXISTS key2"},
			expected: []interface{}{"OK", float64(1), float64(0)},
			delay:    []time.Duration{0, 0, 0},
		},
		{
			name:     "Test EXISTS command with multiple keys",
			commands: []string{"SET key value", "SET key2 value2", "EXISTS key key2 key3", "EXISTS key key2 key3 key4", "DEL key", "EXISTS key key2 key3 key4"},
			expected: []interface{}{"OK", "OK", float64(2), float64(2), float64(1), float64(1)},
			delay:    []time.Duration{0, 0, 0, 0, 0, 0},
		},
		{
			name:     "Test EXISTS an expired key",
			commands: []string{"SET key value ex 2", "EXISTS key", "EXISTS key"},
			expected: []interface{}{"OK", float64(1), float64(0)},
			delay:    []time.Duration{0, 0, 2 * time.Second},
		},
		{
			name:     "Test EXISTS with multiple keys and expired key",
			commands: []string{"SET key value ex 2", "SET key2 value2", "SET key3 value3", "EXISTS key key2 key3", "EXISTS key key2 key3"},
			expected: []interface{}{"OK", "OK", "OK", float64(3), float64(2)},
			delay:    []time.Duration{0, 0, 0, 0, 2 * time.Second},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			conn := exec.ConnectToServer()
			exec.FireCommandAndReadResponse(conn, "FLUSHDB")
			for i, cmd := range tc.commands {
				if tc.delay[i] > 0 {
					time.Sleep(tc.delay[i])
				}
				result, err := exec.FireCommandAndReadResponse(conn, cmd)
				assert.Nil(t, err)
				assert.Equal(t, tc.expected[i], result)
			}
		})
	}
}
