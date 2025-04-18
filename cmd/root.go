package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
)

var path string

var rootCmd = &cobra.Command{
	Use:   "watchdog-go",
	Short: "watchdog-go watches a directory and logs file changes.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := watch(path); err != nil {
			log.Fatalf("Error watching path: %v", err)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&path, "path", "p", ".", "Path to watch")
}

func watch(path string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer func(watcher *fsnotify.Watcher) {
		_ = watcher.Close()
	}(watcher)

	var wg sync.WaitGroup
	wg.Add(1)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		defer wg.Done()
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				fmt.Printf("[EVENT] %s: %s\n", event.Op, event.Name)

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)

			case <-sigs:
				fmt.Println("\nExiting...")
				return
			}
		}
	}()

	err = watcher.Add(path)
	if err != nil {
		return err
	}
	fmt.Printf("Watching directory: %s\n", filepath.Clean(path))

	wg.Wait()

	return nil
}
