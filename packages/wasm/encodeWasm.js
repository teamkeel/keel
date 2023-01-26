const fs = require("fs");

const wasmCode = fs.readFileSync(__dirname + "/dist/keel.wasm");
const encoded = Buffer.from(wasmCode, "binary").toString("base64");

fs.writeFileSync(
  __dirname + "/dist/wasm.js",
  `
function asciiToBinary(str) {
  if (typeof atob === "function") {
    // this works in the browser
    return atob(str);
  } else {
    // this works in node
    return new Buffer(str, "base64").toString("binary");
  }
}

function decode(encoded) {
  var binaryString = asciiToBinary(encoded);
  var bytes = new Uint8Array(binaryString.length);
  for (var i = 0; i < binaryString.length; i++) {
    bytes[i] = binaryString.charCodeAt(i);
  }
  return bytes.buffer;
}

module.exports.wasm = decode("${encoded}");
`
);
