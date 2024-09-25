import {calendar, DefaultApiClient, runOnce, streamEvents} from "/z2/z2-bundle.js";

async function maybeStartFrontendBLE() {
}

function startSSE(url, logCallback) {
    const eventSource = new EventSource(url);

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

function sync(host, days, database, logCallback) {
    const params = {
        host: host,
        days: days,
        database: database
    };

    const queryString = new URLSearchParams(params).toString();
    const sseUrl = `/trigger-sync?${queryString}`;

    startSSE(sseUrl, logCallback);
}

function backup(logCallback) {
    const sseUrl = `/trigger-backup`;

    startSSE(sseUrl, logCallback);
}


async function setup() {
    return {
        apiClient: new DefaultApiClient(),
        calendarClient: new calendar.DefaultApiClient(),
        maybeStartFrontendBLE,
        streamEvents,
        sync,
        backup,
        defaultSyncConfig: {
            host: "{{.Sync.Host}}",
            days: "{{.Sync.Days}}",
            database: "{{.Sync.Database}}",
        }
    };
}

export const getEnv = runOnce(setup);