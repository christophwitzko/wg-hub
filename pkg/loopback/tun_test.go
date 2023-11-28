package loopback

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoopbackTun(t *testing.T) {
	tunDev := CreateTun(1500)
	testData := make([][]byte, 100)
	for i := range testData {
		testData[i] = make([]byte, 500)
		testData[i][0] = byte(i)
		for x := 1; x < len(testData[i]); x++ {
			testData[i][x] = 'A'
		}
	}
	doneChan := make(chan struct{})
	go func() {
		for _, data := range testData {
			_, err := tunDev.Write([][]byte{data}, 0)
			if err != nil {
				t.Errorf("error writing to tun: %v", err)
			}
		}
		close(doneChan)
	}()
	for _, data := range testData {
		buf := [][]byte{make([]byte, 1000)}
		sizes := []int{0}
		_, err := tunDev.Read(buf, sizes, 0)
		if err != nil {
			t.Errorf("error reading from tun: %v", err)
		}
		require.Equal(t, data, buf[0][:sizes[0]])
	}
	<-doneChan
	require.NoError(t, tunDev.Close())
	buf := [][]byte{make([]byte, 1000)}
	sizes := []int{0}
	_, err := tunDev.Read(buf, sizes, 0)
	require.ErrorIs(t, os.ErrClosed, err)
	_, err = tunDev.Write(buf, 0)
	require.ErrorIs(t, os.ErrClosed, err)
}
