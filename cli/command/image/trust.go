package image

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"github.com/harness-community/docker-cli-v23/cli/command"
	"github.com/harness-community/docker-cli-v23/cli/streams"
	"github.com/harness-community/docker-cli-v23/cli/trust"
	"github.com/docker/distribution/reference"
	"github.com/harness-community/docker-v23/api/types"
	registrytypes "github.com/harness-community/docker-v23/api/types/registry"
	"github.com/harness-community/docker-v23/pkg/jsonmessage"
	"github.com/harness-community/docker-v23/registry"
	"github.com/opencontainers/go-digest"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/theupdateframework/notary/client"
	"github.com/theupdateframework/notary/tuf/data"
)

type target struct {
	name   string
	digest digest.Digest
	size   int64
}

// TrustedPush handles content trust pushing of an image
func TrustedPush(ctx context.Context, cli command.Cli, repoInfo *registry.RepositoryInfo, ref reference.Named, authConfig types.AuthConfig, options types.ImagePushOptions) error {
	responseBody, err := cli.Client().ImagePush(ctx, reference.FamiliarString(ref), options)
	if err != nil {
		return err
	}

	defer responseBody.Close()

	return PushTrustedReference(cli, repoInfo, ref, authConfig, responseBody)
}

// PushTrustedReference pushes a canonical reference to the trust server.
//
//nolint:gocyclo
func PushTrustedReference(streams command.Streams, repoInfo *registry.RepositoryInfo, ref reference.Named, authConfig types.AuthConfig, in io.Reader) error {
	// If it is a trusted push we would like to find the target entry which match the
	// tag provided in the function and then do an AddTarget later.
	target := &client.Target{}
	// Count the times of calling for handleTarget,
	// if it is called more that once, that should be considered an error in a trusted push.
	cnt := 0
	handleTarget := func(msg jsonmessage.JSONMessage) {
		cnt++
		if cnt > 1 {
			// handleTarget should only be called once. This will be treated as an error.
			return
		}

		var pushResult types.PushResult
		err := json.Unmarshal(*msg.Aux, &pushResult)
		if err == nil && pushResult.Tag != "" {
			if dgst, err := digest.Parse(pushResult.Digest); err == nil {
				h, err := hex.DecodeString(dgst.Hex())
				if err != nil {
					target = nil
					return
				}
				target.Name = pushResult.Tag
				target.Hashes = data.Hashes{string(dgst.Algorithm()): h}
				target.Length = int64(pushResult.Size)
			}
		}
	}

	var tag string
	switch x := ref.(type) {
	case reference.Canonical:
		return errors.New("cannot push a digest reference")
	case reference.NamedTagged:
		tag = x.Tag()
	default:
		// We want trust signatures to always take an explicit tag,
		// otherwise it will act as an untrusted push.
		if err := jsonmessage.DisplayJSONMessagesToStream(in, streams.Out(), nil); err != nil {
			return err
		}
		fmt.Fprintln(streams.Err(), "No tag specified, skipping trust metadata push")
		return nil
	}

	if err := jsonmessage.DisplayJSONMessagesToStream(in, streams.Out(), handleTarget); err != nil {
		return err
	}

	if cnt > 1 {
		return errors.Errorf("internal error: only one call to handleTarget expected")
	}

	if target == nil {
		return errors.Errorf("no targets found, please provide a specific tag in order to sign it")
	}

	fmt.Fprintln(streams.Out(), "Signing and pushing trust metadata")

	repo, err := trust.GetNotaryRepository(streams.In(), streams.Out(), command.UserAgent(), repoInfo, &authConfig, "push", "pull")
	if err != nil {
		return errors.Wrap(err, "error establishing connection to trust repository")
	}

	// get the latest repository metadata so we can figure out which roles to sign
	_, err = repo.ListTargets()

	switch err.(type) {
	case client.ErrRepoNotInitialized, client.ErrRepositoryNotExist:
		keys := repo.GetCryptoService().ListKeys(data.CanonicalRootRole)
		var rootKeyID string
		// always select the first root key
		if len(keys) > 0 {
			sort.Strings(keys)
			rootKeyID = keys[0]
		} else {
			rootPublicKey, err := repo.GetCryptoService().Create(data.CanonicalRootRole, "", data.ECDSAKey)
			if err != nil {
				return err
			}
			rootKeyID = rootPublicKey.ID()
		}

		// Initialize the notary repository with a remotely managed snapshot key
		if err := repo.Initialize([]string{rootKeyID}, data.CanonicalSnapshotRole); err != nil {
			return trust.NotaryError(repoInfo.Name.Name(), err)
		}
		fmt.Fprintf(streams.Out(), "Finished initializing %q\n", repoInfo.Name.Name())
		err = repo.AddTarget(target, data.CanonicalTargetsRole)
	case nil:
		// already initialized and we have successfully downloaded the latest metadata
		err = AddTargetToAllSignableRoles(repo, target)
	default:
		return trust.NotaryError(repoInfo.Name.Name(), err)
	}

	if err == nil {
		err = repo.Publish()
	}

	if err != nil {
		err = errors.Wrapf(err, "failed to sign %s:%s", repoInfo.Name.Name(), tag)
		return trust.NotaryError(repoInfo.Name.Name(), err)
	}

	fmt.Fprintf(streams.Out(), "Successfully signed %s:%s\n", repoInfo.Name.Name(), tag)
	return nil
}

// AddTargetToAllSignableRoles attempts to add the image target to all the top level delegation roles we can
// (based on whether we have the signing key and whether the role's path allows
// us to).
// If there are no delegation roles, we add to the targets role.
func AddTargetToAllSignableRoles(repo client.Repository, target *client.Target) error {
	signableRoles, err := trust.GetSignableRoles(repo, target)
	if err != nil {
		return err
	}

	return repo.AddTarget(target, signableRoles...)
}

// trustedPull handles content trust pulling of an image
func trustedPull(ctx context.Context, cli command.Cli, imgRefAndAuth trust.ImageRefAndAuth, opts PullOptions) error {
	refs, err := getTrustedPullTargets(cli, imgRefAndAuth)
	if err != nil {
		return err
	}

	ref := imgRefAndAuth.Reference()
	for i, r := range refs {
		displayTag := r.name
		if displayTag != "" {
			displayTag = ":" + displayTag
		}
		fmt.Fprintf(cli.Out(), "Pull (%d of %d): %s%s@%s\n", i+1, len(refs), reference.FamiliarName(ref), displayTag, r.digest)

		trustedRef, err := reference.WithDigest(reference.TrimNamed(ref), r.digest)
		if err != nil {
			return err
		}
		updatedImgRefAndAuth, err := trust.GetImageReferencesAndAuth(ctx, nil, AuthResolver(cli), trustedRef.String())
		if err != nil {
			return err
		}
		if err := imagePullPrivileged(ctx, cli, updatedImgRefAndAuth, PullOptions{
			all:      false,
			platform: opts.platform,
			quiet:    opts.quiet,
			remote:   opts.remote,
		}); err != nil {
			return err
		}

		tagged, err := reference.WithTag(reference.TrimNamed(ref), r.name)
		if err != nil {
			return err
		}

		if err := TagTrusted(ctx, cli, trustedRef, tagged); err != nil {
			return err
		}
	}
	return nil
}

func getTrustedPullTargets(cli command.Cli, imgRefAndAuth trust.ImageRefAndAuth) ([]target, error) {
	notaryRepo, err := cli.NotaryClient(imgRefAndAuth, trust.ActionsPullOnly)
	if err != nil {
		return nil, errors.Wrap(err, "error establishing connection to trust repository")
	}

	ref := imgRefAndAuth.Reference()
	tagged, isTagged := ref.(reference.NamedTagged)
	if !isTagged {
		// List all targets
		targets, err := notaryRepo.ListTargets(trust.ReleasesRole, data.CanonicalTargetsRole)
		if err != nil {
			return nil, trust.NotaryError(ref.Name(), err)
		}
		var refs []target
		for _, tgt := range targets {
			t, err := convertTarget(tgt.Target)
			if err != nil {
				fmt.Fprintf(cli.Err(), "Skipping target for %q\n", reference.FamiliarName(ref))
				continue
			}
			// Only list tags in the top level targets role or the releases delegation role - ignore
			// all other delegation roles
			if tgt.Role != trust.ReleasesRole && tgt.Role != data.CanonicalTargetsRole {
				continue
			}
			refs = append(refs, t)
		}
		if len(refs) == 0 {
			return nil, trust.NotaryError(ref.Name(), errors.Errorf("No trusted tags for %s", ref.Name()))
		}
		return refs, nil
	}

	t, err := notaryRepo.GetTargetByName(tagged.Tag(), trust.ReleasesRole, data.CanonicalTargetsRole)
	if err != nil {
		return nil, trust.NotaryError(ref.Name(), err)
	}
	// Only get the tag if it's in the top level targets role or the releases delegation role
	// ignore it if it's in any other delegation roles
	if t.Role != trust.ReleasesRole && t.Role != data.CanonicalTargetsRole {
		return nil, trust.NotaryError(ref.Name(), errors.Errorf("No trust data for %s", tagged.Tag()))
	}

	logrus.Debugf("retrieving target for %s role", t.Role)
	r, err := convertTarget(t.Target)
	return []target{r}, err
}

// imagePullPrivileged pulls the image and displays it to the output
func imagePullPrivileged(ctx context.Context, cli command.Cli, imgRefAndAuth trust.ImageRefAndAuth, opts PullOptions) error {
	ref := reference.FamiliarString(imgRefAndAuth.Reference())

	encodedAuth, err := command.EncodeAuthToBase64(*imgRefAndAuth.AuthConfig())
	if err != nil {
		return err
	}
	requestPrivilege := command.RegistryAuthenticationPrivilegedFunc(cli, imgRefAndAuth.RepoInfo().Index, "pull")
	options := types.ImagePullOptions{
		RegistryAuth:  encodedAuth,
		PrivilegeFunc: requestPrivilege,
		All:           opts.all,
		Platform:      opts.platform,
	}
	responseBody, err := cli.Client().ImagePull(ctx, ref, options)
	if err != nil {
		return err
	}
	defer responseBody.Close()

	out := cli.Out()
	if opts.quiet {
		out = streams.NewOut(io.Discard)
	}
	return jsonmessage.DisplayJSONMessagesToStream(responseBody, out, nil)
}

// TrustedReference returns the canonical trusted reference for an image reference
func TrustedReference(ctx context.Context, cli command.Cli, ref reference.NamedTagged, rs registry.Service) (reference.Canonical, error) {
	imgRefAndAuth, err := trust.GetImageReferencesAndAuth(ctx, rs, AuthResolver(cli), ref.String())
	if err != nil {
		return nil, err
	}

	notaryRepo, err := cli.NotaryClient(imgRefAndAuth, []string{"pull"})
	if err != nil {
		return nil, errors.Wrap(err, "error establishing connection to trust repository")
	}

	t, err := notaryRepo.GetTargetByName(ref.Tag(), trust.ReleasesRole, data.CanonicalTargetsRole)
	if err != nil {
		return nil, trust.NotaryError(imgRefAndAuth.RepoInfo().Name.Name(), err)
	}
	// Only list tags in the top level targets role or the releases delegation role - ignore
	// all other delegation roles
	if t.Role != trust.ReleasesRole && t.Role != data.CanonicalTargetsRole {
		return nil, trust.NotaryError(imgRefAndAuth.RepoInfo().Name.Name(), client.ErrNoSuchTarget(ref.Tag()))
	}
	r, err := convertTarget(t.Target)
	if err != nil {
		return nil, err
	}
	return reference.WithDigest(reference.TrimNamed(ref), r.digest)
}

func convertTarget(t client.Target) (target, error) {
	h, ok := t.Hashes["sha256"]
	if !ok {
		return target{}, errors.New("no valid hash, expecting sha256")
	}
	return target{
		name:   t.Name,
		digest: digest.NewDigestFromHex("sha256", hex.EncodeToString(h)),
		size:   t.Length,
	}, nil
}

// TagTrusted tags a trusted ref
func TagTrusted(ctx context.Context, cli command.Cli, trustedRef reference.Canonical, ref reference.NamedTagged) error {
	// Use familiar references when interacting with client and output
	familiarRef := reference.FamiliarString(ref)
	trustedFamiliarRef := reference.FamiliarString(trustedRef)

	fmt.Fprintf(cli.Err(), "Tagging %s as %s\n", trustedFamiliarRef, familiarRef)

	return cli.Client().ImageTag(ctx, trustedFamiliarRef, familiarRef)
}

// AuthResolver returns an auth resolver function from a command.Cli
func AuthResolver(cli command.Cli) func(ctx context.Context, index *registrytypes.IndexInfo) types.AuthConfig {
	return func(ctx context.Context, index *registrytypes.IndexInfo) types.AuthConfig {
		return command.ResolveAuthConfig(ctx, cli, index)
	}
}
