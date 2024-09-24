import {calendar, DefaultApiClient, runOnce, streamEvents} from "/z2/z2-bundle.js";

async function maybeStartFrontendBLE() {
}

function sync(host, days, database, logCallback) {
    const params = {
        host: host,
        days: days,
        database: database
    };

    const queryString = new URLSearchParams(params).toString();
    const sseUrl = `/trigger-sync?${queryString}`;

    const eventSource = new EventSource(sseUrl);

    eventSource.addEventListener("info", (event) => {
        logCallback(event.data);
    });

    eventSource.addEventListener("server-error", (event) => {
        logCallback("server error: " + event.data);
    });

    eventSource.addEventListener("close", (event) => {
        logCallback("got server close message");
        eventSource.close();
    });

    eventSource.onerror = function (err) {
        logCallback(`connection error`);
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