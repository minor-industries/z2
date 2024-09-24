declare var Go: any;
import './wasm_exec.js';

export async function runWasm(name: string): Promise<void> {
    const go = new Go();

    const response = await fetch(name);
    const bytes = await response.arrayBuffer();
    const module = await WebAssembly.compile(bytes);
    const instance = await WebAssembly.instantiate(module, go.importObject);
    await go.run(instance);
}
