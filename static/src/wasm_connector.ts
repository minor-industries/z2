import {Handler, Msg} from "rtgraph";
import {decode} from "@msgpack/msgpack"
import {runWasm} from "./run_wasm";

export function runGoWasm(wasmPath: string): Promise<void> {
    const p = new Promise<void>((resolve) => {
        document.addEventListener("wasmReady", () => {
            console.log("WASM is ready!");
            resolve();
        });
    });

    const wasmDonePromise = runWasm(wasmPath);

    return p
}

// TODO: should move into rtgraph itself
export class WASMConnector {
    private readonly subscribe: (
        msg: string,
        callback: (data: any) => void
    ) => void;

    constructor(
        subscribe: (
            msg: string,
            callback: (data: any) => void
        ) => void) {

        this.subscribe = subscribe;
    }

    connect(handler: Handler) {
        const req = handler.subscriptionRequest();

        this.subscribe(JSON.stringify(req), data => {
            const msg = decode(data) as Msg;
            handler.onmessage(msg);
        });
    }
}
