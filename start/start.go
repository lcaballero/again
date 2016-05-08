package start

import (
	cmd "github.com/codegangsta/cli"
	"fmt"
	"github.com/lcaballero/hitman"
	"gopkg.in/fsnotify.v1"
	"github.com/vrecan/death"
	"syscall"
	"os/exec"
	"os"
	"bufio"
	"sync"
)


func Run(cmd *cmd.Context) {
	params := cmd.Args()
	exe := params.First()
	args := params.Tail()

	root, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	targets := hitman.NewTargets()
	targets.AddTarget(NewWatch(root, exe, args))

	death.NewDeath(syscall.SIGTERM, syscall.SIGINT).WaitForDeath(targets)
}

type Watch struct {
	dirs []string
	cmd string
	args []string
}

func NewWatch(dir string, cmd string, args []string) *Watch {
	dirs := walk(dir)
	return &Watch{
		dirs: dirs,
		cmd: cmd,
		args: args,
	}
}

func (w *Watch) Start() hitman.KillChannel {
	kill := hitman.NewKillChannel()
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}

	for _,dir := range w.dirs {
		fmt.Printf("Adding watcher for: %s\n", dir)
		err = watcher.Add(dir)
		if err != nil {
			panic(err)
		}
	}

	//TODO: remove size of channel in favor of putting entire notification
	//TODO: and spin up of command in a separate goroutine.
	handleEvent := make(chan fsnotify.Event, 1000)
	restarting := false

	fmt.Println("Begin watching")
	go func(done hitman.KillChannel) {
		for {
			select {
			case cleaner := <-done:
				watcher.Close()
				cleaner.WaitGroup.Done()
				return

			case ev := <-handleEvent:
				fmt.Println("handling event", ev)
				stop:
				for {
					select {
					case ev = <-handleEvent:
					default:
						break stop
					}
				}
				restarting = true
				w.RestartTriggered()
				restarting = false

			case ev := <-watcher.Events:
				if restarting {
					continue
				}
				if ev.Op&fsnotify.Rename != 0 && ev.Op&ev.Op&fsnotify.Chmod != 0 {
					fmt.Println(ev)
					handleEvent <- ev
				}

			case err := <-watcher.Errors:
				fmt.Println("error:", err)
			}
		}
	}(kill)
	return kill
}

//
func (w *Watch) RestartTriggered() {
	fmt.Println("RestartTriggered -$ ", w.cmd, w.args)
	command := exec.Command(w.cmd, w.args...)

	outReader, err := command.StdoutPipe()
	if err != nil {
		fmt.Println("Error when restarting command.", err)
		return
	}

	errReader, err := command.StderrPipe()
	if err != nil {
		fmt.Println("Error opening err pipe", err)
	}

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		errScan := bufio.NewScanner(errReader)
		for errScan.Scan() {
			fmt.Fprintln(os.Stderr, errScan.Text())
		}
		wg.Done()
	}()
	go func() {
		outScan := bufio.NewScanner(outReader)
		for outScan.Scan() {
			fmt.Fprintln(os.Stdout, outScan.Text())
		}
		wg.Done()
	}()

	err = command.Start()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	wg.Wait()
	err = command.Wait()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}