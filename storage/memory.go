/***************************************************************************
 * COPYRIGHT (C) 2018, Rapid7 LLC, Boston, MA, USA.
 * This code is licensed under MIT license (see LICENSE for details)
 **************************************************************************/
package storage

type memory struct {
	fileName string
	data     []byte
}

func NewMemory(filename string) *memory {
	return &memory{
		fileName: filename,
	}
}

func (m *memory) FileName() string {
	return m.fileName
}

func (m *memory) Write(data []byte) error {
	m.data = append([]byte(nil), data...)

	return nil
}

func (m *memory) Read() ([]byte, error) {
	return m.data, nil
}

func (m *memory) Exists() bool {
	return m.data != nil
}
