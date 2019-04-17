package fstest

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"regexp"
	"syscall"
	"time"
)

const firestoreHostEnvName = "FIRESTORE_EMULATOR_HOST"

var firestoreAddr = regexp.MustCompile(firestoreHostEnvName + `=(.*)`)

type Instance interface {
	Close() error
}

type Options struct {
	ProjectId string
	Port      int
	Rules     string
}

type instance struct {
	projectId             string
	firestoreEmulatorHost string
	child                 *exec.Cmd
	startupTimeout        time.Duration
	port                  int
	rules                 string
}

func findGcloud() (string, error) {
	return exec.LookPath("gcloud")
}

func NewInstance(options *Options) (Instance, error) {
	var (
		projectId = ""
		port      = 0
		rules     = ""
	)
	rand.Seed(time.Now().UnixNano())
	if options == nil {
		projectId = "test-app"
		port = 20000 + rand.Intn(9999)
	} else {
		if options.ProjectId == "" {
			projectId = "test-app"
		}
		if options.Port == 0 {
			port = 20000 + rand.Intn(9999)
		}
		rules = options.Rules
	}
	i := &instance{
		projectId:      projectId,
		startupTimeout: 15 * time.Second,
		port:           port,
		rules:          rules,
	}
	if err := i.startChild(); err != nil {
		return nil, err
	}
	return i, nil
}

func (i *instance) Close() error {
	if i.child == nil {
		return nil
	}
	defer func() {
		i.child = nil
	}()
	if p := i.child.Process; p != nil {
		errc := make(chan error, 1)
		go func() {
			errc <- i.child.Wait()
		}()
		err := p.Signal(syscall.SIGTERM)
		if err != nil {
			return err
		}
		select {
		case <-time.After(15 * time.Second):
			p.Kill()
			return errors.New("timeout killing child process")
		case err = <-errc:
		}
	}
	return nil
}

func (i *instance) startChild() error {
	gcloud, err := findGcloud()
	if err != nil {
		return err
	}
	args := []string{
		"beta",
		"emulators",
		"firestore",
		"start",
		"--host-port=" + fmt.Sprintf("localhost:%d", i.port),
		"--project=" + i.projectId,
	}
	if i.rules != "" {
		args = append(args, "--rules="+i.rules)
	}
	i.child = exec.Command(gcloud, args...)
	i.child.Stdout = os.Stdout
	var stderr io.Reader
	stderr, err = i.child.StderrPipe()
	if err != nil {
		return err
	}
	stderr = io.TeeReader(stderr, os.Stderr)
	if err = i.child.Start(); err != nil {
		return err
	}
	errc := make(chan error, 1)
	go func() {
		s := bufio.NewScanner(stderr)
		for s.Scan() {
			if match := firestoreAddr.FindStringSubmatch(s.Text()); match != nil {
				i.firestoreEmulatorHost = match[1]
				os.Setenv(firestoreHostEnvName, i.firestoreEmulatorHost)
				break
			}
		}
		errc <- s.Err()
	}()
	select {
	case <-time.After(i.startupTimeout):
		if p := i.child.Process; p != nil {
			p.Kill()
		}
		return errors.New("timeout start child process")
	case err := <-errc:
		if err != nil {
			return fmt.Errorf("error reading child process stderr: %v", err)
		}
	}
	return nil
}
