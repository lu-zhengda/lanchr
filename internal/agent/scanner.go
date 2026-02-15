package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/zhengda-lu/lanchr/internal/launchctl"
	"github.com/zhengda-lu/lanchr/internal/platform"
	"github.com/zhengda-lu/lanchr/internal/plist"
)

// Scanner discovers all launch agents and daemons across all domains
// and correlates plist data with live launchctl state.
type Scanner struct {
	parser    *plist.Parser
	launchctl launchctl.Executor
}

// NewScanner creates a new service scanner.
func NewScanner(parser *plist.Parser, executor launchctl.Executor) *Scanner {
	return &Scanner{
		parser:    parser,
		launchctl: executor,
	}
}

// plistResult holds the result of parsing a single plist file.
type plistResult struct {
	pl   *plist.LaunchAgentPlist
	path string
	dir  platform.PlistDir
	err  error
}

// ScanAll returns all services found across all plist directories,
// enriched with live runtime state from launchctl.
func (s *Scanner) ScanAll() ([]Service, error) {
	dirs := platform.PlistDirectories()

	// Step 1: Discover and parse all plist files in parallel.
	plistResults := s.scanPlistDirs(dirs)

	// Step 2: Get live state from launchctl list.
	listEntries, err := s.launchctl.List()
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}

	// Build a lookup map from label to list entry.
	entryMap := make(map[string]launchctl.ListEntry, len(listEntries))
	for _, entry := range listEntries {
		entryMap[entry.Label] = entry
	}

	// Step 3: Get disabled states.
	userDisabled, _ := s.launchctl.PrintDisabled(platform.GUIDomainTarget())
	systemDisabled, _ := s.launchctl.PrintDisabled("system")

	// Merge disabled maps.
	disabledMap := make(map[string]bool)
	for label, disabled := range userDisabled {
		if disabled {
			disabledMap[label] = true
		}
	}
	for label, disabled := range systemDisabled {
		if disabled {
			disabledMap[label] = true
		}
	}

	// Step 4: Build services from plist data, correlating with launchctl state.
	seenLabels := make(map[string]bool)
	var services []Service

	for _, result := range plistResults {
		if result.err != nil {
			continue
		}

		pl := result.pl
		label := pl.Label
		if label == "" {
			// Use the filename (without extension) as a fallback label.
			label = strings.TrimSuffix(filepath.Base(result.path), ".plist")
		}

		svc := Service{
			Label:             label,
			Domain:            result.dir.Domain,
			Type:              result.dir.Type,
			Status:            StatusStopped,
			PID:               -1,
			PlistPath:         result.path,
			Program:           pl.Program,
			ProgramArgs:       pl.ProgramArguments,
			RunAtLoad:         pl.RunAtLoad,
			KeepAlive:         pl.KeepAlive,
			StartInterval:     pl.StartInterval,
			WatchPaths:        pl.WatchPaths,
			QueueDirectories:  pl.QueueDirectories,
			StandardOutPath:   pl.StandardOutPath,
			StandardErrorPath: pl.StandardErrorPath,
			WorkingDirectory:  pl.WorkingDirectory,
			EnvironmentVars:   pl.EnvironmentVariables,
			UserName:          pl.UserName,
			GroupName:         pl.GroupName,
			Disabled:          pl.Disabled,
			ExitTimeout:       pl.ExitTimeOut,
			ThrottleInterval:  pl.ThrottleInterval,
			Nice:              pl.Nice,
			ProcessType:       pl.ProcessType,
			MachServices:      pl.MachServices,
			Sockets:           pl.Sockets,
		}

		// Correlate with launchctl list entry.
		if entry, ok := entryMap[label]; ok {
			svc.PID = entry.PID
			svc.LastExitStatus = entry.Status

			if entry.PID > 0 {
				svc.Status = StatusRunning
			} else if entry.Status != 0 {
				svc.Status = StatusError
			}
		}

		// Check disabled state.
		if disabledMap[label] {
			svc.Disabled = true
			if svc.Status == StatusStopped {
				svc.Status = StatusDisabled
			}
		}

		seenLabels[label] = true
		services = append(services, svc)
	}

	// Step 5: Add services found in launchctl list but with no plist on disk.
	for _, entry := range listEntries {
		if seenLabels[entry.Label] {
			continue
		}

		svc := Service{
			Label:          entry.Label,
			Domain:         platform.DomainUser,
			Type:           platform.TypeAgent,
			PID:            entry.PID,
			LastExitStatus: entry.Status,
			Status:         StatusStopped,
		}

		if entry.PID > 0 {
			svc.Status = StatusRunning
		} else if entry.Status != 0 {
			svc.Status = StatusError
		}

		if disabledMap[entry.Label] {
			svc.Disabled = true
			if svc.Status == StatusStopped {
				svc.Status = StatusDisabled
			}
		}

		services = append(services, svc)
	}

	return services, nil
}

// ScanDomain returns services from a specific domain only.
func (s *Scanner) ScanDomain(domain platform.Domain) ([]Service, error) {
	all, err := s.ScanAll()
	if err != nil {
		return nil, err
	}

	var filtered []Service
	for _, svc := range all {
		if svc.Domain == domain {
			filtered = append(filtered, svc)
		}
	}
	return filtered, nil
}

// FindByLabel finds a service by its label across all domains.
func (s *Scanner) FindByLabel(label string) (*Service, error) {
	all, err := s.ScanAll()
	if err != nil {
		return nil, err
	}

	for i := range all {
		if all[i].Label == label {
			return &all[i], nil
		}
	}
	return nil, fmt.Errorf("service %q not found", label)
}

// scanPlistDirs reads and parses all plist files from the given directories using
// a worker pool bounded by the number of CPUs.
func (s *Scanner) scanPlistDirs(dirs []platform.PlistDir) []plistResult {
	type parseJob struct {
		path string
		dir  platform.PlistDir
	}

	// Collect all plist file paths.
	var jobs []parseJob
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir.Path)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".plist") {
				continue
			}
			jobs = append(jobs, parseJob{
				path: filepath.Join(dir.Path, entry.Name()),
				dir:  dir,
			})
		}
	}

	// Parse in parallel with a bounded worker pool.
	numWorkers := runtime.NumCPU()
	if numWorkers > len(jobs) {
		numWorkers = len(jobs)
	}
	if numWorkers < 1 {
		numWorkers = 1
	}

	jobCh := make(chan parseJob, len(jobs))
	resultCh := make(chan plistResult, len(jobs))

	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobCh {
				pl, err := s.parser.Parse(job.path)
				resultCh <- plistResult{
					pl:   pl,
					path: job.path,
					dir:  job.dir,
					err:  err,
				}
			}
		}()
	}

	for _, job := range jobs {
		jobCh <- job
	}
	close(jobCh)

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	var results []plistResult
	for result := range resultCh {
		results = append(results, result)
	}

	return results
}
