import {calendar, DefaultApiClient, runOnce} from "/dist/z2-bundle.js";

function streamEvents(path, callback) {
    const es = new EventSource(path);
    es.onmessage = (event) => {
        callback(event.data);
    };
}

async function setup() {
    return {
        apiClient: new DefaultApiClient(),
        calendarClient: new calendar.DefaultApiClient(),
        maybeStartFrontendBLE() {
            return Promise.resolve();
        },
        streamEvents,
    };
}

export const getEnv = runOnce(setup);