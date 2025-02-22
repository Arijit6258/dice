// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"fmt"
	"math"
	"testing"

	"github.com/dicedb/dice/internal/server/utils"
	"github.com/dicedb/dicedb-go/wire"
	"gotest.tools/v3/assert"
)

func TestDECR(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []struct {
		name     string
		commands []struct {
			op          string
			key         string
			val         int64
			expectedErr string
		}
	}{
		{
			name: "Decrement multiple keys",
			commands: []struct {
				op          string
				key         string
				val         int64
				expectedErr string
			}{
				{"s", "key1", 3, utils.EmptyStr},
				{"d", "key1", 2, utils.EmptyStr},
				{"d", "key1", 1, utils.EmptyStr},
				{"d", "key2", -1, utils.EmptyStr},
				{"g", "key1", 1, utils.EmptyStr},
				{"g", "key2", -1, utils.EmptyStr},
				{"s", "key3", math.MinInt64 + 1, utils.EmptyStr},
				{"d", "key3", math.MinInt64, utils.EmptyStr},
				{"d", "key3", math.MinInt64, "ERR increment or decrement would overflow"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for _, cmd := range tc.commands {
				switch cmd.op {
				case "s":
					client.FireString(fmt.Sprintf("SET %s %d", cmd.key, cmd.val))
				case "d":
					result := client.FireString(fmt.Sprintf("DECR %s", cmd.key))
					switch v := result.Value.(type) {
					case *wire.Response_VStr:
						assert.Equal(t, cmd.expectedErr, v.VStr)
					case *wire.Response_VInt:
						assert.Equal(t, cmd.val, v.VInt)
					}
				case "g":
					result := client.FireString(fmt.Sprintf("GET %s", cmd.key))
					assert.Equal(t, cmd.val, result)
				}
			}
		})
	}
}

func TestDECRBY(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	type SetCommand struct {
		key string
		val int64
	}

	type DecrByCommand struct {
		key         string
		decrValue   any
		expectedVal int64
		expectedErr string
	}

	type GetCommand struct {
		key         string
		expectedVal int64
	}

	testCases := []struct {
		name           string
		setCommands    []SetCommand
		decrByCommands []DecrByCommand
		getCommands    []GetCommand
	}{
		{
			name: "Decrement multiple keys",
			setCommands: []SetCommand{
				{"key1", 3},
				{"key3", math.MinInt64 + 1},
			},
			decrByCommands: []DecrByCommand{
				{"key1", int64(2), 1, utils.EmptyStr},
				{"key1", int64(1), 0, utils.EmptyStr},
				{"key4", int64(1), -1, utils.EmptyStr},
				{"key3", int64(1), math.MinInt64, utils.EmptyStr},
				{"key3", int64(math.MinInt64), 0, "ERR increment or decrement would overflow"},
				{"key5", "abc", 0, "ERR value is not an integer or out of range"},
			},
			getCommands: []GetCommand{
				{"key1", 0},
				{"key4", -1},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for _, cmd := range tc.setCommands {
				client.FireString(fmt.Sprintf("SET %s %d", cmd.key, cmd.val))
			}

			for _, cmd := range tc.decrByCommands {
				var result any
				switch v := cmd.decrValue.(type) {
				case int64:
					result = client.FireString(fmt.Sprintf("DECRBY %s %d", cmd.key, v))
				case string:
					result = client.FireString(fmt.Sprintf("DECRBY %s %s", cmd.key, v))
				}
				switch v := result.(type) {
				case string:
					assert.Equal(t, cmd.expectedErr, v)
				case int64:
					assert.Equal(t, cmd.expectedVal, v)
				}
			}

			for _, cmd := range tc.getCommands {
				result := client.FireString(fmt.Sprintf("GET %s", cmd.key))
				assert.Equal(t, cmd.expectedVal, result)
			}
		})
	}
}
