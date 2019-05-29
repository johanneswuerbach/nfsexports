package nfsexports

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"
)

func TestAddWithValid(t *testing.T) {
	exportsFile, err := exportsFile(`/Users 192.168.64.1 -alldirs -maproot=root`)
	if err != nil {
		t.Error("Failed creating test exports file", err)
	}

	result, err := Add(exportsFile, "my-id", "/Users 192.168.64.2 -alldirs -maproot=root")
	if err != nil {
		t.Error("Accepts additions resulting in a valid exports file", err)
	}

	if !bytes.Equal(result, []byte(`/Users 192.168.64.1 -alldirs -maproot=root
# BEGIN: my-id
/Users 192.168.64.2 -alldirs -maproot=root
# END: my-id
`)) {
		t.Error("Generates an expected result", string(result))
	}
}

func TestAddWithExistingIdentifier(t *testing.T) {
	exportsFile, err := exportsFile(`/Users 192.168.64.1 -alldirs -maproot=root
# BEGIN: my-id
/Users 192.168.64.2 -alldirs -maproot=root
# END: my-id
`)
	if err != nil {
		t.Error("Failed creating test exports file", err)
	}

	result, err := Add(exportsFile, "my-id", "/Users 192.168.64.2 -alldirs -maproot=root")
	if err != nil {
		t.Error("Accepts additions resulting in a valid exports file", err)
	}

	if !bytes.Equal(result, []byte(`/Users 192.168.64.1 -alldirs -maproot=root
# BEGIN: my-id
/Users 192.168.64.2 -alldirs -maproot=root
# END: my-id
`)) {
		t.Error("Generates an expected result", string(result))
	}
}

func TestAddWithInvalid(t *testing.T) {
	exportsFile, err := exportsFile(`/Users/my-user 192.168.64.1 -alldirs -maproot=root
`)
	if err != nil {
		t.Error("Failed creating test exports file", err)
	}

	result, err := Add(exportsFile, "my-id", "/Users 192.168.64.2 -alldirs -maproot=root")
	if err == nil {
		t.Error("Rejects additions resulting in an invalid exports file", err)
	}

	if result != nil {
		t.Error("Returns no result", string(result))
	}
}

func TestCheckExistsWithValid(t *testing.T) {
	exportsFile, err := exportsFile(`/Users 192.168.64.1 -alldirs -maproot=root
# BEGIN: my-id
/Users 192.168.64.2 -alldirs -maproot=root
# END: my-id
`)
	if err != nil {
		t.Error("Failed creating test exports file", err)
	}

	result, err := Exists(exportsFile, "my-id")
	if err != nil {
		t.Error("Checking existence of valid exports fails", err)
	} else if result == false {
		t.Error("Checking existence of valid exports returned false", result)
	}
}

func TestCheckExistsWithInvalid(t *testing.T) {
	exportsFile, err := exportsFile(`/Users 192.168.64.1 -alldirs -maproot=root
# BEGIN: my-id
/Users 192.168.64.2 -alldirs -maproot=root
# END: my-id
`)
	if err != nil {
		t.Error("Failed creating test exports file", err)
	}

	result, err := Exists(exportsFile, "my-invalid-id")
	if err != nil {
		t.Error("Checking existence of invalid exports fails", err)
	} else if result == true {
		t.Error("Checking existence of invalid exports returned true", result)
	}
}

func TestList(t *testing.T) {

	expected := map[string]string{
		"my-id":  "/Users 192.168.64.2 -alldirs -maproot=root",
		"my-id1": "/Users 192.168.64.3 -alldirs -maproot=root",
		"my-id2": "/Users 192.168.64.4 -alldirs -maproot=root",
	}

	contents := "/Users 192.168.64.1 -alldirs -maproot=root"
	for id, export := range expected {
		contents = fmt.Sprintf("%s\n# BEGIN: %s\n%s\n# END: %s", contents, id, export, id)
	}
	contents += "/Users 192.168.64.6 -alldirs -maproot=root"

	exportsFile, err := exportsFile(contents)
	if err != nil {
		t.Error("Failed creating test exports file", err)
	}

	exports, err := List(exportsFile)

	for id, export := range exports {
		if expected[id] != export {
			t.Error("nfsexport id", id, "not matching", export)
		}
	}
}

func TestListAll(t *testing.T) {
	exportsFile, err := exportsFile(`/Users 192.168.64.1 -alldirs -maproot=root
# BEGIN: my-id
/Users 192.168.64.2 -alldirs -maproot=root
# END: my-id
# BEGIN: my-id1
/Users 192.168.64.3 -alldirs -maproot=root
# END: my-id1
# BEGIN: my-id2
/Users 192.168.64.4 -alldirs -maproot=root
# END: my-id2

/Users 192.168.64.5 -alldirs -maproot=root
`)
	if err != nil {
		t.Error("Failed creating test exports file", err)
	}

	expected := map[string]bool{
		"/Users 192.168.64.1 -alldirs -maproot=root": true,
		"/Users 192.168.64.2 -alldirs -maproot=root": true,
		"/Users 192.168.64.3 -alldirs -maproot=root": true,
		"/Users 192.168.64.4 -alldirs -maproot=root": true,
		"/Users 192.168.64.5 -alldirs -maproot=root": true,
	}

	exports, err := ListAll(exportsFile)

	if len(exports) < len(expected) {
		t.Error("Missing NFS export")
	}

	for _, export := range exports {
		if _, ok := expected[export]; !ok {
			t.Error("Too many NFS exports returned (", export, ")")
		}
	}

}

func TestAddNewFile(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "nfsexports")
	if err != nil {
		t.Error("Failed creating test exports dir", err)
	}

	exportsFile := fmt.Sprintf("%s/exports", tempDir)
	result, err := Add(exportsFile, "my-id", "/Users 192.168.64.2 -alldirs -maproot=root")
	if err != nil {
		t.Error("Accepts additions to an new file", err)
	}

	if !bytes.Equal(result, []byte(`# BEGIN: my-id
/Users 192.168.64.2 -alldirs -maproot=root
# END: my-id
`)) {
		t.Error("Generates an expected result", string(result))
	}
}

func TestRemoveNotExisting(t *testing.T) {
	exportsFile, err := exportsFile(`/Users/my-user 192.168.64.1 -alldirs -maproot=root
`)
	if err != nil {
		t.Error("Failed creating test exports file", err)
	}

	result, err := Remove(exportsFile, "my-id")
	if err == nil {
		t.Error("Errors when removing an unknown identifier", err)
	}

	if result != nil {
		t.Error("Returns no result", string(result))
	}
}

func TestRemoveExisting(t *testing.T) {
	exportsFile, err := exportsFile(`/Users 192.168.64.1 -alldirs -maproot=root
# BEGIN: my-id
/Users 192.168.64.2 -alldirs -maproot=root
# END: my-id
`)
	if err != nil {
		t.Error("Failed creating test exports file", err)
	}

	result, err := Remove(exportsFile, "my-id")
	if err != nil {
		t.Error("Removes an known indentifier without error", err)
	}

	if !bytes.Equal(result, []byte(`/Users 192.168.64.1 -alldirs -maproot=root
`)) {
		t.Error("Generates an expected result", string(result))
	}
}

func TestRemoveLast(t *testing.T) {
	exportsFile, err := exportsFile(`/Users 192.168.64.1 -alldirs -maproot=root
# BEGIN: my-id
/Users 192.168.64.2 -alldirs -maproot=root
# END: my-id
`)
	if err != nil {
		t.Error("Failed creating test exports file", err)
	}

	result, err := Remove(exportsFile, "my-id")
	if err != nil {
		t.Error("Removes an known indentifier without error", err)
	}

	if !bytes.Equal(result, []byte(`/Users 192.168.64.1 -alldirs -maproot=root
`)) {
		t.Error("Generates an expected result", string(result))
	}
}

func TestReloadDaemon(t *testing.T) {
	err := ReloadDaemon()
	if err != nil {
		t.Error("Allows to reload nfsd", err)
	}
}

func exportsFile(content string) (string, error) {
	tmpFile, err := ioutil.TempFile("", "exports-test")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	if _, err := tmpFile.Write([]byte(content)); err != nil {
		return "", err
	}
	tmpFile.Close()

	return tmpFile.Name(), nil
}
