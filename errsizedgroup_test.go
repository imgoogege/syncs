package syncs

import (
	"errors"
	"fmt"
	"log"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestErrorSizedGroup(t *testing.T) {
	ewg := NewErrSizedGroup(10)
	var c uint32

	for i := 0; i < 1000; i++ {
		i := i
		ewg.Go(func() error {
			time.Sleep(time.Millisecond)
			atomic.AddUint32(&c, 1)
			if i == 100 {
				return errors.New("err1")
			}
			if i == 200 {
				return errors.New("err2")
			}
			return nil
		})
	}
	assert.True(t, runtime.NumGoroutine() > 900, "goroutines %d", runtime.NumGoroutine())

	err := ewg.Wait()
	assert.True(t, strings.HasPrefix(err.Error(), "2 error(s) occurred:"))
	assert.Equal(t, uint32(1000), c, fmt.Sprintf("%d, not all routines have been executed.", c))
}

func TestErrorSizedGroupPreGo(t *testing.T) {
	ewg := NewErrSizedGroup(10, Preemptive())
	var c uint32

	for i := 0; i < 1000; i++ {
		i := i
		ewg.Go(func() error {
			assert.True(t, runtime.NumGoroutine() < 20, "goroutines %d", runtime.NumGoroutine())
			atomic.AddUint32(&c, 1)
			if i == 100 {
				return errors.New("err1")
			}
			if i == 200 {
				return errors.New("err2")
			}
			time.Sleep(time.Millisecond)
			return nil
		})
	}

	err := ewg.Wait()
	assert.True(t, strings.HasPrefix(err.Error(), "2 error(s) occurred:"))
	assert.Equal(t, uint32(1000), c, fmt.Sprintf("%d, not all routines have been executed.", c))
}

func TestErrorSizedGroupNoError(t *testing.T) {
	ewg := NewErrSizedGroup(10)
	var c uint32

	for i := 0; i < 1000; i++ {
		ewg.Go(func() error {
			atomic.AddUint32(&c, 1)
			return nil
		})
	}

	err := ewg.Wait()
	assert.Nil(t, err)
	assert.Equal(t, uint32(1000), c, fmt.Sprintf("%d, not all routines have been executed.", c))
}

func TestErrorSizedGroupTerm(t *testing.T) {
	ewg := NewErrSizedGroup(10, TermOnErr())
	var c uint32

	for i := 0; i < 1000; i++ {
		i := i
		ewg.Go(func() error {
			atomic.AddUint32(&c, 1)
			if i == 100 {
				return errors.New("err")
			}
			return nil
		})
	}

	err := ewg.Wait()
	assert.Equal(t, "1 error(s) occurred: [0] {err}", err.Error())
	assert.True(t, c < uint32(1000), fmt.Sprintf("%d, some of routines has to be terminated early", c))
}

// illustrates the use of a SizedGroup for concurrent, limited execution of goroutines.
func ExampleErrorSizedGroup_go() {

	// create sized waiting group allowing maximum 10 goroutines
	grp := NewErrSizedGroup(10)

	var c uint32
	for i := 0; i < 1000; i++ {
		// Go call is non-blocking, like regular go statement
		grp.Go(func() error {
			// do some work in 10 goroutines in parallel
			atomic.AddUint32(&c, 1)
			time.Sleep(10 * time.Millisecond)
			return nil
		})
	}
	// Note: grp.Go acts like go command - never blocks. This code will be executed right away
	log.Print("all 1000 jobs submitted")

	// wait for completion
	if err := grp.Wait(); err != nil {
		panic(err)
	}
}