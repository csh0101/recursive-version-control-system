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
	"errors"
	"fmt"
	"io"
	"os/exec"

	"github.com/google/recursive-version-control-system/snapshot"
	"github.com/google/recursive-version-control-system/storage"
)

func Sign(ctx context.Context, s *storage.LocalFiles, id *snapshot.Identity, h *snapshot.Hash, prevSignature *snapshot.Hash) (*snapshot.Hash, error) {
	if id == nil {
		return nil, errors.New("identity must not be nil")
	}
	if h == nil {
		return nil, errors.New("cannot sign a nil hash")
	}
	helperCommand := fmt.Sprintf("rvcs-sign-%s", id.Algorithm())
	args := []string{id.Contents(), h.String()}
	if prevSignature != nil {
		args = append(args, prevSignature.String())
	}
	signCmd := exec.Command(helperCommand, args...)
	stdout, err := signCmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failure constructing the sign command for %q: %v", helperCommand, err)
	}
	if err := signCmd.Start(); err != nil {
		return nil, fmt.Errorf("failure running the sign helper %q: %v", helperCommand, err)
	}
	outBytes, err := io.ReadAll(stdout)
	if err != nil {
		return nil, fmt.Errorf("failure reading the stdout of the sign helper %q: %v", helperCommand, err)
	}
	h, err = snapshot.ParseHash(string(outBytes))
	if err != nil {
		return nil, fmt.Errorf("failure parsing the stdout of the sign helper %q: %v", helperCommand, err)
	}
	if err := s.UpdateSignatureForIdentity(ctx, id, h); err != nil {
		return nil, fmt.Errorf("failure updating the latest snapshot for %q to %q: %v", id, h, err)
	}
	return h, nil
}
