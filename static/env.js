import {runGoWasm, WASMConnector} from "./dist/z2-bundle.js";

export async function getEnv() {
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

    // setInterval(function () {
    //     window.createValue("heartrate", (new Date()).getTime(), 70.0 + Math.random());
    // }, 1000)

    return {
        apiClient: window.goWasmApi,
        calendarClient: window.goWasmCalendar,
        connector: connector,
    };
}