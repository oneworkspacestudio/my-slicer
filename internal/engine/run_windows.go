//go:build windows

package engine

import (
	"context"
	"errors"
	"golang.org/x/sys/windows"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"unsafe"
)

// Run starts suspended, assigns a bounded kill-on-close Windows job, then resumes.
// Cancellation terminates the process tree, not only the launcher. No shell is used.
func Run(ctx context.Context, exe, dir string, args []string, out io.Writer) error {
	if e := ctx.Err(); e != nil {
		return e
	}
	job, e := windows.CreateJobObject(nil, nil)
	if e != nil {
		return e
	}
	defer windows.CloseHandle(job)
	info := windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION{}
	info.BasicLimitInformation.LimitFlags = windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE | windows.JOB_OBJECT_LIMIT_JOB_MEMORY | windows.JOB_OBJECT_LIMIT_ACTIVE_PROCESS
	info.JobMemoryLimit = 4 << 30
	info.BasicLimitInformation.ActiveProcessLimit = 4
	if _, e = windows.SetInformationJobObject(job, windows.JobObjectExtendedLimitInformation, uintptr(unsafe.Pointer(&info)), uint32(unsafe.Sizeof(info))); e != nil {
		return e
	}
	cmd := exec.Command(exe, args...)
	cmd.Dir = dir
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: windows.CREATE_SUSPENDED | windows.CREATE_NO_WINDOW}
	cmd.Stdout = out
	cmd.Stderr = out
	cmd.Env = []string{"SystemRoot=" + os.Getenv("SystemRoot"), "WINDIR=" + os.Getenv("WINDIR"), "TEMP=" + dir, "TMP=" + dir, "USERPROFILE=" + dir, "APPDATA=" + filepath.Join(dir, "roaming"), "LOCALAPPDATA=" + filepath.Join(dir, "local"), "PATH=" + filepath.Dir(exe) + ";" + filepath.Join(os.Getenv("SystemRoot"), "System32"), "OMP_NUM_THREADS=2"}
	if e = cmd.Start(); e != nil {
		return e
	}
	fail := func(e error) error { _ = cmd.Process.Kill(); _ = cmd.Wait(); return e }
	proc, e := windows.OpenProcess(windows.PROCESS_SET_QUOTA|windows.PROCESS_TERMINATE, false, uint32(cmd.Process.Pid))
	if e != nil {
		return fail(e)
	}
	e = windows.AssignProcessToJobObject(job, proc)
	windows.CloseHandle(proc)
	if e != nil {
		return fail(e)
	}
	snap, e := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPTHREAD, 0)
	if e != nil {
		return fail(e)
	}
	defer windows.CloseHandle(snap)
	entry := windows.ThreadEntry32{Size: uint32(unsafe.Sizeof(windows.ThreadEntry32{}))}
	resumed := false
	for e = windows.Thread32First(snap, &entry); e == nil; e = windows.Thread32Next(snap, &entry) {
		if entry.OwnerProcessID != uint32(cmd.Process.Pid) {
			continue
		}
		thread, e := windows.OpenThread(windows.THREAD_SUSPEND_RESUME, false, entry.ThreadID)
		if e != nil {
			return fail(e)
		}
		_, e = windows.ResumeThread(thread)
		windows.CloseHandle(thread)
		if e != nil {
			return fail(e)
		}
		resumed = true
		break
	}
	if !resumed {
		return fail(errors.New("Could not resume the native slice worker"))
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case e := <-done:
		return e
	case <-ctx.Done():
		_ = windows.TerminateJobObject(job, 1)
		<-done
		return ctx.Err()
	}
}
