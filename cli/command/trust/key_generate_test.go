package trust

import (
	"encoding/pem"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/harness-community/docker-cli-v23/cli/config"
	"github.com/harness-community/docker-cli-v23/internal/test"
	"github.com/theupdateframework/notary"
	"github.com/theupdateframework/notary/passphrase"
	"github.com/theupdateframework/notary/trustmanager"
	tufutils "github.com/theupdateframework/notary/tuf/utils"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestTrustKeyGenerateErrors(t *testing.T) {
	testCases := []struct {
		name          string
		args          []string
		expectedError string
	}{
		{
			name:          "not-enough-args",
			expectedError: "requires exactly 1 argument",
		},
		{
			name:          "too-many-args",
			args:          []string{"key-1", "key-2"},
			expectedError: "requires exactly 1 argument",
		},
	}

	config.SetDir(t.TempDir())

	for _, tc := range testCases {
		cli := test.NewFakeCli(&fakeClient{})
		cmd := newKeyGenerateCommand(cli)
		cmd.SetArgs(tc.args)
		cmd.SetOut(io.Discard)
		assert.ErrorContains(t, cmd.Execute(), tc.expectedError)
	}
}

func TestGenerateKeySuccess(t *testing.T) {
	pubKeyCWD := t.TempDir()
	privKeyStorageDir := t.TempDir()

	passwd := "password"
	cannedPasswordRetriever := passphrase.ConstantRetriever(passwd)
	// generate a single key
	keyName := "alice"
	privKeyFileStore, err := trustmanager.NewKeyFileStore(privKeyStorageDir, cannedPasswordRetriever)
	assert.NilError(t, err)

	pubKeyPEM, err := generateKeyAndOutputPubPEM(keyName, privKeyFileStore)
	assert.NilError(t, err)

	assert.Check(t, is.Equal(keyName, pubKeyPEM.Headers["role"]))
	// the default GUN is empty
	assert.Check(t, is.Equal("", pubKeyPEM.Headers["gun"]))
	// assert public key header
	assert.Check(t, is.Equal("PUBLIC KEY", pubKeyPEM.Type))

	// check that an appropriate ~/<trust_dir>/private/<key_id>.key file exists
	expectedPrivKeyDir := filepath.Join(privKeyStorageDir, notary.PrivDir)
	_, err = os.Stat(expectedPrivKeyDir)
	assert.NilError(t, err)

	keyFiles, err := os.ReadDir(expectedPrivKeyDir)
	assert.NilError(t, err)
	assert.Check(t, is.Len(keyFiles, 1))
	privKeyFilePath := filepath.Join(expectedPrivKeyDir, keyFiles[0].Name())

	// verify the key content
	privFrom, _ := os.OpenFile(privKeyFilePath, os.O_RDONLY, notary.PrivExecPerms)
	defer privFrom.Close()
	fromBytes, _ := io.ReadAll(privFrom)
	privKeyPEM, _ := pem.Decode(fromBytes)
	assert.Check(t, is.Equal(keyName, privKeyPEM.Headers["role"]))
	// the default GUN is empty
	assert.Check(t, is.Equal("", privKeyPEM.Headers["gun"]))
	// assert encrypted header
	assert.Check(t, is.Equal("ENCRYPTED PRIVATE KEY", privKeyPEM.Type))
	// check that the passphrase matches
	_, err = tufutils.ParsePKCS8ToTufKey(privKeyPEM.Bytes, []byte(passwd))
	assert.NilError(t, err)

	// check that the public key exists at the correct path if we use the helper:
	returnedPath, err := writePubKeyPEMToDir(pubKeyPEM, keyName, pubKeyCWD)
	assert.NilError(t, err)
	expectedPubKeyPath := filepath.Join(pubKeyCWD, keyName+".pub")
	assert.Check(t, is.Equal(returnedPath, expectedPubKeyPath))
	_, err = os.Stat(expectedPubKeyPath)
	assert.NilError(t, err)
	// check that the public key is the only file output in CWD
	cwdKeyFiles, err := os.ReadDir(pubKeyCWD)
	assert.NilError(t, err)
	assert.Check(t, is.Len(cwdKeyFiles, 1))
}

func TestValidateKeyArgs(t *testing.T) {
	pubKeyCWD := t.TempDir()

	err := validateKeyArgs("a", pubKeyCWD)
	assert.NilError(t, err)

	err = validateKeyArgs("a/b", pubKeyCWD)
	assert.Error(t, err, "key name \"a/b\" must start with lowercase alphanumeric characters and can include \"-\" or \"_\" after the first character")

	err = validateKeyArgs("-", pubKeyCWD)
	assert.Error(t, err, "key name \"-\" must start with lowercase alphanumeric characters and can include \"-\" or \"_\" after the first character")

	assert.NilError(t, os.WriteFile(filepath.Join(pubKeyCWD, "a.pub"), []byte("abc"), notary.PrivExecPerms))
	err = validateKeyArgs("a", pubKeyCWD)
	assert.Error(t, err, fmt.Sprintf("public key file already exists: \"%s\"", filepath.Join(pubKeyCWD, "a.pub")))

	err = validateKeyArgs("a", "/random/dir/")
	assert.Error(t, err, "public key path does not exist: \"/random/dir/\"")
}
