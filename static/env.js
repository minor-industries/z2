import {runGoWasm, runOnce, WASMConnector} from "./dist/z2-bundle.js";

async function maybeStartFrontendBLE() {
    const url = new URL(window.location.href);
    const pathParts = url.pathname.split('/');
    const currentPage = pathParts[pathParts.length - 1];

    switch (currentPage) {
        case 'bike.html':
            window.z2GoWasm.startReplay("bike");
            return;
    }
}


async function setup() {
    console.log("running setup");
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
        maybeStartFrontendBLE,
    };
}

export const getEnv = runOnce(setup);