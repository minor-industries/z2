import {runGoWasm, runOnce, WASMConnector} from "./dist/z2-bundle.js";

async function setup() {
    window.dbManager = {
        async loadDataAfter() {
            return [];
        },
        async loadDataBetween() {
            return [];
        },
        async insertValue() {
        }
    }; // TODO

    await runGoWasm("/dist/z2.wasm")

    const connector = new WASMConnector(window.subscribe);

    return {
        apiClient: window.goWasmApi,
        calendarClient: window.goWasmCalendar,
        connector: connector,
    };
}

export const getEnv = runOnce(setup);