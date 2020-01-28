/*
Copyright 2020 Cortex Labs, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package clusterconfig

type NATType int

const (
	UnknownNATType NATType = iota
	NoNAT
	OneNAT
	HighlyAvailableNAT
)

// These must match the expected values in eksctl
var _natTypes = []string{
	"unknown",
	"Disable",
	"Single",
	"HighlyAvailable",
}

func NATTypeFromString(s string) NATType {
	for i := 0; i < len(_natTypes); i++ {
		if s == _natTypes[i] {
			return NATType(i)
		}
	}
	return UnknownNATType
}

func NATTypeStrings() []string {
	return _natTypes[1:]
}

func (t NATType) String() string {
	return _natTypes[t]
}

// MarshalText satisfies TextMarshaler
func (t NATType) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

// UnmarshalText satisfies TextUnmarshaler
func (t *NATType) UnmarshalText(text []byte) error {
	enum := string(text)
	for i := 0; i < len(_natTypes); i++ {
		if enum == _natTypes[i] {
			*t = NATType(i)
			return nil
		}
	}

	*t = UnknownNATType
	return nil
}

// UnmarshalBinary satisfies BinaryUnmarshaler
// Needed for msgpack
func (t *NATType) UnmarshalBinary(data []byte) error {
	return t.UnmarshalText(data)
}

// MarshalBinary satisfies BinaryMarshaler
func (t NATType) MarshalBinary() ([]byte, error) {
	return []byte(t.String()), nil
}
