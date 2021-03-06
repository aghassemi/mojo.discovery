// Copyright 2016 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build mojo

package apptest

import (
	"testing"

	"mojo/public/go/application"
	"mojo/public/go/bindings"

	mojom "mojom/v.io/discovery"

	_ "v.io/x/ref/runtime/factories/generic"

	"v.io/mojo/discovery/internal"
)

func newDiscovery(mctx application.Context) *mojom.Discovery_Proxy {
	req, ptr := mojom.CreateMessagePipeForDiscovery()
	mctx.ConnectToApplication("https://mojo.v.io/discovery.mojo").ConnectToService(&req)
	return mojom.NewDiscoveryProxy(ptr, bindings.GetAsyncWaiter())
}

func AppTestDiscoveryBasic(t *testing.T, mctx application.Context) {
	ads := []mojom.Advertisement{
		{
			Id:            &[internal.AdIdLen]uint8{1, 2, 3},
			InterfaceName: "v.io/v23/a",
			Addresses:     []string{"/h1:123/x"},
			Attributes:    &map[string]string{"a1": "v"},
			Attachments:   &map[string][]byte{"a2": []byte{1}},
		},
		{
			InterfaceName: "v.io/v23/b",
			Addresses:     []string{"/h1:123/y"},
			Attributes:    &map[string]string{"b1": "w"},
			Attachments:   &map[string][]byte{"b2": []byte{2}},
		},
	}

	d1 := newDiscovery(mctx)
	defer d1.Close_Proxy()

	var stops []func()
	for i, ad := range ads {
		id, closer, e1, e2 := d1.Advertise(ad, nil)
		if e1 != nil || e2 != nil {
			t.Fatalf("ad[%d]: failed to advertise: %v, %v", i, e1, e2)
		}
		if id == nil {
			t.Errorf("ad[%d]: got nil id", i)
			continue
		}
		if ad.Id == nil {
			ads[i].Id = id
		} else if *id != *ad.Id {
			t.Errorf("ad[%d]: got ad id %v, but wanted %v", i, *id, *ad.Id)
		}

		stop := func() {
			p := mojom.NewCloserProxy(*closer, bindings.GetAsyncWaiter())
			p.Close()
			p.Close_Proxy()
		}
		stops = append(stops, stop)
	}

	// Make sure none of advertisements are discoverable by the same discovery instance.
	if err := scanAndMatch(d1, ``); err != nil {
		t.Error(err)
	}

	// Create a new discovery instance. All advertisements should be discovered with that.
	d2 := newDiscovery(mctx)
	defer d2.Close_Proxy()

	if err := scanAndMatch(d2, `v.InterfaceName="v.io/v23/a"`, ads[0]); err != nil {
		t.Error(err)
	}
	if err := scanAndMatch(d2, `v.InterfaceName="v.io/v23/b"`, ads[1]); err != nil {
		t.Error(err)
	}
	if err := scanAndMatch(d2, ``, ads...); err != nil {
		t.Error(err)
	}

	// Open a new scan channel and consume expected advertisements first.
	scanCh, scanStop, err := scan(d2, `v.InterfaceName="v.io/v23/a"`)
	if err != nil {
		t.Fatal(err)
	}
	defer scanStop()

	update := <-scanCh
	if err := matchFound([]mojom.Update_Pointer{update}, ads[0]); err != nil {
		t.Error(err)
	}

	// Make sure scan returns the lost advertisement when advertising is stopped.
	stops[0]()

	update = <-scanCh
	if err := matchLost([]mojom.Update_Pointer{update}, ads[0]); err != nil {
		t.Error(err)
	}

	// Also it shouldn't affect the other.
	if err := scanAndMatch(d2, `v.InterfaceName="v.io/v23/b"`, ads[1]); err != nil {
		t.Error(err)
	}

	// Stop advertising the remaining one; Shouldn't discover any advertisements.
	stops[1]()
	if err := scanAndMatch(d2, ``); err != nil {
		t.Error(err)
	}
}
