package worker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"github.com/pkg/errors"
	"github.com/shopspring/decimal"

	"github.com/micro/go-micro/client"
	"github.com/micro/protobuf/ptypes"
	"github.com/sirupsen/logrus"
	"github.com/xmc-dev/xmc/dispatcher-srv/db/models/job"
	pjob "github.com/xmc-dev/xmc/dispatcher-srv/proto/job"
	"github.com/xmc-dev/xmc/eval-srv/service"
	"github.com/xmc-dev/xmc/eval-srv/util"
	"github.com/xmc-dev/isowrap"
	"github.com/xmc-dev/xmc/xmc-core/common"
	pattachment "github.com/xmc-dev/xmc/xmc-core/proto/attachment"
	pdataset "github.com/xmc-dev/xmc/xmc-core/proto/dataset"
	pgrader "github.com/xmc-dev/xmc/xmc-core/proto/grader"
	presult "github.com/xmc-dev/xmc/xmc-core/proto/result"
	ptask "github.com/xmc-dev/xmc/xmc-core/proto/task"
)

var log = logrus.WithField("prefix", "worker")
var jobClient = pjob.NewJobsServiceClient("xmc.srv.dispatcher", client.DefaultClient)
var datasetClient = pdataset.NewDatasetServiceClient("xmc.srv.core", client.DefaultClient)
var taskClient = ptask.NewTaskServiceClient("xmc.srv.core", client.DefaultClient)
var attachmentClient = pattachment.NewAttachmentServiceClient("xmc.srv.core", client.DefaultClient)
var graderClient = pgrader.NewGraderServiceClient("xmc.srv.core", client.DefaultClient)

// Worker executes Jobs
type Worker struct {
	job           *job.Job
	m             *sync.Mutex
	srv           *service.Service
	result        *presult.Result
	tempDir       string
	dataset       *pdataset.Dataset
	task          *ptask.Task
	box           *isowrap.Box
	graderProgram *common.Program
	userProgram   *common.Program
	nrTestCases   int
}

// NewWorker creates a new Worker
func NewWorker(srv *service.Service) *Worker {
	w := new(Worker)

	w.job = nil
	w.m = &sync.Mutex{}
	w.srv = srv

	return w
}

// IsIdle returns true if the worker is idle
func (w *Worker) IsIdle() bool {
	return w.job == nil
}

// Work works the job
func (w *Worker) Work(j *job.Job) error {
	w.m.Lock()
	if !w.IsIdle() {
		w.m.Unlock()
		return errors.New("Worker is not idle")
	}
	w.job = j
	w.m.Unlock()
	go w.execute()
	return nil
}

func (w *Worker) execute() {
	if w.IsIdle() {
		return
	}
	id := w.job.UUID
	log.WithField("job_uuid", id).Info("Starting work")
	err := w.prepare()
	if err != nil {
		if len(w.result.ErrorMessage) == 0 {
			w.result.ErrorMessage = "err_prepare:" + err.Error()
		}
	} else {
		err = w.work()
		if err != nil {
			if len(w.result.ErrorMessage) == 0 {
				w.result.ErrorMessage = "err_system:" + err.Error()
			}
		}
	}
	w.finish()
	w.cleanup()
	log.WithField("job_uuid", id).Info("Work finished")
	w.next()
}

func (w *Worker) makeTemp() error {
	tempDir, err := ioutil.TempDir("", "xmc-eval-w-"+w.srv.Name+"-d"+w.job.DatasetID)
	if err != nil {
		return errors.Wrap(err, "couldn't create temp dir")
	}
	w.tempDir = tempDir
	log.WithField("tempDir", tempDir).Debug("Created tempdir")

	return nil
}

func (w *Worker) writeUserProgram() error {
	p := filepath.Join(w.tempDir, "userprogram."+w.job.Language)
	err := ioutil.WriteFile(p, w.job.Code, 0644)
	if err != nil {
		return errors.Wrap(err, "couldn't write user program")
	}
	w.userProgram = common.NewProgram(p, filepath.Join(w.tempDir, "userprogram"), common.Language(w.job.Language))

	return nil
}

func (w *Worker) getDataset() error {
	drsp, err := datasetClient.Read(context.TODO(), &pdataset.ReadRequest{Id: w.job.DatasetID})
	if err != nil {
		return errors.Wrapf(err, "couldn't get dataset %s", w.job.DatasetID)
	}
	w.dataset = drsp.Dataset
	log.WithField("dataset", w.dataset.Id).Debug("Got dataset")

	return nil
}

func (w *Worker) getTask() error {
	rsp, err := taskClient.Read(context.TODO(), &ptask.ReadRequest{Id: w.job.TaskID})
	if err != nil {
		return errors.Wrapf(err, "couldn't get task %s", w.job.TaskID)
	}
	w.task = rsp.Task
	log.WithField("task", w.task.Id).Debug("Got task")

	return nil
}

func (w *Worker) getGrader() error {
	grsp, err := graderClient.Read(context.TODO(), &pgrader.ReadRequest{Id: w.dataset.GraderId})
	if err != nil {
		return errors.Wrapf(err, "couldn't get grader %s", w.dataset.GraderId)
	}
	garsp, err := attachmentClient.GetContents(C(), &pattachment.GetContentsRequest{Id: grsp.Grader.AttachmentId})
	if err != nil {
		return errors.Wrapf(err, "couldn't get grader %s contents", err)
	}
	p := filepath.Join(w.tempDir, "grader."+string(grsp.Grader.Language))
	err = util.Download(garsp.Url, p)
	if err != nil {
		return err
	}
	w.graderProgram = common.NewProgram(p, filepath.Join(w.tempDir, "grader"), common.Language(grsp.Grader.Language))
	log.WithField("grader", w.dataset.GraderId).Debug("Got grader")

	return nil
}

func (w *Worker) getTestCases() error {
	rsp, err := datasetClient.GetTestCases(context.TODO(), &pdataset.GetTestCasesRequest{Id: w.job.DatasetID})
	if err != nil {
		return errors.Wrapf(err, "couldn't get dataset's %s test cases", w.job.DatasetID)
	}
	w.nrTestCases = len(rsp.TestCases)
	for _, tc := range rsp.TestCases {
		rsp, err := attachmentClient.GetContents(C(), &pattachment.GetContentsRequest{Id: tc.InputAttachmentId})
		if err != nil {
			return errors.Wrapf(err, "couldn't get input #%d contents", tc.Number)
		}
		inURL := rsp.Url
		rsp, err = attachmentClient.GetContents(C(), &pattachment.GetContentsRequest{Id: tc.OutputAttachmentId})
		if err != nil {
			return errors.Wrapf(err, "couldn't get output #%d contents", tc.Number)
		}
		outURL := rsp.Url
		err = util.Download(inURL, filepath.Join(w.tempDir, fmt.Sprintf("test%d.in", tc.Number)))
		if err != nil {
			return err
		}
		err = util.Download(outURL, filepath.Join(w.tempDir, fmt.Sprintf("test%d.ok", tc.Number)))
		if err != nil {
			return err
		}
	}
	log.Debug("Successfully downloaded tests")

	return nil
}

func (w *Worker) compilePrograms() error {
	upc := w.userProgram.Compile()
	w.result.BuildCommand = cmdString(upc)
	log.Debug("Compiling user program ", w.result.BuildCommand)
	out, err := upc.CombinedOutput()
	verOut, _ := w.userProgram.Version().CombinedOutput()
	w.result.CompilationMessage = string(verOut) + "\n" + string(out)
	if err != nil {
		w.result.ErrorMessage = "err_userprogram_compilation:" + err.Error()
		return errors.Wrap(err, "couldn't compile user program")
	}
	log.Debug("Successfully compiled user program")

	graderBuildCmd := cmdString(w.graderProgram.Compile())
	log.Debug("Compiling grader ", graderBuildCmd)
	out, err = w.graderProgram.Compile().CombinedOutput()
	if err != nil {
		w.result.ErrorMessage = fmt.Sprintf("err_grader_compilation:%s\n%s\n%s", graderBuildCmd, out, err.Error())
		return errors.Wrap(err, "couldn't compile grader program")
	}
	log.Debug("Successfully compiled grader program")
	return nil
}

func (w *Worker) copyUserProgram() error {
	err := util.CopyFile(w.userProgram.Executable, filepath.Join(w.box.Path, "userprogram"))
	if err != nil {
		return err
	}
	return nil
}

func (w *Worker) prepare() error {
	log = log.WithField("job_uuid", w.job.UUID)
	w.result = &presult.Result{TestResults: []*presult.TestResult{}}
	if err := w.makeTemp(); err != nil {
		return err
	}
	if err := w.writeUserProgram(); err != nil {
		return err
	}
	if err := w.getDataset(); err != nil {
		return err
	}
	if err := w.getTask(); err != nil {
		return err
	}
	if err := w.getGrader(); err != nil {
		return err
	}
	if err := w.getTestCases(); err != nil {
		return err
	}
	if err := w.compilePrograms(); err != nil {
		return err
	}

	return nil
}

func (w *Worker) work() error {
	scoreSum := decimal.Zero
	stdoutFilename := filepath.Join(w.tempDir, "userprogram.out")
	defer w.deinitSandbox()
	for i := 1; i <= w.nrTestCases; i++ {
		tr := &presult.TestResult{
			TestNo: int32(i),
			Score:  "0.00",
		}
		if err := w.initSandbox(); err != nil {
			return err
		}
		if err := w.copyUserProgram(); err != nil {
			return err
		}
		stdout, err := os.Create(stdoutFilename)
		if err != nil {
			return errors.Wrap(err, "couldn't open stdout file")
		}

		testFile := filepath.Join(w.tempDir, fmt.Sprintf("test%d", i))
		if w.task.InputFile != "stdin" {
			err := util.CopyFile(testFile+".in", filepath.Join(w.box.Path, w.task.InputFile))
			if err != nil {
				return err
			}
		}
		var stdin io.Reader = os.Stdin
		if w.task.InputFile == "stdin" {
			in, err := os.Open(testFile + ".in")
			if err != nil {
				return errors.Wrap(err, "couldn't read input file")
			}
			stdin = in
		}
		result, err := w.box.Run(stdin, stdout, os.Stderr, "userprogram")
		if err != nil {
			return errors.Wrap(err, "couldn't execute user program")
		}
		stdout.Close()
		log.Debug(result, err)
		if result.ErrorType != isowrap.NoError {
			switch result.ErrorType {
			case isowrap.RunTimeError:
				tr.GraderMessage = fmt.Sprintf("Program exited with exit status %d", result.ExitCode)
			case isowrap.KilledBySignal:
				tr.GraderMessage = fmt.Sprintf("Killed by signal %d: %v", int(result.Signal.(syscall.Signal)), result.Signal)
			case isowrap.Timeout:
				tr.GraderMessage = "Time limit exceeded"
			case isowrap.MemoryExceeded:
				tr.GraderMessage = "Memory limit exceeded"
			}
		} else {
			if w.task.OutputFile != "stdout" {
				err = util.CopyFile(filepath.Join(w.box.Path, w.task.OutputFile), stdoutFilename)
				if err != nil {
					w.result.ErrorMessage = "err_no_output_file:" + w.task.OutputFile
					return nil
				}
			}

			gProc := w.graderProgram.Execute(testFile+".in", stdoutFilename, testFile+".ok")
			gProc.Dir = w.tempDir
			var gOut, gErr bytes.Buffer
			gProc.Stdout = &gOut
			gProc.Stderr = &gErr
			err = gProc.Run()
			if err != nil {
				return errors.Wrap(err, "couldn't execute grader program")
			}

			// ignores error, because if there's an error then most likely the grader failed
			score, _ := decimal.NewFromString(strings.TrimSpace(string(gOut.Bytes())))
			scoreSum = scoreSum.Add(score)
			tr.Score = score.String()
			tr.GraderMessage = strings.TrimSpace(string(gErr.Bytes()))
		}
		tr.Memory = int32(result.MemUsed)
		tr.Time = ptypes.DurationProto(result.CPUTime)
		w.result.TestResults = append(w.result.TestResults, tr)
		if err := w.deinitSandbox(); err != nil {
			return err
		}
		if f, ok := stdin.(*os.File); ok {
			f.Close()
		}
	}
	w.result.Score = scoreSum.Div(decimal.NewFromFloat(float64(w.nrTestCases))).Mul(decimal.NewFromFloat(100.)).String()

	return nil
}

func (w *Worker) finish() {
	log.Debug(w.result)
	rsp, err := jobClient.Finish(CWithName(w.srv.Name), &pjob.FinishRequest{
		JobUuid: w.job.UUID.String(),
		Result:  w.result,
	})
	log.WithField("job_uuid", w.job.UUID).Info("Work done")
	if err != nil {
		log.Error(err)
	}
	w.m.Lock()
	if rsp == nil || rsp.NextJob == nil {
		w.job = nil
		log.Info("No work left. Idling...")
	} else {
		log.WithField("job_uuid", w.job.UUID).WithField("next_job", rsp.NextJob.Uuid).Info("Next job")
		w.job = job.FromProto(rsp.NextJob)
	}
	w.m.Unlock()
}

func (w *Worker) cleanup() {
	w.result = nil
	err := w.deinitSandbox()
	if err != nil {
		log.Error("Error while cleaning up: couldn't deinit sandbox: ", err)
	}
	err = os.RemoveAll(w.tempDir)
	if err != nil {
		log.WithField("tempDir", w.tempDir).Error("Error while cleaning up: couldn't remove worker's temp dir: ", err)
	}
	w.tempDir = ""
	w.dataset = nil
	w.graderProgram = nil
	w.userProgram = nil
	log = log.WithField("job_uuid", "")
}

func (w *Worker) next() {
	go w.execute()
}
