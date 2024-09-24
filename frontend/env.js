import {runGoWasm, runOnce, WASMConnector} from "./z2/z2-bundle.js";

async function maybeStartFrontendBLE() {
    const url = new URL(window.location.href);
    const pathParts = url.pathname.split('/');
    const currentPage = pathParts[pathParts.length - 1];

    switch (currentPage) {
        case 'bike.html':
            window.z2GoWasm.z2.startReplay("bike");
            window.z2GoWasm.z2.startApp("bike");
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
        },
        async allSeriesNames() {
            return [];
        },
    }; // TODO

    await runGoWasm("/z2/z2.wasm")

    return {
        apiClient: window.z2GoWasm.apiClient,
        calendarClient: window.z2GoWasm.calendarClient,
        connector: new WASMConnector(window.z2GoWasm.z2.subscribe),
        streamEvents: window.z2GoWasm.z2.streamEvents,
        maybeStartFrontendBLE,
        sync: window.z2GoWasm.z2.triggerSync,
    };
}

export const getEnv = runOnce(setup);