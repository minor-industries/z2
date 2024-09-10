import {runGoWasm, WASMConnector} from "./dist/z2-bundle.js";

const resp = {
    "result_sets": [
        {
            "color": "red",
            "date": "2024-06-07",
            "query": "hrm",
            "count": 1
        }
    ]
}

export async function getEnv() {
    window.dbManager = {
        async loadDataAfter() {
            return [];
        },
        async loadDataBetween() {
            return [];
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