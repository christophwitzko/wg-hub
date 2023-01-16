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
			_, err := tunDev.Write(data, 0)
			if err != nil {
				t.Errorf("error writing to tun: %v", err)
			}
		}
		close(doneChan)
	}()
	for _, data := range testData {
		buf := make([]byte, 1000)
		size, err := tunDev.Read(buf, 0)
		if err != nil {
			t.Errorf("error reading from tun: %v", err)
		}
		require.Equal(t, data, buf[:size])
	}
	<-doneChan
	require.NoError(t, tunDev.Close())
	buf := make([]byte, 1000)
	_, err := tunDev.Read(buf, 0)
	require.ErrorIs(t, os.ErrClosed, err)
	_, err = tunDev.Write(buf, 0)
	require.ErrorIs(t, os.ErrClosed, err)
}
