// +build windows
// +build !appengine

package isatty

import (
	"errors"
	"strings"
	"syscall"
	"unsafe"
)

const (
	objectNameInfo uintptr = 1
	fileNameInfo           = 2
	fileTypePipe           = 3
)

var (
	kernel32                         = syscall.NewLazyDLL("kernel32.dll")
	ntdll                            = syscall.NewLazyDLL("ntdll.dll")
	procGetConsoleMode               = kernel32.NewProc("GetConsoleMode")
	procGetFileInformationByHandleEx = kernel32.NewProc("GetFileInformationByHandleEx")
	procGetFileType                  = kernel32.NewProc("GetFileType")
	procNtQueryObject                = ntdll.NewProc("NtQueryObject")
)

func init() {
	// Check if GetFileInformationByHandleEx is available.
	if procGetFileInformationByHandleEx.Find() != nil {
		procGetFileInformationByHandleEx = nil
	}
}

// IsTerminal return true if the file descriptor is terminal.
func IsTerminal(fd uintptr) bool {
	var st uint32
	r, _, e := syscall.Syscall(procGetConsoleMode.Addr(), 2, fd, uintptr(unsafe.Pointer(&st)), 0)
	return r != 0 && e == 0
}

// Check pipe name is used for cygwin/msys2 pty.
// Cygwin/MSYS2 PTY has a name like:
//   \{cygwin,msys}-XXXXXXXXXXXXXXXX-ptyN-{from,to}-master
func isCygwinPipeName(name string) bool {
	token := strings.Split(name, "-")
	if len(token) < 5 {
		return false
	}

	if token[0] != `\msys` &&
		token[0] != `\cygwin` &&
		token[0] != `\Device\NamedPipe\msys` &&
		token[0] != `\Device\NamedPipe\cygwin` {
		return false
	}

	if token[1] == "" {
		return false
	}

	if !strings.HasPrefix(token[2], "pty") {
		return false
	}

	if token[3] != `from` && token[3] != `to` {
		return false
	}

	if token[4] != "master" {
		return false
	}

	return true
}

// getFileNameByHandle use the undocomented ntdll NtQueryObject to get file full name from file handler
// since GetFileInformationByHandleEx is not avilable under windows Vista and still some old fashion
// guys are using Windows XP, this is a workaround for those guys, it will also work on system from
// Windows vista to 10
// see https://stackoverflow.com/a/18792477 for details
func getFileNameByHandle(fd uintptr) (string, error) {
	if procNtQueryObject == nil {
		return "", errors.New("ntdll.dll: NtQueryObject not supported")
	}
	var buf [2 * syscall.MAX_PATH]uint16
	var result uint16
	r, _, e := syscall.Syscall6(procNtQueryObject.Addr(), 5,
		fd, objectNameInfo, uintptr(unsafe.Pointer(&buf)), uintptr(2*len(buf)), uintptr(unsafe.Pointer(&result)), 0)

	if r != 0 {
		return "", e
	}
	//  From the document:
	// If the given object is named and the object name was successfully acquired, the returned string is the name
	// of the given object including as much of the object's full path as possible. In this case,
	// NtQueryObject sets Name.Buffer to the address of the NULL-terminated name of the specified object.
	// The value of Name.MaximumLength is the length of the object name including the NULL termination.
	// The value of Name.Length is length of the only the object name.
	// If the given object is unnamed, or if the object name was not successfully acquired,
	// NtQueryObject sets Name.Buffer to NULL and sets Name.Length and Name.MaximumLength to zero.
	maxLength := buf[1] // Name.MaximumLength

	if maxLength == 0 || result <= maxLength {
		return "", errors.New("cannot acquired file info or unnamed object")
	}

	// It is stupid that microsoft also puts some grabge after the Length and MaximumLength, offset is calculated
	// to skip those grabges
	offset := (result - maxLength) / 2
	end := offset + maxLength/2
	return syscall.UTF16ToString(buf[offset:end]), nil
}

// getFileNameByHandleEx using GetFileInformationByHandleEx syscall to get file name from file handler
func getFileNameByHandleEx(fd uintptr) (string, error) {
	if procGetFileInformationByHandleEx == nil {
		return "", errors.New("kernel32.dll: GetFileInformationByHandleEx not supported")
	}
	// Cygwin/msys's pty is a pipe.
	ft, _, e := syscall.Syscall(procGetFileType.Addr(), 1, fd, 0, 0)
	if ft != fileTypePipe || e != 0 {
		return "", e
	}

	var buf [2 + syscall.MAX_PATH]uint16
	r, _, e := syscall.Syscall6(procGetFileInformationByHandleEx.Addr(),
		4, fd, fileNameInfo, uintptr(unsafe.Pointer(&buf)),
		uintptr(len(buf)*2), 0, 0)
	if r == 0 || e != 0 {
		return "", e
	}

	l := *(*uint32)(unsafe.Pointer(&buf))
	return syscall.UTF16ToString(buf[2 : 2+l/2]), nil
}

// IsCygwinTerminal() return true if the file descriptor is a cygwin or msys2
// terminal.
func IsCygwinTerminal(fd uintptr) bool {
	name, err := getFileNameByHandle(fd)

	if err == nil && len(name) > 0 {
		return isCygwinPipeName(name)
	}

	name, err = getFileNameByHandleEx(fd)

	if err == nil && len(name) > 0 {
		return isCygwinPipeName(name)
	}

	return false
}
