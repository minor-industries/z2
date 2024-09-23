import {calendar, DefaultApiClient, runOnce, streamEvents} from "/dist/z2-bundle.js";

async function maybeStartFrontendBLE() {
}

function sync(host, lookback, database, logCallback) {
    const sseUrl = `/trigger-sync?host=${encodeURIComponent(host)}&lookback=${lookback}&database=${encodeURIComponent(database)}`;

    const eventSource = new EventSource(sseUrl);

    eventSource.onmessage = function (event) {
        logCallback(event.data);
    };

    // TODO: add a close event and differentiate from log events

    eventSource.onerror = function (err) {
        logCallback("connection closed or errored");
        eventSource.close();
    };
}


async function setup() {
    return {
        apiClient: new DefaultApiClient(),
        calendarClient: new calendar.DefaultApiClient(),
        maybeStartFrontendBLE,
        streamEvents,
        sync,
    };
}

export const getEnv = runOnce(setup);