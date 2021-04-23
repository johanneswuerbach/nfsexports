package nfsexports

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

const (
	defaultExportsFile = "/etc/exports"
)

// Add export, if exportsFile is an empty string /etc/exports is used
func Add(exportsFile string, identifier string, export string) ([]byte, error) {
	if exportsFile == "" {
		exportsFile = defaultExportsFile
	}

	exports, err := ioutil.ReadFile(exportsFile)

	if err != nil {
		if os.IsNotExist(err) {
			exports = []byte{}
		} else {
			return nil, err
		}
	}

	if containsExport(exports, identifier) {
		return exports, nil
	}

	newExports := exports
	if len(newExports) > 0 && !bytes.HasSuffix(exports, []byte("\n")) {
		newExports = append(newExports, '\n')
	}

	newExports = append(newExports, []byte(exportEntry(identifier, export))...)

	if err := verifyNewExports(newExports); err != nil {
		return nil, err
	}

	if err := ioutil.WriteFile(exportsFile, newExports, 0644); err != nil {
		return nil, err
	}

	return newExports, nil
}

// Remove export, if exportsFile is an empty string /etc/exports is used
func Remove(exportsFile string, identifier string) ([]byte, error) {
	if exportsFile == "" {
		exportsFile = defaultExportsFile
	}

	exports, err := ioutil.ReadFile(exportsFile)
	if err != nil {
		return nil, err
	}

	beginMark := []byte(fmt.Sprintf("# BEGIN: %s", identifier))
	endMark := []byte(fmt.Sprintf("# END: %s\n", identifier))

	begin := bytes.Index(exports, beginMark)
	end := bytes.Index(exports, endMark)

	if begin == -1 || end == -1 {
		return nil, fmt.Errorf("Couldn't not find export %s in %s", identifier, exportsFile)
	}

	newExports := append(exports[:begin], exports[end+len(endMark):]...)
	newExports = append(bytes.TrimSpace(newExports), '\n')

	if err := ioutil.WriteFile(exportsFile, newExports, 0644); err != nil {
		return nil, err
	}

	return newExports, nil
}

// Exists checks the existence of a given export
// The export must, however, have been created by this library using Add
func Exists(exportsFile string, identifier string) (bool, error) {
	if exportsFile == "" {
		exportsFile = defaultExportsFile
	}

	exports, err := ioutil.ReadFile(exportsFile)
	if err != nil {
		return false, err
	}

	beginMark := []byte(fmt.Sprintf("# BEGIN: %s", identifier))
	endMark := []byte(fmt.Sprintf("# END: %s\n", identifier))

	begin := bytes.Index(exports, beginMark)
	end := bytes.Index(exports, endMark)

	if begin == -1 || end == -1 {
		return false, nil
	}

	return true, nil
}

// List returns the list of exports *created by* nfsexports
// This means other exports might be present in the file but won't
// be returned by this function
func List(exportsFile string) (map[string]string, error) {
	if exportsFile == "" {
		exportsFile = defaultExportsFile
	}

	f, err := os.Open(exportsFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	exports := map[string]string{}
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Index(line, "# BEGIN:") != -1 {
			if scanner.Scan() != false {
				id := strings.TrimLeft(line, "# BEGIN:")
				export := scanner.Text()
				exports[id] = export
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return exports, nil
}

// ListAll returns all nfsexports present in the exports file.
// ListAll does not check the validity of the exports;
// It simply returns any line present in the file that is not a comment
func ListAll(exportsFile string) ([]string, error) {
	if exportsFile == "" {
		exportsFile = defaultExportsFile
	}

	f, err := os.Open(exportsFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	exports := []string{}
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Index(line, "#") != -1 || len(line) == 0 {
			continue
		}
		export := scanner.Text()
		exports = append(exports, export)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return exports, nil
}

// ReloadDaemon reload NFS daemon
func ReloadDaemon() error {
	cmd := exec.Command("sudo", "/sbin/nfsd", "update")
	cmd.Stderr = &bytes.Buffer{}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Reloading nfsd failed: %s\n%s", err.Error(), cmd.Stderr)
	}

	return nil
}

func containsExport(exports []byte, identifier string) bool {
	return bytes.Contains(exports, []byte(fmt.Sprintf("# BEGIN: %s\n", identifier)))
}

func exportEntry(identifier string, export string) string {
	return fmt.Sprintf("# BEGIN: %s\n%s\n# END: %s\n", identifier, export, identifier)
}

func verifyNewExports(newExports []byte) error {
	tmpFile, err := ioutil.TempFile("", "exports")
	if err != nil {
		return err
	}
	defer tmpFile.Close()

	if _, err := tmpFile.Write(newExports); err != nil {
		return err
	}
	tmpFile.Close()

	cmd := exec.Command("/sbin/nfsd", "-F", tmpFile.Name(), "checkexports")
	cmd.Stderr = &bytes.Buffer{}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Export verification failed:\n%s\n%s", cmd.Stderr, err.Error())
	}

	return nil
}
