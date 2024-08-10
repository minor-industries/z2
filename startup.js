import './wasm_exec.js';

async function runWasm() {
    const go = new Go();

    const response = await fetch('z2.wasm');
    const bytes = await response.arrayBuffer();
    const module = await WebAssembly.compile(bytes);
    const instance = await WebAssembly.instantiate(module, go.importObject);
    await go.run(instance);
}

await runWasm();
