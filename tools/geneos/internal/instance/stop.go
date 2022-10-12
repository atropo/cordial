/*
Copyright © 2022 ITRS Group

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package instance

import (
	"errors"
	"fmt"
	"os"
	"syscall"
	"time"

	"github.com/itrs-group/cordial/tools/geneos/internal/geneos"
)

func Stop(c geneos.Instance, force bool) (err error) {
	if !force {
		err = Signal(c, syscall.SIGTERM)
		if err == os.ErrProcessDone {
			return nil
		}

		if errors.Is(err, syscall.EPERM) {
			return nil
		}

		for i := 0; i < 10; i++ {
			time.Sleep(250 * time.Millisecond)
			err = Signal(c, syscall.SIGTERM)
			if err == os.ErrProcessDone {
				break
			}
		}

		if _, err = GetPID(c); err == os.ErrProcessDone {
			fmt.Printf("%s stopped\n", c)
			return nil
		}
	}

	if err = Signal(c, syscall.SIGKILL); err == os.ErrProcessDone {
		return nil
	}

	time.Sleep(250 * time.Millisecond)
	_, err = GetPID(c)
	if err == os.ErrProcessDone {
		fmt.Printf("%s killed\n", c)
		return nil
	}
	return
}
