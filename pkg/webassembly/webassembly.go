package webassembly

// Ignore for now

import (
    "fmt"
    "syscall/js"
)

func add(this js.Value, inputs []js.Value) interface{} {
    var input string = inputs[0].String()

    return "you typed: " + input
}

func main() {
     function init() {
    const go = new Go();
    const result = await WebAssembly.instantiateStreaming(
    fetch("main.wasm"), go.importObject);
    
    var wasmModule = result.instance;
    go.run(wasmModule);
    console.log(wasmModule.exports)
    alt()
}
