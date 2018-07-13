package worker

import (
	"time"

	"github.com/micro/protobuf/ptypes"
	"github.com/pkg/errors"
	"github.com/xmc-dev/isowrap"
)

const wallGraceTime = time.Second / 4

func (w *Worker) initSandbox() error {
	w.box = isowrap.NewBox()
	w.box.Config.CPUTime, _ = ptypes.Duration(w.dataset.TimeLimit)
	w.box.Config.WallTime = w.box.Config.CPUTime + wallGraceTime
	w.box.Config.MemoryLimit = uint(w.dataset.MemoryLimit)
	w.box.Config.ShareNetwork = false
	w.box.ID = 420

	if err := w.box.Init(); err != nil {
		w.box.Cleanup()
		err = w.box.Init()
		if err != nil {
			return errors.Wrap(err, "couldn't init sandbox")
		}
	}

	log.WithField("sandbox_path", w.box.Path).Debug("Initialized sandbox")
	return nil
}

func (w *Worker) deinitSandbox() error {
	if w.box == nil {
		return nil
	}
	err := w.box.Cleanup()
	if err != nil {
		return errors.Wrap(err, "couldn't deinit sandbox")
	}

	log.Debug("Deinitialized sandbox")
	return nil
}
