// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

"use strict";

if (!globalThis.performance) {
  globalThis.performance = {
    now() {
      const [sec, nsec] = process.hrtime();
      return sec * 1000 + nsec / 1000000;
    },
  };
}

if (!globalThis.crypto) {
  const crypto = require("crypto");
  globalThis.crypto = {
    getRandomValues(b) {
      crypto.randomFillSync(b);
    },
  };
}
require("./wasm_exec");
