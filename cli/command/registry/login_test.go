package registry

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	configtypes "github.com/harness-community/docker-cli-v23/cli/config/types"
	"github.com/harness-community/docker-cli-v23/internal/test"
	"github.com/harness-community/docker-v23/api/types"
	registrytypes "github.com/harness-community/docker-v23/api/types/registry"
	"github.com/harness-community/docker-v23/client"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
	"gotest.tools/v3/fs"
)

const (
	userErr        = "userunknownError"
	testAuthErrMsg = "UNKNOWN_ERR"
)

var testAuthErrors = map[string]error{
	userErr: fmt.Errorf(testAuthErrMsg),
}

var (
	expiredPassword = "I_M_EXPIRED"
	useToken        = "I_M_TOKEN"
)

type fakeClient struct {
	client.Client
}

func (c fakeClient) Info(ctx context.Context) (types.Info, error) {
	return types.Info{}, nil
}

func (c fakeClient) RegistryLogin(ctx context.Context, auth types.AuthConfig) (registrytypes.AuthenticateOKBody, error) {
	if auth.Password == expiredPassword {
		return registrytypes.AuthenticateOKBody{}, fmt.Errorf("Invalid Username or Password")
	}
	if auth.Password == useToken {
		return registrytypes.AuthenticateOKBody{
			IdentityToken: auth.Password,
		}, nil
	}
	err := testAuthErrors[auth.Username]
	return registrytypes.AuthenticateOKBody{}, err
}

func TestLoginWithCredStoreCreds(t *testing.T) {
	testCases := []struct {
		inputAuthConfig types.AuthConfig
		expectedMsg     string
		expectedErr     string
	}{
		{
			inputAuthConfig: types.AuthConfig{},
			expectedMsg:     "Authenticating with existing credentials...\n",
		},
		{
			inputAuthConfig: types.AuthConfig{
				Username: userErr,
			},
			expectedMsg: "Authenticating with existing credentials...\n",
			expectedErr: fmt.Sprintf("Login did not succeed, error: %s\n", testAuthErrMsg),
		},
	}
	ctx := context.Background()
	for _, tc := range testCases {
		cli := test.NewFakeCli(&fakeClient{})
		errBuf := new(bytes.Buffer)
		cli.SetErr(errBuf)
		loginWithCredStoreCreds(ctx, cli, &tc.inputAuthConfig)
		outputString := cli.OutBuffer().String()
		assert.Check(t, is.Equal(tc.expectedMsg, outputString))
		errorString := errBuf.String()
		assert.Check(t, is.Equal(tc.expectedErr, errorString))
	}
}

func TestRunLogin(t *testing.T) {
	const storedServerAddress = "reg1"
	const validUsername = "u1"
	const validPassword = "p1"
	const validPassword2 = "p2"

	validAuthConfig := configtypes.AuthConfig{
		ServerAddress: storedServerAddress,
		Username:      validUsername,
		Password:      validPassword,
	}
	expiredAuthConfig := configtypes.AuthConfig{
		ServerAddress: storedServerAddress,
		Username:      validUsername,
		Password:      expiredPassword,
	}
	validIdentityToken := configtypes.AuthConfig{
		ServerAddress: storedServerAddress,
		Username:      validUsername,
		IdentityToken: useToken,
	}
	testCases := []struct {
		inputLoginOption  loginOptions
		inputStoredCred   *configtypes.AuthConfig
		expectedErr       string
		expectedSavedCred configtypes.AuthConfig
	}{
		{
			inputLoginOption: loginOptions{
				serverAddress: storedServerAddress,
			},
			inputStoredCred:   &validAuthConfig,
			expectedErr:       "",
			expectedSavedCred: validAuthConfig,
		},
		{
			inputLoginOption: loginOptions{
				serverAddress: storedServerAddress,
			},
			inputStoredCred: &expiredAuthConfig,
			expectedErr:     "Error: Cannot perform an interactive login from a non TTY device",
		},
		{
			inputLoginOption: loginOptions{
				serverAddress: storedServerAddress,
				user:          validUsername,
				password:      validPassword2,
			},
			inputStoredCred: &validAuthConfig,
			expectedErr:     "",
			expectedSavedCred: configtypes.AuthConfig{
				ServerAddress: storedServerAddress,
				Username:      validUsername,
				Password:      validPassword2,
			},
		},
		{
			inputLoginOption: loginOptions{
				serverAddress: storedServerAddress,
				user:          userErr,
				password:      validPassword,
			},
			inputStoredCred: &validAuthConfig,
			expectedErr:     testAuthErrMsg,
		},
		{
			inputLoginOption: loginOptions{
				serverAddress: storedServerAddress,
				user:          validUsername,
				password:      useToken,
			},
			inputStoredCred:   &validIdentityToken,
			expectedErr:       "",
			expectedSavedCred: validIdentityToken,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			tmpFile := fs.NewFile(t, "test-run-login")
			defer tmpFile.Remove()
			cli := test.NewFakeCli(&fakeClient{})
			configfile := cli.ConfigFile()
			configfile.Filename = tmpFile.Path()

			if tc.inputStoredCred != nil {
				cred := *tc.inputStoredCred
				assert.NilError(t, configfile.GetCredentialsStore(cred.ServerAddress).Store(cred))
			}
			loginErr := runLogin(cli, tc.inputLoginOption)
			if tc.expectedErr != "" {
				assert.Error(t, loginErr, tc.expectedErr)
				return
			}
			assert.NilError(t, loginErr)
			savedCred, credStoreErr := configfile.GetCredentialsStore(tc.inputStoredCred.ServerAddress).Get(tc.inputStoredCred.ServerAddress)
			assert.Check(t, credStoreErr)
			assert.DeepEqual(t, tc.expectedSavedCred, savedCred)
		})
	}
}
