// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Copyright 2014 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// This file is autogenerated by:
//     mojo/public/tools/bindings/mojom_bindings_generator.py
// For:
//     mojom/vanadium/discovery.mojom
//

package io.v.mojo.discovery;

public interface ScanHandler extends org.chromium.mojo.bindings.Interface {

    public interface Proxy extends ScanHandler, org.chromium.mojo.bindings.Interface.Proxy {
    }

    NamedManager<ScanHandler, ScanHandler.Proxy> MANAGER = ScanHandler_Internal.MANAGER;

    void found(Service service);

    void lost(String instanceId);
}
