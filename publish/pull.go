// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package publish defines methods for publishing rvcs snapshots.
package publish

import (
	"context"
	"fmt"
	"io"
	"os/exec"

	"github.com/google/recursive-version-control-system/config"
	"github.com/google/recursive-version-control-system/snapshot"
	"github.com/google/recursive-version-control-system/storage"
)

func pullFrom(ctx context.Context, m *config.Mirror, s *storage.LocalFiles, id *snapshot.Identity, prev *snapshot.Hash) (*snapshot.Hash, error) {
	if m == nil || m.URL == nil {
		return prev, nil
	}
	helperCommand := fmt.Sprintf("rvcs-pull-%s", m.URL.Scheme)
	args := m.HelperFlags
	args = append(args, id.String())
	if prev != nil {
		args = append(args, prev.String())
	}
	pullCmd := exec.Command(helperCommand, args...)
	stdout, err := pullCmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failure constructing the pull command for %q: %v", helperCommand, err)
	}
	if err := pullCmd.Start(); err != nil {
		return nil, fmt.Errorf("failure running the pull helper %q: %v", helperCommand, err)
	}
	outBytes, err := io.ReadAll(stdout)
	if err != nil {
		return nil, fmt.Errorf("failure reading the stdout of the pull helper %q: %v", helperCommand, err)
	}
	h, err := snapshot.ParseHash(string(outBytes))
	if err != nil {
		return nil, fmt.Errorf("failure parsing the stdout of the pull helper %q: %v", helperCommand, err)
	}
	return h, nil
}

func pullFromAndVerify(ctx context.Context, m *config.Mirror, s *storage.LocalFiles, id *snapshot.Identity, prevSignature *snapshot.Hash, prevSigned *snapshot.Hash) (signature *snapshot.Hash, signed *snapshot.Hash, err error) {
	signature, err = pullFrom(ctx, m, s, id, prevSignature)
	if err != nil {
		return nil, nil, fmt.Errorf("failure pulling the latest snapshot for %q from %q: %v", id, m, err)
	}
	if signature.Equal(prevSignature) {
		return prevSignature, prevSigned, nil
	}
	signed, err = Verify(ctx, s, id, signature)
	if err != nil {
		return nil, nil, fmt.Errorf("failure verifying the signature for %q from %q: %v", id, m, err)
	}
	return signature, signed, nil
}

func Pull(ctx context.Context, settings *config.Settings, s *storage.LocalFiles, id *snapshot.Identity) (signature *snapshot.Hash, signed *snapshot.Hash, err error) {
	signature, err = s.LatestSignatureForIdentity(ctx, id)
	if err != nil {
		return nil, nil, fmt.Errorf("failure looking up the previous signature for %q: %v", id, err)
	}
	signed, err = Verify(ctx, s, id, signature)
	if err != nil {
		return nil, nil, fmt.Errorf("failure verifying the previous signature for %q: %v", id, err)
	}
	for _, idSetting := range settings.Identities {
		if idSetting.Name == id.String() {
			for _, mirror := range idSetting.PullMirrors {
				signature, signed, err = pullFromAndVerify(ctx, mirror, s, id, signature, signed)
				if err != nil {
					return nil, nil, fmt.Errorf("failure pulling the latest snapshot for %q from %q: %v", id, mirror, err)
				}
			}
		}
	}
	for _, mirror := range settings.AdditionalPullMirrors {
		signature, signed, err = pullFromAndVerify(ctx, mirror, s, id, signature, signed)
		if err != nil {
			return nil, nil, fmt.Errorf("failure pulling the latest snapshot for %q from %q: %v", id, mirror, err)
		}
	}
	if err := s.UpdateSignatureForIdentity(ctx, id, signature); err != nil {
		return nil, nil, fmt.Errorf("failure updating the latest snapshot for %q to %q: %v", id, signature, err)
	}
	return signature, signed, nil
}
