// Copyright 2026 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build linux

package vfs

import (
	"testing"

	mount "github.com/moby/sys/mountinfo"
)

func TestCanHandle(t *testing.T) {
	p := NewPlugin()

	tests := []struct {
		fsType string
		want   bool
	}{
		// Block-backed filesystems that use plain statfs.
		{"ext2", true},
		{"ext3", true},
		{"ext4", true},
		{"xfs", true},
		{"f2fs", true},
		{"bcachefs", true},
		// Filesystems owned by other plugins.
		{"btrfs", false},
		{"zfs", false},
		{"overlay", false},
		{"tmpfs", false},
		{"nfs4", false},
		// Pseudo filesystems.
		{"proc", false},
		{"sysfs", false},
		{"cgroup2", false},
		// Unknown.
		{"somethingelse", false},
		{"", false},
	}

	for _, tt := range tests {
		if got := p.CanHandle(tt.fsType); got != tt.want {
			t.Errorf("CanHandle(%q) = %v, want %v", tt.fsType, got, tt.want)
		}
	}
}

// bcachefs reports a multi-device source ("/dev/sda:/dev/sdb") in mountinfo.
// The VFS plugin must pass it through untouched: it is only ever used as a
// map key by processMounts and is never parsed as a path.
func TestProcessMountMultiDeviceSource(t *testing.T) {
	p := NewPlugin()
	mnt := &mount.Info{
		Root:       "/",
		Mountpoint: "/",
		Source:     "/dev/sda:/dev/sdb",
		FSType:     "bcachefs",
		Major:      253,
		Minor:      2,
	}

	include, got, err := p.ProcessMount(mnt)
	if err != nil {
		t.Fatalf("ProcessMount returned error: %v", err)
	}
	if !include {
		t.Fatal("ProcessMount excluded a bcachefs mount")
	}
	if got.Source != mnt.Source {
		t.Errorf("ProcessMount rewrote Source: got %q, want %q", got.Source, mnt.Source)
	}
}
