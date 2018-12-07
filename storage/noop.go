/***************************************************************************
 * COPYRIGHT (C) 2018, Rapid7 LLC, Boston, MA, USA.
 * This code is licensed under MIT license (see LICENSE for details)
 **************************************************************************/
package storage

type noop struct {
	fileName string
}

func (n noop) FileName() string {
	return n.fileName
}

func (n noop) Write(data []byte) error {
	return nil
}

func (n noop) Read() ([]byte, error) {
	return nil, nil
}
